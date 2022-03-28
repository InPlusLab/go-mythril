package ethereum

import (
	"fmt"
	"go-mythril/laser/ethereum/state"
	"go-mythril/laser/smt/z3"
	"strconv"
)

const SYMBOLIC_CALLDATA_SIZE = 320

func GetCallData(globalState *state.GlobalState, memStart *z3.Bitvec,
	memSize *z3.Bitvec) state.BaseCalldata {

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
	return state.NewConcreteCalldata(txId, callDataFromMem)
}
