package transaction

// In Golang,

import (
	"go-mythril/disassembler"
	"go-mythril/laser/ethereum/state"
	"go-mythril/laser/smt/z3"
)

type BaseTransaction struct {
	Code          *disassembler.Disasembly
	CalleeAccount *state.Account
	Caller        *z3.Bitvec
	//TODO: calldata
	GasPrice  *z3.Bitvec
	GasLimit  *z3.Bitvec
	CallValue *z3.Bitvec
	Origin    *z3.Bitvec
	Basefee   *z3.Bitvec
	Ctx       *z3.Context
}

func NewBaseTransaction(code string) *BaseTransaction {
	config := z3.NewConfig()
	ctx := z3.NewContext(config)
	txcode := disassembler.NewDisasembly(code)
	return &BaseTransaction{
		Code: txcode,
		// TODO: For test here.
		CalleeAccount: state.NewAccount(ctx.NewBitvecVal(123, 256), ctx.NewArray("balances", 256, 256), txcode),
		Caller:        ctx.NewBitvec("sender", 256),
		GasPrice:      ctx.NewBitvecVal(10, 256),
		GasLimit:      ctx.NewBitvecVal(100, 256),
		CallValue:     ctx.NewBitvecVal(100, 256),
		Origin:        ctx.NewBitvec("origin", 256),
		Basefee:       ctx.NewBitvecVal(1000, 256),
		Ctx:           ctx,
	}
}

func (tx *BaseTransaction) GetCode() *disassembler.Disasembly {
	return tx.Code
}

// TODO: this trait is not for BaseTransaction
func (tx *BaseTransaction) InitialGlobalState() *state.GlobalState {
	environment := state.NewEnvironment(tx.Code, tx.CalleeAccount, tx.Caller, tx.GasPrice, tx.CallValue, tx.Origin, tx.Basefee)
	globalState := state.NewGlobalState(environment, tx.Ctx)
	return globalState
}
