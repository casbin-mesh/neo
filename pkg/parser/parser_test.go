package parser

import (
	"fmt"
	"github.com/casbin-mesh/neo/pkg/parser/ast"
	"github.com/stretchr/testify/assert"
	"regexp"
	"strings"
	"testing"
)

func newLexer(token string) *Lexer {
	return NewLexer(strings.NewReader(token))
}

func TestParse(t *testing.T) {
	s := NewLexer(strings.NewReader("1+3"))
	yyParse(s)
	fmt.Printf("%v\n", s.parseResult)
}

type TestSet struct {
	parseStr string
	expected interface{}
	err      error
}

func asserter(t *testing.T, expected, actual interface{}) {
	switch e := expected.(type) {
	case *ast.RegexOperationExpr:
		a, ok := actual.(*ast.RegexOperationExpr)
		assert.True(t, ok)
		assert.Equal(t, e.Typ, a.Typ)
		assert.Equal(t, e.Target, a.Target)
		assert.Equal(t, e.Pattern.String(), a.Pattern.String())
	default:
		assert.Equal(t, expected, actual)
	}
}

func runTests(sets []TestSet, t *testing.T) {
	for _, set := range sets {
		s := newLexer(set.parseStr)
		yyParse(s)
		asserter(t, set.expected, s.parseResult)
	}
}

func TestPrimitives(t *testing.T) {
	sets := []TestSet{
		{
			parseStr: "\"should be string\"",
			expected: &ast.Primitive{Typ: ast.STRING, Value: "should be string"},
		},
		{
			parseStr: "1",
			expected: &ast.Primitive{Typ: ast.INT, Value: int(1)},
		},
		{
			parseStr: "-1",
			expected: &ast.Primitive{Typ: ast.INT, Value: int(-1)},
		},
		{
			parseStr: "1.00",
			expected: &ast.Primitive{Typ: ast.FLOAT64, Value: float64(1)},
		},
		{
			parseStr: "-1.00",
			expected: &ast.Primitive{Typ: ast.FLOAT64, Value: float64(-1)},
		},
		{
			parseStr: "true",
			expected: &ast.Primitive{Typ: ast.BOOL, Value: true},
		},
		{
			parseStr: "!true",
			expected: &ast.Primitive{Typ: ast.BOOL, Value: false},
		},
		{
			parseStr: "false",
			expected: &ast.Primitive{Typ: ast.BOOL, Value: false},
		},
		{
			parseStr: "TRUE",
			expected: &ast.Primitive{Typ: ast.BOOL, Value: true},
		},
		{
			parseStr: "FALSE",
			expected: &ast.Primitive{Typ: ast.BOOL, Value: false},
		},
		{
			parseStr: "imVariable",
			expected: &ast.Primitive{Typ: ast.VARIABLE, Value: "imVariable"},
		},
		{
			parseStr: "[1,2,\"string\"]",
			expected: &ast.Primitive{Typ: ast.TUPLE, Value: []*ast.Primitive{
				{Typ: ast.INT, Value: 1},
				{Typ: ast.INT, Value: 2},
				{Typ: ast.STRING, Value: "string"},
			}},
		},
		{
			parseStr: "testc(1,2,\"string\")",
			expected: &ast.ScalarFunction{Ident: "testc", Args: []*ast.Primitive{
				{Typ: ast.INT, Value: 1},
				{Typ: ast.INT, Value: 2},
				{Typ: ast.STRING, Value: "string"},
			}},
		},
	}
	runTests(sets, t)
}

