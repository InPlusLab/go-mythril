package modules

import (
	"fmt"
	"go-mythril/analysis"
	"go-mythril/laser/ethereum/state"
	"go-mythril/utils"
)

type AccidentallyKillable struct {
	Name        string
	SWCID       string
	Description string
	PreHooks    []string
	Issues      []*analysis.Issue
	Cache       *utils.Set
}

func NewAccidentallyKillable() *AccidentallyKillable {
	return &AccidentallyKillable{
		Name:  "Contract can be accidentally killed by anyone",
		SWCID: analysis.NewSWCData()["UNPROTECTED_SELFDESTRUCT"],
		Description: "Check if the contact can be 'accidentally' killed by anyone." +
			"For kill-able contracts, also check whether it is possible to direct the contract balance to the attacker.",
		PreHooks: []string{"SELFDESTRUCT"},
		Issues:   make([]*analysis.Issue, 0),
		Cache:    utils.NewSet(),
	}
}

func (dm *AccidentallyKillable) ResetModule() {
	dm.Issues = make([]*analysis.Issue, 0)
}

func (dm *AccidentallyKillable) Execute(target *state.GlobalState) []*analysis.Issue {
	fmt.Println("Entering analysis module: ", dm.Name)
	result := dm._execute(target)
	fmt.Println("Exiting analysis module:", dm.Name)
	return result
}

func (dm *AccidentallyKillable) _execute(globalState *state.GlobalState) []*analysis.Issue {
	if dm.Cache.Contains(globalState.GetCurrentInstruction().Address) {
		return nil
	}
	issues := dm._analyze_state(globalState)
	for _, issue := range issues {
		dm.Cache.Add(issue.Address)
	}
	dm.Issues = append(dm.Issues, issues...)
	return dm.Issues
}

func (dm *AccidentallyKillable) _analyze_state(globalState *state.GlobalState) []*analysis.Issue {
	fmt.Println("Suicide module: Analyzing suicide instruction")
	instruction := globalState.GetCurrentInstruction()
	// to := globalState.Mstate.Stack.RawStack[globalState.Mstate.Stack.Length()-1]
	fmt.Println("SELFDESTRUCT in function ", globalState.Environment.ActiveFuncName)
	// descriptionHead
	// constraints

	issue := analysis.NewIssue(
		globalState.Environment.ActiveAccount.ContractName,
		globalState.Environment.ActiveFuncName,
		instruction.Address,
		analysis.NewSWCData()["UNPROTECTED_SELFDESTRUCT"],
		"Unprotected Selfdestruct",
		globalState.Environment.Code.Bytecode,
		"high",
	)
	issueArr := []*analysis.Issue{issue}
	return issueArr
}
