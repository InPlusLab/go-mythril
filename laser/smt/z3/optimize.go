package z3

// #include <stdlib.h>
// #include "goZ3Config.h"
import "C"
import (
	"unsafe"
)

type Optimize struct {
	rawCtx      C.Z3_context
	rawOptimize C.Z3_optimize
}

// NewOptimize creates a new solver.
func (c *Context) NewOptimize() *Optimize {
	rawOptimize := C.Z3_mk_optimize(c.raw)
	C.Z3_optimize_inc_ref(c.raw, rawOptimize)

	return &Optimize{
		rawOptimize: rawOptimize,
		rawCtx:      c.raw,
	}
}

// SetTimeout sets the timeout of the optimize, timeout is in milliseconds.
func (s *Optimize) SetTimeout(time int) {
	ctx := s.rawCtx
	params := C.Z3_mk_params(ctx)

	ns := C.CString("timeout")
	defer C.free(unsafe.Pointer(ns))
	timeOutSymbol := C.Z3_mk_string_symbol(ctx, ns)

	//C.Z3_params_set_uint(ctx, params, timeOutSymbol, C.uint(time))

	C.Z3_params_set_uint(ctx, params, timeOutSymbol, C.uint(time))
	C.Z3_optimize_set_params(ctx, s.rawOptimize, params)
}

// RLimit set
// TODO: to be tested
func (s *Optimize) RLimit(time int) {
	ctx := s.rawCtx
	params := C.Z3_mk_params(ctx)

	ns := C.CString("rlimit")
	defer C.free(unsafe.Pointer(ns))
	timeOutSymbol := C.Z3_mk_string_symbol(ctx, ns)

	C.Z3_params_set_uint(ctx, params, timeOutSymbol, C.uint(time))
	C.Z3_optimize_set_params(ctx, s.rawOptimize, params)
}

// Close frees the memory associated with this.
func (s *Optimize) Close() error {
	C.Z3_optimize_dec_ref(s.rawCtx, s.rawOptimize)
	return nil
}

// Assert asserts a constraint onto the Optimize.
func (s *Optimize) Assert(args ...*AST) {
	for _, arg := range args {
		C.Z3_optimize_assert(s.rawCtx, s.rawOptimize, arg.rawAST)
	}
}

// Check checks if the currently set formula is consistent.
func (s *Optimize) Check(args ...*AST) LBool {
	length := len(args)
	if length != 0 {
		raws := make([]C.Z3_ast, length)
		for i, arg := range args {
			raws[i] = arg.rawAST
		}
		return LBool(C.Z3_optimize_check(s.rawCtx, s.rawOptimize, C.uint(length),
			(*C.Z3_ast)(unsafe.Pointer(&raws[0]))))
	} else {
		//tmp := AST{rawCtx: s.rawCtx, rawAST: nil}
		//fmt.Println("optimize 2")
		return LBool(C.Z3_optimize_check(s.rawCtx, s.rawOptimize, C.uint(0), nil))
		//return LBool(C.Z3_optimize_check(s.rawCtx, s.rawOptimize, 0, nil))
	}
}

func (s *Optimize) Statistics() *Statistics {
	return &Statistics{
		RawCtx: s.rawCtx,
		Stats:  C.Z3_optimize_get_statistics(s.rawCtx, s.rawOptimize),
	}
}

// Model returns the last model from a Check.
func (s *Optimize) Model() *Model {
	m := &Model{
		rawCtx:   s.rawCtx,
		rawModel: C.Z3_optimize_get_model(s.rawCtx, s.rawOptimize),
	}
	m.IncRef()
	return m
}

// Sexpr returns the formatted string of solver with all constrains.
func (s *Optimize) Sexpr() string {
	return C.GoString(C.Z3_optimize_to_string(s.rawCtx, s.rawOptimize))
}

func (s *Optimize) Minimize(ast *AST) {
	C.Z3_optimize_minimize(s.rawCtx, s.rawOptimize, ast.rawAST)
}
func (s *Optimize) Maximize(ast *AST) {
	C.Z3_optimize_maximize(s.rawCtx, s.rawOptimize, ast.rawAST)
}
