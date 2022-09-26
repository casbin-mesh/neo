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
	"github.com/casbin-mesh/neo/pkg/neo/model"
	"github.com/casbin-mesh/neo/pkg/neo/session"
	"github.com/casbin-mesh/neo/pkg/primitive/bsontype"
)

var (
	mockRequest          = map[string]string{"sub": "alice", "obj": "data1", "act": "read"}
	mockDbWithoutIndexes = &model.DBInfo{
		ID: 1,
		Name: model.CIStr{
			O: "Test",
			L: "test",
		},
		TableInfo: []*model.TableInfo{
			{
				ID: 1,
				Name: model.CIStr{
					O: "policy",
					L: "policy",
				},
				Columns: []*model.ColumnInfo{
					{
						// ID: 1,
						ColName: model.CIStr{
							O: "sub",
							L: "sub",
						},
						Offset: 0,
						Tp:     bsontype.String,
					},
					{
						// ID: 2,
						ColName: model.CIStr{
							O: "obj",
							L: "obj",
						},
						Offset: 1,
						Tp:     bsontype.String,
					},
					{
						// ID: 3,
						ColName: model.CIStr{
							O: "act",
							L: "act",
						},
						Offset: 3,
						Tp:     bsontype.String,
					},
					{
						// ID: 4,
						ColName: model.CIStr{
							O: "eft",
							L: "eft",
						},
						Offset:          4,
						Tp:              bsontype.String,
						DefaultValueBit: []byte("allow"),
					},
				},
			},
			{
				ID: 2,
				Name: model.CIStr{
					O: "group",
					L: "group",
				},
				Columns: []*model.ColumnInfo{
					{
						// ID: 5,
						ColName: model.CIStr{
							O: "member",
							L: "member",
						},
						Tp: bsontype.String,
					},
					{
						// ID: 6,
						ColName: model.CIStr{
							O: "group",
							L: "group",
						},
						Tp: bsontype.String,
					},
					{
						// ID: 7,
						ColName: model.CIStr{
							O: "domain",
							L: "domain",
						},
						DefaultValueBit: []byte("default"),
						Tp:              bsontype.String,
					},
				},
			},
		},
		MatcherInfo: []*model.MatcherInfo{
			{
				ID: 1,
				Name: model.CIStr{
					O: "matcher",
					L: "matcher",
				},
				EffectPolicy: model.AllowOverride,
			},
		},
	}
	mockDbWithIndexes = &model.DBInfo{
		ID: 1,
		Name: model.CIStr{
			O: "Test",
			L: "test",
		},
		TableInfo: []*model.TableInfo{
			{
				ID: 1,
				Name: model.CIStr{
					O: "policy",
					L: "policy",
				},
				Columns: []*model.ColumnInfo{
					{
						// ID: 1,
						ColName: model.CIStr{
							O: "sub",
							L: "sub",
						},
						Offset: 0,
						Tp:     bsontype.String,
					},
					{
						// ID: 2,
						ColName: model.CIStr{
							O: "obj",
							L: "obj",
						},
						Offset: 1,
						Tp:     bsontype.String,
					},
					{
						// ID: 3,
						ColName: model.CIStr{
							O: "act",
							L: "act",
						},
						Offset: 3,
						Tp:     bsontype.String,
					},
					{
						// ID: 4,
						ColName: model.CIStr{
							O: "eft",
							L: "eft",
						},
						Offset:          4,
						Tp:              bsontype.String,
						DefaultValueBit: []byte("allow"),
					},
				},
				Indices: []*model.IndexInfo{
					{
						ID:   1,
						Name: model.CIStr{O: "subject_index", L: "subject_index"},
						Columns: []*model.IndexColumn{
							{
								ColName: model.CIStr{O: "sub", L: "sub"},
								Offset:  0,
							},
						},
					},
					{
						ID:   2,
						Name: model.CIStr{O: "object_index", L: "object_index"},
						Columns: []*model.IndexColumn{
							{
								ColName: model.CIStr{O: "obj", L: "obj"},
								Offset:  1,
							},
						},
					},
					{
						ID:   3,
						Name: model.CIStr{O: "action_index", L: "action_index"},
						Columns: []*model.IndexColumn{
							{
								ColName: model.CIStr{O: "act", L: "act"},
								Offset:  2,
							},
						},
					},
				},
			},
			{
				ID: 2,
				Name: model.CIStr{
					O: "group",
					L: "group",
				},
				Columns: []*model.ColumnInfo{
					{
						// ID: 5,
						ColName: model.CIStr{
							O: "member",
							L: "member",
						},
						Tp: bsontype.String,
					},
					{
						// ID: 6,
						ColName: model.CIStr{
							O: "group",
							L: "group",
						},
						Tp: bsontype.String,
					},
					{
						// ID: 7,
						ColName: model.CIStr{
							O: "domain",
							L: "domain",
						},
						DefaultValueBit: []byte("default"),
						Tp:              bsontype.String,
					},
				},
			},
		},
		MatcherInfo: []*model.MatcherInfo{
			{
				ID: 1,
				Name: model.CIStr{
					O: "matcher",
					L: "matcher",
				},
				EffectPolicy: model.AllowOverride,
			},
		},
	}
	mockDbWithIndexesAndAllowAndDenyMatcher = &model.DBInfo{
		ID: 1,
		Name: model.CIStr{
			O: "Test",
			L: "test",
		},
		TableInfo: []*model.TableInfo{
			{
				ID: 1,
				Name: model.CIStr{
					O: "policy",
					L: "policy",
				},
				Columns: []*model.ColumnInfo{
					{
						// ID: 1,
						ColName: model.CIStr{
							O: "sub",
							L: "sub",
						},
						Offset: 0,
						Tp:     bsontype.String,
					},
					{
						// ID: 2,
						ColName: model.CIStr{
							O: "obj",
							L: "obj",
						},
						Offset: 1,
						Tp:     bsontype.String,
					},
					{
						// ID: 3,
						ColName: model.CIStr{
							O: "act",
							L: "act",
						},
						Offset: 3,
						Tp:     bsontype.String,
					},
					{
						// ID: 4,
						ColName: model.CIStr{
							O: "eft",
							L: "eft",
						},
						Offset:          4,
						Tp:              bsontype.String,
						DefaultValueBit: []byte("allow"),
					},
				},
				Indices: []*model.IndexInfo{
					{
						ID:   1,
						Name: model.CIStr{O: "subject_index", L: "subject_index"},
						Columns: []*model.IndexColumn{
							{
								ColName: model.CIStr{O: "sub", L: "sub"},
								Offset:  0,
							},
						},
					},
					{
						ID:   2,
						Name: model.CIStr{O: "object_index", L: "object_index"},
						Columns: []*model.IndexColumn{
							{
								ColName: model.CIStr{O: "obj", L: "obj"},
								Offset:  1,
							},
						},
					},
					{
						ID:   3,
						Name: model.CIStr{O: "action_index", L: "action_index"},
						Columns: []*model.IndexColumn{
							{
								ColName: model.CIStr{O: "act", L: "act"},
								Offset:  2,
							},
						},
					},
				},
			},
			{
				ID: 2,
				Name: model.CIStr{
					O: "group",
					L: "group",
				},
				Columns: []*model.ColumnInfo{
					{
						// ID: 5,
						ColName: model.CIStr{
							O: "member",
							L: "member",
						},
						Tp: bsontype.String,
					},
					{
						// ID: 6,
						ColName: model.CIStr{
							O: "group",
							L: "group",
						},
						Tp: bsontype.String,
					},
					{
						// ID: 7,
						ColName: model.CIStr{
							O: "domain",
							L: "domain",
						},
						DefaultValueBit: []byte("default"),
						Tp:              bsontype.String,
					},
				},
			},
		},
		MatcherInfo: []*model.MatcherInfo{
			{
				ID: 1,
				Name: model.CIStr{
					O: "matcher",
					L: "matcher",
				},
				EffectPolicy: model.AllowAndDeny,
			},
		},
	}
)

