package z3

// #include <stdlib.h>
// #include "goZ3Config.h"
import "C"

// Sort represents a sort in Z3.
type Sort struct {
	rawCtx  C.Z3_context
	rawSort C.Z3_sort
}

// BoolSort returns the boolean type.
func (c *Context) BoolSort() *Sort {
	return &Sort{
		rawCtx:  c.raw,
		rawSort: C.Z3_mk_bool_sort(c.raw),
	}
}

// IntSort returns the int type.
func (c *Context) IntSort() *Sort {
	return &Sort{
		rawCtx:  c.raw,
		rawSort: C.Z3_mk_int_sort(c.raw),
	}
}

// BvSort returns the bitvector type
// created by chz
func (c *Context) BvSort(size uint) *Sort {
	return &Sort{
		rawCtx:  c.raw,
		rawSort: C.Z3_mk_bv_sort(c.raw, C.uint(size)),
	}
}
