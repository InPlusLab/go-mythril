package state

import (
	"go-mythril/disassembler"
	"go-mythril/laser/smt/z3"
)

type Environment struct {
	Code           *disassembler.Disasembly
	ActiveAccount  *Account
	Address        *z3.Bitvec
	Sender         *z3.Bitvec
	Calldata       BaseCalldata
	GasPrice       int
	CallValue      int
	Origin         *z3.Bitvec
	Basefee        *z3.Bitvec
	ChainId        *z3.Bitvec
	BlockNumber    *z3.Bitvec
	ActiveFuncName string
	Stack          []*z3.Bitvec
}

func NewEnvironment(code *disassembler.Disasembly,
	account *Account,
	sender *z3.Bitvec,
	calldata BaseCalldata,
	gasprice int,
	callvalue int,
	origin *z3.Bitvec,
	basefee *z3.Bitvec) *Environment {
	stack := make([]*z3.Bitvec, 0)
	return &Environment{
		Code:           code,
		ActiveAccount:  account,
		Address:        account.Address,
		Sender:         sender,
		Calldata:       calldata,
		GasPrice:       gasprice,
		CallValue:      callvalue,
		Origin:         origin,
		Basefee:        basefee,
		ChainId:        sender.GetCtx().NewBitvec("chain_id", 256),
		BlockNumber:    sender.GetCtx().NewBitvec("block_number", 256),
		ActiveFuncName: "ActiveFuncName",
		Stack:          stack,
	}
}

// shallow copy
func (env *Environment) Copy() *Environment {
	newStack := make([]*z3.Bitvec, 0)
	for _, v := range env.Stack {
		newStack = append(newStack, v)
	}
	return &Environment{
		Code: env.Code,
		//ActiveAccount:  env.ActiveAccount.Copy(),
		ActiveAccount:  env.ActiveAccount.Copy(),
		Address:        env.Address.Copy(),
		Sender:         env.Sender.Copy(),
		Calldata:       env.Calldata,
		GasPrice:       env.GasPrice,
		CallValue:      env.CallValue,
		Origin:         env.Origin.Copy(),
		Basefee:        env.Basefee.Copy(),
		ChainId:        env.ChainId.Copy(),
		BlockNumber:    env.BlockNumber.Copy(),
		ActiveFuncName: env.ActiveFuncName,
		Stack:          newStack,
	}
}

func (env *Environment) Translate(ctx *z3.Context) *Environment {
	return &Environment{
		Code:           env.Code,
		ActiveAccount:  env.ActiveAccount.Translate(ctx),
		Address:        env.Address.Translate(ctx),
		Sender:         env.Sender.Translate(ctx),
		Calldata:       env.Calldata.Translate(ctx),
		GasPrice:       env.GasPrice,
		CallValue:      env.CallValue,
		Origin:         env.Origin.Translate(ctx),
		Basefee:        env.Basefee.Translate(ctx),
		ChainId:        env.ChainId.Translate(ctx),
		BlockNumber:    env.BlockNumber.Translate(ctx),
		ActiveFuncName: env.ActiveFuncName,
		Stack:          env.Stack,
	}
	//env.ActiveAccount = env.ActiveAccount.Translate(ctx)
	//env.Address = env.Address.Translate(ctx)
	//env.Sender = env.Sender.Translate(ctx)
	//env.Calldata = env.Calldata.Translate(ctx)
	//env.Origin = env.Origin.Translate(ctx)
	//env.Basefee = env.Basefee.Translate(ctx)
	//env.ChainId = env.ChainId.Translate(ctx)
	//env.BlockNumber = env.BlockNumber.Translate(ctx)
	//
	//return env
}
