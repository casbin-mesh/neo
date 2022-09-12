// Copyright 2022 The casbin-neo Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package optimizer

import (
	"github.com/casbin-mesh/neo/pkg/expression/ast"
	"github.com/casbin-mesh/neo/pkg/neo/executor/plan"
	"github.com/casbin-mesh/neo/pkg/neo/model"
	"github.com/casbin-mesh/neo/pkg/neo/optimizer/ctx"
	"github.com/casbin-mesh/neo/pkg/neo/optimizer/heuristic"
	"github.com/casbin-mesh/neo/pkg/neo/optimizer/node"
	"github.com/casbin-mesh/neo/pkg/neo/utils"
)

type LogicalOptimizer struct {
	ctx ctx.Base
}

func NewLogicalOptimizer(ctx ctx.Base) *LogicalOptimizer {
	return &LogicalOptimizer{ctx: ctx}
}

func (lo *LogicalOptimizer) Optimize(tree ast.Evaluable) node.LogicalMatcherPlan {
	predicate := heuristic.Optimize(tree)
	return generatePlans(lo.ctx, predicate)
}

func GetIndexColsName(index []*model.IndexInfo) (result []string) {
	result = make([]string, 0, len(index))
	for _, info := range index {
		result = append(result, info.Columns[0].ColName.L)
	}
	return
}

func getConstAndNonConstExpr(ctx ctx.Base, predicate node.Predicate) (node.Predicate, node.Predicate) {
	switch predicate.Type {
	case node.And, node.Or:
		var constNodes, nonConstNodes []node.Predicate
		for _, arg := range predicate.Args {
			constNode, nonConst := getConstAndNonConstExpr(ctx, arg)
			if constNode.HasPredicates() && !nonConst.HasPredicates() { // const
				if constNode.HasPredicates() {
					constNodes = append(constNodes, constNode)
				}
				if nonConst.HasPredicates() {
					nonConstNodes = append(constNodes, nonConst)
				}
			} else {

			}

		}
		return node.Predicate{Type: predicate.Type, Args: constNodes}, node.Predicate{Type: predicate.Type, Args: nonConstNodes}
	case node.Other:
		if heuristic.IsConstNode(ctx, predicate.Expr) {
			return predicate, node.Predicate{}
		} else {
			return node.Predicate{}, predicate
		}
	}
	return node.Predicate{}, node.Predicate{}
}

func handleConstNode(constNode node.Predicate, plan plan.AbstractPlan, tp node.PredicateType) plan.AbstractPlan {
	switch tp {
	case node.And:
		ret := &node.LogicalAndPlan{}
		ret.Const = append(ret.Const, &node.LogicalConst{Predicate: constNode})
		ret.NonConst = append(ret.NonConst, plan)

		return ret
	case node.Or:
		ret := &node.LogicalOrPlan{}
		ret.Const = append(ret.Const, &node.LogicalConst{Predicate: constNode})
		ret.NonConst = append(ret.NonConst, plan)

		return ret
	}
	return &node.LogicalConst{}
}

func explorePlans(ctx ctx.Base, predicate node.Predicate) plan.AbstractPlan {
	dbId, tableId := ctx.DB().ID, ctx.Table().ID
	if heuristic.IsConstPredicate(ctx, predicate) {
		return &node.LogicalConst{
			Predicate: predicate,
		}
	}
	switch predicate.Type {
	case node.And:

		// without index -> seq scan
		if len(ctx.Table().Indices) == 0 {
			ret := &node.LogicalSeqScan{
				Predicate: predicate,
				DbId:      dbId,
				TableId:   tableId,
			}
			return ret
		}
		// indexes covers few cols -> indexScan + tableRowIdScan
		members := node.GetPredicateAccessorMembers(predicate)
		indexCols := GetIndexColsName(ctx.Table().Indices)
		result := utils.SortedIntersect(members, indexCols)

		if len(result) > 0 {
			return &node.LogicalIndexLookupReader{
				Build: &node.LogicalIndexReader{
					Indexes:   ctx.Table().Indices,
					Predicate: predicate,
				},
				Probe: &node.LogicalRowIdScan{
					TableId: tableId,
				},
				Predicate: predicate,
			}
		} else { // TODO: support multi-index-scan?
			// full-scan
			ret := &node.LogicalSeqScan{
				Predicate: predicate,
				DbId:      dbId,
				TableId:   tableId,
			}
			return ret
		}
	case node.Or:
		var children []plan.AbstractPlan
		var constNodes []plan.AbstractPlan
		for _, arg := range predicate.Args {
			child := explorePlans(ctx, arg)
			if _, ok := child.(*node.LogicalConst); ok {
				constNodes = append(constNodes, child)
			} else {
				children = append(children, child)
			}
		}
		// TODO: indexes covers all predicate -> index merge
		// otherwise -> full-scan
		return &node.LogicalOrPlan{
			Const:    constNodes,
			NonConst: children,
		}

	}
	return nil
}

func generatePlans(ctx ctx.Base, predicate node.Predicate) node.LogicalMatcherPlan {
	subPlans := model.GenerateEffectPolicyAst(ctx.PolicyTableName(), ctx.EffectColName(), ctx.AllowIdent(), ctx.DenyIdent(), ctx.Matcher().EffectPolicy)
	predicates := make([]node.Predicate, 0, len(subPlans))
	// merge policy effect expression
	for _, evaluable := range subPlans {
		p := predicate.Clone()
		heuristic.AppendAst2Predicate(&p, evaluable, func(node ast.Evaluable) bool {
			return heuristic.IsConstNode(ctx, node)
		})
		predicates = append(predicates, p)
	}
	children := make([]plan.AbstractPlan, 0, len(predicates))
	for _, pre := range predicates {
		children = append(children, explorePlans(ctx, pre))
	}
	return node.LogicalMatcherPlan{
		Type:       node.MatcherPlanType(ctx.Matcher().EffectPolicy),
		Predicates: predicates,
		Children:   children,
	}
}