type mockCtx struct {
	matcher *model.MatcherInfo
	db      *model.DBInfo
	table   *model.TableInfo
	stats   staticModel
	req     ast.AccessorValue
}

func (m *mockCtx) SetReqAccessor(a ast.AccessorValue) {
	m.req = a
}

func (m mockCtx) ReqAccessor() ast.AccessorValue {
	return m.req
}

func (m mockCtx) PolicyTableName() string {
	return "p"
}

func (m mockCtx) EffectColName() string {
	return "eft"
}

func (m mockCtx) AllowIdent() string {
	return "allow"
}

func (m mockCtx) DenyIdent() string {
	return "deny"
}

func (m mockCtx) ReqAccessorAncestorName() string {
	return "r"
}

func (m mockCtx) Matcher() *model.MatcherInfo {
	return m.matcher
}

func (m mockCtx) DB() *model.DBInfo {
	return m.db
}

func (m mockCtx) Table() *model.TableInfo {
	return m.table
}

func (m mockCtx) GetTableStatic(name string) session.TableStatic {
	return nil
}

func NewMockCtx(matcher *model.MatcherInfo, db *model.DBInfo, table *model.TableInfo) *mockCtx {
	return &mockCtx{
		matcher: matcher, db: db, table: table,
	}
}

func NewMockCtxWithStatic(matcher *model.MatcherInfo, db *model.DBInfo, table *model.TableInfo, stats staticModel) *mockCtx {
	return &mockCtx{
		matcher: matcher, db: db, table: table, stats: stats,
	}
}

type colStatic struct {
	stats map[string]uint64
	ndv   uint64
}

type tableStatic struct {
	stats map[string]colStatic
	total uint64
}

type staticModel struct {
	stats map[string]tableStatic
}

func (s staticModel) GetTableStatic(table string) session.TableStatic {
	return s.stats[table]
}

func (s tableStatic) GetColEstimatedCardinality(col string) uint64 {
	return s.total / s.stats[col].ndv
}

func (s tableStatic) GetColCardinality(col string, value string) uint64 {
	return s.stats[col].stats[value]
}

func (s tableStatic) GetCount() uint64 {
	return s.total
}
