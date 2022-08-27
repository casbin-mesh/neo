package utils

import (
	"github.com/casbin-mesh/neo/pkg/primitive/bschema"
	"github.com/casbin-mesh/neo/pkg/primitive/btuple"
)

// MergeModifier returns merged modifier, NOTES: returned modifier's elements reference elements from s1, s2
func MergeModifier(m1 btuple.Modifier, s1 bschema.Reader, m2 btuple.Modifier, s2 bschema.Reader) (m btuple.Modifier, s bschema.ReaderWriter) {
	s = bschema.NewReaderWriteFormReader(s1)
	elems := make([]btuple.Elem, 0, s.FieldsLen())

	set := map[string]struct{}{}

	for i := 0; i < s.FieldsLen(); i++ {
		elems = append(elems, m1.ValueAt(i))
		set[string(s.FieldAt(i).Name())] = struct{}{}
	}

	for i := 0; i < s2.FieldsLen(); i++ {
		if _, ok := set[string(s2.FieldAt(i).Name())]; !ok {
			s.AppendFormField(s2.FieldAt(i))
			elems = append(elems, m2.ValueAt(i))
		}
	}

	m = btuple.NewModifier(elems)
	return
}

func MergeSchema(s1 bschema.Reader, s2 bschema.Reader) (s bschema.ReaderWriter) {
	s = bschema.NewReaderWriteFormReader(s1)
	set := map[string]struct{}{}

	for i := 0; i < s.FieldsLen(); i++ {
		set[string(s.FieldAt(i).Name())] = struct{}{}
	}

	for i := 0; i < s2.FieldsLen(); i++ {
		if _, ok := set[string(s2.FieldAt(i).Name())]; !ok {
			s.AppendFormField(s2.FieldAt(i))
		}
	}
	return
}
