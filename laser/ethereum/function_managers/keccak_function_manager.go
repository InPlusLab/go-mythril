package function_managers

import (
	"go-mythril/laser/smt/z3"
	"math/big"
)

type KeccakFunctionManager struct {
	Ctx *z3.Context
	HashResultStore *map[int][]*z3.Bitvec
	HashMatcher string
}

var hashStore = make(map[int][]*z3.Bitvec, 0)
var keccakFunctionManager = &KeccakFunctionManager{
	HashResultStore: &hashStore,
	// This is usually the prefix for the hash in the output
	HashMatcher: "fffffff",
}

func NewKeccakFunctionManager(ctx *z3.Context) *KeccakFunctionManager {
	keccakFunctionManager.Ctx = ctx
	return keccakFunctionManager
}

func (k *KeccakFunctionManager) GetEmptyKeccakHash() *z3.Bitvec {
	val, _ := new(big.Int).SetString("1000", 10)
	return k.Ctx.NewBitvecVal(val, 256)
}

func (k *KeccakFunctionManager) CreateKeccak() *z3.Bitvec {
	return k.Ctx.NewBitvecVal(0, 256)
}

func (k *KeccakFunctionManager) GetConcreteHashData(model *z3.Model) *map[int][]int {
	concreteHashes := make(map[int][]int)
	hashResult := *k.HashResultStore
	for size, _ := range hashResult{
		concreteHashes[size] = make([]int,0)
		for _, val := range hashResult[size] {
			eval_ := model.Eval(val.AsAST(), false)
			concreteVal := eval_.Int()
			concreteHashes[size] = append(concreteHashes[size], concreteVal)
			// TODO: exception: AttributeError
		}
	}
	return &concreteHashes
}

func (k *KeccakFunctionManager) FindConcreteKeccak(data *z3.Bitvec) *z3.Bitvec{
	// TODO: Implementations of sha3 in Golang
	keccak := k.Ctx.NewBitvecVal(0,256)
	return keccak
}