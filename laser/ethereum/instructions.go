package ethereum

import (
	"fmt"
	"go-mythril/laser/ethereum/state"
	"go-mythril/laser/smt/z3"
	"math/big"
	"strconv"
	"strings"
)

type StateTransition struct {
}

// TODO:
func CheckGasUsageLimit(globalState *state.GlobalState) {
	globalState.Mstate.CheckGas()
	value := globalState.CurrentTransaction().GasLimit.Value()
	if value == "" {
		return
	}
	valueInt, _ := strconv.ParseInt(value, 10, 64)
	if globalState.Mstate.MinGasUsed >= int(valueInt) {
		panic("OutOfGasException")
	}
}

type Instruction struct {
	Opcode    string
	PreHooks  []string
	PostHooks []string
}

// NewInstruction Golang don't support default parameters.
func NewInstruction(opcode string, prehooks []string, posthooks []string) *Instruction {
	return &Instruction{
		Opcode:    opcode,
		PreHooks:  prehooks,
		PostHooks: posthooks,
	}
}

func (instr *Instruction) ExePreHooks(globalState *state.GlobalState) {
	for _, funcName := range instr.PreHooks {
		tmpInstr := &Instruction{
			Opcode: funcName,
		}
		// execute the hook
		tmpInstr.Mutator(globalState)
		fmt.Println("exe prehooks!")
	}
}

func (instr *Instruction) ExePostHooks(globalState *state.GlobalState) {
	for _, funcName := range instr.PostHooks {
		tmpInstr := &Instruction{
			Opcode: funcName,
		}
		// execute the hook
		tmpInstr.Mutator(globalState)
		fmt.Println("exe posthooks!")
	}
}

func (instr *Instruction) Evaluate(globalState *state.GlobalState) []*state.GlobalState {

	// TODO: Pre hook
	instr.ExePreHooks(globalState)
	result := instr.Mutator(globalState)
	// TODO: Post hook
	instr.ExePostHooks(globalState)
	fmt.Println("Evaluate", len(result))

	// TODO
	for _, state := range result {
		state.Mstate.Pc++
		// For debug
		state.Mstate.Stack.PrintStack()
	}
	return result
}

// using reflect (getattr) might be too complex? maybe if-else is good
func (instr *Instruction) Mutator(globalState *state.GlobalState) []*state.GlobalState {
	// TODO
	if strings.HasPrefix(instr.Opcode, "PUSH") {
		return instr.push_(globalState)
	} else if strings.HasPrefix(instr.Opcode, "DUP") {
		return instr.dup_(globalState)
	} else if strings.HasPrefix(instr.Opcode, "SWAP") {
		return instr.swap_(globalState)
	} else if instr.Opcode == "ORIGIN" {
		return instr.origin_(globalState)
	} else if instr.Opcode == "STOP" {
		return instr.stop_(globalState)
	} else if instr.Opcode == "AND" {
		return instr.and_(globalState)
	} else if instr.Opcode == "OR" {
		return instr.or_(globalState)
	} else if instr.Opcode == "XOR" {
		return instr.xor_(globalState)
	} else if instr.Opcode == "NOT" {
		return instr.not_(globalState)
	} else if instr.Opcode == "BYTE" {
		return instr.byte_(globalState)
	} else if instr.Opcode == "POP" {
		return instr.pop_(globalState)
	} else if instr.Opcode == "ADD" {
		return instr.add_(globalState)
	} else if instr.Opcode == "SUB" {
		return instr.sub_(globalState)
	} else if instr.Opcode == "MUL" {
		return instr.mul_(globalState)
	} else if instr.Opcode == "DIV" {
		return instr.div_(globalState)
	} else if instr.Opcode == "SDIV" {
		return instr.sdiv_(globalState)
	} else if instr.Opcode == "MOD" {
		return instr.mod_(globalState)
	} else if instr.Opcode == "SMOD" {
		return instr.smod_(globalState)
	} else if instr.Opcode == "ADDMOD" {
		return instr.addmod_(globalState)
	} else if instr.Opcode == "MULMOD" {
		return instr.mulmod_(globalState)
	} else if instr.Opcode == "EXP" {
		return instr.exp_(globalState)
	} else if instr.Opcode == "SIGNEXTEND" {
		return instr.signextend_(globalState)
	} else if instr.Opcode == "SHL" {
		return instr.shl_(globalState)
	} else if instr.Opcode == "SHR" {
		return instr.shr_(globalState)
	} else if instr.Opcode == "SAR" {
		return instr.sar_(globalState)
	} else if instr.Opcode == "LT" {
		return instr.lt_(globalState)
	} else if instr.Opcode == "GT" {
		return instr.gt_(globalState)
	} else if instr.Opcode == "SLT" {
		return instr.slt_(globalState)
	} else if instr.Opcode == "SGT" {
		return instr.sgt_(globalState)
	} else if instr.Opcode == "EQ" {
		return instr.eq_(globalState)
	} else if instr.Opcode == "ISZERO" {
		return instr.iszero_(globalState)
	} else {
		panic("?" + instr.Opcode)
	}

	return nil
}

