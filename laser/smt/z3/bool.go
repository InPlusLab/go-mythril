package z3

// #include <stdlib.h>
// #include "goZ3Config.h"
import "C"
import "unsafe"

type Bool struct {
	rawCtx C.Z3_context
	rawAST C.Z3_ast
}

func (c *Context) NewBool(ast *AST) *Bool {
	return &Bool{
		rawCtx: c.raw,
		rawAST: ast.rawAST,
	}
}

func (b *Bool) AsAST() *AST {
	return &AST{
		rawCtx: b.rawCtx,
		rawAST: b.rawAST,
	}
}

func (b *Bool) IsTrue() bool {
	ast := b.AsAST().Simplify()
	return ast.IsAppOf(OpTrue)
}

func (b *Bool) IsFalse() bool {
	ast := b.AsAST().Simplify()
	return ast.IsAppOf(OpFalse)
}

func (b *Bool) Not() *Bool {
	return &Bool{
		rawCtx: b.rawCtx,
		rawAST: C.Z3_mk_not(b.rawCtx, b.rawAST),
	}
}

func (b *Bool) Simplify() *Bool {
	return &Bool{
		rawCtx: b.rawCtx,
		rawAST: C.Z3_simplify(b.rawCtx, b.rawAST),
	}
}

// Not tested !
func (b *AST) Substitute(args ...*AST) *AST {
	froms := make([]C.Z3_ast, len(args)/2)
	tos := make([]C.Z3_ast, len(args)/2)
	j := 0
	for i, arg := range args {
		if i%2 == 0 {
			froms[j] = arg.rawAST
		} else {
			tos[j] = arg.rawAST
			j++
		}
	}
	return &AST{
		rawCtx: b.rawCtx,
		rawAST: C.Z3_substitute(b.rawCtx,
			b.rawAST,
			C.uint(len(args)/2),
			(*C.Z3_ast)(unsafe.Pointer(&froms[0])),
			(*C.Z3_ast)(unsafe.Pointer(&tos[0]))),
	}
}

func (a *Bool) And(args ...*Bool) *Bool {
	raws := make([]C.Z3_ast, len(args)+1)
	raws[0] = a.rawAST
	for i, arg := range args {
		raws[i+1] = arg.rawAST
	}
	return &Bool{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_and(
			a.rawCtx,
			C.uint(len(raws)),
			(*C.Z3_ast)(unsafe.Pointer(&raws[0]))),
	}
}

func (a *Bool) Or(args ...*Bool) *Bool {
	raws := make([]C.Z3_ast, len(args)+1)
	raws[0] = a.rawAST
	for i, arg := range args {
		raws[i+1] = arg.rawAST
	}
	return &Bool{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_or(
			a.rawCtx,
			C.uint(len(raws)),
			(*C.Z3_ast)(unsafe.Pointer(&raws[0]))),
	}
}

// For debug
func GetBoolCtx(b *Bool) *Context {
	return &Context{
		raw: b.rawCtx,
	}
}
