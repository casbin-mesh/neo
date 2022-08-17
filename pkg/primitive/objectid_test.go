package primitive

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestObjectID_IsEmpty(t *testing.T) {
	assert.False(t, NewObjectID().IsEmpty())
	empty := ObjectID{}
	assert.True(t, empty.IsEmpty())
}
