package ethereum

import (
	"fmt"
	"go-mythril/analysis/module/modules"
	"go-mythril/laser/ethereum/function_managers"
	"go-mythril/laser/ethereum/state"
	"go-mythril/laser/smt/z3"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

type StateTransition struct {
}

func transferEther(globalState *state.GlobalState, sender *z3.Bitvec, receiver *z3.Bitvec, value *z3.Bitvec) {
	fmt.Println("transferEther1")
	globalState.WorldState.Constraints.Add(globalState.WorldState.Balances.GetItem(sender).BvUGe(value))
	fmt.Println("transferEther2")
	globalState.WorldState.Balances.SetItem(receiver, globalState.WorldState.Balances.GetItem(receiver).BvAdd(value))
	fmt.Println("transferEther3")
	globalState.WorldState.Balances.SetItem(sender, globalState.WorldState.Balances.GetItem(sender).BvSub(value))
	fmt.Println("transferEther4")
}

// TODO:
func CheckGasUsageLimit(globalState *state.GlobalState) {
	globalState.Mstate.CheckGas()
	value := globalState.CurrentTransaction().GetGasLimit()
	if value == 0 {
		return
	}
	if globalState.Mstate.MinGasUsed >= value {
		panic("OutOfGasException-Instr-CheckGasUsageLimit")
	}
}

type Instruction struct {
	Opcode    string
	PreHooks  []moduleExecFunc
	PostHooks []moduleExecFunc
}

// NewInstruction Golang don't support default parameters.
func NewInstruction(opcode string, prehooks []moduleExecFunc, posthooks []moduleExecFunc) *Instruction {
	return &Instruction{
		Opcode:    opcode,
		PreHooks:  prehooks,
		PostHooks: posthooks,
	}
}

func (instr *Instruction) ExePreHooks(globalState *state.GlobalState) {
	modules.IsPreHook = true
	for _, hook := range instr.PreHooks {
		// fmt.Println(instr.Opcode, ": preHook execute!")
		hook(globalState)
	}
}

func (instr *Instruction) ExePostHooks(globalState *state.GlobalState) {
	modules.IsPreHook = false
	for _, hook := range instr.PostHooks {
		// fmt.Println(instr.Opcode, ": postHook execute!")
		hook(globalState)
	}
}

var counter int
var cLock sync.Mutex

func (instr *Instruction) Evaluate(globalState *state.GlobalState) []*state.GlobalState {

	//if globalState.GetCurrentInstruction().OpCode.Name == "SUB" && globalState.GetCurrentInstruction().Address == 881 {
	////	fmt.Println("Cons:", globalState.GetCurrentInstruction().Address, globalState.GetCurrentInstruction().OpCode.Name )
	////	globalState.WorldState.Constraints.PrintOneLine()
	////
	//	cLock.Lock()
	////
	//	counter = counter + 1
	//	file, err := os.OpenFile("/home/codepatient/log/"+strconv.Itoa(counter)+".txt", os.O_WRONLY|os.O_APPEND, 0666)
	//	if err != nil {
	//		fmt.Println("file open fail", err)
	//	}
	//	defer file.Close()
	//	write := bufio.NewWriter(file)
	//	write.WriteString("ConstraintsLen-" + strconv.Itoa(globalState.WorldState.Constraints.Length()) + "\r\n")
	//	for i, con := range globalState.WorldState.Constraints.ConstraintList {
	//		str := con.BoolString()
	//		idx := strings.Index(str, "\n")
	//		if idx != -1 {
	//			str = str[:idx]
	//		}
	//		write.WriteString(strconv.Itoa(i) + "-" + str + "\r\n")
	//	}
	//	write.WriteString("+++++++++++++++++++++++++++++++++\r\n")
	//	write.Flush()
	//
	//	cLock.Unlock()
	//}
	//if globalState.WorldState.Constraints.Length() <= 24 && globalState.WorldState.Constraints.Length()>= 20 {
	//	fmt.Println("Cons:", globalState.GetCurrentInstruction().Address, globalState.GetCurrentInstruction().OpCode.Name )
	//	globalState.WorldState.Constraints.PrintOneLine()
	//}

	instr.ExePreHooks(globalState)

	result := instr.Mutator(globalState)

	// fmt.Println("PC:", globalState.Mstate.Pc)
	// fmt.Println("Address:", globalState.GetCurrentInstruction().Address, globalState.GetCurrentInstruction().OpCode.Name)
	// has the same function of StateTransition in instructions.go
	for _, state := range result {
		if instr.Opcode != "JUMPI" && instr.Opcode != "JUMP" && instr.Opcode != "RETURNSUB" {
			state.Mstate.Pc++
		}
	}
	instr.ExePostHooks(globalState)

	// for _, state := range result {
	// For debug
	//fmt.Println("Print:", globalState.GetCurrentInstruction().OpCode.Name)

	//state.Mstate.Stack.PrintStack()
	//for i, con := range state.WorldState.Constraints.ConstraintList {
	//	if i==3{
	//		fmt.Println("PrintCons:", con.BoolString())
	//	}
	//}
	//state.Mstate.Memory.PrintMemoryOneLine()
	//state.Mstate.Memory.PrintMemory()
	// fmt.Println("StackLen:", state.Mstate.Stack.Length())
	// }
	//fmt.Println("------------------------------------------------------------")

	return result
}

// Mutator :using reflect (getattr) might be too complex? maybe if-else is good
// reflect also has a problem of performance
func (instr *Instruction) Mutator(globalState *state.GlobalState) []*state.GlobalState {
	// TODO
	if strings.HasPrefix(instr.Opcode, "PUSH") {
		return instr.push_(globalState)
	} else if strings.HasPrefix(instr.Opcode, "DUP") {
		return instr.dup_(globalState)
	} else if strings.HasPrefix(instr.Opcode, "SWAP") {
		return instr.swap_(globalState)
	} else if strings.HasPrefix(instr.Opcode, "LOG") {
		return instr.log_(globalState)
	} else if instr.Opcode == "JUMPDEST" {
		return instr.jumpdest_(globalState)
	} else if instr.Opcode == "POP" {
		return instr.pop_(globalState)
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
	} else if instr.Opcode == "SHL" {
		return instr.shl_(globalState)
	} else if instr.Opcode == "SHR" {
		return instr.shr_(globalState)
	} else if instr.Opcode == "SAR" {
		return instr.sar_(globalState)
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
	} else if instr.Opcode == "CALLVALUE" {
		return instr.callvalue_(globalState)
	} else if instr.Opcode == "CALLDATALOAD" {
		return instr.calldataload_(globalState)
	} else if instr.Opcode == "CALLDATASIZE" {
		return instr.calldatasize_(globalState)
	} else if instr.Opcode == "CALLDATACOPY" {
		return instr.calldatacopy_(globalState)
	} else if instr.Opcode == "ADDRESS" {
		return instr.address_(globalState)
	} else if instr.Opcode == "BALANCE" {
		return instr.balance_(globalState)
	} else if instr.Opcode == "ORIGIN" {
		return instr.origin_(globalState)
	} else if instr.Opcode == "CALLER" {
		return instr.caller_(globalState)
	} else if instr.Opcode == "CHAINID" {
		return instr.chainid_(globalState)
	} else if instr.Opcode == "SELFBALANCE" {
		return instr.selfbalance_(globalState)
	} else if instr.Opcode == "CODESIZE" {
		return instr.codesize_(globalState)
	} else if instr.Opcode == "SHA3" {
		return instr.sha3_(globalState)
	} else if instr.Opcode == "GASPRICE" {
		return instr.gasprice_(globalState)
	} else if instr.Opcode == "BASEFEE" {
		return instr.basefee_(globalState)
	} else if instr.Opcode == "CODECOPY" {
		return instr.codecopy_(globalState)
	} else if instr.Opcode == "EXTCODESIZE" {
		return instr.extcodesize_(globalState)
	} else if instr.Opcode == "EXTCODECOPY" {
		return instr.extcodecopy_(globalState)
	} else if instr.Opcode == "EXTCODEHASH" {
		return instr.extcodehash_(globalState)
	} else if instr.Opcode == "RETURNDATACOPY" {
		return instr.returndatacopy_(globalState)
	} else if instr.Opcode == "RETURNDATASIZE" {
		return instr.returndatasize_(globalState)
	} else if instr.Opcode == "BLOCKHASH" {
		return instr.blockhash_(globalState)
	} else if instr.Opcode == "COINBASE" {
		return instr.coinbase_(globalState)
	} else if instr.Opcode == "TIMESTAMP" {
		return instr.timestamp_(globalState)
	} else if instr.Opcode == "NUMBER" {
		return instr.number_(globalState)
	} else if instr.Opcode == "DIFFICULTY" {
		return instr.difficulty_(globalState)
	} else if instr.Opcode == "GASLIMIT" {
		return instr.gaslimit_(globalState)
	} else if instr.Opcode == "MLOAD" {
		return instr.mload_(globalState)
	} else if instr.Opcode == "MSTORE" {
		return instr.mstore_(globalState)
	} else if instr.Opcode == "MSTORE8" {
		return instr.mstore8_(globalState)
	} else if instr.Opcode == "SLOAD" {
		return instr.sload_(globalState)
	} else if instr.Opcode == "SSTORE" {
		return instr.sstore_(globalState)
	} else if instr.Opcode == "JUMP" {
		return instr.jump_(globalState)
	} else if instr.Opcode == "JUMPI" {
		return instr.jumpi_(globalState)
	} else if instr.Opcode == "BEGINSUB" {
		return instr.beginsub_(globalState)
	} else if instr.Opcode == "JUMPSUB" {
		return instr.jumpsub_(globalState)
	} else if instr.Opcode == "RETURNSUB" {
		return instr.returnsub_(globalState)
	} else if instr.Opcode == "PC" {
		return instr.pc_(globalState)
	} else if instr.Opcode == "MSIZE" {
		return instr.msize_(globalState)
	} else if instr.Opcode == "GAS" {
		return instr.gas_(globalState)
	} else if instr.Opcode == "CREATE" {
		return instr.create_(globalState)
	} else if instr.Opcode == "CREATE2" {
		return instr.create2_(globalState)
	} else if instr.Opcode == "RETURN" {
		instr.return_(globalState)
		ret := make([]*state.GlobalState, 0)
		return ret
	} else if instr.Opcode == "SELFDESTRUCT" {
		instr.selfdestruct_(globalState)
		ret := make([]*state.GlobalState, 0)
		//ret = append(ret, globalState)
		return ret
	} else if instr.Opcode == "REVERT" {
		instr.revert_(globalState)
		ret := make([]*state.GlobalState, 0)
		return ret
	} else if instr.Opcode == "ASSERTFAIL" {
		instr.assert_fail_(globalState)
		ret := make([]*state.GlobalState, 0)
		return ret
	} else if instr.Opcode == "INVALID" {
		instr.invalid_(globalState)
		ret := make([]*state.GlobalState, 0)
		return ret
	} else if instr.Opcode == "STOP" {
		instr.stop_(globalState)
		ret := make([]*state.GlobalState, 0)
		return ret
	} else if instr.Opcode == "CALL" {
		return instr.call_(globalState)
	} else if instr.Opcode == "CALLCODE" {
		return instr.callcode_(globalState)
	} else if instr.Opcode == "DELEGATECALL" {
		return instr.delegatecall_(globalState)
	} else if instr.Opcode == "STATICCALL" {
		return instr.staticcall_(globalState)
	} else {
		panic("?" + instr.Opcode)
	}
	return nil
}

func (instr *Instruction) jumpdest_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)
	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) push_(globalState *state.GlobalState) []*state.GlobalState {
	mstate := globalState.Mstate
	ret := make([]*state.GlobalState, 0)

	pushInstruction := globalState.GetCurrentInstruction()
	pushValue := pushInstruction.Argument[2:]
	// TODO: check length

	//pushInt, _ := strconv.ParseInt(pushValue, 16, 64)
	// should use big int here
	pushInt, _ := new(big.Int).SetString(pushValue, 16)
	mstate.Stack.Append(globalState.Z3ctx.NewBitvecVal(pushInt, 256))
	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) dup_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)
	value, _ := strconv.ParseInt(globalState.GetCurrentInstruction().OpCode.Name[3:], 10, 64)

	mstate := globalState.Mstate
	// warning: panic: runtime error: index out of range [-2]
	mstate.Stack.Append(mstate.Stack.RawStack[mstate.Stack.Length()-int(value)].Copy())
	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) swap_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	length := mstate.Stack.Length()
	depth, _ := strconv.ParseInt(globalState.GetCurrentInstruction().OpCode.Name[4:], 10, 64)
	mstate.Stack.RawStack[length-int(depth)-1], mstate.Stack.RawStack[length-1] = mstate.Stack.RawStack[length-1], mstate.Stack.RawStack[length-int(depth)-1]

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
			op1str = op1.Simplify().BvString()
		} else {
			op1str = op1.Value()
		}
		result = globalState.NewBitvec(op1str+"["+op0.BvString()+"]", 256)
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
	fmt.Println("exp:", globalState)

	//if (base.ValueInt() == 0 && base.BvString() != "#x0000000000000000000000000000000000000000000000000000000000000000") || (exponent.ValueInt() == 0 && exponent.BvString() != "#x0000000000000000000000000000000000000000000000000000000000000000") {
	if base.Symbolic() || exponent.Symbolic() {
		res := globalState.NewBitvec("invhash("+base.BvString()+")**invhash("+exponent.BvString()+")", 256)
		res.Annotations = base.Annotations.Union(exponent.Annotations)
		mstate.Stack.Append(res)
	} else {
		exponentiation, constraint := function_managers.CreateCondition(base, exponent, globalState.Z3ctx)
		mstate.Stack.Append(exponentiation)
		globalState.WorldState.Constraints.Add(constraint)
	}

	ret = append(ret, globalState)
	return ret
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
	exp := op1.BvUGt(op2)
	mstate.Stack.Append(exp)

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
	result := z3.If(exp, ctx.NewBitvecVal(1, 256), ctx.NewBitvecVal(0, 256))
	mstate.Stack.Append(result.Simplify())

	ret = append(ret, globalState)
	return ret
}

