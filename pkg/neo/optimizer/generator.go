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
	"github.com/casbin-mesh/neo/pkg/neo/session"
	"github.com/casbin-mesh/neo/pkg/neo/utils"
)

type MatcherGenerator struct {
	ctx session.Base
}

func NewMatcherGenerator(ctx session.Base) *MatcherGenerator {
	return &MatcherGenerator{ctx: ctx}
}

func (g *MatcherGenerator) Generate(tree ast.Evaluable) plan.AbstractPlan {
	predicate := Optimize(tree)
	return generatePlans(g.ctx, predicate)
}

func explorePlans(ctx session.Base, predicate Predicate) plan.AbstractPlan {
	dbId, tableId := ctx.DB().ID, ctx.Table().ID
	if IsConstPredicate(ctx, predicate) {
		return &LogicalConst{
			Predicate: predicate,
		}
	}
	switch predicate.Type {
	case And:

		// without index -> seq scan
		if len(ctx.Table().Indices) == 0 {
			ret := &LogicalSeqScan{
				Predicate: predicate,
				DbId:      dbId,
				TableId:   tableId,
			}
			return ret
		}
		// indexes covers few cols -> indexScan + tableRowIdScan
		members := GetPredicateAccessorMembers(predicate, nil)
		indexCols := GetIndexColsName(ctx.Table().Indices)
		result := utils.SortedIntersect(members, indexCols)

		if len(result) > 0 {
			return &LogicalIndexLookupReader{
				Build: &LogicalIndexReader{
					Table:     ctx.Table(),
					Indexes:   ctx.Table().Indices,
					Predicate: predicate,
					TableId:   tableId,
					DbId:      dbId,
				},
				Probe: &LogicalRowIdScan{
					TableId: tableId,
				},
				Predicate: predicate,
			}
		} else { // TODO: support multi-index-scan?
			// full-scan
			ret := &LogicalSeqScan{
				Predicate: predicate,
				DbId:      dbId,
				TableId:   tableId,
			}
			return ret
		}
	case Or:
		var children []plan.AbstractPlan
		var constNodes []plan.AbstractPlan
		for _, arg := range predicate.Args {
			child := explorePlans(ctx, arg)
			if _, ok := child.(*LogicalConst); ok {
				constNodes = append(constNodes, child)
			} else {
				children = append(children, child)
			}
		}
		// TODO: indexes covers all predicate -> index merge
		// otherwise -> full-scan
		return &LogicalOrPlan{
			Const:    constNodes,
			NonConst: children,
		}

	}
	return nil
}

func generatePlans(ctx session.Base, predicate Predicate) plan.AbstractPlan {
	subPlans := model.GenerateMatcherEffectPolicyAst(ctx.PolicyTableName(), ctx.EffectColName(), ctx.AllowIdent(), ctx.DenyIdent(), ctx.Matcher().EffectPolicy)
	predicates := make([]Predicate, 0, len(subPlans))
	// merge policy effect expression
	for _, evaluable := range subPlans {
		p := predicate.Clone()
		AppendAst2Predicate(&p, evaluable, func(node ast.Evaluable) bool {
			return IsConstNode(ctx, node)
		})
		predicates = append(predicates, p)
	}
	children := make([]plan.AbstractPlan, 0, len(predicates))
	for _, pre := range predicates {
		children = append(children, explorePlans(ctx, pre))
	}
	return &LogicalMatcherPlan{
		Type:       MatcherPlanType(ctx.Matcher().EffectPolicy),
		Predicates: predicates,
		Children:   children,
	}
}

func GetIndexColsName(index []*model.IndexInfo) (result []string) {
	result = make([]string, 0, len(index))
	for _, info := range index {
		result = append(result, info.Columns[0].ColName.L)
	}
	return
}

type SelectPlanGenerator struct {
	ctx session.Base
}

func NewSelectPlanGenerator(ctx session.Base) *SelectPlanGenerator {
	return &SelectPlanGenerator{ctx: ctx}
}

func (g *SelectPlanGenerator) Generate(tree ast.Evaluable) plan.AbstractPlan {
	predicate := Optimize(tree)
	child := explorePlans(g.ctx, predicate)
	return child
}
