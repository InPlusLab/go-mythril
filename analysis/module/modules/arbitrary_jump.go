package modules

import (
	"fmt"
	"go-mythril/analysis"
	"go-mythril/laser/ethereum/state"
	"go-mythril/utils"
)

type ArbitraryJump struct {
	Name        string
	SWCID       string
	Description string
	PreHooks    []string
	Issues      []*analysis.Issue
	Cache       *utils.Set
}

func NewArbitraryJump() *ArbitraryJump {
	return &ArbitraryJump{
		Name:        "Caller can redirect execution to arbitrary bytecode locations",
		SWCID:       analysis.NewSWCData()["ARBITRARY_JUMP"],
		Description: "",
		PreHooks:    []string{"JUMP", "JUMPI"},
		Issues:      make([]*analysis.Issue, 0),
		Cache:       utils.NewSet(),
	}
}

func (dm *ArbitraryJump) ResetModule() {
	dm.Issues = make([]*analysis.Issue, 0)
}

func (dm *ArbitraryJump) Execute(target *state.GlobalState) []*analysis.Issue {
	fmt.Println("Entering analysis module: ", dm.Name)
	result := dm._execute(target)
	fmt.Println("Exiting analysis module:", dm.Name)
	return result
}

func (dm *ArbitraryJump) GetIssues() []*analysis.Issue {
	return dm.Issues
}

func (dm *ArbitraryJump) _execute(globalState *state.GlobalState) []*analysis.Issue {
	if dm.Cache.Contains(globalState.GetCurrentInstruction().Address) {
		return nil
	}
	issues := dm._analyze_state(globalState)
	dm.Issues = append(dm.Issues, issues...)
	return dm.Issues
}

func (dm *ArbitraryJump) _analyze_state(globalState *state.GlobalState) []*analysis.Issue {

	issueArr := make([]*analysis.Issue, 0)
	jumpDest := globalState.Mstate.Stack.RawStack[globalState.Mstate.Stack.Length()-1]
	if !jumpDest.Symbolic() {
		return issueArr
	}
	transactionSequence := analysis.GetTransactionSequence(globalState, globalState.WorldState.Constraints)
	if transactionSequence == nil {
		// UnsatError
		return issueArr
	}
	issue := &analysis.Issue{
		Contract:        globalState.Environment.ActiveAccount.ContractName,
		FunctionName:    globalState.Environment.ActiveFuncName,
		Address:         globalState.GetCurrentInstruction().Address,
		Bytecode:        globalState.Environment.Code.Bytecode,
		Title:           "Jump to an arbitrary instruction",
		SWCID:           analysis.NewSWCData()["ARBITRARY_JUMP"],
		Severity:        "High",
		DescriptionHead: "The caller can redirect execution to arbitrary bytecode locations.",
		DescriptionTail: "It is possible to redirect the control flow to arbitrary locations in the code. " +
			"This may allow an attacker to bypass security controls or manipulate the business logic of the " +
			"smart contract. Avoid using low-level-operations and assembly to prevent this issue.",
		GasUsed:             []int{globalState.Mstate.MinGasUsed, globalState.Mstate.MaxGasUsed},
		TransactionSequence: transactionSequence,
	}
	issueArr = append(issueArr, issue)
	return issueArr
}