// Call data
func (instr *Instruction) callvalue_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	//env := globalState.Environment
	//mstate.Stack.Append(globalState.Z3ctx.NewBitvecVal(env.CallValue, 256))
	mstate.Stack.Append(globalState.Z3ctx.NewBitvec("call_value"+globalState.CurrentTransaction().GetId(), 256))
	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) calldataload_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	env := globalState.Environment
	op0 := mstate.Stack.Pop()
	value := env.Calldata.GetWordAt(op0)
	//fmt.Println("calldataload_:", value)
	//fmt.Println("index:", op0.BvString())
	//fmt.Println("value:", value.BvString())
	mstate.Stack.Append(value.Translate(globalState.Z3ctx))

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

	if dstart.Symbolic() {
		fmt.Println("Unsupported symbolic calldata offset in CALLDATACOPY")
		dstart = dstart.Simplify()
	}

	var sizeV int
	if size.Symbolic() {
		fmt.Println("Unsupported symbolic size in CALLDATACOPY")
		sizeV = 320
	} else {
		sizeValue, _ := strconv.Atoi(size.Value())
		sizeV = sizeValue
	}
	fmt.Println("calldatacopy!")

	if sizeV > 0 {
		fmt.Println("size>0", sizeV)
		mstate.MemExtend(mstart, sizeV)
		// TODO: TypeError check for memExtend()

		iData := dstart
		newMemory := make([]*z3.Bitvec, 0)
		for i := 0; i < sizeV; i++ {
			value := env.Calldata.Load(iData)
			newMemory = append(newMemory, value)
			//idataValue, _ := strconv.Atoi(iData.Value())
			//iData = ctx.NewBitvecVal(idataValue+1, iData.BvSize())
			iData = iData.BvAdd(ctx.NewBitvecVal(1, iData.BvSize())).Simplify()
		}

		fmt.Println("idata")
		//mstartValue, _ := strconv.ParseInt(mstart.Value(), 10, 64)
		//fmt.Println(mstartValue, "-mstartOffset ", len(newMemory), "-length")

		for j := 0; j < len(newMemory); j++ {
			//mstate.Memory.SetItem(mstartValue+int64(j), newMemory[j])
			offset := mstart.BvAdd(ctx.NewBitvecVal(j, 256)).Simplify()
			mstate.Memory.SetItem(offset, newMemory[j])
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

	tx := reflect.TypeOf(globalState.CurrentTransaction())
	if tx.String() == "*state.ContractCreationTransaction" {
		fmt.Println("Attempt to use CALLDATACOPY in creation transaction")
		ret := make([]*state.GlobalState, 0)
		ret = append(ret, globalState)
		return ret
	}

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
		balance = globalState.WorldState.AccountsExistOrLoad(address).Balance()
	} else {
		balance = ctx.NewBitvecVal(0, 256)
		for _, acc := range globalState.WorldState.Accounts {
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

func (instr *Instruction) codesize_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	env := globalState.Environment
	disassembly := env.Code
	calldata := env.Calldata
	var noOfBytes int
	switch globalState.CurrentTransaction().(type) {
	case *state.ContractCreationTransaction:
		noOfBytes = len(disassembly.Bytecode) / 2
		switch calldata.(type) {
		case *state.ConcreteCalldata:
			sizeV, _ := strconv.Atoi(calldata.Size().Value())
			noOfBytes += sizeV
		default:
			noOfBytes += 0x200
			tmp := globalState.Z3ctx.NewBitvecVal(noOfBytes, 256)
			globalState.WorldState.Constraints.Add(calldata.Size().Eq(tmp))
		}
	default:
		noOfBytes = len(disassembly.Bytecode) / 2
	}
	mstate.Stack.Append(noOfBytes)

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

func (instr *Instruction) sha3_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	index := mstate.Stack.Pop()
	op1 := mstate.Stack.Pop()

	var length int
	//if op1.Symbolic() {
	if op1.ValueInt() == 0 && op1.BvString() != "#x0000000000000000000000000000000000000000000000000000000000000000" {
		length = 64
		fmt.Println("Can't access symbolic memory offsets")
		globalState.WorldState.Constraints.Add(op1.Eq(globalState.Z3ctx.NewBitvecVal(length, 256)))
	} else {
		lengthV, _ := strconv.Atoi(op1.Value())
		length = lengthV
	}
	instr._sha3_gas_helper(globalState, length)

	mstate.MemExtend(index, length)
	indexV, _ := strconv.ParseInt(index.Value(), 10, 64)
	dataList := mstate.Memory.GetItems(indexV, indexV+int64(length), globalState.Z3ctx)
	var data *z3.Bitvec
	if len(dataList) > 1 {
		//data = dataList[0].Translate(globalState.Z3ctx)
		data = dataList[0]
		for i := 1; i < len(dataList); i++ {
			//data = data.Concat(dataList[i].Translate(globalState.Z3ctx)).Simplify()
			data = data.Concat(dataList[i]).Simplify()
		}
	} else if len(dataList) == 1 {
		//data = dataList[0].Translate(globalState.Z3ctx).Simplify()
		data = dataList[0].Simplify()
	} else {
		result := function_managers.NewKeccakFunctionManager(globalState.Z3ctx).GetEmptyKeccakHash()
		mstate.Stack.Append(result)
		// todo
		globalState.WorldState.Constraints.Add(globalState.Z3ctx.NewBitvecVal(1, 256).Eq(globalState.Z3ctx.NewBitvecVal(1, 256)).Simplify())
		ret = append(ret, globalState)
		fmt.Println("GetEmptyKeccakHash")

		return ret
	}

	result, cons := function_managers.NewKeccakFunctionManager(globalState.Z3ctx).CreateKeccak(data.Translate(globalState.Z3ctx))
	//result := globalState.Z3ctx.NewBitvecVal(1,256)
	//cons := globalState.Z3ctx.NewBitvecVal(1, 256).Eq(globalState.Z3ctx.NewBitvecVal(1, 256)).Simplify()
	mstate.Stack.Append(result)
	globalState.WorldState.Constraints.Add(cons)
	ret = append(ret, globalState)
	fmt.Println("CreateKeccak")

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

// TODO: test
func (instr *Instruction) codecopy_(globalState *state.GlobalState) []*state.GlobalState {

	memoryOffset := globalState.Mstate.Stack.Pop()
	codeOffset := globalState.Mstate.Stack.Pop()
	size := globalState.Mstate.Stack.Pop()
	code := globalState.Environment.Code.Bytecode
	codeSize := len(code) / 2
	switch globalState.CurrentTransaction().(type) {
	case *state.ContractCreationTransaction:
		mstate := globalState.Mstate
		codeOffsetV, _ := strconv.Atoi(codeOffset.Value())
		offset := codeOffsetV - codeSize
		fmt.Println("codeOffset-", codeOffsetV, " codesize-", codeSize)
		fmt.Println("Copying from code offset:", offset, " with size: ", codeSize)

		switch globalState.Environment.Calldata.(type) {
		case *state.SymbolicCalldata:
			if codeOffsetV >= codeSize {
				return instr._calldata_copy_helper(globalState, mstate,
					memoryOffset, globalState.Z3ctx.NewBitvecVal(offset, 256), size)
			}
		default:
			// Copy from both code and calldata appropriately.
			concreteCodeOffset, _ := strconv.Atoi(codeOffset.Value())
			concreteSize, _ := strconv.Atoi(size.Value())

			codeCopyOffset := codeOffset
			var codeCopySize int
			if concreteCodeOffset+concreteSize <= codeSize {
				codeCopySize = concreteSize
			} else {
				codeCopySize = codeSize - concreteCodeOffset
			}
			if codeCopySize < 0 {
				codeCopySize = 0
			}

			var calldataCopyOffset int

			if concreteCodeOffset-codeSize > 0 {
				calldataCopyOffset = concreteCodeOffset - codeSize
			} else {
				// TODO:0?
				//calldataCopyOffset = 0
				calldataCopyOffset = 101
			}

			calldataCopySize := concreteCodeOffset + concreteSize - codeSize
			if calldataCopySize < 0 {
				calldataCopySize = 0
			}
			codeCopySizeBv := globalState.Z3ctx.NewBitvecVal(codeCopySize, 256)
			calldataCopyOffsetBv := globalState.Z3ctx.NewBitvecVal(calldataCopyOffset, 256)
			calldataCopySizeBv := globalState.Z3ctx.NewBitvecVal(calldataCopySize, 256)
			// code_copy_helper has no problem
			globalStateArr := instr._code_copy_helper(globalState.Environment.Code.Bytecode,
				memoryOffset, codeCopyOffset, codeCopySizeBv, "CODECOPY", globalState)
			fmt.Println(memoryOffset.BvAdd(codeCopySizeBv).Simplify().Value())
			fmt.Println(calldataCopyOffsetBv.Value())
			fmt.Println(calldataCopySizeBv.Value())
			return instr._calldata_copy_helper(globalStateArr[0], mstate,
				memoryOffset.BvAdd(codeCopySizeBv).Simplify(), calldataCopyOffsetBv, calldataCopySizeBv)
		}
	default:
		return instr._code_copy_helper(code, memoryOffset, codeOffset, size, "CODECOPY", globalState)
	}
	return instr._code_copy_helper(code, memoryOffset, codeOffset, size, "CODECOPY", globalState)
}

func (instr *Instruction) _code_copy_helper(code []byte, memoryOffset *z3.Bitvec, codeOffset *z3.Bitvec,
	size *z3.Bitvec, op string, globalState *state.GlobalState) []*state.GlobalState {

	ret := make([]*state.GlobalState, 0)

	if memoryOffset.Symbolic() {
		fmt.Println("Unsupported symbolic memory offset in " + op)
		ret = append(ret, globalState)
		return ret
	}
	//mOffsetV, _ := strconv.ParseInt(memoryOffset.Value(), 10, 64)

	if size.Symbolic() {
		globalState.Mstate.MemExtend(memoryOffset, 1)
		globalState.Mstate.Memory.SetItem(memoryOffset,
			globalState.NewBitvec("code("+globalState.Environment.ActiveAccount.ContractName+")", 8))
		ret = append(ret, globalState)
		return ret
	}
	sizeV, _ := strconv.ParseInt(size.Value(), 10, 64)
	globalState.Mstate.MemExtend(memoryOffset, int(sizeV))

	if codeOffset.Symbolic() {
		fmt.Println("Unsupported symbolic code offset in " + op)
		globalState.Mstate.MemExtend(memoryOffset, int(sizeV))
		for i := 0; i < int(sizeV); i++ {
			globalState.Mstate.Memory.SetItem(memoryOffset.BvAdd(globalState.Z3ctx.NewBitvecVal(i, 256)).Simplify(),
				globalState.NewBitvec("code("+globalState.Environment.ActiveAccount.ContractName+")", 8))

		}
		ret = append(ret, globalState)
		return ret
	}
	cOffsetV, _ := strconv.Atoi(codeOffset.Value())

	for i := 0; i < int(sizeV); i++ {
		if (cOffsetV + i + 1) > len(code) {
			break
		}
		globalState.Mstate.Memory.SetItem(memoryOffset.BvAdd(globalState.Z3ctx.NewBitvecVal(i, 256)).Simplify(),
			globalState.Z3ctx.NewBitvecVal(int(code[cOffsetV+i]), 8))
	}

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) extcodesize_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	addr := mstate.Stack.Pop()

	if addr.ValueInt() == 0 && addr.BvString() != "#x0000000000000000000000000000000000000000000000000000000000000000" {
		//if addr.Symbolic() {
		// TypeError
		fmt.Println("unsupported symbolic address for EXTCODESIZE")
		mstate.Stack.Append(globalState.NewBitvec("extcodesize_"+addr.BvString(), 256))
		ret = append(ret, globalState)
		return ret
	}
	code := globalState.WorldState.AccountsExistOrLoad(addr).Code
	if globalState.WorldState.AccountsExistOrLoad(addr).Code == nil {
		fmt.Println("code nil!!!")
	} else {
		//mstate.Stack.Append(globalState.Z3ctx.NewBitvecVal(len(code.Bytecode) / 2, 256))
		mstate.Stack.Append(len(code.Bytecode) / 2)
	}

	ret = append(ret, globalState)
	return ret
}

// TODO
func (instr *Instruction) extcodecopy_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	ret = append(ret, globalState)
	return ret
}

// TODO
func (instr *Instruction) extcodehash_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) returndatacopy_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	//ctx := globalState.Z3ctx
	mstate := globalState.Mstate
	memoryOffset := mstate.Stack.Pop()
	returnOffset := mstate.Stack.Pop()
	size := mstate.Stack.Pop()

	if memoryOffset.Symbolic() {
		fmt.Println("Unsupported symbolic memory offset in RETURNDATACOPY")
		ret = append(ret, globalState)
		return ret
	}
	if returnOffset.Symbolic() {
		fmt.Println("Unsupported symbolic return offset in RETURNDATACOPY")
		ret = append(ret, globalState)
		return ret
	}
	if size.Symbolic() {
		fmt.Println("Unsupported symbolic max_length offset in RETURNDATACOPY")
		ret = append(ret, globalState)
		return ret
	}
	if globalState.LastReturnData == nil {
		ret = append(ret, globalState)
		return ret
	}
	//mOffset, _ := strconv.ParseInt(memoryOffset.Value(), 10, 64)
	rOffset, _ := strconv.ParseInt(returnOffset.Value(), 10, 64)
	sizeV, _ := strconv.ParseInt(size.Value(), 10, 64)

	mstate.MemExtend(memoryOffset, int(sizeV))
	lastReturnData := *globalState.LastReturnData
	for i := 0; i < int(sizeV); i++ {
		var value *z3.Bitvec
		if i < len(lastReturnData) {
			value = lastReturnData[rOffset+int64(i)]
		} else {
			value = globalState.Z3ctx.NewBitvecVal(0, 8)
		}
		//mstate.Memory.SetItem(mOffset+int64(i), value)
		offset := memoryOffset.BvAdd(globalState.Z3ctx.NewBitvecVal(i, 256)).Simplify()
		mstate.Memory.SetItem(offset, value)
	}

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) returndatasize_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	if globalState.LastReturnData == nil {
		fmt.Println("No last_return_data found, adding an unconstrained bitvec to the stack")
		globalState.Mstate.Stack.Append(globalState.NewBitvec("returndatasize", 256))
	} else {
		ctx := globalState.Z3ctx
		globalState.Mstate.Stack.Append(ctx.NewBitvecVal(len(*globalState.LastReturnData), 256))
	}

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) blockhash_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	blocknumber := mstate.Stack.Pop()
	mstate.Stack.Append(globalState.NewBitvec("blockhash_block_"+blocknumber.BvString(), 256))

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) coinbase_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	globalState.Mstate.Stack.Append(globalState.NewBitvec("coinbase", 256))

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) timestamp_(globalState *state.GlobalState) []*state.GlobalState {
	// In EVM, it will push the current time in stack
	// In mythril, it will push a symbolic bv in stack
	ret := make([]*state.GlobalState, 0)
	globalState.Mstate.Stack.Append(globalState.NewBitvec("timestamp", 256))
	// TODO: for test here
	//globalState.Mstate.Stack.Append(globalState.Z3ctx.NewBitvecVal(7, 256))
	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) number_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	//globalState.Mstate.Stack.Append(globalState.Environment.BlockNumber)
	globalState.Mstate.Stack.Append(globalState.Z3ctx.NewBitvec("number", 256))

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) difficulty_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	globalState.Mstate.Stack.Append(globalState.NewBitvec("block_difficulty", 256))

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) gaslimit_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	ctx := globalState.Z3ctx
	globalState.Mstate.Stack.Append(ctx.NewBitvecVal(globalState.Mstate.GasLimit, 256))

	ret = append(ret, globalState)
	return ret
}

