package ethereum

import (
	"errors"
	"fmt"
	"go-mythril/analysis"
	"go-mythril/analysis/module"
	"go-mythril/analysis/module/modules"
	"go-mythril/disassembler"
	"go-mythril/laser/ethereum/state"
	"go-mythril/laser/ethereum/strategy"
	"go-mythril/laser/ethereum/transaction"
	"go-mythril/laser/smt/z3"
	"go-mythril/support"
	"go-mythril/utils"
	"reflect"
	"strconv"
	"time"
)

type moduleExecFunc func(globalState *state.GlobalState) []*analysis.Issue

type JumpdestCountAnnotation struct {
	IndexCounter int
	Trace        []int
}

func NewJumpdestCountAnnotation() *JumpdestCountAnnotation {
	return &JumpdestCountAnnotation{
		IndexCounter: 0,
		//Trace:        sync.Map{},
		Trace: make([]int, 0),
	}
}
func (anno *JumpdestCountAnnotation) PersistToWorldState() bool {
	return false
}
func (anno *JumpdestCountAnnotation) PersistOverCalls() bool {
	return false
}
func (anno *JumpdestCountAnnotation) Copy() state.StateAnnotation {
	newTrace := make([]int, 0)
	for _, v := range anno.Trace {
		newTrace = append(newTrace, v)
	}
	return &JumpdestCountAnnotation{
		IndexCounter: anno.IndexCounter,
		Trace:        newTrace,
	}
}
func (anno *JumpdestCountAnnotation) Translate(ctx *z3.Context) state.StateAnnotation {
	return anno.Copy()
}
func (anno *JumpdestCountAnnotation) getIndex() int {
	anno.IndexCounter = anno.IndexCounter + 1
	return anno.IndexCounter
}
func (anno *JumpdestCountAnnotation) Add(callOffset int) bool {
	//_, exist := anno.Trace.LoadOrStore(anno.getIndex(), callOffset)
	//return !exist
	anno.Trace = append(anno.Trace, callOffset)
	return true
}
func (anno *JumpdestCountAnnotation) GetTrace() []int {
	return anno.Trace
}

// #signal
type Signal struct {
	Id        int
	NewStates int
	Finished  bool
	Time      int64
}

type LaserEVM struct {
	ExecutionTimeout int
	CreateTimeout    int
	TransactionCount int
	OpenStates       []*state.WorldState
	OpenStatesSync   *utils.SyncSlice
	FinalState       *state.GlobalState
	/* LoopBound */
	LoopsStrategy strategy.BoundedLoopsStrategy

	InstrPreHook  *map[string][]moduleExecFunc
	InstrPostHook *map[string][]moduleExecFunc
	/* Parallel */
	AfterExecCh       chan Signal
	BeforeExecCh      chan Signal
	LastOpCodeList    []string
	LastAfterExecList []int
	//LastAfterExecList *utils.SyncSlice
	//CtxList []*z3.Context
	CtxList     *utils.SyncSlice
	NewCtxList  []*z3.Context
	TxCtxList   []*z3.Context
	GofuncCount int
	/* Analysis */
	Loader *module.ModuleLoader

	// pltest
	Manager     *Manager
	RuntimeCode string
}

type Manager struct {
	WorkList chan *state.GlobalState

	// TODO: deterministic
	// WorkLists map[int]chan *state.GlobalState

	SignalCh chan Signal

	TotalStates    int
	FinishedStates int
	Duration       int64

	GofuncCount int
	ReqCh       chan int
	RespChs     map[int]chan bool
}

func NewManager(GofuncCount int) *Manager {
	m := Manager{
		WorkList:       make(chan *state.GlobalState, 100000),
		SignalCh:       make(chan Signal, 100000),
		TotalStates:    0,
		FinishedStates: 0,
		Duration:       0,

		GofuncCount: GofuncCount,
		ReqCh:       make(chan int, 100),
		RespChs:     make(map[int]chan bool),
	}

	for i := 0; i < GofuncCount; i++ {
		m.RespChs[i] = make(chan bool, 100)
	}
	return &m
}

func (m *Manager) LogInfo() {
	fmt.Println("Worklist", m.WorkList, "WorklistLen", len(m.WorkList), "TotalStates", m.TotalStates, "FinishedStates", m.FinishedStates)
}

func (m *Manager) Pop() *state.GlobalState {
	return <-m.WorkList
}

func (m *Manager) SignalLoop() {
	fmt.Println("SignalLoop Start")
	for {
		fmt.Println("wait signal")
		select {
		case signal := <-m.SignalCh:
			if signal.Id != -1 {
				m.FinishedStates += 1
			}
			m.TotalStates += signal.NewStates

			m.Duration += signal.Time

			fmt.Println("got signal", signal.Id, signal.NewStates)
			fmt.Println("total", m.TotalStates, "finished", m.FinishedStates)
			fmt.Println("totalDuration", m.Duration)
			if signal.NewStates == 0 && m.TotalStates == m.FinishedStates {
				goto BREAK
			}
		case id := <-m.ReqCh:
			runningWorkers := m.TotalStates - m.FinishedStates
			canSkip := (runningWorkers+1 < m.GofuncCount)

			// no skip
			canSkip = false

			m.RespChs[id] <- canSkip
		}
	}
BREAK:
	fmt.Println("SignalLoop Stop")
}

func NewLaserEVM(ExecutionTimeout int, CreateTimeout int, TransactionCount int, moduleLoader *module.ModuleLoader, cfg *z3.Config, goFuncCount int) *LaserEVM {

	preHook := make(map[string][]moduleExecFunc)
	postHook := make(map[string][]moduleExecFunc)
	opcodes := *support.NewOpcodes()
	for _, v := range opcodes {
		preHook[v.Name] = make([]moduleExecFunc, 0)
		postHook[v.Name] = make([]moduleExecFunc, 0)
	}

	ctxList := make([]interface{}, goFuncCount, goFuncCount)

	// chz
	//for i, _ := range ctxList {
	//	ctxList[i] = z3.NewContext(cfg)
	//}

	evm := LaserEVM{
		ExecutionTimeout: ExecutionTimeout,
		CreateTimeout:    CreateTimeout,
		TransactionCount: TransactionCount,
		OpenStates:       make([]*state.WorldState, 0),
		OpenStatesSync:   utils.NewSyncSlice(),
		FinalState:       nil,
		LoopsStrategy:    strategy.NewBoundedLoopsStrategy(3),

		InstrPreHook:  &preHook,
		InstrPostHook: &postHook,

		AfterExecCh:       make(chan Signal),
		BeforeExecCh:      make(chan Signal),
		LastOpCodeList:    make([]string, goFuncCount, goFuncCount),
		LastAfterExecList: make([]int, goFuncCount),
		//LastAfterExecList: utils.NewSyncSlice(),
		//CtxList: ctxList,
		CtxList:     utils.NewSyncSliceWithArr(ctxList),
		NewCtxList:  make([]*z3.Context, 0),
		TxCtxList:   make([]*z3.Context, 0),
		GofuncCount: goFuncCount,
		Loader:      moduleLoader,
		Manager:     NewManager(goFuncCount),
	}
	evm.registerInstrHooks()
	for i := 0; i < goFuncCount; i++ {
		go evm.Run2(i, cfg)
	}
	return &evm
}

