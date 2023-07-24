package modules

import (
	"fmt"
	"go-mythril/analysis"
	"go-mythril/laser/ethereum/state"
	"go-mythril/laser/ethereum/transaction"
	"go-mythril/laser/smt/z3"
	"go-mythril/utils"
)

const DESCRIPTION = "Search for cases where Ether can be withdrawn to a user-specified address." +
	"An issue is reported if there is a valid end state where the attacker has successfully increased their Ether balance."

type EtherThief struct {
	Name        string
	SWCID       string
	Description string
	PostHooks   []string
	Issues      *utils.SyncSlice
	Cache       *utils.Set
}

func NewEtherThief() *EtherThief {
	return &EtherThief{
		Name:        "Any sender can withdraw ETH from the contract account",
		SWCID:       analysis.NewSWCData()["UNPROTECTED_ETHER_WITHDRAWAL"],
		Description: DESCRIPTION,
		PostHooks:   []string{"CALL", "STATICCALL"},
		Issues:      utils.NewSyncSlice(),
		Cache:       utils.NewSet(),
	}
}

func (dm *EtherThief) ResetModule() {
	dm.Issues = utils.NewSyncSlice()
}
func (dm *EtherThief) Execute(target *state.GlobalState) []*analysis.Issue {
	// fmt.Println("Entering analysis module: ", dm.Name)
	result := dm._execute(target)
	// fmt.Println("Exiting analysis module:", dm.Name)
	return result
}

func (dm *EtherThief) AddIssue(issue *analysis.Issue) {
	dm.Issues.Append(issue)
}

func (dm *EtherThief) GetIssues() []*analysis.Issue {
	list := make([]*analysis.Issue, 0)
	for _, v := range dm.Issues.Elements() {
		list = append(list, v.(*analysis.Issue))
	}
	return list
}

func (dm *EtherThief) GetPreHooks() []string {
	return make([]string, 0)
}

func (dm *EtherThief) GetPostHooks() []string {
	return dm.PostHooks
}

func (dm *EtherThief) GetCache() *utils.Set {
	return dm.Cache
}

func (dm *EtherThief) _execute(globalState *state.GlobalState) []*analysis.Issue {
	if dm.Cache.Contains(globalState.GetCurrentInstruction().Address) {
		return nil
	}
	potentialIssues := dm._analyze_state(globalState)
	annotation := GetPotentialIssuesAnnotaion(globalState)
	annotation.Append(potentialIssues...)
	return nil
}

func (dm *EtherThief) _analyze_state(globalState *state.GlobalState) []*PotentialIssue {
	//config := z3.GetConfig()
	//newCtx := z3.NewContext(config)

	//Gstate := globalState.Copy()
	Gstate := globalState
	ACTORS := transaction.NewActors(Gstate.Z3ctx)
	instruction := Gstate.GetCurrentInstruction()
	constraints := Gstate.WorldState.Constraints.Copy()

	constraints.Add(Gstate.WorldState.Balances.GetItem(ACTORS.GetAttacker().Translate(Gstate.Z3ctx)).BvUGt(Gstate.WorldState.StartingBalances.GetItem(ACTORS.GetAttacker().Translate(Gstate.Z3ctx))),
		Gstate.Environment.Sender.Translate(Gstate.Z3ctx).Eq(ACTORS.GetAttacker().Translate(Gstate.Z3ctx)),
		Gstate.CurrentTransaction().GetCaller().Translate(Gstate.Z3ctx).Eq(Gstate.CurrentTransaction().GetOrigin().Translate(Gstate.Z3ctx)))
	// Pre-solve so we only add potential issues if the attacker's balance is increased.
	_, sat := state.GetModel(constraints, make([]*z3.Bool, 0), make([]*z3.Bool, 0), true, globalState.Z3ctx)
	potentialIssue := &PotentialIssue{
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
		//Constraints: constraints.Translate(newCtx),
		Detector: dm,
	}
	if sat {
		fmt.Println("etherThief success")
		return []*PotentialIssue{potentialIssue}
	} else {
		// UnsatError
		fmt.Println("etherThief fail")
		return make([]*PotentialIssue, 0)
	}
}
