package utils

import (
	"github.com/casbin-mesh/neo/pkg/primitive/bschema"
	"github.com/casbin-mesh/neo/pkg/primitive/bsontype"
	"github.com/casbin-mesh/neo/pkg/primitive/btuple"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMergeModifier(t *testing.T) {

	t.Run("sample", func(t *testing.T) {
		s1 := bschema.NewReaderWriter()
		s1.Append(bsontype.String, []byte("suj"), nil)
		s1.Append(bsontype.String, []byte("obj"), nil)

		t1 := btuple.NewModifier([]btuple.Elem{btuple.Elem("alice"), btuple.Elem("data1")})

		s2 := bschema.NewReaderWriter()
		s2.Append(bsontype.String, []byte("act"), nil)
		s2.Append(bsontype.String, []byte("eft"), nil)

		t2 := btuple.NewModifier([]btuple.Elem{btuple.Elem("read"), btuple.Elem("allow")})

		m, s := MergeModifier(t1, s1, t2, s2)

		expected := []btuple.Elem{btuple.Elem("alice"), btuple.Elem("data1"), btuple.Elem("read"), btuple.Elem("allow")}
		for i, elem := range expected {
			assert.Equal(t, elem, m.ValueAt(i))
		}

		expectedS := [][]byte{[]byte("suj"), []byte("obj"), []byte("act"), []byte("eft")}
		for i, bytes := range expectedS {
			assert.Equal(t, bytes, s.FieldAt(i).Name())
		}
	})

	t.Run("overlapping", func(t *testing.T) {
		s1 := bschema.NewReaderWriter()
		s1.Append(bsontype.String, []byte("suj"), nil)
		s1.Append(bsontype.String, []byte("obj"), nil)
		s1.Append(bsontype.String, []byte("act"), nil)

		t1 := btuple.NewModifier([]btuple.Elem{btuple.Elem("alice"), btuple.Elem("data1"), btuple.Elem("read")})

		s2 := bschema.NewReaderWriter()
		s2.Append(bsontype.String, []byte("act"), nil)
		s2.Append(bsontype.String, []byte("eft"), nil)

		t2 := btuple.NewModifier([]btuple.Elem{btuple.Elem("read"), btuple.Elem("allow")})

		m, s := MergeModifier(t1, s1, t2, s2)

		expected := []btuple.Elem{btuple.Elem("alice"), btuple.Elem("data1"), btuple.Elem("read"), btuple.Elem("allow")}
		for i, elem := range expected {
			assert.Equal(t, elem, m.ValueAt(i))
		}

		expectedS := [][]byte{[]byte("suj"), []byte("obj"), []byte("act"), []byte("eft")}
		for i, bytes := range expectedS {
			assert.Equal(t, bytes, s.FieldAt(i).Name())
		}
	})
}
