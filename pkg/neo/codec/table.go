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

package codec

import (
	"github.com/casbin-mesh/neo/fb"
	"github.com/casbin-mesh/neo/pkg/neo/model"
	flatbuffers "github.com/google/flatbuffers/go"
)

// TableInfoKey s_t{id}
func TableInfoKey(tableId uint64) []byte {
	buf := make([]byte, 0, 11)
	buf = append(buf, mSchemaPrefix...)
	buf = append(buf, tablePrefixSep...)
	buf = appendUint64(buf, tableId)
	return buf
}

func EncodeTableInfo(info *model.TableInfo) []byte {
	builder := flatbuffers.NewBuilder(1024)
	LName := builder.CreateString(info.Name.L)
	OName := builder.CreateString(info.Name.O)

	// table name
	fb.CIStrStart(builder)
	fb.CIStrAddL(builder, LName)
	fb.CIStrAddO(builder, OName)
	tableName := fb.CIStrEnd(builder)

	// columnIds
	fb.TableInfoStartColumnIdsVector(builder, len(info.Columns))
	for _, column := range info.Columns {
		builder.PrependUint64(column.ID)
	}
	columnIds := builder.EndVector(len(info.Columns))

	// indexIds
	fb.TableInfoStartForeignKeyIdsVector(builder, len(info.Indices))
	for _, index := range info.Indices {
		builder.PrependUint64(index.ID)
	}
	indexIds := builder.EndVector(len(info.Indices))

	// fkInfoIds
	fb.TableInfoStartForeignKeyIdsVector(builder, len(info.ForeignKeys))
	for _, foreignKey := range info.ForeignKeys {
		builder.PrependUint64(foreignKey.ID)
	}
	fkInfoIds := builder.EndVector(len(info.ForeignKeys))

	fb.TableInfoStart(builder)
	fb.TableInfoAddId(builder, info.ID)
	fb.TableInfoAddName(builder, tableName)
	fb.TableInfoAddColumnIds(builder, columnIds)
	fb.TableInfoAddIndexIds(builder, indexIds)
	fb.TableInfoAddForeignKeyIds(builder, fkInfoIds)
	orc := fb.TableInfoEnd(builder)
	builder.Finish(orc)

	return builder.FinishedBytes()
}
