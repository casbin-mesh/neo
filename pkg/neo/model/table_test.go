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

package model

import (
	"github.com/casbin-mesh/neo/pkg/parser"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSelectAst(t *testing.T) {
	t.Run("basic policy", func(t *testing.T) {
		table := &TableInfo{
			Name: NewCIStr("p"),
			Columns: []*ColumnInfo{
				{
					ColName: NewCIStr("sub"),
				},
				{
					ColName: NewCIStr("obj"),
				},
				{
					ColName: NewCIStr("act"),
				},
				{
					ColName: NewCIStr("eft"),
				},
			},
		}
		tree := table.SelectAst("r")
		expected := parser.MustParseFromString("r.sub == p.sub && r.obj == p.obj && r.act == p.act && r.eft == p.eft")
		assert.Equal(t, expected, tree)
	})
	t.Run("basic group", func(t *testing.T) {
		table := &TableInfo{
			Name: NewCIStr("p"),
			Columns: []*ColumnInfo{
				{
					ColName: NewCIStr("child"),
				},
				{
					ColName: NewCIStr("parent"),
				},
			},
		}
		tree := table.SelectAst("r")
		expected := parser.MustParseFromString("r.child == p.child && r.parent == p.parent")
		assert.Equal(t, expected, tree)
	})
	t.Run("basic group with domain", func(t *testing.T) {
		table := &TableInfo{
			Name: NewCIStr("p"),
			Columns: []*ColumnInfo{
				{
					ColName: NewCIStr("child"),
				},
				{
					ColName: NewCIStr("parent"),
				},
				{
					ColName: NewCIStr("domain"),
				},
			},
		}
		tree := table.SelectAst("r")
		expected := parser.MustParseFromString("r.child == p.child && r.parent == p.parent && r.domain == p.domain")
		assert.Equal(t, expected, tree)
	})

}