func (instr *Instruction) stop_(globalState *state.GlobalState) []*state.GlobalState {
	// TODO
	ret := make([]*state.GlobalState, 0)
	return ret
}

func (instr *Instruction) push_(globalState *state.GlobalState) []*state.GlobalState {
	mstate := globalState.Mstate
	ret := make([]*state.GlobalState, 0)

	pushInstruction := globalState.GetCurrentInstruction()
	pushValue := pushInstruction.Argument[2:]
	// TODO: check length

	pushInt, _ := strconv.ParseInt(pushValue, 16, 64)
	mstate.Stack.Append(globalState.Z3ctx.NewBitvecVal(pushInt, 256))
	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) dup_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)
	value, _ := strconv.ParseInt(globalState.GetCurrentInstruction().OpCode.Name[3:], 10, 64)

	mstate := globalState.Mstate
	mstate.Stack.Append(mstate.Stack.RawStack[mstate.Stack.Length()-int(value)])
	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) swap_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	length := mstate.Stack.Length()
	depth, _ := strconv.ParseInt(globalState.GetCurrentInstruction().OpCode.Name[4:], 10, 64)
	mstate.Stack.RawStack[length-int(depth)], mstate.Stack.RawStack[length-1] = mstate.Stack.RawStack[length-1], mstate.Stack.RawStack[length-int(depth)]

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) pop_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	mstate.Stack.Pop()

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) and_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	op1 := mstate.Stack.Pop()
	op2 := mstate.Stack.Pop()
	// TODO: need type check ?
	mstate.Stack.Append(op1.BvAnd(op2))

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) or_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	op1 := mstate.Stack.Pop()
	op2 := mstate.Stack.Pop()
	// TODO: need type check ?
	mstate.Stack.Append(op1.BvOr(op2))

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) xor_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	op1 := mstate.Stack.Pop()
	op2 := mstate.Stack.Pop()
	mstate.Stack.Append(op1.BvXOr(op2))

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) not_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	// Should use big.Int
	val256 := new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil)
	TT256M1 := globalState.Z3ctx.NewBitvecVal(new(big.Int).Sub(val256, big.NewInt(1)), 256)
	op1 := mstate.Stack.Pop()
	mstate.Stack.Append(TT256M1.BvSub(op1))

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) byte_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	op0 := mstate.Stack.Pop()
	op1 := mstate.Stack.Pop()
	// the range of index is [0,31], index 0 is the left side of bv.
	var result *z3.Bitvec
	// here, we should check the op0's type
	if op0.Symbolic() {
		fmt.Println("BYTE: Unsupported symbolic byte offset")
		var op1str string
		if op1.Symbolic() {
			op1str = op1.Simplify().String()
		} else {
			op1str = op1.Value()
		}
		result = globalState.NewBitvec(op1str+"["+op0.String()+"]", 256)
	} else {
		index, _ := strconv.ParseInt(op0.Value(), 10, 64)
		offset := int((31 - index) * 8)

		if offset >= 0 {
			tmp := globalState.Z3ctx.NewBitvecVal(0, 248)
			tmp1 := op1.Extract(offset+7, offset)
			result = tmp.Concat(tmp1).Simplify()
		} else {
			result = globalState.Z3ctx.NewBitvecVal(0, 256)
		}
	}

	mstate.Stack.Append(result)

	ret = append(ret, globalState)
	return ret
}

