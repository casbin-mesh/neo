package parser

import (
	"fmt"
	"github.com/casbin-mesh/neo/pkg/expression/ast"
	"strings"
)

func Parse(lexer yyLexer) {
	yyParse(lexer)
}

func GetParseResult(lexer *Lexer) interface{} {
	return lexer.parseResult
}

func ParseFromLexer(lexer yyLexer) yyLexer {
	yyParse(lexer)
	return lexer
}

func ParseFromString(s string) interface{} {
	l := NewLexer(strings.NewReader(s))
	yyParse(l)
	return GetParseResult(l)
}

func MustParseFromString(s string) ast.Evaluable {
	l := NewLexer(strings.NewReader(s))
	yyParse(l)
	return GetParseResult(l).(ast.Evaluable)
}

func TryParserFromString(s string) (tree ast.Evaluable, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("failed to parse %s: %s", s, r)
		}
	}()
	l := NewLexer(strings.NewReader(s))
	yyParse(l)
	return GetParseResult(l).(ast.Evaluable), nil
}
