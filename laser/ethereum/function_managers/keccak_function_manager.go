package function_managers

import (
	"go-mythril/laser/smt/z3"
	"math/big"
)

type KeccakFunctionManager struct {
	ctx *z3.Context
}

func NewKeccakFunctionManager(ctx *z3.Context) *KeccakFunctionManager {
	return &KeccakFunctionManager{
		ctx: ctx,
	}
}

func (k *KeccakFunctionManager) GetEmptyKeccakHash() *z3.Bitvec {
	val, _ := new(big.Int).SetString("1000", 10)
	return k.ctx.NewBitvecVal(val, 256)
}

func (k *KeccakFunctionManager) CreateKeccak() *z3.Bitvec {
	return k.ctx.NewBitvecVal(0, 256)
}