func (evm *LaserEVM) registerInstrHooks() {
	preHook := *evm.InstrPreHook
	postHook := *evm.InstrPostHook
	for _, module := range evm.Loader.Modules {
		for _, op := range module.GetPreHooks() {
			preHook[op] = append(preHook[op], module.Execute)
		}
		for _, op := range module.GetPostHooks() {
			postHook[op] = append(postHook[op], module.Execute)
		}
	}
}

func (evm *LaserEVM) Refresh() {
	// evm.Manager.WorkList = make(chan *state.GlobalState, 100000)
	evm.OpenStates = make([]*state.WorldState, 0)
	evm.OpenStatesSync = utils.NewSyncSlice()
	evm.FinalState = nil

	// decouple
	//evm.CtxList = make([]*z3.Context, evm.GofuncCount, evm.GofuncCount)
	ctxList := make([]interface{}, evm.GofuncCount, evm.GofuncCount)
	evm.CtxList = utils.NewSyncSliceWithArr(ctxList)

	evm.BeforeExecCh = make(chan Signal)
}

func (evm *LaserEVM) exec() {
	id := 0
LOOP:
	for {
		// When there is no newState in channel, exit the iteration
		fmt.Println("evm workList:", len(evm.Manager.WorkList), evm.Manager.WorkList)
		if len(evm.Manager.WorkList) == 0 {
			break LOOP
		}

		globalState := <-evm.Manager.WorkList

		// get_strategic_global_state in bounded_loops
		annotations := globalState.GetAnnotations(reflect.TypeOf(&JumpdestCountAnnotation{}))
		var annotation *JumpdestCountAnnotation
		if len(annotations) == 0 {
			annotation = NewJumpdestCountAnnotation()
			globalState.Annotate(annotation)
		} else {
			annotation = annotations[0].(*JumpdestCountAnnotation)
		}
		curInstr := globalState.GetCurrentInstruction()
		//evm.Trace = append(evm.Trace, curInstr.Address)
		annotation.Add(curInstr.Address)

		lastInstr := globalState.Environment.Code.InstructionList[globalState.Mstate.LastPc]
		if curInstr.OpCode.Name == "JUMPDEST" {
			fmt.Println("InJumpdest-LastPc:", globalState.Mstate.LastPc)
			if globalState.Mstate.LastPc == 0 {
				count := evm.LoopsStrategy.GetLoopCount(annotation.GetTrace())
				if "*state.ContractCreationTransaction" == reflect.TypeOf(globalState.CurrentTransaction()).String() && count < 8 {
					goto EXEC
				} else if count > evm.LoopsStrategy.Bound {
					fmt.Println("hahah", lastInstr.OpCode.Name, lastInstr.Address)
					fmt.Println("Loop bound reached, skipping state", count, evm.LoopsStrategy.Bound)
					modules.CheckPotentialIssues(globalState)
					fmt.Println("openStatesLen:", len(evm.OpenStates))
					fmt.Println("send signal", id)
					evm.Manager.SignalCh <- Signal{
						Id:        id,
						NewStates: 0,
					}
					continue
				}
			} else {
				// after jumpi
				globalState.Mstate.LastPc = 0
				goto EXEC
			}
		}
	EXEC:
		newStates, opcode := evm.ExecuteState(globalState)
		fmt.Println(id, globalState, opcode)

		if len(newStates) == 2 {
			tmpStates := make([]*state.GlobalState, 0)
			for _, s := range newStates {
				if s.WorldState.Constraints.IsPossible() {
					tmpStates = append(tmpStates, s)
				}
			}
			newStates = tmpStates
		}

		if len(newStates) == 2 {
			ctx := newStates[1].Z3ctx.Copy()
			evm.NewCtxList = append(evm.NewCtxList, ctx)
			newStates[1].Z3ctx = ctx
		}

		// If args.sparse_pruning is False:
		//newPossibleStates := make([]*state.GlobalState, 0)

		//fmt.Println("newPossibleStatesLen:", len(newPossibleStates))
		//for _, newState := range newPossibleStates {
		//	evm.Manager.WorkList <- newState
		//}

		fmt.Println(id, "done", globalState, opcode)

		if opcode == "JUMPI" {
			fmt.Println("#JUMPI", globalState.GetCurrentInstruction().Address, globalState.Mstate.Pc, "lenRes:", len(newStates))
		}
		fmt.Println("evmOpenStatesLen:", len(evm.OpenStates))
		for _, v := range newStates {
			fmt.Println("#LenConstraints:", v.WorldState.Constraints.Length())
		}

		fmt.Println("==============================================================================")

		if len(newStates) == 0 {
			fmt.Println("#LenConstraintsEnd:", globalState.WorldState.Constraints.Length())
			evm.FinalState = globalState
			if "*state.MessageCallTransaction" == reflect.TypeOf(globalState.CurrentTransaction()).String() {
				if opcode != "REVERT" && opcode != "INVALID" {
					fmt.Println("txEnd: append ws!")
					evm.OpenStates = append(evm.OpenStates, globalState.WorldState)
				}
			}
			modules.CheckPotentialIssues(globalState)

			fmt.Println("openStatesLen:", len(evm.OpenStates))
		}

		for _, newState := range newStates {
			evm.Manager.WorkList <- newState
		}
		fmt.Println("send signal", 10086)
		evm.Manager.SignalCh <- Signal{
			Id:        10086,
			NewStates: len(newStates),
		}

		id++
	}
	fmt.Println("evm.Exec: One tx end!")
}

func (evm *LaserEVM) multiExec2(cfg *z3.Config) {
	fmt.Println("multiExec2 evm.Manager.WorkList", len(evm.Manager.WorkList))
	evm.Manager.SignalLoop()
}

