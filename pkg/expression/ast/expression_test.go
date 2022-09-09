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
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestSet struct {
	expr     Evaluable
	expected *Primitive
	ctx      EvaluateCtx
	err      error
}

func runTests(sets []TestSet, t *testing.T) {
	for i, set := range sets {
		actual, err := set.expr.Evaluate(set.ctx)
		assert.Equal(t, set.err, err)
		assert.Equalf(t, set.expected, actual, "set:%d\n", i)
	}
}

func TestBinaryOperationExpr_Evaluate(t *testing.T) {
	sets := []TestSet{
		{
			expr: &BinaryOperationExpr{
				Op: ADD,
				L:  &Primitive{Typ: INT, Value: 1},
				R:  &Primitive{Typ: INT, Value: 2},
			},
			expected: &Primitive{Typ: INT, Value: 3},
		},
		{
			expr: &BinaryOperationExpr{
				Op: ADD,
				L:  &Primitive{Typ: FLOAT, Value: 1.1},
				R:  &Primitive{Typ: INT, Value: 2},
			},
			expected: &Primitive{Typ: FLOAT, Value: 3.1},
		},
		{
			expr: &BinaryOperationExpr{
				Op: ADD,
				L:  &Primitive{Typ: FLOAT, Value: 2.0},
				R:  &Primitive{Typ: FLOAT, Value: 1.1},
			},
			expected: &Primitive{Typ: FLOAT, Value: 3.1},
		},
		{
			expr: &BinaryOperationExpr{
				Op: ADD,
				L:  &Primitive{Typ: INT, Value: 2},
				R:  &Primitive{Typ: FLOAT, Value: 1.1},
			},
			expected: &Primitive{Typ: FLOAT, Value: 3.1},
		},
		{
			expr: &BinaryOperationExpr{
				Op: SUB,
				L:  &Primitive{Typ: INT, Value: 1},
				R:  &Primitive{Typ: INT, Value: 2},
			},
			expected: &Primitive{Typ: INT, Value: -1},
		},
		{
			expr: &BinaryOperationExpr{
				Op: SUB,
				L:  &Primitive{Typ: FLOAT, Value: 1.0},
				R:  &Primitive{Typ: INT, Value: 2},
			},
			expected: &Primitive{Typ: FLOAT, Value: -1.0},
		},
		{
			expr: &BinaryOperationExpr{
				Op: MUL,
				L:  &Primitive{Typ: INT, Value: 1},
				R:  &Primitive{Typ: INT, Value: 2},
			},
			expected: &Primitive{Typ: INT, Value: 2},
		},
		{
			expr: &BinaryOperationExpr{
				Op: MUL,
				L:  &Primitive{Typ: FLOAT, Value: 1.0},
				R:  &Primitive{Typ: INT, Value: 2},
			},
			expected: &Primitive{Typ: FLOAT, Value: 2.0},
		},
		{
			expr: &BinaryOperationExpr{
				Op: DIV,
				L:  &Primitive{Typ: INT, Value: 1},
				R:  &Primitive{Typ: INT, Value: 2},
			},
			expected: &Primitive{Typ: INT, Value: 0},
		},
		{
			expr: &BinaryOperationExpr{
				Op: DIV,
				L:  &Primitive{Typ: INT, Value: 3},
				R:  &Primitive{Typ: INT, Value: 2},
			},
			expected: &Primitive{Typ: INT, Value: 1},
		},
		{
			expr: &BinaryOperationExpr{
				Op: DIV,
				L:  &Primitive{Typ: FLOAT, Value: 1.0},
				R:  &Primitive{Typ: INT, Value: 2},
			},
			expected: &Primitive{Typ: FLOAT, Value: 0.5},
		},
		{
			expr: &BinaryOperationExpr{
				Op: MOD,
				L:  &Primitive{Typ: INT, Value: 3},
				R:  &Primitive{Typ: INT, Value: 2},
			},
			expected: &Primitive{Typ: INT, Value: 1},
		},
		{
			expr: &BinaryOperationExpr{
				Op: MOD,
				L:  &Primitive{Typ: FLOAT, Value: 1.0},
				R:  &Primitive{Typ: INT, Value: 2},
			},
			expected: &Primitive{Typ: FLOAT, Value: 1.0},
		},
		{
			expr: &BinaryOperationExpr{
				Op: POW,
				L:  &Primitive{Typ: INT, Value: 2},
				R:  &Primitive{Typ: INT, Value: 3},
			},
			expected: &Primitive{Typ: INT, Value: 8},
		},
		{
			expr: &BinaryOperationExpr{
				Op: POW,
				L:  &Primitive{Typ: FLOAT, Value: 2.0},
				R:  &Primitive{Typ: INT, Value: 3},
			},
			expected: &Primitive{Typ: FLOAT, Value: 8.0},
		},
		{
			expr: &BinaryOperationExpr{
				Op: EQ_OP,
				L:  &Primitive{Typ: INT, Value: 1},
				R:  &Primitive{Typ: INT, Value: 3},
			},
			expected: &Primitive{Typ: BOOLEAN, Value: false},
		},
		{
			expr: &BinaryOperationExpr{
				Op: NE_OP,
				L:  &Primitive{Typ: INT, Value: 1},
				R:  &Primitive{Typ: INT, Value: 3},
			},
			expected: &Primitive{Typ: BOOLEAN, Value: true},
		},
		{
			expr: &BinaryOperationExpr{
				Op: LT,
				L:  &Primitive{Typ: INT, Value: 1},
				R:  &Primitive{Typ: INT, Value: 1},
			},
			expected: &Primitive{Typ: BOOLEAN, Value: false},
		},
		{
			expr: &BinaryOperationExpr{
				Op: LE,
				L:  &Primitive{Typ: INT, Value: 1},
				R:  &Primitive{Typ: INT, Value: 1},
			},
			expected: &Primitive{Typ: BOOLEAN, Value: true},
		},
		{
			expr: &BinaryOperationExpr{
				Op: LE,
				L:  &Primitive{Typ: INT, Value: 1},
				R:  &Primitive{Typ: INT, Value: 2},
			},
			expected: &Primitive{Typ: BOOLEAN, Value: true},
		},
		{
			expr: &BinaryOperationExpr{
				Op: GT,
				L:  &Primitive{Typ: INT, Value: 1},
				R:  &Primitive{Typ: INT, Value: 1},
			},
			expected: &Primitive{Typ: BOOLEAN, Value: false},
		},
		{
			expr: &BinaryOperationExpr{
				Op: GT,
				L:  &Primitive{Typ: INT, Value: 2},
				R:  &Primitive{Typ: INT, Value: 1},
			},
			expected: &Primitive{Typ: BOOLEAN, Value: true},
		},
		{
			expr: &BinaryOperationExpr{
				Op: GE,
				L:  &Primitive{Typ: INT, Value: 1},
				R:  &Primitive{Typ: INT, Value: 1},
			},
			expected: &Primitive{Typ: BOOLEAN, Value: true},
		},
		{
			expr: &BinaryOperationExpr{
				Op: GE,
				L:  &Primitive{Typ: FLOAT, Value: 2.0},
				R:  &Primitive{Typ: FLOAT, Value: 1.0},
			},
			expected: &Primitive{Typ: BOOLEAN, Value: true},
		},
		{
			expr: &BinaryOperationExpr{
				Op: EQ_OP,
				L:  &Primitive{Typ: BOOLEAN, Value: false},
				R:  &Primitive{Typ: BOOLEAN, Value: false},
			},
			expected: &Primitive{Typ: BOOLEAN, Value: true},
		},
		{
			expr: &BinaryOperationExpr{
				Op: EQ_OP,
				L:  &Primitive{Typ: STRING, Value: "test"},
				R:  &Primitive{Typ: STRING, Value: "test"},
			},
			expected: &Primitive{Typ: BOOLEAN, Value: true},
		},
		{
			expr: &BinaryOperationExpr{
				Op: GT,
				L:  &Primitive{Typ: STRING, Value: "xxx"},
				R:  &Primitive{Typ: STRING, Value: "aaa"},
			},
			expected: &Primitive{Typ: BOOLEAN, Value: true},
		},
		{
			expr: &BinaryOperationExpr{
				Op: EQ_OP,
				L:  &Primitive{Typ: INT, Value: 1},
				R:  &Primitive{Typ: STRING, Value: "aaa"},
			},
			expected: &Primitive{Typ: BOOLEAN, Value: false},
		},
	}

	runTests(sets, t)
}

