package ast

type Op int

var OpToString = []string{
	"PRE_INC_OP", "PRE_DEC_OP", "POST_INC_OP", "POST_DEC_OP",
	"AND", "MUL", "ADD", "SUB", "POW", "DIV", "MOD",
	"LEFT_OP", "RIGHT_OP",
	"IN_OP", "LT", "GT", "LE", "GE",
	"NULL_OP", "EQ_OP", "NE_OP", "RE_OP", "NR_OP", "EX_OR", "IN_OR",
	"AND_OP", "OR_OP",
	"UAND", "UMUL", "UPLUS", "UMINUS", "UBITNOT", "UNOT",
}

const (
	PRE_INC_OP Op = iota + 1
	PRE_DEC_OP
	POST_INC_OP
	POST_DEC_OP

	AND
	MUL
	ADD
	SUB
	POW
	DIV
	MOD

	LEFT_OP
	RIGHT_OP

	IN_OP
	LT
	GT
	LE
	GE

	NULL_OP
	EQ_OP
	NE_OP
	RE_OP
	NR_OP
	EX_OR
	IN_OR

	AND_OP
	OR_OP

	UAND
	UMUL
	UPLUS
	UMINUS
	UBITNOT
	UNOT
)

func (op *Op) String() string {
	return OpToString[*op]
}
