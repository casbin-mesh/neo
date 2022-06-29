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

package btuple

import (
	"encoding/binary"
	"github.com/casbin-mesh/neo/pkg/primitive/codec"
)

type Reader interface {
	ValueAt(pos int) Elem
	Occupied(pos int) bool
	buildHint() error
}

type mapping struct {
	offset uint32
	size   uint32 // it able to addresses 4GiB values
}

type bufferedReader struct {
	raw []byte
	len int
	mt  map[int]mapping
}

func (b *bufferedReader) buildHint() error {
	h := header{}
	h.decode(b.raw[:SizeOfHeader])
	idx := 0
	if h.typ == SmallValueType {
		offset := uint32(SizeOfHeader)
		for i := offset; i < uint32(len(b.raw)); i++ {
			if b.raw[i] == codec.NullTerminator {
				b.mt[idx] = mapping{
					offset: offset,
					size:   i - offset, // skip terminator
				}
				offset = i + 1 // skip terminator
				idx++
			}
		}
	} else if h.typ == LargeValueType {
		dataOffset := SizeOfHeader + 4*h.len
		for i := uint32(0); i < h.len; i++ {
			offset := dataOffset + binary.BigEndian.Uint32(b.raw[SizeOfHeader+i*4:SizeOfHeader+i*4+4])
			end := uint32(len(b.raw))
			if i != h.len-1 {
				end = dataOffset + binary.BigEndian.Uint32(b.raw[SizeOfHeader+i*4+4:SizeOfHeader+i*4+8])
			}
			b.mt[idx] = mapping{
				offset: offset,
				size:   end - offset - 1, // skip terminator
			}
			idx++
		}
	}
	h.len = uint32(idx - 1)
	return nil
}

// ValueAt return the value at position.
// NOTES: you should clone the return value, it doesn't check the bound.
func (b *bufferedReader) ValueAt(pos int) Elem {
	m := b.mt[pos]
	return b.raw[m.offset : m.offset+m.size]
}

func (b *bufferedReader) Occupied(pos int) bool {
	m := b.mt[pos]
	return m.size > 0
}

// NewReader return a tuple reader.
// NOTES: data should be immutable.
func NewReader(data []byte) (Reader, error) {
	r := &bufferedReader{raw: data, mt: make(map[int]mapping)}
	err := r.buildHint()
	return r, err
}
