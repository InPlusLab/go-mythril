package z3

// #include <stdlib.h>
// #include "goZ3Config.h"
import "C"
import (
	"go-mythril/utils"
)

type BaseArray interface {
	GetItem(bitvec *Bitvec) *Bitvec
	SetItem(index *Bitvec, value *Bitvec) BaseArray
	DeepCopy() BaseArray
	Translate(ctx *Context) BaseArray
}
type Array struct {
	name      string
	rawCtx    C.Z3_context
	rawAST    C.Z3_ast
	domSize   int
	rangeSize int
}
type K struct {
	rawCtx    C.Z3_context
	rawAST    C.Z3_ast
	domSize   int
	rangeSize int
}

func (c *Context) NewArray(n string, dom int, vRange int) *Array {
	// In mythril, domain and range of array are bv type.
	//ast := C.Z3_mk_const(c.raw, c.Symbol(n).rawSymbol, c.ArraySort(c.BvSort(uint(dom)), c.BvSort(uint(vRange))).rawSort)
	ast := C.Z3_mk_const(c.raw, c.Symbol(n).rawSymbol, c.ArraySort(c.BvSort(uint(dom)), c.BvSort(uint(vRange))).rawSort)
	return &Array{
		name:      n,
		rawAST:    ast,
		rawCtx:    c.raw,
		domSize:   dom,
		rangeSize: vRange,
	}
}

// K is an array whose values are all constant.
func (c *Context) NewK(dom int, vRange int, value int) *K {
	// In mythril, domain and range of array are bv type.
	// ast := C.Z3_mk_const_array(c.raw, c.BvSort(dom).rawSort, c.Int(value, c.BvSort(vRange)).rawAST)
	ast := C.Z3_mk_const_array(c.raw, c.BvSort(uint(dom)).rawSort, c.NewBitvecVal(value, vRange).rawAST)
	return &K{
		rawAST:    ast,
		rawCtx:    c.raw,
		domSize:   dom,
		rangeSize: vRange,
	}
}

func (a *Array) SetItem(index *Bitvec, value *Bitvec) BaseArray {
	return &Array{
		name:      a.name,
		rawCtx:    a.rawCtx,
		rawAST:    C.Z3_mk_store(a.rawCtx, a.rawAST, index.rawAST, value.rawAST),
		domSize:   a.domSize,
		rangeSize: a.rangeSize,
	}
}

func (a *Array) GetItem(index *Bitvec) *Bitvec {
	return &Bitvec{
		rawCtx:  a.rawCtx,
		rawAST:  C.Z3_mk_select(a.rawCtx, a.rawAST, index.rawAST),
		rawSort: C.Z3_mk_bv_sort(a.rawCtx, C.uint(a.rangeSize)),
		// TODO: maybe wrong
		symbolic:    index.symbolic,
		Annotations: utils.NewSet(),
	}
}
func (a *Array) GetCtx() *Context {
	return &Context{
		raw: a.rawCtx,
	}
}
func (a *Array) DeepCopy() BaseArray {
	return &Array{
		name:      a.name,
		rawCtx:    a.rawCtx,
		rawAST:    a.rawAST,
		domSize:   a.domSize,
		rangeSize: a.rangeSize,
	}
}
func (a *Array) Translate(ctx *Context) BaseArray {
	return &Array{
		name:      a.name,
		rawCtx:    ctx.raw,
		rawAST:    C.Z3_translate(a.rawCtx, a.rawAST, ctx.raw),
		domSize:   a.domSize,
		rangeSize: a.rangeSize,
	}
}

func (a *K) SetItem(index *Bitvec, value *Bitvec) BaseArray {
	return &K{
		rawCtx:    a.rawCtx,
		rawAST:    C.Z3_mk_store(a.rawCtx, a.rawAST, index.rawAST, value.rawAST),
		domSize:   a.domSize,
		rangeSize: a.rangeSize,
	}
}

func (a *K) GetItem(index *Bitvec) *Bitvec {
	return &Bitvec{
		rawCtx:  a.rawCtx,
		rawAST:  C.Z3_mk_select(a.rawCtx, a.rawAST, index.rawAST),
		rawSort: C.Z3_mk_bv_sort(a.rawCtx, C.uint(a.rangeSize)),
		//rawSort: nil,
		// TODO: maybe wrong
		symbolic:    index.symbolic,
		Annotations: utils.NewSet(),
	}
}

func (a *K) GetCtx() *Context {
	return &Context{
		raw: a.rawCtx,
	}
}

func (a *K) DeepCopy() BaseArray {
	return &K{
		rawCtx:    a.rawCtx,
		rawAST:    a.rawAST,
		domSize:   a.domSize,
		rangeSize: a.rangeSize,
	}
}

func (a *K) Translate(ctx *Context) BaseArray {
	return &K{
		rawCtx:    ctx.raw,
		rawAST:    C.Z3_translate(a.rawCtx, a.rawAST, ctx.raw),
		domSize:   a.domSize,
		rangeSize: a.rangeSize,
	}
}
