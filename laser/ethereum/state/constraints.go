package state

import (
	"fmt"
	"go-mythril/laser/smt/z3"
	"strings"
)

type Constraints struct {
	ConstraintList []*z3.Bool
}

func NewConstraints() *Constraints {
	return &Constraints{
		ConstraintList: make([]*z3.Bool, 0),
	}
}

func (c *Constraints) Translate(ctx *z3.Context) *Constraints {
	newCons := NewConstraints()
	for _, v := range c.ConstraintList {
		newV := v.Translate(ctx)
		newCons.ConstraintList = append(newCons.ConstraintList, newV)
	}
	return newCons
}

// In python Mythril, constrains List'copy is shallow copy.
func (c *Constraints) Copy() *Constraints {
	tmp := NewConstraints()
	for _, item := range c.ConstraintList {
		tmp.Add(item)
	}
	return tmp
}

func (c *Constraints) DeepCopy() *Constraints {
	tmp := NewConstraints()
	for _, item := range c.ConstraintList {
		tmp.Add(item.Copy())
	}
	return tmp
}

func (c *Constraints) IsPossible() bool {
	if len(c.ConstraintList) <= 0 {
		fmt.Println("In isPossible: empty Constraints")
		return false
	} else {
		fmt.Println("In ws.Constraints.isPossible")
		ctx := z3.GetBoolCtx(c.ConstraintList[0])
		fmt.Println("After getCtx")
		_, ok := GetModel(c, make([]*z3.Bool, 0), make([]*z3.Bool, 0), false, ctx)
		return ok
	}
}

func (c *Constraints) IsPossibleRlimit() (bool, int) {
	if len(c.ConstraintList) <= 0 {
		fmt.Println("In isPossible: empty Constraints")
		return false, 0
	} else {
		// fmt.Println("In ws.Constraints.isPossible")
		ctx := z3.GetBoolCtx(c.ConstraintList[0])
		// fmt.Println("After getCtx")
		rlimit, ok := GetModelRlimit(c, make([]*z3.Bool, 0), make([]*z3.Bool, 0), false, ctx)
		return ok, rlimit
	}
}

func (c *Constraints) Add(constraints ...*z3.Bool) bool {
	//fmt.Println("addCons!")
	for _, constraint := range constraints {
		//fmt.Println("addConstraints:", constraint.AsAST().String())
		//fmt.Println(constraint.BoolString())
		c.ConstraintList = append(c.ConstraintList, constraint)
	}
	return true
}

func (c *Constraints) Length() int {
	return len(c.ConstraintList)
}

func (c *Constraints) PrintOneLine() {
	if c.Length() == 0 {
		fmt.Println("Cons: null")
	}

	for i := 0; i < c.Length(); i++ {
		str := c.ConstraintList[i].BoolString()
		idx := strings.Index(str, "\n")
		if idx == -1 {
			fmt.Println(i, "-", str)
		} else {
			fmt.Println(i, "-", str[:idx])
		}
	}
}
