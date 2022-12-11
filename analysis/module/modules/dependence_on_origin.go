package modules

import (
	"fmt"
	"go-mythril/analysis"
	"go-mythril/laser/ethereum/state"
	"go-mythril/utils"
	"reflect"
)

type TxOrigin struct {
	Name        string
	SWCID       string
	Description string
	PreHooks    []string
	PostHooks   []string
	Issues      *utils.SyncSlice
	Cache       *utils.Set
}
type TxOriginAnnotation struct {
	//Symbol annotation added to a variable
	//that is initialized with a call to the ORIGIN instruction.
}

func NewTxOrigin() *TxOrigin {
	return &TxOrigin{
		Name:        "Control flow depends on tx.origin",
		SWCID:       analysis.NewSWCData()["TX_ORIGIN_USAGE"],
		Description: "Check whether control flow decisions are influenced by tx.origin",
		PreHooks:    []string{"JUMPI"},
		PostHooks:   []string{"ORIGIN"},
		Issues:      utils.NewSyncSlice(),
		Cache:       utils.NewSet(),
	}
}

func (dm *TxOrigin) ResetModule() {
	dm.Issues = utils.NewSyncSlice()
}
func (dm *TxOrigin) Execute(target *state.GlobalState) []*analysis.Issue {
	fmt.Println("Entering analysis module: ", dm.Name)
	result := dm._execute(target)
	fmt.Println("Exiting analysis module:", dm.Name)
	return result
}

func (dm *TxOrigin) AddIssue(issue *analysis.Issue) {
	dm.Issues.Append(issue)
}

func (dm *TxOrigin) GetIssues() []*analysis.Issue {
	list := make([]*analysis.Issue, 0)
	for _, v := range dm.Issues.Elements() {
		list = append(list, v.(*analysis.Issue))
	}
	return list
}

func (dm *TxOrigin) GetPreHooks() []string {
	return dm.PreHooks
}

func (dm *TxOrigin) GetPostHooks() []string {
	return dm.PostHooks
}

func (dm *TxOrigin) GetCache() *utils.Set {
	return dm.Cache
}

func (dm *TxOrigin) _execute(globalState *state.GlobalState) []*analysis.Issue {
	if dm.Cache.Contains(globalState.GetCurrentInstruction().Address) {
		return nil
	}
	issues := dm._analyze_state(globalState)
	for _, issue := range issues {
		dm.Cache.Add(issue.Address)
	}
	for _, issue := range issues {
		dm.Issues.Append(issue)
	}
	return nil
}

func (dm *TxOrigin) _analyze_state(globalState *state.GlobalState) []*analysis.Issue {
	issues := make([]*analysis.Issue, 0)

	if globalState.GetCurrentInstruction().OpCode.Name == "JUMPI" {
		// In JUMPI prehook
		length := globalState.Mstate.Stack.Length()
		for _, annotation := range globalState.Mstate.Stack.RawStack[length-2].Annotations.Elements() {
			if reflect.TypeOf(annotation).String() == "modules.TxOriginAnnotation" {
				constraints := globalState.WorldState.Constraints.Copy()

				transactionSequence := analysis.GetTransactionSequence(globalState, constraints)
				if transactionSequence == nil {
					// UnsatError
					fmt.Println("unsaterror for getTxSeq")
					continue
				}
				description := "The tx.origin environment variable has been found to influence a control flow decision. " +
					"Note that using tx.origin as a security control might cause a situation where a user " +
					"inadvertently authorizes a smart contract to perform an action on their behalf. It is " +
					"recommended to use msg.sender instead."
				severity := "Low"
				issue := &analysis.Issue{
					Contract:            globalState.Environment.ActiveAccount.ContractName,
					FunctionName:        globalState.Environment.ActiveFuncName,
					Address:             globalState.GetCurrentInstruction().Address,
					SWCID:               analysis.NewSWCData()["TX_ORIGIN_USAGE"],
					Bytecode:            globalState.Environment.Code.Bytecode,
					Title:               "Dependence on tx.origin",
					Severity:            severity,
					DescriptionHead:     "Use of tx.origin as a part of authorization control.",
					DescriptionTail:     description,
					GasUsed:             []int{globalState.Mstate.MinGasUsed, globalState.Mstate.MaxGasUsed},
					TransactionSequence: transactionSequence,
				}
				issues = append(issues, issue)
			}
		}
	} else {
		// In ORIGIN posthook
		length := globalState.Mstate.Stack.Length()
		globalState.Mstate.Stack.RawStack[length-1].Annotate(TxOriginAnnotation{})
	}
	return issues
}
