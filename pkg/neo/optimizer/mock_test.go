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
	"github.com/casbin-mesh/neo/pkg/neo/model"
	"github.com/casbin-mesh/neo/pkg/primitive/bsontype"
)

var (
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
