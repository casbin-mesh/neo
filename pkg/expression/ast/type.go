package ast

type Type int

const (
	BOOLEAN Type = iota + 1
	INT
	FLOAT
	STRING
	TUPLE
	ERROR
	OBJ_ACCESSOR
	METHOD_ACCESSOR
	IDENTIFIER
)