// Arithmetic
func (instr *Instruction) add_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	op1 := mstate.Stack.Pop()
	op2 := mstate.Stack.Pop()
	mstate.Stack.Append(op1.BvAdd(op2))

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) sub_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	op1 := mstate.Stack.Pop()
	op2 := mstate.Stack.Pop()
	mstate.Stack.Append(op1.BvSub(op2))

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) mul_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	op1 := mstate.Stack.Pop()
	op2 := mstate.Stack.Pop()
	mstate.Stack.Append(op1.BvMul(op2))

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) div_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	op0 := mstate.Stack.Pop()
	op1 := mstate.Stack.Pop()
	// unsigned op0 / op1, when op1 == 0
	if op1.Value() == "0" {
		mstate.Stack.Append(globalState.Z3ctx.NewBitvecVal(0, 256))
	} else {
		mstate.Stack.Append(op0.BvUDiv(op1))
		mstate.Stack.RawStack = append(mstate.Stack.RawStack, op0.BvUDiv(op1))
	}

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) sdiv_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	op0 := mstate.Stack.Pop()
	op1 := mstate.Stack.Pop()
	// op0 / op1, when op1 == 0
	if op1.Value() == "0" {
		mstate.Stack.Append(globalState.Z3ctx.NewBitvecVal(0, 256))
	} else {
		mstate.Stack.Append(op0.BvSDiv(op1))
	}

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) mod_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	op0 := mstate.Stack.Pop()
	op1 := mstate.Stack.Pop()

	if op1.Value() == "0" {
		mstate.Stack.Append(globalState.Z3ctx.NewBitvecVal(0, 256))
	} else {
		mstate.Stack.Append(op0.BvURem(op1))
	}

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) shl_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	shift := mstate.Stack.Pop()
	value := mstate.Stack.Pop()
	mstate.Stack.Append(value.BvShL(shift))

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) shr_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	shift := mstate.Stack.Pop()
	value := mstate.Stack.Pop()
	mstate.Stack.Append(value.BvLShR(shift))

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) sar_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	shift := mstate.Stack.Pop()
	value := mstate.Stack.Pop()
	mstate.Stack.Append(value.BvShR(shift))

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) smod_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	op0 := mstate.Stack.Pop()
	op1 := mstate.Stack.Pop()

	if op1.Value() == "0" {
		mstate.Stack.Append(globalState.Z3ctx.NewBitvecVal(0, 256))
	} else {
		mstate.Stack.Append(op0.BvSRem(op1))
	}

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) addmod_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	op0 := mstate.Stack.Pop()
	op1 := mstate.Stack.Pop()
	op2 := mstate.Stack.Pop()

	tmp := (op0.BvURem(op2)).BvAdd(op1.BvURem(op2))
	mstate.Stack.Append(tmp.BvURem(op2))

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) mulmod_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	op0 := mstate.Stack.Pop()
	op1 := mstate.Stack.Pop()
	op2 := mstate.Stack.Pop()

	tmp := (op0.BvURem(op2)).BvMul(op1.BvURem(op2))
	mstate.Stack.Append(tmp.BvURem(op2))

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) exp_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	base := mstate.Stack.Pop()
	exponent := mstate.Stack.Pop()
	exponentiation, constraint := createCondition(base, exponent, globalState.Z3ctx)
	mstate.Stack.Append(exponentiation)
	globalState.WorldState.Constraints.Add(constraint)

	ret = append(ret, globalState)
	return ret
}

// The method should be in the laser/ethereum/function_managers
func createCondition(base *z3.Bitvec, exponent *z3.Bitvec, ctx *z3.Context) (*z3.Bitvec, *z3.Bool) {

	if !base.Symbolic() && !exponent.Symbolic() {
		// TODO: MAX value check
		baseV, _ := strconv.ParseInt(base.Value(), 10, 64)
		exponentV, _ := strconv.ParseInt(exponent.Value(), 10, 64)
		val := new(big.Int).Exp(big.NewInt(baseV), big.NewInt(exponentV), nil)
		constExponentiation := ctx.NewBitvecVal(val, 256)
		// TODO:
		constraint := &z3.Bool{}
		return constExponentiation, constraint
	}
	// TODO:
	return nil, nil
}