func (evm *LaserEVM) multiExec(cfg *z3.Config) {
	evm.multiExec2(cfg)
	return

	for i := 0; i < evm.GofuncCount; i++ {
		go evm.Run(i, cfg)
	}

	// #signal
	beforeExecSignals := make([]bool, evm.GofuncCount)
	evm.LastAfterExecList = make([]int, evm.GofuncCount, evm.GofuncCount)
	for i, _ := range evm.LastAfterExecList {
		evm.LastAfterExecList[i] = -1
	}
	//arr := []interface{}{-1,-1,-1,-1}
	//evm.LastAfterExecList = utils.NewSyncSliceWithArr(arr)
	//endOpCodeList := []string{"STOP", "RETURN", "REVERT", "INVALID", "JUMPDEST"}
	//afterExecSignals := make([]bool, evm.GofuncCount)

LOOP:
	for {
		/*
			There are two situations for exiting.
			1. All goroutines don't generate new globalStates.
			2. There is no globalState in channel, so all goroutines will be blocked.
		*/

		select {
		//case signal := <-evm.AfterExecCh:
		//	// true == didn't produce a newState
		//	afterExecSignals[signal.Id] = signal.Finished
		//	//fmt.Println("afterExecSignal:", signal.Id, signal.Finished)
		//	allFinished := true
		//	for _, finished := range afterExecSignals {
		//		//fmt.Println(i, finished)
		//		if !finished {
		//			allFinished = false
		//		}
		//	}
		//	//fmt.Println("situation 1", allFinished)
		//	if allFinished && len(evm.Manager.WorkList) == 0 {
		//		fmt.Println("break in situation 1")
		//		fmt.Println("workListLen:", len(evm.Manager.WorkList))
		//		break LOOP
		//	}
		case signal := <-evm.BeforeExecCh:
			// true == didn't get a state
			// #signal
			beforeExecSignals[signal.Id] = signal.Finished
			//fmt.Println("afterSingle:", beforeExecSignals)
			// TODO: if the number of goroutines is greater than globalStates, some goroutines initial states will be false.
			// TODO: so the var allNoStates will be false and the program can't stop.
			//allNoStates := true
			//for _, noState := range beforeExecSignals {
			//	if !noState {
			//		// #id goroutine is running a globalState
			//		allNoStates = false
			//		break
			//	}
			//}
			allNoStates := true
			for _, noState := range beforeExecSignals {
				if !noState {
					allNoStates = false
					break
				}
			}

			//// use with readWithSelect3()
			//allEndOpCode := true
			//for _, opcode := range evm.LastOpCodeList {
			//	if !utils.In(opcode, endOpCodeList) && opcode != "" {
			//		allEndOpCode = false
			//		break
			//	}
			//}
			//
			//// flag == true means it is the initial state.
			//flag := true
			//for _, opcode := range evm.LastOpCodeList {
			//	if opcode != "" {
			//		flag = false
			//		break
			//	}
			//}
			//
			// use with readWithSelect3()
			// resFlag == true means all goroutines produce 0 states.
			resFlag := true
			for _, resNum := range evm.LastAfterExecList {
				if resNum != 0 {
					resFlag = false
					break
				}
			}
			//for _, resNum := range evm.LastAfterExecList.Elements() {
			//	if resNum != 0 {
			//		resFlag = false
			//		break
			//	}
			//}

			if allNoStates && resFlag && len(evm.Manager.WorkList) == 0 {
				fmt.Println("break in situation 2")
				fmt.Println("workListLen:", len(evm.Manager.WorkList))

				break LOOP
			}

			//if allNoStates && len(evm.Manager.WorkList) == 0 {
			//	fmt.Println("break in situation 2")
			//	fmt.Println("beforeExecSignals", beforeExecSignals)
			//	fmt.Println("workListLen:", len(evm.Manager.WorkList))
			//	break LOOP
			//}

			//default:
			//	fmt.Println("In Loop for default-break")
			//	break LOOP
		}
	}

	//close(evm.AfterExecCh)
	//close(evm.BeforeExecCh)
}

func (evm *LaserEVM) SingleSymExec(creationCode string, runtimeCode string, contractName string, ctx *z3.Context) {
	fmt.Println("Single Goroutine Symbolic Executing")
	fmt.Println("")
	// CreationTx
	newAccount := ExecuteContractCreation(evm, creationCode, contractName, ctx, false, nil)

	// MessageTx
	inputStrArr := support.GetArgsInstance().TransactionSequences
	for i := 0; i < evm.TransactionCount; i++ {
		fmt.Println("beforeMsgCall-OpenStatesLen:", len(evm.OpenStates))
		tmpOpenStates := make([]*state.WorldState, 0)
		for _, ws := range evm.OpenStates {
			if ws.Constraints.IsPossible() {
				tmpOpenStates = append(tmpOpenStates, ws)
				fmt.Println(ws)
			}
		}
		evm.OpenStates = tmpOpenStates
		fmt.Println("afterMsgCall-OpenStatesLen:", len(evm.OpenStates))

		fmt.Println("msgTx", i, ":start")
		ExecuteMessageCall(evm, runtimeCode, inputStrArr[i], ctx, newAccount.Address, false, nil)
		fmt.Println("workList len:", len(evm.Manager.WorkList))
		fmt.Println("openStates len:", len(evm.OpenStates))
		fmt.Println("msgTx", i, ":end")

	}
}

func (evm *LaserEVM) MultiSymExec(creationCode string, runtimeCode string, contractName string, ctx *z3.Context, cfg *z3.Config) {
	fmt.Println("Multi-Goroutines Symbolic Executing")
	fmt.Println("")
	// CreationTx
	newAccount := ExecuteContractCreation(evm, creationCode, contractName, ctx, false, nil)
	evm.OpenStatesSync.Append(evm.OpenStates[0])
	evm.RuntimeCode = runtimeCode

	// evm.Manager.WorkList = make(chan *state.GlobalState, 100000)

	fmt.Println("beforeMsgCall-OpenStatesLen:", evm.OpenStatesSync.Length())
	tmpOpenStates := make([]interface{}, 0)
	for _, ws := range evm.OpenStatesSync.Elements() {
		//tmpWs := ws.Translate(ctx)
		if ws.(*state.WorldState).Constraints.IsPossible() {
			tmpOpenStates = append(tmpOpenStates, ws)
		}
	}
	evm.OpenStatesSync = utils.NewSyncSliceWithArr(tmpOpenStates)
	fmt.Println("afterMsgCall-OpenStatesLen:", evm.OpenStatesSync.Length())

	ExecuteMessageCallMany(evm, runtimeCode, ctx, newAccount.Address, true, cfg, evm.TransactionCount)
}

func (evm *LaserEVM) SingleSymExecMsgCallOnly(runtimeCode string, contractName string, ctx *z3.Context) {
	fmt.Println("Symbolic Executing: ", runtimeCode)
	fmt.Println("")
	inputStrArr := support.GetArgsInstance().TransactionSequences
	for i := 0; i < evm.TransactionCount; i++ {
		ExecuteMessageCallOnly(evm, runtimeCode, contractName, inputStrArr[i], ctx, false, nil)
		fmt.Println("normalExec:", i, "tx")
	}
}

