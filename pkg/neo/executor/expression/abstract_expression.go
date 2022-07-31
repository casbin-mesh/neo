package expression

import (
	"github.com/casbin-mesh/neo/pkg/neo/session"
	"github.com/casbin-mesh/neo/pkg/primitive/bschema"
	"github.com/casbin-mesh/neo/pkg/primitive/btuple"
)

type Value interface{}

type AbstractExpression interface {
	Evaluate(ctx session.Context, tuple btuple.Reader, schema bschema.Reader) Value
}
