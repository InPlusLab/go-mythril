package ethereum

import (
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
	ForkFlag bool
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
	LastOpCodeList    []string
	LastAfterExecList []int
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
	FinalStates int
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
	fmt.Println("Worklist", m.WorkList, "WorklistLen", len(m.WorkList), "TotalStates", m.TotalStates, "FinishedStates", m.FinishedStates, "Running", m.TotalStates-m.FinishedStates-len(m.WorkList))
}

func (m *Manager) Pop() *state.GlobalState {
	return <-m.WorkList
}

func (m *Manager) SignalLoop() {
	fmt.Println("SignalLoop Start")
	start := time.Now()
	for {
		// fmt.Println("wait signal")
		select {
		case signal := <-m.SignalCh:
			if signal.Id != -1 {
				m.FinishedStates += 1
			}
			m.TotalStates += signal.NewStates

			m.Duration += signal.Time


			//if signal.ForkFlag {
			//	m.FinalStates += int(signal.Time)
			//	//fmt.Println(signal.Id, "relayStates++", m.FinalStates)
			//}
			if !signal.ForkFlag && signal.Id != -1 {
				m.FinalStates += 1
			}

			//if signal.Id == 0 {
			duration := time.Since(start)
			fmt.Println("miaomi:", duration.Seconds(), m.FinalStates)
			//}

			if signal.NewStates == 0 && m.TotalStates == m.FinishedStates {
				goto BREAK
			}
		case id := <-m.ReqCh:
			runningWorkers := m.TotalStates - m.FinishedStates
			canSkip := (runningWorkers+1 < m.GofuncCount)

			// no skip
			if MaxSkipTimes == 0 {
				canSkip = false
			}

			m.RespChs[id] <- canSkip
		}
	}
BREAK:
	fmt.Println("SignalLoop Stop")
	fmt.Println("totalStates:",m.TotalStates, "relayStates:", m.FinalStates)
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

		LastOpCodeList:    make([]string, goFuncCount, goFuncCount),
		LastAfterExecList: make([]int, goFuncCount),
		CtxList:     utils.NewSyncSliceWithArr(ctxList),
		NewCtxList:  make([]*z3.Context, 0),
		TxCtxList:   make([]*z3.Context, 0),
		GofuncCount: goFuncCount,
		Loader:      moduleLoader,
		Manager:     NewManager(goFuncCount),
	}
	//evm.registerInstrHooks()
	for i := 0; i < goFuncCount; i++ {
		go evm.Run(i, cfg)
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
	ctxList := make([]interface{}, evm.GofuncCount, evm.GofuncCount)
	evm.CtxList = utils.NewSyncSliceWithArr(ctxList)
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
		// fmt.Println(id, globalState, opcode)

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

		// fmt.Println(id, "done", globalState, opcode)
		// fmt.Println("==============================================================================")

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

func (evm *LaserEVM) multiExec(cfg *z3.Config) {
	evm.Manager.SignalLoop()
	return
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

var MaxRLimitCount int
var MaxSkipTimes int

func SetMaxRLimitCount(value int) {
	MaxRLimitCount = value
	fmt.Println("MaxRLimitCount:", MaxRLimitCount)
}
func SetMaxSkipTimes(value int) {
	MaxSkipTimes = value
}

func (evm *LaserEVM) inCtxList(globalState *state.GlobalState) bool {

	txFlag := false
	for _, tx := range globalState.TxStack {
		if globalState.Z3ctx.GetRaw() == tx.GetCtx().GetRaw() {
			txFlag = true
			break
		}
	}

	integerFlag := false
	integerArr := []string{"IntegerArithmetics"}
	integerModule := evm.Loader.GetDetectionModules(integerArr)
	for _, item := range integerModule[0].(*modules.IntegerArithmetics).CtxList.Elements() {
		if globalState.Z3ctx.GetRaw() == item.(*z3.Context).GetRaw() {
			integerFlag = true
			break
		}
	}

	return txFlag || integerFlag
}

func (evm *LaserEVM) Run(id int, cfg *z3.Config) {
	for {
		globalState := evm.Manager.Pop()

		if globalState.ForkId != "?" {
			evm.RunWhenFork(id, globalState.ForkId, globalState.RootState, cfg)
		} else {
		EXECFor:
			// TODO canSkip here?
			evm.Manager.ReqCh <- id
			canSkip := <-evm.Manager.RespChs[id]

			if globalState.NeedIsPossible {
				if !canSkip || globalState.SkipTimes >= MaxSkipTimes {
					sat, rlimit := globalState.WorldState.Constraints.IsPossibleRlimit()
					globalState.SkipTimes = 0
					if sat {
						globalState.RLimitCount += rlimit
					} else {
						evm.Manager.SignalCh <- Signal{
							Id:        id,
							NewStates: 0,
						}
						//fmt.Println(id, "throw state1", globalState.ForkId, globalState.GetCurrentInstruction().OpCode.Name)
						continue
					}
				} else {
					// skip success
					globalState.SkipTimes += 1
				}
			}

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

			if curInstr.OpCode.Name == "JUMPDEST" {
				if globalState.Mstate.LastPc == 0 {
					count := evm.LoopsStrategy.GetLoopCount(annotation.GetTrace())
					if "*state.ContractCreationTransaction" == reflect.TypeOf(globalState.CurrentTransaction()).String() && count < 8 {
						goto EXEC
					} else if count > evm.LoopsStrategy.Bound {
						fmt.Println("Loop bound reached, skipping state", count, evm.LoopsStrategy.Bound)
						modules.CheckPotentialIssues(globalState)
						fmt.Println("LenOpenStates:", evm.OpenStatesSync.Length())
						evm.LastOpCodeList[id] = "JUMPDEST"
						evm.LastAfterExecList[id] = 0
						//evm.LastAfterExecList.SetItem(id, 0)
						// decouple
						//evm.CtxList[id] = nil
						evm.CtxList.SetItem(id, nil)
						evm.Manager.SignalCh <- Signal{
							Id:        id,
							NewStates: 0,
						}
						//fmt.Println(id, "throw state2", globalState.ForkId, globalState.GetCurrentInstruction().OpCode.Name)
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

			//fmt.Println(id,  opcode)

			// setNeedIsPossible
			if len(newStates) == 2 {
				if globalState.RLimitCount > MaxRLimitCount {
					//fmt.Println("run out of maxRlimit")
					newStates = make([]*state.GlobalState, 0)
				}
				for _, s := range newStates {
					s.NeedIsPossible = false
				}
			} else {
				for _, s := range newStates {
					s.NeedIsPossible = false
				}
			}
			// translate when jumpi
			if len(newStates) == 2 {
				ctx := z3.NewContext(cfg)
				newStates[1].Translate(ctx)
				ctx0 := z3.NewContext(cfg)
				newStates[0].Translate(ctx0)

				//if !evm.inCtxList(globalState) {
				//	globalState.Z3ctx.Close()
				//}
			}
			//fmt.Println(id, "done", opcode)
			//fmt.Println("===========================================================================")

			endOpcodeList := []string{"STOP", "RETURN", "REVERT", "SELFDESTRUCT"}
			if utils.In(opcode, endOpcodeList) {
				fmt.Println("#LenConstraintsEnd:", globalState.WorldState.Constraints.Length())
				evm.FinalState = globalState
				if "*state.MessageCallTransaction" == reflect.TypeOf(globalState.CurrentTransaction()).String() {
					if opcode != "REVERT" {
						relayGlobalState := OpenStateRelay(globalState.WorldState, evm, evm.RuntimeCode, cfg)
						// put relayState in channel
						if relayGlobalState != nil {
							ctx := z3.NewContext(cfg)
							relayGlobalState.Translate(ctx)
							newStates = append(newStates, relayGlobalState)
						}
						modules.CheckPotentialIssues(globalState)
						// for OpenStateRelay!!!
						if len(newStates) == 1 {
							evm.Manager.WorkList <- newStates[0]
						}
						evm.Manager.SignalCh <- Signal{
							Id:        id,
							NewStates: len(newStates),
						}
						continue
					}
				}
				// decouple
				//evm.CtxList[id] = nil
				evm.CtxList.SetItem(id, nil)
			}

			evm.LastOpCodeList[id] = opcode
			evm.LastAfterExecList[id] = len(newStates)

			if len(newStates) == 2 {
				evm.Manager.WorkList <- newStates[1]
				globalState = newStates[0]
			}
			if len(newStates) == 1 {
				globalState = newStates[0]
			}
			evm.Manager.SignalCh <- Signal{
				Id:        id,
				NewStates: len(newStates),
			}
			if len(newStates) != 0 {
				goto EXECFor
			}
		}
	}

	fmt.Println("Run Stop", id)
}

func (evm *LaserEVM) RunWhenFork(id int, forkId string, rootState *state.GlobalState, cfg *z3.Config) {
	jumpiCount := 0
	relayCount := 0
	// when first come in, globalState.RootState == nil, and it will pass through the path!
	globalState := rootState.Copy()
	gRootState := rootState.CopyForRoot()
	globalState.RootState = gRootState

EXECFor:
	// TODO canSkip here?
	evm.Manager.ReqCh <- id
	canSkip := <-evm.Manager.RespChs[id]

	//newCtx := z3.NewContext(cfg)
	//globalState.Translate(newCtx)

	if globalState.NeedIsPossible {
		if !canSkip || globalState.SkipTimes >= MaxSkipTimes {
			sat, rlimit := globalState.WorldState.Constraints.IsPossibleRlimit()
			globalState.SkipTimes = 0
			if sat {
				globalState.RLimitCount += rlimit
			} else {
				evm.Manager.SignalCh <- Signal{
					Id:        id,
					NewStates: 0,
				}
				//fmt.Println(id, "throw state1", globalState.ForkId, globalState.GetCurrentInstruction().OpCode.Name, "inRunWhenFork")
				return
			}
		} else {
			// skip success
			globalState.SkipTimes += 1
		}
	}

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

	if curInstr.OpCode.Name == "JUMPDEST" {
		if globalState.Mstate.LastPc == 0 {
			count := evm.LoopsStrategy.GetLoopCount(annotation.GetTrace())
			if "*state.ContractCreationTransaction" == reflect.TypeOf(globalState.CurrentTransaction()).String() && count < 8 {
				goto EXEC
			} else if count > evm.LoopsStrategy.Bound {
				fmt.Println("Loop bound reached, skipping state", count, evm.LoopsStrategy.Bound)
				modules.CheckPotentialIssues(globalState)
				fmt.Println("LenOpenStates:", evm.OpenStatesSync.Length())
				evm.LastOpCodeList[id] = "JUMPDEST"
				evm.LastAfterExecList[id] = 0
				//evm.LastAfterExecList.SetItem(id, 0)
				// decouple
				//evm.CtxList[id] = nil
				evm.CtxList.SetItem(id, nil)
				evm.Manager.SignalCh <- Signal{
					Id:        id,
					NewStates: 0,
				}
				//fmt.Println(id, "throw state2", globalState.ForkId, globalState.GetCurrentInstruction().OpCode.Name, "inRunWhenFork")
				return
			}
		} else {
			// after jumpi
			globalState.Mstate.LastPc = 0
			goto EXEC
		}
	}

EXEC:
	newStates, opcode := evm.ExecuteState(globalState)

	//fmt.Println(id, opcode)
	//fmt.Println(id, "done", opcode)
	//fmt.Println("===========================================================================")

	if jumpiCount < len(forkId) {
		relayCount += 1
	}

	if len(newStates) == 2 {
		if globalState.RLimitCount > MaxRLimitCount {
			//fmt.Println("run out of maxRlimit inRunWhenFork")
			newStates = make([]*state.GlobalState, 0)
			evm.Manager.SignalCh <- Signal{
				Id:        id,
				NewStates: len(newStates),
			}
			return
		}
		for _, s := range newStates {
			s.NeedIsPossible = false
		}
		// set forkId
		jumpiCount++
		/* choose one path */
		if jumpiCount > len(forkId){
			// when a new "jumpi" op comes
			ctx := z3.NewContext(cfg)
			newStates[1].Translate(ctx)
			ctx0 := z3.NewContext(cfg)
			newStates[0].Translate(ctx0)

			//if !evm.inCtxList(globalState) {
			//	globalState.Z3ctx.Close()
			//}

			evm.Manager.WorkList <- newStates[1]
			globalState = newStates[0]
			evm.Manager.SignalCh <- Signal{
				Id:        id,
				NewStates: len(newStates),
				Time:      0,
				// not relay part
				ForkFlag: false,
			}
			goto EXECFor
		}else{
			// select one side directly
			pathMark := string(forkId[jumpiCount-1])
			if pathMark == "0" {
				globalState = newStates[0]
			}else{
				globalState = newStates[1]
			}
			// SetNeedIsPossible
			globalState.NeedIsPossible = false

			evm.Manager.SignalCh <- Signal{
				Id:        id,
				NewStates: 1,
				Time:      0,
				ForkFlag: true,
			}
			goto EXECFor
		}
	}else {
		for _, s := range newStates {
			s.NeedIsPossible = false
		}
		relayFlag := jumpiCount < len(forkId)
		if len(newStates) == 1 {
			globalState = newStates[0]

			evm.Manager.SignalCh <- Signal{
				Id:        id,
				NewStates: len(newStates),
				Time:      0,
				ForkFlag: relayFlag,
			}

			goto EXECFor
		}else {
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
						// for OpenStateRelay!!!
						evm.Manager.SignalCh <- Signal{
							Id:        id,
							NewStates: len(newStates),
							Time:      int64(relayCount),
							ForkFlag: false,
						}
						if len(newStates) == 1 {
							evm.Manager.WorkList <- newStates[0]
						}
						return
					}
				}

				// decouple
				//evm.CtxList[id] = nil
				evm.CtxList.SetItem(id, nil)
			}

			evm.Manager.SignalCh <- Signal{
				Id:        id,
				NewStates: len(newStates),
				Time:      int64(relayCount),
				ForkFlag: false,
			}

			evm.LastOpCodeList[id] = opcode
			evm.LastAfterExecList[id] = len(newStates)

		}
	}
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
		//evm.exec()
		evm.Manager.SignalLoop()
	} else {
		evm.multiExec(cfg)
	}
	fmt.Println("CreationTx End!")
	fmt.Println("########################################################################################")

	// Tips: the final Storage pass
	newAccount := evm.FinalState.Environment.ActiveAccount
	evm.OpenStates = []*state.WorldState{evm.FinalState.WorldState}

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

	// setRootState
	rootState := globalState.Copy()
	globalState.RootState = rootState

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

	//if !openState.Constraints.IsPossible() {
	//	fmt.Println("OpenStateRelay", openState.TransactionIdInt, openState.TransactionCount, "notpossible")
	//	return nil
	//}

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
	globalState.NeedIsPossible = false

	// setRoot
	rootState := globalState.Copy()
	globalState.RootState = rootState

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

	// setRoot
	rootState := globalState.Copy()
	globalState.RootState = rootState

	evm.Manager.WorkList <- globalState
	evm.Manager.SignalCh <- Signal{
		Id:        -1,
		NewStates: 1,
	}
	fmt.Println("setupGlobalStateForExecution evm.Manager.WorkList", len(evm.Manager.WorkList), evm.Manager.WorkList)
}