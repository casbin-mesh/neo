package executor

import (
	"context"
	"errors"
	"fmt"
	"github.com/casbin-mesh/neo/pkg/expression/ast"
	"github.com/casbin-mesh/neo/pkg/neo/executor/plan"
	"github.com/casbin-mesh/neo/pkg/neo/session"
	"github.com/casbin-mesh/neo/pkg/primitive"
	"github.com/casbin-mesh/neo/pkg/primitive/btuple"
)

type constExecutor struct {
	baseExecutor
	constPlan plan.ConstPlan
	done      bool
}

func (c constExecutor) Init() {
}

var (
	ErrUnknownEvaluationResult = errors.New("unknown evaluation value")
)

func (c *constExecutor) Next(ctx context.Context, tuple *btuple.Modifier, rid *primitive.ObjectID) (bool, error) {
	if c.done {
		return false, nil
	}
	predicate := c.constPlan.Predicate()
	value, err := predicate.Evaluate(c.GetSessionCtx(), c.constPlan.GetEvalCtx(), nil, nil)
	if err != nil {
		return false, err
	}
	result, ok := value.(*ast.Primitive)
	if ok {
		if result.Typ != ast.BOOLEAN {
			return false, fmt.Errorf("expected BOOLEAN type, but got %s", result.Typ.String())
		}
		if result.Value.(bool) {
			*tuple = btuple.NewModifier([]btuple.Elem{{1}})

		} else {
			*tuple = btuple.NewModifier([]btuple.Elem{{0}})
		}
		c.done = true
		return true, nil
	} else {
		return false, ErrUnknownEvaluationResult
	}
}

func NewConstExecutor(ctx session.Context, constPlan plan.ConstPlan) Executor {
	return &constExecutor{
		baseExecutor: newBaseExecutor(ctx),
		constPlan:    constPlan,
	}
}
