package iterator

import "github.com/casbin-mesh/neo/pkg/expression/ast"

type Iterator interface {
	Next() ast.Evaluable
	HasNext() bool
}

type Stack struct {
	node  ast.Evaluable
	index int
}
