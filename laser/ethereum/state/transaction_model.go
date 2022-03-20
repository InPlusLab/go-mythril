package state

import (
	"go-mythril/disassembler"
	"go-mythril/laser/smt/z3"
)

type BaseTransaction struct {
	Code          *disassembler.Disasembly
	CalleeAccount *Account
	Caller        *z3.Bitvec
	Calldata      BaseCalldata
	GasPrice      *z3.Bitvec
	GasLimit      *z3.Bitvec
	CallValue     *z3.Bitvec
	Origin        *z3.Bitvec
	Basefee       *z3.Bitvec
	Ctx           *z3.Context
}

func NewBaseTransaction(code string) *BaseTransaction {
	config := z3.NewConfig()
	ctx := z3.NewContext(config)
	txcode := disassembler.NewDisasembly(code)
	calldataList := make([]*z3.Bitvec, 0)
	// Function hash: 0xf8a8fd6d, which is the hash of test() in origin.sol
	calldataList = append(calldataList, ctx.NewBitvecVal(248, 8))
	calldataList = append(calldataList, ctx.NewBitvecVal(168, 8))
	calldataList = append(calldataList, ctx.NewBitvecVal(253, 8))
	calldataList = append(calldataList, ctx.NewBitvecVal(109, 8))
	// Parameters
	calldataList = append(calldataList, ctx.NewBitvecVal(0, 8))
	return &BaseTransaction{
		Code: txcode,
		// TODO: For test here.
		CalleeAccount: NewAccount(ctx.NewBitvecVal(123, 256), ctx.NewArray("balances", 256, 256), txcode),
		Caller:        ctx.NewBitvec("sender", 256),
		Calldata:      NewConcreteCalldata("txid123", calldataList),
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
func (tx *BaseTransaction) InitialGlobalState() *GlobalState {
	environment := NewEnvironment(tx.Code, tx.CalleeAccount,
		tx.Caller, tx.Calldata, tx.GasPrice, tx.CallValue, tx.Origin, tx.Basefee)
	globalState := NewGlobalState(environment, tx.Ctx)
	return globalState
}
