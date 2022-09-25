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

package session

import (
	"github.com/casbin-mesh/neo/pkg/expression/ast"
	"github.com/casbin-mesh/neo/pkg/neo/model"
)

type TableStatic interface {
	// GetColEstimatedCardinality returns count / the number of distinct value
	GetColEstimatedCardinality(col string) uint64
	// GetColCardinality returns static collected by CM sketch
	GetColCardinality(col string, value string) uint64
	// GetCount returns total row number
	GetCount() uint64
}

type CostModel interface {
	GetTableStatic(name string) TableStatic
}

type Base interface {
	PolicyTableName() string
	EffectColName() string
	AllowIdent() string
	DenyIdent() string
	ReqAccessorAncestorName() string
	Matcher() *model.MatcherInfo
	DB() *model.DBInfo
	Table() *model.TableInfo
}

type OptimizerCtx interface {
	Base
	CostModel
	ReqAccessor() ast.AccessorValue
}
