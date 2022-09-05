package planner

import "github.com/casbin-mesh/neo/pkg/expression/ast"

type PredicateGroupType uint8

const (
	AndPredicateGroup PredicateGroupType = iota + 1
	OrPredicateGroup
)

type PredicateGroup struct {
	Op   PredicateGroupType
	Cond []ast.Evaluable
}
