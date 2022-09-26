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
	"fmt"
	"github.com/casbin-mesh/neo/pkg/expression"
	"github.com/casbin-mesh/neo/pkg/expression/ast"
	"github.com/casbin-mesh/neo/pkg/neo/codec"
	"github.com/casbin-mesh/neo/pkg/neo/executor/plan"
	"github.com/casbin-mesh/neo/pkg/neo/model"
	"github.com/casbin-mesh/neo/pkg/neo/session"
	"github.com/casbin-mesh/neo/pkg/neo/utils"
	"golang.org/x/exp/slices"
)

type MatcherPlanType int

const (
	AllowOverride MatcherPlanType = iota
	DenyOverride
	AllowAndDeny
	Priority
	PriorityBaseOnRole
)

type PhysicalPlan struct {
	Plan     plan.AbstractPlan
	Children []PhysicalPlan
	Property
}

type LogicalSeqScan struct {
	plan.AbstractPlan
	Predicate Predicate
	DbId      uint64
	TableId   uint64
}

func (p *LogicalSeqScan) String() string {
	return fmt.Sprintf("%s | DB: %d | Table: %d | Predicate: %s", "LogicalSeqScan", p.DbId, p.TableId, p.Predicate.String())
}

func (p *LogicalSeqScan) FindBestPlan(ctx session.OptimizerCtx) plan.AbstractPlan {
	// TODO: support function memoization
	evalCtx := ast.NewContext()
	expr, accessor := expression.NewExpression(Predicate2Evaluable(p.Predicate))
	evalCtx.AddAccessor(ctx.PolicyTableName(), accessor)
	evalCtx.AddAccessor(ctx.ReqAccessorAncestorName(), ctx.ReqAccessor())
	return plan.NewSeqScanPlan(ctx.Table(), expr, evalCtx, p.DbId, p.TableId)
}

type LogicalIndexLookupReader struct {
	plan.AbstractPlan
	Build     plan.AbstractPlan
	Probe     plan.AbstractPlan
	Predicate Predicate
}

func (p *LogicalIndexLookupReader) String() string {
	childStr := make([]string, 0, 2)
	childStr = append(childStr, "(Build)"+p.Build.String())
	childStr = append(childStr, "(Probe)"+p.Probe.String())
	return utils.TreeFormat(fmt.Sprintf("LogicalIndexLookupReader | Predicate: %s", p.Predicate.String()), childStr...)
}

func (p *LogicalIndexLookupReader) FindBestPlan(ctx session.OptimizerCtx) plan.AbstractPlan {
	indexScan := p.Build.FindBestPlan(ctx)
	evalCtx := ast.NewContext()
	expr, accessor := expression.NewExpression(Predicate2Evaluable(p.Predicate))
	evalCtx.AddAccessor(ctx.PolicyTableName(), accessor)
	evalCtx.AddAccessor(ctx.ReqAccessorAncestorName(), ctx.ReqAccessor())
	return plan.NewTableRowIdScan(ctx.Table(), expr, evalCtx, ctx.DB().ID, ctx.Table().ID, indexScan)
}

type LogicalRowIdScan struct {
	plan.AbstractPlan
	TableId uint64
}

func (p *LogicalRowIdScan) String() string {
	return fmt.Sprintf("%s | Table: %d", "LogicalRowIdScan", p.TableId)
}

type LogicalIndexReader struct {
	plan.AbstractPlan
	Table         *model.TableInfo
	Indexes       []*model.IndexInfo
	Predicate     Predicate
	DbId          uint64
	TableId       uint64
	CoveredMember []string
}

func (p *LogicalIndexReader) String() string {
	return fmt.Sprintf("LogicalIndexReader | Predicate: %s", p.Predicate.String())
}

func (p *LogicalIndexReader) FindBestPlan(ctx session.OptimizerCtx) plan.AbstractPlan {
	members := GetPredicateAccessorMembers(p.Predicate, nil)

	maxCoveredIndex := -1
	var coveredMembers []string
	for i, index := range p.Indexes {
		if slices.Contains(members, index.Columns[0].ColName.L) {
			tmpCovered := []string{index.Columns[0].ColName.L}
			for i := 1; i < len(index.Columns); i++ {
				col := index.Columns[i]
				if slices.Contains(members, col.ColName.L) {
					tmpCovered = append(tmpCovered, col.ColName.L)
				}
			}
			if maxCoveredIndex == -1 || len(tmpCovered) > len(coveredMembers) {
				maxCoveredIndex = i
				coveredMembers = tmpCovered
			}
		}
	}

	pruned := *PrunePredicate(p.Predicate, func(evaluable ast.Evaluable) bool {
		member := expression.GetAccessorMembers(evaluable)
		if len(member) > 0 {
			return slices.Contains(coveredMembers, member[0])
		}
		return false
	})

	// coveredMembers[0]
	idx := slices.IndexFunc(p.Indexes, func(i *model.IndexInfo) bool { return i.Columns[0].ColName.L == coveredMembers[0] })
	index := p.Indexes[idx]
	// build prefix
	prefix := codec.PrimaryIndexEntryKey(index.ID, codec.EncodePrimitive(ctx.ReqAccessor().GetMember(coveredMembers[0])))

	evalCtx := ast.NewContext()
	expr, accessor := expression.NewExpression(Predicate2Evaluable(pruned))
	evalCtx.AddAccessor(ctx.PolicyTableName(), accessor)
	evalCtx.AddAccessor(ctx.ReqAccessorAncestorName(), ctx.ReqAccessor())

	return plan.NewIndexScanPlan(model.NewIndexSchemaReader(p.Table, maxCoveredIndex), prefix, expr, evalCtx, p.DbId, p.TableId)
}

