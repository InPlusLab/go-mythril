package z3

// #include <stdlib.h>
// #include "goZ3Config.h"
import "C"
import (
	"go-mythril/utils"
	"unsafe"
)

type Bool struct {
	rawCtx      C.Z3_context
	rawAST      C.Z3_ast
	symbolic    bool
	Annotations *utils.Set
}

func (c *Context) NewBool(ast *AST) *Bool {
	annotations := utils.NewSet()
	return &Bool{
		rawCtx:      c.raw,
		rawAST:      ast.rawAST,
		Annotations: annotations,
	}
}

// BoolString returns a human-friendly string version of the Bool.
func (b *Bool) BoolString() string {
	return C.GoString(C.Z3_ast_to_string(b.rawCtx, b.rawAST))
}

func (b *Bool) Copy() *Bool {
	return &Bool{
		rawCtx:      b.rawCtx,
		rawAST:      b.rawAST,
		symbolic:    b.symbolic,
		Annotations: b.Annotations,
	}
}

//func (b *Bool) Copy(ctx *Context) *Bool {
//	return &Bool{
//		rawCtx:      ctx.raw,
//		rawAST:      C.Z3_translate(b.rawCtx, b.rawAST, ctx.raw),
//		symbolic:    b.symbolic,
//		Annotations: b.Annotations.Copy(),
//	}
//}

// Translate is used to copy ast from one context to another.
func (b *Bool) Translate(c *Context) *Bool {
	// use in integer.go
	if b.rawCtx == c.raw {
		return b
	}
	//fmt.Println("BoolKind:", b.GetAstKind())
	return &Bool{
		rawCtx:      c.raw,
		rawAST:      C.Z3_translate(b.rawCtx, b.rawAST, c.raw),
		symbolic:    b.symbolic,
		Annotations: b.Annotations.Copy(),
	}
}

func (b *Bool) GetAstKind() C.Z3_ast_kind {
	return C.Z3_get_ast_kind(b.rawCtx, b.rawAST)
}
func (b *Bool) GetCtx() *Context {
	return &Context{
		raw: b.rawCtx,
	}
}

func (b *Bool) AsAST() *AST {
	ast := &AST{
		rawCtx:  b.rawCtx,
		rawAST:  b.rawAST,
		rawSort: C.Z3_mk_bool_sort(b.rawCtx),
	}
	return ast.Simplify()
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
		rawCtx:      b.rawCtx,
		rawAST:      C.Z3_mk_not(b.rawCtx, b.rawAST),
		symbolic:    b.symbolic,
		Annotations: b.Annotations,
	}
}

func (b *Bool) Simplify() *Bool {
	return &Bool{
		rawCtx:      b.rawCtx,
		rawAST:      C.Z3_simplify(b.rawCtx, b.rawAST),
		symbolic:    b.symbolic,
		Annotations: b.Annotations,
	}
}

func (b *Bool) ToString() string {
	return C.GoString(C.Z3_ast_to_string(b.rawCtx, b.rawAST))
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
	symbolicSym := a.symbolic
	annotations := a.Annotations
	for i, arg := range args {
		raws[i+1] = arg.rawAST
		annotations.Union(arg.Annotations)
		symbolicSym = symbolicSym || arg.symbolic
	}

	return &Bool{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_and(
			a.rawCtx,
			C.uint(len(raws)),
			(*C.Z3_ast)(unsafe.Pointer(&raws[0]))),
		symbolic:    symbolicSym,
		Annotations: annotations,
	}
}

func (a *Bool) Or(args ...*Bool) *Bool {
	raws := make([]C.Z3_ast, len(args)+1)
	raws[0] = a.rawAST
	symbolicSym := a.symbolic
	annotations := a.Annotations
	for i, arg := range args {
		raws[i+1] = arg.rawAST
		annotations.Union(arg.Annotations)
		symbolicSym = symbolicSym || arg.symbolic
	}
	return &Bool{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_or(
			a.rawCtx,
			C.uint(len(raws)),
			(*C.Z3_ast)(unsafe.Pointer(&raws[0]))),
		symbolic:    symbolicSym,
		Annotations: annotations,
	}
}

// For debug
func GetBoolCtx(b *Bool) *Context {
	return &Context{
		raw: b.rawCtx,
	}
}
