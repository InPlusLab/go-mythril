package modules

import (
	"fmt"
	"go-mythril/analysis"
	"go-mythril/laser/ethereum/state"
	"go-mythril/laser/smt/z3"
	"go-mythril/utils"
	"reflect"
)

type RetVal struct {
	Address int
	Retval  *z3.Bitvec
}
type UncheckedRetvalAnnotation struct {
	//RetVals []*RetVal
	RetVals *utils.Set
}

func (anno UncheckedRetvalAnnotation) PersistToWorldState() bool {
	return false
}
func (anno UncheckedRetvalAnnotation) PersistOverCalls() bool {
	return false
}
func (anno UncheckedRetvalAnnotation) Copy() state.StateAnnotation {
	return UncheckedRetvalAnnotation{
		RetVals: anno.RetVals.Copy(),
	}
}

type UncheckedRetval struct {
	Name        string
	SWCID       string
	Description string
	PreHooks    []string
	PostHooks   []string
	Issues      *utils.SyncIssueSlice
	Cache       *utils.Set
}

func NewUncheckedRetval() *UncheckedRetval {
	return &UncheckedRetval{
		Name:  "Return value of an external call is not checked",
		SWCID: analysis.NewSWCData()["UNCHECKED_RET_VAL"],
		Description: "Test whether CALL return value is checked. For direct calls, the Solidity compiler auto-generates this check. E.g.:\\n" +
			"   Alice c = Alice(address);\\n" +
			"   c.ping(42);\\n" +
			"\"Here the CALL will be followed by IZSERO(retval), if retval = ZERO then state is reverted. " +
			"For low-level-calls this check is omitted. E.g.:\\n" +
			"    c.call.value(0)(bytes4(sha3(\"ping(uint256)\")),1)",
		PreHooks:  []string{"STOP", "RETURN"},
		PostHooks: []string{"CALL", "DELEGATECALL", "STATICCALL", "CALLCODE"},
		Issues:    utils.NewSyncIssueSlice(),
		Cache:     utils.NewSet(),
	}
}

func (dm *UncheckedRetval) ResetModule() {
	dm.Issues = utils.NewSyncIssueSlice()
}
func (dm *UncheckedRetval) Execute(target *state.GlobalState) []*analysis.Issue {
	fmt.Println("Entering analysis module: ", dm.Name)
	result := dm._execute(target)
	fmt.Println("Exiting analysis module:", dm.Name)
	return result
}

func (dm *UncheckedRetval) AddIssue(issue *analysis.Issue) {
	dm.Issues.Append(issue)
}

func (dm *UncheckedRetval) GetIssues() []*analysis.Issue {
	list := make([]*analysis.Issue, 0)
	for _, v := range dm.Issues.Elements() {
		list = append(list, v.(*analysis.Issue))
	}
	return list
}

func (dm *UncheckedRetval) GetPreHooks() []string {
	return dm.PreHooks
}

func (dm *UncheckedRetval) GetPostHooks() []string {
	return dm.PostHooks
}

func (dm *UncheckedRetval) GetCache() *utils.Set {
	return dm.Cache
}

func (dm *UncheckedRetval) _execute(globalState *state.GlobalState) []*analysis.Issue {
	if dm.Cache.Contains(globalState.GetCurrentInstruction().Address) {
		return nil
	}
	issues := dm._analyze_state(globalState)
	for _, issue := range issues {
		dm.Cache.Add(issue.Address)
	}
	for _, issue := range issues {
		fmt.Println("uncheckedRetval push")
		dm.Issues.Append(issue)
	}
	return nil
}

func (dm *UncheckedRetval) _analyze_state(globalState *state.GlobalState) []*analysis.Issue {
	instruction := globalState.GetCurrentInstruction()
	annotations := globalState.GetAnnotations(reflect.TypeOf(UncheckedRetvalAnnotation{}))
	if len(annotations) == 0 {
		globalState.Annotate(UncheckedRetvalAnnotation{
			RetVals: utils.NewSet(),
		})
		annotations = globalState.GetAnnotations(reflect.TypeOf(UncheckedRetvalAnnotation{}))
	}
	retvals := annotations[0].(UncheckedRetvalAnnotation).RetVals

	if instruction.OpCode.Name == "STOP" || instruction.OpCode.Name == "RETURN" {
		issues := make([]*analysis.Issue, 0)

		for _, retval := range retvals.Elements() {
			txCon := globalState.WorldState.Constraints.Copy()
			txCon.Add(retval.(*RetVal).Retval.Translate(globalState.Z3ctx).Eq(globalState.Z3ctx.NewBitvecVal(1, 256)))
			tx := analysis.GetTransactionSequence(globalState, txCon)
			tmpCon := globalState.WorldState.Constraints.Copy()
			tmpCon.Add(retval.(*RetVal).Retval.Translate(globalState.Z3ctx).Eq(globalState.Z3ctx.NewBitvecVal(0, 256)))
			transactionSequence := analysis.GetTransactionSequence(globalState, tmpCon)
			if tx == nil || transactionSequence == nil {
				// UnsatError
				continue
			}

			descriptionTail := "External calls return a boolean value. If the callee halts with an exception, 'false' is " +
				"returned and execution continues in the caller. " +
				"The caller should check whether an exception happened and react accordingly to avoid unexpected behavior. " +
				"For example it is often desirable to wrap external calls in require() so the transaction is reverted if the call fails."
			issue := &analysis.Issue{
				Contract:            globalState.Environment.ActiveAccount.ContractName,
				FunctionName:        globalState.Environment.ActiveFuncName,
				Address:             retval.(*RetVal).Address,
				Bytecode:            globalState.Environment.Code.Bytecode,
				Title:               "Unchecked return value from external call.",
				SWCID:               analysis.NewSWCData()["UNCHECKED_RET_VAL"],
				Severity:            "Medium",
				DescriptionHead:     "The return value of a message call is not checked.",
				DescriptionTail:     descriptionTail,
				GasUsed:             []int{globalState.Mstate.MinGasUsed, globalState.Mstate.MaxGasUsed},
				TransactionSequence: transactionSequence,
			}
			issues = append(issues, issue)
		}
		return issues
	} else {
		fmt.Println("End of call, extracting retval")
		opArr := []string{"CALL", "DELEGATECALL", "STATICCALL", "CALLCODE"}
		opcodeName := globalState.Environment.Code.InstructionList[globalState.Mstate.Pc-1].OpCode.Name
		if !utils.In(opcodeName, opArr) {
			panic("error! In unchecked_retval analyzeState method!")
		}
		returnValue := globalState.Mstate.Stack.RawStack[globalState.Mstate.Stack.Length()-1]

		retvals.Add(&RetVal{
			Address: instruction.Address - 1,
			Retval:  returnValue,
		})
	}
	return make([]*analysis.Issue, 0)
}
