package state

import (
	"go-mythril/disassembler"
	"go-mythril/laser/smt/z3"
	"math/big"
)

type WorldState struct {
	// the key of Accounts is decimal
	Accounts            map[string]*Account
	Balances            *z3.Array
	StartingBalances    *z3.Array
	Constraints         *Constraints
	TransactionSequence []BaseTransaction
}

func NewWordState(ctx *z3.Context) *WorldState {
	accounts := make(map[string]*Account)
	// TODO: just test for balance_()
	caller, _ := new(big.Int).SetString("5B38Da6a701c568545dCfcB03FcB875f56beddC4", 16)
	balances := ctx.NewArray("balance", 256, 256)
	balances.SetItem(ctx.NewBitvecVal(caller, 256), ctx.NewBitvecVal(1, 256))
	return &WorldState{
		Accounts: accounts,
		//Balances:            ctx.NewArray("balance", 256, 256),
		Balances:            balances,
		StartingBalances:    ctx.NewArray("balance", 256, 256),
		Constraints:         NewConstraints(),
		TransactionSequence: make([]BaseTransaction, 0),
	}
}

func (ws *WorldState) Copy() *WorldState {
	//var tmp *WorldState
	//for _, acc := range ws.Accounts {
	//	tmp.PutAccount(acc.Copy())
	//}
	//tmp.Balances = ws.Balances
	//tmp.Constraints = ws.Constraints.Copy()
	//
	//return tmp

	return &WorldState{
		Accounts:            ws.Accounts,
		Balances:            ws.Balances,
		StartingBalances:    ws.StartingBalances,
		Constraints:         ws.Constraints.DeepCopy(),
		TransactionSequence: ws.TransactionSequence,
	}
}

func (ws *WorldState) AccountsExistOrLoad(addr *z3.Bitvec) *Account {
	accounts := ws.Accounts
	acc, ok := accounts[addr.Value()]
	if ok {
		return acc
	} else {
		// TODO: find in dynamicLoader
		return NewAccount(addr, ws.Balances, false, disassembler.NewDisasembly(""), "")
	}
}

func (ws *WorldState) PutAccount(acc *Account) {
	accounts := ws.Accounts
	accounts[acc.Address.Value()] = acc
	acc.Balances = ws.Balances
}

func (ws *WorldState) Translate(ctx *z3.Context) *WorldState {
	newConstraints := NewConstraints()
	for _, v := range ws.Constraints.ConstraintList {
		newV := v.Translate(ctx)
		newConstraints.Add(newV)
	}
	newAccouts := make(map[string]*Account)
	for i, v := range ws.Accounts {
		newAccouts[i] = v.Translate(ctx)
	}

	return &WorldState{
		Accounts:            newAccouts,
		Balances:            ws.Balances.Translate(ctx).(*z3.Array),
		StartingBalances:    ws.StartingBalances.Translate(ctx).(*z3.Array),
		Constraints:         newConstraints,
		TransactionSequence: ws.TransactionSequence,
	}
}
