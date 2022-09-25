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
	"bufio"
	"errors"
	"fmt"
	"github.com/casbin-mesh/neo/pkg/expression"
	"github.com/casbin-mesh/neo/pkg/expression/ast"
	"github.com/casbin-mesh/neo/pkg/neo/model"
	"github.com/casbin-mesh/neo/pkg/neo/utils"
	"github.com/casbin-mesh/neo/pkg/parser"
	"github.com/casbin-mesh/neo/pkg/primitive/bsontype"
	"strings"
)

type Entity struct {
	Name model.CIStr
	Type bsontype.Type
}

func GetEntitiesLName(e []Entity) []string {
	name := make([]string, 0, len(e))
	for _, entity := range e {
		name = append(name, entity.Name.L)
	}
	return name
}

type Generator struct {
	rawModel  string
	request   []Entity
	policy    []Entity
	groups    map[string][]Entity
	eftPolicy model.EffectPolicyType
	predicate ast.Evaluable
	tables    []*model.TableInfo
}

var (
	ErrMultiRequestEntityDefinition = errors.New("multi-request entities definition found")
	ErrMultiPolicyEntityDefinition  = errors.New("multi-policy entities definition found")
	ErrMultiMatcherDefinition       = errors.New("multi matchers definition found")
	ErrInvalidRoleDefinition        = errors.New("invalid role definition")
)

func NewGeneratorFromString(s string) (*Generator, error) {
	return NewGenerator(bufio.NewReader(strings.NewReader(s)))
}

func NewGenerator(buf *bufio.Reader) (*Generator, error) {
	c, err := utils.NewParse(buf)
	if err != nil {
		return nil, err
	}
	if len(c.RequestDef()) != 1 {
		return nil, ErrMultiRequestEntityDefinition
	}
	if len(c.PolicyDef()) != 1 {
		return nil, ErrMultiPolicyEntityDefinition
	}
	if len(c.PolicyDef()) != 1 {
		return nil, ErrMultiMatcherDefinition
	}
	ig := &Generator{
		groups: map[string][]Entity{},
	}
	for _, s := range strings.Split(strings.ReplaceAll(c.RequestDef()["r"], " ", ""), ",") {
		ig.request = append(ig.request, Entity{
			Name: model.NewCIStr(s),
			Type: bsontype.String,
		})
	}
	for _, s := range strings.Split(strings.ReplaceAll(c.PolicyDef()["p"], " ", ""), ",") {
		ig.policy = append(ig.policy, Entity{
			Name: model.NewCIStr(s),
			Type: bsontype.String,
		})
	}
	for i, str := range c.RoleDef() {
		switch len(strings.Split(str, ",")) {
		case 2:
			ig.groups[i] = []Entity{
				{
					Name: model.NewCIStr("child"),
					Type: bsontype.String,
				},
				{
					Name: model.NewCIStr("parent"),
					Type: bsontype.String,
				},
			}
		case 3:
			ig.groups[i] = []Entity{
				{
					Name: model.NewCIStr("child"),
					Type: bsontype.String,
				},
				{
					Name: model.NewCIStr("parent"),
					Type: bsontype.String,
				},
				{
					Name: model.NewCIStr("tenant"),
					Type: bsontype.String,
				},
			}
		default:
			return nil, ErrInvalidRoleDefinition
		}
	}
	typ, err := model.NewEffectPolicyTypeFromString(c.PolicyEffect()["e"])
	if err != nil {
		return nil, err
	}
	ig.eftPolicy = typ
	tree, err := parser.TryParserFromString(c.Matchers()["m"])
	if err != nil {
		return nil, err
	}
	ig.predicate = tree
	return ig, nil
}

func EntitiesToColumns(e []Entity) []*model.ColumnInfo {
	cols := make([]*model.ColumnInfo, 0, len(e))
	for i, entity := range e {
		cols = append(cols, &model.ColumnInfo{
			ColName:         entity.Name,
			Offset:          i,
			Tp:              entity.Type,
			DefaultValue:    nil,
			DefaultValueBit: nil,
		})
	}
	return cols
}

