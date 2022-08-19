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
)

type moduleExecFunc func(globalState *state.GlobalState) []*analysis.Issue

type Signal struct {
	Id       int
	Finished bool
}

type LaserEVM struct {
	ExecutionTimeout int
	CreateTimeout    int
	TransactionCount int
	WorkList         chan *state.GlobalState
	// FinalState       chan *state.GlobalState
	InstrPreHook  *map[string][]moduleExecFunc
	InstrPostHook *map[string][]moduleExecFunc
	// Parallal
	// BeginCh     chan int
	// EndCh       chan int
	SignalCh chan Signal

	GofuncCount int
	// Analysis
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
		// FinalState:       make(chan *state.GlobalState),
		InstrPreHook:  &preHook,
		InstrPostHook: &postHook,

		// BeginCh:     make(chan int),
		// EndCh:       make(chan int),
		SignalCh:    make(chan Signal),
		GofuncCount: 4,
		Loader:      moduleLoader,
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
	//timestampDetectionModule := evm.Loader.Modules[2]
	//preHooksDM := timestampDetectionModule.(*modules.PredictableVariables).PreHooks
	//postHooksDM := timestampDetectionModule.(*modules.PredictableVariables).PostHooks
	//for _, op := range preHooksDM {
	//	//preHook[op] = []moduleExecFunc{reentrancyDetectionModule.Execute}
	//	preHook[op] = append(preHook[op], timestampDetectionModule.Execute)
	//}
	//for _, op := range postHooksDM {
	//	// postHook[op] = []moduleExecFunc{timestampDetectionModule.Execute}
	//	postHook[op] = append(postHook[op], timestampDetectionModule.Execute)
	//}
}

func (evm *LaserEVM) NormalSymExec(CreationCode string, contractName string) {
	fmt.Println("Symbolic Executing: ", CreationCode)
	fmt.Println("")
	//tx := state.NewContractCreationTransaction(CreationCode)
	tx := state.NewMessageCallTransaction(CreationCode, contractName)
	globalState := tx.InitialGlobalState()
	evm.WorkList <- globalState
	id := 0
LOOP:
	for {
		// When there is no newState in channel, exit the iteration
		if len(evm.WorkList) == 0 {
			break LOOP
		}
		globalState := <-evm.WorkList
		fmt.Println(id, globalState)
		fmt.Println(id, "constraints:", globalState.WorldState.Constraints)
		for _, constraint := range globalState.WorldState.Constraints.ConstraintList {
			fmt.Println(constraint)
		}
		newStates, opcode := evm.ExecuteState(globalState)
		// evm.ManageCFG(opcode, newStates)

		for _, newState := range newStates {
			//if opcode == "STOP" || opcode == "RETURN" {
			//	modules.CheckPotentialIssues(newState)
			//}
			evm.WorkList <- newState
		}
		fmt.Println(id, "done", globalState, opcode)
		fmt.Println("==============================================================================")
		if opcode == "STOP" || opcode == "RETURN" {
			modules.CheckPotentialIssues(globalState)
			for _, detector := range evm.Loader.Modules {
				issues := detector.GetIssues()
				for _, issue := range issues {
					fmt.Println("ContractName:", issue.Contract)
					fmt.Println("FunctionName:", issue.FunctionName)
					fmt.Println("Title:", issue.Title)
					fmt.Println("SWCID:", issue.SWCID)
					fmt.Println("Address:", issue.Address)
					fmt.Println("Severity:", issue.Severity)
				}
			}
		}
		id++
	}
}

func (evm *LaserEVM) SymExec(CreationCode string, contractName string) {
	fmt.Println("Symbolic Executing: ", CreationCode)
	// TOOD: actually creation code is not for base tx, but for creation tx, just for test here
	tx := state.NewMessageCallTransaction(CreationCode, contractName)
	globalState := tx.InitialGlobalState()
	evm.WorkList <- globalState
	for i := 0; i < evm.GofuncCount; i++ {
		go evm.Run(i)
	}
	// TODO: not good here
	// beginCount := 0
	// endCount := 0
	latestSignals := make(map[int]bool)
LOOP:
	for {
		// select {
		// case <-evm.BeginCh:
		// 	beginCount++
		// case <-evm.EndCh:
		// 	endCount++
		// 	if endCount == beginCount {
		// 		fmt.Println("finish", beginCount, endCount)
		// 		break LOOP
		// 	} else {
		// 		fmt.Println("not finish", beginCount, endCount)
		// 	}
		// }
		//}
		signal := <-evm.SignalCh
		latestSignals[signal.Id] = signal.Finished
		allFinished := true
		for _, finished := range latestSignals {
			if !finished {
				allFinished = false
			}
		}
		if allFinished {
			break LOOP
		}
	}
	fmt.Println("Finish", len(evm.WorkList))
}

func (evm *LaserEVM) ExecuteState(globalState *state.GlobalState) ([]*state.GlobalState, string) {
	instrs := globalState.Environment.Code.InstructionList
	opcode := instrs[globalState.Mstate.Pc].OpCode.Name

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

func readWithSelect(ch chan *state.GlobalState) (*state.GlobalState, error) {
	select {
	case globalState := <-ch:
		return globalState, nil
	default:
		return nil, errors.New("evm.WorkList is empty")
	}
}

func (evm *LaserEVM) Run(id int) {
	fmt.Println("Run")
	for {
		// globalState := <-evm.WorkList
		globalState, _ := readWithSelect(evm.WorkList)
		//fmt.Println("error:", err, id)

		if globalState != nil {
			//evm.BeginCh <- id
			fmt.Println(id, globalState, &globalState)
			newStates, opcode := evm.ExecuteState(globalState)
			//evm.ManageCFG(opcode, newStates)
			fmt.Println("test")
			for _, newState := range newStates {
				fmt.Println("2")
				evm.WorkList <- newState
				fmt.Println("append:", len(evm.WorkList))
			}
			fmt.Println(id, "done", globalState, opcode, &globalState)
			evm.SignalCh <- Signal{
				Id:       id,
				Finished: (len(newStates) == 0),
			}
			fmt.Println("signal", id, len(newStates) == 0)
			fmt.Println("===========================================================================")
			if opcode == "STOP" || opcode == "RETURN" {
				modules.CheckPotentialIssues(globalState)
				for _, detector := range evm.Loader.Modules {
					issues := detector.GetIssues()
					for _, issue := range issues {
						fmt.Println("ContractName:", issue.Contract)
						fmt.Println("FunctionName:", issue.FunctionName)
						fmt.Println("Title:", issue.Title)
						fmt.Println("SWCID:", issue.SWCID)
						fmt.Println("Address:", issue.Address)
						fmt.Println("Severity:", issue.Severity)
					}
				}
			}
		} else {
			evm.SignalCh <- Signal{
				Id:       id,
				Finished: true,
			}
		}

		// TODO not good for sleep
		// time.Sleep(100 * time.Millisecond)
		// evm.EndCh <- id
	}
}
