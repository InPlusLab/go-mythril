package main

import (
	"fmt"
	"go-mythril/laser/ethereum"
)

func main() {
	fmt.Println("go mythril")
	evm := ethereum.NewLaserEVM(1, 1, 1)

	// testOriginRuntimeBytecode := "6080604052348015600f57600080fd5b506004361060285760003560e01c8063f8a8fd6d14602d575b600080fd5b60336035565b005b336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555056fea264697066735822122030a096a27c01e8d3a63ff01b1934b9c994ea9c5b2ed8819b90056f3b6edc2b8064736f6c63430008070033" // see: temp/origin.sol
	testOriginRuntimeBytecode := "6080604052348015600f57600080fd5b506004361060285760003560e01c8063f8a8fd6d14602d575b600080fd5b60336035565b005b336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555056fea26469706673582212202ea93e8e22a45ad0ed67dd6422256ae66d52bfd16b4b20f0d3928171fba3f69764736f6c63430008070033"
	evm.NormalSymExec(testOriginRuntimeBytecode)
	return

	/*	evm.NormalSymExec("6060320032")
		return*/

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
