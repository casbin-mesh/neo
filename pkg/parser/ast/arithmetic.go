package ast

import "math"

func evalIntAdd(l, r int) int {
	return l + r
}

func evalFloatAdd(l, r float64) float64 {
	return l + r
}

func evalIntSub(l, r int) int {
	return l - r
}

func evalFloatSub(l, r float64) float64 {
	return l - r
}

func evalIntMul(l, r int) int {
	return l * r
}

func evalFloatMul(l, r float64) float64 {
	return l * r
}

func evalIntDiv(l, r int) int {
	return l / r
}

func evalFloatDiv(l, r float64) float64 {
	return l / r
}

func evalIntPow(l, r int) int {
	return int(evalFloatPow(float64(l), float64(r)))
}

func evalFloatPow(l, r float64) float64 {
	return math.Pow(l, r)
}

func evalIntMod(l, r int) int {
	return l % r
}

func evalFloatMod(l, r float64) float64 {
	return float64(evalIntMod(int(l), int(r)))
}

type EvalInt func(int, int) int
type EvalFloat func(float64, float64) float64

type ArithmeticEvalFnGroup struct {
	evalInt   EvalInt
	evalFloat EvalFloat
}

type ArithmeticEvalMap map[Op]ArithmeticEvalFnGroup

var defaultArithmeticEvalMap = map[Op]ArithmeticEvalFnGroup{
	ADD: {evalInt: evalIntAdd, evalFloat: evalFloatAdd},
	SUB: {evalInt: evalIntSub, evalFloat: evalFloatSub},
	DIV: {evalInt: evalIntDiv, evalFloat: evalFloatDiv},
	MUL: {evalInt: evalIntMul, evalFloat: evalFloatMul},
	POW: {evalInt: evalIntPow, evalFloat: evalFloatPow},
	MOD: {evalInt: evalIntMod, evalFloat: evalFloatMod},
}