func TestRegexOperationExpr_Evaluate(t *testing.T) {
	sets := []TestSet{
		{
			expr: &BinaryOperationExpr{
				Op: RE_OP,
				L:  &Primitive{Typ: STRING, Value: "adam[23]"},
				R:  &Primitive{Typ: STRING, Value: "^[a-z]+\\[[0-9]+\\]$"},
			},
			expected: &Primitive{Typ: BOOLEAN, Value: true},
		},
		{
			expr: &BinaryOperationExpr{
				Op: RE_OP,
				L:  &Primitive{Typ: STRING, Value: "Adam[23]"},
				R:  &Primitive{Typ: STRING, Value: "^[a-z]+\\[[0-9]+\\]$"},
			},
			expected: &Primitive{Typ: BOOLEAN, Value: false},
		},
		{
			expr: &BinaryOperationExpr{
				Op: NR_OP,
				L:  &Primitive{Typ: STRING, Value: "adam[23]"},
				R:  &Primitive{Typ: STRING, Value: "^[a-z]+\\[[0-9]+\\]$"},
			},
			expected: &Primitive{Typ: BOOLEAN, Value: false},
		},
	}
	runTests(sets, t)
}

func TestLogicalOperationExpr_Evaluate(t *testing.T) {
	sets := []TestSet{
		{
			expr: &BinaryOperationExpr{
				Op: AND_OP,
				L: &Primitive{
					Typ:   BOOLEAN,
					Value: true,
				},
				R: &Primitive{
					Typ:   BOOLEAN,
					Value: true,
				},
			},
			expected: &Primitive{
				Typ:   BOOLEAN,
				Value: true,
			},
		},
		{
			expr: &BinaryOperationExpr{ // '' && 'foo' => ''
				Op: AND_OP,
				L: &Primitive{
					Typ:   STRING,
					Value: "",
				},
				R: &Primitive{
					Typ:   STRING,
					Value: "foo",
				},
			},
			expected: &Primitive{
				Typ:   STRING,
				Value: "",
			},
		},
		{
			expr: &BinaryOperationExpr{ // 2 && 0 => 0
				Op: AND_OP,
				L: &Primitive{
					Typ:   INT,
					Value: 2,
				},
				R: &Primitive{
					Typ:   INT,
					Value: 0,
				},
			},
			expected: &Primitive{
				Typ:   INT,
				Value: 0,
			},
		},
		{
			expr: &BinaryOperationExpr{ // 'foo' && 4 => 4
				Op: AND_OP,
				L: &Primitive{
					Typ:   STRING,
					Value: "foo",
				},
				R: &Primitive{
					Typ:   INT,
					Value: 4,
				},
			},
			expected: &Primitive{
				Typ:   INT,
				Value: 4,
			},
		},
		{
			expr: &BinaryOperationExpr{ // 'Cat' || 'Dog' => 'Cat'
				Op: OR_OP,
				L: &Primitive{
					Typ:   STRING,
					Value: "Cat",
				},
				R: &Primitive{
					Typ:   STRING,
					Value: "Dog",
				},
			},
			expected: &Primitive{
				Typ:   STRING,
				Value: "Cat",
			},
		},
		{
			expr: &BinaryOperationExpr{ // false || 'Cat' => 'Cat'
				Op: OR_OP,
				L: &Primitive{
					Typ:   BOOLEAN,
					Value: false,
				},
				R: &Primitive{
					Typ:   STRING,
					Value: "Cat",
				},
			},
			expected: &Primitive{
				Typ:   STRING,
				Value: "Cat",
			},
		},
		{
			expr: &BinaryOperationExpr{ // 'Cat' || false  => 'Cat'
				Op: OR_OP,
				L: &Primitive{
					Typ:   STRING,
					Value: "Cat",
				},
				R: &Primitive{
					Typ:   BOOLEAN,
					Value: false,
				},
			},
			expected: &Primitive{
				Typ:   STRING,
				Value: "Cat",
			},
		},
		{
			expr: &BinaryOperationExpr{ // false || '' => ''
				Op: OR_OP,
				L: &Primitive{
					Typ:   BOOLEAN,
					Value: false,
				},
				R: &Primitive{
					Typ:   STRING,
					Value: "",
				},
			},
			expected: &Primitive{
				Typ:   STRING,
				Value: "",
			},
		},
	}

	runTests(sets, t)
}

