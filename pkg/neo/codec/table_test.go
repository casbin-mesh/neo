package codec

import (
	"github.com/casbin-mesh/neo/pkg/neo/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	mockTableData = &model.TableInfo{
		ID: 1,
		Name: model.CIStr{
			L: "table",
			O: "Table",
		},
		Columns:     []*model.ColumnInfo{{ID: 1}, {ID: 2}, {ID: 3}},
		Indices:     []*model.IndexInfo{{ID: 4}, {ID: 5}},
		ForeignKeys: []*model.FKInfo{{ID: 6}},
	}
)

func TestDecodeTableInfo(t *testing.T) {
	buf := EncodeTableInfo(mockTableData)
	decoded := DecodeTableInfo(buf, nil)

	assert.Equal(t, mockTableData, decoded)
}
