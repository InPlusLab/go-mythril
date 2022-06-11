package ethereum

import (
	"go-mythril/laser/ethereum/state"
	"go-mythril/laser/smt/z3"
)

type JumpType int

const (
	CONDITIONAL   JumpType = 1
	UNCONDITIONAL JumpType = 2
	CALL          JumpType = 3
	RETURN        JumpType = 4
	Transaction   JumpType = 5
)

type Node struct {
	// default
	ContractName string
	FunctionName string
	StartAddr    int
	States       []*state.GlobalState
	Constraints  *state.Constraints
}

type Edge struct {
	NodeFrom  int
	NodeTo    int
	EdgeType  JumpType
	Condition *z3.Bool
}
