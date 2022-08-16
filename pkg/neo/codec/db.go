package codec

import (
	"github.com/casbin-mesh/neo/fb"
	"github.com/casbin-mesh/neo/pkg/neo/model"
	flatbuffers "github.com/google/flatbuffers/go"
)

// DBInfoKey s_d{id}
func DBInfoKey(dbId uint64) []byte {
	buf := make([]byte, 0, 11)
	buf = append(buf, mSchemaPrefix...)
	buf = append(buf, databasePrefixSep...)
	buf = appendUint64(buf, dbId)
	return buf
}

func EncodeDBInfo(info *model.DBInfo) []byte {
	builder := flatbuffers.NewBuilder(1024)
	LName := builder.CreateString(info.Name.L)
	OName := builder.CreateString(info.Name.O)
	//  name
	fb.CIStrStart(builder)
	fb.CIStrAddL(builder, LName)
	fb.CIStrAddO(builder, OName)
	name := fb.CIStrEnd(builder)

	// matcherIds
	fb.DBInfoStartMatcherIdsVector(builder, len(info.MatcherInfo))
	for _, matcher := range info.MatcherInfo {
		builder.PrependUint64(matcher.ID)
	}
	matcherIds := builder.EndVector(len(info.MatcherInfo))

	// tableIds
	fb.DBInfoStartTableIdsVector(builder, len(info.TableInfo))
	for _, table := range info.TableInfo {
		builder.PrependUint64(table.ID)
	}
	tableIds := builder.EndVector(len(info.MatcherInfo))

	fb.DBInfoStart(builder)
	fb.DBInfoAddId(builder, info.ID)
	fb.DBInfoAddName(builder, name)
	fb.DBInfoAddMatcherIds(builder, matcherIds)
	fb.DBInfoAddMatcherIds(builder, tableIds)
	fb.DBInfoAddTableIds(builder, tableIds)

	orc := fb.DBInfoEnd(builder)
	builder.Finish(orc)
	return builder.FinishedBytes()
}
