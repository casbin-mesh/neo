package codec

import (
	"github.com/casbin-mesh/neo/pkg/neo/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	mockColumnInfo = &model.ColumnInfo{
		ID: 1,
		ColName: model.CIStr{
			O: "Sub",
			L: "sub",
		},
		Offset:          1,
		Tp:              2,
		DefaultValueBit: []byte("hi"),
	}
)

func TestDecodeColumnInfo(t *testing.T) {
	buf := EncodeColumnInfo(mockColumnInfo)
	decoded := DecodeColumnInfo(buf, nil)
	assert.Equal(t, mockColumnInfo, decoded)
}
