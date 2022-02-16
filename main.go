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
	ctx := z3.NewContext(config)
	config.Close()
	defer ctx.Close()

	s := ctx.NewOptimize()
	defer s.Close()

	x := ctx.Const(ctx.Symbol("X"), ctx.IntSort())
	y := ctx.Const(ctx.Symbol("Y"), ctx.IntSort())
	ten := ctx.Int(10, ctx.IntSort())
	zero := ctx.Int(0, ctx.IntSort())
	s.Assert(x.Ge(zero), y.Ge(zero), x.Add(y).Eq(ten))

	s.Maximize(y)

	if v := s.Check(); v != z3.True {
		fmt.Println("Unsolveable")
		return
	}
	m := s.Model()
	answer := m.Assignments()
	fmt.Println(answer)
}
