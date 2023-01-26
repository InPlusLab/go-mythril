package state

import (
	"encoding/hex"
	"fmt"
	"go-mythril/disassembler"
	"go-mythril/laser/smt/z3"
	"math/big"
	"strconv"
)

var nextTransactionId int

func GetNextTransactionId() string {
	nextTransactionId += 1
	return strconv.Itoa(nextTransactionId)
}

type BaseTransaction interface {
	InitialGlobalStateFromEnvironment(worldState *WorldState, environment *Environment, activeFunc string) *GlobalState
	InitialGlobalState() *GlobalState
	GetId() string
	GetGasLimit() int
	End(state *GlobalState, returnData []byte)
	GetCaller() *z3.Bitvec
	GetOrigin() *z3.Bitvec
	GetCalldata() BaseCalldata
	GetCallValue() *z3.Bitvec
	GetWorldState() *WorldState
	GetCalleeAccount() *Account
	GetCode() *disassembler.Disasembly
	GetCtx() *z3.Context
}

type MessageCallTransaction struct {
	WorldState    *WorldState
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

func NewMessageCallTransaction(code string, contractName string, inputStr string, ctx *z3.Context) *MessageCallTransaction {

	//config := z3.NewConfig()
	//ctx := z3.NewContext(config)
	//config.Close()
	//defer ctx.Close()
	txcode := disassembler.NewDisasembly(code)
	calldataList := make([]*z3.Bitvec, 0)

	// IntegerOverflow: symbolicCallData, sload-0, callValue-0
	//inputStr := "1003e2d2000000000000000000000000000000000000000000000000000000000000000a"
	//inputStr := "1003e2d2ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"

	// Origin: callValue-0, sload-addr
	//inputStr := "f2fde38b000000000000000000000000ab8483f64d9c6d1ecf9b849ae677dd3315835cb2"

	// TimeStamp: callValue-1000000000000000000
	//inputStr := "3ccfd60b"

	// Reentrancy: callvalue-0, sload-addr
	//inputStr := "2e1a7d4d0000000000000000000000000000000000000000000000000000000000000001"

	// Lotttopolo: callvalueInstr-symbolic
	//inputStr := "0eecae21"

	// largeTimeStamp: callvalue-10000000000000000000,
	//inputStr := "17c6abfa"

	// whiteBetting: callvalue-10000000000000000000
	//inputStr := "d5214029"

	for i := 0; i < len(inputStr); i = i + 2 {
		val, _ := strconv.ParseInt(inputStr[i:i+2], 16, 10)
		calldataList = append(calldataList, ctx.NewBitvecVal(val, 8))
	}

	//callerStr, _ := new(big.Int).SetString("5B38Da6a701c568545dCfcB03FcB875f56beddC4", 16)
	txId := GetNextTransactionId()
	caller := ctx.NewBitvec("sender_"+txId, 256)
	//caller := ctx.NewBitvecVal(callerStr, 256)
	origin := ctx.NewBitvec("origin", 256)
	//accAddrStr, _ := new(big.Int).SetString("4a9c121080f6d9250fc0143f41b595fd172e31bf", 16)
	//accAddr := ctx.NewBitvecVal(accAddrStr, 256)

	tx := &MessageCallTransaction{
		WorldState: NewWordState(ctx),
		Code:       txcode,
		// TODO: For test here.
		//CalleeAccount: NewAccount(ctx.NewBitvecVal(caller, 256),
		//	ctx.NewArray("balances", 256, 256), false, txcode, contractName),
		CalleeAccount: NewAccount(caller,
			ctx.NewArray("balances", 256, 256), false, txcode, contractName),
		//Caller:   ctx.NewBitvecVal(caller, 256),
		Caller: caller,
		//Calldata: NewConcreteCalldata(txId, calldataList, ctx),
		Calldata:  NewSymbolicCalldata(txId, ctx),
		GasPrice:  10,
		GasLimit:  100000,
		CallValue: 0, // 1 ether
		Origin:    origin,
		Basefee:   ctx.NewBitvecVal(1000, 256),
		Ctx:       ctx,
		Id:        txId,
	}
	// TODO: maybe wrong?
	tx.WorldState.TransactionSequence = append(tx.WorldState.TransactionSequence, tx)
	return tx
}

func (tx *MessageCallTransaction) InitialGlobalStateFromEnvironment(worldState *WorldState, env *Environment, activeFunc string) *GlobalState {
	txStack := make([]BaseTransaction, 0)
	globalState := NewGlobalState(worldState, env, tx.Ctx, append(txStack, tx))
	globalState.Environment.ActiveFuncName = activeFunc

	// Tips: Account is always the same Account in environment and worldState.Accounts.
	// Tips: So the last round the content of Account may be already translated.
	//env.ActiveAccount = env.ActiveAccount.Translate(tx.Ctx)
	//globalState.Translate(tx.Ctx)

	// make sure the value of sender is enough
	sender := env.Sender
	receiver := env.ActiveAccount.Address
	//value := tx.Ctx.NewBitvecVal(env.CallValue, 256)
	value := tx.GetCallValue().Translate(tx.Ctx)
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
	return tx.InitialGlobalStateFromEnvironment(tx.WorldState, environment, "fallback")
}

func (tx *MessageCallTransaction) End(state *GlobalState, data []byte) {
	tx.ReturnData = data
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Catch-TransactionEndSignal")
		}
	}()
	panic("TransactionEndSignal")
}

func (tx *MessageCallTransaction) GetId() string {
	return tx.Id
}

func (tx *MessageCallTransaction) GetCaller() *z3.Bitvec {
	return tx.Caller
}

func (tx *MessageCallTransaction) GetOrigin() *z3.Bitvec {
	return tx.Origin
}

