package state

import (
	"go-mythril/laser/smt/z3"
)

type Constraints struct {
	ConstraintList []*z3.Bool
}

func NewConstraints() *Constraints {
	return &Constraints{
		ConstraintList: make([]*z3.Bool, 0),
	}
}

func (c *Constraints) IsPossible() bool {
	// TODO: z3 solve
	return true
}

func (c *Constraints) Add(constraint *z3.Bool) bool {
	c.ConstraintList = append(c.ConstraintList, constraint)
	return true
}
