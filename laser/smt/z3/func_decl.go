package z3

// #include <stdlib.h>
// #include "goZ3Config.h"
import "C"
import "go-mythril/utils"

type FuncDecl struct {
	rawCtx      C.Z3_context
	rawFuncDecl C.Z3_func_decl
}

func (c *Context) NewFuncDecl(name string, domain []*Sort, range_ *Sort) *FuncDecl {
	sym := c.Symbol(name)
	cdomain := make([]C.Z3_sort, len(domain))
	for i, sort := range domain {
		cdomain[i] = sort.rawSort
	}
	var cdp *C.Z3_sort
	if len(cdomain) > 0 {
		cdp = &cdomain[0]
	}
	return &FuncDecl{
		rawCtx:      c.raw,
		rawFuncDecl: C.Z3_mk_func_decl(c.raw, sym.rawSymbol, C.uint(len(cdomain)), cdp, range_.rawSort),
	}
}

// Apply creates the application of the given funcdecl and args
func (f *FuncDecl) Apply(args ...*AST) *AST {
	cargs := make([]C.Z3_ast, len(args))
	for i, arg := range args {
		cargs[i] = arg.rawAST
	}
	var cap *C.Z3_ast
	if len(cargs) > 0 {
		cap = &cargs[0]
	}
	return &AST{
		rawCtx: f.rawCtx,
		rawAST: C.Z3_mk_app(f.rawCtx, f.rawFuncDecl, C.uint(len(cargs)), cap),
	}
}

func (f *FuncDecl) ApplyBv(args ...*Bitvec) *Bitvec {
	cargs := make([]C.Z3_ast, len(args))
	symbolicValue := false
	annotations := utils.NewSet()
	for i, arg := range args {
		cargs[i] = arg.rawAST
		annotations.Union(arg.Annotations)
		if arg.Symbolic() {
			symbolicValue = true
		}
	}
	var cap *C.Z3_ast
	if len(cargs) > 0 {
		cap = &cargs[0]
	}

	return &Bitvec{
		rawCtx:      f.rawCtx,
		rawAST:      C.Z3_mk_app(f.rawCtx, f.rawFuncDecl, C.uint(len(cargs)), cap),
		rawSort:     args[0].rawSort,
		symbolic:    symbolicValue,
		Annotations: annotations,
	}
}
