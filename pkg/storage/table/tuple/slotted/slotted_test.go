package slotted

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"unsafe"
)

func TestNewSlotted(t *testing.T) {
	s := NewSlotted()
	data := (*byte)(unsafe.Pointer(s))
	println(data)
}

func TestSample(t *testing.T) {
	s := NewSlotted()
	data := []byte{0, 1, 2, 3}
	insert, err := s.Insert(data)
	assert.Nil(t, err)
	get, _, err := s.Get(insert)
	assert.Nil(t, err)
	assert.Equal(t, insert, uint16(1))
	assert.Equal(t, get, data)

	err = s.Delete(insert)
	assert.Nil(t, err)
	_, deleted, _ := s.Get(insert)
	assert.True(t, deleted)
}

func TestSlotted_RunOutSpace(t *testing.T) {
	s := NewSlotted()
	// total available 4088 B
	// data length 4
	// slot overhead 4
	data := []byte{0, 1, 2, 3}
	for i := 0; i < 4088/8; i++ {
		_, err := s.Insert(data)
		assert.Nil(t, err)
	}
	// should run out of space
	_, err := s.Insert(data)
	assert.Equal(t, ErrOutOfSpace, err)
}

func BenchmarkNewSlotted(b *testing.B) {
	NewSlotted()
}

func BenchmarkNativeNewSlotted(b *testing.B) {
	_ = slotted{}
}
