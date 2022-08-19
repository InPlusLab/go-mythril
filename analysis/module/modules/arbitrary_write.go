package modules

import (
	"fmt"
	"go-mythril/analysis"
	"go-mythril/laser/ethereum/state"
	"go-mythril/utils"
)

type ArbitraryStorage struct {
	Name        string
	SWCID       string
	Description string
	PreHooks    []string
	Issues      []*analysis.Issue
	Cache       *utils.Set
}

func NewArbitraryStorage() *ArbitraryStorage {
	return &ArbitraryStorage{
		Name:        "Caller can write to arbitrary storage locations",
		SWCID:       analysis.NewSWCData()["WRITE_TO_ARBITRARY_STORAGE"],
		Description: "",
		PreHooks:    []string{"SSTORE"},
		Issues:      make([]*analysis.Issue, 0),
		Cache:       utils.NewSet(),
	}
}

func (dm *ArbitraryStorage) ResetModule() {
	dm.Issues = make([]*analysis.Issue, 0)
}

func (dm *ArbitraryStorage) Execute(target *state.GlobalState) []*analysis.Issue {
	fmt.Println("Entering analysis module: ", dm.Name)
	result := dm._execute(target)
	fmt.Println("Exiting analysis module:", dm.Name)
	return result
}

func (dm *ArbitraryStorage) AddIssue(issue *analysis.Issue) {
	dm.Issues = append(dm.Issues, issue)
}

func (dm *ArbitraryStorage) GetIssues() []*analysis.Issue {
	return dm.Issues
}

func (dm *ArbitraryStorage) GetPreHooks() []string {
	return dm.PreHooks
}

func (dm *ArbitraryStorage) GetPostHooks() []string {
	return make([]string, 0)
}

func (dm *ArbitraryStorage) _execute(globalState *state.GlobalState) []*analysis.Issue {
	if dm.Cache.Contains(globalState.GetCurrentInstruction().Address) {
		return nil
	}
	potentialIssues := dm._analyze_state(globalState)
	annotation := GetPotentialIssuesAnnotaion(globalState)
	annotation.PotentialIssues = append(annotation.PotentialIssues, potentialIssues...)
	return nil
}

func (dm *ArbitraryStorage) _analyze_state(globalState *state.GlobalState) []*PotentialIssue {
	writeSlot := globalState.Mstate.Stack.RawStack[globalState.Mstate.Stack.Length()-1]
	ctx := writeSlot.GetCtx()
	constrains := globalState.WorldState.Constraints.Copy()
	constrains.Add(writeSlot.Eq(ctx.NewBitvecVal(324345425435, 256)))
	potentialIssue := &PotentialIssue{
		Contract:        globalState.Environment.ActiveAccount.ContractName,
		FunctionName:    globalState.Environment.ActiveFuncName,
		Address:         globalState.GetCurrentInstruction().Address,
		SWCID:           analysis.NewSWCData()["WRITE_TO_ARBITRARY_STORAGE"],
		Title:           "Write to an arbitrary storage location",
		Severity:        "High",
		Bytecode:        globalState.Environment.Code.Bytecode,
		DescriptionHead: "The caller can write to arbitrary storage locations.",
		DescriptionTail: "It is possible to write to arbitrary storage locations. By modifying the values of storage variables, attackers may bypass security controls or manipulate the business logic of the smart contract.",
		Constraints:     constrains,
	}
	return []*PotentialIssue{potentialIssue}
}
