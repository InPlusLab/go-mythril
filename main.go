package main

import (
	"fmt"
	"go-mythril/laser/ethereum"
)

type C struct {
	Size int
}

func main() {

	fmt.Println("go mythril")
	evm := ethereum.NewLaserEVM(1, 1, 1)
	// CreateBytecode and RuntimeBytecodeForTest()
	//testOriginRuntimeBytecode := "6080604052348015600f57600080fd5b506004361060285760003560e01c8063f8a8fd6d14602d575b600080fd5b60336035565b005b336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555056fea2646970667358221220cc40cae2e419544393419c7a7ea32f42d341094e9dc31099df83cbe79983591164736f6c63430008070033"
	testOriginCreateBytecode := "6080604052348015600f57600080fd5b5060ad8061001e6000396000f3fe6080604052348015600f57600080fd5b506004361060285760003560e01c8063f8a8fd6d14602d575b600080fd5b60336035565b005b336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555056fea2646970667358221220cc40cae2e419544393419c7a7ea32f42d341094e9dc31099df83cbe79983591164736f6c63430008070033"
	evm.NormalSymExec(testOriginCreateBytecode)
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
