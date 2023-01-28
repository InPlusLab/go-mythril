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
	ethereum.SetMaxSkipTimes(5)
	// Set the inputStr for each tx
	args := support.GetArgsInstance()
	args.TransactionSequences = []string{"6aba6fa1", "6aba6fa1", "6aba6fa1"}

	/* 单合约改这 */
	/* code for 0x78c2a1e91b52bca4130b6ed9edd9fbcfd4671c37.sol */
	creationCode := "608060405234801561001057600080fd5b50610328806100206000396000f300608060405260043610610041576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806311be40e014610046575b600080fd5b34801561005257600080fd5b5061012d600480360381019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803573ffffffffffffffffffffffffffffffffffffffff1690602001909291908035906020019082018035906020019080806020026020016040519081016040528093929190818152602001838360200280828437820191505050505050919291929080359060200190820180359060200190808060200260200160405190810160405280939291908181526020018383602002808284378201915050505050509192919290505050610147565b604051808215151515815260200191505060405180910390f35b600080600080855111151561015b57600080fd5b60405180807f7472616e7366657246726f6d28616464726573732c616464726573732c75696e81526020017f7432353629000000000000000000000000000000000000000000000000000000815250602501905060405180910390209150600090505b84518110156102ee578573ffffffffffffffffffffffffffffffffffffffff16827c0100000000000000000000000000000000000000000000000000000000900488878481518110151561020e57fe5b90602001906020020151878581518110151561022657fe5b906020019060200201516040518463ffffffff167c0100000000000000000000000000000000000000000000000000000000028152600401808473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200182815260200193505050506000604051808303816000875af1925050505080806001019150506101be565b6001925050509493505050505600a165627a7a72305820f6d67f11760b8487031f2b00d9c6c8b59210a65f390dc8c3c73ce5178da81c360029"
	runtimeCode := "608060405260043610610041576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806311be40e014610046575b600080fd5b34801561005257600080fd5b5061012d600480360381019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803573ffffffffffffffffffffffffffffffffffffffff1690602001909291908035906020019082018035906020019080806020026020016040519081016040528093929190818152602001838360200280828437820191505050505050919291929080359060200190820180359060200190808060200260200160405190810160405280939291908181526020018383602002808284378201915050505050509192919290505050610147565b604051808215151515815260200191505060405180910390f35b600080600080855111151561015b57600080fd5b60405180807f7472616e7366657246726f6d28616464726573732c616464726573732c75696e81526020017f7432353629000000000000000000000000000000000000000000000000000000815250602501905060405180910390209150600090505b84518110156102ee578573ffffffffffffffffffffffffffffffffffffffff16827c0100000000000000000000000000000000000000000000000000000000900488878481518110151561020e57fe5b90602001906020020151878581518110151561022657fe5b906020019060200201516040518463ffffffff167c0100000000000000000000000000000000000000000000000000000000028152600401808473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200182815260200193505050506000604051808303816000875af1925050505080806001019150506101be565b6001925050509493505050505600a165627a7a72305820f6d67f11760b8487031f2b00d9c6c8b59210a65f390dc8c3c73ce5178da81c360029"

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
