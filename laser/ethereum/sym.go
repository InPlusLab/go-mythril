package ethereum

import (
	"fmt"
	"go-mythril/analysis"
	"go-mythril/analysis/module/modules"
	"go-mythril/laser/ethereum/state"
	"go-mythril/laser/smt/z3"
	"go-mythril/support"
	"go-mythril/utils"
	"time"
)

type moduleExecFunc func(globalState *state.GlobalState) []*analysis.Issue

type LaserEVM struct {
	ExecutionTimeout int
	CreateTimeout    int
	TransactionCount int
	WorkList         chan *state.GlobalState
	// FinalState       chan *state.GlobalState
	InstrPreHook  *map[string][]moduleExecFunc
	InstrPostHook *map[string][]moduleExecFunc
	// Parallal
	BeginCh     chan int
	EndCh       chan int
	GofuncCount int
}

func NewLaserEVM(ExecutionTimeout int, CreateTimeout int, TransactionCount int) *LaserEVM {

	preHook := make(map[string][]moduleExecFunc)
	postHook := make(map[string][]moduleExecFunc)
	opcodes := *support.NewOpcodes()
	for _, v := range opcodes {
		preHook[v.Name] = make([]moduleExecFunc, 0)
		postHook[v.Name] = make([]moduleExecFunc, 0)
	}
	// For hook test here~
	// TODO: svm.py - register_instr_hooks
	integerDetectionModule := modules.NewIntegerArithmetics()
	preHook["PUSH1"] = []moduleExecFunc{integerDetectionModule.Execute}
	postHook["PUSH1"] = []moduleExecFunc{integerDetectionModule.Execute}

	evm := LaserEVM{
		ExecutionTimeout: ExecutionTimeout,
		CreateTimeout:    CreateTimeout,
		TransactionCount: TransactionCount,
		WorkList:         make(chan *state.GlobalState, 1000),
		// FinalState:       make(chan *state.GlobalState),
		InstrPreHook:  &preHook,
		InstrPostHook: &postHook,

		BeginCh:     make(chan int),
		EndCh:       make(chan int),
		GofuncCount: 4,
	}
	return &evm
}

func (evm *LaserEVM) NormalSymExec(CreationCode string) {
	fmt.Println("Symbolic Executing: ", CreationCode)
	fmt.Println("")
	tx := state.NewContractCreationTransaction(CreationCode)
	globalState := tx.InitialGlobalState()
	evm.WorkList <- globalState
	id := 0
	for {
		globalState := <-evm.WorkList
		fmt.Println(id, globalState)
		newStates, opcode := evm.ExecuteState(globalState)
		evm.ManageCFG(opcode, newStates)

		for _, newState := range newStates {
			evm.WorkList <- newState
		}
		fmt.Println(id, "done", globalState, opcode)
		fmt.Println("======================================================")
		if opcode == "STOP" {
			break
		}
		id++
	}
}

func (evm *LaserEVM) SymExec(CreationCode string) {
	fmt.Println("Symbolic Executing: ", CreationCode)

	// TOOD: actually creation code is not for base tx, but for creation tx, just for test here
	tx := state.NewMessageCallTransaction(CreationCode)
	globalState := tx.InitialGlobalState()
	evm.WorkList <- globalState
	for i := 0; i < evm.GofuncCount; i++ {
		go evm.Run(i)
	}
	// TODO: not good here
	beginCount := 0
	endCount := 0
LOOP:
	for {
		select {
		case <-evm.BeginCh:
			beginCount++
		case <-evm.EndCh:
			endCount++
			if endCount == beginCount {
				fmt.Println("finish", beginCount, endCount)
				break LOOP
			} else {
				fmt.Println("not finish", beginCount, endCount)
			}
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


func (evm *LaserEVM) Run(id int) {
	fmt.Println("Run")
	for {
		globalState := <-evm.WorkList
		evm.BeginCh <- id
		fmt.Println(id, globalState)
		newStates, opcode := evm.ExecuteState(globalState)
		evm.ManageCFG(opcode, newStates)

		for _, newState := range newStates {
			evm.WorkList <- newState
		}
		fmt.Println(id, "done", globalState, opcode)

		// TODO not good for sleep
		time.Sleep(time.Second)
		evm.EndCh <- id
	}
}