func (evm *LaserEVM) MultiSymExecMsgCallOnly(runtimeCode string, contractName string, ctx *z3.Context, cfg *z3.Config) {
	fmt.Println("Symbolic Executing: ", runtimeCode)
	fmt.Println("")
	inputStrArr := support.GetArgsInstance().TransactionSequences
	for i := 0; i < evm.TransactionCount; i++ {
		ExecuteMessageCallOnly(evm, runtimeCode, contractName, inputStrArr[i], ctx, true, cfg)
		fmt.Println("normalExec:", i, "tx")
	}
}

func (evm *LaserEVM) ExecuteState(globalState *state.GlobalState) ([]*state.GlobalState, string) {
	instrs := globalState.Environment.Code.InstructionList
	opcode := instrs[globalState.Mstate.Pc].OpCode.Name
	//fmt.Println("opcode:", opcode)
	preHooks := *evm.InstrPreHook
	postHooks := *evm.InstrPostHook

	instr := NewInstruction(opcode, preHooks[opcode], postHooks[opcode])
	newGlobalStates := instr.Evaluate(globalState)

	return newGlobalStates, opcode
}

func (evm *LaserEVM) ManageCFG(opcode string, newStates []*state.GlobalState) {
	statesLen := len(newStates)
	if opcode == "JUMP" {
		if statesLen <= 1 {
			return
		}
		for _, state := range newStates {
			evm.newNodeState(state, UNCONDITIONAL, nil)
		}
	} else if opcode == "JUMPI" {
		if statesLen <= 2 {
			return
		}
		for _, state := range newStates {
			evm.newNodeState(state, CONDITIONAL, state.WorldState.Constraints.ConstraintList[state.WorldState.Constraints.Length()-1])
		}
	} else if utils.In(opcode, []string{"SLOAD", "SSTORE"}) && len(newStates) > 1 {
		for _, state := range newStates {
			evm.newNodeState(state, CONDITIONAL, state.WorldState.Constraints.ConstraintList[state.WorldState.Constraints.Length()-1])
		}
	} else if opcode == "RETURN" {
		for _, state := range newStates {
			evm.newNodeState(state, RETURN, nil)
		}
	}
	for _, state := range newStates {
		// TODO: globalState.node
		fmt.Println(state)
	}
}

// TODO:
func (evm *LaserEVM) newNodeState(state *state.GlobalState, edgeType JumpType, condition *z3.Bool) {
	// default: edge_type=JumpType.UNCONDITIONAL, condition=None
}

func (evm *LaserEVM) executeTransactions(address string) {

	for i := 0; i < evm.TransactionCount; i++ {
		if len(evm.OpenStates) == 0 {
			break
		}
		//oldStatesCount := len(evm.OpenStates)
		fmt.Println("Starting message call transaction, iteration:", i, len(evm.OpenStates), "initial states")
		args := support.NewArgs()
		funcHashes := args.TransactionSequences[i]
		fmt.Println(funcHashes)
		// TODO: ExecuteMessageCall()
	}
}

func readWithSelect(evm *LaserEVM, id int) (*state.GlobalState, error) {
	select {
	case globalState := <-evm.Manager.WorkList:
		//evm.BeforeExecCh <- Signal{
		//	Id:       id,
		//	Finished: false,
		//}
		return globalState, nil

	default:
		//evm.BeforeExecCh <- Signal{
		//	Id:       id,
		//	Finished: true,
		//}
		return nil, errors.New("evm.Manager.WorkList is empty")
	}

}

func readWithSelect3(evm *LaserEVM, id int) (*state.GlobalState, error) {
	//select {
	//case globalState := <-evm.Manager.WorkList:
	//	ctx := evm.CtxList[id]
	//	//ctx := evm.CtxList.Load(id)
	//	if globalState.Z3ctx == ctx {
	//		return globalState, nil
	//	} else {
	//		for i, m := range evm.CtxList {
	//			if i != id {
	//				if globalState.Z3ctx == m {
	//					evm.Manager.WorkList <- globalState
	//					return nil, errors.New("get other's state, push it back to WorkList")
	//					//return globalState, errors.New("get other's state, push it back to WorkList")
	//				}
	//			}
	//		}
	//		// chz
	//		if ctx == nil {
	//			evm.CtxList[id] = globalState.Z3ctx
	//			//evm.CtxList.SetItem(id, globalState.Z3ctx)
	//			return globalState, nil
	//		} else {
	//			evm.Manager.WorkList <- globalState
	//			return nil, errors.New("my own states haven't been executed")
	//		}
	//	}
	//default:
	//	return nil, errors.New("evm.Manager.WorkList is empty")
	//}
	return nil, nil
}

func (evm *LaserEVM) goroutineAvailable(id int) bool {
	flag := false
	arr := evm.CtxList
	fmt.Println(id, "*****************************")
	//fmt.Println("ctxListLen:", len(evm.CtxList), arr)

	for i, item := range arr.Elements() {
		fmt.Println(i, "cp", item, item == nil)
		if i != id {
			if item == nil {
				flag = true
				break
			}
		}
	}
	fmt.Println("*****************************")
	return flag
}

var MaxRLimitCount int64

func SetMaxRLimitCount(value int64) {
	MaxRLimitCount = value
	fmt.Println("MaxRLimitCount:", MaxRLimitCount)
}

