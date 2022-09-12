package planner

import (
	"github.com/casbin-mesh/neo/pkg/expression/ast"
	"github.com/casbin-mesh/neo/pkg/neo/executor/plan"
	"github.com/casbin-mesh/neo/pkg/neo/model"
	"github.com/casbin-mesh/neo/pkg/neo/optimizer"
	"github.com/casbin-mesh/neo/pkg/neo/optimizer/heuristic"
	"github.com/casbin-mesh/neo/pkg/neo/optimizer/node"
	"sort"
	"strings"
)

type TableStatic interface {
	// GetColEstimatedCardinality returns count / the number of distinct value
	GetColEstimatedCardinality(col string) int64
	// GetColCardinality returns static collected by CM sketch
	GetColCardinality(col string, value string) int64
	// GetCount returns total row number
	GetCount(col string) int64
}

type CostModel interface {
	GetTable(name string) TableStatic
}

type Ctx interface {
	PolicyTableName() string
	EffectColName() string
	AllowIdent() string
	DenyIdent() string
	ReqAccessorAncestorName() string
}

func GeneratePlans(ctx Ctx, tree ast.Evaluable, matcher *model.MatcherInfo) {
	subplan := model.GenerateEffectPolicyAst(ctx.PolicyTableName(), ctx.EffectColName(), ctx.AllowIdent(), ctx.DenyIdent(), matcher.EffectPolicy)
	logicalOptimizer := optimizer.LogicalOptimizer{}
	predicate := logicalOptimizer.Optimize(tree)
	predicates := make([]node.Predicate, 0, len(subplan))
	for _, evaluable := range subplan {
		p := predicate.Clone()
		heuristic.AppendAst2Predicate(&p, evaluable, func(node ast.Evaluable) bool {
			return IsConstNode(ctx, node)
		})
		predicates = append(predicates, p)
	}
}

func GeneratePhysicalPlans(ctx Ctx, ast ast.Evaluable, db *model.DBInfo) (plans []plan.AbstractPlan, err error) {
	tableInfo, err := db.TableByLName(ctx.PolicyTableName())
	if err != nil {
		return nil, err
	}
	if len(tableInfo.Indices) > 0 {
		//for _, index := range tableInfo.Indices {
		//
		//}
	}
	return nil, nil
}

func IsBinaryOperationExpr(node ast.Evaluable) (ast.Evaluable, bool) {
	n, ok := node.(*ast.BinaryOperationExpr)
	return n, ok
}

func IsReqAccessor(node ast.Evaluable, reqAccessorAncestorName string) bool {
	n, ok := node.(*ast.Accessor)
	if ok {
		if p, ok := n.Ancestor.(*ast.Primitive); ok && p.Typ == ast.IDENTIFIER {
			return p.Value.(string) == reqAccessorAncestorName
		}
	}
	return false
}

func IsConstNode(ctx Ctx, node ast.Evaluable) (ok bool) {
	switch n := node.(type) {
	case *ast.Primitive:
		switch n.Typ {
		case ast.INT, ast.FLOAT, ast.STRING, ast.NULL, ast.BOOLEAN:
			return true
		case ast.TUPLE:
			for i := 0; i < n.ChildrenLen(); i++ {
				if !IsConstNode(ctx, n.GetChildAt(i)) {
					return false
				}
			}
			return true

		case ast.MEMBER_ACCESSOR:
			return IsReqAccessor(node, ctx.ReqAccessorAncestorName())
		}
	case *ast.Accessor:
		return IsReqAccessor(node, ctx.ReqAccessorAncestorName())
	case *ast.UnaryOperationExpr:
		if !IsConstNode(ctx, n.Child) {
			return
		}
		return true
	case *ast.BinaryOperationExpr:
		if !IsConstNode(ctx, n.L) || !IsConstNode(ctx, n.R) {
			return
		}
		return true
	case *ast.TernaryOperationExpr:
		if !IsConstNode(ctx, n.Cond) || !IsConstNode(ctx, n.True) || !IsConstNode(ctx, n.False) {
			return
		}
		return true
	}
	return
}

func sortStrings(target []string) {
	sort.Slice(target, func(i, j int) bool {
		return strings.Compare(target[i], target[j]) < 0
	})
}

type Property struct {
	indexes       []*model.IndexInfo
	secondScan    bool
	fullTableScan bool
	constConds    []ast.Evaluable
}

type physicalPlan struct {
	plan.MatcherPlan
	Property
	Children []physicalPlan
}