// TODO: NOT tested
func (instr *Instruction) signextend_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	ctx := globalState.Z3ctx
	mstate := globalState.Mstate
	s0 := mstate.Stack.Pop()
	s1 := mstate.Stack.Pop()

	testbit := s0.BvMul(ctx.NewBitvecVal(8, 256)).
		BvAdd(ctx.NewBitvecVal(7, 256))
	setTestbit := ctx.NewBitvecVal(1, 256).BvShL(testbit)
	signBitset := s1.BvAnd(setTestbit).Eq(ctx.NewBitvecVal(0, 256)).Not().Simplify()

	TT256 := ctx.NewBitvecVal(0, 256)
	if0 := z3.If(signBitset, s1.BvOr(TT256.BvSub(setTestbit)),
		s1.BvAnd(setTestbit.BvSub(ctx.NewBitvecVal(1, 256))))
	if1 := z3.If(s0.BvSLe(ctx.NewBitvecVal(31, 256)), if0, s1)
	mstate.Stack.Append(if1.Simplify())

	ret = append(ret, globalState)
	return ret
}

// Comparisons
func (instr *Instruction) lt_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	op1 := mstate.Stack.Pop()
	op2 := mstate.Stack.Pop()
	mstate.Stack.Append(op1.BvULt(op2))

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) gt_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	op1 := mstate.Stack.Pop()
	op2 := mstate.Stack.Pop()
	mstate.Stack.Append(op1.BvUGt(op2))

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) slt_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	op1 := mstate.Stack.Pop()
	op2 := mstate.Stack.Pop()
	mstate.Stack.Append(op1.BvSLt(op2))

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) sgt_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	op1 := mstate.Stack.Pop()
	op2 := mstate.Stack.Pop()
	mstate.Stack.Append(op1.BvSGt(op2))

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) eq_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	op1 := mstate.Stack.Pop()
	op2 := mstate.Stack.Pop()
	mstate.Stack.Append(op1.Eq(op2))

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) iszero_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	ctx := globalState.Z3ctx
	val := mstate.Stack.Pop()
	exp := val.Eq(ctx.NewBitvecVal(0, 256))
	mstate.Stack.Append(z3.If(exp, ctx.NewBitvecVal(1, 256),
		ctx.NewBitvecVal(0, 256)).Simplify())

	ret = append(ret, globalState)
	return ret
}

// Call data
func (instr *Instruction) callvalue_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	env := globalState.Environment
	mstate.Stack.Append(env.CallValue)

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) calldataload_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	env := globalState.Environment
	op0 := mstate.Stack.Pop()
	value := env.Calldata.GetWordAt(op0)
	mstate.Stack.Append(value)

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) calldatasize_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	env := globalState.Environment
	mstate.Stack.Append(env.Calldata.Calldatasize())

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) _calldata_copy_helper(globalState *state.GlobalState,
	mstate *state.MachineState, mstart *z3.Bitvec, dstart *z3.Bitvec, size *z3.Bitvec) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	env := globalState.Environment
	ctx := globalState.Z3ctx

	if mstart.Symbolic() {
		fmt.Println("Unsupported symbolic memory offset in CALLDATACOPY")
		ret = append(ret, globalState)
		return ret
	}
	vSize, _ := strconv.ParseInt(size.Value(), 10, 64)
	if vSize > 0 {
		mstate.MemExtend(mstart, 1)
		// TypeError check
		iData := dstart
		newMemory := make([]*z3.Bitvec, 0)
		for i := 0; i < int(vSize); i++ {
			value := env.Calldata.GetWordAt(iData)
			newMemory = append(newMemory, value)
			idataValue, _ := strconv.ParseInt(iData.Value(), 10, 64)
			iData = ctx.NewBitvecVal(idataValue+1, iData.BvSize())
		}
		for j := 0; j < len(newMemory); j++ {
			mstartValue, _ := strconv.ParseInt(mstart.Value(), 10, 64)
			mstate.Memory.WriteWordAt(j+int(mstartValue), newMemory[j])
		}
		// IndexError check
	}
	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) calldatacopy_(globalState *state.GlobalState) []*state.GlobalState {
	mstate := globalState.Mstate
	op0 := mstate.Stack.Pop()
	op1 := mstate.Stack.Pop()
	op2 := mstate.Stack.Pop()
	// TODO: tx warning
	return instr._calldata_copy_helper(globalState, mstate, op0, op1, op2)
}

