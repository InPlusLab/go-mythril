package ethereum

import (
	"fmt"
	"go-mythril/laser/ethereum/state"
	"strconv"
	//"go-mythril/laser/smt"
)

type Instruction struct {
	Opcode string
}

func NewInstruction(opcode string) *Instruction {
	return &Instruction{opcode}
}

func (instr *Instruction) Evaluate(globalState *state.GlobalState) []*state.GlobalState {

	// TODO: Pre hook
	result := instr.Mutator(globalState)
	// TODO: Post hook
	fmt.Println("Evaluate", len(result))

	// TODO
	for _, state := range result {
		state.Mstate.Pc++
	}
	return result
}

// using reflect (getattr) might be too complex? maybe if-else is good
func (instr *Instruction) Mutator(globalState *state.GlobalState) []*state.GlobalState {
	// TODO
	if instr.Opcode == "PUSH1" {
		return instr.push_(globalState)
	} else if instr.Opcode == "ORIGIN" {
		return instr.origin_(globalState)
	} else if instr.Opcode == "STOP" {
		return instr.stop_(globalState)
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
	mstate.Stack = append(mstate.Stack, globalState.Z3ctx.NewBitvecVal(uint(pushInt), 256))
	ret = append(ret, globalState)
	return ret
}

func (instr *Instruction) origin_(globalState *state.GlobalState) []*state.GlobalState {
	mstate := globalState.Mstate

	ret := make([]*state.GlobalState, 0)
	// TODO: this append is not right, should be desinged as a trait of mstate
	mstate.Stack = append(mstate.Stack, globalState.Z3ctx.NewBitvecVal(10086, 256))
	ret = append(ret, globalState)
	return ret
}
