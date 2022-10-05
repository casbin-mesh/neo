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
	"github.com/casbin-mesh/neo/pkg/neo/session"
)

type LogicalOptimizer struct {
	ctx session.Base
}

func NewLogicalOptimizer(ctx session.Base) *LogicalOptimizer {
	return &LogicalOptimizer{ctx: ctx}
}

func (lo *LogicalOptimizer) Optimize(tree ast.Evaluable) plan.AbstractPlan {
	predicate := Optimize(tree)
	return generatePlans(lo.ctx, predicate)
}

func getConstAndNonConstExpr(ctx session.Base, predicate Predicate) (Predicate, Predicate) {
	switch predicate.Type {
	case And, Or:
		var constNodes, nonConstNodes []Predicate
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
		return Predicate{Type: predicate.Type, Args: constNodes}, Predicate{Type: predicate.Type, Args: nonConstNodes}
	case Other:
		if IsConstNode(ctx, predicate.Expr) {
			return predicate, Predicate{}
		} else {
			return Predicate{}, predicate
		}
	}
	return Predicate{}, Predicate{}
}

func handleConstNode(constNode Predicate, plan plan.AbstractPlan, tp PredicateType) plan.AbstractPlan {
	switch tp {
	case And:
		ret := &LogicalAndPlan{}
		ret.Const = append(ret.Const, &LogicalConst{Predicate: constNode})
		ret.NonConst = append(ret.NonConst, plan)

		return ret
	case Or:
		ret := &LogicalOrPlan{}
		ret.Const = append(ret.Const, &LogicalConst{Predicate: constNode})
		ret.NonConst = append(ret.NonConst, plan)

		return ret
	}
	return &LogicalConst{}
}
