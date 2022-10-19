package function_managers

import (
	"crypto"
	"go-mythril/laser/smt/z3"
	"go-mythril/utils"
	"math/big"
	"strconv"
)

type KeccakFunctionManager struct {
	Ctx             *z3.Context
	HashResultStore *map[int][]*z3.Bitvec
	HashMatcher     string
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
	val, _ := new(big.Int).SetString("89477152217924674838424037953991966239322087453347756267410168184682657981552", 10)
	return k.Ctx.NewBitvecVal(val, 256)
}

func (k *KeccakFunctionManager) CreateKeccak(data *z3.Bitvec) (*z3.Bitvec, *z3.Bool) {

	var funcData *z3.Bitvec
	if data.Symbolic() {
		funcData = k.Ctx.NewBitvec("kecak256_"+data.BvString(), 256)
	} else {
		dataV, _ := strconv.ParseInt(data.BvString()[2:], 16, 10)
		srcBuf := utils.IntToBytes(dataV)
		writer := crypto.SHA256.New()
		writer.Write(srcBuf)
		resBuf := writer.Sum(nil)
		resInt := utils.BytesToInt(resBuf)
		funcData = k.Ctx.NewBitvecVal(resInt, 256)
	}

	cons := k.Ctx.NewBitvecVal(1, 256).Eq(k.Ctx.NewBitvecVal(1, 256)).Simplify()
	return funcData, cons
}

//func (k *KeccakFunctionManager) CreateKeccak(data *z3.Bitvec) (*z3.Bitvec, *z3.Bool) {
//	length := data.BvSize()
//	//funcData := k.Ctx.NewBitvec("keccak256_"+strconv.Itoa(length)+"_"+data.BvString(), 256)
//	funcData := k.Ctx.NewBitvec("keccak256_"+strconv.Itoa(length), 256)
//	//funcData := k.Ctx.NewBitvecVal(1,256)
//	cons := k.Ctx.NewBitvecVal(1, 256).Eq(k.Ctx.NewBitvecVal(1, 256)).Simplify()
//	return funcData, cons
//}

func (k *KeccakFunctionManager) GetConcreteHashData(model *z3.Model) *map[int][]int {
	concreteHashes := make(map[int][]int)
	hashResult := *k.HashResultStore
	for size, _ := range hashResult {
		concreteHashes[size] = make([]int, 0)
		for _, val := range hashResult[size] {
			eval_ := model.Eval(val.AsAST(), false)
			concreteVal := eval_.Int()
			concreteHashes[size] = append(concreteHashes[size], concreteVal)
			// TODO: exception: AttributeError
		}
	}
	return &concreteHashes
}

func (k *KeccakFunctionManager) FindConcreteKeccak(data *z3.Bitvec) *z3.Bitvec {
	// TODO: Implementations of sha3 in Golang
	keccak := k.Ctx.NewBitvecVal(0, 256)
	return keccak
}
