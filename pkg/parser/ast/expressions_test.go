package ast

import (
	"github.com/stretchr/testify/assert"
	"regexp"
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
				L:  &Primitive{Typ: FLOAT64, Value: 1.1},
				R:  &Primitive{Typ: INT, Value: 2},
			},
			expected: &Primitive{Typ: FLOAT64, Value: 3.1},
		},
		{
			expr: &BinaryOperationExpr{
				Op: ADD,
				L:  &Primitive{Typ: FLOAT64, Value: 2.0},
				R:  &Primitive{Typ: FLOAT64, Value: 1.1},
			},
			expected: &Primitive{Typ: FLOAT64, Value: 3.1},
		},
		{
			expr: &BinaryOperationExpr{
				Op: ADD,
				L:  &Primitive{Typ: INT, Value: 2},
				R:  &Primitive{Typ: FLOAT64, Value: 1.1},
			},
			expected: &Primitive{Typ: FLOAT64, Value: 3.1},
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
				L:  &Primitive{Typ: FLOAT64, Value: 1.0},
				R:  &Primitive{Typ: INT, Value: 2},
			},
			expected: &Primitive{Typ: FLOAT64, Value: -1.0},
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
				L:  &Primitive{Typ: FLOAT64, Value: 1.0},
				R:  &Primitive{Typ: INT, Value: 2},
			},
			expected: &Primitive{Typ: FLOAT64, Value: 2.0},
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
				L:  &Primitive{Typ: FLOAT64, Value: 1.0},
				R:  &Primitive{Typ: INT, Value: 2},
			},
			expected: &Primitive{Typ: FLOAT64, Value: 0.5},
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
				L:  &Primitive{Typ: FLOAT64, Value: 1.0},
				R:  &Primitive{Typ: INT, Value: 2},
			},
			expected: &Primitive{Typ: FLOAT64, Value: 1.0},
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
				L:  &Primitive{Typ: FLOAT64, Value: 2.0},
				R:  &Primitive{Typ: INT, Value: 3},
			},
			expected: &Primitive{Typ: FLOAT64, Value: 8.0},
		},
		{
			expr: &BinaryOperationExpr{
				Op: EQ,
				L:  &Primitive{Typ: INT, Value: 1},
				R:  &Primitive{Typ: INT, Value: 3},
			},
			expected: &Primitive{Typ: BOOL, Value: false},
		},
		{
			expr: &BinaryOperationExpr{
				Op: NE,
				L:  &Primitive{Typ: INT, Value: 1},
				R:  &Primitive{Typ: INT, Value: 3},
			},
			expected: &Primitive{Typ: BOOL, Value: true},
		},
		{
			expr: &BinaryOperationExpr{
				Op: LT,
				L:  &Primitive{Typ: INT, Value: 1},
				R:  &Primitive{Typ: INT, Value: 1},
			},
			expected: &Primitive{Typ: BOOL, Value: false},
		},
		{
			expr: &BinaryOperationExpr{
				Op: LTE,
				L:  &Primitive{Typ: INT, Value: 1},
				R:  &Primitive{Typ: INT, Value: 1},
			},
			expected: &Primitive{Typ: BOOL, Value: true},
		},
		{
			expr: &BinaryOperationExpr{
				Op: LTE,
				L:  &Primitive{Typ: INT, Value: 1},
				R:  &Primitive{Typ: INT, Value: 2},
			},
			expected: &Primitive{Typ: BOOL, Value: true},
		},
		{
			expr: &BinaryOperationExpr{
				Op: GT,
				L:  &Primitive{Typ: INT, Value: 1},
				R:  &Primitive{Typ: INT, Value: 1},
			},
			expected: &Primitive{Typ: BOOL, Value: false},
		},
		{
			expr: &BinaryOperationExpr{
				Op: GT,
				L:  &Primitive{Typ: INT, Value: 2},
				R:  &Primitive{Typ: INT, Value: 1},
			},
			expected: &Primitive{Typ: BOOL, Value: true},
		},
		{
			expr: &BinaryOperationExpr{
				Op: GTE,
				L:  &Primitive{Typ: INT, Value: 1},
				R:  &Primitive{Typ: INT, Value: 1},
			},
			expected: &Primitive{Typ: BOOL, Value: true},
		},
		{
			expr: &BinaryOperationExpr{
				Op: GTE,
				L:  &Primitive{Typ: FLOAT64, Value: 2.0},
				R:  &Primitive{Typ: FLOAT64, Value: 1.0},
			},
			expected: &Primitive{Typ: BOOL, Value: true},
		},
		{
			expr: &BinaryOperationExpr{
				Op: EQ,
				L:  &Primitive{Typ: BOOL, Value: false},
				R:  &Primitive{Typ: BOOL, Value: false},
			},
			expected: &Primitive{Typ: BOOL, Value: true},
		},
		{
			expr: &BinaryOperationExpr{
				Op: EQ,
				L:  &Primitive{Typ: STRING, Value: "test"},
				R:  &Primitive{Typ: STRING, Value: "test"},
			},
			expected: &Primitive{Typ: BOOL, Value: true},
		},
		{
			expr: &BinaryOperationExpr{
				Op: GT,
				L:  &Primitive{Typ: STRING, Value: "xxx"},
				R:  &Primitive{Typ: STRING, Value: "aaa"},
			},
			expected: &Primitive{Typ: BOOL, Value: true},
		},
		{
			expr: &BinaryOperationExpr{
				Op: EQ,
				L:  &Primitive{Typ: INT, Value: 1},
				R:  &Primitive{Typ: STRING, Value: "aaa"},
			},
			expected: &Primitive{Typ: BOOL, Value: false},
		},
	}
	runTests(sets, t)
}