func GetIndexInfos(col []*model.ColumnInfo, table model.CIStr, l ...string) (res []*model.IndexInfo) {
	for _, s := range l {
		for _, c := range col {
			if c.ColName.L == s {
				res = append(res, &model.IndexInfo{
					Name:    model.NewCIStr(fmt.Sprintf("%s_asc", c.ColName.O)),
					Table:   table,
					Columns: []*model.IndexColumn{{ColName: c.ColName, Offset: c.Offset}},
				})
			}
		}
	}
	return res
}

func columnInfo2IndexInfo(col []*model.ColumnInfo) ([]*model.IndexColumn, string) {
	result := make([]*model.IndexColumn, 0, len(col))
	names := make([]string, 0, len(col))
	for _, info := range col {
		result = append(result, &model.IndexColumn{
			ColName: info.ColName,
			Offset:  info.Offset,
		})
		names = append(names, info.ColName.O)
	}
	return result, strings.Join(names, "_")
}

// inverseColumns inverse first and second element in column
func inverseColumns(col []*model.ColumnInfo) []*model.ColumnInfo {
	if len(col) < 2 {
		return col
	}
	inverse := make([]*model.ColumnInfo, 0, len(col))
	for _, info := range col {
		inverse = append(inverse, info)
	}
	inverse[0], inverse[1] = inverse[1], inverse[0]
	return inverse
}

func generateGroupIndexInfo(col []*model.ColumnInfo, table model.CIStr) []*model.IndexInfo {
	res := make([]*model.IndexInfo, 0, 2)
	columns, name := columnInfo2IndexInfo(col)
	res = append(res, &model.IndexInfo{
		Name:    model.NewCIStr(fmt.Sprintf("%s_asc", name)),
		Table:   table,
		Columns: columns,
	})
	columns, name = columnInfo2IndexInfo(inverseColumns(col))
	res = append(res, &model.IndexInfo{
		Name:    model.NewCIStr(fmt.Sprintf("%s_asc", name)),
		Table:   table,
		Columns: columns,
	})

	return res
}

func (ig *Generator) GenerateTables() []*model.TableInfo {
	if ig.tables != nil {
		return ig.tables
	}
	// policy table
	policyTableName := model.NewCIStr("p")
	members := expression.GetAccessorMembers(ig.predicate)
	utils.SortStrings(members)
	policyMembers := GetEntitiesLName(ig.policy)
	utils.SortStrings(policyMembers)
	indexMembers := utils.SortedIntersect(members, policyMembers)
	columns := EntitiesToColumns(ig.policy)
	policyTable := &model.TableInfo{
		Name:    policyTableName,
		Columns: columns,
		Indices: GetIndexInfos(columns, policyTableName, indexMembers...),
	}
	ig.tables = append(ig.tables, policyTable)

	// groups
	for s, group := range ig.groups {
		groupTableName := model.NewCIStr(s)
		columns = EntitiesToColumns(group)
		groupTable := &model.TableInfo{
			Name:    groupTableName,
			Columns: columns,
			Indices: generateGroupIndexInfo(columns, groupTableName),
		}
		ig.tables = append(ig.tables, groupTable)
	}

	return ig.tables
}

func (ig *Generator) GenerateDB(db string) *model.DBInfo {
	dbInfo := &model.DBInfo{
		Name:        model.NewCIStr(db),
		TableInfo:   ig.GenerateTables(),
		MatcherInfo: []*model.MatcherInfo{ig.GenerateMatcher()},
	}
	return dbInfo
}

func (ig *Generator) GenerateMatcher() *model.MatcherInfo {
	return &model.MatcherInfo{
		ID:           0,
		Name:         model.NewCIStr("matcher"),
		Raw:          ig.rawModel,
		EffectPolicy: ig.eftPolicy,
		Predicate:    ig.predicate,
	}
}
