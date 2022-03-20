package z3

// #include <stdlib.h>
// #include "goZ3Config.h"
import "C"
import (
	"fmt"
	"math/big"
	"strconv"
)

type Bitvec struct {
	rawCtx   C.Z3_context
	rawAST   C.Z3_ast
	rawSort  C.Z3_sort
	symbolic bool
}

// NewBitvec corresponds to BitVec() in Python
// It implements the class BitVecRef
func (c *Context) NewBitvec(bvSymbol string, size int) *Bitvec {
	ast := C.Z3_mk_const(c.raw, c.Symbol(bvSymbol).rawSymbol, c.BvSort(uint(size)).rawSort)
	return &Bitvec{
		rawCtx:   c.raw,
		rawAST:   ast,
		rawSort:  c.BvSort(uint(size)).rawSort,
		symbolic: true,
	}
}

// NewBitvecVal corresponds to BitVecVal() in Python
// It implements the class BitVecNumRef
// type int supports the normal usage, big.Int supports the big number usage.
func (c *Context) NewBitvecVal(value interface{}, size int) *Bitvec {
	var valueStr string

	switch t := value.(type) {
	case int:
		valueStr = strconv.FormatInt(int64(value.(int)), 10)
	case int64:
		valueStr = strconv.FormatInt(value.(int64), 10)
	case *big.Int:
		b := value.(*big.Int)
		valueStr = b.String()
	default:
		fmt.Println(t)
		panic("type error for NewBitvecVal")
	}
	// Can't use mk_int API here.
	ast := C.Z3_mk_numeral(c.raw, C.CString(valueStr), c.BvSort(uint(size)).rawSort)
	return &Bitvec{
		rawCtx:   c.raw,
		rawAST:   ast,
		rawSort:  c.BvSort(uint(size)).rawSort,
		symbolic: false,
	}
}

// The attribute of bitvec

/*func (a *Bitvec) GetCtx() *Context {
	return a.rawCtx
}*/

// BvSize returns the size of bitvector
// created by chz
func (a *Bitvec) BvSize() int {
	return int(C.Z3_get_bv_sort_size(a.rawCtx, a.rawSort))
}

// AsAST returns the ast of bv
func (b *Bitvec) AsAST() *AST {
	return &AST{
		rawCtx: b.rawCtx,
		rawAST: b.rawAST,
	}
}

// Value returns the value of BitvecVal, "" for Bitvec
// should use get_numeral_string rather than get_numeral_int API
func (b *Bitvec) Value() string {
	if !b.symbolic {
		value := C.GoString(C.Z3_get_numeral_string(b.rawCtx, b.rawAST))
		return value
	} else {
		return ""
	}
}

// Symbolic tells whether this bv is symbolic
func (b *Bitvec) Symbolic() bool {
	return b.symbolic
}

// String returns a human-friendly string version of the AST.
// eg: String(BitVec("x",256)) == "x"
func (b *Bitvec) String() string {
	return C.GoString(C.Z3_ast_to_string(b.rawCtx, b.rawAST))
}

// GetAstKind returns the ast kind of bv.
// For debug.
func (b *Bitvec) GetAstKind() C.Z3_ast_kind {
	return C.Z3_get_ast_kind(b.rawCtx, b.rawAST)
}

// GetCtx returns the context of bitvec b.
func (b *Bitvec) GetCtx() *Context {
	return &Context{
		raw: b.rawCtx,
	}
}

// Simplify equations
func (b *Bitvec) Simplify() *Bitvec {
	return &Bitvec{
		rawCtx: b.rawCtx,
		rawAST: C.Z3_simplify(b.rawCtx, b.rawAST),
	}
}

// The arith op in Bitvec

// BvAdd creates an "addition" for bitvector
// created by chz
func (a *Bitvec) BvAdd(t *Bitvec) *Bitvec {
	return &Bitvec{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvadd(
			a.rawCtx,
			a.rawAST,
			t.rawAST),
		rawSort: a.rawSort,
	}
}

