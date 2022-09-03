package codec

import (
	"github.com/casbin-mesh/neo/fb"
	"github.com/casbin-mesh/neo/pkg/neo/model"
	"github.com/casbin-mesh/neo/pkg/primitive/bsontype"
	flatbuffers "github.com/google/flatbuffers/go"
)

// ColumnInfoKey s_c{id}
func ColumnInfoKey(columnId uint64) []byte {
	buf := make([]byte, 0, 11)
	buf = append(buf, mSchemaPrefix...)
	buf = append(buf, columnPrefixSep...)
	buf = appendUint64(buf, columnId)
	return buf
}

func EncodeColumnInfo(info *model.ColumnInfo) []byte {
	builder := flatbuffers.NewBuilder(1024)
	LName := builder.CreateString(info.ColName.L)
	OName := builder.CreateString(info.ColName.O)

	//name
	fb.CIStrStart(builder)
	fb.CIStrAddL(builder, LName)
	fb.CIStrAddO(builder, OName)
	name := fb.CIStrEnd(builder)

	defaultValueBit := builder.CreateByteString(info.DefaultValueBit)

	fb.ColumnInfoStart(builder)
	fb.ColumnInfoAddId(builder, info.ID)
	fb.ColumnInfoAddTp(builder, byte(info.Tp))
	fb.ColumnInfoAddDefaultValue(builder, defaultValueBit)
	fb.ColumnInfoAddName(builder, name)
	fb.ColumnInfoAddOffset(builder, int64(info.Offset))

	orc := fb.ColumnInfoEnd(builder)
	builder.Finish(orc)
	return builder.FinishedBytes()
}

func DecodeColumnInfo(buf []byte, dst *model.ColumnInfo) *model.ColumnInfo {
	if dst == nil {
		dst = &model.ColumnInfo{}
	}
	fbInfo := fb.GetRootAsColumnInfo(buf, 0)

	// ID
	dst.ID = fbInfo.Id()
	// col name
	name := fbInfo.Name(nil)
	dst.ColName.L = string(name.L())
	dst.ColName.O = string(name.O())
	// type
	dst.Tp = bsontype.Type(fbInfo.Tp())
	// default value
	dst.DefaultValueBit = fbInfo.DefaultValueBytes()
	//offset
	dst.Offset = int(fbInfo.Offset())
	return dst
}
