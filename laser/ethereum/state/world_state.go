package state

import (
	"go-mythril/laser/smt/z3"
	"strconv"
)

type WorldState struct {
	Accounts            *map[int64]*Account
	Balances            *z3.Array
	StartingBalances    *z3.Array
	Constraints         *Constraints
	TransactionSequence []BaseTransaction
}

func NewWordState(ctx *z3.Context) *WorldState {
	accounts := make(map[int64]*Account)
	return &WorldState{
		Accounts:            &accounts,
		Balances:            ctx.NewArray("balance", 256, 256),
		StartingBalances:    ctx.NewArray("balance", 256, 256),
		Constraints:         NewConstraints(),
		TransactionSequence: make([]BaseTransaction, 0),
	}
}

func (ws *WorldState) Copy() *WorldState {
	var tmp *WorldState
	for _, acc := range *ws.Accounts {
		tmp.PutAccount(acc.Copy())
	}
	tmp.Balances = ws.Balances
	tmp.Constraints = ws.Constraints.Copy()

	return tmp
}

func (ws *WorldState) AccountsExistOrLoad(bvValue string) *Account {
	// Big int here?
	value, _ := strconv.ParseInt(bvValue, 10, 64)
	accounts := *ws.Accounts
	acc, ok := accounts[value]
	if ok {
		return acc
	} else {
		// TODO: find in dynamicLoader
		return nil
	}
}

func (ws *WorldState) PutAccount(acc *Account) {
	addrV, _ := strconv.ParseInt(acc.Address.Value(), 10, 64)
	accounts := *ws.Accounts
	accounts[addrV] = acc
	acc.Balances = ws.Balances
}
