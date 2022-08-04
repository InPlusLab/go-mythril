package analysis

import (
	"encoding/hex"
	"fmt"
	"go-mythril/disassembler"
	"go-mythril/laser/ethereum/function_managers"
	"go-mythril/laser/ethereum/state"
	"go-mythril/laser/smt/z3"
	"go-mythril/utils"
	"math/big"
	"strconv"
	"strings"
)

func GetTransactionSequence(globalState *state.GlobalState, constraints *state.Constraints) map[string]interface{} {
	transactionSequence := globalState.WorldState.TransactionSequence
	concreteTransactions := make([]*map[string]string, 0)
	txConstraints, minimize := _set_minimisation_constraints(transactionSequence, constraints.Copy(),
		make([]*z3.Bool, 0), 5000, globalState.WorldState, globalState.Z3ctx)
	model, ok := state.GetModel(txConstraints, minimize, make([]*z3.Bool, 0), false, globalState.Z3ctx)
	if !ok {
		// UnsatError
		fmt.Println("unsat in getTxSeq")
		return nil
	}

	initialWorldState := transactionSequence[0].GetWorldState()
	initialAccounts := initialWorldState.Accounts
	for _, transaction := range transactionSequence {
		concreteTx := _get_concrete_transaction(model, transaction)
		concreteTransactions = append(concreteTransactions, concreteTx)
	}
	minPriceDict := make(map[string]int)
	ctx := globalState.Z3ctx

	for address, _ := range initialAccounts {
		minPriceDict[address] = model.Eval(
			initialWorldState.StartingBalances.GetItem(ctx.NewBitvecVal(address, 256)).AsAST(), true).Int()
	}
	concreteInitialState := _get_concrete_state(initialAccounts, &minPriceDict)
	switch transactionSequence[0].(type) {
	case *state.ContractCreationTransaction:
		code := transactionSequence[0].(*state.ContractCreationTransaction).Code
		_replace_with_actual_sha(concreteTransactions, model, code, globalState.Z3ctx)
	default:
		_replace_with_actual_sha(concreteTransactions, model, nil, globalState.Z3ctx)
	}
	_add_calldata_placeholder(concreteTransactions, transactionSequence)

	steps := make(map[string]interface{})
	steps["initialState"] = concreteInitialState
	steps["steps"] = concreteTransactions
	return steps
}

func _get_concrete_state(initialAccounts map[string]*state.Account, minPriceDict *map[string]int) *map[string]*map[string]string {
	accounts := make(map[string]*map[string]string)
	for address, _ := range initialAccounts {
		data := make(map[string]string)
		// data["nonce"] = strconv.Itoa( account.Nonce )
		// data["code"] = string(account.Code.Bytecode)
		// TODO: storage 2 str ?
		// data["storage"] = account.Storage
		// data["balance"] = hex(minPriceDict["address"])
		accounts[address] = &data
	}
	return &accounts
}

func _get_concrete_transaction(model *z3.Model, transacton state.BaseTransaction) *map[string]string {
	// Get concrete values from transaction
	// String() return #x123, not 0x123
	address := transacton.GetCalleeAccount().Address.String()
	value := model.Eval(transacton.GetCallValue().AsAST(), true).String()
	caller := model.Eval(transacton.GetCaller().AsAST(), true).String()
	caller = "#x" + utils.Zfill(caller[2:], 40)
	input := ""
	switch transacton.(type) {
	case *state.ContractCreationTransaction:
		address = ""
		input += "bytecode"
	}
	concreteTx := make(map[string]string)
	concreteTx["input"] = input
	concreteTx["value"] = value
	concreteTx["origin"] = caller
	concreteTx["address"] = address

	return &concreteTx
}

