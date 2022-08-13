%{
package parser

import (
    "regexp"
    "github.com/casbin-mesh/neo/pkg/parser/ast"
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
  op            ast.Op
  ex            ast.Evaluable
  t             []*ast.Primitive
  p             *ast.Primitive
}

%token <op> OB CB BETWEEN AND_AND OR_OR OP CP ADD SUB MUL MOD DIV AND OR XOR RSHIFT LSHIFT SEPARATOR NULL_COALESCENCE POW BOOL_NOT UMINUS "-"
%token <i> INT
%token <f> FLOAT
%token <b> BOOL
%token <s> STRING IDENT

%type <ex> expr cond_expr
%type <p> primtive tuple_primtive
%type <t> tuple

/* http://www.eecs.northwestern.edu/~wkliao/op-prec.htm */
%left  SEPARATOR /* Lowest precedence */
%right TERNARY_TRUE TERNARY_FALSE
%left NULL_COALESCENCE  /* https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators/Operator_Precedence */
%left OR_OR
%left AND_AND
%left BIT_OR
%left BIT_XOR
%left BIT_AND
%left EQ NE RE NRE
%left GT GTE LT LTE BETWEEN
%left LSHIFT RSHIFT
%left ADD SUB
%left DIV MUL MOD
%right POW
%nonassoc UMINUS BIT_NOT BOOL_NOT
%left OP

%%

line
    : expr                   { setScannerData(yylex, $1) }
;

expr
    : cond_expr                 { $$ = $1 }
    | expr ADD expr          { $$ = &ast.BinaryOperationExpr{ Op: ast.ADD, L:$1, R:$3 } }
    | expr SUB expr          { $$ = &ast.BinaryOperationExpr{ Op: ast.SUB, L:$1, R:$3 } }
    | expr MUL expr          { $$ = &ast.BinaryOperationExpr{ Op: ast.MUL, L:$1, R:$3 } }
    | expr DIV expr          { $$ = &ast.BinaryOperationExpr{ Op: ast.DIV, L:$1, R:$3 } }
    | expr MOD expr          { $$ = &ast.BinaryOperationExpr{ Op: ast.MOD, L:$1, R:$3 } }
    | expr POW expr          { $$ = &ast.BinaryOperationExpr{ Op: ast.POW, L:$1, R:$3 } }
    | BOOL_NOT OP cond_expr CP  { $$ = &ast.UnaryOperationExpr{ Op: ast.BOOL_NOT, Child:$3  }  }
    | SUB OP expr CP %prec UMINUS  { $$ = &ast.UnaryOperationExpr{ Op: ast.UMINUS, Child:$3 } }
    | OP expr CP             { $$ = $2 }
    | primtive               { $$ = $1 }
    | primtive BETWEEN tuple_primtive { $$ = &ast.BinaryOperationExpr{ Op:ast.BETWEEN,L:$1,R:$3 } }
    | IDENT OP tuple CP      { $$ = &ast.ScalarFunction{ Ident:$1, Args: $3 } }
    | expr TERNARY_TRUE expr TERNARY_FALSE expr { $$ = &ast.TernaryOperationExpr{ Cond:$1, True:$3, False: $5 } }
;

cond_expr
    : expr EQ expr           { $$ = &ast.BinaryOperationExpr{ Op: ast.EQ, L:$1, R:$3 } }
    | expr NE expr           { $$ = &ast.BinaryOperationExpr{ Op: ast.NE, L:$1, R:$3 } }
    | expr GT expr           { $$ = &ast.BinaryOperationExpr{ Op: ast.GT, L:$1, R:$3 } }
    | expr GTE expr          { $$ = &ast.BinaryOperationExpr{ Op: ast.GTE, L:$1, R:$3 } }
    | expr LT expr           { $$ = &ast.BinaryOperationExpr{ Op: ast.LT, L:$1, R:$3 } }
    | expr LTE expr          { $$ = &ast.BinaryOperationExpr{ Op: ast.LTE, L:$1, R:$3 } }
    | expr AND_AND expr      { $$ = &ast.BinaryOperationExpr{ Op: ast.AND_AND, L:$1, R:$3 } }
    | expr OR_OR expr        { $$ = &ast.BinaryOperationExpr{ Op: ast.OR_OR, L:$1, R:$3 } }
    | expr NULL_COALESCENCE expr { $$ = &ast.BinaryOperationExpr{ Op: ast.NULL_COALESCENCE, L:$1, R:$3 } }
    | STRING RE STRING       { $$ = &ast.RegexOperationExpr{ Typ: ast.RE, Pattern: regexp.MustCompile($3), Target: $1 } } /* regex operation */
    | IDENT RE STRING        { $$ = &ast.RegexOperationExpr{ Typ: ast.RE, Pattern: regexp.MustCompile($3), Target: $1 } }
    | STRING NRE STRING      { $$ = &ast.RegexOperationExpr{ Typ: ast.NRE, Pattern: regexp.MustCompile($3),Target: $1 } }
    | IDENT NRE STRING       { $$ = &ast.RegexOperationExpr{ Typ: ast.NRE, Pattern: regexp.MustCompile($3),Target: $1 } }
;

tuple
    : 			          { $$ = nil }
    | primtive                    { $$ = append([]*ast.Primitive{},$1) }
    | tuple SEPARATOR primtive    { $$ = append($1,$3) }
;

tuple_primtive
    : OB tuple CB                 { $$ = &ast.Primitive{ Typ: ast.TUPLE, Value:$2 } }
;

primtive
    : INT                         { $$ = &ast.Primitive{ Typ: ast.INT, Value:$1 } }
    | FLOAT                       { $$ = &ast.Primitive{ Typ: ast.FLOAT64, Value:$1 } }
    | SUB INT %prec UMINUS        { $$ = &ast.Primitive{ Typ: ast.INT, Value:-$2 } }
    | SUB FLOAT %prec UMINUS      { $$ = &ast.Primitive{ Typ: ast.FLOAT64, Value:-$2 } }
    | BOOL                        { $$ = &ast.Primitive{ Typ: ast.BOOL, Value:$1 } }
    | BOOL_NOT BOOL               { $$ = &ast.Primitive{ Typ: ast.BOOL, Value:!$2 } }
    | STRING                      { $$ = &ast.Primitive{ Typ: ast.STRING, Value:$1 } }
    | IDENT                       { $$ = &ast.Primitive{ Typ: ast.VARIABLE, Value:$1 } }
    | tuple_primtive              { $$ = $1 }
;

%%