func TestRegexOperationExpr_Evaluate(t *testing.T) {
	sets := []TestSet{
		{
			expr: &RegexOperationExpr{
				Typ:     RE,
				Target:  "adam[23]",
				Pattern: regexp.MustCompile(`^[a-z]+\[[0-9]+\]$`),
			},
			expected: &Primitive{Typ: BOOL, Value: true},
		},
		{
			expr: &RegexOperationExpr{
				Typ:     RE,
				Target:  "Adam[23]",
				Pattern: regexp.MustCompile(`^[a-z]+\[[0-9]+\]$`),
			},
			expected: &Primitive{Typ: BOOL, Value: false},
		},
		{
			expr: &RegexOperationExpr{
				Typ:     NRE,
				Target:  "Adam[23]",
				Pattern: regexp.MustCompile(`^[a-z]+\[[0-9]+\]$`),
			},
			expected: &Primitive{Typ: BOOL, Value: true},
		},
	}
	runTests(sets, t)
}

func TestLogicalOperationExpr_Evaluate(t *testing.T) {
	sets := []TestSet{
		{
			expr: &BinaryOperationExpr{
				Op: AND_AND,
				L: &Primitive{
					Typ:   BOOL,
					Value: true,
				},
				R: &Primitive{
					Typ:   BOOL,
					Value: true,
				},
			},
			expected: &Primitive{
				Typ:   BOOL,
				Value: true,
			},
		},
		{
			expr: &BinaryOperationExpr{ // '' && 'foo' => ''
				Op: AND_AND,
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
				Op: AND_AND,
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
				Op: AND_AND,
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
				Op: OR_OR,
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
			expr: &BinaryOperationExpr{ // false || 'Dog' => 'Cat'
				Op: OR_OR,
				L: &Primitive{
					Typ:   BOOL,
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
				Op: OR_OR,
				L: &Primitive{
					Typ:   STRING,
					Value: "Cat",
				},
				R: &Primitive{
					Typ:   BOOL,
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
				Op: OR_OR,
				L: &Primitive{
					Typ:   BOOL,
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
				Op: NULL_COALESCENCE,
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
				Op: NULL_COALESCENCE,
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

func TestBetweenOperationExpr_Evaluate(t *testing.T) {
	sets := []TestSet{
		{
			expr: &BinaryOperationExpr{
				Op: BETWEEN,
				L:  &Primitive{Typ: INT, Value: 1},
				R: &Primitive{Typ: TUPLE, Value: []*Primitive{
					{Typ: INT, Value: 1},
					{Typ: INT, Value: 2},
					{Typ: INT, Value: 3},
				}},
			},
			expected: &Primitive{Typ: BOOL, Value: true},
		},
		{
			expr: &BinaryOperationExpr{
				Op: BETWEEN,
				L:  &Primitive{Typ: STRING, Value: "test"},
				R: &Primitive{Typ: TUPLE, Value: []*Primitive{
					{Typ: INT, Value: 1},
					{Typ: INT, Value: 2},
					{Typ: STRING, Value: "test"},
				}},
			},
			expected: &Primitive{Typ: BOOL, Value: true},
		},
		{
			expr: &BinaryOperationExpr{
				Op: BETWEEN,
				L:  &Primitive{Typ: STRING, Value: "foo"},
				R: &Primitive{Typ: TUPLE, Value: []*Primitive{
					{Typ: INT, Value: 1},
					{Typ: INT, Value: 2},
					{Typ: STRING, Value: "test"},
				}},
			},
			expected: &Primitive{Typ: BOOL, Value: false},
		},
	}
	runTests(sets, t)
}

type mockFunc struct {
	naiveFn func(args ...interface{}) (interface{}, error)
	fn      func(args ...*Primitive) (*Primitive, error)
}

func (f *mockFunc) Eval(args ...*Primitive) (*Primitive, error) {
	return f.fn(args...)
}

func (f *mockFunc) NaiveEval(args ...interface{}) (interface{}, error) {
	return f.naiveFn(args...)
}

func TestVariable_Evaluate(t *testing.T) {

	fn1 := mockFunc{
		naiveFn: func(args ...interface{}) (interface{}, error) {
			return 1, nil
		},
		fn: func(args ...*Primitive) (*Primitive, error) {
			return &Primitive{Typ: INT, Value: 1}, nil
		},
	}
	fn2 := mockFunc{
		naiveFn: func(args ...interface{}) (interface{}, error) {
			return args[0].(int) + args[1].(int), nil
		},
		fn: func(args ...*Primitive) (*Primitive, error) {
			return &Primitive{Typ: INT, Value: args[0].Value.(int) + args[1].Value.(int)}, nil
		},
	}

	ctx := NewContext().(*Context)
	ctx.parameters.AddNaiveParameter("a", 1)
	ctx.parameters.AddNaiveParameter("b", 1)

	ctx.functions.AddFunction("fn1", &fn1)
	ctx.functions.AddFunction("fn2", &fn2)
	sets := []TestSet{
		{
			expr: &BinaryOperationExpr{
				Op: ADD,
				L:  &Primitive{Typ: VARIABLE, Value: "a"},
				R:  &Primitive{Typ: VARIABLE, Value: "b"},
			},
			expected: &Primitive{Typ: INT, Value: 2},
			ctx:      ctx,
		},
		{
			expr: &BinaryOperationExpr{
				Op: ADD,
				L:  &ScalarFunction{Ident: "fn1", Args: nil},
				R:  &Primitive{Typ: VARIABLE, Value: "b"},
			},
			expected: &Primitive{Typ: INT, Value: 2},
			ctx:      ctx,
		},
		{
			expr: &BinaryOperationExpr{
				Op: ADD,
				L: &ScalarFunction{Ident: "fn2", Args: []*Primitive{
					{Typ: INT, Value: 1},
					{Typ: INT, Value: 2},
				}},
				R: &Primitive{Typ: VARIABLE, Value: "b"},
			},
			expected: &Primitive{Typ: INT, Value: 4},
			ctx:      ctx,
		},
	}
	runTests(sets, t)

	ctx.functions.AddNaiveFunction("fn1", &fn1)
	ctx.functions.AddNaiveFunction("fn2", &fn2)
	runTests(sets, t)
}
