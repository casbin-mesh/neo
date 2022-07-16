// Copyright 2022 The casbin-mesh Authors. All Rights Reserved.
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

package executor

import (
	"context"
	"github.com/casbin-mesh/neo/pkg/neo/codec"
	"github.com/casbin-mesh/neo/pkg/neo/model"
	"github.com/casbin-mesh/neo/pkg/neo/session"
)

type SchemaExec struct {
	ctx session.Context
}

func NewSchemaExec(ctx session.Context) *SchemaExec {
	return &SchemaExec{
		ctx: ctx,
	}
}

func (s *SchemaExec) createDatabase(ctx context.Context, info *model.DBInfo) (dbId uint64, err error) {
	rw := s.ctx.GetMetaReaderWriter()
	if dbId, err = rw.NewDb(info.Name.L); err != nil {
		return
	}

	for _, matcherInfo := range info.MatcherInfo {
		if _, err = s.createMatcher(ctx, dbId, matcherInfo); err != nil {
			return dbId, err
		}
	}

	for _, tableInfo := range info.TableInfo {
		if _, err = s.createTable(ctx, dbId, tableInfo); err != nil {
			return dbId, err
		}
	}

	schemaRW := s.ctx.GetSchemaReaderWriter()
	err = schemaRW.Set(codec.DBInfoKey(dbId), info)
	if err != nil {
		return 0, err
	}

	return
}

func (s *SchemaExec) createTable(ctx context.Context, did uint64, info *model.TableInfo) (tableId uint64, err error) {
	rw := s.ctx.GetMetaReaderWriter()
	if tableId, err = rw.NewTable(did, info.Name.L); err != nil {
		return
	}
	info.ID = tableId

	for _, column := range info.Columns {
		if _, err = s.createColumn(ctx, tableId, column); err != nil {
			return
		}
	}

	for _, index := range info.Indices {
		if _, err = s.createIndex(ctx, tableId, index); err != nil {
			return
		}
	}

	//TODO(weny) :foreign keys

	return
}

func (s *SchemaExec) createColumn(ctx context.Context, tid uint64, info *model.ColumnInfo) (columnId uint64, err error) {
	rw := s.ctx.GetMetaReaderWriter()
	if columnId, err = rw.NewColumn(tid, info.Name.L); err != nil {
		return
	}
	info.ID = columnId

	return
}

func (s *SchemaExec) createIndex(ctx context.Context, tid uint64, info *model.IndexInfo) (indexId uint64, err error) {
	rw := s.ctx.GetMetaReaderWriter()
	if indexId, err = rw.NewIndex(tid, info.Name.L); err != nil {
		return
	}
	info.ID = indexId

	return
}

func (s *SchemaExec) createMatcher(ctx context.Context, did uint64, info *model.MatcherInfo) (matcherId uint64, err error) {
	rw := s.ctx.GetMetaReaderWriter()
	if matcherId, err = rw.NewMatcher(did, info.Name.L); err != nil {
		return
	}
	info.ID = matcherId

	return
}
