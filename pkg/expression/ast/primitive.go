package ast

type Error struct {
	error
}

func (e *Error) Evaluate(ctx EvaluateCtx) (*Primitive, error) {
	return nil, e.error
}

type Accessor struct {
	Typ      Type
	Ancestor Evaluable
	Ident    Evaluable
}

func (a Accessor) Evaluate(ctx EvaluateCtx) (*Primitive, error) {
	//TODO implement me
	panic("implement me")
}

type ScalarFunction struct {
	Ident Evaluable
	Args  []Evaluable
}

func (s ScalarFunction) Evaluate(ctx EvaluateCtx) (*Primitive, error) {
	//TODO implement me
	panic("implement me")
}

type Primitive struct {
	Typ   Type
	Value interface{}
}

func (e *Primitive) Evaluate(ctx EvaluateCtx) (*Primitive, error) {
	return nil, nil
}
