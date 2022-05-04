package modules

import (
	"fmt"
	"go-mythril/analysis"
	"go-mythril/laser/ethereum/state"
	"go-mythril/utils"
	"math/big"
)

type UserAssertions struct {
	Name        string
	SWCID       string
	Description string
	PreHooks    []string
	Issues      []*analysis.Issue
	Cache       *utils.Set
}

func NewUserAssertions() *UserAssertions {
	return &UserAssertions{
		Name:        "A user-defined assertion has been triggered",
		SWCID:       analysis.NewSWCData()["ASSERT_VIOLATION"],
		Description: "Search for reachable user-supplied exceptions. Report a warning if an log message is emitted: 'emit AssertionFailed(string)",
		PreHooks:    []string{"LOG1", "MSTORE"},
		Issues:      make([]*analysis.Issue, 0),
		Cache:       utils.NewSet(),
	}
}

func (dm *UserAssertions) ResetModule() {
	dm.Issues = make([]*analysis.Issue, 0)
}
func (dm *UserAssertions) Execute(target *state.GlobalState) []*analysis.Issue {
	fmt.Println("Entering analysis module: ", dm.Name)
	result := dm._execute(target)
	fmt.Println("Exiting analysis module:", dm.Name)
	return result
}
func (dm *UserAssertions) _execute(globalState *state.GlobalState) []*analysis.Issue {
	issues := dm._analyze_state(globalState)
	for _, issue := range issues {
		dm.Cache.Add(issue.Address)
	}
	dm.Issues = append(dm.Issues, issues...)
	return nil
}

func (dm *UserAssertions) _analyze_state(globalState *state.GlobalState) []*analysis.Issue {
	opcode := globalState.GetCurrentInstruction().OpCode
	var message string
	stackLen := globalState.Mstate.Stack.Length()

	if opcode.Name == "MSTORE" {
		value := globalState.Mstate.Stack.RawStack[stackLen-2]
		if value.Symbolic() {
			return make([]*analysis.Issue, 0)
		}
		// mstore_pattern := new(big.Int).SetString("cafecafecafecafecafecafecafecafecafecafecafecafecafecafecafe",16)
		// TODO:
		message = "Failed property id" + value.Extract(15, 0).Value()
	} else {
		topic := globalState.Mstate.Stack.RawStack[stackLen-3]
		size := globalState.Mstate.Stack.RawStack[stackLen-2]
		memStart := globalState.Mstate.Stack.RawStack[stackLen-1]

		assertion_failed_hash, _ := new(big.Int).SetString("B42604CB105A16C8F6DB8A41E6B00C0C1B4826465E8BC504B3EB3E88B3E6A4A0", 16)
		if topic.Symbolic() || topic.Value() != assertion_failed_hash.String() {
			return make([]*analysis.Issue, 0)
		}
		if !memStart.Symbolic() && !size.Symbolic() {
			// TODO: eth_abi
			message = "eth_abi"
		}
	}
	// TODO: solver.getTxSeq
	var descriptionTail string
	if message != "" {
		descriptionTail = "A user-provided assertion failed with the message " + message
	} else {
		descriptionTail = "A user-provided assertion failed."
	}
	fmt.Println("MythX assertion emitted:" + descriptionTail)
	address := globalState.GetCurrentInstruction().Address
	issue := &analysis.Issue{
		Contract:        globalState.Environment.ActiveAccount.ContractName,
		FunctionName:    globalState.Environment.ActiveFuncName,
		Address:         address,
		SWCID:           analysis.NewSWCData()["ASSERT_VIOLATION"],
		Title:           "Exception State",
		Severity:        "Medium",
		DescriptionHead: "A user-provided assertion failed.",
		DescriptionTail: descriptionTail,
		Bytecode:        globalState.Environment.Code.Bytecode,
		// TxSeq
		GasUsed: []int{globalState.Mstate.MinGasUsed, globalState.Mstate.MaxGasUsed},
	}
	return []*analysis.Issue{issue}
	// TODO: unsatError RETURN []
}
