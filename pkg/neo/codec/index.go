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

// IndexInfoKey s_i{id}
func IndexInfoKey(matcherId uint64) []byte {
	buf := make([]byte, 0, 11)
	buf = append(buf, mSchemaPrefix...)
	buf = append(buf, indexPrefixSep...)
	buf = appendUint64(buf, matcherId)
	return buf
}

func EncodeIndexInfo(info *model.IndexInfo) []byte {
	builder := flatbuffers.NewBuilder(1024)
	LName := builder.CreateString(info.Name.L)
	OName := builder.CreateString(info.Name.O)
	// name
	fb.CIStrStart(builder)
	fb.CIStrAddL(builder, LName)
	fb.CIStrAddO(builder, OName)
	name := fb.CIStrEnd(builder)

	tableLName := builder.CreateString(info.Table.L)
	tableOName := builder.CreateString(info.Table.O)
	// table name
	fb.CIStrStart(builder)
	fb.CIStrAddL(builder, tableLName)
	fb.CIStrAddO(builder, tableOName)
	tableName := fb.CIStrEnd(builder)

	colLNames := make([]flatbuffers.UOffsetT, len(info.Columns))
	colONames := make([]flatbuffers.UOffsetT, len(info.Columns))
	for i, column := range info.Columns {
		colLNames[i] = builder.CreateString(column.ColName.L)
		colONames[i] = builder.CreateString(column.ColName.O)
	}

	colNames := make([]flatbuffers.UOffsetT, len(info.Columns))
	for i, _ := range info.Columns {
		fb.CIStrStart(builder)
		fb.CIStrAddL(builder, colLNames[i])
		fb.CIStrAddO(builder, colONames[i])
		colNames[i] = fb.CIStrEnd(builder)
	}

	indexColumns := make([]flatbuffers.UOffsetT, len(info.Columns))
	for i, column := range info.Columns {
		fb.IndexColumnStart(builder)
		fb.IndexColumnAddName(builder, colNames[i])
		fb.IndexColumnAddOffset(builder, int64(column.Offset))
		indexColumns[i] = fb.IndexColumnEnd(builder)
	}

	fb.IndexInfoStartColumnsVector(builder, len(info.Columns))
	for _, id := range indexColumns {
		builder.PrependUOffsetT(id)
	}
	columns := builder.EndVector(len(info.Columns))

	fb.IndexInfoStart(builder)
	fb.IndexInfoAddId(builder, info.ID)
	fb.IndexInfoAddName(builder, name)
	fb.IndexInfoAddTableName(builder, tableName)
	fb.IndexInfoAddPrimary(builder, info.Primary)
	fb.IndexInfoAddUnique(builder, info.Unique)
	fb.IndexInfoAddColumns(builder, columns)

	orc := fb.IndexInfoEnd(builder)
	builder.Finish(orc)
	return builder.FinishedBytes()
}
