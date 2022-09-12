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
	"fmt"
	"github.com/casbin-mesh/neo/pkg/expression/ast"
	"github.com/casbin-mesh/neo/pkg/neo/optimizer/node"
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
	expected node.Predicate
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
