package function_managers

import (
	"crypto"
	"fmt"
	"go-mythril/laser/smt/z3"
	"go-mythril/utils"
	"math/big"
	"strconv"
	"sync"
)

type KeccakFunctionManager struct {
	Ctx             *z3.Context
	HashResultStore *map[int][]*z3.Bitvec
	ConcreteHashes  *map[*z3.Bitvec]*z3.Bitvec
	HashMatcher     string
}

var keccakFunctionManager *KeccakFunctionManager
var once sync.Once

func NewKeccakFunctionManager(ctx *z3.Context) *KeccakFunctionManager {
	once.Do(func() {
		hashStore := make(map[int][]*z3.Bitvec, 0)
		concreteHashes := make(map[*z3.Bitvec]*z3.Bitvec)
		keccakFunctionManager = &KeccakFunctionManager{
			HashResultStore: &hashStore,
			ConcreteHashes:  &concreteHashes,
			// This is usually the prefix for the hash in the output
			HashMatcher: "fffffff",
		}
	})
	keccakFunctionManager.Ctx = ctx
	return keccakFunctionManager
}

func RefreshKeccak() {
	hashStore := make(map[int][]*z3.Bitvec, 0)
	concreteHashes := make(map[*z3.Bitvec]*z3.Bitvec)
	keccakFunctionManager.HashResultStore = &hashStore
	keccakFunctionManager.ConcreteHashes = &concreteHashes
}

func (k *KeccakFunctionManager) GetEmptyKeccakHash() *z3.Bitvec {
	val, _ := new(big.Int).SetString("89477152217924674838424037953991966239322087453347756267410168184682657981552", 10)
	return k.Ctx.NewBitvecVal(val, 256)
}

func (k *KeccakFunctionManager) getFunction(length int) (*z3.FuncDecl, *z3.FuncDecl) {
	domain1 := []*z3.Sort{k.Ctx.BvSort(uint(length))}
	keccakFun := k.Ctx.NewFuncDecl("kecak256_"+strconv.Itoa(length), domain1, k.Ctx.BvSort(256))
	domain2 := []*z3.Sort{k.Ctx.BvSort(256)}
	inverse := k.Ctx.NewFuncDecl("kecak256_"+strconv.Itoa(length)+"-1", domain2, k.Ctx.BvSort(uint(length)))
	return keccakFun, inverse
}

func (k *KeccakFunctionManager) createCondition(funcInput *z3.Bitvec) *z3.Bool {
	length := funcInput.BvSize()
	keccakFun, inverse := k.getFunction(length)

	keccakRes := keccakFun.ApplyBv(funcInput)

	TOTAL_PARTS := new(big.Int).Exp(big.NewInt(10), big.NewInt(40), nil)
	PART_Molecular := new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil), big.NewInt(1))
	PART := new(big.Int).Div(PART_Molecular, TOTAL_PARTS)

	index := new(big.Int).Sub(TOTAL_PARTS, big.NewInt(34534))
	lowerBound := new(big.Int).Mul(index, PART)
	upperBound := new(big.Int).Add(lowerBound, PART)

	cond := inverse.ApplyBv(keccakRes).Eq(funcInput).And(
		k.Ctx.NewBitvecVal(lowerBound, 256).BvULe(keccakRes),
		keccakRes.BvULt(k.Ctx.NewBitvecVal(upperBound, 256)),
		keccakRes.BvURem(k.Ctx.NewBitvecVal(64, 256)).Eq(k.Ctx.NewBitvecVal(0, 256)),
	)

	concreteCond := k.Ctx.NewBitvecVal(1, 256).Eq(k.Ctx.NewBitvecVal(0, 256)).Simplify()
	for key, keccak := range *k.ConcreteHashes {
		key = key.Translate(k.Ctx)
		keccak = keccak.Translate(k.Ctx)
		if key.BvSize() != funcInput.BvSize() {
			continue
		}
		hashEq := keccakFun.ApplyBv(funcInput).Eq(keccak).And(key.Eq(funcInput))
		concreteCond = concreteCond.Or(hashEq)
	}
	return inverse.ApplyBv(keccakRes).Eq(funcInput).And(cond.Or(concreteCond))
	//return k.Ctx.NewBitvecVal(1,256).Eq(k.Ctx.NewBitvecVal(1,256)).Simplify()
}

func (k *KeccakFunctionManager) CreateKeccak(data *z3.Bitvec) (*z3.Bitvec, *z3.Bool) {

	var funcData *z3.Bitvec
	var condition *z3.Bool

	length := data.BvSize()
	keccakFun, inverse := k.getFunction(length)

	if data.Symbolic() {
		//funcData = k.Ctx.NewBitvec("kecak256_"+funcInput.BvString(), 256)
		funcData = keccakFun.ApplyBv(data)
		condition = k.createCondition(data)
	} else {
		funcData = k.FindConcreteKeccak(data)
		concreteHashes := *k.ConcreteHashes
		cfg := z3.GetConfig()
		ctx := z3.NewContext(cfg)
		concreteHashes[data.Translate(ctx)] = funcData.Translate(ctx)
		fmt.Println("beforeHere!", inverse)
		condition = keccakFun.ApplyBv(data).Eq(funcData).And(inverse.ApplyBv(keccakFun.ApplyBv(data)).Eq(data))
		fmt.Println("afterHere!")
	}
	//condition = k.Ctx.NewBitvecVal(1, 256).Eq(k.Ctx.NewBitvecVal(1, 256)).Simplify()
	return funcData, condition
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
	//srcBuf := utils.IntToBytes(dataV)
	srcBuf := []byte(data.BvString()[2:])
	writer := crypto.SHA256.New()
	writer.Write(srcBuf)
	fmt.Println("resBuf")
	resBuf := writer.Sum(nil)
	fmt.Println(resBuf)

	resInt := utils.BytesToInt(resBuf)
	keccak := k.Ctx.NewBitvecVal(resInt, 256)
	return keccak
}
