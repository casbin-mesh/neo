%{
package parser

import (
	"github.com/casbin-mesh/neo/pkg/expression/ast"
)

func setScannerData(yylex interface{}, data interface{}) {
	yylex.(*Lexer).parseResult = data
}
%}

%union {
  i             int
  f             float64
  b             bool
  s             string
  op		ast.Op
  expr		ast.Evaluable
  exprs 	[]ast.Evaluable
  empty 	struct{}
  error 	*ast.Error
}

// The precedence followings the https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators/Operator_Precedence

%token <empty> INC_OP DEC_OP LEFT_OP RIGHT_OP
%token <empty> LE_OP GE_OP EQ_OP NE_OP RE_OP NR_OP
%token <empty> AND_OP OR_OP NULL_OP
%token <empty> EXP_OP IN_OP

%token <i> INT
%token <f> FLOAT
%token <b> BOOLEAN
%token <s> STRING_LITERAL IDENTIFIER

%type <error> error
%type <op> unary_operator
%type <exprs> argument_expression_list expresions
%type <expr> expression
%type <expr> conditional_expression
%type <expr> logical_or_expression logical_and_expression
%type <expr> inclusive_or_expression exclusive_or_expression and_expression
%type <expr> equality_expression relational_expression
%type <expr> shift_expression
%type <expr> additive_expression multiplicative_expression
%type <expr> exponentiation_expression
%type <expr> unary_expression
%type <expr> postfix_expression
%type <expr> primitive primary_expression

%%
line
    : expression                   { setScannerData(yylex, $1) }
;

primitive
  : BOOLEAN								{ $$ = &ast.Primitive{ Typ:ast.BOOLEAN, Value:$1 } }
  | INT									{ $$ = &ast.Primitive{ Typ:ast.INT, Value:$1 } }
  | FLOAT								{ $$ = &ast.Primitive{ Typ:ast.FLOAT, Value:$1 } }
  | IDENTIFIER								{ $$ = &ast.Primitive{ Typ:ast.IDENTIFIER, Value:$1 } }
  | STRING_LITERAL							{ $$ = &ast.Primitive{ Typ:ast.STRING, Value:$1 } }
;

expresions
    : 									{ $$ = nil }
    | expression							{ $$ = append($$,$1) }
    | expresions ',' expression						{ $$ = append($1,$3) }
;

primary_expression
  : primitive 								{ $$ = $1 }
  | '[' expresions ']'							{ $$ = &ast.Primitive{ Typ:ast.TUPLE, Value:$2 } }
  | '(' expression ')'							{ $$ = $2 }
  | '(' error ')'							{ $$ = &ast.Primitive{ Typ:ast.ERROR, Value:$2 } }
  ;

postfix_expression
  : primary_expression							{ $$ = $1 }
  | postfix_expression '[' expression ']'				{ $$ = &ast.Accessor{ Typ:ast.OBJ_ACCESSOR, Ancestor:$1, Ident:$3 } }
  | postfix_expression '(' argument_expression_list ')'			{ $$ = &ast.ScalarFunction{ Ident: $1,Args:$3 } }
  | postfix_expression '.' IDENTIFIER					{ $$ = &ast.Accessor{ Typ:ast.METHOD_ACCESSOR, Ancestor:$1, Ident:&ast.Primitive{ Typ:ast.STRING, Value:$3 } } }
  | postfix_expression INC_OP						{ $$ = &ast.UnaryOperationExpr{ Op:ast.POST_INC_OP, Child:$1 } }
  | postfix_expression DEC_OP						{ $$ = &ast.UnaryOperationExpr{ Op:ast.POST_DEC_OP, Child:$1 } }
  ;

argument_expression_list
  : 									{ $$ = nil }
  | conditional_expression						{ $$ = append($$,$1) }
  | argument_expression_list ',' conditional_expression			{ $$ = append($1,$3) }
  ;

