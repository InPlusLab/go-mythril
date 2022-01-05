package z3

import (
	"unsafe"
)

// #include <stdlib.h>
// #cgo CFLAGS: -IC:/Z3/src/api
// #cgo LDFLAGS: -LC:/Z3/build -llibz3
// #include "z3.h"
import "C"

// Add creates an AST node representing adding.
//
// All AST values must be part of the same context.
func (a *AST) Add(args ...*AST) *AST {
	raws := make([]C.Z3_ast, len(args)+1)
	raws[0] = a.rawAST
	for i, arg := range args {
		raws[i+1] = arg.rawAST
	}

	return &AST{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_add(
			a.rawCtx,
			C.uint(len(raws)),
			(*C.Z3_ast)(unsafe.Pointer(&raws[0]))),
	}
}

// Mul creates an AST node representing multiplication.
//
// All AST values must be part of the same context.
func (a *AST) Mul(args ...*AST) *AST {
	raws := make([]C.Z3_ast, len(args)+1)
	raws[0] = a.rawAST
	for i, arg := range args {
		raws[i+1] = arg.rawAST
	}

	return &AST{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_mul(
			a.rawCtx,
			C.uint(len(raws)),
			(*C.Z3_ast)(unsafe.Pointer(&raws[0]))),
	}
}

// Sub creates an AST node representing subtraction.
//
// All AST values must be part of the same context.
func (a *AST) Sub(args ...*AST) *AST {
	raws := make([]C.Z3_ast, len(args)+1)
	raws[0] = a.rawAST
	for i, arg := range args {
		raws[i+1] = arg.rawAST
	}

	return &AST{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_sub(
			a.rawCtx,
			C.uint(len(raws)),
			(*C.Z3_ast)(unsafe.Pointer(&raws[0]))),
	}
}

// Lt creates a "less than" comparison.
//
// Maps to: Z3_mk_lt
func (a *AST) Lt(a2 *AST) *AST {
	return &AST{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_lt(a.rawCtx, a.rawAST, a2.rawAST),
	}
}

// Le creates a "less than" comparison.
//
// Maps to: Z3_mk_le
func (a *AST) Le(a2 *AST) *AST {
	return &AST{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_le(a.rawCtx, a.rawAST, a2.rawAST),
	}
}

// Gt creates a "greater than" comparison.
//
// Maps to: Z3_mk_gt
func (a *AST) Gt(a2 *AST) *AST {
	return &AST{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_gt(a.rawCtx, a.rawAST, a2.rawAST),
	}
}

// Ge creates a "less than" comparison.
//
// Maps to: Z3_mk_ge
func (a *AST) Ge(a2 *AST) *AST {
	return &AST{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_ge(a.rawCtx, a.rawAST, a2.rawAST),
	}
}

// BvAdd creates an "addition" for bitvector
// created by chz
func (a *AST) BvAdd(t *AST) *AST {
	return &AST{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvadd(
			a.rawCtx,
			a.rawAST,
			t.rawAST)}
}

// BvSub creates a "subtraction" for bitvector
// created by chz
func (a *AST) BvSub(t *AST) *AST {
	return &AST{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvsub(
			a.rawCtx,
			a.rawAST,
			t.rawAST)}
}

// BvMul creates a "multiplication" for bitvector
// created by chz
func (a *AST) BvMul(t *AST) *AST {
	return &AST{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvmul(
			a.rawCtx,
			a.rawAST,
			t.rawAST)}
}

// BvSDiv creates a "signed division" for bitvector
// created by chz
func (a *AST) BvSDiv(t *AST) *AST {
	return &AST{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvsdiv(
			a.rawCtx,
			a.rawAST,
			t.rawAST)}
}

// BvUDiv creates an "unsigned division" for bitvector
// created by chz
func (a *AST) BvUDiv(t *AST) *AST {
	return &AST{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvudiv(
			a.rawCtx,
			a.rawAST,
			t.rawAST)}
}

// BvAddNoOverflow checks the addition of Node a & t doesn't overflow
// created by chz
func (a *AST) BvAddNoOverflow(t *AST, isSigned bool) *AST {
	return &AST{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvadd_no_overflow(
			a.rawCtx,
			a.rawAST,
			t.rawAST,
			C.bool(isSigned))}
}

