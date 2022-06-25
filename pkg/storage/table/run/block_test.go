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

package run

import (
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"testing"
)

func newBlock(size int) (source []byte, b *block) {
	source = make([]byte, size)
	b = NewBlock(source)
	return source, b
}

func TestNewBlock(t *testing.T) {
	_, b := newBlock(1024)
	assert.NotNil(t, b)
}

func intKeyGenerator(i int) []byte {
	buf := make([]byte, 8)
	base := 12345
	binary.PutUvarint(buf, uint64(base+i))
	return buf
}

func appendEntry(t *testing.T, b *block, count int, kg func(i int) []byte, vg func(i int) []byte) {
	for i := 0; i < count; i++ {
		k := kg(i)
		v := vg(i)
		err := b.AppendEntry(k, v)
		assert.Nil(t, err)
	}
}

func TestBlock_AppendEntry(t *testing.T) {
	size := 56 // 16B+4B entry * 2,  offset 4B * 2, baseKey 8B = 56B
	_, b := newBlock(size)
	appendEntry(t, b, 2, intKeyGenerator, intKeyGenerator)
}

func TestBlock_ValueAt(t *testing.T) {
	size := 56 // 16B+4B entry * 2,  offset 4B * 2, baseKey 8B = 56B
	_, b := newBlock(size)
	count := 2
	appendEntry(t, b, count, intKeyGenerator, intKeyGenerator)

	for i := 0; i < count; i++ {
		exp := intKeyGenerator(i)
		key, value := b.ValueAt(i)
		assert.Equal(t, exp, key)
		assert.Equal(t, exp, value)
	}
}

func TestBlock_Occupied(t *testing.T) {
	size := 56 // 16B+4B entry * 2,  offset 4B * 2, baseKey 8B = 56B
	_, b := newBlock(size)
	count := 2
	appendEntry(t, b, count, intKeyGenerator, intKeyGenerator)

	for i := 0; i < count; i++ {
		assert.True(t, b.Occupied(i))
	}
}
