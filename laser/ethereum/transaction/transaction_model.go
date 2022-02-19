package transaction

import (
	"go-mythril/disassembler"
	"go-mythril/laser/ethereum/state"
)

type BaseTransaction struct {
	code *disassembler.Disasembly
}

func NewBaseTransaction(code string) *BaseTransaction {
	return &BaseTransaction{
		code: disassembler.NewDisasembly(code),
	}
}

// TODO: this trait is not for BaseTransaction
func (tx *BaseTransaction) InitialGlobalState() *state.GlobalState {
	environment := state.NewEnvironment(tx.code)
	globalState := state.NewGlobalState(environment)
	return globalState
}
