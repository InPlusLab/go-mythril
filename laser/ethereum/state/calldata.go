package state

import "C"
import (
	"go-mythril/laser/smt/z3"
	"strconv"
)

type BaseCalldata interface {
	Calldatasize() *z3.Bitvec
	GetWordAt(offset *z3.Bitvec) *z3.Bitvec
	Load(item *z3.Bitvec) *z3.Bitvec
	Size() *z3.Bitvec
	Concrete(model *z3.Model) []*z3.Bitvec
	Translate(ctx *z3.Context) BaseCalldata
}

type ConcreteCalldata struct {
	TxId             string
	ConcreteCalldata []*z3.Bitvec
	Calldata         *z3.K
}

// BasicConcreteCalldata class
// Initializes the ConcreteCalldata object, that doesn't use z3 arrays.
type BasicConcreteCalldata struct {
	TxId     string
	Calldata []*z3.Bitvec
}
type SymbolicCalldata struct {
	TxId     string
	Calldata *z3.Array
	SymSize  *z3.Bitvec
	Ctx      *z3.Context
}
type BasicSymbolicCalldata struct {
	TxId    string
	Reads   *map[int64]*z3.Bitvec
	SymSize *z3.Bitvec
	Ctx     *z3.Context
}

func NewConcreteCalldata(id string, calldata []*z3.Bitvec, ctx *z3.Context) *ConcreteCalldata {
	k := ctx.NewK(256, 8, 0)
	for i := 0; i < len(calldata); i++ {
		k = k.SetItem(ctx.NewBitvecVal(i, 256), calldata[i]).(*z3.K)
	}
	return &ConcreteCalldata{
		TxId:             id,
		ConcreteCalldata: calldata,
		Calldata:         k,
	}
}
func NewBasicConcreteCalldata(id string, calldata []*z3.Bitvec) *BasicConcreteCalldata {
	return &BasicConcreteCalldata{
		TxId:     id,
		Calldata: calldata,
	}
}
func NewSymbolicCalldata(id string, ctx *z3.Context) *SymbolicCalldata {
	return &SymbolicCalldata{
		TxId:     id,
		Calldata: ctx.NewArray(id+"_calldata", 256, 8),
		SymSize:  ctx.NewBitvec(id+"_calldatasize", 256),
		Ctx:      ctx,
	}
}
func NewBasicymbolicCalldata(id string, ctx *z3.Context) *BasicSymbolicCalldata {
	r := make(map[int64]*z3.Bitvec)
	return &BasicSymbolicCalldata{
		TxId:    id,
		Reads:   &r,
		SymSize: ctx.NewBitvec(id+"_calldatasize", 256),
		Ctx:     ctx,
	}
}

func (ccd *ConcreteCalldata) Calldatasize() *z3.Bitvec {
	return ccd.Size()
}
func (ccd *ConcreteCalldata) GetWordAt(offset *z3.Bitvec) *z3.Bitvec {
	tmp := ccd.Calldata.GetItem(offset)
	ctx := tmp.GetCtx()
	// OutofIndex check
	index, _ := strconv.ParseInt(offset.Value(), 10, 64)
	for i := index + 1; i < index+32; i++ {
		tmp = tmp.Concat(ccd.Calldata.GetItem(ctx.NewBitvecVal(i, 256)))
	}
	return tmp.Simplify()
}
func (ccd *ConcreteCalldata) Load(item *z3.Bitvec) *z3.Bitvec {
	return ccd.Calldata.GetItem(item).Simplify()
}
func (ccd *ConcreteCalldata) Concrete(model *z3.Model) []*z3.Bitvec {
	return ccd.ConcreteCalldata
}
func (ccd *ConcreteCalldata) Size() *z3.Bitvec {
	item := ccd.ConcreteCalldata[0]
	ctx := item.GetCtx()
	return ctx.NewBitvecVal(len(ccd.ConcreteCalldata), 256)
	//return ctx.NewBitvec(ccd.TxId+"_calldatasize", 256)
}
func (ccd *ConcreteCalldata) Translate(ctx *z3.Context) BaseCalldata {
	newCalldata := make([]*z3.Bitvec, 0)
	for _, v := range ccd.ConcreteCalldata {
		newV := v.Translate(ctx)
		newCalldata = append(newCalldata, newV)
	}
	return &ConcreteCalldata{
		TxId:             ccd.TxId,
		ConcreteCalldata: newCalldata,
		Calldata:         ccd.Calldata.Translate(ctx).(*z3.K),
	}
}