func TestBinaryOperationExprs(t *testing.T) {
	sets := []TestSet{
		/* arithmetic */
		{
			parseStr: "1+2",
			expected: &ast.BinaryOperationExpr{
				Op: ast.ADD,
				L:  &ast.Primitive{Typ: ast.INT, Value: 1},
				R:  &ast.Primitive{Typ: ast.INT, Value: 2},
			},
		},
		{
			parseStr: "1*2",
			expected: &ast.BinaryOperationExpr{
				Op: ast.MUL,
				L:  &ast.Primitive{Typ: ast.INT, Value: 1},
				R:  &ast.Primitive{Typ: ast.INT, Value: 2},
			},
		},
		{
			parseStr: "1/2",
			expected: &ast.BinaryOperationExpr{
				Op: ast.DIV,
				L:  &ast.Primitive{Typ: ast.INT, Value: 1},
				R:  &ast.Primitive{Typ: ast.INT, Value: 2},
			},
		},
		{
			parseStr: "1-2",
			expected: &ast.BinaryOperationExpr{
				Op: ast.SUB,
				L:  &ast.Primitive{Typ: ast.INT, Value: 1},
				R:  &ast.Primitive{Typ: ast.INT, Value: 2},
			},
		},
		{
			parseStr: "1**2",
			expected: &ast.BinaryOperationExpr{
				Op: ast.POW,
				L:  &ast.Primitive{Typ: ast.INT, Value: 1},
				R:  &ast.Primitive{Typ: ast.INT, Value: 2},
			},
		},
		/* conditional expr */
		{
			parseStr: "1==2",
			expected: &ast.BinaryOperationExpr{
				Op: ast.EQ,
				L:  &ast.Primitive{Typ: ast.INT, Value: 1},
				R:  &ast.Primitive{Typ: ast.INT, Value: 2},
			},
		},
		{
			parseStr: "1!=2",
			expected: &ast.BinaryOperationExpr{
				Op: ast.NE,
				L:  &ast.Primitive{Typ: ast.INT, Value: 1},
				R:  &ast.Primitive{Typ: ast.INT, Value: 2},
			},
		},
		{
			parseStr: "1>2",
			expected: &ast.BinaryOperationExpr{
				Op: ast.GT,
				L:  &ast.Primitive{Typ: ast.INT, Value: 1},
				R:  &ast.Primitive{Typ: ast.INT, Value: 2},
			},
		},
		{
			parseStr: "1>=2",
			expected: &ast.BinaryOperationExpr{
				Op: ast.GTE,
				L:  &ast.Primitive{Typ: ast.INT, Value: 1},
				R:  &ast.Primitive{Typ: ast.INT, Value: 2},
			},
		},
		{
			parseStr: "1<2",
			expected: &ast.BinaryOperationExpr{
				Op: ast.LT,
				L:  &ast.Primitive{Typ: ast.INT, Value: 1},
				R:  &ast.Primitive{Typ: ast.INT, Value: 2},
			},
		},
		{
			parseStr: "1<=2",
			expected: &ast.BinaryOperationExpr{
				Op: ast.LTE,
				L:  &ast.Primitive{Typ: ast.INT, Value: 1},
				R:  &ast.Primitive{Typ: ast.INT, Value: 2},
			},
		},
		/* Logical AND / OR */
		{
			parseStr: "1 && 2",
			expected: &ast.BinaryOperationExpr{
				Op: ast.AND_AND,
				L:  &ast.Primitive{Typ: ast.INT, Value: 1},
				R:  &ast.Primitive{Typ: ast.INT, Value: 2},
			},
		},
		{
			parseStr: "1 || 2",
			expected: &ast.BinaryOperationExpr{
				Op: ast.OR_OR,
				L:  &ast.Primitive{Typ: ast.INT, Value: 1},
				R:  &ast.Primitive{Typ: ast.INT, Value: 2},
			},
		},
		/* NULL coalescing  */
		{
			parseStr: "1 ?? 2",
			expected: &ast.BinaryOperationExpr{
				Op: ast.NULL_COALESCENCE,
				L:  &ast.Primitive{Typ: ast.INT, Value: 1},
				R:  &ast.Primitive{Typ: ast.INT, Value: 2},
			},
		},
		/* Regex Operations */
		{
			parseStr: " \"adam[23]\" =~ `^[a-z]+\\[[0-9]+\\]$` ",
			expected: &ast.RegexOperationExpr{
				Typ:     ast.RE,
				Target:  "adam[23]",
				Pattern: regexp.MustCompile(`^[a-z]+\[[0-9]+\]$`),
			},
		},
		{
			parseStr: " \"adam[23]\" !~ `^[a-z]+\\[[0-9]+\\]$` ",
			expected: &ast.RegexOperationExpr{
				Typ:     ast.NRE,
				Target:  "adam[23]",
				Pattern: regexp.MustCompile(`^[a-z]+\[[0-9]+\]$`),
			},
		},
		/* Between Operations */
		{
			parseStr: "1 in [1,2,3]",
			expected: &ast.BinaryOperationExpr{
				Op: ast.BETWEEN,
				L:  &ast.Primitive{Typ: ast.INT, Value: 1},
				R: &ast.Primitive{Typ: ast.TUPLE, Value: []*ast.Primitive{
					{Typ: ast.INT, Value: 1},
					{Typ: ast.INT, Value: 2},
					{Typ: ast.INT, Value: 3},
				}},
			},
		},
	}
	runTests(sets, t)
}

