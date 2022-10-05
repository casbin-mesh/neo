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
	"strings"
)

type PredicateType int

const (
	Or PredicateType = iota + 1
	And
	Other
)

type Predicate struct {
	Type PredicateType
	Args []Predicate
	Expr ast.Evaluable
}

func (p *Predicate) HasPredicates() bool {
	return len(p.Args) > 0 || p.Expr != nil
}

func (p *Predicate) String() string {
	switch p.Type {
	case And:
		conds := make([]string, 0, len(p.Args))
		for _, arg := range p.Args {
			conds = append(conds, arg.String())
		}

		return fmt.Sprintf("( %s )", strings.Join(conds, " && "))
	case Or:
		conds := make([]string, 0, len(p.Args))
		for _, arg := range p.Args {
			conds = append(conds, arg.String())
		}
		return fmt.Sprintf("( %s )", strings.Join(conds, " || "))
	case Other:
		return p.Expr.String()
	default:
		return "unknown predicate"
	}
}

func (p *Predicate) Clone() Predicate {
	np := *p
	np.Args = make([]Predicate, len(p.Args))
	for i, arg := range p.Args {
		np.Args[i] = arg.Clone()
	}
	if p.Expr != nil {
		np.Expr = p.Expr.Clone()
	}
	return np
}

// IncludeAccessorOnly returns true where both sides of binary operation expr are accessor
func IncludeAccessorOnly(cur ast.Evaluable) bool {
	switch v := cur.(type) {
	case *ast.BinaryOperationExpr:
		_, lok := v.L.(*ast.Accessor)
		_, rok := v.R.(*ast.Accessor)
		return !(lok && rok)
	default:
		return true
	}
}

func GetPredicateAccessorMembers(p Predicate, skip func(ast ast.Evaluable) bool) (result []string) {
	nameSet := make(map[string]struct{})
	switch p.Type {
	case Or, And:
		for _, arg := range p.Args {
			for _, name := range GetPredicateAccessorMembers(arg, skip) {
				nameSet[name] = struct{}{}
			}
		}
	case Other:
		if skip == nil || (skip != nil && !skip(p.Expr)) {
			return expression.GetAccessorMembers(p.Expr)
		}
	}
	result = make([]string, 0, len(nameSet))
	for name, _ := range nameSet {
		result = append(result, name)
	}
	return result
}

func predicateType2AstOp(predicateType PredicateType) ast.Op {
	switch predicateType {
	case And:
		return ast.AND_OP
	case Or:
		return ast.OR_OP
	}
	return 0
}

func PrunePredicate(predicate Predicate, prune func(evaluable ast.Evaluable) bool) (pruned *Predicate, remaining *Predicate) {
	var result []Predicate
	var remained []Predicate
	switch predicate.Type {
	case And, Or:
		for _, arg := range predicate.Args {
			r, rr := PrunePredicate(arg, prune)
			if r != nil {
				result = append(result, *r)
			}
			if rr != nil {
				remained = append(remained, *rr)
			}
		}
	case Other:
		if prune(predicate.Expr) {
			return &predicate, nil
		}
		return nil, &predicate
	}
	if len(result) == 0 {
		pruned = nil
	} else if len(result) == 1 {
		pruned = &result[0]
	} else {
		pruned = &Predicate{Type: predicate.Type, Args: result}
	}

	if len(remained) == 0 {
		remaining = nil
	} else if len(remained) == 1 {
		remaining = &remained[0]
	} else {
		remaining = &Predicate{Type: predicate.Type, Args: remained}
	}

	return
}

func NewPredicate(root ast.Evaluable) Predicate {
	switch root.(type) {
	case *ast.BinaryOperationExpr:
		binExpr := root.(*ast.BinaryOperationExpr)
		if binExpr.Op == ast.AND_OP {
			args := make([]Predicate, 0, 2)

			args = append(args, NewPredicate(binExpr.L))
			args = append(args, NewPredicate(binExpr.R))
			return Predicate{
				Type: And,
				Args: args,
			}
		} else if binExpr.Op == ast.OR_OP {
			args := make([]Predicate, 0, 2)
			args = append(args, NewPredicate(binExpr.L))
			args = append(args, NewPredicate(binExpr.R))
			return Predicate{
				Type: Or,
				Args: args,
			}
		}
	}
	return Predicate{
		Type: Other,
		Expr: root,
	}
}

func Predicate2Evaluable(predicate Predicate) ast.Evaluable {
	switch predicate.Type {
	case And, Or:
		if len(predicate.Args) == 0 {
			return nil
		}
		if len(predicate.Args) == 1 {
			return Predicate2Evaluable(predicate.Args[0])
		}
		cur := &ast.BinaryOperationExpr{
			Op: predicateType2AstOp(predicate.Type),
			L:  Predicate2Evaluable(predicate.Args[0]),
			R:  nil,
		}
		for i, arg := range predicate.Args {
			if i == 0 {
				continue
			}
			cur.R = Predicate2Evaluable(arg)
			cur = &ast.BinaryOperationExpr{
				Op: predicateType2AstOp(predicate.Type),
				L:  cur,
			}
		}
		if cur.R == nil {
			return cur.L
		}
		return cur
	case Other:
		return predicate.Expr
	}
	return nil
}
