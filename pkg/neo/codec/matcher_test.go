package codec

import (
	"github.com/casbin-mesh/neo/pkg/neo/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	mockMatcherData = &model.MatcherInfo{
		ID: 1,
		Name: model.CIStr{
			O: "Matcher",
			L: "matcher",
		},
		Raw:          "r.sub==p.sub && r.obj==p.obj",
		EffectPolicy: 1,
	}
)

func TestDecodeMatcherInfo(t *testing.T) {
	buf := EncodeMatcherInfo(mockMatcherData)
	decoded := DecodeMatcherInfo(buf, nil)
	assert.Equal(t, mockMatcherData, decoded)
}
