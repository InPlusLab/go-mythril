package ethereum

import (
	"errors"
	"fmt"
	"go-mythril/analysis"
	"go-mythril/analysis/module"
	"go-mythril/analysis/module/modules"
	"go-mythril/disassembler"
	"go-mythril/laser/ethereum/state"
	"go-mythril/laser/ethereum/transaction"
	"go-mythril/laser/smt/z3"
	"go-mythril/support"
	"go-mythril/utils"
	"strconv"
	"sync"
)

type moduleExecFunc func(globalState *state.GlobalState) []*analysis.Issue

var l sync.Mutex

type Signal struct {
	Id       int
	Finished bool
}

type LaserEVM struct {
	ExecutionTimeout int
	CreateTimeout    int
	TransactionCount int
	WorkList         chan *state.GlobalState
	OpenStates       []*state.WorldState
	// FinalState       chan *state.GlobalState
	InstrPreHook  *map[string][]moduleExecFunc
	InstrPostHook *map[string][]moduleExecFunc
	/* Parallal */
	SignalCh       chan Signal
	NoStatesFlag   bool
	NoStatesSignal []bool
	NoStatesCh     chan []bool
	GofuncCount    int
	/* Analysis */
	Loader *module.ModuleLoader
}

func NewLaserEVM(ExecutionTimeout int, CreateTimeout int, TransactionCount int, moduleLoader *module.ModuleLoader) *LaserEVM {

	preHook := make(map[string][]moduleExecFunc)
	postHook := make(map[string][]moduleExecFunc)
	opcodes := *support.NewOpcodes()
	for _, v := range opcodes {
		preHook[v.Name] = make([]moduleExecFunc, 0)
		postHook[v.Name] = make([]moduleExecFunc, 0)
	}

	evm := LaserEVM{
		ExecutionTimeout: ExecutionTimeout,
		CreateTimeout:    CreateTimeout,
		TransactionCount: TransactionCount,
		WorkList:         make(chan *state.GlobalState, 1000),
		OpenStates:       make([]*state.WorldState, 0),
		// FinalState:       make(chan *state.GlobalState),
		InstrPreHook:  &preHook,
		InstrPostHook: &postHook,

		SignalCh:       make(chan Signal),
		NoStatesFlag:   false,
		NoStatesSignal: make([]bool, 4, 4),
		NoStatesCh:     make(chan []bool),
		GofuncCount:    4,
		Loader:         moduleLoader,
	}
	evm.registerInstrHooks()
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

func (evm *LaserEVM) exec() {
	id := 0
LOOP:
	for {
		// When there is no newState in channel, exit the iteration
		fmt.Println("evm workList:", len(evm.WorkList))
		if len(evm.WorkList) == 0 {
			break LOOP
		}
		globalState := <-evm.WorkList

		newStates, opcode := evm.ExecuteState(globalState)
		fmt.Println(id, globalState, opcode)

		for _, newState := range newStates {
			evm.WorkList <- newState
		}

		fmt.Println(id, "done", globalState, opcode)
		fmt.Println("==============================================================================")

		if len(newStates) == 0 {
			modules.CheckPotentialIssues(globalState)
			for _, detector := range evm.Loader.Modules {
				issues := detector.GetIssues()
				fmt.Println("number of issues:", len(issues))
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
		}
		id++
	}
	fmt.Println("evm.Exec: One tx end!")
}

func (evm *LaserEVM) multiExec(cfg *z3.Config) {
	for i := 0; i < evm.GofuncCount; i++ {
		go evm.Run(i, cfg)
	}

	latestSignals := make(map[int]bool)
LOOP:
	for {
		/*
			There are two situations for exiting.
			1. All goroutines don't generate new globalStates.
			2. There is no globalState in channel, so all goroutines will be blocked.
		*/

		//Situation 2
		fmt.Println("noStatesFlag:", evm.NoStatesFlag)
		if evm.NoStatesFlag {
			fmt.Println("break in situation 2")
			break LOOP
		}

		// Situation 1
		signal := <-evm.SignalCh
		latestSignals[signal.Id] = signal.Finished
		fmt.Println(signal.Id, signal.Finished)
		allFinished := true
		for _, finished := range latestSignals {
			//fmt.Println(i, finished)
			if !finished {
				allFinished = false
			}
		}
		fmt.Println("situation 1", allFinished)
		// TODO: 2022.10.12- Situation: goroutine 0-2 don't generate new states at the last execution.
		// TODO: 2022.10.12- Now goroutine 3 don't generate new states, and then it get a state in channel to execute.
		// TODO: 2022.10.12- But now it has broken in situation 1.
		if allFinished && len(evm.WorkList) == 0 && evm.NoStatesSignal[signal.Id] {
			fmt.Println("break in situation 1")
			break LOOP
		}
	}
}

func (evm *LaserEVM) SingleSymExec(creationCode string, runtimeCode string, contractName string, ctx *z3.Context) {
	fmt.Println("Single Goroutine Symbolic Executing")
	fmt.Println("")
	// CreationTx
	newAccount := ExecuteContractCreation(evm, creationCode, contractName, ctx)
	// MessageTx
	inputStrArr := support.GetArgsInstance().TransactionSequences
	for i := 0; i < evm.TransactionCount; i++ {
		ExecuteMessageCall(evm, runtimeCode, inputStrArr[i], ctx, newAccount.Address, false, nil)
		fmt.Println("msgTx", i, ":done")
	}
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
	fmt.Println("opcode:", opcode)
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

func readWithSelect(evm *LaserEVM) (*state.GlobalState, error) {
	select {
	case globalState := <-evm.WorkList:
		return globalState, nil
	default:
		return nil, errors.New("evm.WorkList is empty")
	}
}

func (evm *LaserEVM) Run(id int, cfg *z3.Config) {
	fmt.Println("Run")
	ctx := z3.NewContext(cfg)
	for {
		globalState, _ := readWithSelect(evm)
		//evm.SignalCh <- Signal{
		//	Id:       id,
		//	Finished: false,
		//}
		//l.Lock()
		evm.NoStatesSignal[id] = globalState == nil
		//l.Unlock()

		if globalState != nil {
			//if ctx != globalState.Z3ctx {
			//	globalState.Translate(ctx)
			//}
			globalState.Translate(ctx)

			newStates, opcode := evm.ExecuteState(globalState)
			fmt.Println("last", len(newStates) == 0)

			fmt.Println(id, globalState, opcode)
			//fmt.Println(id, globalState, opcode)
			//evm.ManageCFG(opcode, newStates)
			for _, newState := range newStates {
				evm.WorkList <- newState
			}
			evm.SignalCh <- Signal{
				Id:       id,
				Finished: len(newStates) == 0,
			}

			fmt.Println(id, "done", globalState, opcode)

			//fmt.Println("produceNoStates:", id, len(newStates) == 0)
			//fmt.Println(id, opcode)
			//for i := 0; i < evm.GofuncCount; i++ {
			//	fmt.Println(evm.NoStatesSignal[i])
			//}
			fmt.Println("===========================================================================")
			//if opcode == "STOP" || opcode == "RETURN" {
			if len(newStates) == 0 {
				modules.CheckPotentialIssues(globalState)
				for _, detector := range evm.Loader.Modules {
					issues := detector.GetIssues()
					fmt.Println("number of issues:", len(issues))
					for _, issue := range issues {
						fmt.Println("+++++++++++++++++++++++++++++++++++", id)
						fmt.Println("ContractName:", issue.Contract)
						fmt.Println("FunctionName:", issue.FunctionName)
						fmt.Println("Title:", issue.Title)
						fmt.Println("SWCID:", issue.SWCID)
						fmt.Println("Address:", issue.Address)
						fmt.Println("Severity:", issue.Severity)
					}
				}

				fmt.Println("+++++++++++++++++++++++++++++++++++")

				// when the other goroutines have no globalStates to solve.
				l.Lock()
				flag := true
				for i, v := range evm.NoStatesSignal {
					if i != id {
						flag = flag && v
					}
				}
				l.Unlock()

				if flag {
					fmt.Println("all goroutines have no globalStates")
					evm.NoStatesFlag = true
					// Here, we push a value in channel to trigger the LOOP iteration in symExec.
					evm.SignalCh <- Signal{
						Id:       id,
						Finished: false,
					}
				}
			}
		}

		/* peilin
		globalState := <-evm.WorkList
		//evm.BeginCh <- id
		fmt.Println(id, globalState)
		newStates, opcode := evm.ExecuteState(globalState)
		evm.ManageCFG(opcode, newStates)

		for _, newState := range newStates {
			evm.WorkList <- newState
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
}

func ExecuteContractCreation(evm *LaserEVM, creationCode string, contractName string, ctx *z3.Context) *state.Account {

	worldState := state.NewWordState(ctx)
	evm.OpenStates = append(evm.OpenStates, worldState)
	ACTORS := transaction.NewActors(ctx)
	txId := state.GetNextTransactionId()

	tx := &state.ContractCreationTransaction{
		WorldState:    worldState,
		Code:          disassembler.NewDisasembly(creationCode),
		CalleeAccount: worldState.CreateAccount(0, true, ACTORS.GetCreator(), nil, nil, contractName),
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

	newAccount := tx.CalleeAccount

	fmt.Println("########################################################################################")
	fmt.Println("CreationTx Execute!")
	evm.exec()
	fmt.Println("CreationTx Done!")
	fmt.Println("########################################################################################")

	return newAccount
}

func ExecuteMessageCall(evm *LaserEVM, runtimeCode string, inputStr string, ctx *z3.Context, address *z3.Bitvec, multiple bool, cfg *z3.Config) {
	txId := state.GetNextTransactionId()
	externalSender := ctx.NewBitvec("sender_"+txId, 256)
	txCode := disassembler.NewDisasembly(runtimeCode)

	calldataList := make([]*z3.Bitvec, 0)
	for i := 0; i < len(inputStr); i = i + 2 {
		val, _ := strconv.ParseInt(inputStr[i:i+2], 16, 10)
		calldataList = append(calldataList, ctx.NewBitvecVal(val, 8))
	}

	var openState *state.WorldState
	if len(evm.OpenStates) > 0 {
		openState = evm.OpenStates[0]
	} else {
		panic("empty openStates for msgTx!")
	}

	tx := &state.MessageCallTransaction{
		WorldState:    openState,
		Code:          txCode,
		CalleeAccount: openState.AccountsExistOrLoad(address),
		Caller:        externalSender,
		Calldata:      state.NewSymbolicCalldata(txId, ctx),
		//Calldata: NewConcreteCalldata(txId, calldataList, ctx),
		GasPrice:  10,
		GasLimit:  8000000,
		CallValue: 0,
		Origin:    externalSender,
		Basefee:   ctx.NewBitvecVal(1000, 256),
		Ctx:       ctx,
		Id:        txId,
	}
	//tx.WorldState.TransactionSequence = append(tx.WorldState.TransactionSequence, tx)
	setupGlobalStateForExecution(evm, tx)
	fmt.Println("########################################################################################")
	fmt.Println("MessageTx Execute!")
	if !multiple {
		evm.exec()
	} else {
		evm.multiExec(cfg)
	}
	fmt.Println("MessageTx Done!")
	fmt.Println("########################################################################################")
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
	fmt.Println("MessageTx Done!")
	fmt.Println("########################################################################################")
}

func setupGlobalStateForExecution(evm *LaserEVM, tx state.BaseTransaction) {
	globalState := tx.InitialGlobalState()
	ACTORS := transaction.NewActors(globalState.Z3ctx)
	constraint := tx.GetCaller().Eq(ACTORS.GetCreator()).Or(tx.GetCaller().Eq(ACTORS.GetAttacker()), tx.GetCaller().Eq(ACTORS.GetSomeGuy()))
	globalState.WorldState.Constraints.Add(constraint)
	globalState.WorldState.TransactionSequence = append(globalState.WorldState.TransactionSequence, tx)
	evm.WorkList <- globalState
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
//			if allFinished && len(evm.WorkList) == 0 && evm.NoStatesSignal[signal.Id] {
//				fmt.Println("break in situation 1")
//				break LOOP
//			}
//		}
//		//fmt.Println("Finish", i, len(evm.WorkList))
//		// Reset the flag
//		//evm.NoStatesFlag = false
//		//for j := 0; j < evm.GofuncCount; j++ {
//		//	evm.NoStatesSignal[j] = true
//		//}
//	}
//}