// Environment
func (instr *Instruction) address_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	env := globalState.Environment
	mstate.Stack.Append(env.Address)

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) balance_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	ctx := globalState.Z3ctx
	address := mstate.Stack.Pop()
	var balance *z3.Bitvec
	if !address.Symbolic() {
		balance = globalState.WorldState.AccountsExistOrLoad(address.Value()).Balance()
	} else {
		balance = ctx.NewBitvecVal(0, 256)
		for _, acc := range *globalState.WorldState.Accounts {
			balance = z3.If(address.Eq(acc.Address), acc.Balance(), balance)
		}
	}
	mstate.Stack.Append(balance)

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) origin_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	env := globalState.Environment
	mstate.Stack.Append(env.Origin)

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) caller_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	env := globalState.Environment
	mstate.Stack.Append(env.Sender)

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) chainid_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	env := globalState.Environment
	mstate.Stack.Append(env.ChainId)

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) selfbalance_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	env := globalState.Environment
	mstate.Stack.Append(env.ActiveAccount.Balance())

	ret = append(ret, globalState)
	return ret
}

// TODO
func (instr *Instruction) codesize_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) _sha3_gas_helper(globalState *state.GlobalState, length int) *state.GlobalState {
	minGas, maxGas := CalculateSha3Gas(length)
	globalState.Mstate.MinGasUsed += minGas
	globalState.Mstate.MaxGasUsed += maxGas
	CheckGasUsageLimit(globalState)
	return globalState
}

// TODO
func (instr *Instruction) sha3_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	index := mstate.Stack.Pop()
	op1 := mstate.Stack.Pop()

	length, _ := strconv.ParseInt(op1.Value(), 10, 64)
	if length == 0 {
		// can't access symbolic memory offsets
		length = 64
		globalState.WorldState.Constraints.Add(op1.Eq(globalState.Z3ctx.NewBitvecVal(length, 256)))
	}
	instr._sha3_gas_helper(globalState, int(length))

	mstate.MemExtend(index, int(length))
	// TODO: Memory

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) gasprice_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	env := globalState.Environment
	mstate.Stack.Append(env.GasPrice)

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) basefee_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	env := globalState.Environment
	mstate.Stack.Append(env.Basefee)

	ret = append(ret, globalState)
	return ret
}

// TODO
func (instr *Instruction) codecopy_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	ret = append(ret, globalState)
	return ret
}

// TODO
func (instr *Instruction) _code_copy_helper(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) extcodesize_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	addr := mstate.Stack.Pop()

	if addr.Symbolic() {
		// TypeError
		fmt.Println("unsupported symbolic address for EXTCODESIZE")
		mstate.Stack.Append(globalState.NewBitvec("extcodesize_"+addr.String(), 256))
		ret = append(ret, globalState)
		return ret
	}
	code := globalState.WorldState.AccountsExistOrLoad(addr.Value()).Code.Bytecode
	mstate.Stack.Append(len(code) / 2)

	ret = append(ret, globalState)
	return ret
}

// Memory operations
// TODO: NOT tested
/*func (instr *Instruction) mload_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	ctx := globalState.Z3ctx
	mstate := globalState.Mstate
	offset := mstate.Stack.Pop()

	mstate.MemExtend(offset, 32)
	data := mstate.Memory.GetWordAt(offset, ctx)
	mstate.Stack.Append(data)

	ret = append(ret, globalState)
	return ret
}

// TODO: NOT tested
func (instr *Instruction) mstore_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	ctx := globalState.Z3ctx
	mstate := globalState.Mstate
	mstart := mstate.Stack.Pop()
	value := mstate.Stack.Pop()
	// TODO: exception
	mstate.MemExtend(mstart, 32)
	mstate.Memory.WriteWordAt(mstart, ctx, value)

	ret = append(ret, globalState)
	return ret
}*/
