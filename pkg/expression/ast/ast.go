package ast

type EvaluateCtx interface {
	GetParameters() Parameters
	GetFunctions() Functions
}

type Evaluable interface {
	Evaluate(ctx EvaluateCtx) (*Primitive, error)
}
