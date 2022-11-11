package z3

// #include <stdlib.h>
// #include "goZ3Config.h"
import "C"
import (
	"fmt"
	"go-mythril/utils"
	"math/big"
	"strconv"
	"strings"
	"unsafe"
)

type Bitvec struct {
	rawCtx      C.Z3_context
	rawAST      C.Z3_ast
	rawSort     C.Z3_sort
	symbolic    bool
	Annotations *utils.Set
}

// NewBitvec corresponds to BitVec() in Python
// It implements the class BitVecRef
func (c *Context) NewBitvec(bvSymbol string, size int) *Bitvec {
	ast := C.Z3_mk_const(c.raw, c.Symbol(bvSymbol).rawSymbol, c.BvSort(uint(size)).rawSort)
	Annotations := utils.NewSet()
	return &Bitvec{
		rawCtx:      c.raw,
		rawAST:      ast,
		rawSort:     c.BvSort(uint(size)).rawSort,
		symbolic:    true,
		Annotations: Annotations,
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
	tmp := C.CString(valueStr)
	defer C.free(unsafe.Pointer(tmp))
	//ast := C.Z3_mk_numeral(c.raw, C.CString(valueStr), c.BvSort(uint(size)).rawSort)
	ast := C.Z3_mk_numeral(c.raw, tmp, c.BvSort(uint(size)).rawSort)
	Annotations := utils.NewSet()
	return &Bitvec{
		rawCtx:      c.raw,
		rawAST:      ast,
		rawSort:     c.BvSort(uint(size)).rawSort,
		symbolic:    false,
		Annotations: Annotations,
	}
}

// The attribute of bitvec

//func (b *Bitvec) Annotations() *utils.Set {
//	return b.Annotations
//}

func (b *Bitvec) Annotate(item interface{}) {
	b.Annotations.Add(item)
}

// BvSize returns the size of bitvector
// created by chz
func (a *Bitvec) BvSize() int {
	return int(C.Z3_get_bv_sort_size(a.rawCtx, a.rawSort))
}

// AsAST returns the ast of bv
func (b *Bitvec) AsAST() *AST {
	return &AST{
		rawCtx:  b.rawCtx,
		rawAST:  b.rawAST,
		rawSort: b.rawSort,
	}
}

// AsBool returns the BOOL of bv
// used in /solver  _set_minimisation_constraints
func (b *Bitvec) AsBool() *Bool {
	return &Bool{
		rawCtx:      b.rawCtx,
		rawAST:      b.rawAST,
		symbolic:    b.symbolic,
		Annotations: b.Annotations,
	}
}

// Value returns the value of BitvecVal, "" for Bitvec
// should use get_numeral_string rather than get_numeral_int API
// why? 2022.05.06
// Because it only succeeds if the value can fit in a machine int.
func (b *Bitvec) Value() string {
	if b == nil {
		return ""
	}
	if !strings.Contains(b.BvString(), "_") && !b.symbolic && strings.Contains(b.BvString(), "#x") {
		tmp := b.Simplify()
		value := C.GoString(C.Z3_get_numeral_string(tmp.rawCtx, tmp.rawAST))
		//value := "0"
		return value
	} else {
		return "??"
	}
}

func (b *Bitvec) ValueInt() int {
	var dst C.int
	C.Z3_get_numeral_int(b.rawCtx, b.rawAST, &dst)
	return int(dst)
}

// Symbolic tells whether this bv is symbolic
func (b *Bitvec) Symbolic() bool {
	return b.symbolic
}

// BvString returns a human-friendly string version of the AST.
// eg: BvString(BitVec("x",256)) == "x"
func (b *Bitvec) BvString() string {
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

//
func (b *Bitvec) Copy() *Bitvec {
	//var anno *utils.Set
	//if b.Annotations == nil {
	//	anno = utils.NewSet()
	//} else {
	//	anno = b.Annotations.Copy()
	//}

	return &Bitvec{
		rawCtx:   b.rawCtx,
		rawAST:   b.rawAST,
		rawSort:  b.rawSort,
		symbolic: b.symbolic,
		//Annotations: anno,
		Annotations: b.Annotations,
	}
}

// Translate is used to copy ast from one context to another.
func (b *Bitvec) Translate(c *Context) *Bitvec {
	if b.rawCtx == c.raw {
		return b
	}
	return &Bitvec{
		rawCtx: c.raw,
		rawAST: C.Z3_translate(b.rawCtx, b.rawAST, c.raw),
		// TODO: sort translate?
		rawSort:     b.rawSort,
		symbolic:    b.symbolic,
		Annotations: b.Annotations.Copy(),
	}
}

// Simplify equations
func (b *Bitvec) Simplify() *Bitvec {
	return &Bitvec{
		rawCtx: b.rawCtx,
		rawAST: C.Z3_simplify(b.rawCtx, b.rawAST),
		// TODO: wrong use
		rawSort:     b.rawSort,
		symbolic:    b.symbolic,
		Annotations: b.Annotations,
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
		rawSort:     a.rawSort,
		symbolic:    a.symbolic || t.symbolic,
		Annotations: a.Annotations.Union(t.Annotations),
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
			t.rawAST),
		rawSort:     a.rawSort,
		symbolic:    a.symbolic || t.symbolic,
		Annotations: a.Annotations.Union(t.Annotations),
	}
}

// BvMul creates a "multiplication" for bitvector
// created by chz
func (a *Bitvec) BvMul(t *Bitvec) *Bitvec {
	return &Bitvec{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvmul(
			a.rawCtx,
			a.rawAST,
			t.rawAST),
		rawSort:     a.rawSort,
		symbolic:    a.symbolic || t.symbolic,
		Annotations: a.Annotations.Union(t.Annotations),
	}
}

// BvSDiv creates a "signed division" for bitvector
// created by chz
func (a *Bitvec) BvSDiv(t *Bitvec) *Bitvec {
	return &Bitvec{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvsdiv(
			a.rawCtx,
			a.rawAST,
			t.rawAST),
		rawSort:     a.rawSort,
		symbolic:    a.symbolic || t.symbolic,
		Annotations: a.Annotations.Union(t.Annotations),
	}
}

// BvUDiv creates an "unsigned division" for bitvector
// created by chz
func (a *Bitvec) BvUDiv(t *Bitvec) *Bitvec {
	return &Bitvec{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvudiv(
			a.rawCtx,
			a.rawAST,
			t.rawAST),
		rawSort:     a.rawSort,
		symbolic:    a.symbolic || t.symbolic,
		Annotations: a.Annotations.Union(t.Annotations),
	}
}

// BvAddNoOverflow checks the addition of Node a & t doesn't overflow
// created by chz
func (a *Bitvec) BvAddNoOverflow(t *Bitvec, isSigned bool) *Bool {
	return &Bool{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvadd_no_overflow(
			a.rawCtx,
			a.rawAST,
			t.rawAST,
			C.bool(isSigned)),
		symbolic:    a.symbolic || t.symbolic,
		Annotations: a.Annotations.Union(t.Annotations),
	}
}

// BvSubNoUnderflow checks the subtraction of Node a & t doesn't underflow
// created by chz
func (a *Bitvec) BvSubNoUnderflow(t *Bitvec, isSigned bool) *Bool {
	return &Bool{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvsub_no_underflow(
			a.rawCtx,
			a.rawAST,
			t.rawAST,
			C.bool(isSigned)),
		symbolic:    a.symbolic || t.symbolic,
		Annotations: a.Annotations.Union(t.Annotations),
	}
}

// BvMulNoOverflow checks the multiplication of Node a & t doesn't overflow
// created by chz
func (a *Bitvec) BvMulNoOverflow(t *Bitvec, isSigned bool) *Bool {
	return &Bool{
		rawCtx: a.rawCtx,
		rawAST: C.Z3_mk_bvmul_no_overflow(
			a.rawCtx,
			a.rawAST,
			t.rawAST,
			C.bool(isSigned)),
		symbolic:    a.symbolic || t.symbolic,
		Annotations: a.Annotations.Union(t.Annotations),
	}
}

// BvSLt creates a "signed <" for bitvector
// created by chz
func (a *Bitvec) BvSLt(a2 *Bitvec) *Bool {
	return &Bool{
		rawCtx:      a.rawCtx,
		rawAST:      C.Z3_mk_bvslt(a.rawCtx, a.rawAST, a2.rawAST),
		symbolic:    a.symbolic || a2.symbolic,
		Annotations: a.Annotations.Union(a2.Annotations),
	}
}

// BvSLe creates a "signed <=" for bitvector
// created by chz
func (a *Bitvec) BvSLe(a2 *Bitvec) *Bool {
	return &Bool{
		rawCtx:      a.rawCtx,
		rawAST:      C.Z3_mk_bvsle(a.rawCtx, a.rawAST, a2.rawAST),
		symbolic:    a.symbolic || a2.symbolic,
		Annotations: a.Annotations.Union(a2.Annotations),
	}
}

// BvSGt creates a "signed >" for bitvector
// created by chz
func (a *Bitvec) BvSGt(a2 *Bitvec) *Bool {
	return &Bool{
		rawCtx:      a.rawCtx,
		rawAST:      C.Z3_mk_bvsgt(a.rawCtx, a.rawAST, a2.rawAST),
		symbolic:    a.symbolic || a2.symbolic,
		Annotations: a.Annotations.Union(a2.Annotations),
	}
}

// BvSGe creates a "signed >=" for bitvector
// created by chz
func (a *Bitvec) BvSGe(a2 *Bitvec) *Bool {
	return &Bool{
		rawCtx:      a.rawCtx,
		rawAST:      C.Z3_mk_bvsge(a.rawCtx, a.rawAST, a2.rawAST),
		symbolic:    a.symbolic || a2.symbolic,
		Annotations: a.Annotations.Union(a2.Annotations),
	}
}

// BvULt creates an "unsigned <" for bitvector
// created by chz
func (a *Bitvec) BvULt(a2 *Bitvec) *Bool {
	return &Bool{
		rawCtx:      a.rawCtx,
		rawAST:      C.Z3_mk_bvult(a.rawCtx, a.rawAST, a2.rawAST),
		symbolic:    a.symbolic || a2.symbolic,
		Annotations: a.Annotations.Union(a2.Annotations),
	}
}

// BvULe creates an "unsigned <=" for bitvector
// created by chz
func (a *Bitvec) BvULe(a2 *Bitvec) *Bool {
	return &Bool{
		rawCtx:      a.rawCtx,
		rawAST:      C.Z3_mk_bvule(a.rawCtx, a.rawAST, a2.rawAST),
		symbolic:    a.symbolic || a2.symbolic,
		Annotations: a.Annotations.Union(a2.Annotations),
	}
}

// BvUGt creates an "unsigned >" for bitvector
// created by chz
func (a *Bitvec) BvUGt(a2 *Bitvec) *Bool {
	return &Bool{
		rawCtx:      a.rawCtx,
		rawAST:      C.Z3_mk_bvugt(a.rawCtx, a.rawAST, a2.rawAST),
		symbolic:    a.symbolic || a2.symbolic,
		Annotations: a.Annotations.Union(a2.Annotations),
	}
}

// BvUGe creates an "unsigned >=" for bitvector
// created by chz
func (a *Bitvec) BvUGe(a2 *Bitvec) *Bool {
	return &Bool{
		rawCtx:      a.rawCtx,
		rawAST:      C.Z3_mk_bvuge(a.rawCtx, a.rawAST, a2.rawAST),
		symbolic:    a.symbolic || a2.symbolic,
		Annotations: a.Annotations.Union(a2.Annotations),
	}
}

// BvURem gets the unsigned remainder for bitvector
// created by chz
func (a *Bitvec) BvURem(a2 *Bitvec) *Bitvec {
	return &Bitvec{
		rawCtx:      a.rawCtx,
		rawAST:      C.Z3_mk_bvurem(a.rawCtx, a.rawAST, a2.rawAST),
		rawSort:     a.rawSort,
		symbolic:    a.symbolic || a2.symbolic,
		Annotations: a.Annotations.Union(a2.Annotations),
	}
}

// BvSRem gets the signed remainder for bitvector
// created by chz
func (a *Bitvec) BvSRem(a2 *Bitvec) *Bitvec {
	return &Bitvec{
		rawCtx:      a.rawCtx,
		rawAST:      C.Z3_mk_bvsrem(a.rawCtx, a.rawAST, a2.rawAST),
		rawSort:     a.rawSort,
		symbolic:    a.symbolic || a2.symbolic,
		Annotations: a.Annotations.Union(a2.Annotations),
	}
}

// BvShL gets the shift left of node "a", "a2" is number of shift op.
// created by chz
func (a *Bitvec) BvShL(a2 *Bitvec) *Bitvec {
	return &Bitvec{
		rawCtx:      a.rawCtx,
		rawAST:      C.Z3_mk_bvshl(a.rawCtx, a.rawAST, a2.rawAST),
		rawSort:     a.rawSort,
		symbolic:    a.symbolic || a2.symbolic,
		Annotations: a.Annotations.Union(a2.Annotations),
	}
}

// BvShR gets the arithmetical shift right of node "a", "a2" is number of shift op.
// created by chz
func (a *Bitvec) BvShR(a2 *Bitvec) *Bitvec {
	return &Bitvec{
		rawCtx:      a.rawCtx,
		rawAST:      C.Z3_mk_bvashr(a.rawCtx, a.rawAST, a2.rawAST),
		rawSort:     a.rawSort,
		symbolic:    a.symbolic || a2.symbolic,
		Annotations: a.Annotations.Union(a2.Annotations),
	}
}

// BvLShR gets the logical shift right of node "a", "a2" is number of shift op.
// created by chz
func (a *Bitvec) BvLShR(a2 *Bitvec) *Bitvec {
	return &Bitvec{
		rawCtx:      a.rawCtx,
		rawAST:      C.Z3_mk_bvlshr(a.rawCtx, a.rawAST, a2.rawAST),
		rawSort:     a.rawSort,
		symbolic:    a.symbolic || a2.symbolic,
		Annotations: a.Annotations.Union(a2.Annotations),
	}
}

// The logic op in Bitvec

// Eq gets the ast of bv a == bv a2
// created by chz
func (a *Bitvec) Eq(a2 *Bitvec) *Bool {
	return &Bool{
		rawCtx:      a.rawCtx,
		rawAST:      C.Z3_mk_eq(a.rawCtx, a.rawAST, a2.rawAST),
		symbolic:    a.symbolic || a2.symbolic,
		Annotations: a.Annotations.Union(a2.Annotations),
	}
}

func (a *Bitvec) Neq(a2 *Bitvec) *Bool {
	raws := make([]C.Z3_ast, 2)
	raws[0] = a.rawAST
	raws[1] = a2.rawAST

	return &Bool{
		rawCtx: a.rawCtx,
		//rawAST:      C.Z3_mk_eq(a.rawCtx, a.rawAST, a2.rawAST),
		rawAST:      C.Z3_mk_distinct(a.rawCtx, C.uint(2), (*C.Z3_ast)(unsafe.Pointer(&raws[0]))),
		symbolic:    a.symbolic || a2.symbolic,
		Annotations: a.Annotations.Union(a2.Annotations),
	}
}

// BvAnd gets the and of bv a & bv a2
// created by chz
func (a *Bitvec) BvAnd(a2 *Bitvec) *Bitvec {
	return &Bitvec{
		rawCtx:      a.rawCtx,
		rawAST:      C.Z3_mk_bvand(a.rawCtx, a.rawAST, a2.rawAST),
		rawSort:     a.rawSort,
		symbolic:    a.symbolic || a2.symbolic,
		Annotations: a.Annotations.Union(a2.Annotations),
	}
}

// BvOr gets the or of bv a & bv a2
// created by chz
func (a *Bitvec) BvOr(a2 *Bitvec) *Bitvec {
	return &Bitvec{
		rawCtx:      a.rawCtx,
		rawAST:      C.Z3_mk_bvor(a.rawCtx, a.rawAST, a2.rawAST),
		rawSort:     a.rawSort,
		symbolic:    a.symbolic || a2.symbolic,
		Annotations: a.Annotations.Union(a2.Annotations),
	}
}

// BvXOr gets the exclusive-or of bv a & bv a2
// created by chz
func (a *Bitvec) BvXOr(a2 *Bitvec) *Bitvec {
	return &Bitvec{
		rawCtx:      a.rawCtx,
		rawAST:      C.Z3_mk_bvxor(a.rawCtx, a.rawAST, a2.rawAST),
		rawSort:     a.rawSort,
		symbolic:    a.symbolic || a2.symbolic,
		Annotations: a.Annotations.Union(a2.Annotations),
	}
}

// Concat gets the concatenation of bv a & bv a2
// created by chz
func (a *Bitvec) Concat(a2 *Bitvec) *Bitvec {
	bvSize1 := a.BvSize()
	bvSize2 := a2.BvSize()
	return &Bitvec{
		rawCtx:  a.rawCtx,
		rawAST:  C.Z3_mk_concat(a.rawCtx, a.rawAST, a2.rawAST),
		rawSort: C.Z3_mk_bv_sort(a.rawCtx, C.uint(bvSize1+bvSize2)),
		//rawSort: a.rawSort,
		symbolic:    a.symbolic || a2.symbolic,
		Annotations: a.Annotations.Union(a2.Annotations),
	}
}

// Extract extracts the bv bits from index high to low.
// created by chz
func (a *Bitvec) Extract(high int, low int) *Bitvec {
	return &Bitvec{
		rawCtx:  a.rawCtx,
		rawAST:  C.Z3_mk_extract(a.rawCtx, C.uint(high), C.uint(low), a.rawAST),
		rawSort: C.Z3_mk_bv_sort(a.rawCtx, C.uint(high-low+1)),
		//rawSort: a.rawSort,
		symbolic:    a.symbolic,
		Annotations: a.Annotations,
	}
}

// If creates an if(a) t2 then t3 structure. t1 is bool sort, t2 and t3 must be the same sort.
// created by chz
func If(a *Bool, t2 *Bitvec, t3 *Bitvec) *Bitvec {
	anno1 := a.Annotations.Union(t2.Annotations)
	return &Bitvec{
		rawCtx:      a.rawCtx,
		rawAST:      C.Z3_mk_ite(a.rawCtx, a.rawAST, t2.rawAST, t3.rawAST),
		rawSort:     t2.rawSort,
		symbolic:    a.symbolic || t2.symbolic || t3.symbolic,
		Annotations: anno1.Union(t3.Annotations),
	}
}

// GetASTHash gets the hash code for the given AST.
// created by chz
func (a *Bitvec) GetASTHash() uint {
	return uint(C.Z3_get_ast_hash(a.rawCtx, a.rawAST))
}
