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
	"github.com/casbin-mesh/neo/pkg/primitive"
	"github.com/casbin-mesh/neo/pkg/primitive/btuple"
	flatbuffers "github.com/google/flatbuffers/go"
	"hash/fnv"
)

// IndexInfoKey s_i{id}
func IndexInfoKey(matcherId uint64) []byte {
	buf := make([]byte, 0, 11)
	buf = append(buf, mSchemaPrefix...)
	buf = append(buf, indexPrefixSep...)
	buf = appendUint64(buf, matcherId)
	return buf
}

// PrimaryIndexEntryKey i{index_id}_{columns_value}}
func PrimaryIndexEntryKey(indexId uint64, columnValue []byte) []byte {
	buf := make([]byte, 0, 10+len(columnValue))
	buf = append(buf, indexPrefix...)
	buf = appendUint64(buf, indexId)
	buf = append(buf, Sep...)
	buf = append(buf, columnValue...)
	return buf
}

// SecondaryIndexEntryKey i{index_id}_{leftmost_column_value}_{r_id}
func SecondaryIndexEntryKey(indexId uint64, columnValue []byte, rId []byte) []byte {
	buf := make([]byte, 0, 19+len(columnValue))
	buf = append(buf, indexPrefix...)
	buf = appendUint64(buf, indexId)
	buf = append(buf, Sep...)
	buf = append(buf, columnValue...)
	buf = append(buf, Sep...)
	buf = append(buf, rId...)
	return buf
}

func IndexEntryKey(indexInfo *model.IndexInfo, columns []*model.ColumnInfo, tuple btuple.Reader, rid primitive.ObjectID) []byte {

	switch indexInfo.Tp {
	case model.SingleColumnIndex:
		var (
			leftmost []byte
		)
		index := indexInfo.Leftmost()
		col := columns[index.Offset]
		// retrieve actual value, then encode to mem-comparable format
		leftmost = EncodeCmpValue(DecodeValue(tuple.ValueAt(index.Offset), col.Tp))
		key := SecondaryIndexEntryKey(indexInfo.ID, leftmost, rid.Bytes())

		return key
	case model.HashIndex:

		h := fnv.New128()
		for _, index := range indexInfo.Columns {
			h.Write(tuple.ValueAt(index.Offset))
		}

		return SecondaryIndexEntryKey(indexInfo.ID, h.Sum(nil), rid.Bytes())
	}
	panic("unreachable")

}

func IndexEntry(indexInfo *model.IndexInfo, columns []*model.ColumnInfo, tuple btuple.Reader, rid primitive.ObjectID) (key, value []byte) {
	switch indexInfo.Tp {
	case model.SingleColumnIndex:
		tupleBuilder := btuple.NewTupleBuilder(btuple.SmallValueType)

		for _, index := range indexInfo.Columns {
			tupleBuilder.Append(tuple.ValueAt(index.Offset))
		}

		key = IndexEntryKey(indexInfo, columns, tuple, rid)
		return key, tupleBuilder.Encode()
	case model.HashIndex:
		key = IndexEntryKey(indexInfo, columns, tuple, rid)
		return key, nil
	}
	panic("unreachable")
}

func IndexEntries(index *model.IndexInfo, tuple btuple.Reader, rid primitive.ObjectID, iter func(key, value []byte) error) (err error) {
	for _, column := range index.Columns {
		var key, value []byte
		columnValue := (tuple).ValueAt(column.Offset)
		if index.Unique {
			key = PrimaryIndexEntryKey(index.ID, columnValue)
			value = rid[:]
		} else {
			key = SecondaryIndexEntryKey(index.ID, columnValue, rid[:])
		}
		if err = iter(key, value); err != nil {
			return err
		}
	}
	return
}

// ParseTupleRecordKeyFromSecondaryIndex parse r_id form i{index_id}_{index_column_value}_{r_id} form
func ParseTupleRecordKeyFromSecondaryIndex(b []byte) (primitive.ObjectID, error) {
	if len(b) < 19 {
		return primitive.ObjectID{}, ErrInvalidKey
	}
	if b[len(b)-9] != '_' {
		return primitive.ObjectID{}, ErrInvalidKey
	}

	data := [8]byte{}
	copy(data[:], b[len(b)-8:])
	return data, nil
}

func ParseTupleRecordKeyFromPrimaryIndex(b []byte) (primitive.ObjectID, error) {
	if len(b) != 8 {
		return primitive.ObjectID{}, ErrInvalidKey
	}
	data := [8]byte{}
	copy(data[:], b[:])
	return data, nil
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
	fb.IndexInfoAddTp(builder, byte(info.Tp))
	fb.IndexInfoAddUnique(builder, info.Unique)
	fb.IndexInfoAddColumns(builder, columns)

	orc := fb.IndexInfoEnd(builder)
	builder.Finish(orc)
	return builder.FinishedBytes()
}

func DecodeIndexInfo(buf []byte, dst *model.IndexInfo) *model.IndexInfo {
	if dst == nil {
		dst = &model.IndexInfo{}
	}
	fbInfo := fb.GetRootAsIndexInfo(buf, 0)

	// ID
	dst.ID = fbInfo.Id()
	// name
	name := fbInfo.Name(nil)
	dst.Name.L = string(name.L())
	dst.Name.O = string(name.O())
	// table name
	tableName := fbInfo.TableName(nil)
	dst.Table.L = string(tableName.L())
	dst.Table.O = string(tableName.O())
	// index type
	dst.Tp = model.IndexType(fbInfo.Tp())
	// primary
	dst.Primary = fbInfo.Primary()
	// unique
	dst.Unique = fbInfo.Unique()

	// info columns
	columnLen := fbInfo.ColumnsLength()
	dst.Columns = make([]*model.IndexColumn, 0, columnLen)
	for i := columnLen - 1; i >= 0; i-- {
		info := new(fb.IndexColumn)
		if fbInfo.Columns(info, i) {
			colName := info.Name(nil)
			dst.Columns = append(dst.Columns, &model.IndexColumn{
				ColName: model.CIStr{
					O: string(colName.O()),
					L: string(colName.L()),
				},
				Offset: int(info.Offset()),
			})
		}
	}

	return dst
}