// Memory operations
func (instr *Instruction) mload_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	offset := mstate.Stack.Pop()

	//fmt.Println("mload_")
	//fmt.Println("offsetSymbolic:", offset.Symbolic())
	//fmt.Println("offset:", offset.BvString())

	//if offset.Symbolic() {
	//	fmt.Println("can't access memory with symbolic index!")
	//	mstate.Stack.Append(globalState.Z3ctx.NewBitvecVal(0, 256))
	//	//return ret
	//} else {
	//	mstate.MemExtend(offset, 32)
	//	offsetV, _ := strconv.ParseInt(offset.Value(), 10, 64)
	//
	//	data := mstate.Memory.GetWordAt(offsetV)
	//
	//	fmt.Println("data:", data.BvString())
	//
	//	mstate.Stack.Append(data)
	//}

	mstate.MemExtend(offset, 32)

	data := mstate.Memory.GetWordAt(offset)

	//fmt.Println("data:", data.BvString())

	mstate.Stack.Append(data)

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) mstore_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	mstart := mstate.Stack.Pop()
	value := mstate.Stack.Pop()

	//fmt.Println("mstore_")
	//fmt.Println("mstartSymbolic:", mstart.Symbolic())
	//fmt.Println("mstart:", mstart.BvString())
	//fmt.Println("value:", value.BvString())

	//if mstart.Symbolic() {
	//	fmt.Println("fail for mstore because of the symbolic index")
	//	ret = append(ret, globalState)
	//	return ret
	//}

	mstate.MemExtend(mstart, 32)
	//mstartV, _ := strconv.ParseInt(mstart.Value(), 10, 64)
	//mstate.Memory.WriteWordAt(mstartV, value)
	mstate.Memory.WriteWordAt(mstart, value)

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) mstore8_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	offset := mstate.Stack.Pop()
	value := mstate.Stack.Pop()

	mstate.MemExtend(offset, 1)
	value2write := value.Extract(7, 0)

	//offsetV, _ := strconv.ParseInt(offset.Value(), 10, 64)
	//mstate.Memory.SetItem(offsetV, value2write)
	mstate.Memory.SetItem(offset, value2write)
	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) sload_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	index := mstate.Stack.Pop()
	// TODO: DynLoader to get the storage ?
	//globalState.Environment.ActiveAccount.Storage.SetItem(index, globalState.Z3ctx.NewBitvecVal(0, 256))
	//fmt.Println("sload_")
	//fmt.Println("index:", index.BvString())
	//fmt.Println("value:", globalState.Environment.ActiveAccount.Storage.GetItem(index).Translate(globalState.Z3ctx).BvString())
	mstate.Stack.Append(globalState.Environment.ActiveAccount.Storage.GetItem(index).Translate(globalState.Z3ctx))
	//mstate.Stack.Append(globalState.Z3ctx.NewBitvec("sload_",256))
	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) sstore_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	index := mstate.Stack.Pop()
	value := mstate.Stack.Pop()

	//fmt.Println("sstore_:")
	//fmt.Println("index:", index.BvString(), "-", value.BvString())
	globalState.Environment.ActiveAccount.Storage.SetItem(index, value)

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) jump_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	disassembly := globalState.Environment.Code
	jumpAddr := mstate.Stack.Pop()
	if jumpAddr.Symbolic() {
		//if jumpAddr.ValueInt() == 0 && jumpAddr.BvString() != "#x0000000000000000000000000000000000000000000000000000000000000000" {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Catch-Invalid jump argument (symbolic address)")
			}
		}()
		panic("Invalid jump argument (symbolic address)")
	}
	jumpAddrV, _ := strconv.ParseInt(jumpAddr.Value(), 10, 64)
	index := GetInstructionIndex(disassembly.InstructionList, int(jumpAddrV))
	if index == -1 {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Catch-JUMP to invalid address")
			}
		}()
		panic("JUMP to invalid address")
	}
	opCode := disassembly.InstructionList[index].OpCode

	if opCode.Name != "JUMPDEST" {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Catch-Skipping JUMP to invalid destination (not JUMPDEST): " + jumpAddr.BvString())
			}
		}()
		panic("Skipping JUMP to invalid destination (not JUMPDEST): " + jumpAddr.BvString())
	}

	newState := globalState.Copy()
	// "jump" address 0x56
	minGas, maxGas := GetOpcodeGas(0x56)
	newState.Mstate.MinGasUsed += minGas
	newState.Mstate.MaxGasUsed += maxGas
	// manually set PC to destination
	newState.Mstate.Pc = index
	//newState.Mstate.LastPc = globalState.Mstate.Pc
	newState.Mstate.Depth += 1

	ret = append(ret, newState)
	return ret
}

