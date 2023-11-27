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
	"strconv"
	"time"
)
func main() {

	ethereum.SetMaxRLimitCount(9600000)
	state.SetGetModelRLimit(3200000)
	ethereum.SetMaxSkipTimes(0)

	runtime.GOMAXPROCS(16)
	fmt.Println("go mythril")

	LOADER := module.NewModuleLoader()
	config := z3.GetConfig()
	ctx := z3.NewContext(config)

	evm := ethereum.NewLaserEVM(1, 1, 2, LOADER, config, 1)

	args := support.GetArgsInstance()
	args.TransactionSequences = []string{"6aba6fa1", "6aba6fa1", "6aba6fa1"}

	/* 单合约改这 */
	/* code for 0x78c2a1e91b52bca4130b6ed9edd9fbcfd4671c37.sol */
	creationCode := "608060405234801561001057600080fd5b5060c88061001f6000396000f300608060405260043610603f576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff168063a5843f08146044575b600080fd5b348015604f57600080fd5b50607660048036038101908080359060200190929190803590602001909291905050506078565b005b806000808481526020019081526020016000206000828254039250508190555050505600a165627a7a72305820ea338418a76151d448e291eba41eb6fb0d30ee5924c10f363f9b91b2a9de90d30029"
	runtimeCode := "608060405260043610603f576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff168063a5843f08146044575b600080fd5b348015604f57600080fd5b50607660048036038101908080359060200190929190803590602001909291905050506078565b005b806000808481526020019081526020016000206000828254039250508190555050505600a165627a7a72305820ea338418a76151d448e291eba41eb6fb0d30ee5924c10f363f9b91b2a9de90d30029"

	start := time.Now()
	//evm.SingleSymExec(creationCode, runtimeCode, "test", ctx)
	evm.MultiSymExec(creationCode, runtimeCode, "test", ctx, config)
	duration := time.Since(start)

	// analysis
	issueList := ""
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
			str := issue.SWCID + "-" + strconv.Itoa(issue.Address) + " "
			issueList += str
		}
	}
	evm.Manager.LogInfo()
	fmt.Println("+++++++++++++++++++++++++++++++++++")
	fmt.Println(issueList)
	fmt.Println("Duration:", duration.Seconds())
	fmt.Println("TotalStates:", evm.Manager.TotalStates)

	// symLog
	ethereum.SymTreeLog(ethereum.GetHeadNode())
	ethereum.SymTreeCount()

	return
}