func (bcd *BasicConcreteCalldata) Calldatasize() *z3.Bitvec {
	return bcd.Size()
}
func (bcd *BasicConcreteCalldata) GetWordAt(offset *z3.Bitvec) *z3.Bitvec {
	tmp := bcd.Load(offset)
	// OutofIndex check
	index, _ := strconv.ParseInt(offset.Value(), 10, 64)
	for i := index + 1; i < index+32; i++ {
		tmp = tmp.Concat(bcd.Calldata[int(index)])
	}
	return tmp.Simplify()
}
func (bcd *BasicConcreteCalldata) Load(item *z3.Bitvec) *z3.Bitvec {
	index, _ := strconv.ParseInt(item.Value(), 10, 64)
	return bcd.Calldata[int(index)]
}
func (bcd *BasicConcreteCalldata) Concrete(model *z3.Model) []*z3.Bitvec {
	return bcd.Calldata
}
func (bcd *BasicConcreteCalldata) Size() *z3.Bitvec {
	item := bcd.Calldata[0]
	ctx := item.GetCtx()
	return ctx.NewBitvecVal(len(bcd.Calldata), 256)
}

func (scd *SymbolicCalldata) Calldatasize() *z3.Bitvec {
	return scd.Size()
}
func (scd *SymbolicCalldata) GetWordAt(offset *z3.Bitvec) *z3.Bitvec {
	tmp := scd.Load(offset)
	// OutofIndex check
	index, _ := strconv.ParseInt(offset.Value(), 10, 64)
	for i := index + 1; i < index+32; i++ {
		tmp = tmp.Concat(scd.Load(scd.Ctx.NewBitvecVal(i, 256)))
	}
	return tmp.Simplify()
	//return scd.Ctx.NewBitvec("SymbolicInput-"+offset.BvString(), 256)
}
func (scd *SymbolicCalldata) Load(item *z3.Bitvec) *z3.Bitvec {
	//return z3.If(item.BvSLt(scd.SymSize),
	//	scd.Calldata.GetItem(item).Simplify(),
	//	scd.Ctx.NewBitvecVal(0, 8)).Simplify()
	//return scd.Calldata.GetItem(item).Simplify()
	return z3.If(item.BvSLt(scd.Size()), scd.Calldata.GetItem(item), scd.Ctx.NewBitvecVal(0, 8)).Simplify()
}

// TODO: z3.model should be changed. In Mythril, model.py is a list of models.
func (scd *SymbolicCalldata) Concrete(model *z3.Model) []*z3.Bitvec {
	result := make([]*z3.Bitvec, 0)
	return result
}
func (scd *SymbolicCalldata) Size() *z3.Bitvec {
	return scd.SymSize
}
func (scd *SymbolicCalldata) Translate(ctx *z3.Context) BaseCalldata {
	return &SymbolicCalldata{
		TxId:     scd.TxId,
		Calldata: scd.Calldata.Translate(ctx).(*z3.Array),
		SymSize:  scd.SymSize.Translate(ctx),
		Ctx:      ctx,
	}
}

func (bsd *BasicSymbolicCalldata) Calldatasize() *z3.Bitvec {
	return bsd.Size()
}
func (bsd *BasicSymbolicCalldata) GetWordAt(offset *z3.Bitvec) *z3.Bitvec {
	tmp := bsd.Load(bsd.Ctx.NewBitvecVal(offset, 256))
	// OutofIndex check
	index, _ := strconv.ParseInt(offset.Value(), 10, 64)
	for i := index + 1; i < index+32; i++ {
		tmp = tmp.Concat(bsd.Load(bsd.Ctx.NewBitvecVal(i, 256)))
	}
	return tmp.Simplify()
}
func (bsd *BasicSymbolicCalldata) Load(item *z3.Bitvec) *z3.Bitvec {
	symbolicBaseValue := z3.If(item.BvSGe(bsd.SymSize),
		bsd.Ctx.NewBitvecVal(0, 8),
		bsd.Ctx.NewBitvec(bsd.TxId+"_calldata_"+item.BvString(), 8))
	returnValue := symbolicBaseValue
	reads := *bsd.Reads

	for i, v := range reads {
		iBv := bsd.Ctx.NewBitvecVal(i, 256)
		returnValue = z3.If(iBv.Eq(item), v, returnValue)
	}
	// default clean==false
	itemIndex, _ := strconv.ParseInt(item.Value(), 10, 64)
	reads[itemIndex] = symbolicBaseValue
	return returnValue.Simplify()
}

// TODO: z3.model should be changed. In Mythril, model.py is a list of models.
func (bsd *BasicSymbolicCalldata) Concrete(model *z3.Model) []*z3.Bitvec {
	result := make([]*z3.Bitvec, 0)
	return result
}
func (bsd *BasicSymbolicCalldata) Size() *z3.Bitvec {
	return bsd.SymSize
}
