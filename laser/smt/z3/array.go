package z3

// #include <stdlib.h>
// #include "goZ3Config.h"
import "C"

type BaseArray struct {
	name   string
	rawCtx C.Z3_context
	rawAST C.Z3_ast
}
type Array struct {
	*BaseArray
}
type K struct {
	*BaseArray
}

func NewBaseArray(n string, ctx *Context, ast *AST) *BaseArray {
	return &BaseArray{
		name:   n,
		rawCtx: ctx.raw,
		rawAST: ast.rawAST,
	}
}

func (c *Context) NewArray(n string, dom uint, vRange uint) *Array {
	// In mythril, domain and range of array are bv type.
	ast := C.Z3_mk_const(c.raw, c.Symbol(n).rawSymbol, c.ArraySort(c.BvSort(dom), c.BvSort(vRange)).rawSort)
	newAST := &AST{
		rawAST: ast,
		rawCtx: c.raw,
	}
	return &Array{
		BaseArray: NewBaseArray(n, c, newAST),
	}
}

// K is an array whose values are all a constant value.
func (c *Context) NewK(dom uint, vRange uint, value int) *K {
	// In mythril, domain and range of array are bv type.
	ast := C.Z3_mk_const_array(c.raw, c.BvSort(dom).rawSort, c.Int(value, c.BvSort(vRange)).rawAST)
	newAST := &AST{
		rawAST: ast,
		rawCtx: c.raw,
	}
	return &K{
		// K's name is "";
		BaseArray: NewBaseArray("", c, newAST),
	}
}

func (a *BaseArray) SetItem(index *Bitvec, value *Bitvec) *Array {
	newBA := &BaseArray{
		name:   a.name,
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_store(a.rawCtx, a.rawAST, index.rawAST, value.rawAST),
	}
	return &Array{
		BaseArray: newBA,
	}
}

func (a *BaseArray) GetItem(index *Bitvec) *Bitvec {
	return &Bitvec{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_select(a.rawCtx, a.rawAST, index.rawAST),
	}
}
