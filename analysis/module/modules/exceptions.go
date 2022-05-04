package modules

import (
	"fmt"
	"go-mythril/analysis"
	"go-mythril/laser/ethereum/state"
	"go-mythril/utils"
	"reflect"
	"strconv"
)

type Exceptions struct {
	Name        string
	SWCID       string
	Description string
	PreHooks    []string
	Issues      []*analysis.Issue
	Cache       *utils.Set
}

type LastJumpAnnotation struct {
	LastJump int
}

func (anno LastJumpAnnotation) PersistToWorldState() bool {
	return false
}
func (anno LastJumpAnnotation) PersistOverCalls() bool {
	return false
}
func (anno LastJumpAnnotation) SetLastJump(a int) {
	anno.LastJump = a
}

func NewExceptions() *Exceptions {
	return &Exceptions{
		Name:        "Assertion violation",
		SWCID:       analysis.NewSWCData()["ASSERT_VIOLATION"],
		Description: "Checks whether any exception states are reachable.",
		PreHooks:    []string{"INVALID", "JUMP", "REVERT"},
		Issues:      make([]*analysis.Issue, 0),
		Cache:       utils.NewSet(),
	}
}

func (dm *Exceptions) ResetModule() {
	dm.Issues = make([]*analysis.Issue, 0)
}
func (dm *Exceptions) Execute(target *state.GlobalState) []*analysis.Issue {
	fmt.Println("Entering analysis module: ", dm.Name)
	result := dm._execute(target)
	fmt.Println("Exiting analysis module:", dm.Name)
	return result
}
func (dm *Exceptions) _execute(globalState *state.GlobalState) []*analysis.Issue {
	for _, v := range dm.Cache.Elements() {
		if reflect.TypeOf(v).String() == "map[int]string" {
			if v.(map[int]string)[globalState.GetCurrentInstruction().Address] == globalState.Environment.ActiveFuncName {
				return nil
			}
		}
	}
	issues := dm._analyze_state(globalState)
	for _, issue := range issues {
		addrFuncMap := make(map[int]string)
		addrFuncMap[issue.Address] = issue.FunctionName
		dm.Cache.Add(addrFuncMap)
	}
	dm.Issues = append(dm.Issues, issues...)
	return nil
}

func (dm *Exceptions) _analyze_state(globalState *state.GlobalState) []*analysis.Issue {
	opcode := globalState.GetCurrentInstruction().OpCode
	address := globalState.GetCurrentInstruction().Address
	annotations := globalState.GetAnnotations(reflect.TypeOf(LastJumpAnnotation{}))
	if len(annotations) == 0 {
		globalState.Annotate(LastJumpAnnotation{})
		annotations = globalState.GetAnnotations(reflect.TypeOf(LastJumpAnnotation{}))
	}
	if opcode.Name == "JUMP" {
		annotations[0].(LastJumpAnnotation).SetLastJump(address)
		return make([]*analysis.Issue, 0)
	}
	if opcode.Name == "REVERT" && !(is_assertion_failure(globalState)) {
		return make([]*analysis.Issue, 0)
	}
	fmt.Println("ASSERT_FAIL/REVERT in function " + globalState.Environment.ActiveFuncName)

	descriptionTail := "It is possible to trigger an assertion violation. Note that Solidity assert() statements should " +
		"only be used to check invariants. Review the transaction trace generated for this issue and " +
		"either make sure your program logic is correct, or use require() instead of assert() if your goal " +
		"is to constrain user inputs or enforce preconditions. Remember to validate inputs from both callers " +
		"(for instance, via passed arguments) and callees (for instance, via return values)."
	// TODO: TxSeq solver.getTxSeq
	issue := &analysis.Issue{
		Contract:        globalState.Environment.ActiveAccount.ContractName,
		FunctionName:    globalState.Environment.ActiveFuncName,
		Address:         address,
		SWCID:           analysis.NewSWCData()["ASSERT_VIOLATION"],
		Title:           "Exception State",
		Severity:        "Medium",
		DescriptionHead: "An assertion violation was triggered.",
		DescriptionTail: descriptionTail,
		Bytecode:        globalState.Environment.Code.Bytecode,
		// TxSeq
		GasUsed:        []int{globalState.Mstate.MinGasUsed, globalState.Mstate.MaxGasUsed},
		SourceLocation: annotations[0].(LastJumpAnnotation).LastJump,
	}
	return []*analysis.Issue{issue}
	// TODO: unsatError RETURN []
}

func is_assertion_failure(globalState *state.GlobalState) bool {
	state := globalState.Mstate
	stackLength := state.Stack.Length()
	offset := state.Stack.RawStack[stackLength-1]
	length := state.Stack.RawStack[stackLength-2]
	offsetV, _ := strconv.ParseInt(offset.Value(), 10, 64)
	lengthV, _ := strconv.ParseInt(length.Value(), 10, 64)
	returnData := state.Memory.GetItems(offsetV, offsetV+lengthV)
	// The function signature of Panic(uint256)
	PANIC_SIGNATURE := []int{78, 72, 123, 113}
	flag := true
	for i := 0; i < 4; i++ {
		val, _ := strconv.Atoi(returnData[i].Value())
		if val != PANIC_SIGNATURE[i] {
			flag = false
		}
	}
	return flag && returnData[len(returnData)-1].Value() == "1"
}
