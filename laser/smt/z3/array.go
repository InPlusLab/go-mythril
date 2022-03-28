package z3

// #include <stdlib.h>
// #include "goZ3Config.h"
import "C"

type BaseArray interface {
	GetItem(bitvec *Bitvec) *Bitvec
	SetItem(index *Bitvec, value *Bitvec) BaseArray
	DeepCopy() BaseArray
}
type Array struct {
	name   string
	rawCtx C.Z3_context
	rawAST C.Z3_ast
}
type K struct {
	rawCtx C.Z3_context
	rawAST C.Z3_ast
}

func (c *Context) NewArray(n string, dom uint, vRange uint) *Array {
	// In mythril, domain and range of array are bv type.
	ast := C.Z3_mk_const(c.raw, c.Symbol(n).rawSymbol, c.ArraySort(c.BvSort(dom), c.BvSort(vRange)).rawSort)
	return &Array{
		name:   n,
		rawAST: ast,
		rawCtx: c.raw,
	}
}

// K is an array whose values are all a constant value.
func (c *Context) NewK(dom uint, vRange uint, value int) *K {
	// In mythril, domain and range of array are bv type.
	ast := C.Z3_mk_const_array(c.raw, c.BvSort(dom).rawSort, c.Int(value, c.BvSort(vRange)).rawAST)
	return &K{
		rawAST: ast,
		rawCtx: c.raw,
	}
}

func (a *Array) SetItem(index *Bitvec, value *Bitvec) BaseArray {
	return &Array{
		name:   a.name,
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_store(a.rawCtx, a.rawAST, index.rawAST, value.rawAST),
	}
}

func (a *Array) GetItem(index *Bitvec) *Bitvec {
	return &Bitvec{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_select(a.rawCtx, a.rawAST, index.rawAST),
	}
}
func (a *Array) GetCtx() *Context {
	return &Context{
		raw: a.rawCtx,
	}
}
func (a *Array) DeepCopy() BaseArray {
	return &Array{
		name:   a.name,
		rawCtx: a.rawCtx,
		rawAST: a.rawAST,
	}
}

func (a *K) SetItem(index *Bitvec, value *Bitvec) BaseArray {
	return &K{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_store(a.rawCtx, a.rawAST, index.rawAST, value.rawAST),
	}
}

func (a *K) GetItem(index *Bitvec) *Bitvec {
	return &Bitvec{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_select(a.rawCtx, a.rawAST, index.rawAST),
	}
}

func (a *K) GetCtx() *Context {
	return &Context{
		raw: a.rawCtx,
	}
}

func (a *K) DeepCopy() BaseArray {
	return &K{
		rawCtx: a.rawCtx,
		rawAST: a.rawAST,
	}
}