func (evm *LaserEVM) Run2(id int, cfg *z3.Config) {
	fmt.Println("Run2 Start", id)

	for {
		fmt.Println("Run2 Wait", id)
		evm.Manager.LogInfo()
		globalState := evm.Manager.Pop()
		// TODO canSkip here?
		evm.Manager.ReqCh <- id
		canSkip := <-evm.Manager.RespChs[id]

		if globalState.NeedIsPossible {
			sat, rlimit := globalState.WorldState.Constraints.IsPossibleRlimit()
			if sat {
				globalState.RLimitCount += rlimit
			} else {
				fmt.Println("send signal", id)
				evm.Manager.SignalCh <- Signal{
					Id:        id,
					NewStates: 0,
				}
				continue
			}
		}
		fmt.Println("Run2 Got", id)

		// decouple
		//evm.CtxList[id] = globalState.Z3ctx
		evm.CtxList.SetItem(id, globalState.Z3ctx)

		evm.Manager.LogInfo()
		annotations := globalState.GetAnnotations(reflect.TypeOf(&JumpdestCountAnnotation{}))
		var annotation *JumpdestCountAnnotation
		if len(annotations) == 0 {
			annotation = NewJumpdestCountAnnotation()
			globalState.Annotate(annotation)
		} else {
			annotation = annotations[0].(*JumpdestCountAnnotation)
		}
		curInstr := globalState.GetCurrentInstruction()
		//evm.Trace = append(evm.Trace, curInstr.Address)
		annotation.Add(curInstr.Address)

		lastInstr := globalState.Environment.Code.InstructionList[globalState.Mstate.LastPc]
		if curInstr.OpCode.Name == "JUMPDEST" {
			fmt.Println("InJumpdest-LastPc:", globalState.Mstate.LastPc)

			if globalState.Mstate.LastPc == 0 {
				count := evm.LoopsStrategy.GetLoopCount(annotation.GetTrace())
				if "*state.ContractCreationTransaction" == reflect.TypeOf(globalState.CurrentTransaction()).String() && count < 8 {
					goto EXEC
				} else if count > evm.LoopsStrategy.Bound {
					fmt.Println("hahah", lastInstr.OpCode.Name, lastInstr.Address)
					fmt.Println("Loop bound reached, skipping state", count, evm.LoopsStrategy.Bound)
					modules.CheckPotentialIssues(globalState)
					fmt.Println("LenOpenStates:", evm.OpenStatesSync.Length())
					evm.LastOpCodeList[id] = "JUMPDEST"
					evm.LastAfterExecList[id] = 0
					//evm.LastAfterExecList.SetItem(id, 0)
					// decouple
					//evm.CtxList[id] = nil
					evm.CtxList.SetItem(id, nil)

					fmt.Println("send signal", id)
					evm.Manager.SignalCh <- Signal{
						Id:        id,
						NewStates: 0,
					}
					continue
				}
			} else {
				// after jumpi
				globalState.Mstate.LastPc = 0
				goto EXEC
			}
		}

	EXEC:
		newStates, opcode := evm.ExecuteState(globalState)

		fmt.Println(id, globalState, opcode)

		// Decouple
		// TODO canSkip here?
		// canskip
		//evm.Manager.ReqCh <- id
		//canSkip := <-evm.Manager.RespChs[id]
		if len(newStates) == 2 && !canSkip {
			// if globalState.RLimitCount > MaxRLimitCount {
			// 	newStates = make([]*state.GlobalState, 0)
			// }
			for _, s := range newStates {
				s.NeedIsPossible = true
			}
		} else {
			for _, s := range newStates {
				s.NeedIsPossible = false
			}
		}

		start := time.Now()

		if len(newStates) == 2 {
			ctx := z3.NewContext(cfg)
			newStates[1].Translate(ctx)
			// TODO: beng?
			ctx0 := z3.NewContext(cfg)
			newStates[0].Translate(ctx0)

		}
		var duration int64
		duration = time.Since(start).Milliseconds()
		fmt.Println("DurationInSym.go:", duration)

		fmt.Println(id, "done", globalState, opcode, globalState.Z3ctx)
		fmt.Println("evmOpenStatesLen:", evm.OpenStatesSync.Length(), "evmWorkListLen:", len(evm.Manager.WorkList))
		if opcode == "JUMPI" {
			fmt.Println("#JUMPI", globalState.GetCurrentInstruction().Address, globalState.Mstate.Pc, "lenRes:", len(newStates))
		}
		for _, v := range newStates {
			fmt.Println("#LenConstraints:", v.WorldState.Constraints.Length())
		}
		fmt.Println("===========================================================================")

		endOpcodeList := []string{"STOP", "RETURN", "REVERT", "SELFDESTRUCT"}
		if utils.In(opcode, endOpcodeList) {
			fmt.Println("#LenConstraintsEnd:", globalState.WorldState.Constraints.Length())
			evm.FinalState = globalState
			if "*state.MessageCallTransaction" == reflect.TypeOf(globalState.CurrentTransaction()).String() {
				if opcode != "REVERT" {
					// evm.OpenStatesSync.Append(globalState.WorldState)
					relayGlobalState := OpenStateRelay(globalState.WorldState, evm, evm.RuntimeCode, cfg)
					if relayGlobalState != nil {
						newStates = append(newStates, relayGlobalState)
					}
					modules.CheckPotentialIssues(globalState)
				}
			}

			// decouple
			//evm.CtxList[id] = nil
			evm.CtxList.SetItem(id, nil)
		}

		evm.LastOpCodeList[id] = opcode
		evm.LastAfterExecList[id] = len(newStates)

		for _, newState := range newStates {
			evm.Manager.WorkList <- newState
		}
		fmt.Println("Run2 send signal", id)
		evm.Manager.SignalCh <- Signal{
			Id:        id,
			NewStates: len(newStates),
			Time:      duration,
		}
	}
	fmt.Println("Run2 Stop", id)
}

