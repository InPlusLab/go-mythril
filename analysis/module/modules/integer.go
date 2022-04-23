package modules

import (
	"fmt"
	"go-mythril/analysis"
	"go-mythril/laser/ethereum/state"
	"go-mythril/laser/smt/z3"
	"go-mythril/utils"
	"math"
	"reflect"
	"strconv"
)

type handelFunc func(globalState *state.GlobalState)

type OverUnderflowAnnotation struct {
	// Symbol Annotation used if a BitVector can overflow
	OverflowingState *state.GlobalState
	Operator         string
	Constraint       *z3.Bool
}

type OverUnderflowStateAnnotation struct {
	OverflowingStateAnnotations *utils.Set
}

func NewOverUnderflowStateAnnotation() OverUnderflowStateAnnotation {
	return OverUnderflowStateAnnotation{
		OverflowingStateAnnotations: utils.NewSet(),
	}
}
func (anno OverUnderflowStateAnnotation) PersistToWorldState() bool {
	return false
}
func (anno OverUnderflowStateAnnotation) PersistOverCalls() bool {
	return false
}

type IntegerArithmetics struct {
	Name                 string
	SWCID                string
	Description          string
	PreHooks             []string
	Issues               []*analysis.Issue
	Cache                *utils.Set
	OstatesSatisfiable   *utils.Set
	OstatesUnsatisfiable *utils.Set
}

func NewIntegerArithmetics() *IntegerArithmetics {
	return &IntegerArithmetics{
		Name:  "Integer overflow or underflow",
		SWCID: analysis.NewSWCData()["INTEGER_OVERFLOW_AND_UNDERFLOW"],
		Description: "For every SUB instruction, " +
			"check if there's a possible state where op1 > op0. " +
			"For every ADD, MUL instruction, " +
			"check if there's a possible state where op1 + op0 > 2^32 - 1",
		PreHooks: []string{"ADD", "MUL", "EXP", "SUB", "SSTORE",
			"JUMPI", "STOP", "RETURN", "CALL"},
		Issues:               make([]*analysis.Issue, 0),
		Cache:                utils.NewSet(),
		OstatesSatisfiable:   utils.NewSet(),
		OstatesUnsatisfiable: utils.NewSet(),
	}
}

func (dm *IntegerArithmetics) ResetModule() {
	dm.Issues = make([]*analysis.Issue, 0)
	dm.OstatesSatisfiable = utils.NewSet()
	dm.OstatesUnsatisfiable = utils.NewSet()
}

func (dm *IntegerArithmetics) Execute(target *state.GlobalState) []*analysis.Issue {
	fmt.Println("Entering analysis module: ", dm.Name)
	result := dm._execute(target)
	fmt.Println("Exiting analysis module:", dm.Name)
	return result
}

func (dm *IntegerArithmetics) _get_args(state *state.GlobalState) (*z3.Bitvec, *z3.Bitvec) {
	stack := state.Mstate.Stack
	op0 := stack.Pop()
	op1 := stack.Pop()
	return op0, op1
}

func (dm *IntegerArithmetics) _execute(globalState *state.GlobalState) []*analysis.Issue {

	fmt.Println("IntegerArithmetics - _execute!")
	address := getAddressFromState(globalState)
	if dm.Cache.Contains(address) {
		return dm.Issues
	}
	opcode := globalState.GetCurrentInstruction().OpCode.Name
	funcs := make(map[string][]handelFunc)
	funcs["ADD"] = []handelFunc{dm._handel_add}
	funcs["SUB"] = []handelFunc{dm._handel_sub}
	funcs["MUL"] = []handelFunc{dm._handel_mul}
	funcs["SSTORE"] = []handelFunc{dm._handel_sstore}
	funcs["JUMPI"] = []handelFunc{dm._handel_jumpi}
	funcs["CALL"] = []handelFunc{dm._handel_call}
	funcs["RETURN"] = []handelFunc{dm._handel_return, dm._handel_transaction_end}
	funcs["STOP"] = []handelFunc{dm._handel_transaction_end}
	funcs["EXP"] = []handelFunc{dm._handel_exp}

	for _, f := range funcs[opcode] {
		f(globalState)
	}

	return dm.Issues
}

func (dm *IntegerArithmetics) _handel_add(globalState *state.GlobalState) {
	op0, op1 := dm._get_args(globalState)
	c := op0.BvAddNoOverflow(op1, false).Not()
	annotation := OverUnderflowAnnotation{
		OverflowingState: globalState,
		Operator:         "addition",
		Constraint:       c,
	}
	op0.Annotate(annotation)
}
func (dm *IntegerArithmetics) _handel_mul(globalState *state.GlobalState) {
	op0, op1 := dm._get_args(globalState)
	c := op0.BvMulNoOverflow(op1, false).Not()
	annotation := OverUnderflowAnnotation{
		OverflowingState: globalState,
		Operator:         "multiplication",
		Constraint:       c,
	}
	op0.Annotate(annotation)
}
func (dm *IntegerArithmetics) _handel_sub(globalState *state.GlobalState) {
	op0, op1 := dm._get_args(globalState)
	c := op0.BvSubNoUnderflow(op1, false).Not()
	annotation := OverUnderflowAnnotation{
		OverflowingState: globalState,
		Operator:         "subtraction",
		Constraint:       c,
	}
	op0.Annotate(annotation)
}
func (dm *IntegerArithmetics) _handel_exp(globalState *state.GlobalState) {
	op0, op1 := dm._get_args(globalState)
	ctx := op0.GetCtx()

	op0V, _ := strconv.Atoi(op0.Value())
	op1V, _ := strconv.Atoi(op1.Value())

	if (!op1.Symbolic() && op1V == 0) || (!op0.Symbolic() && op0V < 2) {
		return
	}

	var constraint *z3.Bool
	if op0.Symbolic() && op1.Symbolic() {
		constraint = op1.BvSGt(ctx.NewBitvecVal(256, 256)).And(
			op0.BvSGt(ctx.NewBitvecVal(1, 256)))
	} else if op0.Symbolic() {
		constraint = op0.BvSGe(ctx.NewBitvecVal(int64(math.Pow(2, math.Ceil(256/float64(op1V)))), 256))
	} else {
		constraint = op1.BvSGe(ctx.NewBitvecVal(int64(math.Ceil(256/math.Log2(float64(op0V)))), 256))
	}
	annotation := OverUnderflowAnnotation{
		OverflowingState: globalState,
		Operator:         "exponentiation",
		Constraint:       constraint,
	}
	op0.Annotate(annotation)
}

