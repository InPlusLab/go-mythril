package analysis

import (
	"go-mythril/laser/ethereum/state"
	"go-mythril/laser/smt/z3"
	"go-mythril/utils"
	"math/big"
)

//func GetTransactionSequence(globalState *state.GlobalState, constraints *state.Constraints) {
//	transactionSequence := globalState.WorldState.TransactionSequence
//	concreteTransactions := make([]state.BaseTransaction, 0)
//	txConstraints, minimize := _set_minimisation_constraints(transactionSequence, constraints.Copy(),
//		make([]*z3.Bitvec, 0), 5000, globalState.WorldState, globalState.Z3ctx)
//	model, ok := support.GetModel(txConstraints, minimize, make([]*z3.Bool, 0), true, globalState.Z3ctx)
//	if !ok {
//		// UnsatError
//		return
//	}
//	initialWorldState := transactionSequence[0].GetWorldState()
//	initialAccounts := initialWorldState.Accounts
//	// TODO: _get_concrete_transaction
//	//for _, tx := range transactionSequence{
//	//	concreteTx :=
//	//}
//	minPriceDict := make(map[string]int)
//
//}

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
	for _, account := range *worldState.Accounts {
		// Lazy way to prevent overflows and to ensure "reasonable" balances
		// Each account starts with less than 100 ETH
		constraints.Add(ctx.NewBitvecVal(tmpValue, 256).BvUGe(worldState.StartingBalances.GetItem(account.Address)))
	}
	return constraints, minimize
}
