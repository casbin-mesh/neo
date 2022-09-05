package planner

import (
	"github.com/casbin-mesh/neo/pkg/expression"
	"github.com/casbin-mesh/neo/pkg/expression/ast"
	"github.com/casbin-mesh/neo/pkg/neo/executor/plan"
	"github.com/casbin-mesh/neo/pkg/neo/model"
	"sort"
	"strings"
)

type Ctx interface {
	PolicyTableName() string
	EffectColName() string
	AllowIdent() string
	DenyIdent() string
	ReqAccessorAncestorName() string
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

func FlatExprTree(ctx Ctx, root ast.Evaluable) PredicateGroup {
	if n, ok := root.(*ast.BinaryOperationExpr); ok && (n.Op == ast.AND_OP || n.Op == ast.OR_OP) {
		if n.Op == ast.AND_OP {
			return PredicateGroup{
				Op:   AndPredicateGroup,
				Cond: expression.FlatAndSubtree(n),
			}
		} else { // ast.OR_OP
			return PredicateGroup{
				Op:   OrPredicateGroup,
				Cond: expression.FlatOrSubtree(n),
			}
		}
	}
	return PredicateGroup{Op: AndPredicateGroup, Cond: []ast.Evaluable{root}}
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

func MergePolicyEffect(ctx Ctx, predicateGroup PredicateGroup, policyEffect ast.Evaluable) PredicateGroup {
	if predicateGroup.Op == AndPredicateGroup {
		predicateGroup.Cond = append(predicateGroup.Cond, policyEffect)
	} else if predicateGroup.Op == OrPredicateGroup {
		for i, cond := range predicateGroup.Cond {
			if !IsConstNode(ctx, cond) {
				predicateGroup.Cond[i] = expression.ConnectSubtree(cond, policyEffect)
			}
		}
	}
	return predicateGroup
}

//func GeneratePredicates(ctx Ctx, predicate ast.Evaluable, info *model.MatcherInfo) PredicateGroup {
//	switch info.EffectPolicy {
//	// some(where (p.eft == allow))
//	case model.AllowOverride:
//		exprTree := FlatExprTree(ctx, predicate)
//		policyEffect := &ast.BinaryOperationExpr{
//			Op: ast.EQ_OP,
//			L: &ast.Accessor{
//				Typ: ast.MEMBER_ACCESSOR,
//				Ancestor: &ast.Primitive{
//					Typ:   ast.IDENTIFIER,
//					Value: ctx.PolicyTableName()},
//				Ident: &ast.Primitive{
//					Typ:   ast.IDENTIFIER,
//					Value: ctx.EffectColName(),
//				},
//			},
//			// TODO: the type of allow ident?
//			R: &ast.Primitive{Typ: ast.STRING, Value: ctx.AllowIdent()},
//		}
//
//	}
//	return PredicateGroup{}
//}

func GenerateNaivePlan(
	ctx Ctx,
	constEvalCtx ast.EvaluateCtx,
	condType plan.EffectType,
	constConds []ast.Evaluable,
	conds []ast.Evaluable, // un-flatted expressions
	table *model.TableInfo,
	dbId uint64) plan.AbstractPlan {
	children := make([]plan.AbstractPlan, 0, len(conds)+len(constConds))
	for _, cond := range constConds {
		children = append(children, plan.NewConstPlan(expression.NewAbstractExpression(cond), constEvalCtx))
	}
	for _, cond := range conds {
		evalCtx := ast.NewContext()
		expr, accessor := expression.NewExpression(cond)
		evalCtx.AddAccessor(ctx.PolicyTableName(), accessor)
		children = append(children, plan.NewSeqScanPlan(table, expr, evalCtx, table.ID, dbId))
	}
	matcherPlan := plan.NewMatcherPlan(children, condType)
	return matcherPlan
}

func sortStrings(target []string) {
	sort.Slice(target, func(i, j int) bool {
		return strings.Compare(target[i], target[j]) < 0
	})
}

type CircuitType int

const (
	None CircuitType = iota
	AndCircuit
	OrCircuit
)

type Property struct {
	indexes          []*model.IndexInfo
	secondScan       bool
	fullTableScan    bool
	shortCircuitType CircuitType
	constConds       []ast.Evaluable
}

type physicalPlan struct {
	plan.MatcherPlan
	Property
	Children []physicalPlan
}

func GeneratePlanWithIndex(
	ctx Ctx,
	constEvalCtx ast.EvaluateCtx,
	condType plan.EffectType,
	constConds []ast.Evaluable,
	conds []ast.Evaluable, // un-flatted expressions
	table *model.TableInfo,
	dbId uint64) []physicalPlan {

	for i := 0; i < len(table.Indices); i++ {
		children := make([]plan.AbstractPlan, 0, len(conds)+len(constConds))
		for _, cond := range constConds {
			children = append(children, plan.NewConstPlan(expression.NewAbstractExpression(cond), constEvalCtx))
		}
		if len(conds) == 0 {
			return nil
		}
		for _, cond := range conds {
			evalCtx := ast.NewContext()
			expr, accessor := expression.NewExpression(cond)
			pg := FlatExprTree(ctx, cond)

			// if the type is and
			// the most selective index
			//
			evalCtx.AddAccessor(ctx.PolicyTableName(), accessor)
			children = append(children, plan.NewSeqScanPlan(table, expr, evalCtx, table.ID, dbId))
		}
		matcherPlan := plan.NewMatcherPlan(children, condType)
	}
	return nil
}

//func GeneratePlanWithIndex(ctx Ctx, predicate ast.Evaluable, matcher *model.MatcherInfo, index *model.IndexInfo, table *model.TableInfo, dbId uint64) []plan.AbstractPlan {
//	members := expression.GetAccessorMembers(predicate)
//	indexFields := make([]string, 0, len(index.Columns))
//	for _, col := range index.Columns {
//		indexFields = append(indexFields, col.ColName.L)
//	}
//	leftmost := index.Leftmost().ColName.L
//
//	sortStrings(members)
//	sortStrings(indexFields)
//	// intersect
//	intersect := utils.SortedGeneric(members, indexFields)
//
//	// if and only if intersect include the leftmost column name
//	// then we can push down the predicate
//	_, found := slices.BinarySearch(intersect, leftmost)
//	effectPolicyAst := model.GenerateEffectPolicyAst(ctx.PolicyTableName(), ctx.EffectColName(), ctx.AllowIdent(), ctx.DenyIdent(), matcher.EffectPolicy)
//	if found {
//		// generates plans
//		if len(index.Columns) > 1 {
//			// try push all index-covered predicate to index-scan
//			pruned, remained := expression.PruneSubtree(predicate.Clone(), func(subtree ast.Evaluable) bool {
//				subtreeMembers := expression.GetAccessorMembers(subtree)
//				sortStrings(subtreeMembers)
//				return slices.Compare(subtreeMembers, members) == 0
//			})
//		}
//		// push the leftmost column covered predicate to index-scan
//		pruned, remained := expression.PruneSubtree(predicate.Clone(), func(subtree ast.Evaluable) bool {
//			subtreeMembers := expression.GetAccessorMembers(subtree)
//			return len(subtreeMembers) == 1 && subtreeMembers[0] == leftmost
//		})
//	}
//
//	return nil
//}
