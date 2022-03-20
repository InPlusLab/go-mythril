package state

import (
	"go-mythril/laser/smt"
	"strconv"
)

type WorldState struct {
	Accounts    *map[int64]*Account
	Balances    *smt.Array
	Constraints *Constraints
}

func NewWordState() *WorldState {
	accounts := make(map[int64]*Account)
	return &WorldState{
		Accounts:    &accounts,
		Balances:    &smt.Array{},
		Constraints: NewConstraints(),
	}
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
