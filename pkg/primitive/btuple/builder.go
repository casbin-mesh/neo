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
	"errors"
	"github.com/casbin-mesh/neo/pkg/primitive/codec"
)

type Builder interface {
	Append(...Elem)
	Encode() []byte
	Write([]byte) (int, error)
	Size() int
}

type builder struct {
	tupleType BTupleType
	elems     []Elem
	offset    []uint32
	len       int
}

// Reset resets all fields excludes BTupleType
func (b *builder) Reset() {
	b.elems = nil
	b.offset = nil
	b.len = 0
}

func (b *builder) Size() int {
	return SizeOfHeader + len(b.offset)*4 + b.len
}

var (
	ErrOutOfSpace = errors.New("run out of space")
)

// writeTo encode BTuple, return written size
func (b *builder) writeTo(dst []byte) int {
	//TODO: determine binary tuple types
	writeTo := NewHeader(
		b.tupleType,          // tuple type
		uint32(len(b.elems)), // tuple count
	).writeTo(dst)

	if b.tupleType == LargeValueType {
		for i := 0; i < len(b.offset); i++ {
			binary.BigEndian.PutUint32(dst[writeTo:], b.offset[i])
			writeTo += 4
		}
	}
	for _, tuple := range b.elems {
		writeTo += copy(dst[writeTo:], tuple)
		dst[writeTo] = codec.NullTerminator
		writeTo += 1
	}
	return writeTo
}

func (b *builder) Write(dst []byte) (int, error) {
	if len(dst) < b.Size() {
		return 0, ErrOutOfSpace
	}
	w := b.writeTo(dst)
	return w, nil
}

func (b *builder) Encode() []byte {
	buf := make([]byte, b.Size())
	b.writeTo(buf)
	return buf
}

// Append elements to builder's buffer.
// NOTES: element should be immutable.
func (b *builder) Append(e ...Elem) {
	for _, elem := range e {
		if b.tupleType == LargeValueType {
			b.offset = append(b.offset, uint32(b.len)) // points to start
		}
		b.len += len(elem) + 1 // +1 for CString terminator
	}
	b.elems = append(b.elems, e...)
}

func (b *builder) SetTupleType(t BTupleType) {
	b.tupleType = t
}

func NewTupleBuilder(t BTupleType, tuple ...Elem) Builder {
	l := 0
	b := &builder{tupleType: t, len: l}
	b.Append(tuple...)
	return b
}