func (dm *IntegerArithmetics) _handel_sstore(globalState *state.GlobalState) {
	stack := globalState.Mstate.Stack
	value := stack.RawStack[stack.Length()-2]
	stateAnnotation := getOverflowUnderflowStateAnnotation(globalState)
	for _, annotation := range value.Annotations().Elements() {
		if reflect.TypeOf(annotation).String() == "OverUnderflowAnnotation" {
			stateAnnotation.OverflowingStateAnnotations.Add(annotation)
		}
	}
}

func (dm *IntegerArithmetics) _handel_jumpi(globalState *state.GlobalState) {
	stack := globalState.Mstate.Stack
	value := stack.RawStack[stack.Length()-2]
	stateAnnotation := getOverflowUnderflowStateAnnotation(globalState)
	for _, annotation := range value.Annotations().Elements() {
		if reflect.TypeOf(annotation).String() == "OverUnderflowAnnotation" {
			stateAnnotation.OverflowingStateAnnotations.Add(annotation)
		}
	}
}

func (dm *IntegerArithmetics) _handel_call(globalState *state.GlobalState) {
	stack := globalState.Mstate.Stack
	value := stack.RawStack[stack.Length()-3]
	stateAnnotation := getOverflowUnderflowStateAnnotation(globalState)
	for _, annotation := range value.Annotations().Elements() {
		if reflect.TypeOf(annotation).String() == "OverUnderflowAnnotation" {
			stateAnnotation.OverflowingStateAnnotations.Add(annotation)
		}
	}
}

func (dm *IntegerArithmetics) _handel_return(globalState *state.GlobalState) {
	stack := globalState.Mstate.Stack
	offset := stack.RawStack[stack.Length()-1]
	length := stack.RawStack[stack.Length()-2]
	offsetV, _ := strconv.ParseInt(offset.Value(), 10, 64)
	lengthV, _ := strconv.ParseInt(length.Value(), 10, 64)

	stateAnnotation := getOverflowUnderflowStateAnnotation(globalState)

	for _, element := range globalState.Mstate.Memory.GetItems(offsetV, offsetV+lengthV) {
		for _, annotation := range element.Annotations().Elements() {
			if reflect.TypeOf(annotation).String() == "OverUnderflowAnnotation" {
				stateAnnotation.OverflowingStateAnnotations.Add(annotation)
			}
		}
	}
}

func (dm *IntegerArithmetics) _handel_transaction_end(globalState *state.GlobalState) {
	stateAnnotation := getOverflowUnderflowStateAnnotation(globalState)

	for _, annotation := range stateAnnotation.OverflowingStateAnnotations.Elements() {
		ostate := annotation.(OverUnderflowAnnotation).OverflowingState
		if dm.OstatesUnsatisfiable.Contains(ostate) {
			continue
		}
		if !dm.OstatesSatisfiable.Contains(ostate) {
			// TODO: support/model
			// constraints SAT
			if true {
				dm.OstatesSatisfiable.Add(ostate)
			} else {
				// constrains UNSAT
				dm.OstatesUnsatisfiable.Add(ostate)
				continue
			}
		}
		fmt.Println("Checking overflow in", globalState.GetCurrentInstruction().OpCode.Name,
			"at transaction end address", globalState.GetCurrentInstruction().Address, "ostate address",
			ostate.GetCurrentInstruction().Address)
		// TODO: analysis/model

		issue := analysis.NewIssue(
			ostate.Environment.ActiveAccount.ContractName,
			ostate.Environment.ActiveFuncName,
			ostate.GetCurrentInstruction().Address,
			analysis.NewSWCData()["INTEGER_OVERFLOW_AND_UNDERFLOW"],
			"Integer Arithmetic Bugs",
			ostate.Environment.Code.Bytecode,
			"High",
		)
		address := getAddressFromState(ostate)
		dm.Cache.Add(address)
		dm.Issues = append(dm.Issues, issue)
	}
}

func getAddressFromState(globalState *state.GlobalState) int {
	return globalState.GetCurrentInstruction().Address
}

func getOverflowUnderflowStateAnnotation(globalState *state.GlobalState) OverUnderflowStateAnnotation {
	typeInstance := OverUnderflowStateAnnotation{}
	stateAnnotations := globalState.GetAnnotations(reflect.TypeOf(typeInstance))

	if len(stateAnnotations) == 0 {
		stateAnnotation := NewOverUnderflowStateAnnotation()
		globalState.Annotate(stateAnnotation)
		return stateAnnotation
	} else {
		return stateAnnotations[0].(OverUnderflowStateAnnotation)
	}
}
