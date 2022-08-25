package ast

type TernaryOperationExpr struct {
	Cond  Evaluable
	True  Evaluable
	False Evaluable
}

func (e *TernaryOperationExpr) Evaluate(ctx EvaluateCtx) (*Primitive, error) {
	return nil, nil
}

type BinaryOperationExpr struct {
	Op
	L Evaluable
	R Evaluable
}

func (e *BinaryOperationExpr) Evaluate(ctx EvaluateCtx) (*Primitive, error) {
	return nil, nil
}

type UnaryOperationExpr struct {
	Child Evaluable
	Op
}

func (e *UnaryOperationExpr) Evaluate(ctx EvaluateCtx) (*Primitive, error) {
	return nil, nil
}
