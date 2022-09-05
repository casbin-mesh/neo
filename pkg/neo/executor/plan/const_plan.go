package plan

import (
	"github.com/casbin-mesh/neo/pkg/expression"
	"github.com/casbin-mesh/neo/pkg/expression/ast"
)

type ConstPlan interface {
	AbstractPlan
	Predicate() expression.Expression
	GetEvalCtx() ast.EvaluateCtx
}

type constPlan struct {
	AbstractPlan
	predicate expression.Expression
	ctx       ast.EvaluateCtx
}

func (c constPlan) GetEvalCtx() ast.EvaluateCtx {
	return c.ctx
}

func (c constPlan) Predicate() expression.Expression {
	return c.predicate
}

func NewConstPlan(predicate expression.Expression, ctx ast.EvaluateCtx) ConstPlan {
	return &constPlan{
		AbstractPlan: NewAbstractPlan(nil, nil),
		predicate:    predicate,
		ctx:          ctx,
	}
}
