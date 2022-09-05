package planner

import (
	"github.com/casbin-mesh/neo/pkg/expression/ast"
)

// SplitConditions returns const expressions and non const expression separately
func SplitConditions(ctx Ctx, conditions []ast.Evaluable) (constExpr, nonConst []ast.Evaluable) {
	isConst := make([]bool, len(conditions))
	count := 0
	for i, condition := range conditions {
		if IsConstNode(ctx, condition) {
			isConst[i] = true
			count++
		}
	}
	if count > 0 {
		constExpr = make([]ast.Evaluable, 0, count)
	}
	if len(conditions)-count > 0 {
		nonConst = make([]ast.Evaluable, 0, len(conditions)-count)
	}
	for i, condition := range conditions {
		if isConst[i] {
			constExpr = append(constExpr, condition)
		} else {
			nonConst = append(nonConst, condition)
		}
	}
	return
}
