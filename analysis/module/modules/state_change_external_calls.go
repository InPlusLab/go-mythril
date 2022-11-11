package modules

import (
	"fmt"
	"go-mythril/analysis"
	"go-mythril/laser/ethereum/state"
	"go-mythril/laser/smt/z3"
	"go-mythril/utils"
	"math/big"
	"reflect"
	"strconv"
)

type StateChangeAfterCall struct {
	Name        string
	SWCID       string
	Description string
	PreHooks    []string
	Issues      *utils.SyncIssueSlice
	Cache       *utils.Set
}
type StateChangeCallsAnnotation struct {
	CallState          *state.GlobalState
	StateChangeStates  []*state.GlobalState
	UserDefinedAddress bool
}

func NewStateChangeCallsAnnotation(globalState *state.GlobalState, userDefinedAddress bool) *StateChangeCallsAnnotation {
	stateList := make([]*state.GlobalState, 0)
	return &StateChangeCallsAnnotation{
		CallState:          globalState,
		StateChangeStates:  stateList,
		UserDefinedAddress: userDefinedAddress,
	}
}

func (anno *StateChangeCallsAnnotation) PersistToWorldState() bool {
	return false
}
func (anno *StateChangeCallsAnnotation) PersistOverCalls() bool {
	return false
}
func (anno *StateChangeCallsAnnotation) Copy() state.StateAnnotation {
	stateChangeStatesNew := make([]*state.GlobalState, 0)
	for _, v := range anno.StateChangeStates {
		stateChangeStatesNew = append(stateChangeStatesNew, v)
	}
	return &StateChangeCallsAnnotation{
		CallState:          anno.CallState.Copy(),
		StateChangeStates:  stateChangeStatesNew,
		UserDefinedAddress: anno.UserDefinedAddress,
	}
}
func (anno *StateChangeCallsAnnotation) AppendState(globalState *state.GlobalState) {
	anno.StateChangeStates = append(anno.StateChangeStates, globalState)
	fmt.Println("appendState!")
}
func (anno *StateChangeCallsAnnotation) GetIssue(globalState *state.GlobalState, dm *StateChangeAfterCall) *PotentialIssue {
	if len(anno.StateChangeStates) == 0 {
		fmt.Println("AnnoGetIssue Len(States) == 0")
		return nil
	}
	constraints := state.NewConstraints()
	stackLen := anno.CallState.Mstate.Stack.Length()

	anno.CallState.Translate(globalState.Z3ctx)

	ctx := anno.CallState.Z3ctx
	gas := anno.CallState.Mstate.Stack.RawStack[stackLen-1]
	to := anno.CallState.Mstate.Stack.RawStack[stackLen-2]

	constraints.Add(gas.BvUGt(ctx.NewBitvecVal(2300, 256)),
		(to.BvSGt(ctx.NewBitvecVal(16, 256))).Or(to.Eq(ctx.NewBitvecVal(0, 256))))
	var severity string
	var addressType string
	if anno.UserDefinedAddress {
		tmpVal, _ := new(big.Int).SetString("DEADBEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEF", 16)
		// stmpV, _ := new(big.Int).SetString("DEADBEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEF", 16)
		constraints.Add(to.Eq(ctx.NewBitvecVal(tmpVal, 256)))
		severity = "Medium"
		addressType = "user defined"
	} else {
		severity = "Low"
		addressType = "fixed"
	}
	constraints.Add(globalState.WorldState.Constraints.ConstraintList...)
	fmt.Println("getIssues- GetTx")
	transactionSequence := analysis.GetTransactionSequence(globalState, constraints)
	if transactionSequence == nil {
		// UnsatError
		fmt.Println("stateChangeExternalCall unsat")
		return nil
	}
	address := globalState.GetCurrentInstruction().Address
	fmt.Println("[EXTERNAL_CALLS] Detected state changes at addresses: ", address)
	readOrWrite := "Write to"
	if globalState.GetCurrentInstruction().OpCode.Name == "SLOAD" {
		readOrWrite = "Read of"
	}
	descriptionHead := readOrWrite + " persistent state following external call"
	descriptionTail := "The contract account state is accessed after an external call to a " + addressType + " address." +
		"To prevent reentrancy issues, consider accessing the state only before the call, especially if the callee is untrusted. " +
		"Alternatively, a reentrancy lock can be used to prevent untrusted callees from re-entering the contract in an intermediate state."
	fmt.Println("stateChangeExternalCall push")
	return &PotentialIssue{
		Contract:        globalState.Environment.ActiveAccount.ContractName,
		FunctionName:    globalState.Environment.ActiveFuncName,
		Address:         address,
		Title:           "State access after external call",
		Severity:        severity,
		DescriptionHead: descriptionHead,
		DescriptionTail: descriptionTail,
		SWCID:           analysis.NewSWCData()["REENTRANCY"],
		Bytecode:        globalState.Environment.Code.Bytecode,
		Constraints:     constraints,
		Detector:        dm,
	}
}

