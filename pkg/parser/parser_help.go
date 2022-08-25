package parser

func Parse(lexer yyLexer) {
	yyParse(lexer)
}

func GetParseResult(lexer *Lexer) interface{} {
	return lexer.parseResult
}