func TestNullishCoalescingOperationExpr_Evaluate(t *testing.T) {
	sets := []TestSet{
		{
			expr: &BinaryOperationExpr{
				Op: NULL_OP,
				L: &Primitive{
					Typ:   STRING,
					Value: "default",
				},
				R: &Primitive{
					Typ:   TUPLE,
					Value: nil,
				},
			},
			expected: &Primitive{
				Typ:   STRING,
				Value: "default",
			},
		},
		{
			expr: &BinaryOperationExpr{
				Op: NULL_OP,
				L: &Primitive{
					Typ:   STRING,
					Value: "foo",
				},
				R: &Primitive{
					Typ:   STRING,
					Value: "default",
				},
			},
			expected: &Primitive{
				Typ:   STRING,
				Value: "foo",
			},
		},
	}

	runTests(sets, t)
}

type mockFunc struct {
	fn func(args ...Evaluable) (*Primitive, error)
}

func (f *mockFunc) Eval(ctx EvaluateCtx, args ...Evaluable) (*Primitive, error) {
	return f.fn(args...)
}

func TestIdentifier_Evaluate(t *testing.T) {
	ctx := NewContext()
	ctx.AddParameter("a", Primitive{Typ: INT, Value: 1})
	ctx.AddParameter("b", Primitive{Typ: INT, Value: 1})
	ctx.AddParameter("c", Primitive{Typ: FLOAT, Value: 2.0})
	ctx.AddParameter("str", Primitive{Typ: STRING, Value: "Cat"})
	ctx.AddParameter("bool", Primitive{Typ: BOOLEAN, Value: false})
	sets := []TestSet{
		{
			expr: &BinaryOperationExpr{
				Op: ADD,
				L:  &Primitive{Typ: IDENTIFIER, Value: "a"},
				R:  &Primitive{Typ: IDENTIFIER, Value: "b"},
			},
			expected: &Primitive{Typ: INT, Value: 2},
			ctx:      ctx,
		},
		{
			expr: &BinaryOperationExpr{
				Op: ADD,
				L:  &Primitive{Typ: IDENTIFIER, Value: "a"},
				R:  &Primitive{Typ: IDENTIFIER, Value: "c"},
			},
			expected: &Primitive{Typ: FLOAT, Value: 3.0},
			ctx:      ctx,
		},
		{
			expr: &BinaryOperationExpr{
				Op: OR_OP,
				L:  &Primitive{Typ: IDENTIFIER, Value: "bool"},
				R:  &Primitive{Typ: IDENTIFIER, Value: "str"},
			},
			expected: &Primitive{Typ: STRING, Value: "Cat"},
			ctx:      ctx,
		},
		{
			expr:     &Primitive{Typ: IDENTIFIER, Value: "bool"},
			expected: &Primitive{Typ: BOOLEAN, Value: false},
			ctx:      ctx,
		},
	}
	runTests(sets, t)

}