func NewStateChangeAfterCall() *StateChangeAfterCall {
	return &StateChangeAfterCall{
		Name:        "State change after an external call",
		SWCID:       analysis.NewSWCData()["REENTRANCY"],
		Description: "Check whether the account state is accesses after the execution of an external call",
		// CALL_LIST = ["CALL", "DELEGATECALL", "CALLCODE"]
		// STATE_READ_WRITE_LIST = ["SSTORE", "SLOAD", "CREATE", "CREATE2"]
		PreHooks: []string{"CALL", "DELEGATECALL", "CALLCODE", "SSTORE", "SLOAD", "CREATE", "CREATE2"},
		Issues:   utils.NewSyncIssueSlice(),
		Cache:    utils.NewSet(),
	}
}

func (dm *StateChangeAfterCall) ResetModule() {
	dm.Issues = utils.NewSyncIssueSlice()
}
func (dm *StateChangeAfterCall) Execute(target *state.GlobalState) []*analysis.Issue {
	fmt.Println("Entering analysis module: ", dm.Name)
	result := dm._execute(target)
	fmt.Println("Exiting analysis module:", dm.Name)
	return result
}

func (dm *StateChangeAfterCall) AddIssue(issue *analysis.Issue) {
	dm.Issues.Append(issue)
}

func (dm *StateChangeAfterCall) GetIssues() []*analysis.Issue {
	list := make([]*analysis.Issue, 0)
	for _, v := range dm.Issues.Elements() {
		list = append(list, v.(*analysis.Issue))
	}
	return list
}

func (dm *StateChangeAfterCall) GetPreHooks() []string {
	return dm.PreHooks
}

func (dm *StateChangeAfterCall) GetPostHooks() []string {
	return make([]string, 0)
}

func (dm *StateChangeAfterCall) GetCache() *utils.Set {
	return dm.Cache
}

func (dm *StateChangeAfterCall) _execute(globalState *state.GlobalState) []*analysis.Issue {
	if dm.Cache.Contains(globalState.GetCurrentInstruction().Address) {
		return nil
	}
	issues := dm._analyze_state(globalState)
	annotation := GetPotentialIssuesAnnotaion(globalState)
	annotation.Append(issues...)
	return nil
}

