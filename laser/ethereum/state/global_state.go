package state

import (
	"fmt"
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
	newAnnotations := make([]StateAnnotation, 0)
	for _, anno := range globalState.Annotations {
		newAnnotations = append(newAnnotations, anno)
	}
	return &GlobalState{
		WorldState:     globalState.WorldState.Copy(),
		Environment:    globalState.Environment,
		Mstate:         globalState.Mstate.DeepCopy(),
		TxStack:        globalState.TxStack,
		Z3ctx:          globalState.Z3ctx,
		LastReturnData: globalState.LastReturnData,
		Annotations:    newAnnotations,
		//LastReturnData: nil,
		//Annotations: make([]StateAnnotation, 0),
	}
}

func (globalState *GlobalState) Translate(ctx *z3.Context) {
	if globalState.Z3ctx.GetRaw() == ctx.GetRaw() {
		return
	}
	fmt.Println("before changeStateContext")
	globalState.Z3ctx = ctx
	// machineState stack & memory
	//fmt.Println("glTrans1")
	newMachineState := globalState.Mstate.Translate(ctx)
	globalState.Mstate = newMachineState
	// worldState constraints
	//fmt.Println("glTrans2")
	newWorldState := globalState.WorldState.Translate(ctx)
	globalState.WorldState = newWorldState
	// env
	//fmt.Println("glTrans3")
	newEnv := globalState.Environment.Translate(ctx)
	globalState.Environment = newEnv
	//fmt.Println("glTrans4")
	// lastReturnData
	//fmt.Println("changeStateContext done")
}

//func changeStateContext(globalState *state.GlobalState, ctx *z3.Context) {
//	fmt.Println("before changeStateContext")
//	globalState.Z3ctx = ctx
//	// machineState stack & memory
//	newMachineState := globalState.Mstate.Translate(ctx)
//	globalState.Mstate = newMachineState
//	// worldState constraints
//	newWorldState := globalState.WorldState.Translate(ctx)
//	globalState.WorldState = newWorldState
//	// env
//	newEnv := globalState.Environment.Translate(ctx)
//	globalState.Environment = newEnv
//	// lastReturnData
//	//fmt.Println("changeStateContext done")
//}

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
