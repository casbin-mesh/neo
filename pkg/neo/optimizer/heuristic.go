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
	"github.com/casbin-mesh/neo/pkg/neo/session"
)

func Optimize(root ast.Evaluable) Predicate {
	return RewritePredicate(CNFPredicates(RewritePredicate(NewPredicate(root))))
}

//func predicateType2AstOp(predicateType PredicateType) ast.Op {
//	switch predicateType {
//	case And:
//		return ast.AND_OP
//	case Or:
//		return ast.OR_OP
//	}
//	return 0
//}

//func Predicate2Evaluable(predicate Predicate) ast.Evaluable {
//	switch predicate.Type {
//	case And, Or:
//		if len(predicate.Args) == 0 {
//			return Predicate2Evaluable(predicate.Args[0])
//		}
//		cur := &ast.BinaryOperationExpr{
//			Op: predicateType2AstOp(predicate.Type),
//			L:  Predicate2Evaluable(predicate.Args[0]),
//			R:  nil,
//		}
//		for i, arg := range predicate.Args {
//			if i == 0 {
//				continue
//			}
//			cur.R = Predicate2Evaluable(arg)
//			cur = &ast.BinaryOperationExpr{
//				Op: predicateType2AstOp(predicate.Type),
//				L:  cur,
//			}
//		}
//		if cur.R == nil {
//			return cur.L
//		}
//		return cur
//	case Other:
//		return predicate.Expr
//	}
//	return nil
//}

func flatPredicates(predicate Predicate) []Predicate {
	switch predicate.Type {
	case And:
		return []Predicate{predicate}
	case Or:
		return predicate.Args
	default:
		return []Predicate{predicate}
	}
}

func buildAndPredicate(target Predicate, args ...Predicate) Predicate {
	switch target.Type {
	case And:
		target.Args = append(target.Args, args...)
		return target
	default:
		var predicates []Predicate
		predicates = append(predicates, target)
		predicates = append(predicates, args...)
		return Predicate{Type: And, Args: predicates}
	}
}

// CNFPredicates
// (a || b) && (c || d) && (e || f)
// tmpArgs: (a || b)
// arg: (c || d)
// newArgs: a && c || a && d || b && c || b && d
// tmpArgs: a && c || a && d || b && c || b && d
// arg: (e || f)
//
func CNFPredicates(predicate Predicate) Predicate {
	switch predicate.Type {
	case And:
		for i, arg := range predicate.Args {
			predicate.Args[i] = CNFPredicates(arg)
		}

		if len(predicate.Args) == 1 {
			return predicate.Args[0]
		}
		var tmpArgs []Predicate
		var newArgs []Predicate

		tmpArgs = append(tmpArgs, flatPredicates(predicate.Args[0])...)
		for i, arg := range predicate.Args {
			if i == 0 {
				continue
			}
			for _, pred := range flatPredicates(arg) {
				for _, tmpArg := range tmpArgs {
					newArgs = append(newArgs, buildAndPredicate(tmpArg, pred))
				}
			}
			tmpArgs = newArgs
			newArgs = nil
		}
		if len(tmpArgs) == 1 {
			return tmpArgs[0]
		}
		return Predicate{Type: Or, Args: tmpArgs}
	case Or:
		for i, arg := range predicate.Args {
			predicate.Args[i] = CNFPredicates(arg)
		}
		return predicate
	default:
		return predicate
	}
}

func RewritePredicate(predicate Predicate) Predicate {
	switch predicate.Type {
	case Or:
		args := make([]Predicate, 0, len(predicate.Args))
		for _, arg := range predicate.Args {
			rewritten := RewritePredicate(arg)
			args = append(args, rewritten)
		}
		flatten := FlatOrs(args)
		return Predicate{Type: Or, Args: flatten}
	case And:
		args := make([]Predicate, 0, len(predicate.Args))
		for _, arg := range predicate.Args {
			rewritten := RewritePredicate(arg)
			args = append(args, rewritten)
		}
		return Predicate{Type: And, Args: FlatAnds(args)}
	default:
		return predicate
	}
}

func FlatOrs(predicates []Predicate) (flatten []Predicate) {
	for _, pre := range predicates {
		switch pre.Type {
		case Or:
			flatten = append(flatten, FlatOrs(pre.Args)...)
		default:
			flatten = append(flatten, pre)
		}
	}
	return
}

func FlatAnds(predicates []Predicate) (flatten []Predicate) {
	for _, pre := range predicates {
		switch pre.Type {
		case And:
			flatten = append(flatten, FlatAnds(pre.Args)...)
		default:
			flatten = append(flatten, pre)
		}
	}
	return
}

func ApplySkip(predicate *Predicate, fn func(evaluable ast.Evaluable) bool) bool {
	switch predicate.Type {
	case And, Or:
		for _, arg := range predicate.Args {
			if !ApplySkip(&arg, fn) {
				return false
			}
		}
		return true
	case Other:
		return fn(predicate.Expr)
	}
	return true
}

func AppendAst2Predicate(predicate *Predicate, ast ast.Evaluable, skip func(ast ast.Evaluable) bool) {
	switch predicate.Type {
	case And:
		if skip == nil || !ApplySkip(predicate, skip) {
			predicate.Args = append(predicate.Args, NewPredicate(ast))
		}
	case Or:
		if skip == nil || !ApplySkip(predicate, skip) {
			for i, arg := range predicate.Args {
				switch arg.Type {
				case And:
					AppendAst2Predicate(&predicate.Args[i], ast, skip)
				case Or:
					AppendAst2Predicate(&predicate.Args[i], ast, skip)
				case Other:
					AppendAst2Predicate(&predicate.Args[i], ast, skip)
				}
			}
		}
	case Other:
		if skip == nil || !ApplySkip(predicate, skip) {
			*predicate = Predicate{
				Type: And,
				Args: []Predicate{*predicate, NewPredicate(ast)},
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

func IsConstPredicate(ctx session.Base, p Predicate) bool {
	switch p.Type {
	case And, Or:
		for _, arg := range p.Args {
			if !IsConstPredicate(ctx, arg) {
				return false
			}
		}
		return true
	case Other:
		return IsConstNode(ctx, p.Expr)
	}
	return true
}

func IsConstNode(ctx session.Base, node ast.Evaluable) (ok bool) {
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
