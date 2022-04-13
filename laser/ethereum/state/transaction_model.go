package state

import (
	"encoding/hex"
	"go-mythril/disassembler"
	"go-mythril/laser/smt/z3"
	"math/big"
	"strconv"
)

var nextTransactionId int

func GetNextTransactionId() string {
	nextTransactionId += 1
	return string(nextTransactionId)
}

type BaseTransaction interface {
	InitialGlobalStateFromEnvironment(environment *Environment, activeFunc string) *GlobalState
	InitialGlobalState() *GlobalState
	GetId() string
	GetGasLimit() int
	End(state *GlobalState, returnData []byte)
}

type MessageCallTransaction struct {
	Code          *disassembler.Disasembly
	CalleeAccount *Account
	Caller        *z3.Bitvec
	Calldata      BaseCalldata
	GasPrice      int
	GasLimit      int
	CallValue     int
	Origin        *z3.Bitvec
	Basefee       *z3.Bitvec
	Ctx           *z3.Context
	Id            string
	ReturnData    []byte
}

func NewMessageCallTransaction(code string) *MessageCallTransaction {
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
	// calldataList = append(calldataList, ctx.NewBitvecVal(0, 8))
	// For test
	caller, _ := new(big.Int).SetString("5B38Da6a701c568545dCfcB03FcB875f56beddC4", 16)
	return &MessageCallTransaction{
		Code: txcode,
		// TODO: For test here.
		CalleeAccount: NewAccount(ctx.NewBitvecVal(123, 256),
			ctx.NewArray("balances", 256, 256), true, txcode),
		Caller:    ctx.NewBitvecVal(caller, 256),
		Calldata:  NewConcreteCalldata("txid123", calldataList),
		GasPrice:  10,
		GasLimit:  100,
		CallValue: 0,
		Origin:    ctx.NewBitvec("origin", 256),
		Basefee:   ctx.NewBitvecVal(1000, 256),
		Ctx:       ctx,
		Id:        GetNextTransactionId(),
	}
}

func (tx *MessageCallTransaction) InitialGlobalStateFromEnvironment(env *Environment, activeFunc string) *GlobalState {
	txStack := make([]BaseTransaction, 0)
	globalState := NewGlobalState(env, tx.Ctx, append(txStack, tx))
	globalState.Environment.ActiveFuncName = activeFunc
	sender := env.Sender
	receiver := env.ActiveAccount.Address
	value := tx.Ctx.NewBitvecVal(env.CallValue, 256)
	constrain := globalState.WorldState.Balances.GetItem(sender).BvUGe(value)
	globalState.WorldState.Constraints.Add(constrain)

	receiverV := globalState.WorldState.Balances.GetItem(receiver)
	senderV := globalState.WorldState.Balances.GetItem(sender)
	globalState.WorldState.Balances.SetItem(receiver, receiverV.BvAdd(value).Simplify())
	globalState.WorldState.Balances.SetItem(sender, senderV.BvSub(value).Simplify())

	return globalState
}

func (tx *MessageCallTransaction) InitialGlobalState() *GlobalState {
	environment := NewEnvironment(tx.Code, tx.CalleeAccount,
		tx.Caller, tx.Calldata, tx.GasPrice, tx.CallValue, tx.Origin, tx.Basefee)
	return tx.InitialGlobalStateFromEnvironment(environment, "fallback")
}

func (tx *MessageCallTransaction) End(state *GlobalState, data []byte) {
	tx.ReturnData = data
	panic("TransactionEndSignal")
}

func (tx *MessageCallTransaction) GetId() string {
	return tx.Id
}

func (tx *MessageCallTransaction) GetGasLimit() int {
	return tx.GasLimit
}

type ContractCreationTransaction struct {
	Code          *disassembler.Disasembly
	CalleeAccount *Account
	Caller        *z3.Bitvec
	Calldata      BaseCalldata
	GasPrice      int
	GasLimit      int
	CallValue     int
	Origin        *z3.Bitvec
	Basefee       *z3.Bitvec
	Ctx           *z3.Context
	Id            string
	ReturnData    []byte
}

