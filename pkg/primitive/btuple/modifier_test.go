package btuple

import (
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
