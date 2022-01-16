package main

import (
	"fmt"
	"go-mythril/laser/smt/z3"
)

func main() {
	/*	fmt.Println("go mythril")
		evm := ethereum.NewLaserEVM(1, 1, 1)
		evm.SymExec("0x6060")*/
	fmt.Println("go mythril-testForGoz3")
	config := z3.NewConfig()
	config.Close()
	ctx := z3.NewContext(config)
	defer ctx.Close()
	s := ctx.NewSolver()
	defer s.Close()

	x := ctx.Const(ctx.Symbol("x"), ctx.BvSort(16))
	y := ctx.Const(ctx.Symbol("y"), ctx.BvSort(16))
	s.Assert(x.BvAdd(y).Eq(ctx.Int(2, ctx.BvSort(16))))
	s.Assert(x.BvULt(ctx.Int(10, ctx.BvSort(16))))
	s.Assert(y.BvULt(ctx.Int(10, ctx.BvSort(16))))

	if v := s.Check(); v != z3.True {
		fmt.Println("Unsolveable")
		return
	}

	m := s.Model()
	assignments := m.Assignments()
	m.Close()

	fmt.Println(assignments)
}
