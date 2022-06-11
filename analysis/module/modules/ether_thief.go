package modules

import (
	"fmt"
	"go-mythril/analysis"
	"go-mythril/laser/ethereum/state"
	"go-mythril/laser/ethereum/transaction"
	"go-mythril/utils"
)

const DESCRIPTION = "Search for cases where Ether can be withdrawn to a user-specified address." +
	"An issue is reported if there is a valid end state where the attacker has successfully increased their Ether balance."

type EtherThief struct {
	Name        string
	SWCID       string
	Description string
	PostHooks   []string
	Issues      []*analysis.Issue
	Cache       *utils.Set
}

func NewEtherThief() *EtherThief {
	return &EtherThief{
		Name:        "Any sender can withdraw ETH from the contract account",
		SWCID:       analysis.NewSWCData()["UNPROTECTED_ETHER_WITHDRAWAL"],
		Description: DESCRIPTION,
		PostHooks:   []string{"CALL", "STATICCALL"},
		Issues:      make([]*analysis.Issue, 0),
		Cache:       utils.NewSet(),
	}
}

func (dm *EtherThief) ResetModule() {
	dm.Issues = make([]*analysis.Issue, 0)
}
func (dm *EtherThief) Execute(target *state.GlobalState) []*analysis.Issue {
	fmt.Println("Entering analysis module: ", dm.Name)
	result := dm._execute(target)
	fmt.Println("Exiting analysis module:", dm.Name)
	return result
}
func (dm *EtherThief) _execute(globalState *state.GlobalState) []*analysis.Issue {
	if dm.Cache.Contains(globalState.GetCurrentInstruction().Address) {
		return nil
	}
	potentialIssues := dm._analyze_state(globalState)
	annotation := analysis.GetPotentialIssuesAnnotaion(globalState)
	annotation.PotentialIssues = append(annotation.PotentialIssues, potentialIssues...)
	return nil
}

func (dm *EtherThief) _analyze_state(globalState *state.GlobalState) []*analysis.PotentialIssue {
	Gstate := globalState.Copy()
	ACTORS := transaction.NewActors(Gstate.Z3ctx)
	instruction := Gstate.GetCurrentInstruction()
	constraints := Gstate.WorldState.Constraints.Copy()

	constraints.Add(Gstate.WorldState.Balances.GetItem(ACTORS.GetAttacker()).BvUGt(Gstate.WorldState.StartingBalances.GetItem(ACTORS.GetAttacker())),
		Gstate.Environment.Sender.Eq(ACTORS.GetAttacker()),
		Gstate.CurrentTransaction().GetCaller().Eq(Gstate.CurrentTransaction().GetOrigin()))
	// Pre-solve so we only add potential issues if the attacker's balance is increased.
	_, sat := state.GetModel(constraints, nil,nil, true, globalState.Z3ctx)
	potentialIssue := &analysis.PotentialIssue{
		Contract:     Gstate.Environment.ActiveAccount.ContractName,
		FunctionName: Gstate.Environment.ActiveFuncName,
		// In post hook we use offset of previous instruction
		Address:         instruction.Address - 1,
		SWCID:           analysis.NewSWCData()["UNPROTECTED_ETHER_WITHDRAWAL"],
		Title:           "Unprotected Ether Withdrawal",
		Severity:        "High",
		Bytecode:        Gstate.Environment.Code.Bytecode,
		DescriptionHead: "Any sender can withdraw Ether from the contract account.",
		DescriptionTail: "Arbitrary senders other than the contract creator can profitably extract Ether " +
			"from the contract account. Verify the business logic carefully and make sure that appropriate " +
			"security controls are in place to prevent unexpected loss of funds.",
		Constraints: constraints,
	}
	if sat{
		return []*analysis.PotentialIssue{potentialIssue}
	}else{
		// UnsatError
		return make([]*analysis.PotentialIssue, 0)
	}
}
