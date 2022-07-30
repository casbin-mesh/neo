package btuple

type Modifier interface {
	Reader
	Set(pos int, elem Elem)
	Delete(pos int)
}

type modifier struct {
	elems []Elem
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
