package z3

// #include <stdlib.h>
// #include "goZ3Config.h"
import "C"

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
