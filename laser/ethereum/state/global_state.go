package state

import (
	"go-mythril/disassembler"
	"go-mythril/laser/smt/z3"
)

type GlobalState struct {
	WorldState  *WorldState
	Mstate      *MachineState
	Z3ctx       *z3.Context
	Environment *Environment
}

func NewGlobalState(env *Environment) *GlobalState {
	config := z3.NewConfig()
	ctx := z3.NewContext(config)
	return &GlobalState{
		WorldState:  NewWordState(),
		Mstate:      NewMachineState(),
		Z3ctx:       ctx,
		Environment: env,
	}
}

func (globalState *GlobalState) GetCurrentInstruction() *disassembler.EvmInstruction {
	instructions := globalState.Environment.Code.InstructionList
	pc := globalState.Mstate.Pc
	// Important: PC in mythril seems to be different with others (e.g. Etherscan). Mythril.address = Others.pc. Mythril.pc seems to be the index of the evminstruction.
	if pc < len(instructions) {
		return instructions[pc]
	}
	// TODO
	return nil
}

func (globalState *GlobalState) NewBitvec(name string, size int) *z3.Bitvec {
	// TODO: tx
	// txId := globalState.currentTx.id
	str := "0" + name
	return globalState.Z3ctx.NewBitvec(str, size)
}
