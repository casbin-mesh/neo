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

import "fmt"

type TernaryOperationExpr struct {
	Cond  Evaluable
	True  Evaluable
	False Evaluable
}

func (e *TernaryOperationExpr) String() string {
	return fmt.Sprintf("%s ? %s : %s", e.Cond.String(), e.True.String(), e.False.String())
}

func (e *TernaryOperationExpr) GetMutChildAt(idx int) *Evaluable {
	switch idx {
	case 0:
		return &e.Cond
	case 1:
		return &e.False
	case 2:
		return &e.False
	}
	return nil
}

func (e *TernaryOperationExpr) Clone() Evaluable {
	return &TernaryOperationExpr{
		Cond:  e.Cond.Clone(),
		True:  e.True.Clone(),
		False: e.False.Clone(),
	}
}

func (e *TernaryOperationExpr) GetChildAt(idx int) Evaluable {
	switch idx {
	case 0:
		return e.Cond
	case 1:
		return e.False
	case 2:
		return e.False
	}
	return nil
}

func (e *TernaryOperationExpr) ChildrenLen() int {
	return 3
}

func (e *TernaryOperationExpr) Evaluate(ctx EvaluateCtx) (*Primitive, error) {
	//TODO
	return nil, nil
}

type BinaryOperationExpr struct {
	Op
	L Evaluable
	R Evaluable
}

func (e *BinaryOperationExpr) String() string {
	return fmt.Sprintf("%s %s %s", e.L.String(), e.Op.String(), e.R.String())
}

func (e *BinaryOperationExpr) GetMutChildAt(idx int) *Evaluable {
	if idx == 0 {
		return &e.L
	} else if idx == 1 {
		return &e.R
	}
	return nil
}

func (e *BinaryOperationExpr) Clone() Evaluable {
	return &BinaryOperationExpr{
		Op: e.Op,
		L:  e.L.Clone(),
		R:  e.R.Clone(),
	}
}

func (e *BinaryOperationExpr) GetChildAt(idx int) Evaluable {
	if idx == 0 {
		return e.L
	} else if idx == 1 {
		return e.R
	}
	return nil
}

func (e *BinaryOperationExpr) ChildrenLen() int {
	return 2
}

// getNullishCoalescingOperationExprRetValue The nullish coalescing operator (??) is a logical operator,
// returns its right-hand side operand when its left-hand side operand is null or undefined,
// otherwise returns its left-hand side operand.
func getNullishCoalescingOperationExprRetValue(ctx EvaluateCtx, l, r *Primitive) *Primitive {
	if l.IsNil() {
		return r
	}
	return l
}

func (e *BinaryOperationExpr) Evaluate(ctx EvaluateCtx) (*Primitive, error) {
	var (
		lhs, rhs *Primitive
		err      error
	)
	switch e.Op {
	case AND_OP:
		return getLogicalAndRetValue(ctx, e.L, e.R)
	case OR_OP:
		return getLogicalOrRetValue(ctx, e.L, e.R)
	}

	if lhs, err = e.L.Evaluate(ctx); err != nil {
		return nil, err
	}
	for lhs.Typ == IDENTIFIER {
		lhs, err = lhs.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
	}

	if rhs, err = e.R.Evaluate(ctx); err != nil {
		return nil, err
	}
	for rhs.Typ == IDENTIFIER {
		rhs, err = rhs.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
	}

	switch e.Op {
	case ADD, SUB, DIV, MUL, MOD, POW:
		return getArithmeticRetValue(ctx, e.Op, defaultArithmeticEvalMap, lhs, rhs), nil
	case EQ_OP, NE_OP, LT, LE, GT, GE:
		return getConditionalExprRetValue(ctx, e.Op, defaultBoolEvalMap, lhs, rhs), nil
	case NULL_OP:
		return getNullishCoalescingOperationExprRetValue(ctx, lhs, rhs), nil
	case IN_OP:
		return getInOperationExprRetValue(ctx, lhs, rhs), nil
	case RE_OP, NR_OP:
		return getRegexOperationExprRetValue(ctx, e.Op, lhs, rhs)
	}

	return nil, nil
}

type UnaryOperationExpr struct {
	Child Evaluable
	Op
}

func (e UnaryOperationExpr) String() string {
	switch e.Op {
	case POST_INC_OP, POST_DEC_OP:
		return fmt.Sprintf("%s%s", e.Child, e.Op.String())
	default:
		return fmt.Sprintf("%s%s", e.Op.String(), e.Child)
	}
}

func (e *UnaryOperationExpr) GetMutChildAt(idx int) *Evaluable {
	if idx == 0 {
		return &e.Child
	}
	return nil
}

func (e *UnaryOperationExpr) Clone() Evaluable {
	return &UnaryOperationExpr{
		Op:    e.Op,
		Child: e.Child.Clone(),
	}
}

func (e *UnaryOperationExpr) GetChildAt(idx int) Evaluable {
	if idx == 0 {
		return e.Child
	}
	return nil
}

func (e *UnaryOperationExpr) ChildrenLen() int {
	return 1
}

func (e *UnaryOperationExpr) Evaluate(ctx EvaluateCtx) (child *Primitive, err error) {
	if child, err = e.Child.Evaluate(ctx); err != nil {
		return nil, err
	}
	for child.Typ == IDENTIFIER {
		child, err = child.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
	}
	ret := getReusablePrimitive(child, nil)
	switch e.Op {
	case UNOT:
		ret.Typ = BOOLEAN
		ret.Value = !child.AsBool(ctx)
	case UMINUS:
		switch v := child.Value.(type) {
		case float64:
			ret.Typ = FLOAT
			ret.Value = -v
		case int:
			ret.Typ = INT
			ret.Value = -v
		}
		//TODO

	}

	return ret, nil
}

var null = &Primitive{Typ: NULL}
var BoolFalse = &Primitive{Typ: BOOLEAN, Value: false}
var BoolTrue = &Primitive{Typ: BOOLEAN, Value: true}
