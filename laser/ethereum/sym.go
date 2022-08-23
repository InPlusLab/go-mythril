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
	/* Parallal */
	SignalCh       chan Signal
	NoStatesFlag   bool
	NoStatesSignal []bool
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
		// FinalState:       make(chan *state.GlobalState),
		InstrPreHook:  &preHook,
		InstrPostHook: &postHook,

		SignalCh:       make(chan Signal),
		NoStatesFlag:   false,
		NoStatesSignal: make([]bool, 4),
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
	//integerDetectionModule := evm.Loader.Modules[0]
	//preHooksDM := integerDetectionModule.(*modules.IntegerArithmetics).PreHooks
	//for _, op := range preHooksDM {
	//	preHook[op] = []moduleExecFunc{integerDetectionModule.Execute}
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
		fmt.Println("loop in currentObj")
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

	latestSignals := make(map[int]bool)
LOOP:
	for {
		/*
			There are two situations for exiting.
			1. All goroutines don't generate new globalStates.
			2. There is no globalState in channel, so all goroutines will be blocked.
		*/

		// Situation 2
		if evm.NoStatesFlag {
			fmt.Println("break in situation 2")
			break LOOP
		}

		// Situation 1
		signal := <-evm.SignalCh
		latestSignals[signal.Id] = signal.Finished
		allFinished := true
		for i, finished := range latestSignals {
			fmt.Println(i, finished)
			if !finished {
				allFinished = false
			}
		}
		if allFinished {
			fmt.Println("break in situation 1")
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

func readWithSelect(evm *LaserEVM) (*state.GlobalState, error) {
	//fmt.Println(id)
	select {
	case globalState := <-evm.WorkList:
		return globalState, nil
	default:
		return nil, errors.New("evm.WorkList is empty")
	}
}

func (evm *LaserEVM) Run(id int) {
	fmt.Println("Run")
	for {
		globalState, _ := readWithSelect(evm)
		evm.NoStatesSignal[id] = globalState == nil
		if globalState != nil {
			fmt.Println(id, globalState, &globalState)
			newStates, opcode := evm.ExecuteState(globalState)
			//evm.ManageCFG(opcode, newStates)
			for _, newState := range newStates {
				evm.WorkList <- newState
				fmt.Println("append:", len(evm.WorkList))
			}
			fmt.Println(id, "done", globalState, opcode, &globalState)
			evm.SignalCh <- Signal{
				Id:       id,
				Finished: len(newStates) == 0,
			}
			fmt.Println("signal", id, len(newStates) == 0)
			fmt.Println("===========================================================================")
			if opcode == "STOP" || opcode == "RETURN" || opcode == "REVERT" {
				fmt.Println("before potentialIssues")
				modules.CheckPotentialIssues(globalState)
				fmt.Println("finish potentialIssues")
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
				// when the other goroutines have no globalStates to solve.
				flag := true
				for i, v := range evm.NoStatesSignal {
					if i != id {
						flag = flag && v
					}
				}
				if flag {
					fmt.Println("all goroutines have no globalStates")
					evm.NoStatesFlag = true
				}
			}
		}
	}
}
