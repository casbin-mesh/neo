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
	"github.com/casbin-mesh/neo/pkg/expression/ast"
	"github.com/casbin-mesh/neo/pkg/parser"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewPredicate(t *testing.T) {
	root := parser.MustParseFromString("r.sub == p.sub && (r.obj == p.obj || r.obj == \"public\") && r.act == p.act || r.sub == \"root\" ")
	pre := NewPredicate(root)
	re := RewritePredicate(pre)
	fmt.Println(re)

}

type testAppendAst2Predicate struct {
	root     ast.Evaluable
	eft      ast.Evaluable
	skip     func(ast ast.Evaluable) bool
	expected Predicate
}

func runTest(sets []testAppendAst2Predicate, t *testing.T) {
	for _, set := range sets {
		predicate := RewritePredicate(NewPredicate(set.root))
		AppendAst2Predicate(&predicate, set.eft, set.skip)
		assert.Equal(t, set.expected, predicate)
	}
}

func TestAppendAst2Predicate2(t *testing.T) {
	sets := []testAppendAst2Predicate{
		{
			root:     parser.MustParseFromString("r.sub == p.sub && r.obj == p.obj && r.act == p.act"),
			eft:      parser.MustParseFromString("r.eft == allow"),
			expected: RewritePredicate(NewPredicate(parser.MustParseFromString("r.sub == p.sub && r.obj == p.obj && r.act == p.act && r.eft == allow"))),
		},
		{
			root:     parser.MustParseFromString("r.sub == p.sub && r.obj == p.obj && r.act == p.act || r.sub == \"root\""),
			eft:      parser.MustParseFromString("r.eft == allow"),
			expected: RewritePredicate(NewPredicate(parser.MustParseFromString("r.sub == p.sub && r.obj == p.obj && r.act == p.act && r.eft == allow || r.sub == \"root\""))),
			skip: func(node ast.Evaluable) bool { // mock skip const node
				n, ok := node.(*ast.BinaryOperationExpr)
				if ok {
					if v, ok := n.R.(*ast.Primitive); ok {
						return v.Typ == ast.STRING
					}
				}
				return false
			},
		},
		{
			// r.sub == p.sub && (r.obj == p.obj || p.obj == "public" ) && r.act == p.act
			root:     parser.MustParseFromString("r.sub == p.sub && (r.obj == p.obj || r.obj == \"public\") && r.act == p.act || r.sub == \"root\""),
			eft:      parser.MustParseFromString("r.eft == allow"),
			expected: RewritePredicate(NewPredicate(parser.MustParseFromString("r.sub == p.sub && (r.obj == p.obj || r.obj == \"public\") && r.act == p.act && r.eft == allow || r.sub == \"root\""))),
			skip: func(node ast.Evaluable) bool {
				n, ok := node.(*ast.BinaryOperationExpr)
				if ok {
					if v, ok := n.R.(*ast.Primitive); ok {
						return v.Typ == ast.STRING
					}
				}
				return false
			},
		},
		{
			root:     parser.MustParseFromString("r.sub == p.sub || r.obj == p.obj || r.act == p.act"),
			eft:      parser.MustParseFromString("r.eft == allow"),
			expected: RewritePredicate(NewPredicate(parser.MustParseFromString("r.sub == p.sub && r.eft == allow || r.obj == p.obj && r.eft == allow || r.act == p.act && r.eft == allow"))),
			skip: func(node ast.Evaluable) bool {
				n, ok := node.(*ast.BinaryOperationExpr)
				if ok {
					if v, ok := n.R.(*ast.Primitive); ok {
						return v.Typ == ast.STRING
					}
				}
				return false
			},
		},
		{
			root:     parser.MustParseFromString("r.sub == \"root\" || r.obj == \"public\" || r.act == \"public\""),
			eft:      parser.MustParseFromString("r.eft == allow"),
			expected: RewritePredicate(NewPredicate(parser.MustParseFromString("r.sub == \"root\" || r.obj == \"public\" || r.act == \"public\""))),
			skip: func(node ast.Evaluable) bool {
				n, ok := node.(*ast.BinaryOperationExpr)
				if ok {
					if v, ok := n.R.(*ast.Primitive); ok {
						return v.Typ == ast.STRING
					}
				}
				return false
			},
		},
	}
	runTest(sets, t)
}

type testPredicate2Evaluable struct {
	input ast.Evaluable
}

func runPredicate2EvaluableTests(sets []testPredicate2Evaluable, t *testing.T) {
	for _, set := range sets {
		root := set.input
		pre := NewPredicate(root.Clone())
		tree := Predicate2Evaluable(pre)
		assert.Equal(t, root.String(), tree.String())
	}
}

func TestPredicate2Evaluable(t *testing.T) {
	sets := []testPredicate2Evaluable{
		{
			input: parser.MustParseFromString("r.sub == p.sub && r.obj == p.obj && r.act == p.act"),
		},
		{
			input: parser.MustParseFromString("r.sub == p.sub && r.obj == p.obj && r.act == p.act || r.sub == \"root\""),
		},
		{
			input: parser.MustParseFromString("r.sub == p.sub && (r.obj == p.obj || p.obj == \"public\" ) && r.act == p.act"),
		},
		{
			input: parser.MustParseFromString("r.sub == \"root\" || r.obj == \"public\" || r.act == \"public\""),
		},
		{
			input: parser.MustParseFromString("(r.subOwner == p.subOwner || p.subOwner == \"*\") && \\\n    (r.subName == p.subName || p.subName == \"*\" || r.subName != \"anonymous\" && p.subName == \"!anonymous\") && \\\n    (r.method == p.method || p.method == \"*\") && \\\n    (r.urlPath == p.urlPath || p.urlPath == \"*\") && \\\n    (r.objOwner == p.objOwner || p.objOwner == \"*\") && \\\n    (r.objName == p.objName || p.objName == \"*\") || \\\n    (r.subOwner == r.objOwner && r.subName == r.objName)"),
		},
	}
	runPredicate2EvaluableTests(sets, t)
}
