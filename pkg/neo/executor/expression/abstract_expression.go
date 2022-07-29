package expression

import (
	"context"
	"github.com/casbin-mesh/neo/pkg/primitive/bschema"
	"github.com/casbin-mesh/neo/pkg/primitive/btuple"
)

type Value interface{}

type AbstractExpression interface {
	Evaluate(ctx context.Context, tuple btuple.Reader, schema bschema.Reader) Value
}