// BvSub creates a "subtraction" for bitvector
// created by chz
func (a *Bitvec) BvSub(t *Bitvec) *Bitvec {
	return &Bitvec{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvsub(
			a.rawCtx,
			a.rawAST,
			t.rawAST)}
}

// BvMul creates a "multiplication" for bitvector
// created by chz
func (a *Bitvec) BvMul(t *Bitvec) *Bitvec {
	return &Bitvec{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvmul(
			a.rawCtx,
			a.rawAST,
			t.rawAST)}
}

// BvSDiv creates a "signed division" for bitvector
// created by chz
func (a *Bitvec) BvSDiv(t *Bitvec) *Bitvec {
	return &Bitvec{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvsdiv(
			a.rawCtx,
			a.rawAST,
			t.rawAST)}
}

// BvUDiv creates an "unsigned division" for bitvector
// created by chz
func (a *Bitvec) BvUDiv(t *Bitvec) *Bitvec {
	return &Bitvec{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvudiv(
			a.rawCtx,
			a.rawAST,
			t.rawAST)}
}

// BvAddNoOverflow checks the addition of Node a & t doesn't overflow
// created by chz
func (a *Bitvec) BvAddNoOverflow(t *Bitvec, isSigned bool) *Bitvec {
	return &Bitvec{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvadd_no_overflow(
			a.rawCtx,
			a.rawAST,
			t.rawAST,
			C.bool(isSigned))}
}

// BvSubNoUnderflow checks the subtraction of Node a & t doesn't underflow
// created by chz
func (a *Bitvec) BvSubNoUnderflow(t *Bitvec, isSigned bool) *Bitvec {
	return &Bitvec{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvsub_no_underflow(
			a.rawCtx,
			a.rawAST,
			t.rawAST,
			C.bool(isSigned))}
}

// BvMulNoOverflow checks the multiplication of Node a & t doesn't overflow
// created by chz
func (a *Bitvec) BvMulNoOverflow(t *Bitvec, isSigned bool) *Bitvec {
	return &Bitvec{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvmul_no_overflow(
			a.rawCtx,
			a.rawAST,
			t.rawAST,
			C.bool(isSigned))}
}

// BvSLt creates a "signed <" for bitvector
// created by chz
func (a *Bitvec) BvSLt(a2 *Bitvec) *Bool {
	return &Bool{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvslt(a.rawCtx, a.rawAST, a2.rawAST),
	}
}

// BvSLe creates a "signed <=" for bitvector
// created by chz
func (a *Bitvec) BvSLe(a2 *Bitvec) *Bool {
	return &Bool{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvsle(a.rawCtx, a.rawAST, a2.rawAST),
	}
}

// BvSGt creates a "signed >" for bitvector
// created by chz
func (a *Bitvec) BvSGt(a2 *Bitvec) *Bool {
	return &Bool{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvsgt(a.rawCtx, a.rawAST, a2.rawAST),
	}
}

// BvSGe creates a "signed >=" for bitvector
// created by chz
func (a *Bitvec) BvSGe(a2 *Bitvec) *Bool {
	return &Bool{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvsge(a.rawCtx, a.rawAST, a2.rawAST),
	}
}

// BvULt creates an "unsigned <" for bitvector
// created by chz
func (a *Bitvec) BvULt(a2 *Bitvec) *Bool {
	return &Bool{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvult(a.rawCtx, a.rawAST, a2.rawAST),
	}
}

// BvULe creates an "unsigned <=" for bitvector
// created by chz
func (a *Bitvec) BvULe(a2 *Bitvec) *Bool {
	return &Bool{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvule(a.rawCtx, a.rawAST, a2.rawAST),
	}
}