type LogicalMatcherPlan struct {
	plan.AbstractPlan
	Type       MatcherPlanType
	Predicates []Predicate
	Children   []plan.AbstractPlan
}

var eftPolicy2Str = []string{
	"AllowOverride",
	"DenyOverride",
	"AllowAndDeny",
	"Priority",
	"PriorityBaseOnRole",
}

func (p *LogicalMatcherPlan) String() string {
	childStr := make([]string, 0, len(p.Children))
	for _, child := range p.Children {
		childStr = append(childStr, child.String())
	}
	return utils.TreeFormat(fmt.Sprintf("LogicalMatcherPlan | Type: %s", eftPolicy2Str[p.Type]), childStr...)
}

func (p *LogicalMatcherPlan) FindBestPlan(ctx session.OptimizerCtx) plan.AbstractPlan {
	children := make([]plan.AbstractPlan, 0, len(p.Children))
	for _, child := range p.Children {
		children = append(children, child.FindBestPlan(ctx))
	}
	return plan.NewMatcherPlan(children, plan.EffectType(p.Type))
}

type LogicalConst struct {
	plan.AbstractPlan
	Predicate Predicate
}

func (p *LogicalConst) FindBestPlan(ctx session.OptimizerCtx) plan.AbstractPlan {
	evalCtx := ast.NewContext()
	expr, accessor := expression.NewExpression(Predicate2Evaluable(p.Predicate))
	evalCtx.AddAccessor(ctx.PolicyTableName(), accessor)
	evalCtx.AddAccessor(ctx.ReqAccessorAncestorName(), ctx.ReqAccessor())
	return plan.NewConstPlan(expr, evalCtx)
}

func (p *LogicalConst) String() string {
	return utils.TreeFormat(fmt.Sprintf("LogicalConst | Predicate: %s", p.Predicate.String()))
}

type LogicalPredicate struct {
	plan.AbstractPlan
	Predicate Predicate
}

func (p *LogicalPredicate) String() string {
	return utils.TreeFormat(fmt.Sprintf("LogicalPredicate | Predicate: %s", p.Predicate.String()))
}

type LogicalAndPlan struct {
	plan.AbstractPlan
	Const    []plan.AbstractPlan
	NonConst []plan.AbstractPlan
}

func (p *LogicalAndPlan) String() string {
	childStr := make([]string, 0, len(p.Const)+len(p.NonConst))
	for _, child := range p.Const {
		childStr = append(childStr, fmt.Sprintf("(Const)%s", child.String()))
	}
	for _, child := range p.NonConst {
		childStr = append(childStr, fmt.Sprintf("(Non-Const)%s", child.String()))
	}
	return utils.TreeFormat("LogicalAndPlan", childStr...)
}

func (p *LogicalAndPlan) FindBestPlan(ctx session.OptimizerCtx) plan.AbstractPlan {
	constChild := make([]plan.AbstractPlan, 0, len(p.Const))
	nonConstChild := make([]plan.AbstractPlan, 0, len(p.NonConst))
	for _, child := range p.Const {
		constChild = append(constChild, child.FindBestPlan(ctx))
	}
	for _, child := range p.NonConst {
		nonConstChild = append(nonConstChild, child.FindBestPlan(ctx))
	}
	return plan.NewShortCircuitPlan(nonConstChild, constChild, plan.AND)
}

type LogicalOrPlan struct {
	plan.AbstractPlan
	Const    []plan.AbstractPlan
	NonConst []plan.AbstractPlan
}

func (p *LogicalOrPlan) FindBestPlan(ctx session.OptimizerCtx) plan.AbstractPlan {
	constChild := make([]plan.AbstractPlan, 0, len(p.Const))
	nonConstChild := make([]plan.AbstractPlan, 0, len(p.NonConst))
	for _, child := range p.Const {
		constChild = append(constChild, child.FindBestPlan(ctx))
	}
	for _, child := range p.NonConst {
		nonConstChild = append(nonConstChild, child.FindBestPlan(ctx))
	}
	return plan.NewShortCircuitPlan(nonConstChild, constChild, plan.OR)
}

func (p *LogicalOrPlan) String() string {
	childStr := make([]string, 0, len(p.Const)+len(p.NonConst))
	for _, child := range p.Const {
		childStr = append(childStr, fmt.Sprintf("(Const)%s", child.String()))
	}
	for _, child := range p.NonConst {
		childStr = append(childStr, fmt.Sprintf("(Non-Const)%s", child.String()))
	}
	return utils.TreeFormat("LogicalOrPlan", childStr...)
}

type Property struct {
	Indexes       []*model.IndexInfo
	SecondScan    bool
	FullTableScan bool
	Cardinality   uint64
}