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

type handelFuncSingle func(globalState *state.GlobalState)
type OverUnderflowAnnotationSingle struct {
	// Symbol Annotation used if a BitVector can overflow
	OverflowingState *state.GlobalState
	Operator         string
	Constraint       *z3.Bool
}
type OverUnderflowStateAnnotationSingle struct {
	OverflowingStateAnnotations *utils.Set
}

func NewOverUnderflowStateAnnotationSingle() *OverUnderflowStateAnnotationSingle {
	return &OverUnderflowStateAnnotationSingle{
		OverflowingStateAnnotations: utils.NewSet(),
	}
}
func (anno OverUnderflowStateAnnotationSingle) PersistToWorldState() bool {
	return false
}
func (anno OverUnderflowStateAnnotationSingle) PersistOverCalls() bool {
	return false
}
func (anno OverUnderflowStateAnnotationSingle) Copy() state.StateAnnotation {
	return &OverUnderflowStateAnnotationSingle{
		OverflowingStateAnnotations: anno.OverflowingStateAnnotations.Copy(),
	}
}
func (anno OverUnderflowStateAnnotationSingle) Translate(ctx *z3.Context) state.StateAnnotation {
	return anno.Copy()
}

type IntegerArithmeticsSingle struct {
	Name                 string
	SWCID                string
	Description          string
	PreHooks             []string
	Issues               *utils.SyncSlice
	Cache                *utils.Set
	OstatesSatisfiable   *utils.Set
	OstatesUnsatisfiable *utils.Set
}

func NewIntegerArithmeticsSingle() *IntegerArithmeticsSingle {
	return &IntegerArithmeticsSingle{
		Name:  "Integer overflow or underflow",
		SWCID: analysis.NewSWCData()["INTEGER_OVERFLOW_AND_UNDERFLOW"],
		Description: "For every SUB instruction, " +
			"check if there's a possible state where op1 > op0. " +
			"For every ADD, MUL instruction, " +
			"check if there's a possible state where op1 + op0 > 2^32 - 1",
		PreHooks: []string{"ADD", "SUB", "MUL", "EXP", "SSTORE",
			"JUMPI", "STOP", "RETURN", "CALL"},
		Issues:               utils.NewSyncSlice(),
		Cache:                utils.NewSet(),
		OstatesSatisfiable:   utils.NewSet(),
		OstatesUnsatisfiable: utils.NewSet(),
	}
}
func (dm *IntegerArithmeticsSingle) ResetModule() {
	dm.Issues = utils.NewSyncSlice()
	dm.OstatesSatisfiable = utils.NewSet()
	dm.OstatesUnsatisfiable = utils.NewSet()
}
func (dm *IntegerArithmeticsSingle) Execute(target *state.GlobalState) []*analysis.Issue {
	// fmt.Println("Entering analysis module: ", dm.Name)
	result := dm._execute(target)
	// fmt.Println("Exiting analysis module:", dm.Name)
	return result
}
func (dm *IntegerArithmeticsSingle) AddIssue(issue *analysis.Issue) {
	dm.Issues.Append(issue)
}
func (dm *IntegerArithmeticsSingle) GetIssues() []*analysis.Issue {
	list := make([]*analysis.Issue, 0)
	for _, v := range dm.Issues.Elements() {
		list = append(list, v.(*analysis.Issue))
	}
	return list
}
func (dm *IntegerArithmeticsSingle) GetPreHooks() []string {
	return dm.PreHooks
}
func (dm *IntegerArithmeticsSingle) GetPostHooks() []string {
	return make([]string, 0)
}
func (dm *IntegerArithmeticsSingle) GetCache() *utils.Set {
	return dm.Cache
}
func (dm *IntegerArithmeticsSingle) _get_args(state *state.GlobalState) (*z3.Bitvec, *z3.Bitvec) {
	stack := state.Mstate.Stack
	op0 := stack.RawStack[stack.Length()-1]
	op1 := stack.RawStack[stack.Length()-2]
	return op0, op1
}
func (dm *IntegerArithmeticsSingle) _execute(globalState *state.GlobalState) []*analysis.Issue {
	address := getAddressFromStateSingle(globalState)
	if dm.Cache.Contains(address) {
		//return dm.Issues
		return nil
	}
	opcode := globalState.GetCurrentInstruction().OpCode.Name
	funcs := make(map[string][]handelFuncSingle)
	funcs["ADD"] = []handelFuncSingle{dm._handel_add}
	funcs["SUB"] = []handelFuncSingle{dm._handel_sub}
	funcs["MUL"] = []handelFuncSingle{dm._handel_mul}
	funcs["SSTORE"] = []handelFuncSingle{dm._handel_sstore}
	funcs["JUMPI"] = []handelFuncSingle{dm._handel_jumpi}
	funcs["CALL"] = []handelFuncSingle{dm._handel_call}
	funcs["RETURN"] = []handelFuncSingle{dm._handel_return, dm._handel_transaction_end}
	funcs["STOP"] = []handelFuncSingle{dm._handel_transaction_end}
	funcs["EXP"] = []handelFuncSingle{dm._handel_exp}
	for _, f := range funcs[opcode] {
		f(globalState)
	}
	return nil
	//return dm.Issues
}

