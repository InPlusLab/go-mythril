package main

import (
	"fmt"
	"go-mythril/analysis"
	"go-mythril/analysis/module"
	"go-mythril/laser/ethereum"
	"go-mythril/laser/ethereum/state"
	"go-mythril/laser/smt/z3"
	"go-mythril/support"
	"runtime"
	"time"
)

func main() {
	runtime.GOMAXPROCS(16)

	fmt.Println("go mythril")

	LOADER := module.NewModuleLoader()

	//config := z3.NewConfig()
	config := z3.GetConfig()
	//z3.SetGlobalParam("parallel.enable", "true")
	ctx := z3.NewContext(config)

	evm := ethereum.NewLaserEVM(1, 1, 2, LOADER, config, 8)
	ethereum.SetMaxRLimitCount(1008610086)
	state.SetGetModelRLimit(3200000)
	// Set the inputStr for each tx
	args := support.GetArgsInstance()
	args.TransactionSequences = []string{"6aba6fa1", "6aba6fa1", "6aba6fa1"}

	/* 单合约改这 */
	/* code for 0x78c2a1e91b52bca4130b6ed9edd9fbcfd4671c37.sol */
	creationCode := "6080604052336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555034801561005057600080fd5b506104ed806100606000396000f300608060405260043610610062576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680633ccfd60b14610064578063b4a99a4e1461006e578063ba21d62a146100c5578063e0b0452114610141575b005b61006c61014b565b005b34801561007a57600080fd5b506100836102c4565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b61013f600480360381019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803590602001908201803590602001908080601f01602080910402602001604051908101604052809392919081815260200183838082843782019150505050505091929192905050506102e9565b005b6101496103d1565b005b732f61e7e1023bc22063b8da897d8323965a7712b773ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614156101e857732f61e7e1023bc22063b8da897d8323965a7712b76000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055505b6000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614151561024357600080fd5b6000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166108fc3073ffffffffffffffffffffffffffffffffffffffff16319081150290604051600060405180830381858888f193505050501580156102c1573d6000803e3d6000fd5b50565b6000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b6000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614151561034457600080fd5b8173ffffffffffffffffffffffffffffffffffffffff16348260405180828051906020019080838360005b8381101561038a57808201518184015260208101905061036f565b50505050905090810190601f1680156103b75780820380516001836020036101000a031916815260200191505b5091505060006040518083038185875af192505050505050565b670de0b6b3a76400003411156104bf576000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166108fc3073ffffffffffffffffffffffffffffffffffffffff16319081150290604051600060405180830381858888f1935050505015801561045f573d6000803e3d6000fd5b503373ffffffffffffffffffffffffffffffffffffffff166108fc3073ffffffffffffffffffffffffffffffffffffffff16319081150290604051600060405180830381858888f193505050501580156104bd573d6000803e3d6000fd5b505b5600a165627a7a72305820f22b8b27c88af7c98a8f0bd9542438002eb1ee266384bb289a1038cd6508b7140029"
	runtimeCode := "608060405260043610610062576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680633ccfd60b14610064578063b4a99a4e1461006e578063ba21d62a146100c5578063e0b0452114610141575b005b61006c61014b565b005b34801561007a57600080fd5b506100836102c4565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b61013f600480360381019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803590602001908201803590602001908080601f01602080910402602001604051908101604052809392919081815260200183838082843782019150505050505091929192905050506102e9565b005b6101496103d1565b005b732f61e7e1023bc22063b8da897d8323965a7712b773ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614156101e857732f61e7e1023bc22063b8da897d8323965a7712b76000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055505b6000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614151561024357600080fd5b6000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166108fc3073ffffffffffffffffffffffffffffffffffffffff16319081150290604051600060405180830381858888f193505050501580156102c1573d6000803e3d6000fd5b50565b6000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b6000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614151561034457600080fd5b8173ffffffffffffffffffffffffffffffffffffffff16348260405180828051906020019080838360005b8381101561038a57808201518184015260208101905061036f565b50505050905090810190601f1680156103b75780820380516001836020036101000a031916815260200191505b5091505060006040518083038185875af192505050505050565b670de0b6b3a76400003411156104bf576000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166108fc3073ffffffffffffffffffffffffffffffffffffffff16319081150290604051600060405180830381858888f1935050505015801561045f573d6000803e3d6000fd5b503373ffffffffffffffffffffffffffffffffffffffff166108fc3073ffffffffffffffffffffffffffffffffffffffff16319081150290604051600060405180830381858888f193505050501580156104bd573d6000803e3d6000fd5b505b5600a165627a7a72305820f22b8b27c88af7c98a8f0bd9542438002eb1ee266384bb289a1038cd6508b7140029"

	start := time.Now()
	/* 单文件中有多个合约，单协程多协程用下面 */
	//for i := 0; i < len(creationCodes); i++ {
	//	evm.SingleSymExec(creationCodes[i], runtimeCodes[i], "CHZ", ctx)
	//	evm.Refresh()
	//	x := function_managers.NewKeccakFunctionManager(ctx)
	//	fmt.Println(x)
	//	function_managers.RefreshKeccak()
	//}
	//for i := 0; i < len(creationCodes); i++ {
	//	evm.MultiSymExec(creationCodes[i], runtimeCodes[i], "CHZ", ctx, config)
	//	evm.Refresh()
	//	x := function_managers.NewKeccakFunctionManager(ctx)
	//	fmt.Println(x)
	//	function_managers.RefreshKeccak()
	//}

	/* 单文件中只有1个合约，单协程多协程用下面 */
	//evm.SingleSymExec(creationCode, runtimeCode, "CHZ", ctx)
	evm.MultiSymExec(creationCode, runtimeCode, "CHZ", ctx, config)

	duration := time.Since(start)

	// analysis
	for _, detector := range evm.Loader.Modules {
		issues := detector.GetIssues()

		// de-duplicate
		finalIssues := make([]*analysis.Issue, 0)
		for _, issue := range issues {
			existFlag := false
			for _, item := range finalIssues {
				if item.Address == issue.Address {
					existFlag = true
					break
				}
			}
			if !existFlag {
				finalIssues = append(finalIssues, issue)
			}
		}

		if len(finalIssues) > 0 {
			fmt.Println("number of issues:", len(finalIssues))
		}
		for _, issue := range finalIssues {
			fmt.Println("+++++++++++++++++++++++++++++++++++")
			fmt.Println("ContractName:", issue.Contract)
			fmt.Println("FunctionName:", issue.FunctionName)
			fmt.Println("Title:", issue.Title)
			fmt.Println("SWCID:", issue.SWCID)
			fmt.Println("Address:", issue.Address)
			fmt.Println("Severity:", issue.Severity)
		}
	}
	fmt.Println("+++++++++++++++++++++++++++++++++++")

	fmt.Println("Duration:", duration.Seconds())

	return
}
