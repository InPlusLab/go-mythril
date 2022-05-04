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
	// TODO: getTxSequence()

	issue := analysis.NewIssue(
		globalState.Environment.ActiveAccount.ContractName,
		globalState.Environment.ActiveFuncName,
		globalState.GetCurrentInstruction().Address,
		analysis.NewSWCData()["ARBITRARY_JUMP"],
		"Jump to an arbitrary instruction",
		globalState.Environment.Code.Bytecode,
		"High",
	)
	issueArr = append(issueArr, issue)
	return issueArr
}