func NewContractCreationTransaction(code string) *ContractCreationTransaction {
	config := z3.NewConfig()
	ctx := z3.NewContext(config)
	txcode := disassembler.NewDisasembly(code)
	calldataList := make([]*z3.Bitvec, 0)
	// TODO: For test here
	// Remix input here
	//
	inputStr := "6080604052348015600f57600080fd5b5060ad8061001e6000396000f3fe6080604052348015600f57600080fd5b506004361060285760003560e01c8063f8a8fd6d14602d575b600080fd5b60336035565b005b336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555056fea2646970667358221220cc40cae2e419544393419c7a7ea32f42d341094e9dc31099df83cbe79983591164736f6c63430008070033"
	for i := 0; i < len(inputStr); i = i + 2 {
		tmp := string(inputStr[i]) + string(inputStr[i+1])
		tmpV, _ := strconv.ParseInt(tmp, 16, 64)
		calldataList = append(calldataList, ctx.NewBitvecVal(tmpV, 8))
	}
	// set your callerValue in remix to test
	// For test
	caller, _ := new(big.Int).SetString("5B38Da6a701c568545dCfcB03FcB875f56beddC4", 16)
	return &ContractCreationTransaction{
		Code: txcode,
		// TODO: For test here.
		CalleeAccount: NewAccount(ctx.NewBitvecVal(123, 256),
			ctx.NewArray("balances", 256, 256), true, txcode),
		Caller:    ctx.NewBitvecVal(caller, 256),
		Calldata:  NewConcreteCalldata("txid123", calldataList),
		GasPrice:  10,
		GasLimit:  100000,
		CallValue: 0,
		Origin:    ctx.NewBitvecVal(caller, 256),
		Basefee:   ctx.NewBitvecVal(1000, 256),
		Ctx:       ctx,
		Id:        GetNextTransactionId(),
	}
}

func (tx *ContractCreationTransaction) InitialGlobalStateFromEnvironment(env *Environment, activeFunc string) *GlobalState {
	txStack := make([]BaseTransaction, 0)
	globalState := NewGlobalState(env, tx.Ctx, append(txStack, tx))
	globalState.Environment.ActiveFuncName = activeFunc
	sender := env.Sender
	receiver := env.ActiveAccount.Address
	value := tx.Ctx.NewBitvecVal(env.CallValue, 256)
	constrain := globalState.WorldState.Balances.GetItem(sender).BvUGe(value)
	globalState.WorldState.Constraints.Add(constrain)

	receiverV := globalState.WorldState.Balances.GetItem(receiver)
	senderV := globalState.WorldState.Balances.GetItem(sender)
	globalState.WorldState.Balances.SetItem(receiver, receiverV.BvAdd(value).Simplify())
	globalState.WorldState.Balances.SetItem(sender, senderV.BvSub(value).Simplify())

	return globalState
}

func (tx *ContractCreationTransaction) InitialGlobalState() *GlobalState {
	environment := NewEnvironment(tx.Code, tx.CalleeAccount,
		tx.Caller, tx.Calldata, tx.GasPrice, tx.CallValue, tx.Origin, tx.Basefee)
	return tx.InitialGlobalStateFromEnvironment(environment, "constructor")
}

func (tx *ContractCreationTransaction) End(globalState *GlobalState, data []byte) {
	if len(data) == 0 {
		tx.ReturnData = nil
		panic("TransactionEndSignal")
	}

	globalState.Environment.ActiveAccount.Code.AssignBytecode(data)
	newData, _ := hex.DecodeString(globalState.Environment.ActiveAccount.Address.Value())
	tx.ReturnData = newData

	if len(globalState.Environment.ActiveAccount.Code.InstructionList) == 0 {
		panic("AssertError: instructionList == []")
	} else {
		panic("TransactionEndSignal")
	}
}

func (tx *ContractCreationTransaction) GetId() string {
	return tx.Id
}

func (tx *ContractCreationTransaction) GetGasLimit() int {
	return tx.GasLimit
}
