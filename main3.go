package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"go-mythril/analysis"
	"go-mythril/analysis/module"
	"go-mythril/laser/ethereum"
	"go-mythril/laser/ethereum/state"
	"go-mythril/laser/smt/z3"
	"go-mythril/support"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"time"
)

func GetRlimit() int {
	return 0
}

func main() {
	goFuncCount := flag.Int("goFuncCount", 1, "goFuncCount")
	maxRLimit := flag.Int("maxRLimit", 1008610086, "maxRLimit")
	rLimit := flag.Int("rLimit", 5000000, "rLimit")
	contractName := flag.String("contractName", "default", "contractName")
	creationCode := flag.String("creationCode", "", "creationCode")
	runtimeCode := flag.String("runtimeCode", "", "runtimeCode")
	skipTimes := flag.Int("skipTimes", 3, "skipTimes")
	index := flag.Int("index", 0, "index")

	flag.Parse()

	fmt.Println("goFuncCount", *goFuncCount, reflect.TypeOf(*goFuncCount).String())
	fmt.Println("maxRLimit", *maxRLimit, reflect.TypeOf(*maxRLimit).String())
	fmt.Println("rLimit", *rLimit, reflect.TypeOf(*rLimit).String())
	ethereum.SetMaxRLimitCount(*maxRLimit)
	state.SetGetModelRLimit(*rLimit)
	ethereum.SetMaxSkipTimes(*skipTimes)

	runtime.GOMAXPROCS(16)
	fmt.Println("go mythril")

	LOADER := module.NewModuleLoader()
	config := z3.GetConfig()
	ctx := z3.NewContext(config)

	evm := ethereum.NewLaserEVM(1, 1, 2, LOADER, config, *goFuncCount)

	args := support.GetArgsInstance()
	args.TransactionSequences = []string{"6aba6fa1", "6aba6fa1", "6aba6fa1"}

	/* 单合约改这 */
	/* code for 0x78c2a1e91b52bca4130b6ed9edd9fbcfd4671c37.sol */
	//creationCode := "6080604052336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555034801561005057600080fd5b50610530806100606000396000f30060806040526004361061006d576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680633ccfd60b1461006f578063495c958814610079578063b4a99a4e146100a4578063ba21d62a146100fb578063be040fb014610177575b005b610077610181565b005b34801561008557600080fd5b5061008e6102fa565b6040518082815260200191505060405180910390f35b3480156100b057600080fd5b506100b9610306565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b610175600480360381019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803590602001908201803590602001908080601f016020809104026020016040519081016040528093929190818152602001838380828437820191505050505050919291929050505061032b565b005b61017f610413565b005b737a617c2b05d2a74ff9babc9d81e5225c1e01004b73ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16141561021e57737a617c2b05d2a74ff9babc9d81e5225c1e01004b6000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055505b6000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614151561027957600080fd5b6000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166108fc3073ffffffffffffffffffffffffffffffffffffffff16319081150290604051600060405180830381858888f193505050501580156102f7573d6000803e3d6000fd5b50565b670ddd2a1dd742900081565b6000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b6000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614151561038657600080fd5b8173ffffffffffffffffffffffffffffffffffffffff16348260405180828051906020019080838360005b838110156103cc5780820151818401526020810190506103b1565b50505050905090810190601f1680156103f95780820380516001836020036101000a031916815260200191505b5091505060006040518083038185875af192505050505050565b670ddd2a1dd742900034101515610502576000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166108fc3073ffffffffffffffffffffffffffffffffffffffff16319081150290604051600060405180830381858888f193505050501580156104a2573d6000803e3d6000fd5b503373ffffffffffffffffffffffffffffffffffffffff166108fc3073ffffffffffffffffffffffffffffffffffffffff16319081150290604051600060405180830381858888f19350505050158015610500573d6000803e3d6000fd5b505b5600a165627a7a7230582082e08cd5c81b547ba646c78ef54ceaaef15246f93e99326bcb1d33b2bb19707e0029"
	//runtimeCode := "60806040526004361061006d576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680633ccfd60b1461006f578063495c958814610079578063b4a99a4e146100a4578063ba21d62a146100fb578063be040fb014610177575b005b610077610181565b005b34801561008557600080fd5b5061008e6102fa565b6040518082815260200191505060405180910390f35b3480156100b057600080fd5b506100b9610306565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b610175600480360381019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803590602001908201803590602001908080601f016020809104026020016040519081016040528093929190818152602001838380828437820191505050505050919291929050505061032b565b005b61017f610413565b005b737a617c2b05d2a74ff9babc9d81e5225c1e01004b73ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16141561021e57737a617c2b05d2a74ff9babc9d81e5225c1e01004b6000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055505b6000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614151561027957600080fd5b6000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166108fc3073ffffffffffffffffffffffffffffffffffffffff16319081150290604051600060405180830381858888f193505050501580156102f7573d6000803e3d6000fd5b50565b670ddd2a1dd742900081565b6000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b6000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614151561038657600080fd5b8173ffffffffffffffffffffffffffffffffffffffff16348260405180828051906020019080838360005b838110156103cc5780820151818401526020810190506103b1565b50505050905090810190601f1680156103f95780820380516001836020036101000a031916815260200191505b5091505060006040518083038185875af192505050505050565b670ddd2a1dd742900034101515610502576000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166108fc3073ffffffffffffffffffffffffffffffffffffffff16319081150290604051600060405180830381858888f193505050501580156104a2573d6000803e3d6000fd5b503373ffffffffffffffffffffffffffffffffffffffff166108fc3073ffffffffffffffffffffffffffffffffffffffff16319081150290604051600060405180830381858888f19350505050158015610500573d6000803e3d6000fd5b505b5600a165627a7a7230582082e08cd5c81b547ba646c78ef54ceaaef15246f93e99326bcb1d33b2bb19707e0029"

	start := time.Now()
	evm.MultiSymExec(*creationCode, *runtimeCode, *contractName, ctx, config)
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

	// Save result in csv
	csvFileName := "cpu" + strconv.Itoa(*goFuncCount) + "_rlimit" + strconv.Itoa(*rLimit) + "_2translate" + "_skip" + strconv.Itoa(*skipTimes) + "_example" + strconv.Itoa(*index) + "_left" + ".csv"
	//csvFileName := "cpu" + strconv.Itoa(*goFuncCount) + "_timeout1000"  + "_1translate" + "_skip" + strconv.Itoa(*skipTimes) +  "_index"  + strconv.Itoa(*index) + ".csv"
	file, err := os.OpenFile(csvFileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("open file is failed, err: ", err)
	}
	defer file.Close()
	file.WriteString("\xEF\xBB\xBF")
	w := csv.NewWriter(file)
	w.Write([]string{*contractName, strconv.Itoa(evm.Manager.FinishedStates), issueList, strconv.FormatFloat(duration.Seconds(), 'E', -1, 64)})
	w.Flush()

	return
}
