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
	RLimitCount    int
	NeedIsPossible bool
	SkipTimes      int
	// cloud9
	ForkId string
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
		RLimitCount:    0,
		NeedIsPossible: false,
		SkipTimes:      0,
		ForkId: "?",
	}
}

func (globalState *GlobalState) Copy() *GlobalState {

	newAnnotations := make([]StateAnnotation, 0)
	for _, anno := range globalState.Annotations {
		//if reflect.TypeOf(anno).String() == "*ethereum.JumpdestCountAnnotation" {
		//	newAnnotations = append(newAnnotations, anno.Copy())
		//	continue
		//}
		newAnnotations = append(newAnnotations, anno.Copy())
	}
	newWs := globalState.WorldState.Copy()
	newEnv := globalState.Environment.Copy()
	newEnv.ActiveAccount = newWs.AccountsExistOrLoad(newEnv.ActiveAccount.Address)

	newMs := globalState.Mstate.DeepCopy()

	return &GlobalState{
		WorldState:     newWs,
		Environment:    newEnv,
		Mstate:         newMs,
		TxStack:        globalState.TxStack,
		Z3ctx:          globalState.Z3ctx,
		LastReturnData: globalState.LastReturnData,
		Annotations:    newAnnotations,
		RLimitCount:    globalState.RLimitCount,
		NeedIsPossible: globalState.NeedIsPossible,
		SkipTimes:      globalState.SkipTimes,
		ForkId: globalState.ForkId,
	}
}

func (globalState *GlobalState) Translate(ctx *z3.Context) {
	if globalState.Z3ctx.GetRaw() == ctx.GetRaw() {
		return
	}
	// fmt.Println("before changeStateContext:", globalState.GetCurrentInstruction().OpCode.Name, globalState, globalState.LastReturnData)

	globalState.Z3ctx = ctx
	// machineState stack & memory
	// fmt.Println("glTrans1")
	globalState.Mstate = globalState.Mstate.Translate(ctx)
	// worldState constraints
	// fmt.Println("glTrans2")
	globalState.WorldState = globalState.WorldState.Translate(ctx)
	// env
	// fmt.Println("glTrans3")
	globalState.Environment = globalState.Environment.Translate(ctx)
	// annotations
	// fmt.Println("glTrans4")
	newAnnotations := make([]StateAnnotation, 0)
	for _, anno := range globalState.Annotations {
		newAnnotations = append(newAnnotations, anno.Translate(ctx))
	}
	globalState.Annotations = newAnnotations

	// lastReturnData
	//fmt.Println("changeStateContext done")
}

func (globalState *GlobalState) TranslateR(ctx *z3.Context) *GlobalState {
	newAnnotations := make([]StateAnnotation, 0)
	for _, anno := range globalState.Annotations {
		newAnnotations = append(newAnnotations, anno.Translate(ctx))
	}
	newWs := globalState.WorldState.Translate(ctx)
	newEnv := globalState.Environment.Translate(ctx)
	newEnv.ActiveAccount = newWs.AccountsExistOrLoad(newEnv.ActiveAccount.Address)
	newMs := globalState.Mstate.Translate(ctx)

	return &GlobalState{
		WorldState:     newWs,
		Environment:    newEnv,
		Mstate:         newMs,
		TxStack:        globalState.TxStack,
		Z3ctx:          ctx,
		LastReturnData: globalState.LastReturnData,
		Annotations:    newAnnotations,
		RLimitCount: globalState.RLimitCount,
		NeedIsPossible: globalState.NeedIsPossible,
		SkipTimes: globalState.SkipTimes,
		ForkId: globalState.ForkId,
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
	if globalState.Annotations != nil {
		for _, v := range globalState.Annotations {
			if annoType == reflect.TypeOf(v) {
				annoList = append(annoList, v)
			}
		}
	}

	return annoList
}
