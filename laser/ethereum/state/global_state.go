package state

import (
	"go-mythril/disassembler"
	"go-mythril/laser/smt/z3"
	"reflect"
)

type GlobalState struct {
	WorldState  *WorldState
	Mstate      *MachineState
	Z3ctx       *z3.Context
	Environment *Environment
	// TODO: tx is not baseTx
	TxStack        []BaseTransaction
	LastReturnData *map[int64]*z3.Bitvec
	Annotations    []StateAnnotation
}

func NewGlobalState(worldState *WorldState, env *Environment, ctx *z3.Context, txStack []BaseTransaction) *GlobalState {

	return &GlobalState{
		WorldState:     worldState,
		Mstate:         NewMachineState(),
		Z3ctx:          ctx,
		Environment:    env,
		TxStack:        txStack,
		LastReturnData: nil,
		Annotations:    make([]StateAnnotation, 0),
	}
}

// TODO: test
func (globalState *GlobalState) Copy() *GlobalState {
	//worldState := globalState.WorldState.Copy()
	//environment := globalState.Environment.Copy()
	//mstate := globalState.Mstate.DeepCopy()
	//var txStack []BaseTransaction
	//// shallow copy seems to be different in python and golang?
	//copy(txStack, globalState.TxStack)
	//var anno []StateAnnotation
	//copy(globalState.Annotations, anno)
	//return &GlobalState{
	//	WorldState:     worldState,
	//	Environment:    environment,
	//	Mstate:         mstate,
	//	TxStack:        txStack,
	//	Z3ctx:          globalState.Z3ctx,
	//	LastReturnData: globalState.LastReturnData,
	//	Annotations:    anno,
	//}
	return &GlobalState{
		WorldState:     globalState.WorldState,
		Environment:    globalState.Environment,
		Mstate:         globalState.Mstate.DeepCopy(),
		TxStack:        globalState.TxStack,
		Z3ctx:          globalState.Z3ctx,
		LastReturnData: globalState.LastReturnData,
		Annotations:    globalState.Annotations,
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

func (globalState *GlobalState) CurrentTransaction() BaseTransaction {
	length := len(globalState.TxStack)
	if length != 0 {
		return globalState.TxStack[length-1]
	} else {
		return nil
	}
}

func (globalState *GlobalState) NewBitvec(name string, size int) *z3.Bitvec {
	txId := globalState.CurrentTransaction().GetId()
	str := txId + "_" + name
	return globalState.Z3ctx.NewBitvec(str, size)
}

func (globalState *GlobalState) Annotate(annotation StateAnnotation) {
	globalState.Annotations = append(globalState.Annotations, annotation)
	// TODO:
	//if annotation.PersistToWorldState(){
	//
	//}
}

func (globalState *GlobalState) GetAnnotations(annoType reflect.Type) []StateAnnotation {
	annoList := make([]StateAnnotation, 0)
	for _, v := range globalState.Annotations {
		if annoType == reflect.TypeOf(v) {
			annoList = append(annoList, v)
		}
	}
	return annoList
}
