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

func TestExpendPredicate(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		exp := parser.MustParseFromString("(a || b) && (c || d)")
		pred := NewPredicate(exp)
		fmt.Println(pred.String())
		expended := CNFPredicates(pred)
		fmt.Println(expended.String())
		expected := `( ( a && c ) || ( b && c ) || ( a && d ) || ( b && d ) )`
		assert.Equal(t, expected, expended.String())
	})
	t.Run("basic2", func(t *testing.T) {
		exp := parser.MustParseFromString(" (a || b) && (c || d) && (e || f)")
		pred := NewPredicate(exp)
		fmt.Println(pred.String())
		expended := CNFPredicates(pred)
		fmt.Println(expended.String())
		expected := `( ( a && c && e ) || ( b && c && e ) || ( a && d && e ) || ( b && d && e ) || ( a && c && f ) || ( b && c && f ) || ( a && d && f ) || ( b && d && f ) )`
		assert.Equal(t, expected, expended.String())
	})
	t.Run("basic3", func(t *testing.T) {
		exp := parser.MustParseFromString(" (a || b) && (c || d) || c")
		pred := NewPredicate(exp)
		fmt.Println(pred.String())
		expended := CNFPredicates(pred)
		fmt.Println(expended.String())
		expected := `( ( ( a && c ) || ( b && c ) || ( a && d ) || ( b && d ) ) || c )`
		assert.Equal(t, expected, expended.String())
	})
	t.Run("basic4", func(t *testing.T) {
		exp := parser.MustParseFromString("(a || b) && (c || d) && (e || f) || g")
		pred := NewPredicate(exp)
		fmt.Println(pred.String())
		expended := CNFPredicates(pred)
		fmt.Println(expended.String())
	})
	t.Run("basic5", func(t *testing.T) {
		exp := parser.MustParseFromString("r.sub == p.sub && (r.obj == p.obj || p.obj ==\"public\") && r.act == p.act || r.sub == \"root\"")
		pred := RewritePredicate(NewPredicate(exp))
		fmt.Println(pred.String())
		expended := CNFPredicates(pred)
		fmt.Println(expended.String())
		expected := `( ( ( (r.sub == p.sub) && (r.obj == p.obj) && (r.act == p.act) ) || ( (r.sub == p.sub) && (p.obj == "public") && (r.act == p.act) ) ) || (r.sub == "root") )`
		assert.Equal(t, expected, expended.String())
	})
	t.Run("basic6", func(t *testing.T) {
		exp := parser.MustParseFromString(" c || (d || f) && e")
		// c || (d || f) && e -> c || (d && e || f && e) -> c || d && e || f && e
		pred := NewPredicate(exp)
		fmt.Println(pred.String())
		expended := RewritePredicate(CNFPredicates(pred))
		fmt.Println(expended.String())
		expected := `( c || ( d && e ) || ( f && e ) )`
		assert.Equal(t, expected, expended.String())
	})
	t.Run("complex0", func(t *testing.T) {
		exp := parser.MustParseFromString(" (a || b) && (c || d || e != f && e != g) || c")
		pred := RewritePredicate(NewPredicate(exp))
		fmt.Println(pred.String())
		expended := RewritePredicate(CNFPredicates(pred))
		fmt.Println(expended.String())
		expected := `( ( a && c ) || ( b && c ) || ( a && d ) || ( b && d ) || ( a && (e != f) && (e != g) ) || ( b && (e != f) && (e != g) ) || c )`
		assert.Equal(t, expected, expended.String())
	})
	t.Run("complex1", func(t *testing.T) {
		exp := parser.MustParseFromString(` (r.subOwner == p.subOwner || p.subOwner == "*") && \
		    (r.subName == p.subName || p.subName == "*" || r.subName != "anonymous" && p.subName == "!anonymous") && \
		    (r.method == p.method || p.method == "*") && \
		    (r.urlPath == p.urlPath || p.urlPath == "*") && \
		    (r.objOwner == p.objOwner || p.objOwner == "*") && \
		    (r.objName == p.objName || p.objName == "*") || \
		    (r.subOwner == r.objOwner && r.subName == r.objName)`)
		pred := RewritePredicate(NewPredicate(exp))
		fmt.Println(pred.String())
		expended := CNFPredicates(pred)
		fmt.Println(expended.String())
		// ( ( ( (r.subOwner == p.subOwner) && (r.subName == p.subName) && (r.method == p.method) && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (p.subOwner == "*") && (r.subName == p.subName) && (r.method == p.method) && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && (p.subName == "*") && (r.method == p.method) && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (p.subOwner == "*") && (p.subName == "*") && (r.method == p.method) && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && ( (r.subName != "anonymous") && (p.subName == "!anonymous") ) && (r.method == p.method) && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (p.subOwner == "*") && ( (r.subName != "anonymous") && (p.subName == "!anonymous") ) && (r.method == p.method) && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && (r.subName == p.subName) && (p.method == "*") && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (p.subOwner == "*") && (r.subName == p.subName) && (p.method == "*") && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && (p.subName == "*") && (p.method == "*") && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (p.subOwner == "*") && (p.subName == "*") && (p.method == "*") && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && ( (r.subName != "anonymous") && (p.subName == "!anonymous") ) && (p.method == "*") && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (p.subOwner == "*") && ( (r.subName != "anonymous") && (p.subName == "!anonymous") ) && (p.method == "*") && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && (r.subName == p.subName) && (r.method == p.method) && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (p.subOwner == "*") && (r.subName == p.subName) && (r.method == p.method) && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && (p.subName == "*") && (r.method == p.method) && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (p.subOwner == "*") && (p.subName == "*") && (r.method == p.method) && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && ( (r.subName != "anonymous") && (p.subName == "!anonymous") ) && (r.method == p.method) && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (p.subOwner == "*") && ( (r.subName != "anonymous") && (p.subName == "!anonymous") ) && (r.method == p.method) && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && (r.subName == p.subName) && (p.method == "*") && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (p.subOwner == "*") && (r.subName == p.subName) && (p.method == "*") && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && (p.subName == "*") && (p.method == "*") && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (p.subOwner == "*") && (p.subName == "*") && (p.method == "*") && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && ( (r.subName != "anonymous") && (p.subName == "!anonymous") ) && (p.method == "*") && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (p.subOwner == "*") && ( (r.subName != "anonymous") && (p.subName == "!anonymous") ) && (p.method == "*") && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && (r.subName == p.subName) && (r.method == p.method) && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (p.subOwner == "*") && (r.subName == p.subName) && (r.method == p.method) && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && (p.subName == "*") && (r.method == p.method) && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (p.subOwner == "*") && (p.subName == "*") && (r.method == p.method) && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && ( (r.subName != "anonymous") && (p.subName == "!anonymous") ) && (r.method == p.method) && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (p.subOwner == "*") && ( (r.subName != "anonymous") && (p.subName == "!anonymous") ) && (r.method == p.method) && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && (r.subName == p.subName) && (p.method == "*") && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (p.subOwner == "*") && (r.subName == p.subName) && (p.method == "*") && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && (p.subName == "*") && (p.method == "*") && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (p.subOwner == "*") && (p.subName == "*") && (p.method == "*") && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && ( (r.subName != "anonymous") && (p.subName == "!anonymous") ) && (p.method == "*") && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (p.subOwner == "*") && ( (r.subName != "anonymous") && (p.subName == "!anonymous") ) && (p.method == "*") && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && (r.subName == p.subName) && (r.method == p.method) && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (p.subOwner == "*") && (r.subName == p.subName) && (r.method == p.method) && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && (p.subName == "*") && (r.method == p.method) && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (p.subOwner == "*") && (p.subName == "*") && (r.method == p.method) && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && ( (r.subName != "anonymous") && (p.subName == "!anonymous") ) && (r.method == p.method) && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (p.subOwner == "*") && ( (r.subName != "anonymous") && (p.subName == "!anonymous") ) && (r.method == p.method) && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && (r.subName == p.subName) && (p.method == "*") && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (p.subOwner == "*") && (r.subName == p.subName) && (p.method == "*") && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && (p.subName == "*") && (p.method == "*") && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (p.subOwner == "*") && (p.subName == "*") && (p.method == "*") && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && ( (r.subName != "anonymous") && (p.subName == "!anonymous") ) && (p.method == "*") && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (p.subOwner == "*") && ( (r.subName != "anonymous") && (p.subName == "!anonymous") ) && (p.method == "*") && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && (r.subName == p.subName) && (r.method == p.method) && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (p.subOwner == "*") && (r.subName == p.subName) && (r.method == p.method) && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && (p.subName == "*") && (r.method == p.method) && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (p.subOwner == "*") && (p.subName == "*") && (r.method == p.method) && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && ( (r.subName != "anonymous") && (p.subName == "!anonymous") ) && (r.method == p.method) && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (p.subOwner == "*") && ( (r.subName != "anonymous") && (p.subName == "!anonymous") ) && (r.method == p.method) && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && (r.subName == p.subName) && (p.method == "*") && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (p.subOwner == "*") && (r.subName == p.subName) && (p.method == "*") && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && (p.subName == "*") && (p.method == "*") && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (p.subOwner == "*") && (p.subName == "*") && (p.method == "*") && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && ( (r.subName != "anonymous") && (p.subName == "!anonymous") ) && (p.method == "*") && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (p.subOwner == "*") && ( (r.subName != "anonymous") && (p.subName == "!anonymous") ) && (p.method == "*") && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && (r.subName == p.subName) && (r.method == p.method) && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (p.subOwner == "*") && (r.subName == p.subName) && (r.method == p.method) && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && (p.subName == "*") && (r.method == p.method) && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (p.subOwner == "*") && (p.subName == "*") && (r.method == p.method) && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && ( (r.subName != "anonymous") && (p.subName == "!anonymous") ) && (r.method == p.method) && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (p.subOwner == "*") && ( (r.subName != "anonymous") && (p.subName == "!anonymous") ) && (r.method == p.method) && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && (r.subName == p.subName) && (p.method == "*") && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (p.subOwner == "*") && (r.subName == p.subName) && (p.method == "*") && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && (p.subName == "*") && (p.method == "*") && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (p.subOwner == "*") && (p.subName == "*") && (p.method == "*") && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && ( (r.subName != "anonymous") && (p.subName == "!anonymous") ) && (p.method == "*") && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (p.subOwner == "*") && ( (r.subName != "anonymous") && (p.subName == "!anonymous") ) && (p.method == "*") && (p.urlPath == "*") && (r.objOwner == p.objOwner) && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && (r.subName == p.subName) && (r.method == p.method) && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (p.subOwner == "*") && (r.subName == p.subName) && (r.method == p.method) && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && (p.subName == "*") && (r.method == p.method) && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (p.subOwner == "*") && (p.subName == "*") && (r.method == p.method) && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && ( (r.subName != "anonymous") && (p.subName == "!anonymous") ) && (r.method == p.method) && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (p.subOwner == "*") && ( (r.subName != "anonymous") && (p.subName == "!anonymous") ) && (r.method == p.method) && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && (r.subName == p.subName) && (p.method == "*") && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (p.subOwner == "*") && (r.subName == p.subName) && (p.method == "*") && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && (p.subName == "*") && (p.method == "*") && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (p.subOwner == "*") && (p.subName == "*") && (p.method == "*") && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && ( (r.subName != "anonymous") && (p.subName == "!anonymous") ) && (p.method == "*") && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (p.subOwner == "*") && ( (r.subName != "anonymous") && (p.subName == "!anonymous") ) && (p.method == "*") && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && (r.subName == p.subName) && (r.method == p.method) && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (p.subOwner == "*") && (r.subName == p.subName) && (r.method == p.method) && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && (p.subName == "*") && (r.method == p.method) && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (p.subOwner == "*") && (p.subName == "*") && (r.method == p.method) && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && ( (r.subName != "anonymous") && (p.subName == "!anonymous") ) && (r.method == p.method) && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (p.subOwner == "*") && ( (r.subName != "anonymous") && (p.subName == "!anonymous") ) && (r.method == p.method) && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && (r.subName == p.subName) && (p.method == "*") && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (p.subOwner == "*") && (r.subName == p.subName) && (p.method == "*") && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && (p.subName == "*") && (p.method == "*") && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (p.subOwner == "*") && (p.subName == "*") && (p.method == "*") && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (r.subOwner == p.subOwner) && ( (r.subName != "anonymous") && (p.subName == "!anonymous") ) && (p.method == "*") && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") )
		// || ( (p.subOwner == "*") && ( (r.subName != "anonymous") && (p.subName == "!anonymous") ) && (p.method == "*") && (p.urlPath == "*") && (p.objOwner == "*") && (p.objName == "*") ) )
		// || ( ( (r.subOwner == r.objOwner) && (r.subName == r.objName) ) ) )
	})
}