func TestComplexExprs(t *testing.T) {
	sets := []TestSet{
		{
			parseStr: "r_obj == t_obj && r_act == t_act",
			expected: &ast.BinaryOperationExpr{
				Op: ast.AND_AND,
				L: &ast.BinaryOperationExpr{
					Op: ast.EQ,
					L:  &ast.Primitive{Typ: ast.VARIABLE, Value: "r_obj"},
					R:  &ast.Primitive{Typ: ast.VARIABLE, Value: "t_obj"},
				},
				R: &ast.BinaryOperationExpr{
					Op: ast.EQ,
					L:  &ast.Primitive{Typ: ast.VARIABLE, Value: "r_act"},
					R:  &ast.Primitive{Typ: ast.VARIABLE, Value: "t_act"},
				},
			},
		},
		{
			parseStr: "r_obj == t_obj && r_act == t_act && r_sub == t_sub",
			expected: &ast.BinaryOperationExpr{
				Op: ast.AND_AND,
				L: &ast.BinaryOperationExpr{
					Op: ast.AND_AND,
					L: &ast.BinaryOperationExpr{
						Op: ast.EQ,
						L:  &ast.Primitive{Typ: ast.VARIABLE, Value: "r_obj"},
						R:  &ast.Primitive{Typ: ast.VARIABLE, Value: "t_obj"},
					},
					R: &ast.BinaryOperationExpr{
						Op: ast.EQ,
						L:  &ast.Primitive{Typ: ast.VARIABLE, Value: "r_act"},
						R:  &ast.Primitive{Typ: ast.VARIABLE, Value: "t_act"},
					},
				},
				R: &ast.BinaryOperationExpr{
					Op: ast.EQ,
					L:  &ast.Primitive{Typ: ast.VARIABLE, Value: "r_sub"},
					R:  &ast.Primitive{Typ: ast.VARIABLE, Value: "t_sub"},
				},
			},
		},
		{
			parseStr: "r_obj == t_obj && keyMatch(r_sub,t_sub)",
			expected: &ast.BinaryOperationExpr{
				Op: ast.AND_AND,
				L: &ast.BinaryOperationExpr{
					Op: ast.EQ,
					L:  &ast.Primitive{Typ: ast.VARIABLE, Value: "r_obj"},
					R:  &ast.Primitive{Typ: ast.VARIABLE, Value: "t_obj"},
				},
				R: &ast.ScalarFunction{
					Ident: "keyMatch",
					Args:  []*ast.Primitive{{Typ: ast.VARIABLE, Value: "r_sub"}, {Typ: ast.VARIABLE, Value: "t_sub"}},
				},
			},
		},
		{
			parseStr: "r_obj == t_obj && foo()",
			expected: &ast.BinaryOperationExpr{
				Op: ast.AND_AND,
				L: &ast.BinaryOperationExpr{
					Op: ast.EQ,
					L:  &ast.Primitive{Typ: ast.VARIABLE, Value: "r_obj"},
					R:  &ast.Primitive{Typ: ast.VARIABLE, Value: "t_obj"},
				},
				R: &ast.ScalarFunction{
					Ident: "foo",
				},
			},
		},
		{
			parseStr: "g(r_sub, p_sub) && r_obj == p_obj && r_act == p_act",
			expected: &ast.BinaryOperationExpr{
				Op: ast.AND_AND,
				L: &ast.BinaryOperationExpr{
					Op: ast.AND_AND,
					L: &ast.ScalarFunction{
						Ident: "g",
						Args:  []*ast.Primitive{{Typ: ast.VARIABLE, Value: "r_sub"}, {Typ: ast.VARIABLE, Value: "p_sub"}},
					},
					R: &ast.BinaryOperationExpr{
						Op: ast.EQ,
						L:  &ast.Primitive{Typ: ast.VARIABLE, Value: "r_obj"},
						R:  &ast.Primitive{Typ: ast.VARIABLE, Value: "p_obj"},
					},
				},
				R: &ast.BinaryOperationExpr{
					Op: ast.EQ,
					L:  &ast.Primitive{Typ: ast.VARIABLE, Value: "r_act"},
					R:  &ast.Primitive{Typ: ast.VARIABLE, Value: "p_act"},
				},
			},
		},
	}
	runTests(sets, t)
}
