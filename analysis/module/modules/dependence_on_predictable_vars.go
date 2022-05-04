package modules

import (
	"fmt"
	"go-mythril/analysis"
	"go-mythril/laser/ethereum/state"
	"go-mythril/utils"
	"reflect"
	"strings"
)

type PredictableVariables struct {
	Name        string
	SWCID       string
	Description string
	PreHooks    []string
	PostHooks   []string
	Issues      []*analysis.Issue
	Cache       *utils.Set
}
type PredictableValueAnnotation struct {
	// Symbol annotation used if a variable is initialized from a predictable environment variable.
	Operation string
}
type OldBlockNumberUsedAnnotation struct {
	// Symbol annotation used if a variable is initialized from a predictable environment variable.
}

func (anno OldBlockNumberUsedAnnotation) PersistToWorldState() bool {
	return false
}
func (anno OldBlockNumberUsedAnnotation) PersistOverCalls() bool {
	return false
}

func NewPredictableVariables() *PredictableVariables {
	swcData := analysis.NewSWCData()
	return &PredictableVariables{
		Name:  "Control flow depends on a predictable environment variable",
		SWCID: swcData["TIMESTAMP_DEPENDENCE"] + " " + swcData["WEAK_RANDOMNESS"],
		Description: "Check whether control flow decisions are influenced by block.coinbase," +
			"block.gaslimit, block.timestamp or block.number.",
		PreHooks:  []string{"JUMPI", "BLOCKHASH"},
		PostHooks: []string{"BLOCKHASH", "COINBASE", "GASLIMIT", "TIMESTAMP", "NUMBER"},
		Issues:    make([]*analysis.Issue, 0),
		Cache:     utils.NewSet(),
	}
}
func (dm *PredictableVariables) ResetModule() {
	dm.Issues = make([]*analysis.Issue, 0)
}
func (dm *PredictableVariables) Execute(target *state.GlobalState) []*analysis.Issue {
	fmt.Println("Entering analysis module: ", dm.Name)
	result := dm._execute(target)
	fmt.Println("Exiting analysis module:", dm.Name)
	return result
}
func (dm *PredictableVariables) _execute(globalState *state.GlobalState) []*analysis.Issue {
	if dm.Cache.Contains(globalState.GetCurrentInstruction().Address) {
		return nil
	}
	issues := dm._analyze_state(globalState)
	for _, issue := range issues {
		dm.Cache.Add(issue)
	}
	dm.Issues = append(dm.Issues, issues...)
	return nil
}

func (dm *PredictableVariables) _analyze_state(globalState *state.GlobalState) []*analysis.Issue {
	issues := make([]*analysis.Issue, 0)

	// TODO: isPrehook() ?
	isPrehook := true
	if isPrehook {
		opcode := globalState.GetCurrentInstruction().OpCode
		length := globalState.Mstate.Stack.Length()
		if opcode.Name == "JUMPI" {
			// Look for predictable state variables in jump condition
			for _, annotation := range globalState.Mstate.Stack.RawStack[length-2].Annotations().Elements() {
				if reflect.TypeOf(annotation).String() == "PredictableValueAnnotation" {
					// constraints := globalState.WorldState.Constraints
					// TODO: solver.getTxSeq
					description := annotation.(PredictableValueAnnotation).Operation + " is used to determine a control flow decision." +
						"Note that the values of variables like coinbase, gaslimit, block number and timestamp are " +
						"predictable and can be manipulated by a malicious miner. Also keep in mind that " +
						"attackers know hashes of earlier blocks. Don't use any of those environment variables " +
						"as sources of randomness and be aware that use of these variables introduces a certain level of trust into miners."
					severity := "Low"
					var swcId string
					if strings.Contains(annotation.(PredictableValueAnnotation).Operation, "timestamp") {
						swcId = analysis.NewSWCData()["TIMESTAMP_DEPENDENCE"]
					} else {
						swcId = analysis.NewSWCData()["WEAK_RANDOMNESS"]
					}
					issue := &analysis.Issue{
						Contract:        globalState.Environment.ActiveAccount.ContractName,
						FunctionName:    globalState.Environment.ActiveFuncName,
						Address:         globalState.GetCurrentInstruction().Address,
						SWCID:           swcId,
						Bytecode:        globalState.Environment.Code.Bytecode,
						Title:           "Dependence on predictable environment variable",
						Severity:        severity,
						DescriptionHead: "A control flow decision is made based on " + annotation.(PredictableValueAnnotation).Operation,
						DescriptionTail: description,
						GasUsed:         []int{globalState.Mstate.MinGasUsed, globalState.Mstate.MaxGasUsed},
						// txSeq
					}
					issues = append(issues, issue)
				}
			}
		} else if opcode.Name == "BLOCKHASH" {
			// TODO: solver.get_model
			globalState.Annotate(OldBlockNumberUsedAnnotation{})
		}
	} else {
		opcode := globalState.Environment.Code.InstructionList[globalState.Mstate.Pc-1].OpCode
		length := globalState.Mstate.Stack.Length()
		if opcode.Name == "BLOCKHASH" {
			// if we're in the post hook of a BLOCKHASH op, check if an old block number was used to create it.
			annotations := globalState.GetAnnotations(reflect.TypeOf(OldBlockNumberUsedAnnotation{}))
			if len(annotations) != 0 {
				globalState.Mstate.Stack.RawStack[length-1].Annotate(PredictableValueAnnotation{
					Operation: "The block hash of a previous block",
				})
			}
		} else {
			// Always create an annotation when COINBASE, GASLIMIT, TIMESTAMP or NUMBER is executed.
			globalState.Mstate.Stack.RawStack[length-1].Annotate(PredictableValueAnnotation{
				Operation: "The blokc " + opcode.Name + " environment variable",
			})
		}
	}
	return issues
}
