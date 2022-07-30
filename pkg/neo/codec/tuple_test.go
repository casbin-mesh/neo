package codec

import (
	"github.com/casbin-mesh/neo/pkg/primitive"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseTupleRecordKey(t *testing.T) {
	oid := primitive.ObjectID{}
	key := TupleRecordKey(1, oid)
	got, err := ParseTupleRecordKey(key)
	assert.Nil(t, err)
	assert.Equal(t, oid, got)
}
