package state

import (
	"go-mythril/laser/smt"
)

type WorldState struct {
	Balances *smt.Array
}

func NewWordState() *WorldState {
	return &WorldState{
		Balances: &smt.Array{},
	}
}
