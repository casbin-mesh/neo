// Copyright 2022 The casbin-neo Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

func hasInt(l, r *Primitive) (bool, bool) {
	return l.Typ == INT, r.Typ == INT
}

func hasFloat64(l, r *Primitive) (bool, bool) {
	return l.Typ == FLOAT, r.Typ == FLOAT
}

func getArithmeticRetValue(ctx EvaluateCtx, op Op, evalMap ArithmeticEvalMap, l, r *Primitive) *Primitive {
	lInt, rInt := hasInt(l, r)
	lFloat, rFloat := hasFloat64(l, r)
	evalInt := evalMap[op].evalInt
	evalFloat := evalMap[op].evalFloat
	ret := getReusablePrimitive(l, r)

	if lInt && rInt {
		ret.Typ = INT
		ret.Value = evalInt(l.Value.(int), r.Value.(int))
	} else if lFloat && rFloat {
		ret.Typ = FLOAT
		ret.Value = evalFloat(l.Value.(float64), r.Value.(float64))
	} else {
		if lFloat && rInt {
			ret.Typ = FLOAT
			ret.Value = evalFloat(l.Value.(float64), float64(r.Value.(int)))
		} else if lInt && rFloat {
			ret.Typ = FLOAT
			ret.Value = evalFloat(float64(l.Value.(int)), r.Value.(float64))
		}
	}
	return ret
}