func (evm *LaserEVM) Run(id int, cfg *z3.Config) {
	fmt.Println("Run")
	/*
		for {
			//ctx := z3.NewContext(cfg)
			//globalState, err := readWithSelect(evm, id)
			//globalState, err := readWithSelect2(evm, id)
			globalState, err := readWithSelect3(evm, id)
			//fmt.Println("Run", id, globalState == nil)
			evm.BeforeExecCh <- Signal{
				Id:       id,
				Finished: globalState == nil && err.Error() == "evm.Manager.WorkList is empty",
				//Finished: globalState == nil,
			}

			if globalState != nil && err == nil {

				fmt.Println("evm workList:", len(evm.Manager.WorkList))
				// chz
				//globalState.Translate(evm.CtxList[id])
				// get_strategic_global_state in bounded_loops
				annotations := globalState.GetAnnotations(reflect.TypeOf(&JumpdestCountAnnotation{}))
				var annotation *JumpdestCountAnnotation
				if len(annotations) == 0 {
					annotation = NewJumpdestCountAnnotation()
					globalState.Annotate(annotation)
				} else {
					annotation = annotations[0].(*JumpdestCountAnnotation)
				}
				curInstr := globalState.GetCurrentInstruction()
				//evm.Trace = append(evm.Trace, curInstr.Address)
				annotation.Add(curInstr.Address)

				lastInstr := globalState.Environment.Code.InstructionList[globalState.Mstate.LastPc]
				if curInstr.OpCode.Name == "JUMPDEST" {
					fmt.Println("InJumpdest-LastPc:", globalState.Mstate.LastPc)

					if globalState.Mstate.LastPc == 0 {
						count := evm.LoopsStrategy.GetLoopCount(annotation.GetTrace())
						if "*state.ContractCreationTransaction" == reflect.TypeOf(globalState.CurrentTransaction()).String() && count < 8 {
							goto EXEC
						} else if count > evm.LoopsStrategy.Bound {
							fmt.Println("hahah", lastInstr.OpCode.Name, lastInstr.Address)
							fmt.Println("Loop bound reached, skipping state", count, evm.LoopsStrategy.Bound)
							modules.CheckPotentialIssues(globalState)
							fmt.Println("LenOpenStates:", evm.OpenStatesSync.Length())
							evm.LastOpCodeList[id] = "JUMPDEST"
							evm.LastAfterExecList[id] = 0
							//evm.LastAfterExecList.SetItem(id, 0)
							evm.CtxList[id] = nil
							//evm.CtxList.SetItem(id, nil)
							continue
						}
					} else {
						// after jumpi
						globalState.Mstate.LastPc = 0
						goto EXEC
					}
				}

			EXEC:
				newStates, opcode := evm.ExecuteState(globalState)

				fmt.Println(id, globalState, opcode)

				// Decouple
				if len(newStates) == 2 {
					//if evm.goroutineAvailable(id) {
					//	tmpStates := make([]*state.GlobalState, 0)
					//	for _, s := range newStates {
					//		if s.WorldState.Constraints.IsPossible() {
					//			tmpStates = append(tmpStates, s)
					//		}
					//	}
					//	newStates = tmpStates
					//}
					tmpStates := make([]*state.GlobalState, 0)
					for _, s := range newStates {
						if s.WorldState.Constraints.IsPossible() {
							tmpStates = append(tmpStates, s)
						}
					}
					newStates = tmpStates
				}

				if len(newStates) == 2 {
					ctx := z3.NewContext(cfg)
					evm.NewCtxList = append(evm.NewCtxList, ctx)
					newStates[1].Translate(ctx)
				}

				//evm.AfterExecCh <- Signal{
				//	Id:       id,
				//	Finished: len(newStates) == 0,
				//}

				fmt.Println(id, "done", globalState, opcode, globalState.Z3ctx)
				fmt.Println("evmOpenStatesLen:", evm.OpenStatesSync.Length(), "evmWorkListLen:", len(evm.Manager.WorkList))
				if opcode == "JUMPI" {
					fmt.Println("#JUMPI", globalState.GetCurrentInstruction().Address, globalState.Mstate.Pc, "lenRes:", len(newStates))
				}
				for _, v := range newStates {
					fmt.Println("#LenConstraints:", v.WorldState.Constraints.Length())
				}
				fmt.Println("===========================================================================")
				if len(newStates) == 0 {

					fmt.Println("#LenConstraintsEnd:", globalState.WorldState.Constraints.Length())
					evm.FinalState = globalState
					if "*state.MessageCallTransaction" == reflect.TypeOf(globalState.CurrentTransaction()).String() {
						if opcode != "REVERT" && opcode != "INVALID" {
							evm.OpenStatesSync.Append(globalState.WorldState)
							//evm.OpenStates = append(evm.OpenStates, globalState.WorldState)
						}
					}
					modules.CheckPotentialIssues(globalState)
					evm.CtxList[id] = nil
					//evm.CtxList.SetItem(id, nil)
				}

				evm.LastOpCodeList[id] = opcode
				evm.LastAfterExecList[id] = len(newStates)
				//evm.LastAfterExecList.SetItem(id, len(newStates))

				for _, newState := range newStates {
					evm.Manager.WorkList <- newState
				}
			}
	*/
	/* peilin
	globalState := <-evm.Manager.WorkList
	//evm.BeginCh <- id
	fmt.Println(id, globalState)
	newStates, opcode := evm.ExecuteState(globalState)
	evm.ManageCFG(opcode, newStates)

	for _, newState := range newStates {
		evm.Manager.WorkList <- newState
	}
	fmt.Println(id, "done", globalState, opcode)
	fmt.Println("======================================================")
	if opcode == "STOP" || opcode == "RETURN" {
		modules.CheckPotentialIssues(globalState)
		for _, detector := range evm.Loader.Modules {
			issues := detector.GetIssues()
			for _, issue := range issues {
				fmt.Println("+++++++++++++++++++++++++++++++++++")
				fmt.Println("ContractName:", issue.Contract)
				fmt.Println("FunctionName:", issue.FunctionName)
				fmt.Println("Title:", issue.Title)
				fmt.Println("SWCID:", issue.SWCID)
				fmt.Println("Address:", issue.Address)
				fmt.Println("Severity:", issue.Severity)

			}
		}
		fmt.Println("+++++++++++++++++++++++++++++++++++")
		// TODO: a better way for exiting
		//panic("we have already reached the end of code!!!")
	}
	// TODO not good for sleep
	// time.Sleep(100 * time.Millisecond)
	// evm.EndCh <- id
	evm.SignalCh <- Signal{
		Id:       id,
		Finished: (len(newStates) == 0),
	}
	fmt.Println("signal", id, len(newStates) == 0)
	*/
}

func ExecuteContractCreation(evm *LaserEVM, creationCode string, contractName string, ctx *z3.Context, multiple bool, cfg *z3.Config) *state.Account {

	worldState := state.NewWordState(ctx)
	evm.OpenStates = append(evm.OpenStates, worldState)
	//evm.OpenStates.Append(worldState)
	ACTORS := transaction.NewActors(ctx)
	txId := state.GetNextTransactionId()
	code := disassembler.NewDisasembly(creationCode)
	fmt.Println("In creationId:", txId)

	account := worldState.CreateAccount(0, true, ACTORS.GetCreator(), nil, code, contractName)

	tx := &state.ContractCreationTransaction{
		WorldState:    worldState,
		Code:          code,
		CalleeAccount: account,
		Caller:        ACTORS.GetCreator(),
		Calldata:      state.NewSymbolicCalldata(txId, ctx),
		GasPrice:      10,
		GasLimit:      8000000,
		CallValue:     0,
		Origin:        ACTORS.GetCreator(),
		Basefee:       ctx.NewBitvecVal(1000, 256),
		Ctx:           ctx,
		Id:            txId,
	}
	setupGlobalStateForExecution(evm, tx)

	fmt.Println("########################################################################################")
	fmt.Println("CreationTx Execute!")
	if !multiple {
		evm.exec()
		evm.Manager.SignalLoop()
	} else {
		evm.multiExec(cfg)
	}
	fmt.Println("CreationTx End!")
	fmt.Println("########################################################################################")

	// Tips: the final Storage pass
	newAccount := evm.FinalState.Environment.ActiveAccount
	evm.OpenStates = []*state.WorldState{evm.FinalState.WorldState}
	//evm.OpenStates = utils.NewSyncIssueSlice()
	//evm.OpenStates.Append(evm.FinalState.WorldState)

	return newAccount
}