func (instr *Instruction) jumpi_(globalState *state.GlobalState) []*state.GlobalState {

	ret := make([]*state.GlobalState, 0)

	ctx := globalState.Z3ctx
	mstate := globalState.Mstate
	disassembly := globalState.Environment.Code
	// "jumpi" address 0x57
	minGas, maxGas := GetOpcodeGas(0x57)
	op0 := mstate.Stack.Pop()
	condition := mstate.Stack.Pop()

	if op0.Symbolic() {
		fmt.Println("Skipping JUMPI to invalid destination.")
		mstate.Pc += 1
		mstate.MinGasUsed += minGas
		mstate.MaxGasUsed += maxGas
		ret = append(ret, globalState)
		return ret
	}

	jumpAddr, _ := strconv.ParseInt(op0.Value(), 10, 64)
	zero := ctx.NewBitvecVal(0, 256)

	negated := condition.Eq(zero).Simplify()
	condi := condition.Neq(zero).Simplify()
	//condi := condition.Eq(zero).Not().Simplify()
	//negated := condition.Eq(zero)
	//condi := condition.Eq(zero).Not()
	negatedCond := !negated.IsFalse()
	positiveCond := !condi.IsFalse()

	// False case
	if negatedCond {
		// fmt.Println("test negative nil")
		newState := globalState.Copy()
		newState.Mstate.MinGasUsed += minGas
		newState.Mstate.MaxGasUsed += maxGas
		// manually increment PC
		newState.Mstate.Depth += 1
		newState.Mstate.Pc += 1
		newState.WorldState.Constraints.Add(negated)

		//returnData := make(map[int64]*z3.Bitvec)
		//returnData[0] = newState.Z3ctx.NewBitvecVal(0, 256)
		//newState.LastReturnData = &returnData

		//tmpCons := state.NewConstraints()
		//for i, c := range newState.WorldState.Constraints.ConstraintList {
		//	if i==36 {
		//		continue
		//	}
		//	tmpCons.Add(c)
		//}

		// fmt.Println("negativeState:", newState)
		ret = append(ret, newState)
		//if newState.WorldState.Constraints.IsPossible() {
		//	//if tmpCons.IsPossible() {
		//	ret = append(ret, newState)
		//} else {
		//	fmt.Println("negativeStateGet, but ws.Constraints isn't possible")
		//}
		//if globalState.GetCurrentInstruction().Address == 1005 {
		//	fmt.Println("Jumpi1005 negativeCONSTRAINT:", newState.WorldState.Constraints.IsPossible(),  "haha", tmpCons.IsPossible())
		//	for i, c := range newState.WorldState.Constraints.ConstraintList {
		//		if i==9 || i==15 || i==18 || i==21 || i==32 || i==36 {
		//			fmt.Println(i,":", c.BoolString())
		//			continue
		//		}
		//		idx := strings.Index(c.BoolString(), "\n")
		//		if idx == -1 {
		//			fmt.Println(i,":", c.BoolString())
		//		}else{
		//			fmt.Println(i,":", c.BoolString()[:idx])
		//		}
		//	}
		//}

	} else {
		fmt.Println("Pruned unreachable states-negative.")
	}

	// True case
	index := GetInstructionIndex(disassembly.InstructionList, int(jumpAddr))
	if index == -1 {
		fmt.Println("Invalid jump destination: " + op0.BvString())
		return ret
	}
	instruction := disassembly.InstructionList[index]
	if instruction.OpCode.Name == "JUMPDEST" {
		if positiveCond {
			// fmt.Println("test positive nil")

			newState := globalState.Copy()
			newState.Mstate.MinGasUsed += minGas
			newState.Mstate.MaxGasUsed += maxGas

			newState.Mstate.Pc = index
			newState.Mstate.LastPc = globalState.Mstate.Pc
			newState.Mstate.Depth += 1
			newState.WorldState.Constraints.Add(condi)

			//returnData := make(map[int64]*z3.Bitvec)
			//returnData[0] = newState.Z3ctx.NewBitvecVal(1, 256)
			//newState.LastReturnData = &returnData

			// fmt.Println("positiveState:", newState)

			//if globalState.GetCurrentInstruction().Address == 1005 {
			//	fmt.Println("Jumpi1005 positiveCONSTRAINT:", newState.WorldState.Constraints.IsPossible())
			//	for i, c := range newState.WorldState.Constraints.ConstraintList {
			//		idx := strings.Index(c.BoolString(), "\n")
			//		if idx == -1 {
			//			fmt.Println(i,":", c.BoolString())
			//		}else{
			//			fmt.Println(i,":", c.BoolString()[:idx])
			//		}
			//	}
			//}
			ret = append(ret, newState)
			//if newState.WorldState.Constraints.IsPossible() {
			//	//if tmpCons.IsPossible(){
			//	ret = append(ret, newState)
			//} else {
			//	fmt.Println("positiveStateGet, but ws.Constraints isn't possible")
			//}
		} else {
			fmt.Println("Pruned unreachable states-positive")
		}
	}
	return ret
}

