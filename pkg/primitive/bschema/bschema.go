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

package bschema

import (
	"github.com/casbin-mesh/neo/pkg/primitive/bsontype"
)

type BSchema interface {
	Key() []byte
	// ValueCopy returns a copy of the value of the item from the value bytes, writing it to dst slice.
	// If nil is passed, or capacity of dst isn't sufficient, a new slice would be allocated and
	// returned.
	ValueCopy(dst []byte) []byte
}

type ReaderWriter interface {
	Writer
	Reader
}

type Writer interface {
	EncodeVal() []byte
	EncodeKey() []byte
}

type Reader interface {
	DecodeVal(src []byte)
	DecodeKey(src []byte)
}

// readerWriter represents a bschema of a model, managing a table.
type readerWriter struct {
	name      []byte
	namespace []byte
	fields    []Field
	valLen    int
}

func NewReaderWriter(namespace, name []byte) ReaderWriter {
	return &readerWriter{namespace: namespace, name: name}
}

func (bs *readerWriter) Append(typ bsontype.Type, name []byte) {
	bs.fields = append(bs.fields, Field{
		name: name,
		typ:  typ,
	})
	bs.valLen += len(name) + 2 // 1B for type, 1B for NULL terminator
}

func (bs *readerWriter) Namespace() []byte { return bs.namespace }

func (bs *readerWriter) FieldsLen() int { return len(bs.fields) }

func (bs *readerWriter) FieldAt(pos int) Field {
	return bs.fields[pos]
}

// EncodeKey
//
// key format: | namespace \x00 | name \x00 |
func (bs *readerWriter) EncodeKey() []byte {
	dst := make([]byte, len(bs.namespace)+1+len(bs.name)+1)
	written := 0
	written = copy(dst[written:], bs.namespace) + 1
	written = copy(dst[written:], bs.name)
	return dst
}

func (bs *readerWriter) EncodeVal() []byte {
	dst := make([]byte, 0, bs.valLen)
	written := 0
	for _, field := range bs.fields {
		written = copy(dst[written:], field.Encode()) + 1
	}
	return dst
}

func (bs *readerWriter) DecodeVal(src []byte) {
	bs.valLen = len(src)
	lastIdx := 0
	for i := 0; i < len(src); i++ {
		if src[i] == 0 {
			f := Field{}
			f.Decode(src[lastIdx:i])
			bs.fields = append(bs.fields, f)
			lastIdx = i
		}
	}
}

// DecodeKey from bytes.
//
// key format: | namespace \x00 | name \x00 |
func (bs *readerWriter) DecodeKey(src []byte) {
	idx := len(src) - 2 // skip the last null terminator
	for ; idx >= 0 && src[idx] != 0; idx-- {
	}
	// key ref
	cloned := make([]byte, len(src))
	copy(cloned, src)
	bs.namespace = cloned[:idx]
	// skip the null terminator after the namespace
	bs.name = cloned[idx+1 : len(src)-1] // ignore the null terminator of name
}
