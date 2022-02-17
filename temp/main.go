package main

import (
	"fmt"
	"go-mythril/laser/smt/z3"
)

func main() {
	//test1()

	test2()
}

func solve(ctx *z3.Context, realTen int, resultCh chan int) {
	x := ctx.Const(ctx.Symbol("X"), ctx.IntSort())
	y := ctx.Const(ctx.Symbol("Y"), ctx.IntSort())

	s := ctx.NewOptimize()
	// defer s.Close()
	ten := ctx.Int(realTen, ctx.IntSort())
	zero := ctx.Int(0, ctx.IntSort())
	s.Assert(x.Ge(zero), y.Ge(zero), x.Add(y).Eq(ten))
	s.Maximize(x)

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
	s := ctx.NewOptimize()
	// defer s.Close()

	ten := ctx.Int(realTen, ctx.IntSort())
	zero := ctx.Int(0, ctx.IntSort())
	s.Assert(x.Ge(zero), y.Ge(zero), x.Add(y).Eq(ten))
	s.Maximize(x)

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

	total := 10
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

	total := 3
	for i := 0; i < total; i++ {
		go solve(ctx, i, resultCh)
	}

	for i := 0; i < total; i++ {
		res := <-resultCh
		fmt.Println("finish", res)
	}
}
