package parser

import (
	"fmt"
	"github.com/casbin-mesh/neo/pkg/parser/ast"
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
	fmt.Printf("%v\n", s.parseResult)
}

type TestSet struct {
	parseStr string
	expected interface{}
	err      error
}

func runTests(sets []TestSet, t *testing.T) {
	for _, set := range sets {
		s := newLexer(set.parseStr)
		yyParse(s)
		assert.Equal(t, set.expected, s.parseResult)
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
