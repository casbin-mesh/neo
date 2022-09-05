package expression

import (
	"errors"
	"fmt"
	"github.com/casbin-mesh/neo/pkg/expression/ast"
	"github.com/casbin-mesh/neo/pkg/neo/session"
	"github.com/casbin-mesh/neo/pkg/primitive/bschema"
	"github.com/casbin-mesh/neo/pkg/primitive/btuple"
)

type Value interface{}

type AbstractExpression interface {
	Evaluate(ctx session.Context, tuple btuple.Reader, schema bschema.Reader) Value
}

type MockExpr struct {
	Expr func(ctx session.Context, tuple btuple.Reader, schema bschema.Reader) Value
}

func (m MockExpr) Evaluate(ctx session.Context, tuple btuple.Reader, schema bschema.Reader) Value {
	return m.Expr(ctx, tuple, schema)
}

var (
	ErrUnknownEvaluationResult = errors.New("unknown evaluation result")
)

func TryGetBool(value Value) (bool, error) {
	if v, ok := value.(*ast.Primitive); ok {
		if v.Typ != ast.BOOLEAN {
			return false, fmt.Errorf("expected BOOLEAN type, but got %s", v.Typ.String())
		}
		return v.Value.(bool), nil
	}
	return false, ErrUnknownEvaluationResult
}
