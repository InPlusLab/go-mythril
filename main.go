package main

import (
	"fmt"
	"go-mythril/laser/ethereum"
)

func main() {
	fmt.Println("go mythril")
	evm := ethereum.NewLaserEVM(1, 1, 1)
	evm.SymExec("0x6060")
}
