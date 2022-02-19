package state

import "go-mythril/laser/smt/z3"

const STACK_LIMIT = 1024

// TODO
type MachineStack []*z3.Bitvec

type MachineState struct {
	// Important: PC in mythril seems to be different with others (e.g. Etherscan). Mythril.address = Others.pc. Mythril.pc seems to be the index of the evminstruction.
	Pc    int
	Stack []*z3.Bitvec
}

func NewMachineState() *MachineState {
	stack := make([]*z3.Bitvec, 0)
	return &MachineState{
		Pc:    0,
		Stack: stack,
	}
}
