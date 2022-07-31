package btuple

import (
	"github.com/casbin-mesh/neo/pkg/primitive/bschema"
	"github.com/casbin-mesh/neo/pkg/primitive/bsontype"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestModifier_Set(t *testing.T) {
	m := NewModifier([]Elem{{1}, {2}, {3}})
	m.Set(1, Elem{3})
	assert.Equal(t, []Elem{{1}, {3}, {3}}, m.Values())
}

func TestModifier_Delete(t *testing.T) {
	m := NewModifier([]Elem{{1}, {2}, {3}})
	m.Delete(1)
	assert.Equal(t, []Elem{{1}, {3}}, m.Values())
}

func TestModifier_MergeDefaultValue(t *testing.T) {
	rw := bschema.NewReaderWriter(nil, nil)
	rw.Append(bsontype.String, []byte("sub"), nil)
	rw.Append(bsontype.String, []byte("obj"), nil)
	rw.Append(bsontype.String, []byte("act"), nil)
	rw.Append(bsontype.String, []byte("eft"), []byte("allow"))

	exp := NewModifier([]Elem{
		[]byte("alice"),
		[]byte(""),
		[]byte("read"),
		[]byte("allow"),
	})

	m := NewModifier([]Elem{
		[]byte("alice"),
		[]byte(""),
		[]byte("read"),
	})
	err := m.MergeDefaultValue(rw)
	assert.Nil(t, err)
	assert.Equal(t, exp, m)
}

func TestModifier_Clone(t *testing.T) {
	m := NewModifier([]Elem{
		[]byte("alice"),
		[]byte(""),
		[]byte("read"),
	})
	cloned := m.Clone()
	cloned.Append([]byte("test"))

	assert.NotEqual(t, m, cloned)
}
