package modules

import (
	"fmt"
	"go-mythril/analysis"
	"go-mythril/laser/ethereum/state"
	"go-mythril/utils"
	"reflect"
)

type MultipleSends struct {
	Name        string
	SWCID       string
	Description string
	PreHooks    []string
	Issues      []*analysis.Issue
	Cache       *utils.Set
}

type MultipleSendsAnnotation struct {
	CallOffsets []int
}

func (anno MultipleSendsAnnotation) PersistToWorldState() bool {
	return false
}
func (anno MultipleSendsAnnotation) PersistOverCalls() bool {
	return false
}

func NewMultipleSends() *MultipleSends {
	return &MultipleSends{
		Name:        "Multiple external calls in the same transaction",
		SWCID:       analysis.NewSWCData()["MULTIPLE_SENDS"],
		Description: "Check for multiple sends in a single transaction",
		PreHooks:    []string{"CALL", "DELEGATECALL", "STATICCALL", "CALLCODE", "RETURN", "STOP"},
		Issues:      make([]*analysis.Issue, 0),
		Cache:       utils.NewSet(),
	}
}
func (dm *MultipleSends) ResetModule() {
	dm.Issues = make([]*analysis.Issue, 0)
}
func (dm *MultipleSends) Execute(target *state.GlobalState) []*analysis.Issue {
	fmt.Println("Entering analysis module: ", dm.Name)
	result := dm._execute(target)
	fmt.Println("Exiting analysis module:", dm.Name)
	return result
}
func (dm *MultipleSends) _execute(globalState *state.GlobalState) []*analysis.Issue {
	if dm.Cache.Contains(globalState.GetCurrentInstruction().Address) {
		return nil
	}
	issues := dm._analyze_state(globalState)
	for _, issue := range issues {
		dm.Cache.Add(issue.Address)
	}
	dm.Issues = append(dm.Issues, issues...)
	return nil
}

func (dm *MultipleSends) _analyze_state(globalState *state.GlobalState) []*analysis.Issue {
	instruction := globalState.GetCurrentInstruction()
	annotations := globalState.GetAnnotations(reflect.TypeOf(MultipleSendsAnnotation{}))

	if len(annotations) == 0 {
		globalState.Annotate(MultipleSendsAnnotation{})
		annotations = globalState.GetAnnotations(reflect.TypeOf(MultipleSendsAnnotation{}))
	}
	callOffsets := annotations[0].(MultipleSendsAnnotation).CallOffsets

	if instruction.OpCode.Name == "CALL" || instruction.OpCode.Name == "DELEGATECALL" ||
		instruction.OpCode.Name == "STATICCALL" || instruction.OpCode.Name == "CALLCODE" {
		callOffsets = append(callOffsets, globalState.GetCurrentInstruction().Address)
	} else {
		// RETURN OR STOP
		for i := 1; i < len(callOffsets); i++ {
			//TODO: get_transaction_sequence   unsatError
			descriptionTail := "This call is executed following another call within the same transaction. It is possible " +
				"that the call never gets executed if a prior call fails permanently. This might be caused " +
				"intentionally by a malicious callee. If possible, refactor the code such that each transaction " +
				"only executes one external call or make sure that all callees can be trusted (i.e. they’re part of your own codebase)."
			issue := &analysis.Issue{
				Contract:        globalState.Environment.ActiveAccount.ContractName,
				FunctionName:    globalState.Environment.ActiveFuncName,
				Address:         callOffsets[i],
				SWCID:           analysis.NewSWCData()["MULTIPLE_SENDS"],
				Bytecode:        globalState.Environment.Code.Bytecode,
				Title:           "Multiple Calls in a Single Transaction",
				Severity:        "Low",
				DescriptionHead: "Multiple calls are executed in the same transaction.",
				DescriptionTail: descriptionTail,
				GasUsed:         []int{globalState.Mstate.MinGasUsed, globalState.Mstate.MaxGasUsed},
				// TxSeq
			}
			return []*analysis.Issue{issue}
		}
	}
	return make([]*analysis.Issue, 0)
}