type mockAccessorValue struct {
}

func (m *mockAccessorValue) GetMember(ident string) *Primitive {
	switch ident {
	case "a":
		return &Primitive{Typ: INT, Value: 1}
	case "b":
		return &Primitive{Typ: INT, Value: 2}
	default:
		return &Primitive{Typ: INT, Value: 0}
	}
}

func TestAccessor_Evaluate(t *testing.T) {
	ctx := NewContext()
	ctx.AddAccessor("obj", &mockAccessorValue{})
	sets := []TestSet{
		{
			expr: &BinaryOperationExpr{
				Op: ADD,
				L: &Accessor{
					Typ:      MEMBER_ACCESSOR,
					Ancestor: &Primitive{Typ: IDENTIFIER, Value: "obj"},
					Ident:    &Primitive{Typ: IDENTIFIER, Value: "a"},
				},
				R: &Accessor{
					Typ:      MEMBER_ACCESSOR,
					Ancestor: &Primitive{Typ: IDENTIFIER, Value: "obj"},
					Ident:    &Primitive{Typ: IDENTIFIER, Value: "b"},
				},
			},
			expected: &Primitive{Typ: INT, Value: 3},
			ctx:      ctx,
		},
	}

	runTests(sets, t)

}

func TestUnaryOperationExpr_Evaluate(t *testing.T) {
	sets := []TestSet{
		{
			expr: &UnaryOperationExpr{
				Child: &Primitive{
					Typ:   BOOLEAN,
					Value: true,
				},
				Op: UNOT,
			},
			expected: &Primitive{
				Typ:   BOOLEAN,
				Value: false,
			},
		},
		{
			expr: &UnaryOperationExpr{
				Child: &Primitive{
					Typ:   INT,
					Value: 1,
				},
				Op: UMINUS,
			},
			expected: &Primitive{
				Typ:   INT,
				Value: -1,
			},
		},
		{
			expr: &UnaryOperationExpr{
				Child: &Primitive{
					Typ:   FLOAT,
					Value: 1.0,
				},
				Op: UMINUS,
			},
			expected: &Primitive{
				Typ:   FLOAT,
				Value: -1.0,
			},
		},
	}
	runTests(sets, t)

}
