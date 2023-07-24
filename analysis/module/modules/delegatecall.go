package modules

import (
	"fmt"
	"go-mythril/analysis"
	"go-mythril/laser/ethereum/state"
	"go-mythril/laser/ethereum/transaction"
	"go-mythril/utils"
	"reflect"
	"strconv"
)

type ArbitraryDelegateCall struct {
	Name        string
	SWCID       string
	Description string
	PreHooks    []string
	Issues      *utils.SyncSlice
	Cache       *utils.Set
}

func NewArbitraryDelegateCall() *ArbitraryDelegateCall {
	return &ArbitraryDelegateCall{
		Name:        "Delegatecall to a user-specified address",
		SWCID:       analysis.NewSWCData()["DELEGATECALL_TO_UNTRUSTED_CONTRACT"],
		Description: "Check for invocations of delegatecall to a user-supplied address.",
		PreHooks:    []string{"DELEGATECALL"},
		Issues:      utils.NewSyncSlice(),
		Cache:       utils.NewSet(),
	}
}

func (dm *ArbitraryDelegateCall) ResetModule() {
	dm.Issues = utils.NewSyncSlice()
}

func (dm *ArbitraryDelegateCall) Execute(target *state.GlobalState) []*analysis.Issue {
	// fmt.Println("Entering analysis module: ", dm.Name)
	result := dm._execute(target)
	// fmt.Println("Exiting analysis module:", dm.Name)
	return result
}

func (dm *ArbitraryDelegateCall) AddIssue(issue *analysis.Issue) {
	dm.Issues.Append(issue)
}

func (dm *ArbitraryDelegateCall) GetIssues() []*analysis.Issue {
	list := make([]*analysis.Issue, 0)
	for _, v := range dm.Issues.Elements() {
		list = append(list, v.(*analysis.Issue))
	}
	return list
}

func (dm *ArbitraryDelegateCall) GetPreHooks() []string {
	return dm.PreHooks
}

func (dm *ArbitraryDelegateCall) GetPostHooks() []string {
	return make([]string, 0)
}

func (dm *ArbitraryDelegateCall) GetCache() *utils.Set {
	return dm.Cache
}

func (dm *ArbitraryDelegateCall) _execute(globalState *state.GlobalState) []*analysis.Issue {
	if dm.Cache.Contains(globalState.GetCurrentInstruction().Address) {
		return nil
	}
	potentialIssues := dm._analyze_state(globalState)
	annotation := GetPotentialIssuesAnnotaion(globalState)
	//annotation.PotentialIssues = append(annotation.PotentialIssues, potentialIssues...)
	annotation.Append(potentialIssues...)
	return nil
}

func (dm *ArbitraryDelegateCall) _analyze_state(globalState *state.GlobalState) []*PotentialIssue {
	//config := z3.GetConfig()
	//newCtx := z3.NewContext(config)

	length := globalState.Mstate.Stack.Length()
	gas := globalState.Mstate.Stack.RawStack[length-1]
	to := globalState.Mstate.Stack.RawStack[length-2]

	ctx := globalState.Z3ctx
	ACTORS := transaction.NewActors(ctx)

	constraints := state.NewConstraints()
	constraints.Add(to.Eq(ACTORS.GetAttacker()), gas.BvUGt(ctx.NewBitvecVal(2300, 256)),
		globalState.NewBitvec("retval_"+strconv.Itoa(globalState.GetCurrentInstruction().Address), 256).Eq(
			ctx.NewBitvecVal(1, 256)))
	for _, tx := range globalState.WorldState.TransactionSequence {
		if reflect.TypeOf(tx).String() == "*state.ContractCreationTransaction" {
			constraints.Add(tx.(*state.ContractCreationTransaction).Caller.Translate(ctx).Eq(ACTORS.GetAttacker()))
		}
	}
	address := globalState.GetCurrentInstruction().Address
	fmt.Println("[DELEGATECALL] Detected potential delegatecall to a user-supplied address :", address)
	descriptionHead := "The contract delegates execution to another contract with a user-supplied address."
	descriptionTail := "The smart contract delegates execution to a user-supplied address.This could allow an attacker to " +
		"execute arbitrary code in the context of this contract account and manipulate the state of the " +
		"contract account or execute actions on its behalf."
	issueArr := make([]*PotentialIssue, 0)
	potentialIssue := &PotentialIssue{
		Contract:        globalState.Environment.ActiveAccount.ContractName,
		FunctionName:    globalState.Environment.ActiveFuncName,
		Address:         address,
		SWCID:           analysis.NewSWCData()["DELEGATECALL_TO_UNTRUSTED_CONTRACT"],
		Bytecode:        globalState.Environment.Code.Bytecode,
		Title:           "Delegatecall to user-supplied address",
		Severity:        "High",
		DescriptionHead: descriptionHead,
		DescriptionTail: descriptionTail,
		Constraints:     constraints,
		//Constraints: constraints.Translate(newCtx),
		Detector: dm,
	}

	issueArr = append(issueArr, potentialIssue)
	return issueArr
}