// TODO
func (instr *Instruction) beginsub_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	ret = append(ret, globalState)
	return ret
}

// TODO
func (instr *Instruction) jumpsub_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	ret = append(ret, globalState)
	return ret
}

// TODO
func (instr *Instruction) returnsub_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) pc_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	ctx := globalState.Z3ctx

	index := globalState.Mstate.Pc
	programCounter := globalState.Environment.Code.InstructionList[index].Address
	globalState.Mstate.Stack.Append(ctx.NewBitvecVal(programCounter, 256))

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) msize_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	globalState.Mstate.Stack.Append(globalState.Mstate.MemorySize())

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) gas_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)
	globalState.Mstate.Stack.Append(globalState.NewBitvec("gas", 256))
	//globalState.Mstate.Stack.Append(globalState.Z3ctx.NewBitvecVal(6431, 256))
	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) log_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	mstate := globalState.Mstate
	depth := instr.Opcode[3:]
	depthV, _ := strconv.ParseInt(depth, 10, 64)

	mstate.Stack.Pop()
	mstate.Stack.Pop()
	for i := 0; i < int(depthV); i++ {
		mstate.Stack.Pop()
	}

	ret = append(ret, globalState)
	return ret
}

// TODO
func (instr *Instruction) _create_transaction_helper(globalState *state.GlobalState, callValue *z3.Bitvec,
	memOffset *z3.Bitvec, memSize *z3.Bitvec) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	/*	mstate := globalState.Mstate
		env := globalState.Environment
		worldState := globalState.WorldState

		callData := GetCallData(globalState, memOffset, memOffset.BvAdd(memSize).Simplify())
		codeRaw := make([]string, 0)
		codeEnd, _ := strconv.Atoi(callData.Size().Value())
		size := callData.Size()
		var sizeV int64
		if size.Symbolic() {
			sizeV = int64(math.Pow(10, 5))
		} else {
			value, _ := strconv.ParseInt(size.Value(), 10, 64)
			sizeV = value
		}
		for i := 0; int64(i) < sizeV; i++ {
			item := callData.Load(globalState.Z3ctx.NewBitvecVal(i, 256))
			if item.Symbolic() {
				codeEnd = i
				break
			}
			codeRaw = append(codeRaw, item.Value())
		}
		if len(codeRaw) < 1 {
			globalState.Mstate.Stack.Append(1)
			fmt.Println("No code found for trying to execute a create type instruction.")
			ret = append(ret, globalState)
			return ret
		}
		nextTxId := state.GetNextTransactionId()
		constructorArgs := state.NewConcreteCalldata(nextTxId, callData)*/

	ret = append(ret, globalState)
	return ret
}

