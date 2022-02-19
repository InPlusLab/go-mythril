package ethereum

import (
	"fmt"
	"go-mythril/laser/ethereum/state"
	"go-mythril/laser/ethereum/transaction"
	"time"
	//"go-mythril/laser/smt"
)

type LaserEVM struct {
	ExecutionTimeout int
	CreateTimeout    int
	TransactionCount int
	WorkList         chan *state.GlobalState
	// FinalState       chan *state.GlobalState

	// Parallal
	BeginCh     chan int
	EndCh       chan int
	GofuncCount int
}

func NewLaserEVM(ExecutionTimeout int, CreateTimeout int, TransactionCount int) *LaserEVM {
	evm := LaserEVM{
		ExecutionTimeout: ExecutionTimeout,
		CreateTimeout:    CreateTimeout,
		TransactionCount: TransactionCount,
		WorkList:         make(chan *state.GlobalState, 1000),
		// FinalState:       make(chan *state.GlobalState),
		BeginCh:     make(chan int),
		EndCh:       make(chan int),
		GofuncCount: 4,
	}
	return &evm
}

func (evm *LaserEVM) SymExec(CreationCode string) {
	fmt.Println("Symbolic Executing: ", CreationCode)

	// TOOD: actually creation code is not for base tx, but for creation tx, just for test here
	tx := transaction.NewBaseTransaction(CreationCode)
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

	instr := NewInstruction(opcode)
	newGlobalStates := instr.Evaluate(globalState)

	return newGlobalStates, opcode
}

func (evm *LaserEVM) ManageCFG(opcode string, newStates []*state.GlobalState) {
	// TODO
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
