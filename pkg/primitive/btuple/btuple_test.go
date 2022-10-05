// Copyright 2022 The casbin-mesh Authors. All Rights Reserved.
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

package btuple

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func newValueBTuple(t assert.TestingT, typ BTupleType, e ...Elem) []byte {
	b := NewTupleBuilder(typ, e...)
	assert.NotNil(t, b)
	encoded := b.Encode()
	return encoded
}

func TestBufferedReader_Test(t *testing.T) {
	t.Run("SmallValueType", func(t *testing.T) {
		e := []Elem{Elem("Alice"), Elem("data1"), Elem("read")}
		buf := newValueBTuple(t, SmallValueType, e...)
		reader, err := NewReader(buf)
		assert.NotNil(t, reader)
		assert.Nil(t, err)
		for i := 0; i < len(e); i++ {
			assert.Equal(t, e[i], reader.ValueAt(i))
			assert.True(t, reader.Occupied(i))
		}
	})
	t.Run("SmallValueType2", func(t *testing.T) {
		e := []Elem{Elem(""), Elem(""), Elem("")}
		buf := newValueBTuple(t, SmallValueType, e...)
		reader, err := NewReader(buf)
		assert.NotNil(t, reader)
		assert.Nil(t, err)
		for i := 0; i < len(e); i++ {
			assert.Equal(t, e[i], reader.ValueAt(i))
			assert.True(t, reader.Occupied(i))
		}
	})
	t.Run("LargeValueType", func(t *testing.T) {
		e := []Elem{Elem("Alice"), Elem("data1"), Elem("read")}
		buf := newValueBTuple(t, LargeValueType, e...)
		reader, err := NewReader(buf)
		assert.NotNil(t, reader)
		assert.Nil(t, err)
		for i := 0; i < len(e); i++ {
			assert.Equal(t, e[i], reader.ValueAt(i))
			assert.True(t, reader.Occupied(i))
		}
	})
	t.Run("LargeValueType2", func(t *testing.T) {
		e := []Elem{Elem(""), Elem(""), Elem("")}
		buf := newValueBTuple(t, LargeValueType, e...)
		reader, err := NewReader(buf)
		assert.NotNil(t, reader)
		assert.Nil(t, err)
		for i := 0; i < len(e); i++ {
			assert.Equal(t, e[i], reader.ValueAt(i))
			assert.True(t, reader.Occupied(i))
		}
	})
}

func BenchmarkBufferedReader_ValueAt(b *testing.B) {
	e := []Elem{Elem("Alice"), Elem("data1"), Elem("read")}
	buf := newValueBTuple(b, LargeValueType, e...)
	reader, err := NewReader(buf)
	assert.Nil(b, err)
	b.ResetTimer()
	for i := 0; i < len(e); i++ {
		reader.ValueAt(i)
	}
}

func BenchmarkMap_ValueAt(b *testing.B) {
	e := []Elem{Elem("Alice"), Elem("data1"), Elem("read")}
	data := make(map[int]Elem)
	for i, elem := range e {
		data[i] = elem
	}
	b.ResetTimer()
	for i := 0; i < len(e); i++ {
		_ = data[i]
	}
}
