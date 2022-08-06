package modules

import (
	"fmt"
	"go-mythril/analysis"
	"go-mythril/laser/ethereum/state"
	"go-mythril/laser/ethereum/transaction"
	"go-mythril/utils"
	"math/big"
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

func (dm *AccidentallyKillable) AddIssue(issue *analysis.Issue) {
	dm.Issues = append(dm.Issues, issue)
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
	to := globalState.Mstate.Stack.RawStack[globalState.Mstate.Stack.Length()-1]
	fmt.Println("SELFDESTRUCT in function ", globalState.Environment.ActiveFuncName)

	descriptionHead := "Any sender can cause the contract to self-destruct."
	constraints := state.NewConstraints()
	actors := transaction.NewActors(globalState.Z3ctx)
	attackerV, _ := new(big.Int).SetString(actors.Attacker, 16)
	for _, tx := range globalState.WorldState.TransactionSequence {
		switch tx.(type) {
		case *state.MessageCallTransaction:
			constraints.Add((tx.GetCaller().Eq(globalState.Z3ctx.NewBitvecVal(attackerV, 256))).And(
				tx.GetCaller().Eq(tx.GetOrigin())))
		}
	}

	var descriptionTail string
	var transactionSequence map[string]interface{}

	tmpCon := globalState.WorldState.Constraints.Copy()
	tmpCon.Add(constraints.ConstraintList...)
	tmpCon.Add(to.Eq(globalState.Z3ctx.NewBitvecVal(attackerV, 256)))
	transactionSequence = analysis.GetTransactionSequence(globalState, tmpCon)
	if transactionSequence == nil {
		tmpCon2 := globalState.WorldState.Constraints.Copy()
		tmpCon2.Add(constraints.ConstraintList...)
		transactionSequence = analysis.GetTransactionSequence(globalState, tmpCon2)
		if transactionSequence == nil {
			fmt.Println("No model found")
			return make([]*analysis.Issue, 0)
		} else {
			descriptionTail = "Any sender can trigger execution of the SELFDESTRUCT instruction to destroy this " +
				"contract account. Review the transaction trace generated for this issue and make sure that " +
				"appropriate security controls are in place to prevent unrestricted access."
		}
	} else {
		descriptionTail = "Any sender can trigger execution of the SELFDESTRUCT instruction to destroy this " +
			"contract account and withdraw its balance to an arbitrary address. Review the transaction trace " +
			"generated for this issue and make sure that appropriate security controls are in place to prevent " +
			"unrestricted access."
	}

	issue := &analysis.Issue{
		Contract:            globalState.Environment.ActiveAccount.ContractName,
		FunctionName:        globalState.Environment.ActiveFuncName,
		Address:             instruction.Address,
		SWCID:               analysis.NewSWCData()["UNPROTECTED_SELFDESTRUCT"],
		Bytecode:            globalState.Environment.Code.Bytecode,
		Title:               "Unprotected Selfdestruct",
		Severity:            "High",
		DescriptionHead:     descriptionHead,
		DescriptionTail:     descriptionTail,
		TransactionSequence: transactionSequence,
		GasUsed:             []int{globalState.Mstate.MinGasUsed, globalState.Mstate.MaxGasUsed},
	}
	issueArr := []*analysis.Issue{issue}
	return issueArr
}
