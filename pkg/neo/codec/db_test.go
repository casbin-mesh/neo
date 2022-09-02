package codec

import (
	"github.com/casbin-mesh/neo/pkg/neo/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	mockDBInfo = &model.DBInfo{
		ID: 1,
		Name: model.CIStr{
			L: "table",
			O: "Table",
		},
		TableInfo:   []*model.TableInfo{{ID: 1}, {ID: 2}},
		MatcherInfo: []*model.MatcherInfo{{ID: 3}, {ID: 4}},
	}
)

func TestDecodeBDInfo(t *testing.T) {
	buf := EncodeDBInfo(mockDBInfo)
	decoded := DecodeBDInfo(buf)
	assert.Equal(t, mockDBInfo, decoded)
}
