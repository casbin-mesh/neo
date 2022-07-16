package codec

import (
	"github.com/casbin-mesh/neo/fb"
	"github.com/casbin-mesh/neo/pkg/neo/model"
	flatbuffers "github.com/google/flatbuffers/go"
)

// DBInfoKey t{id}
func DBInfoKey(dbId uint64) []byte {
	buf := make([]byte, 0, 9)
	buf = append(buf, databasePrefix...)
	appendUint64(buf, dbId)
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

	fb.DBInfoStart(builder)
	fb.DBInfoAddId(builder, info.ID)
	fb.DBInfoAddName(builder, name)

	// matcherIds
	fb.DBInfoStartMatcherIdsVector(builder, len(info.MatcherInfo))
	for _, matcher := range info.MatcherInfo {
		builder.PrependUint64(matcher.ID)
	}
	matcherIds := builder.EndVector(len(info.MatcherInfo))
	fb.DBInfoAddMatcherIds(builder, matcherIds)

	// tableIds
	fb.DBInfoStartTableIdsVector(builder, len(info.TableInfo))
	for _, table := range info.TableInfo {
		builder.PrependUint64(table.ID)
	}
	tableIds := builder.EndVector(len(info.MatcherInfo))
	fb.DBInfoAddMatcherIds(builder, tableIds)
	fb.DBInfoAddTableIds(builder, tableIds)

	orc := fb.DBInfoEnd(builder)
	builder.Finish(orc)
	return builder.FinishedBytes()
}
