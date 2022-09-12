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
	"github.com/casbin-mesh/neo/pkg/neo/model"
	"github.com/casbin-mesh/neo/pkg/primitive/bsontype"
	"github.com/casbin-mesh/neo/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	basic_model_table = []*model.TableInfo{{
		Name: model.NewCIStr("p"),
		Columns: []*model.ColumnInfo{
			{
				ColName: model.NewCIStr("sub"),
				Offset:  0,
				Tp:      bsontype.String,
			},
			{
				ColName: model.NewCIStr("obj"),
				Offset:  1,
				Tp:      bsontype.String,
			},
			{
				ColName: model.NewCIStr("act"),
				Offset:  2,
				Tp:      bsontype.String,
			},
		},
		Indices: []*model.IndexInfo{
			{
				Name:  model.NewCIStr("act_asc"),
				Table: model.NewCIStr("p"),
				Columns: []*model.IndexColumn{{
					ColName: model.NewCIStr("act"),
					Offset:  2,
				}},
			},
			{
				Name:  model.NewCIStr("obj_asc"),
				Table: model.NewCIStr("p"),
				Columns: []*model.IndexColumn{{
					ColName: model.NewCIStr("obj"),
					Offset:  1,
				}},
			},
			{
				Name:  model.NewCIStr("sub_asc"),
				Table: model.NewCIStr("p"),
				Columns: []*model.IndexColumn{{
					ColName: model.NewCIStr("sub"),
					Offset:  0,
				}},
			},
		},
	}}
	basic_without_resources_model_table = []*model.TableInfo{{
		Name: model.NewCIStr("p"),
		Columns: []*model.ColumnInfo{
			{
				ColName: model.NewCIStr("sub"),
				Offset:  0,
				Tp:      bsontype.String,
			},
			{
				ColName: model.NewCIStr("act"),
				Offset:  1,
				Tp:      bsontype.String,
			},
		},
		Indices: []*model.IndexInfo{
			{
				Name:  model.NewCIStr("act_asc"),
				Table: model.NewCIStr("p"),
				Columns: []*model.IndexColumn{{
					ColName: model.NewCIStr("act"),
					Offset:  1,
				}},
			},
			{
				Name:  model.NewCIStr("sub_asc"),
				Table: model.NewCIStr("p"),
				Columns: []*model.IndexColumn{{
					ColName: model.NewCIStr("sub"),
					Offset:  0,
				}},
			},
		},
	}}
	basic_model_without_users_table = []*model.TableInfo{{
		Name: model.NewCIStr("p"),
		Columns: []*model.ColumnInfo{
			{
				ColName: model.NewCIStr("obj"),
				Offset:  0,
				Tp:      bsontype.String,
			},
			{
				ColName: model.NewCIStr("act"),
				Offset:  1,
				Tp:      bsontype.String,
			},
		},
		Indices: []*model.IndexInfo{
			{
				Name:  model.NewCIStr("act_asc"),
				Table: model.NewCIStr("p"),
				Columns: []*model.IndexColumn{{
					ColName: model.NewCIStr("act"),
					Offset:  1,
				}},
			},
			{
				Name:  model.NewCIStr("obj_asc"),
				Table: model.NewCIStr("p"),
				Columns: []*model.IndexColumn{{
					ColName: model.NewCIStr("obj"),
					Offset:  0,
				}},
			},
		},
	}}
)

type testSet struct {
	modelPath string
	expected  []*model.TableInfo
}

func runTests(sets []testSet, t *testing.T) {
	for _, set := range sets {
		r, err := utils.ReadFile(set.modelPath)
		assert.Nil(t, err)
		g, err := NewIndexGenerator(r)
		assert.Nil(t, err)
		res := g.Generate()
		assert.Equal(t, set.expected, res)
	}
}

func TestIndexGenerator_Generate(t *testing.T) {
	sets := []testSet{
		{
			modelPath: "../../../../examples/assets/model/basic_model.conf",
			expected:  basic_model_table,
		},
		{
			modelPath: "../../../../examples/assets/model/basic_model_without_spaces.conf",
			expected:  basic_model_table,
		},
		{
			modelPath: "../../../../examples/assets/model/basic_with_root_model.conf",
			expected:  basic_model_table,
		},
		{
			modelPath: "../../../../examples/assets/model/comment_model.conf",
			expected:  basic_model_table,
		},
		{
			modelPath: "../../../../examples/assets/model/basic_without_resources_model.conf",
			expected:  basic_without_resources_model_table,
		},
		{
			modelPath: "../../../../examples/assets/model/basic_without_users_model.conf",
			expected:  basic_model_without_users_table,
		},
	}
	runTests(sets, t)
}
