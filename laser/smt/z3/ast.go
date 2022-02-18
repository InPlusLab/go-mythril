package z3

// #include <stdlib.h>
// #include "goZ3Config.h"
import "C"

// AST represents an AST value in Z3.
//
// AST memory management is automatically managed by the Context it
// is contained within. When the Context is freed, so are the AST nodes.
type AST struct {
	rawCtx C.Z3_context
	rawAST C.Z3_ast
	// created by chz
	rawSort C.Z3_sort
}

// String returns a human-friendly string version of the AST.
func (a *AST) String() string {
	return C.GoString(C.Z3_ast_to_string(a.rawCtx, a.rawAST))
}

// DeclName returns the name of a declaration. The AST value must be a
// func declaration for this to work.
func (a *AST) DeclName() *Symbol {
	return &Symbol{
		rawCtx: a.rawCtx,
		rawSymbol: C.Z3_get_decl_name(
			a.rawCtx, C.Z3_to_func_decl(a.rawCtx, a.rawAST)),
	}
}

//-------------------------------------------------------------------
// Var, Literal Creation
//-------------------------------------------------------------------

// Const declares a variable. It is called "Const" since internally
// this is equivalent to create a function that always returns a constant
// value. From an initial user perspective this may be confusing but go-z3
// is following identical naming convention.
func (c *Context) Const(s *Symbol, typ *Sort) *AST {
	return &AST{
		rawCtx:  c.raw,
		rawAST:  C.Z3_mk_const(c.raw, s.rawSymbol, typ.rawSort),
		rawSort: typ.rawSort,
	}
}

// Int creates an integer type. Sort can be int, bv or finite-domain sort.
//	created by chz
// Maps: Z3_mk_int, which is faster than Z3_mk_numeral
func (c *Context) Int(v int, typ *Sort) *AST {
	return &AST{
		rawCtx: c.raw,
		rawAST: C.Z3_mk_int(c.raw, C.int(v), typ.rawSort),
	}
}

// True creates the value "true".
//
// Maps: Z3_mk_true
func (c *Context) True() *AST {
	return &AST{
		rawCtx: c.raw,
		rawAST: C.Z3_mk_true(c.raw),
	}
}

// False creates the value "false".
//
// Maps: Z3_mk_false
func (c *Context) False() *AST {
	return &AST{
		rawCtx: c.raw,
		rawAST: C.Z3_mk_false(c.raw),
	}
}

//-------------------------------------------------------------------
// Value Readers
//-------------------------------------------------------------------

// Int gets the integer value of this AST. The value must be able to fit
// into a machine integer.
func (a *AST) Int() int {
	var dst C.int
	C.Z3_get_numeral_int(a.rawCtx, a.rawAST, &dst)
	return int(dst)
}

// Simplify
// created by chz
func (a *AST) Simplify() *AST {
	return &AST{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_simplify(a.rawCtx, a.rawAST),
	}
}

// IsAppOf is used to determine the AST's type, such as true, false, add, etc.
// Part of the Z3_decl_kind constants in enum.go
// created by chz
func (a *AST) IsAppOf(k C.Z3_decl_kind) bool {
	var res bool
	res = bool(C.Z3_is_app(a.rawCtx, a.rawAST)) &&
		C.Z3_get_decl_kind(a.rawCtx, C.Z3_get_app_decl(a.rawCtx, C.Z3_to_app(a.rawCtx, a.rawAST))) == k

	return res
}

// Translate is used to copy ast from one context to another.
func (a *AST) Translate(c *Context) *AST {
	return &AST{
		rawCtx: c.raw,
		rawAST: C.Z3_translate(a.rawCtx, a.rawAST, c.raw),
	}
}
