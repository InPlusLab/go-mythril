package main

import (
	"fmt"
	"go-mythril/analysis/module"
	"go-mythril/laser/ethereum"
	"time"
)

type C struct {
	Size int
}

func main() {

	fmt.Println("go mythril")
	LOADER := module.NewModuleLoader()
	evm := ethereum.NewLaserEVM(1, 1, 1, LOADER)

	/* code for IntegerOverflow.sol */
	// integerOverflowCreateBytecode := "6080604052600160005534801561001557600080fd5b506101e3806100256000396000f3fe608060405234801561001057600080fd5b50600436106100365760003560e01c80631003e2d21461003b578063b69ef8a814610057575b600080fd5b610055600480360381019061005091906100ab565b610075565b005b61005f610090565b60405161006c91906100e7565b60405180910390f35b806000808282546100869190610102565b9250508190555050565b60005481565b6000813590506100a581610196565b92915050565b6000602082840312156100c1576100c0610191565b5b60006100cf84828501610096565b91505092915050565b6100e181610158565b82525050565b60006020820190506100fc60008301846100d8565b92915050565b600061010d82610158565b915061011883610158565b9250827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0382111561014d5761014c610162565b5b828201905092915050565b6000819050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b600080fd5b61019f81610158565b81146101aa57600080fd5b5056fea264697066735822122087b595ab091f2f063af4481cbf3f5906249a224c0b0e343c050f64821d66d84864736f6c63430008070033"
	//integetOverflowRuntimeBytecode := "608060405234801561001057600080fd5b50600436106100365760003560e01c80631003e2d21461003b578063b69ef8a814610057575b600080fd5b610055600480360381019061005091906100d1565b610075565b005b61005f610090565b60405161006c919061010d565b60405180910390f35b806000808282546100869190610157565b9250508190555050565b60005481565b600080fd5b6000819050919050565b6100ae8161009b565b81146100b957600080fd5b50565b6000813590506100cb816100a5565b92915050565b6000602082840312156100e7576100e6610096565b5b60006100f5848285016100bc565b91505092915050565b6101078161009b565b82525050565b600060208201905061012260008301846100fe565b92915050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b60006101628261009b565b915061016d8361009b565b925082820190508082111561018557610184610128565b5b9291505056fea264697066735822122074145123a5d293c019a6db8699647b33d5be9eca827645a3e76c43e3d7cf7d7664736f6c63430008100033"
	//integerOverflowRuntimeBytecode := "6080604052600436106049576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806306661abd14604e578063a444f5e9146076575b600080fd5b348015605957600080fd5b50606060a0565b6040518082815260200191505060405180910390f35b348015608157600080fd5b50609e6004803603810190808035906020019092919050505060a6565b005b60005481565b806000808282540392505081905550505600a165627a7a72305820ec7c669611e9bf0e29af59657d4b10d43a6eef7537f52ac2dcbfa47e47af50970029"
	/* code: PUSH32(max of 256bits) PUSH1(0x2) ADD PUSH1(0x29) JUMPI PUSH1(0x01) JUMPDEST STOP */
	// mytestcode := "7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff60020160295760015b00"

	/* code for Origin.sol */
	// originCreateBytecode := "608060405234801561001057600080fd5b50336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555061025a806100606000396000f3fe608060405234801561001057600080fd5b50600436106100365760003560e01c80638da5cb5b1461003b578063f2fde38b14610059575b600080fd5b610043610075565b60405161005091906101bb565b60405180910390f35b610073600480360381019061006e919061017f565b610099565b005b60008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b60008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163273ffffffffffffffffffffffffffffffffffffffff1614156100f257600080fd5b600073ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff161461016757806000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055505b50565b6000813590506101798161020d565b92915050565b60006020828403121561019557610194610208565b5b60006101a38482850161016a565b91505092915050565b6101b5816101d6565b82525050565b60006020820190506101d060008301846101ac565b92915050565b60006101e1826101e8565b9050919050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b600080fd5b610216816101d6565b811461022157600080fd5b5056fea26469706673582212202219ec9676147dd8fad11007fa97b79075f0279094148f69b86ee8276d7fcf8464736f6c63430008070033"
	//originRuntimeBytecode := "60806040526004361061004c576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680638da5cb5b14610051578063f2fde38b146100a8575b600080fd5b34801561005d57600080fd5b506100666100f9565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b3480156100b457600080fd5b506100f7600480360360208110156100cb57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919050505061011e565b005b6000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b6000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163273ffffffffffffffffffffffffffffffffffffffff161415151561017a57600080fd5b600073ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff161415156101f157806000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055505b5056fea165627a7a72305820323ae5791f7e8e7617378072b1a3c68a016512f6a961fd74c25e8f5540f0a9f60029"
	//originRuntimeBytecode := "608060405234801561001057600080fd5b50600436106100365760003560e01c80638da5cb5b1461003b578063f2fde38b1461006f575b600080fd5b6100436100b3565b604051808273ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b6100b16004803603602081101561008557600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506100d7565b005b60008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b60008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163273ffffffffffffffffffffffffffffffffffffffff161461012f57600080fd5b600073ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff16146101a457806000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055505b5056fea264697066735822122044fda70dd58730e017e83dbdc779fc814720ecbc7f11cb487408899e5796ab0264736f6c63430007000033"
	/* code for TimeStamp.sol */
	//timestampRuntimeBytecode := "60806040526004361060265760003560e01c80633ccfd60b14602b57806358d02b09146033575b600080fd5b6031605b565b005b348015603e57600080fd5b50604560de565b6040518082815260200191505060405180910390f35b670de0b6b3a7640000341015606f57600080fd5b600054421415607d57600080fd5b426000819055506000600f4281608f57fe5b06141560dc573373ffffffffffffffffffffffffffffffffffffffff166108fc479081150290604051600060405180830381858888f1935050505015801560da573d6000803e3d6000fd5b505b565b6000548156fea26469706673582212207964489ee642f167fbbe9e9da0ebe57c7f4fda94eba52cae1b111a92a7c5f94b64736f6c63430007000033"

	/* code for Reentrancy.sol */
	reentrancyRuntimeBytecode := "608060405260043610610056576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff168062362a951461005b5780632e1a7d4d14610091578063d5d44d80146100be575b600080fd5b61008f600480360381019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610115565b005b34801561009d57600080fd5b506100bc60048036038101908080359060200190929190505050610164565b005b3480156100ca57600080fd5b506100ff600480360381019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610232565b6040518082815260200191505060405180910390f35b346000808373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000828254019250508190555050565b806000803373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205410151561022f573373ffffffffffffffffffffffffffffffffffffffff168160405160006040518083038185875af19250505015156101e257600080fd5b806000803373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600082825403925050819055505b50565b600060205280600052604060002060009150905054815600a165627a7a723058207948469cd963b4c9fe220456690d1a527b6ae59eb1ee5876a2b3996ce93dc1e80029"
	start := time.Now()
	//evm.NormalSymExec(reentrancyRuntimeBytecode, "Time")
	evm.SymExec(reentrancyRuntimeBytecode, "Time")
	duration := time.Since(start)
	fmt.Println("Duration:", duration)
	fmt.Println("end of code")
	return

	/*	evm.NormalSymExec("6060320032")
		return*/
	/*
		fmt.Println("go mythril-testForGoz3")
		config := z3.NewConfig()
		ctx := z3.NewContext(config)
		config.Close()
		defer ctx.Close()

		s := ctx.NewSolver()
		defer s.Close()

		//zero := ctx.NewBitvecVal(0, 256)
		one := ctx.NewBitvecVal(1, 256)
		timestamp := ctx.NewBitvec("timestamp", 256)
		z := one.BvUGt(timestamp)
		k := one.BvSGt(timestamp)
		fmt.Println(z.ToString(), z.IsTrue())
		fmt.Println(k.ToString(), k.IsTrue())
		//s.Assert(z.AsAST().Simplify())
		//`
		//if v := s.Check(); v != z3.True {
		//	fmt.Println("Unsolveable")
		//	return
		//}
		//
		//m := s.Model()
		//answer := m.Assignments()
		//fmt.Println("solver0's result:")
		//fmt.Println(answer)
	*/
}
