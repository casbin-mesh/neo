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

package heuristic

import (
	"github.com/casbin-mesh/neo/pkg/expression/ast"
	"github.com/casbin-mesh/neo/pkg/neo/optimizer/ctx"
	"github.com/casbin-mesh/neo/pkg/neo/optimizer/node"
)

func Optimize(root ast.Evaluable) node.Predicate {
	return RewritePredicate(NewPredicate(root))
}

func NewPredicate(root ast.Evaluable) node.Predicate {
	switch root.(type) {
	case *ast.BinaryOperationExpr:
		binExpr := root.(*ast.BinaryOperationExpr)
		if binExpr.Op == ast.AND_OP {
			args := make([]node.Predicate, 0, 2)

			args = append(args, NewPredicate(binExpr.L))
			args = append(args, NewPredicate(binExpr.R))
			return node.Predicate{
				Type: node.And,
				Args: args,
			}
		} else if binExpr.Op == ast.OR_OP {
			args := make([]node.Predicate, 0, 2)
			args = append(args, NewPredicate(binExpr.L))
			args = append(args, NewPredicate(binExpr.R))
			return node.Predicate{
				Type: node.Or,
				Args: args,
			}
		}
	}
	return node.Predicate{
		Type: node.Other,
		Expr: root,
	}
}

func RewritePredicate(predicate node.Predicate) node.Predicate {
	switch predicate.Type {
	case node.Or:
		args := make([]node.Predicate, 0, len(predicate.Args))
		for _, arg := range predicate.Args {
			rewritten := RewritePredicate(arg)
			args = append(args, rewritten)
		}
		flatten := FlatOrs(args)
		return node.Predicate{Type: node.Or, Args: flatten}
	case node.And:
		args := make([]node.Predicate, 0, len(predicate.Args))
		for _, arg := range predicate.Args {
			rewritten := RewritePredicate(arg)
			args = append(args, rewritten)
		}
		return node.Predicate{Type: node.And, Args: FlatAnds(args)}
	default:
		return predicate
	}
}

func FlatOrs(predicates []node.Predicate) (flatten []node.Predicate) {
	for _, pre := range predicates {
		switch pre.Type {
		case node.Or:
			flatten = append(flatten, FlatOrs(pre.Args)...)
		default:
			flatten = append(flatten, pre)
		}
	}
	return
}

func FlatAnds(predicates []node.Predicate) (flatten []node.Predicate) {
	for _, pre := range predicates {
		switch pre.Type {
		case node.And:
			flatten = append(flatten, FlatAnds(pre.Args)...)
		default:
			flatten = append(flatten, pre)
		}
	}
	return
}

func ApplySkip(predicate *node.Predicate, fn func(evaluable ast.Evaluable) bool) bool {
	switch predicate.Type {
	case node.And, node.Or:
		for _, arg := range predicate.Args {
			if !ApplySkip(&arg, fn) {
				return false
			}
		}
		return true
	case node.Other:
		return fn(predicate.Expr)
	}
	return true
}

func AppendAst2Predicate(predicate *node.Predicate, ast ast.Evaluable, skip func(ast ast.Evaluable) bool) {
	switch predicate.Type {
	case node.And:
		if skip == nil || !ApplySkip(predicate, skip) {
			predicate.Args = append(predicate.Args, NewPredicate(ast))
		}
	case node.Or:
		if skip == nil || !ApplySkip(predicate, skip) {
			for i, arg := range predicate.Args {
				switch arg.Type {
				case node.And:
					AppendAst2Predicate(&predicate.Args[i], ast, skip)
				case node.Or:
					AppendAst2Predicate(&predicate.Args[i], ast, skip)
				case node.Other:
					AppendAst2Predicate(&predicate.Args[i], ast, skip)
				}
			}
		}
	case node.Other:
		if skip == nil || !ApplySkip(predicate, skip) {
			*predicate = node.Predicate{
				Type: node.And,
				Args: []node.Predicate{*predicate, NewPredicate(ast)},
			}
		}
	}
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

func IsConstPredicate(ctx ctx.Base, p node.Predicate) bool {
	switch p.Type {
	case node.And, node.Or:
		for _, arg := range p.Args {
			if !IsConstPredicate(ctx, arg) {
				return false
			}
		}
		return true
	case node.Other:
		return IsConstNode(ctx, p.Expr)
	}
	return true
}

func IsConstNode(ctx ctx.Base, node ast.Evaluable) (ok bool) {
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
