package state

import (
	"go-mythril/disassembler"
	"go-mythril/laser/smt/z3"
)

type Environment struct {
	Code          *disassembler.Disasembly
	ActiveAccount *Account
	Address       *z3.Bitvec
	Sender        *z3.Bitvec
	Calldata      BaseCalldata
	GasPrice      *z3.Bitvec
	CallValue     *z3.Bitvec
	Origin        *z3.Bitvec
	Basefee       *z3.Bitvec
	ChainId       *z3.Bitvec
	Stack         []*z3.Bitvec
}

func NewEnvironment(code *disassembler.Disasembly,
	account *Account,
	sender *z3.Bitvec,
	calldata BaseCalldata,
	gasprice *z3.Bitvec,
	callvalue *z3.Bitvec,
	origin *z3.Bitvec,
	basefee *z3.Bitvec) *Environment {
	stack := make([]*z3.Bitvec, 0)
	return &Environment{
		Code:          code,
		ActiveAccount: account,
		Address:       account.Address,
		Sender:        sender,
		Calldata:      calldata,
		GasPrice:      gasprice,
		CallValue:     callvalue,
		Origin:        origin,
		Basefee:       basefee,
		ChainId:       sender.GetCtx().NewBitvec("chain_id", 256),
		Stack:         stack,
	}
}
