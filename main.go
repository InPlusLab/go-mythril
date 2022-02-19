package main

import (
	"fmt"
	"go-mythril/laser/ethereum"
	"go-mythril/laser/smt/z3"
)

func main() {
	fmt.Println("go mythril")
	evm := ethereum.NewLaserEVM(1, 1, 1)
	evm.SymExec("6060320032")
	return

	fmt.Println("go mythril-testForGoz3")
	config := z3.NewConfig()
	ctx := z3.NewContext(config)
	config.Close()
	defer ctx.Close()

	s := ctx.NewSolver()
	defer s.Close()
	s1 := ctx.NewSolver()
	defer s1.Close()

	x := ctx.Const(ctx.Symbol("X"), ctx.IntSort())
	ten := ctx.Int(10, ctx.IntSort())
	five := ctx.Int(5, ctx.IntSort())
	zero := ctx.Int(0, ctx.IntSort())
	s.Assert(x.Ge(zero), x.Lt(five))
	s1.Assert(x.Ge(five), x.Lt(ten))

	if v := s.Check(); v != z3.True {
		fmt.Println("Unsolveable")
		return
	}

	if v := s1.Check(); v != z3.True {
		fmt.Println("Unsolveable")
		return
	}

	m := s.Model()
	answer := m.Assignments()
	fmt.Println("solver0's result:")
	fmt.Println(answer)

	m1 := s1.Model()
	answer1 := m1.Assignments()
	fmt.Println("solver1's result:")
	fmt.Println(answer1)
}
