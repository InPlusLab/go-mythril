package state

import (
	"go-mythril/disassembler"
	"go-mythril/laser/smt/z3"
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
	return &WorldState{
		Accounts:            accounts,
		Balances:            ctx.NewArray("balance", 256, 256),
		StartingBalances:    ctx.NewArray("balance", 256, 256),
		Constraints:         NewConstraints(),
		TransactionSequence: make([]BaseTransaction, 0),
	}
}

func (ws *WorldState) Copy() *WorldState {
	var tmp *WorldState
	for _, acc := range ws.Accounts {
		tmp.PutAccount(acc.Copy())
	}
	tmp.Balances = ws.Balances
	tmp.Constraints = ws.Constraints.Copy()

	return tmp
}

func (ws *WorldState) AccountsExistOrLoad(addr *z3.Bitvec) *Account {
	accounts := ws.Accounts
	acc, ok := accounts[addr.Value()]
	if ok {
		return acc
	} else {
		// TODO: find in dynamicLoader
		return NewAccount(addr, nil, false, disassembler.NewDisasembly(""))
	}
}

func (ws *WorldState) PutAccount(acc *Account) {
	accounts := ws.Accounts
	accounts[acc.Address.Value()] = acc
	acc.Balances = ws.Balances
}
