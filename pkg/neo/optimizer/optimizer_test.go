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

func TestOptimizer_Optimizer2(t *testing.T) {
	t.Run("basic expression without indexes", func(t *testing.T) {
		tree := parser.MustParseFromString("r.sub == p.sub && r.obj == p.obj && r.act == p.act")
		c := NewMockCtx(mockDbWithoutIndexes.MatcherInfo[0], mockDbWithoutIndexes, mockDbWithoutIndexes.TableInfo[0])
		lo := NewMatcherGenerator(c)
		o := NewOptimizer(c)
		output := o.Optimizer(lo.Generate(tree))
		expected := `MatcherPlan | Type: AllowOverride
└─LimitPlan | Limit:1
  └─SeqScanPlan | Predicate: ((((r.sub == p.sub) && (r.obj == p.obj)) && (r.act == p.act)) && (p.eft == "allow"))`
		assert.Equal(t, expected, output.String())
		fmt.Println(output)
	})
	t.Run("basic expression with indexes", func(t *testing.T) {
		tree := parser.MustParseFromString("r.sub == p.sub && r.obj == p.obj && r.act == p.act")
		c := NewMockCtxWithStatic(mockDbWithIndexes.MatcherInfo[0], mockDbWithIndexes, mockDbWithIndexes.TableInfo[0], staticModel{})
		c.SetReqAccessor(ast.NewMockAccessor(mockRequest))
		lo := NewMatcherGenerator(c)
		o := NewOptimizer(c)
		output := o.Optimizer(lo.Generate(tree))
		expected := `MatcherPlan | Type: AllowOverride
└─LimitPlan | Limit:1
  └─TableRowIdScan
    └─IndexScanPlan | Predicate: ((((r.sub == p.sub) && (r.obj == p.obj)) && (r.act == p.act)) && (p.eft == "allow"))`
		assert.Equal(t, expected, output.String())
		fmt.Println(output.String())
	})
	t.Run("root or expression with indexes", func(t *testing.T) {
		tree := parser.MustParseFromString("r.sub == p.sub && r.obj == p.obj && r.act == p.act || r.sub == \"root\"")
		c := NewMockCtx(mockDbWithIndexesAndAllowAndDenyMatcher.MatcherInfo[0], mockDbWithIndexesAndAllowAndDenyMatcher, mockDbWithIndexesAndAllowAndDenyMatcher.TableInfo[0])
		c.SetReqAccessor(ast.NewMockAccessor(mockRequest))
		lo := NewMatcherGenerator(c)
		o := NewOptimizer(c)
		output := o.Optimizer(lo.Generate(tree))
		expected := `MatcherPlan | Type: AllowAndDeny
├─LimitPlan | Limit:1
│ └─ShortCircuitPlan | Type: OR
│   ├─(Const)SeqScanPlan | Predicate: (r.sub == "root")
│   └─(Non-Const)TableRowIdScan
│     └─IndexScanPlan | Predicate: ((((r.sub == p.sub) && (r.obj == p.obj)) && (r.act == p.act)) && (p.eft == "allow"))
└─LimitPlan | Limit:1
  └─ShortCircuitPlan | Type: OR
    ├─(Const)SeqScanPlan | Predicate: (r.sub == "root")
    └─(Non-Const)TableRowIdScan
      └─IndexScanPlan | Predicate: ((((r.sub == p.sub) && (r.obj == p.obj)) && (r.act == p.act)) && (p.eft == "deny"))`
		assert.Equal(t, expected, output.String())
		fmt.Println(output.String())
	})
	t.Run("complex expression with indexes", func(t *testing.T) {
		tree := parser.MustParseFromString("r.sub == p.sub && (r.obj == p.obj || p.obj ==\"public\") && r.act == p.act || r.sub == \"root\"")
		c := NewMockCtx(mockDbWithIndexesAndAllowAndDenyMatcher.MatcherInfo[0], mockDbWithIndexesAndAllowAndDenyMatcher, mockDbWithIndexesAndAllowAndDenyMatcher.TableInfo[0])
		c.SetReqAccessor(ast.NewMockAccessor(mockRequest))
		lo := NewMatcherGenerator(c)
		o := NewOptimizer(c)
		output := o.Optimizer(lo.Generate(tree))
		expected := `MatcherPlan | Type: AllowAndDeny
├─LimitPlan | Limit:1
│ └─ShortCircuitPlan | Type: OR
│   ├─(Const)SeqScanPlan | Predicate: (r.sub == "root")
│   ├─(Non-Const)TableRowIdScan
│   │ └─IndexScanPlan | Predicate: ((((r.sub == p.sub) && (r.obj == p.obj)) && (r.act == p.act)) && (p.eft == "allow"))
│   └─(Non-Const)TableRowIdScan
│     └─IndexScanPlan | Predicate: ((((r.sub == p.sub) && (p.obj == "public")) && (r.act == p.act)) && (p.eft == "allow"))
└─LimitPlan | Limit:1
  └─ShortCircuitPlan | Type: OR
    ├─(Const)SeqScanPlan | Predicate: (r.sub == "root")
    ├─(Non-Const)TableRowIdScan
    │ └─IndexScanPlan | Predicate: ((((r.sub == p.sub) && (r.obj == p.obj)) && (r.act == p.act)) && (p.eft == "deny"))
    └─(Non-Const)TableRowIdScan
      └─IndexScanPlan | Predicate: ((((r.sub == p.sub) && (p.obj == "public")) && (r.act == p.act)) && (p.eft == "deny"))`
		assert.Equal(t, expected, output.String())
		fmt.Println(output.String())
	})
	t.Run("complex or expression with func and indexes", func(t *testing.T) {
		tree := parser.MustParseFromString("r.sub == p.sub && keyMatch(r.obj, p.obj) && r.act == p.act || isPublic(r.obj) || r.obj == \"public\" || r.sub == \"root\" ")
		c := NewMockCtx(mockDbWithIndexes.MatcherInfo[0], mockDbWithIndexes, mockDbWithIndexes.TableInfo[0])
		c.SetReqAccessor(ast.NewMockAccessor(mockRequest))
		lo := NewMatcherGenerator(c)
		o := NewOptimizer(c)
		output := o.Optimizer(lo.Generate(tree))
		expected := `MatcherPlan | Type: AllowOverride
└─LimitPlan | Limit:1
  └─ShortCircuitPlan | Type: OR
    ├─(Const)SeqScanPlan | Predicate: (r.obj == "public")
    ├─(Const)SeqScanPlan | Predicate: (r.sub == "root")
    ├─(Non-Const)TableRowIdScan
    │ └─IndexScanPlan | Predicate: ((((r.sub == p.sub) && keyMatch(r.obj, p.obj)) && (r.act == p.act)) && (p.eft == "allow"))
    └─(Non-Const)TableRowIdScan | Predicate: (p.eft == "allow")
      └─IndexScanPlan | Predicate: isPublic(r.obj)`
		assert.Equal(t, expected, output.String())
		fmt.Println(output.String())
	})
	t.Run("complex 2", func(t *testing.T) {
		/**
		(a || b) && (c || d) -> a && c || a && d || b && c || b && d
		(a || b) && (c || d) && (e || f) -> (a && c || a && d || b && c || b && d) && (e || f)



		(a || b || c) && (d || e) && (f || g) -> a && d && f || a && d && g || a && e && f || a && e && g ||
		(a || b || c) && (d || e || f) && (g || h) ->


		m = (r.subOwner == p.subOwner || p.subOwner == "*") && \
		    (r.subName == p.subName || p.subName == "*" || r.subName != "anonymous" && p.subName == "!anonymous") && \
		    (r.method == p.method || p.method == "*") && \
		    (r.urlPath == p.urlPath || p.urlPath == "*") && \
		    (r.objOwner == p.objOwner || p.objOwner == "*") && \
		    (r.objName == p.objName || p.objName == "*") || \
		    (r.subOwner == r.objOwner && r.subName == r.objName)
		*/
		tree := parser.MustParseFromString("(r.subOwner == p.subOwner || p.subOwner == \"*\") && \\\n    (r.subName == p.subName || p.subName == \"*\" || r.subName != \"anonymous\" && p.subName == \"!anonymous\") && \\\n    (r.method == p.method || p.method == \"*\") && \\\n    (r.urlPath == p.urlPath || p.urlPath == \"*\") && \\\n    (r.objOwner == p.objOwner || p.objOwner == \"*\") && \\\n    (r.objName == p.objName || p.objName == \"*\") || \\\n    (r.subOwner == r.objOwner && r.subName == r.objName)")
		c := NewMockCtx(mockDbWithIndexes.MatcherInfo[0], mockDbWithIndexes, mockDbWithIndexes.TableInfo[0])
		c.SetReqAccessor(ast.NewMockAccessor(mockRequest))
		lo := NewMatcherGenerator(c)
		o := NewOptimizer(c)
		output := o.Optimizer(lo.Generate(tree))
		expected := `MatcherPlan | Type: AllowOverride
└─LimitPlan | Limit:1
  └─ShortCircuitPlan | Type: OR
    ├─(Const)SeqScanPlan | Predicate: ((r.subOwner == r.objOwner) && (r.subName == r.objName))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((r.subOwner == p.subOwner) && (r.subName == p.subName)) && (r.method == p.method)) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((p.subOwner == "*") && (r.subName == p.subName)) && (r.method == p.method)) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((r.subOwner == p.subOwner) && (p.subName == "*")) && (r.method == p.method)) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((p.subOwner == "*") && (p.subName == "*")) && (r.method == p.method)) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: ((((((((r.subOwner == p.subOwner) && (r.subName != "anonymous")) && (p.subName == "!anonymous")) && (r.method == p.method)) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: ((((((((p.subOwner == "*") && (r.subName != "anonymous")) && (p.subName == "!anonymous")) && (r.method == p.method)) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((r.subOwner == p.subOwner) && (r.subName == p.subName)) && (p.method == "*")) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((p.subOwner == "*") && (r.subName == p.subName)) && (p.method == "*")) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((r.subOwner == p.subOwner) && (p.subName == "*")) && (p.method == "*")) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((p.subOwner == "*") && (p.subName == "*")) && (p.method == "*")) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: ((((((((r.subOwner == p.subOwner) && (r.subName != "anonymous")) && (p.subName == "!anonymous")) && (p.method == "*")) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: ((((((((p.subOwner == "*") && (r.subName != "anonymous")) && (p.subName == "!anonymous")) && (p.method == "*")) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((r.subOwner == p.subOwner) && (r.subName == p.subName)) && (r.method == p.method)) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((p.subOwner == "*") && (r.subName == p.subName)) && (r.method == p.method)) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((r.subOwner == p.subOwner) && (p.subName == "*")) && (r.method == p.method)) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((p.subOwner == "*") && (p.subName == "*")) && (r.method == p.method)) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: ((((((((r.subOwner == p.subOwner) && (r.subName != "anonymous")) && (p.subName == "!anonymous")) && (r.method == p.method)) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: ((((((((p.subOwner == "*") && (r.subName != "anonymous")) && (p.subName == "!anonymous")) && (r.method == p.method)) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((r.subOwner == p.subOwner) && (r.subName == p.subName)) && (p.method == "*")) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((p.subOwner == "*") && (r.subName == p.subName)) && (p.method == "*")) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((r.subOwner == p.subOwner) && (p.subName == "*")) && (p.method == "*")) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((p.subOwner == "*") && (p.subName == "*")) && (p.method == "*")) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: ((((((((r.subOwner == p.subOwner) && (r.subName != "anonymous")) && (p.subName == "!anonymous")) && (p.method == "*")) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: ((((((((p.subOwner == "*") && (r.subName != "anonymous")) && (p.subName == "!anonymous")) && (p.method == "*")) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((r.subOwner == p.subOwner) && (r.subName == p.subName)) && (r.method == p.method)) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((p.subOwner == "*") && (r.subName == p.subName)) && (r.method == p.method)) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((r.subOwner == p.subOwner) && (p.subName == "*")) && (r.method == p.method)) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((p.subOwner == "*") && (p.subName == "*")) && (r.method == p.method)) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: ((((((((r.subOwner == p.subOwner) && (r.subName != "anonymous")) && (p.subName == "!anonymous")) && (r.method == p.method)) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: ((((((((p.subOwner == "*") && (r.subName != "anonymous")) && (p.subName == "!anonymous")) && (r.method == p.method)) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((r.subOwner == p.subOwner) && (r.subName == p.subName)) && (p.method == "*")) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((p.subOwner == "*") && (r.subName == p.subName)) && (p.method == "*")) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((r.subOwner == p.subOwner) && (p.subName == "*")) && (p.method == "*")) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((p.subOwner == "*") && (p.subName == "*")) && (p.method == "*")) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: ((((((((r.subOwner == p.subOwner) && (r.subName != "anonymous")) && (p.subName == "!anonymous")) && (p.method == "*")) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: ((((((((p.subOwner == "*") && (r.subName != "anonymous")) && (p.subName == "!anonymous")) && (p.method == "*")) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((r.subOwner == p.subOwner) && (r.subName == p.subName)) && (r.method == p.method)) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((p.subOwner == "*") && (r.subName == p.subName)) && (r.method == p.method)) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((r.subOwner == p.subOwner) && (p.subName == "*")) && (r.method == p.method)) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((p.subOwner == "*") && (p.subName == "*")) && (r.method == p.method)) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: ((((((((r.subOwner == p.subOwner) && (r.subName != "anonymous")) && (p.subName == "!anonymous")) && (r.method == p.method)) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: ((((((((p.subOwner == "*") && (r.subName != "anonymous")) && (p.subName == "!anonymous")) && (r.method == p.method)) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((r.subOwner == p.subOwner) && (r.subName == p.subName)) && (p.method == "*")) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((p.subOwner == "*") && (r.subName == p.subName)) && (p.method == "*")) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((r.subOwner == p.subOwner) && (p.subName == "*")) && (p.method == "*")) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((p.subOwner == "*") && (p.subName == "*")) && (p.method == "*")) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: ((((((((r.subOwner == p.subOwner) && (r.subName != "anonymous")) && (p.subName == "!anonymous")) && (p.method == "*")) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: ((((((((p.subOwner == "*") && (r.subName != "anonymous")) && (p.subName == "!anonymous")) && (p.method == "*")) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((r.subOwner == p.subOwner) && (r.subName == p.subName)) && (r.method == p.method)) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((p.subOwner == "*") && (r.subName == p.subName)) && (r.method == p.method)) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((r.subOwner == p.subOwner) && (p.subName == "*")) && (r.method == p.method)) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((p.subOwner == "*") && (p.subName == "*")) && (r.method == p.method)) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: ((((((((r.subOwner == p.subOwner) && (r.subName != "anonymous")) && (p.subName == "!anonymous")) && (r.method == p.method)) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: ((((((((p.subOwner == "*") && (r.subName != "anonymous")) && (p.subName == "!anonymous")) && (r.method == p.method)) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((r.subOwner == p.subOwner) && (r.subName == p.subName)) && (p.method == "*")) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((p.subOwner == "*") && (r.subName == p.subName)) && (p.method == "*")) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((r.subOwner == p.subOwner) && (p.subName == "*")) && (p.method == "*")) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((p.subOwner == "*") && (p.subName == "*")) && (p.method == "*")) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: ((((((((r.subOwner == p.subOwner) && (r.subName != "anonymous")) && (p.subName == "!anonymous")) && (p.method == "*")) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: ((((((((p.subOwner == "*") && (r.subName != "anonymous")) && (p.subName == "!anonymous")) && (p.method == "*")) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((r.subOwner == p.subOwner) && (r.subName == p.subName)) && (r.method == p.method)) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((p.subOwner == "*") && (r.subName == p.subName)) && (r.method == p.method)) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((r.subOwner == p.subOwner) && (p.subName == "*")) && (r.method == p.method)) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((p.subOwner == "*") && (p.subName == "*")) && (r.method == p.method)) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: ((((((((r.subOwner == p.subOwner) && (r.subName != "anonymous")) && (p.subName == "!anonymous")) && (r.method == p.method)) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: ((((((((p.subOwner == "*") && (r.subName != "anonymous")) && (p.subName == "!anonymous")) && (r.method == p.method)) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((r.subOwner == p.subOwner) && (r.subName == p.subName)) && (p.method == "*")) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((p.subOwner == "*") && (r.subName == p.subName)) && (p.method == "*")) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((r.subOwner == p.subOwner) && (p.subName == "*")) && (p.method == "*")) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((p.subOwner == "*") && (p.subName == "*")) && (p.method == "*")) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: ((((((((r.subOwner == p.subOwner) && (r.subName != "anonymous")) && (p.subName == "!anonymous")) && (p.method == "*")) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: ((((((((p.subOwner == "*") && (r.subName != "anonymous")) && (p.subName == "!anonymous")) && (p.method == "*")) && (p.urlPath == "*")) && (r.objOwner == p.objOwner)) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((r.subOwner == p.subOwner) && (r.subName == p.subName)) && (r.method == p.method)) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((p.subOwner == "*") && (r.subName == p.subName)) && (r.method == p.method)) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((r.subOwner == p.subOwner) && (p.subName == "*")) && (r.method == p.method)) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((p.subOwner == "*") && (p.subName == "*")) && (r.method == p.method)) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: ((((((((r.subOwner == p.subOwner) && (r.subName != "anonymous")) && (p.subName == "!anonymous")) && (r.method == p.method)) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: ((((((((p.subOwner == "*") && (r.subName != "anonymous")) && (p.subName == "!anonymous")) && (r.method == p.method)) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((r.subOwner == p.subOwner) && (r.subName == p.subName)) && (p.method == "*")) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((p.subOwner == "*") && (r.subName == p.subName)) && (p.method == "*")) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((r.subOwner == p.subOwner) && (p.subName == "*")) && (p.method == "*")) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((p.subOwner == "*") && (p.subName == "*")) && (p.method == "*")) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: ((((((((r.subOwner == p.subOwner) && (r.subName != "anonymous")) && (p.subName == "!anonymous")) && (p.method == "*")) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: ((((((((p.subOwner == "*") && (r.subName != "anonymous")) && (p.subName == "!anonymous")) && (p.method == "*")) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((r.subOwner == p.subOwner) && (r.subName == p.subName)) && (r.method == p.method)) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((p.subOwner == "*") && (r.subName == p.subName)) && (r.method == p.method)) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((r.subOwner == p.subOwner) && (p.subName == "*")) && (r.method == p.method)) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((p.subOwner == "*") && (p.subName == "*")) && (r.method == p.method)) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: ((((((((r.subOwner == p.subOwner) && (r.subName != "anonymous")) && (p.subName == "!anonymous")) && (r.method == p.method)) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: ((((((((p.subOwner == "*") && (r.subName != "anonymous")) && (p.subName == "!anonymous")) && (r.method == p.method)) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((r.subOwner == p.subOwner) && (r.subName == p.subName)) && (p.method == "*")) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((p.subOwner == "*") && (r.subName == p.subName)) && (p.method == "*")) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((r.subOwner == p.subOwner) && (p.subName == "*")) && (p.method == "*")) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: (((((((p.subOwner == "*") && (p.subName == "*")) && (p.method == "*")) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    ├─(Non-Const)SeqScanPlan | Predicate: ((((((((r.subOwner == p.subOwner) && (r.subName != "anonymous")) && (p.subName == "!anonymous")) && (p.method == "*")) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))
    └─(Non-Const)SeqScanPlan | Predicate: ((((((((p.subOwner == "*") && (r.subName != "anonymous")) && (p.subName == "!anonymous")) && (p.method == "*")) && (p.urlPath == "*")) && (p.objOwner == "*")) && (p.objName == "*")) && (p.eft == "allow"))`
		assert.Equal(t, expected, output.String())
		fmt.Println(output.String())
	})
	t.Run("basic expression with indexes", func(t *testing.T) {
		tree := parser.MustParseFromString("r.sub == p.sub")
		c := NewMockCtx(mockDbWithIndexes.MatcherInfo[0], mockDbWithIndexes, mockDbWithIndexes.TableInfo[0])
		c.SetReqAccessor(ast.NewMockAccessor(mockRequest))
		op := &SelectPlanGenerator{ctx: c}
		plan := op.Generate(tree)
		o := NewOptimizer(c)
		output := o.Optimizer(plan)
		expected := `TableRowIdScan
└─IndexScanPlan | Predicate: (r.sub == p.sub)`
		assert.Equal(t, expected, output.String())
		fmt.Println(output.String())
	})

}
