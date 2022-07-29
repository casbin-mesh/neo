package codec

import "github.com/casbin-mesh/neo/pkg/primitive"

// TupleRecordKey t{tableId}_r{rid}
func TupleRecordKey(tableId uint64, rid primitive.ObjectID) []byte {
	buf := make([]byte, 0, 19)
	buf = append(buf, tablePrefix...)
	buf = appendUint64(buf, tableId)
	buf = append(buf, tupleRecordPrefix...)
	buf = append(buf, rid[:]...)
	return buf
}