unary_expression
  : postfix_expression							{ $$ =$1 }
  | INC_OP unary_expression						{ $$ = &ast.UnaryOperationExpr{ Op:ast.PRE_INC_OP, Child:$2 } }
  | DEC_OP unary_expression						{ $$ = &ast.UnaryOperationExpr{ Op:ast.PRE_DEC_OP, Child:$2 } }
  | unary_operator unary_expression					{ $$ = &ast.UnaryOperationExpr{ Op:$1, Child:$2 } }
  ;

unary_operator
  : '&'    { $$ = ast.UAND }
  | '*'    { $$ = ast.UMUL }
  | '+'    { $$ = ast.UPLUS }
  | '-'    { $$ = ast.UMINUS }
  | '~'    { $$ = ast.UBITNOT }
  | '!'    { $$ = ast.UNOT }
  ;

exponentiation_expression
  : unary_expression							{ $$ = $1 }
  | unary_expression EXP_OP exponentiation_expression			{ $$ = &ast.BinaryOperationExpr{ Op:ast.POW, L:$1, R:$3 } }
;

multiplicative_expression
  : exponentiation_expression						{ $$ = $1 }
  | multiplicative_expression '*' exponentiation_expression		{ $$ = &ast.BinaryOperationExpr{ Op:ast.MUL, L:$1, R:$3 } }
  | multiplicative_expression '/' exponentiation_expression		{ $$ = &ast.BinaryOperationExpr{ Op:ast.DIV, L:$1, R:$3 } }
  | multiplicative_expression '%' exponentiation_expression		{ $$ = &ast.BinaryOperationExpr{ Op:ast.MOD, L:$1, R:$3 } }
  | error '*' exponentiation_expression					{ $$ = &ast.BinaryOperationExpr{ Op:ast.MUL, L:$1, R:$3 } }
  | error '/' exponentiation_expression					{ $$ = &ast.BinaryOperationExpr{ Op:ast.DIV, L:$1, R:$3 } }
  | error '%' exponentiation_expression					{ $$ = &ast.BinaryOperationExpr{ Op:ast.MOD, L:$1, R:$3 } }
  ;

additive_expression
  : multiplicative_expression						{ $$ = $1 }
  | additive_expression '+' multiplicative_expression			{ $$ = &ast.BinaryOperationExpr{ Op:ast.ADD, L:$1, R:$3 } }
  | additive_expression '-' multiplicative_expression			{ $$ = &ast.BinaryOperationExpr{ Op:ast.SUB, L:$1, R:$3 } }
  | error '+' multiplicative_expression					{ $$ = &ast.BinaryOperationExpr{ Op:ast.ADD, L:$1, R:$3 } }
  | error '-' multiplicative_expression					{ $$ = &ast.BinaryOperationExpr{ Op:ast.SUB, L:$1, R:$3 } }
  ;

shift_expression
  : additive_expression							{ $$ = $1 }
  | shift_expression LEFT_OP additive_expression			{ $$ = &ast.BinaryOperationExpr{ Op:ast.LEFT_OP, L:$1, R:$3 } }
  | shift_expression RIGHT_OP additive_expression			{ $$ = &ast.BinaryOperationExpr{ Op:ast.RIGHT_OP, L:$1, R:$3 } }
  ;

relational_expression
  : shift_expression							{ $$ = $1 }
  | relational_expression IN_OP shift_expression			{ $$ = &ast.BinaryOperationExpr{ Op:ast.IN_OP, L:$1, R:$3 } }
  | relational_expression '<' shift_expression				{ $$ = &ast.BinaryOperationExpr{ Op:ast.LT, L:$1, R:$3 } }
  | relational_expression '>' shift_expression				{ $$ = &ast.BinaryOperationExpr{ Op:ast.GT, L:$1, R:$3 } }
  | relational_expression LE_OP shift_expression			{ $$ = &ast.BinaryOperationExpr{ Op:ast.LE, L:$1, R:$3 } }
  | relational_expression GE_OP shift_expression			{ $$ = &ast.BinaryOperationExpr{ Op:ast.GE, L:$1, R:$3 } }
  | error IN_OP shift_expression					{ $$ = &ast.BinaryOperationExpr{ Op:ast.IN_OP, L:$1, R:$3 } }
  | error '<' shift_expression						{ $$ = &ast.BinaryOperationExpr{ Op:ast.LT, L:$1, R:$3 } }
  | error '>' shift_expression						{ $$ = &ast.BinaryOperationExpr{ Op:ast.GT, L:$1, R:$3 } }
  | error LE_OP shift_expression					{ $$ = &ast.BinaryOperationExpr{ Op:ast.LE, L:$1, R:$3 } }
  | error GE_OP shift_expression					{ $$ = &ast.BinaryOperationExpr{ Op:ast.GE, L:$1, R:$3 } }
  ;

