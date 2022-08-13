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

	// table name
	fb.CIStrStart(builder)
	fb.CIStrAddL(builder, builder.CreateString(info.Table.L))
	fb.CIStrAddO(builder, builder.CreateString(info.Table.O))
	tableName := fb.CIStrEnd(builder)

	fb.TableInfoStartColumnIdsVector(builder, len(info.Columns))
	for _, column := range info.Columns {
		fb.CIStrStart(builder)
		fb.CIStrAddL(builder, builder.CreateString(column.ColName.L))
		fb.CIStrAddO(builder, builder.CreateString(column.ColName.O))
		indexColName := fb.CIStrEnd(builder)
		fb.IndexColumnStart(builder)
		fb.IndexColumnAddName(builder, indexColName)
		fb.IndexColumnAddOffset(builder, int64(column.Offset))
		builder.PrependUOffsetT(fb.IndexColumnEnd(builder))
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
