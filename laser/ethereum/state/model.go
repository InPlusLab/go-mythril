package state

import (
	"fmt"
	"go-mythril/laser/smt/z3"
)

// TODO: implementation of LRU cache
// default: enforceExecutionTime = true, minimize=maximize=[]
func GetModel(constraints *Constraints, minimize []*z3.Bool, maximize []*z3.Bool,
	enforceExecutionTime bool, ctx *z3.Context) (*z3.Model, bool) {
	s := ctx.NewOptimize()
	defer s.Close()
	//timeout := support.NewArgs().SolverTimeout
	//timeout := 30000
	//timeout := 1000000000
	timeout := 1000000000
	if enforceExecutionTime {
		// GetTimeHandlerInstance().TimeRemaining()-500
		//timeout = min(timeout, GetTimeHandlerInstance().TimeRemaining()-500)
		if timeout <= 0 {
			fmt.Println("timeout")
			return nil, false
		}
	}
	//s.SetTimeout(timeout)
	s.RLimit(timeout)
	for _, constraint := range constraints.ConstraintList {
		if constraint == nil {
			fmt.Println("constraint nil")
			return nil, false
		}
	}
	for _, constraint := range constraints.ConstraintList {
		// TODO: constraint == nil
		s.Assert(constraint.AsAST())
		//if constraint.Annotations != nil {
		//	s.Assert(constraint.AsAST())
		//}
	}

	for _, e := range minimize {
		s.Minimize(e.AsAST())
	}
	for _, e := range maximize {
		s.Maximize(e.AsAST())
	}
	// TODO: args.solverLog
	fmt.Println("beforeCheck")
	result := s.Check()
	fmt.Println("afterCheck")
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