func (dm *StateChangeAfterCall) _analyze_state(globalState *state.GlobalState) []*PotentialIssue {
	config := z3.GetConfig()
	newCtx := z3.NewContext(config)

	annotations := globalState.GetAnnotations(reflect.TypeOf(&StateChangeCallsAnnotation{}))
	opcode := globalState.GetCurrentInstruction().OpCode

	CALL_LIST := []string{"CALL", "DELEGATECALL", "CALLCODE"}
	STATE_READ_WRITE_LIST := []string{"SSTORE", "SLOAD", "CREATE", "CREATE2"}

	if len(annotations) == 0 && utils.In(opcode.Name, STATE_READ_WRITE_LIST) {
		fmt.Println("rwList: annotations == 0")
		return make([]*PotentialIssue, 0)
	}
	if utils.In(opcode.Name, STATE_READ_WRITE_LIST) {
		fmt.Println("rwList: annotations != 0")
		for _, annotation := range annotations {
			fmt.Println("beforeAppend:", len(annotation.(*StateChangeCallsAnnotation).StateChangeStates))
			fmt.Println("beforeAppend:", globalState.GetAnnotations(reflect.TypeOf(&StateChangeCallsAnnotation{}))[0])
			annotation.(*StateChangeCallsAnnotation).AppendState(globalState.Copy())
			fmt.Println("afterAppend:", len(annotation.(*StateChangeCallsAnnotation).StateChangeStates))
			fmt.Println("afterAppend:", globalState.GetAnnotations(reflect.TypeOf(&StateChangeCallsAnnotation{}))[0])
		}
	}
	// Record state changes following from a transfer of ether
	if utils.In(opcode.Name, CALL_LIST) {
		stackLen := globalState.Mstate.Stack.Length()
		value := globalState.Mstate.Stack.RawStack[stackLen-3]
		if dm._balance_change(value, globalState) {
			for _, annotation := range annotations {
				fmt.Println("beforeAppend:", len(annotation.(*StateChangeCallsAnnotation).StateChangeStates))
				annotation.(*StateChangeCallsAnnotation).AppendState(globalState.Copy())
				fmt.Println("afterAppend:", len(annotation.(*StateChangeCallsAnnotation).StateChangeStates))
			}
		}
	}
	// Record external calls
	if utils.In(opcode.Name, CALL_LIST) {
		dm._add_external_call(globalState)
	}
	// Check for vulnerabilities
	vulnerabilities := make([]*PotentialIssue, 0)
	for _, annotation := range annotations {
		if len(annotation.(*StateChangeCallsAnnotation).StateChangeStates) == 0 {
			fmt.Println("stateList == 0")
			continue
		}
		issue := annotation.(*StateChangeCallsAnnotation).GetIssue(globalState, dm)
		if issue != nil {
			issue.Constraints = issue.Constraints.Translate(newCtx)
			vulnerabilities = append(vulnerabilities, issue)
		}
	}
	return vulnerabilities
}

func (dm *StateChangeAfterCall) _add_external_call(globalState *state.GlobalState) {
	stackLen := globalState.Mstate.Stack.Length()
	gas := globalState.Mstate.Stack.RawStack[stackLen-1]
	to := globalState.Mstate.Stack.RawStack[stackLen-2]
	ctx := globalState.Z3ctx

	constraints := globalState.WorldState.Constraints.Copy()
	tmpCon := globalState.WorldState.Constraints.Copy()
	tmpCon.Add(gas.BvUGt(ctx.NewBitvecVal(2300, 256)),
		(to.BvSGt(ctx.NewBitvecVal(16, 256))).Or(to.Eq(ctx.NewBitvecVal(0, 256))))
	_, sat := state.GetModel(tmpCon, make([]*z3.Bool, 0), make([]*z3.Bool, 0), true, ctx)
	if !sat {
		return
	}
	// Check whether we can also set the callee address
	tmpVal, _ := new(big.Int).SetString("DEADBEEFDEADBEEFDEADBEEFDEADBEEFDEADBEEF", 16)
	constraints.Add(to.Eq(ctx.NewBitvecVal(tmpVal, 256)))
	_, sat2 := state.GetModel(constraints, make([]*z3.Bool, 0), make([]*z3.Bool, 0), true, ctx)
	if sat2 {
		globalState.Annotate(NewStateChangeCallsAnnotation(globalState.Copy(), true))
		//globalState.Annotate(NewStateChangeCallsAnnotation(globalState, true))
		fmt.Println("NewStateAnno: sat")
	} else {
		globalState.Annotate(NewStateChangeCallsAnnotation(globalState.Copy(), false))
		//globalState.Annotate(NewStateChangeCallsAnnotation(globalState, false))
		fmt.Println("NewStateAnno: unsat")
	}
}

func (dm *StateChangeAfterCall) _balance_change(value *z3.Bitvec, globalState *state.GlobalState) bool {
	//fmt.Println("value:", value.BvString())
	if !value.Symbolic() {
		v, _ := strconv.Atoi(value.Value())
		return v > 0
	} else {
		constraints := globalState.WorldState.Constraints.Copy()
		constraints.Add(value.BvSGt(globalState.Z3ctx.NewBitvecVal(0, 256)))
		_, sat := state.GetModel(constraints, make([]*z3.Bool, 0), make([]*z3.Bool, 0), true, globalState.Z3ctx)
		if sat {
			return true
		} else {
			return false
		}
	}
}
