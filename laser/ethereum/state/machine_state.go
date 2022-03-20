package state

import "C"
import (
	"fmt"
	"go-mythril/laser/smt/z3"
	"strconv"
)

const STACK_LIMIT = 1024
const GAS_MEMORY = 3
const GAS_MEMORY_QUADRATIC_DENOMINATOR = 512

// TODO
type MachineStack struct {
	RawStack []*z3.Bitvec
}

func NewMachineStack() *MachineStack {
	stack := make([]*z3.Bitvec, 0)
	return &MachineStack{
		RawStack: stack,
	}
}

func (m *MachineStack) Append(b interface{}) {
	// TODO: STACK LIMIT CHECK
	len := m.Length()
	if len >= STACK_LIMIT {
		panic("Reached the EVM stack limit, you can't reach more.")
	}
	switch b.(type) {
	case *z3.Bitvec:
		// TODO: simplify this bitvec?
		m.RawStack = append(m.RawStack, b.(*z3.Bitvec))
	case *z3.Bool:
		ctx := z3.GetBoolCtx(b.(*z3.Bool))
		tmp := z3.If(b.(*z3.Bool), ctx.NewBitvecVal(1, 256), ctx.NewBitvecVal(0, 256))
		m.RawStack = append(m.RawStack, tmp)
	}
}

func (m *MachineStack) Length() int {
	return len(m.RawStack)
}

func (m *MachineStack) Pop() *z3.Bitvec {
	length := len(m.RawStack)
	if length == 0 {
		panic("trying to pop from an empty stack")
	} else {
		item := m.RawStack[length-1]
		m.RawStack = m.RawStack[:length-1]
		return item
	}
}

// For debug
func (m *MachineStack) PrintStack() {
	if len(m.RawStack) == 0 {
		fmt.Println("PrintStack: 0 instruction")
	}
	for _, item := range m.RawStack {
		if item.GetAstKind() == z3.NumeralKindAST {
			fmt.Println("PrintStack: ", item.Value())
		} else {
			fmt.Println("PrintStack: ", item.String())
		}
	}
}

type MachineState struct {
	// Important: PC in mythril seems to be different with others (e.g. Etherscan). Mythril.address = Others.pc. Mythril.pc seems to be the index of the evminstruction.
	GasLimit   int
	Pc         int
	Stack      *MachineStack
	Memory     *Memory
	MinGasUsed int
	MaxGasUsed int
}

func NewMachineState() *MachineState {
	stack := NewMachineStack()
	return &MachineState{
		GasLimit:   0,
		Pc:         0,
		Stack:      stack,
		Memory:     NewMemory(),
		MinGasUsed: 0,
		MaxGasUsed: 0,
	}
}

func (m *MachineState) CalculateMemorySize(start int, size int) int {
	if m.memorySize() > (start + size) {
		return 0
	}
	// In python, we use // for floor division.
	// In golang, / represents the floor division.
	newSize := Ceil32(start+size) / 32
	oldSize := m.memorySize() / 32
	return (newSize - oldSize) * 32
}

func (m *MachineState) CalculateMemoryGas(start int, size int) int {
	oldSize := m.memorySize() / 32
	oldTotalFee := oldSize*GAS_MEMORY +
		oldSize*oldSize/GAS_MEMORY_QUADRATIC_DENOMINATOR
	newSize := Ceil32(start+size) / 32
	newTotalFee := newSize*GAS_MEMORY +
		newSize*newSize/GAS_MEMORY_QUADRATIC_DENOMINATOR
	return newTotalFee - oldTotalFee
}

func (m *MachineState) CheckGas() {
	if m.MinGasUsed > m.GasLimit {
		panic("OutOfGasException")
	}
}

func (m *MachineState) MemExtend(start *z3.Bitvec, size int) {
	if start.Symbolic() {
		return
	} else {
		startValue, _ := strconv.ParseInt(start.Value(), 10, 64)
		mExtend := m.CalculateMemorySize(int(startValue), size)
		if mExtend != 0 {
			extendGas := m.CalculateMemoryGas(int(startValue), size)
			m.MinGasUsed += extendGas
			m.MaxGasUsed += extendGas
			m.CheckGas()
			m.Memory.Extend(mExtend)
		}
	}
}

func (m *MachineState) memorySize() int {
	return m.Memory.length()
}

// Ceil32 the implementation is in
// https://github.com/ethereum/py-evm/blob/master/eth/_utils/numeric.py
func Ceil32(value int) int {
	remainder := value % 32
	if remainder == 0 {
		return value
	} else {
		return value + 32 - remainder
	}
}