func ExecuteMessageCall(evm *LaserEVM, runtimeCode string, inputStr string, ctx *z3.Context, address *z3.Bitvec, multiple bool, cfg *z3.Config) {

	if !multiple {
		if len(evm.OpenStates) <= 0 {
			//panic("ExecuteMessageCall empty openStates!")
			fmt.Println("ExecuteMessageCall empty openStates!")
			// only in multipleSymExec
			return
		}
		for _, openState := range evm.OpenStates {
			getNewMsgTx(openState, evm, runtimeCode, inputStr, ctx, address, multiple, cfg)
		}
		evm.OpenStates = make([]*state.WorldState, 0)
	} else {
		if evm.OpenStatesSync.Length() <= 0 {
			//panic("ExecuteMessageCall empty openStates!")
			fmt.Println("ExecuteMessageCall empty openStates!Multiple")
			// only in multipleSymExec
			return
		}
		for _, openState := range evm.OpenStatesSync.Elements() {
			getNewMsgTx(openState.(*state.WorldState), evm, runtimeCode, inputStr, ctx, address, multiple, cfg)
		}
		evm.OpenStatesSync = utils.NewSyncSlice()
	}

	fmt.Println("########################################################################################")
	fmt.Println("MessageTx Execute!")
	if !multiple {
		evm.exec()
	} else {
		evm.multiExec(cfg)
	}
	fmt.Println("MessageTx End!")
	fmt.Println("########################################################################################")
}

func ExecuteMessageCallMany(evm *LaserEVM, runtimeCode string, ctx *z3.Context, address *z3.Bitvec, multiple bool, cfg *z3.Config, txCount int) {

	if !multiple {
		if len(evm.OpenStates) <= 0 {
			//panic("ExecuteMessageCall empty openStates!")
			fmt.Println("ExecuteMessageCall empty openStates!")
			// only in multipleSymExec
			return
		}
		for _, openState := range evm.OpenStates {
			OpenStateInit(openState, evm, runtimeCode, ctx, address, multiple, cfg, txCount)
		}
		evm.OpenStates = make([]*state.WorldState, 0)
	} else {
		if evm.OpenStatesSync.Length() <= 0 {
			//panic("ExecuteMessageCall empty openStates!")
			fmt.Println("ExecuteMessageCall empty openStates!Multiple")
			// only in multipleSymExec
			return
		}
		for _, openState := range evm.OpenStatesSync.Elements() {
			OpenStateInit(openState.(*state.WorldState), evm, runtimeCode, ctx, address, multiple, cfg, txCount)
		}
		evm.OpenStatesSync = utils.NewSyncSlice()
	}

	fmt.Println("########################################################################################")
	fmt.Println("MessageTx Execute!")
	if !multiple {
		evm.exec()
	} else {
		evm.multiExec(cfg)
	}
	fmt.Println("MessageTx End!")
	fmt.Println("########################################################################################")
}

func OpenStateInit(openState *state.WorldState, evm *LaserEVM, runtimeCode string, ctx *z3.Context, address *z3.Bitvec, multiple bool, cfg *z3.Config, txCount int) {
	var txCtx *z3.Context
	txCtx = z3.NewContext(cfg)

	openState.TransactionCount = txCount
	openState.TransactionIdInt++
	fmt.Println("OpenStateInit", openState.TransactionIdInt, openState.TransactionCount)
	openState.ContractAddress = address
	txId := strconv.Itoa(openState.TransactionIdInt)

	externalSender := txCtx.NewBitvec("sender_"+txId, 256)
	txCode := disassembler.NewDisasembly(runtimeCode)
	fmt.Println("In msgId:", txId)
	tx := &state.MessageCallTransaction{
		WorldState:    openState.Translate(txCtx),
		Code:          txCode,
		CalleeAccount: openState.Translate(txCtx).AccountsExistOrLoad(openState.ContractAddress.Translate(txCtx)),
		//CalleeAccount: openState.AccountsExistOrLoad(address),
		Caller:   externalSender,
		Calldata: state.NewSymbolicCalldata(txId, txCtx),
		//Calldata: NewConcreteCalldata(txId, calldataList, ctx),
		GasPrice:  10,
		GasLimit:  8000000,
		CallValue: 0,
		Origin:    externalSender,
		Basefee:   txCtx.NewBitvecVal(1000, 256),
		Ctx:       txCtx,
		Id:        txId,
	}

	globalState := tx.InitialGlobalState()
	ACTORS := transaction.NewActors(globalState.Z3ctx)
	constraint := tx.GetCaller().Eq(ACTORS.GetCreator()).Or(tx.GetCaller().Eq(ACTORS.GetAttacker()), tx.GetCaller().Eq(ACTORS.GetSomeGuy()))
	globalState.WorldState.Constraints.Add(constraint)
	globalState.WorldState.TransactionSequence = append(globalState.WorldState.TransactionSequence, tx)
	evm.Manager.WorkList <- globalState
	evm.Manager.SignalCh <- Signal{
		Id:        -1,
		NewStates: 1,
	}
}

func OpenStateRelay(openState *state.WorldState, evm *LaserEVM, runtimeCode string, cfg *z3.Config) *state.GlobalState {
	var txCtx *z3.Context
	txCtx = z3.NewContext(cfg)
	openState.TransactionIdInt++

	if openState.TransactionIdInt-1 > openState.TransactionCount {
		fmt.Println("OpenStateRelay", openState.TransactionIdInt, openState.TransactionCount, "notneed")
		return nil
	}

	if !openState.Constraints.IsPossible() {
		fmt.Println("OpenStateRelay", openState.TransactionIdInt, openState.TransactionCount, "notpossible")
		return nil
	}

	fmt.Println("OpenStateRelay", openState.TransactionIdInt, openState.TransactionCount, "ojbk")

	txId := strconv.Itoa(openState.TransactionIdInt)

	externalSender := txCtx.NewBitvec("sender_"+txId, 256)
	txCode := disassembler.NewDisasembly(runtimeCode)
	fmt.Println("In msgId:", txId)
	tx := &state.MessageCallTransaction{
		WorldState:    openState.Translate(txCtx),
		Code:          txCode,
		CalleeAccount: openState.Translate(txCtx).AccountsExistOrLoad(openState.ContractAddress.Translate(txCtx)),
		//CalleeAccount: openState.AccountsExistOrLoad(address),
		Caller:   externalSender,
		Calldata: state.NewSymbolicCalldata(txId, txCtx),
		//Calldata: NewConcreteCalldata(txId, calldataList, ctx),
		GasPrice:  10,
		GasLimit:  8000000,
		CallValue: 0,
		Origin:    externalSender,
		Basefee:   txCtx.NewBitvecVal(1000, 256),
		Ctx:       txCtx,
		Id:        txId,
	}

	globalState := tx.InitialGlobalState()
	ACTORS := transaction.NewActors(globalState.Z3ctx)
	constraint := tx.GetCaller().Eq(ACTORS.GetCreator()).Or(tx.GetCaller().Eq(ACTORS.GetAttacker()), tx.GetCaller().Eq(ACTORS.GetSomeGuy()))
	globalState.WorldState.Constraints.Add(constraint)
	globalState.WorldState.TransactionSequence = append(globalState.WorldState.TransactionSequence, tx)
	return globalState
}