func _set_minimisation_constraints(txSeq []state.BaseTransaction, constraints *state.Constraints,
	minimize []*z3.Bool, maxSize int, worldState *state.WorldState, ctx *z3.Context) (*state.Constraints, []*z3.Bool) {
	tmpValue, _ := new(big.Int).SetString("1000000000000000000000", 10)
	for _, tx := range txSeq {
		// Set upper bound on calldata size
		maxCalldataSize := ctx.NewBitvecVal(maxSize, 256)
		constraints.Add(maxCalldataSize.BvUGe(tx.GetCalldata().Calldatasize()))
		// Minimize
		minimize = append(minimize, tx.GetCalldata().Calldatasize().AsBool())
		minimize = append(minimize, tx.GetCallValue().AsBool())

		constraints.Add(ctx.NewBitvecVal(tmpValue, 256).BvUGe(worldState.StartingBalances.GetItem(tx.GetCaller())))
	}
	for _, account := range worldState.Accounts {
		// Lazy way to prevent overflows and to ensure "reasonable" balances
		// Each account starts with less than 100 ETH
		constraints.Add(ctx.NewBitvecVal(tmpValue, 256).BvUGe(worldState.StartingBalances.GetItem(account.Address)))
	}
	return constraints, minimize
}

func _replace_with_actual_sha(concreteTransactions []*map[string]string, model *z3.Model, code *disassembler.Disasembly, ctx *z3.Context) {
	keccakFunctionManager := function_managers.NewKeccakFunctionManager(ctx)
	concreteHashes := keccakFunctionManager.GetConcreteHashData(model)
	for _, tx := range concreteTransactions {
		realTx := *tx
		if !(strings.Contains(realTx["input"], keccakFunctionManager.HashMatcher)) {
			continue
		}
		var sIndex int
		codeStr := hex.EncodeToString(code.Bytecode)
		if code != nil && strings.Contains(realTx["input"], codeStr) {
			// TODO: len(codeStr) or len(code.Bytecode) here ?
			sIndex = len(codeStr) + 2
		} else {
			sIndex = 10
		}
		for i := sIndex; i < len(realTx["input"]); i++ {
			dataSlice := realTx["input"][i : i+64]
			if !(strings.Contains(dataSlice, keccakFunctionManager.HashMatcher)) || len(dataSlice) != 64 {
				continue
			}
			dataSliceInt, _ := strconv.ParseInt(dataSlice, 16, 10)
			findInput := ctx.NewBitvecVal(dataSliceInt, 256)
			var input_ *z3.Bitvec
			concreteHashesContent := *concreteHashes
			for size, _ := range concreteHashesContent {
				// TODO: keccakFunctionManager.storeFunction
				findInputValue, _ := strconv.Atoi(findInput.Value())
				findInputInFlag := false
				for _, item := range concreteHashesContent[size] {
					if findInputValue == item {
						findInputInFlag = true
					}
				}
				if !findInputInFlag {
					continue
				}
				// TODO: keccakFunctionManager.storeFunction
				input_ = ctx.NewBitvecVal(model.Eval(findInput.AsAST(), false), size)
			}
			if input_ == nil {
				continue
			}
			keccak := keccakFunctionManager.FindConcreteKeccak(input_)
			keccakValue, _ := strconv.Atoi(keccak.Value())
			hexKeccak := utils.ToHexStr(keccakValue)
			if len(hexKeccak) != 64 {
				hexKeccak = utils.Zfill(hexKeccak, 64)
			}
			realTx["input"] = realTx["intput"][:sIndex] + strings.Replace(realTx["intput"][sIndex:], realTx["input"][i:64+i], hexKeccak, -1)
		}
	}
}

func _add_calldata_placeholder(concreteTransactions []*map[string]string, transactionSequence []state.BaseTransaction) {
	for _, tx := range concreteTransactions {
		realTx := *tx
		realTx["calldata"] = realTx["input"]
	}
	switch transactionSequence[0].(type) {
	case *state.MessageCallTransaction:
		return
	}
	codeLen := len(transactionSequence[0].GetCode().Bytecode)
	realTx0 := *concreteTransactions[0]
	realTx0["calldata"] = realTx0["input"][codeLen+2:]
}
