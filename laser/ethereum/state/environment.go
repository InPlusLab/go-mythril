package state

import (
	"go-mythril/disassembler"
	"go-mythril/laser/smt/z3"
)

type Environment struct {
	Code *disassembler.Disasembly
	//TODO z3.ExprRef
	Stack []*z3.Bitvec
}

func NewEnvironment(code *disassembler.Disasembly) *Environment {
	stack := make([]*z3.Bitvec, 0)
	return &Environment{
		Code:  code,
		Stack: stack,
	}
}
