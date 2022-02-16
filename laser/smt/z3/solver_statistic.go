package z3

// #include <stdlib.h>
// #include "goZ3Config.h"
import "C"
import "time"

// SolverStatistic is used to measure statistics
// for smt query check function
// Corresponding to the solver_statistic.py in Mythril
type SolverStatistic struct {
	QueryCount int
	SolverTime time.Duration
}

// Singleton
var solverStatistic SolverStatistic

func GetSolverStatistic() *SolverStatistic {
	return &solverStatistic
}

// CheckWrapper is used to store the statistic data of Check
// including queryCount and solverTime
func (s *Solver) CheckWrapper() LBool {
	start := time.Now()
	result := s.Check()
	end := time.Now()

	instance := GetSolverStatistic()
	instance.QueryCount++
	instance.SolverTime += end.Sub(start)

	return result
}
func (s *Optimize) CheckWrapper() LBool {
	start := time.Now()
	result := s.Check()
	end := time.Now()

	instance := GetSolverStatistic()
	instance.QueryCount++
	instance.SolverTime += end.Sub(start)

	return result
}
