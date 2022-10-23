package state

import (
	"fmt"
	"go-mythril/disassembler"
	"go-mythril/laser/smt/z3"
	"go-mythril/utils"
	"strconv"
)

type Account struct {
	Address      *z3.Bitvec
	Balances     *z3.Array
	Storage      *Storage
	Code         *disassembler.Disasembly
	ContractName string
	Deleted      bool
}

func NewAccount(addr *z3.Bitvec, balances *z3.Array, concreteStorage bool, code *disassembler.Disasembly, contractName string) *Account {
	return &Account{
		Address:      addr,
		Balances:     balances,
		Storage:      NewStorage(addr, concreteStorage),
		Code:         code,
		ContractName: contractName,
		Deleted:      false,
	}
}
func (acc *Account) Copy() *Account {
	tmp := &Account{
		Address:      acc.Address,
		Code:         acc.Code,
		Balances:     acc.Balances,
		Storage:      acc.Storage.DeepCopy(),
		ContractName: acc.ContractName,
		Deleted:      acc.Deleted,
	}
	return tmp
}
func (acc *Account) Translate(ctx *z3.Context) *Account {
	//return &Account{
	//	Address:      acc.Address.Translate(ctx),
	//	Balances:     acc.Balances.Translate(ctx).(*z3.Array),
	//	Storage:      acc.Storage.Translate(ctx),
	//	Code:         acc.Code,
	//	ContractName: acc.ContractName,
	//	Deleted:      acc.Deleted,
	//}
	// Tips: Account == Storage, and Translate != Copy
	acc.Address = acc.Address.Translate(ctx)
	acc.Balances = acc.Balances.Translate(ctx).(*z3.Array)
	acc.Storage = acc.Storage.Translate(ctx)
	return acc
}
func (acc *Account) Balance() *z3.Bitvec {
	return acc.Balances.GetItem(acc.Address)
}
func (acc *Account) SetBalance(balance *z3.Bitvec) {
	acc.Balances.SetItem(acc.Address, balance)
}

type Storage struct {
	Address          *z3.Bitvec
	StandardStorage  z3.BaseArray
	PrintableStorage map[*z3.Bitvec]*z3.Bitvec
	// set(int)
	StorageKeysLoaded *utils.Set
	// set(bv)
	KeysSet *utils.Set
}

func NewStorage(addr *z3.Bitvec, concrete bool) *Storage {
	// concrete: bool indicating whether to interpret
	// uninitialized storage as concrete versus symbolic
	ctx := addr.GetCtx()
	var sstorage z3.BaseArray
	if concrete {
		sstorage = ctx.NewK(256, 256, 0)
	} else {
		sstorage = ctx.NewArray("Storage"+addr.BvString(), 256, 256)
	}

	return &Storage{
		Address:           addr,
		StandardStorage:   sstorage,
		PrintableStorage:  make(map[*z3.Bitvec]*z3.Bitvec),
		StorageKeysLoaded: utils.NewSet(),
		KeysSet:           utils.NewSet(),
	}
}
func (s *Storage) Translate(ctx *z3.Context) *Storage {
	newPrintableStorage := make(map[*z3.Bitvec]*z3.Bitvec)
	for i, v := range s.PrintableStorage {
		newPrintableStorage[i.Translate(ctx)] = v.Translate(ctx)
	}
	newKeysSet := utils.NewSet()
	for _, v := range s.KeysSet.Elements() {
		newKeysSet.Add(v.(*z3.Bitvec).Translate(ctx))
	}
	return &Storage{
		Address:           s.Address.Translate(ctx),
		StandardStorage:   s.StandardStorage.Translate(ctx),
		PrintableStorage:  newPrintableStorage,
		StorageKeysLoaded: s.StorageKeysLoaded,
		KeysSet:           newKeysSet,
	}
}
func (s *Storage) DeepCopy() *Storage {
	var concrete bool
	switch s.StandardStorage.(type) {
	case *z3.K:
		concrete = true
	default:
		concrete = false
	}
	newStorage := NewStorage(s.Address, concrete)
	newStorage.StandardStorage = s.StandardStorage.DeepCopy()
	newStorage.PrintableStorage = s.PrintableStorage
	newStorage.StorageKeysLoaded = s.StorageKeysLoaded
	newStorage.KeysSet = s.KeysSet

	return newStorage
}
func (s *Storage) GetItem(item *z3.Bitvec) *z3.Bitvec {
	storage := s.StandardStorage
	fmt.Println("index:", item.BvString())
	//ctx := item.GetCtx()
	//itemV, _ := strconv.ParseInt(s.Address.Value(), 10, 64)
	//storageKeysLoaded := s.StorageKeysLoaded
	//inKeysLoaded := storageKeysLoaded.Contains(itemV)
	// TODO: dynLoader
	//if !item.Symbolic() && itemV != 0 && !inKeysLoaded {
	//	fmt.Println("#1")
	//
	//	value := ctx.NewBitvecVal(0, 256)
	//	for _, key := range s.KeysSet.Elements() {
	//		value = z3.If(key.(*z3.Bitvec).Eq(item), storage.GetItem(item).Simplify(), value)
	//	}
	//	fmt.Println("#2")
	//	// storage.SetItem(item, value)
	//	s.StorageKeysLoaded.Add(itemV)
	//	s.PrintableStorage[item] = value
	//}
	if item.Symbolic() {
		panic("can't get item in Storage using symbolic index!")
	}
	result := storage.GetItem(item).Simplify()
	//if strings.Contains(result.BvString(), "_") {
	//	fmt.Println("StorageSymbolicResult: ", result.BvString())
	//	return ctx.NewBitvecVal(0, 256)
	//}
	//itemStr := item.BvString()
	//for k, v := range s.PrintableStorage {
	//	if k.BvString() == itemStr {
	//		return v.Translate(ctx)
	//	}
	//}
	return result
}
func (s *Storage) SetItem(key *z3.Bitvec, value *z3.Bitvec) {
	printableStorage := s.PrintableStorage
	printableStorage[key] = value
	s.StandardStorage.SetItem(key, value)
	s.KeysSet.Add(key)

	if !key.Symbolic() {
		keyV, _ := strconv.ParseInt(key.Value(), 10, 64)
		storageKeysLoaded := s.StorageKeysLoaded
		storageKeysLoaded.Add(keyV)
	}
}
