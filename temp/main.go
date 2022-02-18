package main

import (
	"fmt"
	"go-mythril/laser/smt/z3"
)

func main() {
	//test1()

	//test2()

	//test3()

	test4()
}

func solve(ctx *z3.Context, realTen int, resultCh chan int) {
	x := ctx.Const(ctx.Symbol("X"), ctx.IntSort())
	y := ctx.Const(ctx.Symbol("Y"), ctx.IntSort())

	s := ctx.NewSolver()
	// defer s.Close()
	ten := ctx.Int(realTen, ctx.IntSort())
	zero := ctx.Int(0, ctx.IntSort())
	s.Assert(x.Ge(zero), y.Ge(zero), x.Add(y).Eq(ten))

	if v := s.CheckWrapper(); v != z3.True {
		fmt.Println("Unsolveable")
		return
	}

	m := s.Model()
	answer := m.Assignments()
	fmt.Println(answer)

	instance := z3.GetSolverStatistic()
	fmt.Println(instance.SolverTime)

	resultCh <- realTen
}
func solveXY(ctx *z3.Context, x *z3.AST, y *z3.AST, realTen int, resultCh chan int) {
	s := ctx.NewSolver()
	//defer s.Close()

	ten := ctx.Int(realTen, ctx.IntSort())
	zero := ctx.Int(0, ctx.IntSort())
	s.Assert(x.Ge(zero), y.Ge(zero), x.Add(y).Eq(ten))

	if v := s.CheckWrapper(); v != z3.True {
		fmt.Println("Unsolveable")
		return
	}

	m := s.Model()
	answer := m.Assignments()
	fmt.Println(answer)

	instance := z3.GetSolverStatistic()
	fmt.Println(instance.SolverTime)

	resultCh <- realTen
}
func solveCtx(cg *z3.Config, realTen int, resultCh chan int) {
	ctx := z3.NewContext(cg)

	x := ctx.Const(ctx.Symbol("X"), ctx.IntSort())
	y := ctx.Const(ctx.Symbol("Y"), ctx.IntSort())

	s := ctx.NewSolver()
	//defer s.Close()

	ten := ctx.Int(realTen, ctx.IntSort())
	zero := ctx.Int(0, ctx.IntSort())
	s.Assert(x.Ge(zero), y.Ge(zero), x.Add(y).Eq(ten))

	if v := s.CheckWrapper(); v != z3.True {
		fmt.Println("Unsolveable")
		return
	}

	m := s.Model()
	answer := m.Assignments()
	fmt.Println(answer)

	instance := z3.GetSolverStatistic()
	fmt.Println(instance.SolverTime)

	resultCh <- realTen
}
func solveTranslate(newCtx *z3.Context, x *z3.AST, y *z3.AST, realTen int, resultCh chan int) {

	newX := x.Translate(newCtx)
	newY := y.Translate(newCtx)

	s := newCtx.NewSolver()
	// defer s.Close()
	ten := newCtx.Int(realTen, newCtx.IntSort())
	zero := newCtx.Int(0, newCtx.IntSort())
	s.Assert(newX.Ge(zero), newY.Ge(zero), newX.Add(newY).Eq(ten))

	if v := s.CheckWrapper(); v != z3.True {
		fmt.Println("Unsolveable")
		return
	}

	m := s.Model()
	answer := m.Assignments()
	fmt.Println(answer)

	instance := z3.GetSolverStatistic()
	fmt.Println(instance.SolverTime)

	resultCh <- realTen
}

func test1() {

	fmt.Println("go mythril-testForGoz3")
	config := z3.NewConfig()
	ctx := z3.NewContext(config)
	config.Close()

	// do not close it as go func need it
	// defer ctx.Close()

	x := ctx.Const(ctx.Symbol("X"), ctx.IntSort())
	y := ctx.Const(ctx.Symbol("Y"), ctx.IntSort())

	resultCh := make(chan int)

	total := 5
	for i := 0; i < total; i++ {
		go solveXY(ctx, x, y, i, resultCh)
	}

	for i := 0; i < total; i++ {
		res := <-resultCh
		fmt.Println("finish", res)
	}
}

func test2() {

	fmt.Println("go mythril-testForGoz3")
	config := z3.NewConfig()
	ctx := z3.NewContext(config)
	config.Close()

	// do not close it as go func need it
	// defer ctx.Close()

	resultCh := make(chan int)

	total := 5
	for i := 0; i < total; i++ {
		go solve(ctx, i, resultCh)
	}

	for i := 0; i < total; i++ {
		res := <-resultCh
		fmt.Println("finish", res)
	}
}
func test3() {

	fmt.Println("go mythril-testForGoz3")
	config := z3.NewConfig()
	defer config.Close()

	// do not close it as go func need it
	// defer ctx.Close()

	resultCh := make(chan int)

	total := 100
	for i := 0; i < total; i++ {
		go solveCtx(config, i, resultCh)
	}

	for i := 0; i < total; i++ {
		res := <-resultCh
		fmt.Println("finish", res)
	}
}
func test4() {
	fmt.Println("go mythril-testForGoz3")
	config := z3.NewConfig()
	ctx := z3.NewContext(config)
	defer config.Close()

	// do not close it as go func need it
	// defer ctx.Close()

	x := ctx.Const(ctx.Symbol("X"), ctx.IntSort())
	y := ctx.Const(ctx.Symbol("Y"), ctx.IntSort())

	resultCh := make(chan int)

	total := 100
	for i := 0; i < total; i++ {
		newCtx := z3.NewContext(config)
		go solveTranslate(newCtx, x, y, i, resultCh)
	}

	for i := 0; i < total; i++ {
		res := <-resultCh
		fmt.Println("finish", res)
	}
}
