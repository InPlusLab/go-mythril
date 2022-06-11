package state

import (
	"fmt"
	"go-mythril/laser/smt/z3"
	"go-mythril/support"
)

// TODO: implementation of LRU cache
// default: enforceExecutionTime = true, minimize=maximize=[]
func GetModel(constraints *Constraints, minimize []*z3.Bool, maximize []*z3.Bool,
	enforceExecutionTime bool, ctx *z3.Context) (*z3.Model, bool) {
	s := ctx.NewOptimize()
	timeout := support.NewArgs().SolverTimeout
	if enforceExecutionTime {
		// GetTimeHandlerInstance().TimeRemaining()-500
		timeout = min(timeout, GetTimeHandlerInstance().TimeRemaining()-500)
		if timeout <= 0 {
			return nil, false
		}
	}
	s.SetTimeout(timeout)
	for _, constraint := range constraints.ConstraintList {
		if constraint == nil {
			return nil, false
		}
	}

	for _, constraint := range constraints.ConstraintList {
		s.Assert(constraint.AsAST())
	}
	for _, e := range minimize {
		s.Minimize(e.AsAST())
	}
	for _, e := range maximize {
		s.Maximize(e.AsAST())
	}
	// TODO: args.solverLog
	result := s.Check()
	if result == z3.True {
		return s.Model(), true
	} else {
		fmt.Println("Timeout/Error encountered while solving expression using z3")
		return nil, false
	}
}
func min(a int, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

