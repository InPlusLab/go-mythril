package ethereum

import (
	"fmt"
	"go-mythril/disassembler"
	"go-mythril/laser/ethereum/state"
	"go-mythril/laser/smt/z3"
	"strconv"
)

const SYMBOLIC_CALLDATA_SIZE = 320
const GAS_CALLSTIPEND = 2300

// default: withValue=false
func GetCallParameters(globalState *state.GlobalState, withValue bool) (*z3.Bitvec, *state.Account, state.BaseCalldata, *z3.Bitvec, *z3.Bitvec, *z3.Bitvec, *z3.Bitvec) {
	gas := globalState.Mstate.Stack.Pop()
	to := globalState.Mstate.Stack.Pop()
	var value *z3.Bitvec
	if withValue {
		value = globalState.Mstate.Stack.Pop()
	} else {
		value = globalState.Z3ctx.NewBitvecVal(0, 256)
	}
	memoryInputOffset := globalState.Mstate.Stack.Pop()
	memoryInputSize := globalState.Mstate.Stack.Pop()
	memoryOutOffset := globalState.Mstate.Stack.Pop()
	memoryOutSize := globalState.Mstate.Stack.Pop()

	calleeAddress := getCalleeAddress(globalState, to)
	calleeAccount := getCalleeAccount(globalState, calleeAddress)
	fmt.Println("beforeGetCallData", memoryInputOffset, " ", memoryInputSize)
	callData := GetCallData(globalState, memoryInputOffset, memoryInputSize)

	ctx := globalState.Z3ctx
	gas = gas.BvAdd(z3.If(value.BvSGt(ctx.NewBitvecVal(0, 256)), ctx.NewBitvecVal(GAS_CALLSTIPEND, gas.BvSize()), ctx.NewBitvecVal(0, gas.BvSize())))

	return calleeAddress, calleeAccount, callData, value, gas, memoryOutOffset, memoryOutSize
}

func getCalleeAccount(globalState *state.GlobalState, calleeAddress *z3.Bitvec) *state.Account {
	if calleeAddress.Symbolic() {
		return state.NewAccount(calleeAddress, globalState.WorldState.Balances, false, disassembler.NewDisasembly(""))
	} else {
		return globalState.WorldState.AccountsExistOrLoad(calleeAddress)
	}
}

func getCalleeAddress(globalState *state.GlobalState, symbolicToAddress *z3.Bitvec) *z3.Bitvec {
	environment := globalState.Environment
	fmt.Println("getCalleeAddress", environment)
	// TODO: symbolicAddress situation
	return symbolicToAddress
}

func GetCallData(globalState *state.GlobalState, memStart *z3.Bitvec, memSize *z3.Bitvec) state.BaseCalldata {

	mstate := globalState.Mstate
	txId := globalState.CurrentTransaction().GetId() + "_internalcall"

	if memStart.Symbolic() {
		fmt.Println("Unsupported symbolic memory offset")
		return state.NewSymbolicCalldata(txId, globalState.Z3ctx)
	}
	memStartV, _ := strconv.ParseInt(memStart.Value(), 10, 64)
	var memSizeV int64
	if memSize.Symbolic() {
		memSizeV = SYMBOLIC_CALLDATA_SIZE
	} else {
		value, _ := strconv.ParseInt(memSize.Value(), 10, 64)
		memSizeV = value
	}
	callDataFromMem := mstate.Memory.GetItems(memStartV, memStartV+memSizeV)
	fmt.Println(memStartV, memSizeV)
	fmt.Println("calldataFromMem:", callDataFromMem, " ", len(callDataFromMem))
	return state.NewConcreteCalldata(txId, callDataFromMem, globalState.Z3ctx)
}

//func NativeCall(globalState *state.GlobalState, calleeAddress *z3.Bitvec, callData state.BaseCalldata, memoryOutOffset *z3.Bitvec, memoryOutSzie *z3.Bitvec) []*state.GlobalState{
//
//}
