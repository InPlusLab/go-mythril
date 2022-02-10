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

	s := ctx.NewSolver()
	defer s.Close()

	x := ctx.NewBitvecVal(1, 16)
	y := ctx.NewBitvecVal(1, 16)
	k := x.BvSLe(y)

	fmt.Println(k.IsTrue())

	if v := s.Check(); v != z3.True {
		fmt.Println("Unsolveable")
		return
	}
	m := s.Model()
	answer := m.Assignments()
	fmt.Println(answer)
}
