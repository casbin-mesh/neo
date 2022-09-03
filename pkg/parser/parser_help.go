package parser

import (
	"github.com/casbin-mesh/neo/pkg/expression/ast"
	"strings"
)

func Parse(lexer yyLexer) {
	yyParse(lexer)
}

func GetParseResult(lexer *Lexer) interface{} {
	return lexer.parseResult
}

func ParseFormLexer(lexer yyLexer) yyLexer {
	yyParse(lexer)
	return lexer
}

func ParseFormString(s string) interface{} {
	l := NewLexer(strings.NewReader(s))
	yyParse(l)
	return GetParseResult(l)
}

func MustParseFormString(s string) ast.Evaluable {
	l := NewLexer(strings.NewReader(s))
	yyParse(l)
	return GetParseResult(l).(ast.Evaluable)
}