// TODO
func (instr *Instruction) create_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	ret = append(ret, globalState)
	return ret
}

// TODO
func (instr *Instruction) create2_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	ret = append(ret, globalState)
	return ret
}

// TODO
func (instr *Instruction) _handle_create_type_post(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) return_(globalState *state.GlobalState) {
	mstate := globalState.Mstate
	offset := mstate.Stack.Pop()
	length := mstate.Stack.Pop()
	var returnData []byte
	if length.Symbolic() {
		//returnData = append(returnData, globalState.NewBitvec("return_data", 8))
		fmt.Println("Return with symbolic length or offset. Not supported")
	} else {
		lenV, _ := strconv.ParseInt(length.Value(), 10, 64)
		mstate.MemExtend(offset, int(lenV))
		CheckGasUsageLimit(globalState)
		offsetV, _ := strconv.ParseInt(offset.Value(), 10, 64)
		returnData = mstate.Memory.GetItems2Bytes(offsetV, offsetV+lenV, globalState.Z3ctx)
	}
	globalState.CurrentTransaction().End(globalState, returnData)
}

func (instr *Instruction) selfdestruct_(globalState *state.GlobalState) {

	target := globalState.Mstate.Stack.Pop()
	transferAmount := globalState.Environment.ActiveAccount.Balance()
	originBalnace := globalState.WorldState.Balances.GetItem(target)
	globalState.WorldState.Balances.SetItem(target, originBalnace.BvAdd(transferAmount).Simplify())

	// TODO: deepcopy account?

	globalState.Environment.ActiveAccount.SetBalance(globalState.Z3ctx.NewBitvecVal(0, 256))
	globalState.Environment.ActiveAccount.Deleted = true
	items := make([]byte, 0)
	globalState.CurrentTransaction().End(globalState, items)
}

