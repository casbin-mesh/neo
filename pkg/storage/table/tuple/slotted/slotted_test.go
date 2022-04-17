// Copyright 2022 The casbin-neo Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
