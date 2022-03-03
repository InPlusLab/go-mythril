package main

import (
	"fmt"
	"go-mythril/laser/ethereum"
)

func main() {

	fmt.Println("go mythril")
	evm := ethereum.NewLaserEVM(1, 1, 1)
	evm.NormalSymExec("6060320032")

	return

	/*	fmt.Println("go mythril-testForGoz3")
		config := z3.NewConfig()
		ctx := z3.NewContext(config)
		config.Close()
		defer ctx.Close()

		s := ctx.NewSolver()
		defer s.Close()

		x := ctx.NewBitvecVal(7, 16)
		y := ctx.NewBitvecVal(1, 16)
		z := ctx.NewBitvec("z", 16)
		fmt.Println(x.Value(), x.BvSize())
		fmt.Println(y.Value(), y.BvSize())
		fmt.Println(z.Value(), z.BvSize())
		fmt.Println(x.BvAdd(y).Value(), x.BvAdd(y).BvSize())

		s.Assert(z.Eq(x.BvAdd(y)).AsAST())

		if v := s.Check(); v != z3.True {
			fmt.Println("Unsolveable")
			return
		}

		m := s.Model()
		answer := m.Assignments()
		fmt.Println("solver0's result:")
		fmt.Println(answer)
		fmt.Println(z.Value(), z.BvSize())*/
}
