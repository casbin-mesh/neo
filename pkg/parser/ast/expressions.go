package ast

type Type int

const (
	INT Type = iota + 1
	FLOAT64
	BOOL
	STRING
	VARIABLE
	TUPLE
)

type Op int

const (
	ADD Op = iota + 1
	SUB
	MUL
	DIV
	MOD

	SEPARATOR

	NULL_COALESCENCE
	TERNARY_FALSE
	TERNARY_TRUE

	POW
	LSHIFT
	RSHIFT

	OR
	AND
	XOR
	NOT
	GT
	GTE
	LT
	LTE
	EQ
	NE
	RE
	NRE
	BETWEEN
	BOOL_NOT
	AND_AND
	OR_OR
)

// BinaryOperationExpr is for binary operation like `1 + 1`, `1 - 1`, etc.
type BinaryOperationExpr struct {
	// Op is the operator code for BinaryOperation.
	Op
	// L is the left expression in BinaryOperation.
	L interface{}
	// R is the right expression in BinaryOperation.
	R interface{}
}

type RegexOperationExpr struct {
	Typ     Op
	Target  string
	Pattern string
}

type ScalarFunction struct {
	Ident string
	Args  []*Primitive
}

type UnaryOperationExpr struct {
	Child interface{}
	Op
}

type TernaryOperationExpr struct {
	Cond  interface{}
	True  interface{}
	False interface{}
}

type Primitive struct {
	Typ   Type
	Value interface{}
	Text  string
}
