package iterator

import "github.com/casbin-mesh/neo/pkg/expression/ast"

type Filter = func(node ast.Evaluable) bool

var (
	defaultIncludeAll          = func(n ast.Evaluable) bool { return true }
	defaultOpt                 = &Option{defaultIncludeAll}
	DefaultBinaryTreeTraversal = &Option{
		Filter: func(node ast.Evaluable) bool {
			_, ok := node.(*ast.BinaryOperationExpr)
			return ok
		},
	}
)

type Option struct {
	Filter
}

func getOpt(opts ...*Option) *Option {
	if len(opts) == 0 {
		return defaultOpt
	} else {
		return opts[0]
	}
}
