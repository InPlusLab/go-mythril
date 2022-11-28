package modules

import (
	"fmt"
	"go-mythril/analysis"
	"go-mythril/laser/ethereum/state"
	"go-mythril/laser/smt/z3"
	"go-mythril/utils"
	"reflect"
	"sync"
)

type MultipleSends struct {
	Name        string
	SWCID       string
	Description string
	PreHooks    []string
	Issues      *utils.SyncIssueSlice
	Cache       *utils.Set
}

type MultipleSendsAnnotation struct {
	IndexCounter int
	CallOffsets  sync.Map
}

func NewMultipleSendsAnnotation() *MultipleSendsAnnotation {
	return &MultipleSendsAnnotation{
		IndexCounter: 0,
		CallOffsets:  sync.Map{},
	}
}

func (anno *MultipleSendsAnnotation) PersistToWorldState() bool {
	return false
}
func (anno *MultipleSendsAnnotation) PersistOverCalls() bool {
	return false
}
func (anno *MultipleSendsAnnotation) Copy() state.StateAnnotation {
	var offsetList sync.Map
	for i, v := range anno.Elements() {
		offsetList.Store(i, v)
	}
	return &MultipleSendsAnnotation{
		IndexCounter: anno.IndexCounter,
		CallOffsets:  offsetList,
	}
}
func (anno *MultipleSendsAnnotation) Translate(ctx *z3.Context) state.StateAnnotation {
	return anno.Copy()
}
func (anno *MultipleSendsAnnotation) getIndex() int {
	anno.IndexCounter = anno.IndexCounter + 1
	return anno.IndexCounter
}
func (anno *MultipleSendsAnnotation) Add(callOffset int) bool {
	_, exist := anno.CallOffsets.LoadOrStore(anno.getIndex(), callOffset)
	return !exist
}
func (anno *MultipleSendsAnnotation) Elements() []int {
	res := make([]int, 0)
	anno.CallOffsets.Range(func(k, v interface{}) bool {
		res = append(res, v.(int))
		return true
	})
	return res
}
func (anno *MultipleSendsAnnotation) Len() int {
	return len(anno.Elements())
}

func NewMultipleSends() *MultipleSends {
	return &MultipleSends{
		Name:        "Multiple external calls in the same transaction",
		SWCID:       analysis.NewSWCData()["MULTIPLE_SENDS"],
		Description: "Check for multiple sends in a single transaction",
		PreHooks:    []string{"CALL", "DELEGATECALL", "STATICCALL", "CALLCODE", "RETURN", "STOP"},
		Issues:      utils.NewSyncIssueSlice(),
		Cache:       utils.NewSet(),
	}
}
func (dm *MultipleSends) ResetModule() {
	dm.Issues = utils.NewSyncIssueSlice()
}
func (dm *MultipleSends) Execute(target *state.GlobalState) []*analysis.Issue {
	fmt.Println("Entering analysis module: ", dm.Name)
	result := dm._execute(target)
	fmt.Println("Exiting analysis module:", dm.Name)
	return result
}

func (dm *MultipleSends) AddIssue(issue *analysis.Issue) {
	dm.Issues.Append(issue)
}

func (dm *MultipleSends) GetIssues() []*analysis.Issue {
	list := make([]*analysis.Issue, 0)
	for _, v := range dm.Issues.Elements() {
		list = append(list, v.(*analysis.Issue))
	}
	return list
}

func (dm *MultipleSends) GetPreHooks() []string {
	return dm.PreHooks
}

func (dm *MultipleSends) GetPostHooks() []string {
	return make([]string, 0)
}

func (dm *MultipleSends) GetCache() *utils.Set {
	return dm.Cache
}

func (dm *MultipleSends) _execute(globalState *state.GlobalState) []*analysis.Issue {
	if dm.Cache.Contains(globalState.GetCurrentInstruction().Address) {
		return nil
	}
	issues := dm._analyze_state(globalState)
	for _, issue := range issues {
		dm.Cache.Add(issue.Address)
	}
	for _, issue := range issues {
		fmt.Println("multipleSends push")
		dm.Issues.Append(issue)
	}
	return nil
}

func (dm *MultipleSends) _analyze_state(globalState *state.GlobalState) []*analysis.Issue {
	instruction := globalState.GetCurrentInstruction()
	annotations := globalState.GetAnnotations(reflect.TypeOf(&MultipleSendsAnnotation{}))

	if len(annotations) == 0 {
		globalState.Annotate(NewMultipleSendsAnnotation())
		annotations = globalState.GetAnnotations(reflect.TypeOf(&MultipleSendsAnnotation{}))
	}
	callOffsets := annotations[0].(*MultipleSendsAnnotation)

	if instruction.OpCode.Name == "CALL" || instruction.OpCode.Name == "DELEGATECALL" ||
		instruction.OpCode.Name == "STATICCALL" || instruction.OpCode.Name == "CALLCODE" {
		callOffsets.Add(globalState.GetCurrentInstruction().Address)
	} else {
		// RETURN OR STOP
		for i, v := range callOffsets.Elements() {
			if i == 0 {
				continue
			}

			fmt.Println("MultipleSends:")

			transactionSequence := analysis.GetTransactionSequence(globalState, globalState.WorldState.Constraints)
			if transactionSequence == nil {
				// UnsatError
				continue
			}
			descriptionTail := "This call is executed following another call within the same transaction. It is possible " +
				"that the call never gets executed if a prior call fails permanently. This might be caused " +
				"intentionally by a malicious callee. If possible, refactor the code such that each transaction " +
				"only executes one external call or make sure that all callees can be trusted (i.e. they’re part of your own codebase)."
			issue := &analysis.Issue{
				Contract:            globalState.Environment.ActiveAccount.ContractName,
				FunctionName:        globalState.Environment.ActiveFuncName,
				Address:             v,
				SWCID:               analysis.NewSWCData()["MULTIPLE_SENDS"],
				Bytecode:            globalState.Environment.Code.Bytecode,
				Title:               "Multiple Calls in a Single Transaction",
				Severity:            "Low",
				DescriptionHead:     "Multiple calls are executed in the same transaction.",
				DescriptionTail:     descriptionTail,
				GasUsed:             []int{globalState.Mstate.MinGasUsed, globalState.Mstate.MaxGasUsed},
				TransactionSequence: transactionSequence,
			}
			return []*analysis.Issue{issue}
		}
	}
	return make([]*analysis.Issue, 0)
}
