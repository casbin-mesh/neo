package parser

import (
	"encoding/json"
	"fmt"
	"github.com/casbin-mesh/neo/pkg/expression/ast"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func newLexer(token string) *Lexer {
	return NewLexer(strings.NewReader(token))
}

func TestParse(t *testing.T) {
	s := NewLexer(strings.NewReader("1+3"))
	yyParse(s)

	result, _ := json.Marshal(s.parseResult)
	fmt.Printf("%s\n", result)
}

type TestSet struct {
	parseStr string
	expected interface{}
	err      error
}

func asserter(t *testing.T, expected, actual interface{}) {
	actualJSON, err := json.Marshal(actual)
	if err != nil {
		fmt.Println(err)
	}
	expectedJSON, err := json.Marshal(expected)
	if err != nil {
		fmt.Println(err)
	}
	assert.Equalf(t, expected, actual, "expected %s, but got %s\n", expectedJSON, actualJSON)
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
			expected: &ast.UnaryOperationExpr{Op: ast.UMINUS, Child: &ast.Primitive{Typ: ast.INT, Value: 1}},
		},
		{
			parseStr: "1.00",
			expected: &ast.Primitive{Typ: ast.FLOAT, Value: float64(1)},
		},
		{
			parseStr: "-1.00",
			expected: &ast.UnaryOperationExpr{Op: ast.UMINUS, Child: &ast.Primitive{Typ: ast.FLOAT, Value: 1.0}},
		},
		{
			parseStr: "true",
			expected: &ast.Primitive{Typ: ast.BOOLEAN, Value: true},
		},
		{
			parseStr: "!true",
			expected: &ast.UnaryOperationExpr{Op: ast.UNOT, Child: &ast.Primitive{Typ: ast.BOOLEAN, Value: true}},
		},
		{
			parseStr: "false",
			expected: &ast.Primitive{Typ: ast.BOOLEAN, Value: false},
		},
		{
			parseStr: "TRUE",
			expected: &ast.Primitive{Typ: ast.BOOLEAN, Value: true},
		},
		{
			parseStr: "FALSE",
			expected: &ast.Primitive{Typ: ast.BOOLEAN, Value: false},
		},
		{
			parseStr: "imVariable",
			expected: &ast.Primitive{Typ: ast.IDENTIFIER, Value: "imVariable"},
		},
		{
			parseStr: "(1,2,\"string\")",
			expected: &ast.Primitive{Typ: ast.TUPLE, Value: []ast.Evaluable{
				&ast.Primitive{Typ: ast.INT, Value: 1},
				&ast.Primitive{Typ: ast.INT, Value: 2},
				&ast.Primitive{Typ: ast.STRING, Value: "string"},
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
				Op: ast.EQ_OP,
				L:  &ast.Primitive{Typ: ast.INT, Value: 1},
				R:  &ast.Primitive{Typ: ast.INT, Value: 2},
			},
		},
		{
			parseStr: "1!=2",
			expected: &ast.BinaryOperationExpr{
				Op: ast.NE_OP,
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
				Op: ast.GE,
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
				Op: ast.LE,
				L:  &ast.Primitive{Typ: ast.INT, Value: 1},
				R:  &ast.Primitive{Typ: ast.INT, Value: 2},
			},
		},
		/* Logical AND / OR */
		{
			parseStr: "1 && 2",
			expected: &ast.BinaryOperationExpr{
				Op: ast.AND_OP,
				L:  &ast.Primitive{Typ: ast.INT, Value: 1},
				R:  &ast.Primitive{Typ: ast.INT, Value: 2},
			},
		},
		{
			parseStr: "1 || 2",
			expected: &ast.BinaryOperationExpr{
				Op: ast.OR_OP,
				L:  &ast.Primitive{Typ: ast.INT, Value: 1},
				R:  &ast.Primitive{Typ: ast.INT, Value: 2},
			},
		},
		/* NULL coalescing  */
		{
			parseStr: "1 ?? 2",
			expected: &ast.BinaryOperationExpr{
				Op: ast.NULL_OP,
				L:  &ast.Primitive{Typ: ast.INT, Value: 1},
				R:  &ast.Primitive{Typ: ast.INT, Value: 2},
			},
		},
		/* Regex Operations */
		{
			parseStr: " \"adam[23]\" =~ \"^[a-z]+\\[[0-9]+\\]$\" ",
			expected: &ast.BinaryOperationExpr{
				Op: ast.RE_OP,
				L:  &ast.Primitive{Typ: ast.STRING, Value: "adam[23]"},
				R:  &ast.Primitive{Typ: ast.STRING, Value: "^[a-z]+\\[[0-9]+\\]$"},
			},
		},
		{
			parseStr: " \"adam[23]\" =~ '^[a-z]+\\[[0-9]+\\]$' ",
			expected: &ast.BinaryOperationExpr{
				Op: ast.RE_OP,
				L:  &ast.Primitive{Typ: ast.STRING, Value: "adam[23]"},
				R:  &ast.Primitive{Typ: ast.STRING, Value: "^[a-z]+\\[[0-9]+\\]$"},
			},
		},
		{
			parseStr: " \"adam[23]\" =~ `^[a-z]+\\[[0-9]+\\]$` ",
			expected: &ast.BinaryOperationExpr{
				Op: ast.RE_OP,
				L:  &ast.Primitive{Typ: ast.STRING, Value: "adam[23]"},
				R:  &ast.Primitive{Typ: ast.STRING, Value: "^[a-z]+\\[[0-9]+\\]$"},
			},
		},
		/* Between Operations */
		{
			parseStr: "1 in (1,2,3)",
			expected: &ast.BinaryOperationExpr{
				Op: ast.IN_OP,
				L:  &ast.Primitive{Typ: ast.INT, Value: 1},
				R: &ast.Primitive{Typ: ast.TUPLE, Value: []ast.Evaluable{
					&ast.Primitive{Typ: ast.INT, Value: 1},
					&ast.Primitive{Typ: ast.INT, Value: 2},
					&ast.Primitive{Typ: ast.INT, Value: 3},
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
				Op: ast.AND_OP,
				L: &ast.BinaryOperationExpr{
					Op: ast.EQ_OP,
					L:  &ast.Primitive{Typ: ast.IDENTIFIER, Value: "r_obj"},
					R:  &ast.Primitive{Typ: ast.IDENTIFIER, Value: "t_obj"},
				},
				R: &ast.BinaryOperationExpr{
					Op: ast.EQ_OP,
					L:  &ast.Primitive{Typ: ast.IDENTIFIER, Value: "r_act"},
					R:  &ast.Primitive{Typ: ast.IDENTIFIER, Value: "t_act"},
				},
			},
		},
		{
			parseStr: "r_obj == t_obj && r_act == t_act && r_sub == t_sub",
			expected: &ast.BinaryOperationExpr{
				Op: ast.AND_OP,
				L: &ast.BinaryOperationExpr{
					Op: ast.AND_OP,
					L: &ast.BinaryOperationExpr{
						Op: ast.EQ_OP,
						L:  &ast.Primitive{Typ: ast.IDENTIFIER, Value: "r_obj"},
						R:  &ast.Primitive{Typ: ast.IDENTIFIER, Value: "t_obj"},
					},
					R: &ast.BinaryOperationExpr{
						Op: ast.EQ_OP,
						L:  &ast.Primitive{Typ: ast.IDENTIFIER, Value: "r_act"},
						R:  &ast.Primitive{Typ: ast.IDENTIFIER, Value: "t_act"},
					},
				},
				R: &ast.BinaryOperationExpr{
					Op: ast.EQ_OP,
					L:  &ast.Primitive{Typ: ast.IDENTIFIER, Value: "r_sub"},
					R:  &ast.Primitive{Typ: ast.IDENTIFIER, Value: "t_sub"},
				},
			},
		},
		{
			parseStr: "r_obj == t_obj && keyMatch(r_sub,t_sub)",
			expected: &ast.BinaryOperationExpr{
				Op: ast.AND_OP,
				L: &ast.BinaryOperationExpr{
					Op: ast.EQ_OP,
					L:  &ast.Primitive{Typ: ast.IDENTIFIER, Value: "r_obj"},
					R:  &ast.Primitive{Typ: ast.IDENTIFIER, Value: "t_obj"},
				},
				R: &ast.ScalarFunction{
					Ident: &ast.Primitive{Typ: ast.IDENTIFIER, Value: "keyMatch"},
					Args:  []ast.Evaluable{&ast.Primitive{Typ: ast.IDENTIFIER, Value: "r_sub"}, &ast.Primitive{Typ: ast.IDENTIFIER, Value: "t_sub"}},
				},
			},
		},
		{
			parseStr: "r_obj == t_obj && foo()",
			expected: &ast.BinaryOperationExpr{
				Op: ast.AND_OP,
				L: &ast.BinaryOperationExpr{
					Op: ast.EQ_OP,
					L:  &ast.Primitive{Typ: ast.IDENTIFIER, Value: "r_obj"},
					R:  &ast.Primitive{Typ: ast.IDENTIFIER, Value: "t_obj"},
				},
				R: &ast.ScalarFunction{
					Ident: &ast.Primitive{Typ: ast.IDENTIFIER, Value: "foo"},
				},
			},
		},
		{
			parseStr: "g(r_sub, p_sub) && r_obj == p_obj && r_act == p_act",
			expected: &ast.BinaryOperationExpr{
				Op: ast.AND_OP,
				L: &ast.BinaryOperationExpr{
					Op: ast.AND_OP,
					L: &ast.ScalarFunction{
						Ident: &ast.Primitive{Typ: ast.IDENTIFIER, Value: "g"},
						Args:  []ast.Evaluable{&ast.Primitive{Typ: ast.IDENTIFIER, Value: "r_sub"}, &ast.Primitive{Typ: ast.IDENTIFIER, Value: "p_sub"}},
					},
					R: &ast.BinaryOperationExpr{
						Op: ast.EQ_OP,
						L:  &ast.Primitive{Typ: ast.IDENTIFIER, Value: "r_obj"},
						R:  &ast.Primitive{Typ: ast.IDENTIFIER, Value: "p_obj"},
					},
				},
				R: &ast.BinaryOperationExpr{
					Op: ast.EQ_OP,
					L:  &ast.Primitive{Typ: ast.IDENTIFIER, Value: "r_act"},
					R:  &ast.Primitive{Typ: ast.IDENTIFIER, Value: "p_act"},
				},
			},
		},
		{
			parseStr: "!g(r_sub, p_sub)",
			expected: &ast.UnaryOperationExpr{
				Op: ast.UNOT,
				Child: &ast.ScalarFunction{
					Ident: &ast.Primitive{Typ: ast.IDENTIFIER, Value: "g"},
					Args:  []ast.Evaluable{&ast.Primitive{Typ: ast.IDENTIFIER, Value: "r_sub"}, &ast.Primitive{Typ: ast.IDENTIFIER, Value: "p_sub"}},
				},
			},
		},
	}
	runTests(sets, t)
}
