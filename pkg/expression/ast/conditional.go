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

import (
	"errors"
	"regexp"
	"strings"
)

type compareFn[T any] func(l, r T) int

func gte[T any](cmp compareFn[T]) EvalCond[T] {
	return func(l, r T) bool {
		return cmp(l, r) >= 0
	}
}

func gt[T any](cmp compareFn[T]) EvalCond[T] {
	return func(l, r T) bool {
		return cmp(l, r) > 0
	}
}

func lt[T any](cmp compareFn[T]) EvalCond[T] {
	return func(l, r T) bool {
		return cmp(l, r) < 0
	}
}

func lte[T any](cmp compareFn[T]) EvalCond[T] {
	return func(l, r T) bool {
		return cmp(l, r) <= 0
	}
}

func ne[T any](cmp compareFn[T]) EvalCond[T] {
	return func(l, r T) bool {
		return cmp(l, r) != 0
	}
}

func eq[T any](cmp compareFn[T]) EvalCond[T] {
	return func(l, r T) bool {
		return cmp(l, r) == 0
	}
}

func compareString(l, r string) int {
	return strings.Compare(l, r)
}

func compareBool(l, r bool) int {
	if l == r {
		return 0
	}
	if !l {
		return -1
	}
	return 1
}

func compareFloat(l, r float64) int {
	if l == r {
		return 0
	} else if l < r {
		return -1
	}
	return 1
}

func compareInt(l, r int) int {
	if l == r {
		return 0
	} else if l < r {
		return -1
	}
	return 1
}

type EvalCond[T any] func(l, r T) bool

type CondEvalMap map[Op]CondEvalFnGroup

type CondEvalFnGroup struct {
	evalInt    EvalCond[int]
	evalFloat  EvalCond[float64]
	evalBool   EvalCond[bool]
	evalString EvalCond[string]
}

var defaultBoolEvalMap = map[Op]CondEvalFnGroup{
	EQ_OP: {
		evalInt:    eq(compareInt),
		evalFloat:  eq(compareFloat),
		evalBool:   eq(compareBool),
		evalString: eq(compareString),
	},
	NE_OP: {
		evalInt:    ne(compareInt),
		evalFloat:  ne(compareFloat),
		evalBool:   ne(compareBool),
		evalString: ne(compareString),
	},
	LT: {
		evalInt:    lt(compareInt),
		evalFloat:  lt(compareFloat),
		evalBool:   lt(compareBool),
		evalString: lt(compareString),
	},
	LE: {
		evalInt:    lte(compareInt),
		evalFloat:  lte(compareFloat),
		evalBool:   lte(compareBool),
		evalString: lte(compareString),
	},
	GT: {
		evalInt:    gt(compareInt),
		evalFloat:  gt(compareFloat),
		evalBool:   gt(compareBool),
		evalString: gt(compareString),
	},
	GE: {
		evalInt:    gte(compareInt),
		evalFloat:  gte(compareFloat),
		evalBool:   gte(compareBool),
		evalString: gte(compareString),
	},
}

func getConditionalExprRetValue(ctx EvaluateCtx, op Op, evalMap CondEvalMap, l, r *Primitive) *Primitive {
	evalGroup := evalMap[op]
	ret := getReusablePrimitive(l, r)
	if l.Typ != r.Typ {
		ret.Typ = BOOLEAN
		ret.Value = false
		return ret
	}
	switch l.Typ {
	case INT:
		ret.Typ = BOOLEAN
		ret.Value = evalGroup.evalInt(l.Value.(int), r.Value.(int))
		return ret
	case FLOAT:
		ret.Typ = BOOLEAN
		ret.Value = evalGroup.evalFloat(l.Value.(float64), r.Value.(float64))
		return ret
	case STRING:
		ret.Typ = BOOLEAN
		ret.Value = evalGroup.evalString(l.Value.(string), r.Value.(string))
		return ret
	case BOOLEAN:
		ret.Typ = BOOLEAN
		ret.Value = evalGroup.evalBool(l.Value.(bool), r.Value.(bool))
		return ret
	case IDENTIFIER:
		// TODO: eval identifier
	}

	return BoolFalse
}

func getInOperationExprRetValue(ctx EvaluateCtx, l, r *Primitive) *Primitive {
	if r.IsNil() {
		return BoolFalse
	}
	for _, elem := range r.Value.([]*Primitive) {
		if l.Equal(elem, ctx) {
			return BoolTrue
		}
	}
	return BoolFalse
}

var (
	ErrInvalidRegexExpr   = errors.New("invalid regex expr")
	ErrCompileRegexFailed = errors.New("compile regex failed")
)

func getRegexOperationExprRetValue(ctx EvaluateCtx, op Op, l, r *Primitive) (*Primitive, error) {
	if r.Typ != STRING || l.Typ != STRING {
		return nil, ErrInvalidRegexExpr
	}
	ret := getReusablePrimitive(l, r)
	ret.Typ = BOOLEAN
	reg, err := regexp.Compile(r.Value.(string))
	if err != nil {
		return nil, ErrCompileRegexFailed
	}

	if op == RE_OP {
		ret.Value = reg.MatchString(l.Value.(string))
	} else {
		ret.Value = !reg.MatchString(l.Value.(string))
	}

	return ret, nil
}
