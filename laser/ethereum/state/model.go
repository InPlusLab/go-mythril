package state

import (
	"fmt"
	"go-mythril/laser/smt/z3"
)

// TODO: implementation of LRU cache
// default: enforceExecutionTime = true, minimize=maximize=[]
func GetModel(constraints *Constraints, minimize []*z3.Bool, maximize []*z3.Bool,
	enforceExecutionTime bool, ctx *z3.Context) (*z3.Model, bool) {

	//fmt.Println("Constraints:")
	//for i, constraint := range constraints.ConstraintList {
	//	fmt.Println(i, "-", constraint.BoolString())
	//}
	//fmt.Println("Minimize:")
	//for i, e := range minimize {
	//	fmt.Println(i, "-", e.BoolString())
	//}
	//for i, e := range maximize {
	//	fmt.Println(i, "-", e.BoolString())
	//}

	s := ctx.NewOptimize()
	//s := ctx.NewSolver()
	defer s.Close()
	//timeout := support.NewArgs().SolverTimeout
	timeout := 20000000
	//timeout := 100000000000
	//timeout := 30000
	//if enforceExecutionTime {
	//	// GetTimeHandlerInstance().TimeRemaining()-500
	//	//timeout = min(timeout, GetTimeHandlerInstance().TimeRemaining()-500)
	//	if timeout <= 0 {
	//		//fmt.Println("timeout")
	//		return nil, false
	//	}
	//}
	//s.SetTimeout(timeout)
	s.RLimit(timeout)
	//for _, constraint := range constraints.ConstraintList {
	//	if constraint == nil {
	//		fmt.Println("constraint nil")
	//		return nil, false
	//	}
	//}
	for _, constraint := range constraints.ConstraintList {
		// TODO: constraint == nil
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
		//return s.Model(), true
		return nil, true
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