func (instr *Instruction) revert_(globalState *state.GlobalState) {
	mstate := globalState.Mstate
	offset := mstate.Stack.Pop()
	length := mstate.Stack.Pop()
	var returnData []byte
	// returnData = append(returnData, globalState.Z3ctx.NewBitvec("return_data", 8))
	if offset.Symbolic() || length.Symbolic() {
		fmt.Println("Return with symbolic length or offset. Not supported")
	} else {
		startV, _ := strconv.ParseInt(offset.Value(), 10, 64)
		lengthV, _ := strconv.ParseInt(length.Value(), 10, 64)
		returnData = mstate.Memory.GetItems2Bytes(startV, startV+lengthV, globalState.Z3ctx)
	}
	globalState.CurrentTransaction().End(globalState, returnData)
}

func (instr *Instruction) assert_fail_(globalState *state.GlobalState) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Catch-InvalidInstruction-assertFail")
		}
	}()
	panic("InvalidInstruction-assertFail")
}

func (instr *Instruction) invalid_(globalState *state.GlobalState) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Catch-InvalidInstruction-invalid")
		}
	}()
	panic("InvalidInstruction-invalid")
}

func (instr *Instruction) stop_(globalState *state.GlobalState) {
	returnData := make([]byte, 0)
	globalState.CurrentTransaction().End(globalState, returnData)
}

