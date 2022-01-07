package state

import (
	"go-mythril/laser/smt"
)

type WorldState struct {
	Balances    *smt.Array
	Constraints *Constraints
}

func NewWordState() *WorldState {
	return &WorldState{
		Balances:    &smt.Array{},
		Constraints: NewConstraints(),
	}
}