func getNewMsgTx(openState *state.WorldState, evm *LaserEVM, runtimeCode string, inputStr string, ctx *z3.Context, address *z3.Bitvec, multiple bool, cfg *z3.Config) {
	var txCtx *z3.Context
	if cfg == nil && !multiple {
		txCtx = ctx
	} else {
		txCtx = z3.NewContext(cfg)
	}

	evm.TxCtxList = append(evm.TxCtxList, txCtx)
	txId := state.GetNextTransactionId()
	externalSender := txCtx.NewBitvec("sender_"+txId, 256)
	txCode := disassembler.NewDisasembly(runtimeCode)
	fmt.Println("In msgId:", txId)
	calldataList := make([]*z3.Bitvec, 0)
	for i := 0; i < len(inputStr); i = i + 2 {
		val, _ := strconv.ParseInt(inputStr[i:i+2], 16, 10)
		calldataList = append(calldataList, txCtx.NewBitvecVal(val, 8))
	}
	tx := &state.MessageCallTransaction{
		WorldState:    openState.Translate(txCtx),
		Code:          txCode,
		CalleeAccount: openState.Translate(txCtx).AccountsExistOrLoad(address.Translate(txCtx)),
		//CalleeAccount: openState.AccountsExistOrLoad(address),
		Caller:   externalSender,
		Calldata: state.NewSymbolicCalldata(txId, txCtx),
		//Calldata: NewConcreteCalldata(txId, calldataList, ctx),
		GasPrice:  10,
		GasLimit:  8000000,
		CallValue: 0,
		Origin:    externalSender,
		Basefee:   txCtx.NewBitvecVal(1000, 256),
		Ctx:       txCtx,
		Id:        txId,
	}
	setupGlobalStateForExecution(evm, tx)
}

func ExecuteMessageCallOnly(evm *LaserEVM, runtimeCode string, contractName string, inputStr string, ctx *z3.Context, multiple bool, cfg *z3.Config) {

	txId := state.GetNextTransactionId()
	externalSender := ctx.NewBitvec("sender_"+txId, 256)
	txCode := disassembler.NewDisasembly(runtimeCode)

	calldataList := make([]*z3.Bitvec, 0)
	for i := 0; i < len(inputStr); i = i + 2 {
		val, _ := strconv.ParseInt(inputStr[i:i+2], 16, 10)
		calldataList = append(calldataList, ctx.NewBitvecVal(val, 8))
	}

	tx := &state.MessageCallTransaction{
		WorldState: state.NewWordState(ctx),
		Code:       txCode,
		CalleeAccount: state.NewAccount(externalSender, ctx.NewArray("balances", 256, 256),
			false, txCode, contractName),
		Caller:   externalSender,
		Calldata: state.NewSymbolicCalldata(txId, ctx),
		//Calldata: NewConcreteCalldata(txId, calldataList, ctx),
		GasPrice:  10,
		GasLimit:  8000000,
		CallValue: 0,
		Origin:    externalSender,
		Basefee:   ctx.NewBitvecVal(1000, 256),
		Ctx:       ctx,
		Id:        txId,
	}

	setupGlobalStateForExecution(evm, tx)
	fmt.Println("########################################################################################")
	fmt.Println("MessageTx Execute!")
	if !multiple {
		evm.exec()
	} else {
		evm.multiExec(cfg)
	}
	fmt.Println("MessageTx End!")
	fmt.Println("########################################################################################")
}

func setupGlobalStateForExecution(evm *LaserEVM, tx state.BaseTransaction) {
	globalState := tx.InitialGlobalState()
	ACTORS := transaction.NewActors(globalState.Z3ctx)
	constraint := tx.GetCaller().Eq(ACTORS.GetCreator()).Or(tx.GetCaller().Eq(ACTORS.GetAttacker()), tx.GetCaller().Eq(ACTORS.GetSomeGuy()))
	globalState.WorldState.Constraints.Add(constraint)
	globalState.WorldState.TransactionSequence = append(globalState.WorldState.TransactionSequence, tx)
	evm.Manager.WorkList <- globalState
	evm.Manager.SignalCh <- Signal{
		Id:        -1,
		NewStates: 1,
	}
	fmt.Println("setupGlobalStateForExecution evm.Manager.WorkList", len(evm.Manager.WorkList), evm.Manager.WorkList)
}

//func (evm *LaserEVM) executeTransactionNormal(creationCode string, contractName string, ctx *z3.Context) {
//	inputStrArr := support.GetArgsInstance().TransactionSequences
//	for i := 0; i < evm.TransactionCount; i++ {
//		ExecuteMessageCallOnly(evm, creationCode, contractName, inputStrArr[i], ctx)
//		fmt.Println("normalExec:", i, "tx")
//	}
//}
//func (evm *LaserEVM) executeTransaction(creationCode string, contractName string, ctx *z3.Context, cfg *z3.Config) {
//	inputStrArr := support.GetArgsInstance().TransactionSequences
//	for i := 0; i < evm.TransactionCount; i++ {
//		ExecuteMessageCallOnly(evm, creationCode, contractName, inputStrArr[i], ctx, true, cfg)
//		for i := 0; i < evm.GofuncCount; i++ {
//			go evm.Run(i, cfg)
//		}
//
//		latestSignals := make(map[int]bool)
//	LOOP:
//		for {
//			/*
//				There are two situations for exiting.
//				1. All goroutines don't generate new globalStates.
//				2. There is no globalState in channel, so all goroutines will be blocked.
//			*/
//
//			//Situation 2
//			fmt.Println("noStatesFlag:", evm.NoStatesFlag)
//			if evm.NoStatesFlag {
//				fmt.Println("break in situation 2")
//				break LOOP
//			}
//
//			// Situation 1
//			signal := <-evm.SignalCh
//			latestSignals[signal.Id] = signal.Finished
//			fmt.Println(signal.Id, signal.Finished)
//			allFinished := true
//			for _, finished := range latestSignals {
//				//fmt.Println(i, finished)
//				if !finished {
//					allFinished = false
//				}
//			}
//			fmt.Println("situation 1", allFinished)
//			// TODO: 2022.10.12- Situation: goroutine 0-2 don't generate new states at the last execution.
//			// TODO: 2022.10.12- Now goroutine 3 don't generate new states, and then it get a state in channel to execute.
//			// TODO: 2022.10.12- But now it has broken in situation 1.
//			if allFinished && len(evm.Manager.WorkList) == 0 && evm.NoStatesSignal[signal.Id] {
//				fmt.Println("break in situation 1")
//				break LOOP
//			}
//		}
//		//fmt.Println("Finish", i, len(evm.Manager.WorkList))
//		// Reset the flag
//		//evm.NoStatesFlag = false
//		//for j := 0; j < evm.GofuncCount; j++ {
//		//	evm.NoStatesSignal[j] = true
//		//}
//	}
//}