func (instr *Instruction) call_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)
	//
	instruction := globalState.GetCurrentInstruction()
	environment := globalState.Environment
	length := globalState.Mstate.Stack.Length()
	memoryOutSize := globalState.Mstate.Stack.RawStack[length-7]
	memoryOutOffset := globalState.Mstate.Stack.RawStack[length-6]

	calleeAddres, calleeAccount, callData, value, gas, memoryOutOffset, memoryOutSize := GetCallParameters(globalState, true)
	fmt.Println("call_", memoryOutSize, memoryOutOffset, calleeAddres, callData, gas, instruction)

	if calleeAccount != nil && len(calleeAccount.Code.Bytecode) == 0 {
		fmt.Println("The call is related to ether transfer between accounts")
		sender := environment.ActiveAccount.Address
		receiver := calleeAccount.Address
		//fmt.Println(sender,receiver,value)
		transferEther(globalState, sender, receiver, value)

		globalState.Mstate.Stack.Append(globalState.NewBitvec("retval_"+strconv.Itoa(instruction.Address), 256))
		//globalState.Mstate.Stack.Append(globalState.Z3ctx.NewBitvecVal(1, 256))
		ret = append(ret, globalState)
		return ret
	}
	// TODO: symbolic ValueError exception
	// TODO: nativeCall
	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) callcode_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	instruction := globalState.GetCurrentInstruction()
	environment := globalState.Environment
	length := globalState.Mstate.Stack.Length()
	memoryOutSize := globalState.Mstate.Stack.RawStack[length-7]
	memoryOutOffset := globalState.Mstate.Stack.RawStack[length-6]

	calleeAddres, calleeAccount, callData, value, gas, memoryOutOffset, memoryOutSize := GetCallParameters(globalState, true)
	fmt.Println("call_", memoryOutSize, memoryOutOffset, calleeAddres, callData, gas, instruction)

	if calleeAccount != nil && len(calleeAccount.Code.Bytecode) == 0 {
		fmt.Println("The call is related to ether transfer between accounts")
		sender := environment.ActiveAccount.Address
		receiver := calleeAccount.Address
		//fmt.Println(sender,receiver,value)
		transferEther(globalState, sender, receiver, value)

		globalState.Mstate.Stack.Append(globalState.NewBitvec("retval_"+strconv.Itoa(instruction.Address), 256))
		//globalState.Mstate.Stack.Append(globalState.Z3ctx.NewBitvecVal(1, 256))
		ret = append(ret, globalState)
		return ret
	} else {
		// TODO: TransactionStartSignal
		fmt.Println("callCodeNewTx")
	}
	// TODO: valueError
	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) delegatecall_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	instruction := globalState.GetCurrentInstruction()
	environment := globalState.Environment
	length := globalState.Mstate.Stack.Length()
	memoryOutSize := globalState.Mstate.Stack.RawStack[length-6]
	memoryOutOffset := globalState.Mstate.Stack.RawStack[length-5]

	_, calleeAccount, callData, value, gas, _, _ := GetCallParameters(globalState, false)

	fmt.Println("delegateCall_:", memoryOutSize, memoryOutOffset, callData, gas)

	if calleeAccount != nil && len(calleeAccount.Code.Bytecode) == 0 {
		fmt.Println("The call is related to ether transfer between accounts")
		sender := environment.ActiveAccount.Address
		receiver := calleeAccount.Address
		//fmt.Println(sender,receiver,value)
		transferEther(globalState, sender, receiver, value)

		globalState.Mstate.Stack.Append(globalState.NewBitvec("retval_"+strconv.Itoa(instruction.Address), 256))
		//globalState.Mstate.Stack.Append(globalState.Z3ctx.NewBitvecVal(1, 256))
		ret = append(ret, globalState)
		return ret
	} else {
		// TODO: TransactionStartSignal
		fmt.Println("delegateCallNewTx")
	}
	// TODO: valueError
	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) staticcall_(globalState *state.GlobalState) []*state.GlobalState {
	ret := make([]*state.GlobalState, 0)

	instruction := globalState.GetCurrentInstruction()
	environment := globalState.Environment
	stack := globalState.Mstate.Stack
	memoryOutSize := stack.RawStack[stack.Length()-6]
	memoryOutOffset := stack.RawStack[stack.Length()-5]

	calleeAddres, calleeAccount, callData, value, gas, memoryOutOffset, memoryOutSize := GetCallParameters(globalState, false)
	fmt.Println("call_", memoryOutSize, memoryOutOffset, calleeAddres, callData, gas, instruction)

	if calleeAccount != nil && len(calleeAccount.Code.Bytecode) == 0 {
		fmt.Println("The call is related to ether transfer between accounts")
		sender := environment.ActiveAccount.Address
		receiver := calleeAccount.Address
		//fmt.Println(sender,receiver,value)
		transferEther(globalState, sender, receiver, value)

		globalState.Mstate.Stack.Append(globalState.NewBitvec("retval_"+strconv.Itoa(instruction.Address), 256))
		//globalState.Mstate.Stack.Append(globalState.Z3ctx.NewBitvecVal(1, 256))
		ret = append(ret, globalState)
		return ret
	} else {
		// TODO: nativeCall & TransactionStartSignal
		fmt.Println("delegateCallNewTx")
	}
	// TODO: valueError
	ret = append(ret, globalState)
	return ret
}
