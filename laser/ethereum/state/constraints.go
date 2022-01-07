package state

import (
	"go-mythril/laser/smt"
)

type Constraints struct {
	ConstraintList []*smt.Bool
}

func NewConstraints() *Constraints {
	return &Constraints{
		ConstraintList: make([]*smt.Bool, 0),
	}
}

func (c *Constraints) IsPossible() bool {
	// TODO: z3 solve
	return true
}

func (c *Constraints) Add(constraint *smt.Bool) bool {
	c.ConstraintList = append(c.ConstraintList, constraint)
	return true
}
