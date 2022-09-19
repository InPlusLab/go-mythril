package ethereum

import (
	"errors"
	"fmt"
	"go-mythril/analysis"
	"go-mythril/analysis/module"
	"go-mythril/analysis/module/modules"
	"go-mythril/laser/ethereum/state"
	"go-mythril/laser/smt/z3"
	"go-mythril/support"
	"go-mythril/utils"
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
	//timeStampDetectionModule := evm.Loader.Modules[2]
	//preHooksDM := timeStampDetectionModule.(*modules.PredictableVariables).PreHooks
	//for _, op := range preHooksDM {
	//	preHook[op] = []moduleExecFunc{timeStampDetectionModule.Execute}
	//}
	//postHooksDM := timeStampDetectionModule.(*modules.PredictableVariables).PostHooks
	//for _, op := range postHooksDM {
	//	postHook[op] = []moduleExecFunc{timeStampDetectionModule.Execute}
	//}
}

func (evm *LaserEVM) NormalSymExec(creationCode string, contractName string, ctx *z3.Context) {
	fmt.Println("Symbolic Executing: ", creationCode)
	fmt.Println("")
	evm.executeTransactionNormal(creationCode, contractName, ctx)
}

func (evm *LaserEVM) executeTransactionNormal(creationCode string, contractName string, ctx *z3.Context) {
	inputStrArr := support.GetArgsInstance().TransactionSequences
	for i := 0; i < evm.TransactionCount; i++ {
		tx := state.NewMessageCallTransaction(creationCode, contractName, inputStrArr[i], ctx)
		globalState := tx.InitialGlobalState()
		evm.WorkList <- globalState
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
			// evm.ManageCFG(opcode, newStates)

			for _, newState := range newStates {
				evm.WorkList <- newState
			}

			fmt.Println(id, "done", globalState, opcode)
			fmt.Println("==============================================================================")
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
			}
			id++
		}
		fmt.Println("normalExec:", i, "tx")
	}
}

func (evm *LaserEVM) SymExec(creationCode string, contractName string, ctx *z3.Context, cfg *z3.Config) {
	fmt.Println("Symbolic Executing: ", creationCode)
	fmt.Println("")
	evm.executeTransaction(creationCode, contractName, ctx, cfg)
}

func (evm *LaserEVM) executeTransaction(creationCode string, contractName string, ctx *z3.Context, cfg *z3.Config) {
	inputStrArr := support.GetArgsInstance().TransactionSequences
	for i := 0; i < evm.TransactionCount; i++ {
		tx := state.NewMessageCallTransaction(creationCode, contractName, inputStrArr[i], ctx)
		globalState := tx.InitialGlobalState()
		evm.WorkList <- globalState

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
			allFinished := true
			for _, finished := range latestSignals {
				//fmt.Println(i, finished)
				if !finished {
					allFinished = false
				}
			}
			if allFinished && len(evm.WorkList) == 0 {
				fmt.Println("break in situation 1")
				break LOOP
			}
		}
		//fmt.Println("Finish", i, len(evm.WorkList))
		// Reset the flag
		//evm.NoStatesFlag = false
		//for j := 0; j < evm.GofuncCount; j++ {
		//	evm.NoStatesSignal[j] = true
		//}
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

func changeStateContext(globalState *state.GlobalState, ctx *z3.Context) {
	fmt.Println("before changeStateContext")
	globalState.Z3ctx = ctx
	// machineState stack & memory
	newMachineState := globalState.Mstate.Translate(ctx)
	globalState.Mstate = newMachineState
	// worldState constraints
	newWorldState := globalState.WorldState.Translate(ctx)
	globalState.WorldState = newWorldState
	// env
	newEnv := globalState.Environment.Translate(ctx)
	globalState.Environment = newEnv
	// lastReturnData
	//fmt.Println("changeStateContext done")
}

func (evm *LaserEVM) Run(id int, cfg *z3.Config) {
	fmt.Println("Run")
	ctx := z3.NewContext(cfg)
	for {
		globalState, _ := readWithSelect(evm)

		//l.Lock()
		evm.NoStatesSignal[id] = globalState == nil
		//l.Unlock()

		if globalState != nil {
			if ctx != globalState.Z3ctx {
				changeStateContext(globalState, ctx)
			}

			newStates, opcode := evm.ExecuteState(globalState)
			fmt.Println(id, globalState, opcode)
			//fmt.Println(id, globalState, opcode)
			//evm.ManageCFG(opcode, newStates)
			for _, newState := range newStates {
				evm.WorkList <- newState
			}
			fmt.Println(id, "done", globalState, opcode)
			evm.SignalCh <- Signal{
				Id:       id,
				Finished: len(newStates) == 0,
			}
			//fmt.Println("produceNoStates:", id, len(newStates) == 0)
			//fmt.Println(id, opcode)
			//for i := 0; i < evm.GofuncCount; i++ {
			//	fmt.Println(evm.NoStatesSignal[i])
			//}
			fmt.Println("===========================================================================")
			if opcode == "STOP" || opcode == "RETURN" {

				modules.CheckPotentialIssues(globalState)
				for _, detector := range evm.Loader.Modules {
					issues := detector.GetIssues()
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
