package ethereum

import (
	"fmt"
	"go-mythril/laser/ethereum/state"
	//"go-mythril/laser/smt"
)

type LaserEVM struct {
	ExecutionTimeout int
	CreateTimeout    int
	TransactionCount int
	WorkList         chan *state.GlobalState
	FinalState       chan *state.GlobalState
}

func NewLaserEVM(ExecutionTimeout int, CreateTimeout int, TransactionCount int) *LaserEVM {
	evm := LaserEVM{
		ExecutionTimeout: ExecutionTimeout,
		CreateTimeout:    CreateTimeout,
		TransactionCount: TransactionCount,
		WorkList:         make(chan *state.GlobalState),
		FinalState:       make(chan *state.GlobalState),
	}
	go evm.Run()
	go evm.Run()
	go evm.Run()
	return &evm
}

func (evm *LaserEVM) SymExec(CreationCode string) {
	fmt.Println("Symbolic Executing: ", CreationCode)
	evm.WorkList <- state.NewGlobalState()
	finalState := <-evm.FinalState
	fmt.Println("Finish", finalState)
}

func (evm *LaserEVM) Run() {
	globalState := <-evm.WorkList
	fmt.Println("Run", globalState)
	// TODO: test smt here:
	// globalState.WorldState.Constraints.Add(&smt.Bool) (a<1)
	// globalState.WorldState.Constraints.IsPossible()
	evm.FinalState <- globalState
}
