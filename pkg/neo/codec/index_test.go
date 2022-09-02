package codec

import (
	"encoding/binary"
	"fmt"
	"github.com/casbin-mesh/neo/pkg/neo/model"
	"github.com/casbin-mesh/neo/pkg/primitive"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPrimaryIndexEntryKey(t *testing.T) {
	id := uint64(1)
	bid := [8]byte{}
	binary.BigEndian.PutUint64(bid[:], id)
	assert.Equal(t, []byte(fmt.Sprintf("i%s_hello", bid)), PrimaryIndexEntryKey(1, []byte("hello")))
}

func TestSecondaryIndexEntryKey(t *testing.T) {
	id := uint64(1)
	bid := [8]byte{}
	binary.BigEndian.PutUint64(bid[:], id)
	assert.Equal(t, []byte(fmt.Sprintf("i%s_hello_%s", bid, bid)), SecondaryIndexEntryKey(1, []byte("hello"), bid[:]))
}

func TestParseTupleRecordKeyFromSecondaryIndex(t *testing.T) {
	bid := primitive.NewObjectID()
	key := SecondaryIndexEntryKey(1, []byte("hello"), bid[:])

	oid, err := ParseTupleRecordKeyFromSecondaryIndex(key)
	assert.Nil(t, err)
	assert.Equal(t, bid, oid)
}

var (
	mockIndexInfoData = &model.IndexInfo{
		ID: 1,
		Name: model.CIStr{
			O: "SUB_INDEX",
			L: "sub_index",
		},
		Table: model.CIStr{
			O: "Sub",
			L: "sub",
		},
		Columns: []*model.IndexColumn{
			{
				ColName: model.CIStr{
					O: "Sub",
					L: "sub",
				},
				Offset: 0,
			},
			{
				ColName: model.CIStr{
					O: "Act",
					L: "act",
				},
				Offset: 2,
			},
			{
				ColName: model.CIStr{
					O: "Eft",
					L: "Eft",
				},
				Offset: 2,
			},
		},
		Unique:  false,
		Primary: false,
		Tp:      1,
	}
)

func TestDecodeIndexInfo(t *testing.T) {
	buf := EncodeIndexInfo(mockIndexInfoData)
	decoded := DecodeIndexInfo(buf, nil)
	assert.Equal(t, mockIndexInfoData, decoded)
}
