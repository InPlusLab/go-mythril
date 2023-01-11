package z3

// #include <stdlib.h>
// #include "goZ3Config.h"
import "C"
import (
	"unsafe"
)

// Solver is a single solver tied to a specific Context within Z3.
//
// It is created via the NewSolver methods on Context. When a solver is
// no longer needed, the Close method must be called. This will remove the
// solver from the context and no more APIs on Solver may be called
// thereafter.
//
// Freeing the context (Context.Close) will NOT automatically close associated
// solvers. They must be managed separately.
type Solver struct {
	rawCtx    C.Z3_context
	rawSolver C.Z3_solver
}

// NewSolver creates a new solver.
func (c *Context) NewSolver() *Solver {
	rawSolver := C.Z3_mk_solver(c.raw)
	C.Z3_solver_inc_ref(c.raw, rawSolver)

	return &Solver{
		rawSolver: rawSolver,
		rawCtx:    c.raw,
	}
}

// SetTimeout sets the timeout of the solver, timeout is in milliseconds.
func (s *Solver) SetTimeout(time uint) {
	ctx := s.rawCtx
	params := C.Z3_mk_params(ctx)

	ns := C.CString("timeout")
	defer C.free(unsafe.Pointer(ns))
	timeOutSymbol := C.Z3_mk_string_symbol(ctx, ns)

	C.Z3_params_set_uint(ctx, params, timeOutSymbol, C.uint(time))
	C.Z3_solver_set_params(ctx, s.rawSolver, params)
}

// Close frees the memory associated with this.
func (s *Solver) Close() error {
	C.Z3_solver_dec_ref(s.rawCtx, s.rawSolver)
	return nil
}

// Assert asserts a constraint onto the Solver.
//
// Maps to: Z3_solver_assert
func (s *Solver) Assert(args ...*AST) {
	for _, arg := range args {
		C.Z3_solver_assert(s.rawCtx, s.rawSolver, arg.rawAST)
	}
}

// Check checks if the currently set formula is consistent.
//
// Maps to: Z3_solver_check
func (s *Solver) Check() LBool {
	return LBool(C.Z3_solver_check(s.rawCtx, s.rawSolver))
}

func (s *Solver) Statistics() *Statistics {
	return &Statistics{
		RawCtx: s.rawCtx,
		Stats:  C.Z3_solver_get_statistics(s.rawCtx, s.rawSolver),
	}
}

// Model returns the last model from a Check.
//
// Maps to: Z3_solver_get_model
func (s *Solver) Model() *Model {
	m := &Model{
		rawCtx:   s.rawCtx,
		rawModel: C.Z3_solver_get_model(s.rawCtx, s.rawSolver),
	}
	m.IncRef()
	return m
}

// Sexpr returns the formatted string of solver with all constrains.
func (s *Solver) Sexpr() string {
	return C.GoString(C.Z3_solver_to_string(s.rawCtx, s.rawSolver))
}

// Reset just reset the solver
func (s *Solver) Reset() {
	C.Z3_solver_reset(s.rawCtx, s.rawSolver)
}

// Push sets the drawbacking points of the solver
func (s *Solver) Push() {
	C.Z3_solver_push(s.rawCtx, s.rawSolver)
}

// Pop just pops num constrains from the solver
func (s *Solver) Pop(num uint) {
	C.Z3_solver_pop(s.rawCtx, s.rawSolver, C.uint(num))
}

// Translate is used to copy solver from one context to another.
func (s *Solver) Translate(c *Context) *Solver {
	return &Solver{
		rawCtx:    c.raw,
		rawSolver: C.Z3_solver_translate(s.rawCtx, s.rawSolver, c.raw),
	}
}
