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

	arr := ctx.NewArray("arr1", 16, 16)
	// SetItem will return a new array.
	// arrNew[1] = 2
	arrNew := arr.SetItem(ctx.Int(1, ctx.BvSort(16)), ctx.Int(2, ctx.BvSort(16)))
	item := ctx.Const(ctx.Symbol("item"), ctx.BvSort(16))

	s.Assert(arrNew.GetItem(ctx.Int(1, ctx.BvSort(16))).Eq(item))

	if v := s.Check(); v != z3.True {
		fmt.Println("Unsolveable")
		return
	}

	m := s.Model()
	assignments := m.Assignments()
	m.Close()

	fmt.Println(assignments)
}
