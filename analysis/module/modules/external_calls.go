package modules

import (
	"fmt"
	"go-mythril/analysis"
	"go-mythril/laser/ethereum/state"
	"go-mythril/laser/ethereum/transaction"
	"go-mythril/utils"
)

type ExternalCalls struct {
	Name        string
	SWCID       string
	Description string
	PreHooks    []string
	Issues      []*analysis.Issue
	Cache       *utils.Set
}

func NewExternalCalls() *ExternalCalls {
	return &ExternalCalls{
		Name:        "External call to another contract",
		SWCID:       analysis.NewSWCData()["REENTRANCY"],
		Description: "Search for external calls with unrestricted gas to a user-specified address.",
		PreHooks:    []string{"CALL"},
		Issues:      make([]*analysis.Issue, 0),
		Cache:       utils.NewSet(),
	}
}
func (dm *ExternalCalls) ResetModule() {
	dm.Issues = make([]*analysis.Issue, 0)
}
func (dm *ExternalCalls) Execute(target *state.GlobalState) []*analysis.Issue {
	fmt.Println("Entering analysis module: ", dm.Name)
	result := dm._execute(target)
	fmt.Println("Exiting analysis module:", dm.Name)
	return result
}

func (dm *ExternalCalls) AddIssue(issue *analysis.Issue) {
	dm.Issues = append(dm.Issues, issue)
}

func (dm *ExternalCalls) GetIssues() []*analysis.Issue {
	return dm.Issues
}

func (dm *ExternalCalls) GetPreHooks() []string {
	return dm.PreHooks
}

func (dm *ExternalCalls) GetPostHooks() []string {
	return make([]string, 0)
}

func (dm *ExternalCalls) _execute(globalState *state.GlobalState) []*analysis.Issue {
	potentialIssues := dm._analyze_state(globalState)
	annotation := GetPotentialIssuesAnnotaion(globalState)
	annotation.PotentialIssues = append(annotation.PotentialIssues, potentialIssues...)
	return nil
}

func (dm *ExternalCalls) _analyze_state(globalState *state.GlobalState) []*PotentialIssue {
	gas := globalState.Mstate.Stack.RawStack[globalState.Mstate.Stack.Length()-1]
	to := globalState.Mstate.Stack.RawStack[globalState.Mstate.Stack.Length()-2]
	address := globalState.GetCurrentInstruction().Address
	ACTORS := transaction.NewActors(globalState.Z3ctx)

	constraints := state.NewConstraints()
	constraints.Add(gas.BvUGt(globalState.Z3ctx.NewBitvecVal(2300, 256)), to.Eq(ACTORS.GetAttacker()))

	tmpCon := constraints.Copy()
	tmpCon.Add(globalState.WorldState.Constraints.ConstraintList...)
	fmt.Println("Constraints in externalCalls: ")
	fmt.Println(tmpCon.ConstraintList)
	for _, con := range tmpCon.ConstraintList {
		fmt.Println(con.AsAST().String())
	}

	transactionSequence := analysis.GetTransactionSequence(globalState, tmpCon)
	if transactionSequence == nil {
		// UnsatError
		fmt.Println("[EXTERNAL_CALLS] No model found.")
		return make([]*PotentialIssue, 0)
	}
	descriptionHead := "A call to a user-supplied address is executed."
	descriptionTail := "An external message call to an address specified by the caller is executed. Note that " +
		"the callee account might contain arbitrary code and could re-enter any function " +
		"within this contract. Reentering the contract in an intermediate state may lead to " +
		"unexpected behaviour. Make sure that no state modifications are executed after this call and/or reentrancy guards are in place."
	issue := &PotentialIssue{
		Contract:        globalState.Environment.ActiveAccount.ContractName,
		FunctionName:    globalState.Environment.ActiveFuncName,
		Address:         address,
		SWCID:           analysis.NewSWCData()["REENTRANCY"],
		Title:           "External Call To User-Supplied Address",
		Bytecode:        globalState.Environment.Code.Bytecode,
		Severity:        "Low",
		DescriptionHead: descriptionHead,
		DescriptionTail: descriptionTail,
		Constraints:     constraints,
		Detector:        dm,
	}
	return []*PotentialIssue{issue}
}