equality_expression
  : relational_expression						{ $$ = $1 }
  | equality_expression EQ_OP relational_expression			{ $$ = &ast.BinaryOperationExpr{ Op:ast.EQ_OP, L:$1, R:$3 } }
  | equality_expression NE_OP relational_expression			{ $$ = &ast.BinaryOperationExpr{ Op:ast.NE_OP, L:$1, R:$3 } }
  | equality_expression RE_OP relational_expression			{ $$ = &ast.BinaryOperationExpr{ Op:ast.RE_OP, L:$1, R:$3 } }
  | equality_expression NR_OP relational_expression			{ $$ = &ast.BinaryOperationExpr{ Op:ast.NR_OP, L:$1, R:$3 } }
  | error EQ_OP relational_expression					{ $$ = &ast.BinaryOperationExpr{ Op:ast.EQ_OP, L:$1, R:$3 } }
  | error NE_OP relational_expression					{ $$ = &ast.BinaryOperationExpr{ Op:ast.NE_OP, L:$1, R:$3 } }
  | error RE_OP relational_expression					{ $$ = &ast.BinaryOperationExpr{ Op:ast.RE_OP, L:$1, R:$3 } }
  | error NR_OP relational_expression					{ $$ = &ast.BinaryOperationExpr{ Op:ast.NR_OP, L:$1, R:$3 } }
  ;

and_expression
  : equality_expression							{ $$ = $1 }
  | and_expression '&' equality_expression				{ $$ = &ast.BinaryOperationExpr{ Op:ast.AND, L:$1, R:$3 } }
  ;

exclusive_or_expression
  : and_expression							{ $$ = $1 }
  | exclusive_or_expression '^' and_expression				{ $$ = &ast.BinaryOperationExpr{ Op:ast.EX_OR, L:$1, R:$3 } }
  ;

inclusive_or_expression
  : exclusive_or_expression						{ $$ = $1 }
  | inclusive_or_expression '|' exclusive_or_expression			{ $$ = &ast.BinaryOperationExpr{ Op:ast.IN_OR, L:$1, R:$3 } }
  ;

logical_and_expression
  : inclusive_or_expression						{ $$ = $1 }
  | logical_and_expression AND_OP inclusive_or_expression		{ $$ = &ast.BinaryOperationExpr{ Op:ast.AND_OP, L:$1, R:$3 } }
  ;

logical_or_expression
  : logical_and_expression						{ $$ = $1 }
  | logical_or_expression OR_OP logical_and_expression			{ $$ = &ast.BinaryOperationExpr{ Op:ast.OR_OP, L:$1, R:$3 } }
  | logical_and_expression NULL_OP logical_and_expression		{ $$ = &ast.BinaryOperationExpr{ Op:ast.NULL_OP, L:$1, R:$3 } }
  ;

conditional_expression
  : logical_or_expression { $$ = $1 }
  | logical_or_expression '?' expression ':' conditional_expression	{ $$ = &ast.TernaryOperationExpr{ Cond:$1,True:$3,False:$5 } }
  | error '?' error ':' conditional_expression 				{ $$ = &ast.TernaryOperationExpr{ Cond:$1,True:$3,False:$5 } }
  | logical_or_expression '?' error ':' conditional_expression		{ $$ = &ast.TernaryOperationExpr{ Cond:$1,True:$3,False:$5 } }
  | error '?' expression ':' conditional_expression			{ $$ = &ast.TernaryOperationExpr{ Cond:$1,True:$3,False:$5 } }
  ;

expression
  : conditional_expression { $$ = $1 }
  ;

%%