// BvUGt creates an "unsigned >" for bitvector
// created by chz
func (a *Bitvec) BvUGt(a2 *Bitvec) *Bool {
	return &Bool{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvugt(a.rawCtx, a.rawAST, a2.rawAST),
	}
}

// BvUGe creates an "unsigned >=" for bitvector
// created by chz
func (a *Bitvec) BvUGe(a2 *Bitvec) *Bool {
	return &Bool{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvuge(a.rawCtx, a.rawAST, a2.rawAST),
	}
}

// BvURem gets the unsigned remainder for bitvector
// created by chz
func (a *Bitvec) BvURem(a2 *Bitvec) *Bitvec {
	return &Bitvec{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvurem(a.rawCtx, a.rawAST, a2.rawAST),
	}
}

// BvSRem gets the signed remainder for bitvector
// created by chz
func (a *Bitvec) BvSRem(a2 *Bitvec) *Bitvec {
	return &Bitvec{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvsrem(a.rawCtx, a.rawAST, a2.rawAST),
	}
}

// BvShL gets the shift left of node "a", "a2" is number of shift op.
// created by chz
func (a *Bitvec) BvShL(a2 *Bitvec) *Bitvec {
	return &Bitvec{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvshl(a.rawCtx, a.rawAST, a2.rawAST),
	}
}

// BvShR gets the arithmetical shift right of node "a", "a2" is number of shift op.
// created by chz
func (a *Bitvec) BvShR(a2 *Bitvec) *Bitvec {
	return &Bitvec{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvashr(a.rawCtx, a.rawAST, a2.rawAST),
	}
}

// BvLShR gets the logical shift right of node "a", "a2" is number of shift op.
// created by chz
func (a *Bitvec) BvLShR(a2 *Bitvec) *Bitvec {
	return &Bitvec{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvlshr(a.rawCtx, a.rawAST, a2.rawAST),
	}
}

// The logic op in Bitvec

// Eq gets the ast of bv a == bv a2
// created by chz
func (a *Bitvec) Eq(a2 *Bitvec) *Bool {
	return &Bool{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_eq(a.rawCtx, a.rawAST, a2.rawAST),
	}
}

// BvAnd gets the and of bv a & bv a2
// created by chz
func (a *Bitvec) BvAnd(a2 *Bitvec) *Bitvec {
	return &Bitvec{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvand(a.rawCtx, a.rawAST, a2.rawAST),
	}
}

// BvOr gets the or of bv a & bv a2
// created by chz
func (a *Bitvec) BvOr(a2 *Bitvec) *Bitvec {
	return &Bitvec{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvor(a.rawCtx, a.rawAST, a2.rawAST),
	}
}

// BvXOr gets the exclusive-or of bv a & bv a2
// created by chz
func (a *Bitvec) BvXOr(a2 *Bitvec) *Bitvec {
	return &Bitvec{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvxor(a.rawCtx, a.rawAST, a2.rawAST),
	}
}

// Concat gets the concatenation of bv a & bv a2
// created by chz
func (a *Bitvec) Concat(a2 *Bitvec) *Bitvec {
	return &Bitvec{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_concat(a.rawCtx, a.rawAST, a2.rawAST),
	}
}

// Extract extracts the bv bits from index high to low.
// created by chz
func (a *Bitvec) Extract(high int, low int) *Bitvec {
	return &Bitvec{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_extract(a.rawCtx, C.uint(high), C.uint(low), a.rawAST),
	}
}

// If creates an if(a) t2 then t3 structure. t1 is bool sort, t2 and t3 must be the same sort.
// created by chz
func If(a *Bool, t2 *Bitvec, t3 *Bitvec) *Bitvec {
	return &Bitvec{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_ite(a.rawCtx, a.rawAST, t2.rawAST, t3.rawAST),
	}
}

// GetASTHash gets the hash code for the given AST.
// created by chz
func (a *Bitvec) GetASTHash() uint {
	return uint(C.Z3_get_ast_hash(a.rawCtx, a.rawAST))
}