func (tx *MessageCallTransaction) GetCalldata() BaseCalldata {
	return tx.Calldata
}

func (tx *MessageCallTransaction) GetCallValue() *z3.Bitvec {
	//return tx.Ctx.NewBitvecVal(tx.CallValue, 256)
	return tx.Ctx.NewBitvec("call_value"+tx.GetId(), 256)
}

func (tx *MessageCallTransaction) GetWorldState() *WorldState {
	return tx.WorldState
}

func (tx *MessageCallTransaction) GetCalleeAccount() *Account {
	return tx.CalleeAccount
}

func (tx *MessageCallTransaction) GetGasLimit() int {
	return tx.GasLimit
}

func (tx *MessageCallTransaction) GetCode() *disassembler.Disasembly {
	return tx.Code
}

func (tx *MessageCallTransaction) GetCtx() *z3.Context {
	return tx.Ctx
}

type ContractCreationTransaction struct {
	WorldState    *WorldState
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

func NewContractCreationTransaction(code string, contractName string) *ContractCreationTransaction {
	config := z3.NewConfig()
	ctx := z3.NewContext(config)
	txcode := disassembler.NewDisasembly(code)
	calldataList := make([]*z3.Bitvec, 0)
	// TODO: For test here
	// Remix input here
	//
	inputStr := "6080604052600160005534801561001557600080fd5b506101e3806100256000396000f3fe608060405234801561001057600080fd5b50600436106100365760003560e01c80631003e2d21461003b578063b69ef8a814610057575b600080fd5b610055600480360381019061005091906100ab565b610075565b005b61005f610090565b60405161006c91906100e7565b60405180910390f35b806000808282546100869190610102565b9250508190555050565b60005481565b6000813590506100a581610196565b92915050565b6000602082840312156100c1576100c0610191565b5b60006100cf84828501610096565b91505092915050565b6100e181610158565b82525050565b60006020820190506100fc60008301846100d8565b92915050565b600061010d82610158565b915061011883610158565b9250827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0382111561014d5761014c610162565b5b828201905092915050565b6000819050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b600080fd5b61019f81610158565b81146101aa57600080fd5b5056fea264697066735822122087b595ab091f2f063af4481cbf3f5906249a224c0b0e343c050f64821d66d84864736f6c63430008070033"
	for i := 0; i < len(inputStr); i = i + 2 {
		tmp := string(inputStr[i]) + string(inputStr[i+1])
		tmpV, _ := strconv.ParseInt(tmp, 16, 64)
		calldataList = append(calldataList, ctx.NewBitvecVal(tmpV, 8))
	}
	// set your callerValue in remix to test
	// For test
	caller, _ := new(big.Int).SetString("5B38Da6a701c568545dCfcB03FcB875f56beddC4", 16)
	tx := &ContractCreationTransaction{
		WorldState: NewWordState(ctx),
		Code:       txcode,
		// TODO: For test here.
		CalleeAccount: NewAccount(ctx.NewBitvecVal(123, 256),
			ctx.NewArray("balances", 256, 256), true, txcode, contractName),
		Caller:    ctx.NewBitvecVal(caller, 256),
		Calldata:  NewConcreteCalldata("txid123", calldataList, ctx),
		GasPrice:  10,
		GasLimit:  100000,
		CallValue: 0,
		Origin:    ctx.NewBitvecVal(caller, 256),
		Basefee:   ctx.NewBitvecVal(1000, 256),
		Ctx:       ctx,
		Id:        GetNextTransactionId(),
	}
	// TODO: maybe wrong?
	tx.WorldState.TransactionSequence = append(tx.WorldState.TransactionSequence, tx)
	return tx
}

func (tx *ContractCreationTransaction) InitialGlobalStateFromEnvironment(worldState *WorldState, env *Environment, activeFunc string) *GlobalState {
	txStack := make([]BaseTransaction, 0)
	globalState := NewGlobalState(worldState, env, tx.Ctx, append(txStack, tx))
	globalState.Environment.ActiveFuncName = activeFunc

	sender := env.Sender
	receiver := env.ActiveAccount.Address
	//value := tx.Ctx.NewBitvecVal(env.CallValue, 256)
	value := tx.GetCallValue()
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
	return tx.InitialGlobalStateFromEnvironment(tx.WorldState, environment, "constructor")
}

func (tx *ContractCreationTransaction) End(globalState *GlobalState, data []byte) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Catch-TransactionEndSignal")
		}
	}()

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

func (tx *ContractCreationTransaction) GetCaller() *z3.Bitvec {
	return tx.Caller
}

func (tx *ContractCreationTransaction) GetOrigin() *z3.Bitvec {
	return tx.Origin
}

func (tx *ContractCreationTransaction) GetCalldata() BaseCalldata {
	return tx.Calldata
}

func (tx *ContractCreationTransaction) GetCallValue() *z3.Bitvec {
	//return tx.Ctx.NewBitvecVal(tx.CallValue, 256)
	return tx.Ctx.NewBitvec("call_value"+tx.GetId(), 256)
}

func (tx *ContractCreationTransaction) GetWorldState() *WorldState {
	return tx.WorldState
}

func (tx *ContractCreationTransaction) GetCalleeAccount() *Account {
	return tx.CalleeAccount
}

func (tx *ContractCreationTransaction) GetGasLimit() int {
	return tx.GasLimit
}

func (tx *ContractCreationTransaction) GetPreWorldState() *WorldState {
	// TODO: Deepcopy ?
	//return tx.WorldState.Copy()
	return nil
}

func (tx *ContractCreationTransaction) GetCode() *disassembler.Disasembly {
	return tx.Code
}

func (tx *ContractCreationTransaction) GetCtx() *z3.Context {
	return tx.Ctx
}
