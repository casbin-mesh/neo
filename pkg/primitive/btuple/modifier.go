package btuple

import (
	"errors"
	"github.com/casbin-mesh/neo/pkg/primitive/bschema"
)

var ErrInvalidSchema = errors.New("invalid schema")

type Modifier interface {
	Reader
	Set(pos int, elem Elem)
	Append(elem Elem)
	Delete(pos int)
	MergeDefaultValue(schema bschema.Reader) error
	Clone() Modifier
}

type modifier struct {
	elems []Elem
}

func (m *modifier) Clone() Modifier {
	return &modifier{elems: append([]Elem{}, m.elems...)}
}

func (m *modifier) MergeDefaultValue(schema bschema.Reader) error {
	ll, rl := len(m.elems), schema.FieldsLen()

	if ll > rl {
		return ErrInvalidSchema
	}
	for i := 0; i < rl; i++ {
		if i < ll {
			if len(m.elems[i]) == 0 && len(schema.FieldAt(i).GetDefaultValue()) != 0 {
				m.elems[i] = schema.FieldAt(i).GetDefaultValue()
			}
		} else {
			m.Append(schema.FieldAt(i).GetDefaultValue())
		}
	}
	return nil
}

func (m *modifier) Append(elem Elem) {
	m.elems = append(m.elems, elem)
}

func (m *modifier) Set(pos int, elem Elem) {
	m.elems[pos] = elem
}

func (m *modifier) Delete(pos int) {
	m.elems = append(m.elems[:pos], m.elems[pos+1:]...)
}

func (m *modifier) Values() []Elem {
	return m.elems
}

func (m *modifier) ValueAt(pos int) Elem {
	return m.elems[pos]
}

func (m *modifier) Occupied(pos int) bool {
	return pos >= 0 && pos < len(m.elems)
}

func NewModifier(elems []Elem) Modifier {
	return &modifier{elems: elems}
}

func NewModifierFromBytes(elems [][]byte) Modifier {
	e := make([]Elem, len(elems))
	for i, elem := range elems {
		e[i] = elem
	}
	return &modifier{elems: e}
}
