package function_managers

import (
	"go-mythril/laser/smt/z3"
	"math/big"
	"strconv"
)

func CreateCondition(base *z3.Bitvec, exponent *z3.Bitvec, ctx *z3.Context) (*z3.Bitvec, *z3.Bool) {

	domain := make([]*z3.Sort, 0)
	domain = append(domain, ctx.BvSort(uint(256)), ctx.BvSort(uint(256)))
	power := ctx.NewFuncDecl("Power", domain, ctx.BvSort(uint(256)))
	exponentiation := power.ApplyBv(base, exponent)

	if !base.Symbolic() && !exponent.Symbolic() {
		// TODO: MAX value check
		baseV, _ := strconv.ParseInt(base.Value(), 10, 64)
		exponentV, _ := strconv.ParseInt(exponent.Value(), 10, 64)
		val := new(big.Int).Exp(big.NewInt(baseV), big.NewInt(exponentV), nil)
		constExponentiation := ctx.NewBitvecVal(val, 256)
		constExponentiation.Annotations = base.Annotations.Union(exponent.Annotations)

		constraint := constExponentiation.Eq(exponentiation)
		//constraint = ctx.NewBitvecVal(1,256).Eq(ctx.NewBitvecVal(1,256)).Simplify()
		return constExponentiation, constraint
	}
	// TODO:
	return nil, nil
}