// BvSubNoUnderflow checks the subtraction of Node a & t doesn't underflow
// created by chz
func (a *AST) BvSubNoUnderflow(t *AST, isSigned bool) *AST {
	return &AST{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvsub_no_underflow(
			a.rawCtx,
			a.rawAST,
			t.rawAST,
			C.bool(isSigned))}
}

// BvMulNoOverflow checks the multiplication of Node a & t doesn't overflow
// created by chz
func (a *AST) BvMulNoOverflow(t *AST, isSigned bool) *AST {
	return &AST{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvmul_no_overflow(
			a.rawCtx,
			a.rawAST,
			t.rawAST,
			C.bool(isSigned))}
}

// BvSLt creates a "signed <" for bitvector
// created by chz
func (a *AST) BvSLt(a2 *AST) *AST {
	return &AST{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvslt(a.rawCtx, a.rawAST, a2.rawAST),
	}
}

// BvSLe creates a "signed <=" for bitvector
// created by chz
func (a *AST) BvSLe(a2 *AST) *AST {
	return &AST{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvsle(a.rawCtx, a.rawAST, a2.rawAST),
	}
}

// BvSGt creates a "signed >" for bitvector
// created by chz
func (a *AST) BvSGt(a2 *AST) *AST {
	return &AST{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvsgt(a.rawCtx, a.rawAST, a2.rawAST),
	}
}

// BvSGe creates a "signed >=" for bitvector
// created by chz
func (a *AST) BvSGe(a2 *AST) *AST {
	return &AST{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvsge(a.rawCtx, a.rawAST, a2.rawAST),
	}
}

// BvULt creates an "unsigned <" for bitvector
// created by chz
func (a *AST) BvULt(a2 *AST) *AST {
	return &AST{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvult(a.rawCtx, a.rawAST, a2.rawAST),
	}
}

// BvULe creates an "unsigned <=" for bitvector
// created by chz
func (a *AST) BvULe(a2 *AST) *AST {
	return &AST{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvule(a.rawCtx, a.rawAST, a2.rawAST),
	}
}

// BvUGt creates an "unsigned >" for bitvector
// created by chz
func (a *AST) BvUGt(a2 *AST) *AST {
	return &AST{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvugt(a.rawCtx, a.rawAST, a2.rawAST),
	}
}

// BvUGe creates an "unsigned >=" for bitvector
// created by chz
func (a *AST) BvUGe(a2 *AST) *AST {
	return &AST{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvuge(a.rawCtx, a.rawAST, a2.rawAST),
	}
}

// BvURem gets the unsigned remainder for bitvector
// created by chz
func (a *AST) BvURem(a2 *AST) *AST {
	return &AST{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvurem(a.rawCtx, a.rawAST, a2.rawAST),
	}
}

// BvSRem gets the signed remainder for bitvector
// created by chz
func (a *AST) BvSRem(a2 *AST) *AST {
	return &AST{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvsrem(a.rawCtx, a.rawAST, a2.rawAST),
	}
}

// BvShL gets the shift left of node "a", "a2" is number of shift op.
// created by chz
func (a *AST) BvShL(a2 *AST) *AST {
	return &AST{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvshl(a.rawCtx, a.rawAST, a2.rawAST),
	}
}

// BvShR gets the arithmetical shift right of node "a", "a2" is number of shift op.
// created by chz
func (a *AST) BvShR(a2 *AST) *AST {
	return &AST{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvashr(a.rawCtx, a.rawAST, a2.rawAST),
	}
}

// BvLShR gets the logical shift right of node "a", "a2" is number of shift op.
// created by chz
func (a *AST) BvLShR(a2 *AST) *AST {
	return &AST{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvlshr(a.rawCtx, a.rawAST, a2.rawAST),
	}
}

// Int2Bv transforms the Int type to Bitvector type
// created by chz
func (a *AST) Int2Bv(bits uint) *AST {
	return &AST{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_int2bv(a.rawCtx, C.uint(bits), a.rawAST),
	}
}
