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

// In python Mythril, constrains List'copy is shallow copy.
func (c *Constraints) Copy() *Constraints {
	tmp := NewConstraints()
	for _, item := range c.ConstraintList {
		tmp.Add(item)
	}
	return tmp
}

func (c *Constraints) IsPossible() bool {
	// TODO: z3 solve
	return true
}

func (c *Constraints) Add(constraints ...*z3.Bool) bool {
	for _, constraint := range constraints {
		c.ConstraintList = append(c.ConstraintList, constraint)
	}
	return true
}

func (c *Constraints) Length() int{
	return len(c.ConstraintList)
}