func (dm *IntegerArithmeticsSingle) _handel_add(globalState *state.GlobalState) {
	//config := z3.GetConfig()
	//ctx := z3.NewContext(config)
	//newState := globalState.Copy()
	//newState.Translate(ctx)

	op0, op1 := dm._get_args(globalState)
	c := op0.BvAddNoOverflow(op1, false).Not()
	//c := op0.BvAddNoOverflow(op1, false).Not().Translate(ctx)
	annotation := OverUnderflowAnnotationSingle{
		//OverflowingState: newState,
		OverflowingState: globalState.Copy(),
		Operator:         "addition",
		Constraint:       c,
	}
	op0.Annotate(annotation)
}

func (dm *IntegerArithmeticsSingle) _handel_mul(globalState *state.GlobalState) {
	//config := z3.GetConfig()
	//ctx := z3.NewContext(config)
	//newState := globalState.Copy()
	//newState.Translate(ctx)

	op0, op1 := dm._get_args(globalState)
	c := op0.BvMulNoOverflow(op1, false).Not()
	//c := op0.BvMulNoOverflow(op1, false).Not().Translate(ctx)
	annotation := OverUnderflowAnnotationSingle{
		//OverflowingState: newState,
		OverflowingState: globalState.Copy(),
		Operator:         "multiplication",
		Constraint:       c,
	}
	op0.Annotate(annotation)
}
func (dm *IntegerArithmeticsSingle) _handel_sub(globalState *state.GlobalState) {
	//config := z3.GetConfig()
	//ctx := z3.NewContext(config)
	//newState := globalState.Copy()
	//newState.Translate(ctx)

	op0, op1 := dm._get_args(globalState)
	c := op0.BvSubNoUnderflow(op1, false).Not()
	//c := op0.BvSubNoUnderflow(op1, false).Not().Translate(ctx)
	annotation := OverUnderflowAnnotationSingle{
		//OverflowingState: newState,
		OverflowingState: globalState.Copy(),
		Operator:         "subtraction",
		Constraint:       c,
	}
	op0.Annotate(annotation)
}
func (dm *IntegerArithmeticsSingle) _handel_exp(globalState *state.GlobalState) {
	//config := z3.GetConfig()
	//newCtx := z3.NewContext(config)
	//newState := globalState.Copy()
	//newState.Translate(newCtx)

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

	//constraint = constraint.Translate(newCtx)

	annotation := OverUnderflowAnnotationSingle{
		//OverflowingState: newState,
		OverflowingState: globalState.Copy(),
		Operator:         "exponentiation",
		Constraint:       constraint,
	}
	op0.Annotate(annotation)
}
func (dm *IntegerArithmeticsSingle) _handel_sstore(globalState *state.GlobalState) {
	stack := globalState.Mstate.Stack
	value := stack.RawStack[stack.Length()-2]
	stateAnnotation := getOverflowUnderflowStateAnnotationSingle(globalState)
	for _, annotation := range value.Annotations.Elements() {
		if reflect.TypeOf(annotation).String() == "modules.OverUnderflowAnnotationSingle" {
			stateAnnotation.OverflowingStateAnnotations.Add(annotation)
		}
	}
}
func (dm *IntegerArithmeticsSingle) _handel_jumpi(globalState *state.GlobalState) {
	stack := globalState.Mstate.Stack
	value := stack.RawStack[stack.Length()-2]
	stateAnnotation := getOverflowUnderflowStateAnnotationSingle(globalState)
	//fmt.Println(stateAnnotation, " ", stateAnnotation.OverflowingStateAnnotations.Len())
	for _, annotation := range value.Annotations.Elements() {
		if reflect.TypeOf(annotation).String() == "modules.OverUnderflowAnnotationSingle" {
			stateAnnotation.OverflowingStateAnnotations.Add(annotation)
		}
	}
}
func (dm *IntegerArithmeticsSingle) _handel_call(globalState *state.GlobalState) {
	stack := globalState.Mstate.Stack
	value := stack.RawStack[stack.Length()-3]
	stateAnnotation := getOverflowUnderflowStateAnnotationSingle(globalState)
	for _, annotation := range value.Annotations.Elements() {
		if reflect.TypeOf(annotation).String() == "modules.OverUnderflowAnnotationSingle" {
			stateAnnotation.OverflowingStateAnnotations.Add(annotation)
		}
	}
}
func (dm *IntegerArithmeticsSingle) _handel_return(globalState *state.GlobalState) {
	stack := globalState.Mstate.Stack
	offset := stack.RawStack[stack.Length()-1]
	length := stack.RawStack[stack.Length()-2]
	stateAnnotation := getOverflowUnderflowStateAnnotationSingle(globalState)

	if length.Symbolic() {
		return
	}
	if offset.Symbolic() {
		element := globalState.Mstate.Memory.GetWordAt(offset)
		for _, annotation := range element.Annotations.Elements() {
			if reflect.TypeOf(annotation).String() == "modules.OverUnderflowAnnotationSingle" {
				stateAnnotation.OverflowingStateAnnotations.Add(annotation)
			}
		}
	}

	offsetV, _ := strconv.ParseInt(offset.Value(), 10, 64)
	lengthV, _ := strconv.ParseInt(length.Value(), 10, 64)
	for _, element := range globalState.Mstate.Memory.GetItems(offsetV, offsetV+lengthV, nil) {
		for _, annotation := range element.Annotations.Elements() {
			if reflect.TypeOf(annotation).String() == "modules.OverUnderflowAnnotationSingle" {
				stateAnnotation.OverflowingStateAnnotations.Add(annotation)
			}
		}
	}
}
func (dm *IntegerArithmeticsSingle) _handel_transaction_end(globalState *state.GlobalState) {
	stateAnnotation := getOverflowUnderflowStateAnnotationSingle(globalState)
	for _, annotation := range stateAnnotation.OverflowingStateAnnotations.Elements() {
		ostate := annotation.(OverUnderflowAnnotationSingle).OverflowingState
		if dm.OstatesUnsatisfiable.Contains(ostate) {
			continue
		}
		//if ostate.Z3ctx.GetRaw() != globalState.Z3ctx.GetRaw() {
		//	ostate.Translate(globalState.Z3ctx)
		//}

		if !dm.OstatesSatisfiable.Contains(ostate) {
			constraints := ostate.WorldState.Constraints.DeepCopy()
			constraints.Add(annotation.(OverUnderflowAnnotationSingle).Constraint)
			_, sat := state.GetModel(constraints, nil, nil, false, ostate.Z3ctx)
			if sat {
				// fmt.Println("sat")
				dm.OstatesSatisfiable.Add(ostate)
			} else {
				// UnsatError
				// fmt.Println("unsat")
				dm.OstatesUnsatisfiable.Add(ostate)
				continue
			}
		}
		fmt.Println("Checking overflow in", globalState.GetCurrentInstruction().OpCode.Name,
			"at transaction end address", globalState.GetCurrentInstruction().Address, "ostate address",
			ostate.GetCurrentInstruction().Address)

		constraints := globalState.WorldState.Constraints.DeepCopy()
		constraints.Add(annotation.(OverUnderflowAnnotationSingle).Constraint)
		//constraints.Add(annotation.(OverUnderflowAnnotationSingle).Constraint)

		transactionSequence := analysis.GetTransactionSequence(globalState, constraints)
		if transactionSequence == nil {
			// UnsatError
			fmt.Println("unsaterror for getTxSeq")
			continue
		}
		var flowStr string
		if annotation.(OverUnderflowAnnotationSingle).Operator == "subtraction" {
			flowStr = "underflow"
		} else {
			flowStr = "overflow"
		}
		descriptionHead := "The arithmetic operator can " + flowStr
		descriptionTail := "It is possible to cause an integer overflow or underflow in the arithmetic operation. " +
			"Prevent this by constraining inputs using the require() statement or use the OpenZeppelin SafeMath library for integer arithmetic operations. " +
			"Refer to the transaction trace generated for this issue to reproduce the issue."
		issue := &analysis.Issue{
			Contract:            ostate.Environment.ActiveAccount.ContractName,
			FunctionName:        ostate.Environment.ActiveFuncName,
			Address:             ostate.GetCurrentInstruction().Address,
			SWCID:               analysis.NewSWCData()["INTEGER_OVERFLOW_AND_UNDERFLOW"],
			Bytecode:            ostate.Environment.Code.Bytecode,
			Title:               "Integer Arithmetic Bugs",
			Severity:            "High",
			DescriptionHead:     descriptionHead,
			DescriptionTail:     descriptionTail,
			GasUsed:             []int{globalState.Mstate.MinGasUsed, globalState.Mstate.MaxGasUsed},
			TransactionSequence: transactionSequence,
		}
		address := getAddressFromStateSingle(ostate)
		dm.Cache.Add(address)
		//dm.Issues = append(dm.Issues, issue)
		dm.Issues.Append(issue)
	}
}
func getAddressFromStateSingle(globalState *state.GlobalState) int {
	return globalState.GetCurrentInstruction().Address
}
func getOverflowUnderflowStateAnnotationSingle(globalState *state.GlobalState) *OverUnderflowStateAnnotationSingle {
	typeInstance := &OverUnderflowStateAnnotationSingle{}
	stateAnnotations := globalState.GetAnnotations(reflect.TypeOf(typeInstance))
	if len(stateAnnotations) == 0 {
		stateAnnotation := NewOverUnderflowStateAnnotationSingle()
		globalState.Annotate(stateAnnotation)
		return stateAnnotation
	} else {
		return stateAnnotations[0].(*OverUnderflowStateAnnotationSingle)
	}
}
