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
	"errors"
	"unsafe"
)

type BlockReaderWriter interface {
	BlockWriter
	BlockReader
}

type BlockReader interface {
	ValueAt(pos int) ([]byte, []byte)
	Occupied(pos int) bool
}

type BlockWriter interface {
	Encode(dst []byte) int
	AppendEntry(key, value []byte) error
}

type block struct {
	data         []byte
	baseKey      []byte // Base key for current block, use for the key's prefix compaction.
	entryOffsets []uint32
	end          int // points to the end offset of the block. NOTES: MAX addresses 4GiB.
}

// NewBlock returns a block containing a specific memory space.
// The length of the source slice MUST be exact; otherwise, it causes a page fault.
func NewBlock(source []byte) *block {
	return &block{
		data:         source, // memory space
		baseKey:      nil,
		entryOffsets: nil,
		end:          0,
	}
}

type entryKeyHeader struct {
	overlap uint16 // Overlap with base key.
	diff    uint16 // Length of the diff.
}

// Encode encodes the header.
func (h *entryKeyHeader) Encode(dst []byte) int {
	written := 0
	// BigEndian
	// set overlap
	dst[written] = byte(h.overlap >> 8)
	dst[written+1] = byte(h.overlap)
	// set diff
	dst[written+2] = byte(h.diff >> 8)
	dst[written+3] = byte(h.diff)
	written += 4
	return written
}

// Decode decodes the header.
func (h *entryKeyHeader) Decode(s []byte) {
	h.overlap = uint16(s[1]) | uint16(s[0])<<8
	h.diff = uint16(s[3]) | uint16(s[2])<<8
}

func (b *block) parseKeyValue(raw []byte) ([]byte, []byte) {
	h := entryKeyHeader{}
	h.Decode(raw[:SizeOfKeyHeader]) // header
	// copy key
	keyCopy := make([]byte, h.overlap+h.diff)
	read := 0
	read = copy(keyCopy, b.baseKey[:h.overlap])
	read = copy(keyCopy[read:], raw[SizeOfKeyHeader:SizeOfKeyHeader+int(h.diff)])
	// TODO: assert read equals keyCopy len

	var valueCopy []byte
	// make a copy
	valueCopy = append(valueCopy[:0], raw[SizeOfKeyHeader+read:len(raw)]...)
	return keyCopy, valueCopy
}

func (b *block) read(pos int) (begin, end int) {
	end = b.end
	begin = int(b.entryOffsets[pos])

	if pos != len(b.entryOffsets)-1 {
		end = int(b.entryOffsets[pos+1])
	}
	return
}

func (b *block) Occupied(pos int) bool {
	if pos < 0 || pos > len(b.entryOffsets)-1 { // out of bound
		return false
	}
	begin, end := b.read(pos)
	return end-begin > 0
}

// ValueAt returns the key value COPY at position.
func (b *block) ValueAt(pos int) ([]byte, []byte) {
	if pos < 0 || pos > len(b.entryOffsets)-1 { // out of bound
		return nil, nil
	}
	begin, end := b.read(pos)
	raw := b.data[begin:end]
	return b.parseKeyValue(raw)
}

// keyDiff returns a suffix of newKey that is different from b.baseKey.
func (b *block) keyDiff(newKey []byte) []byte {
	var i int
	for i = 0; i < len(newKey) && i < len(b.baseKey); i++ {
		if newKey[i] != b.baseKey[i] {
			break
		}
	}
	return newKey[i:]
}

func (b *block) ensureRoom(need int) error {
	fixed := len(b.baseKey) + // base key len
		(len(b.entryOffsets)+1)*4 //size of offsets

	if len(b.data)-fixed-b.end < need {
		return ErrOutOfSpace
	}
	return nil
}

var (
	ErrOutOfSpace   = errors.New("out of space")
	SizeOfKeyHeader = int(unsafe.Sizeof(entryKeyHeader{}))
)

func (b *block) AppendEntry(key, value []byte) error {
	var (
		diffKey []byte
	)
	if len(b.baseKey) == 0 {
		// eliminate side-effect
		b.baseKey = append(b.baseKey[:], key...)
		diffKey = key
	} else {
		diffKey = b.keyDiff(key)
	}

	h := entryKeyHeader{
		overlap: uint16(len(key) - len(diffKey)),
		diff:    uint16(len(diffKey)),
	}

	// ensure room
	if err := b.ensureRoom(
		SizeOfKeyHeader + // key header
			len(diffKey) + // size of diff key
			len(value), // size of value
	); err != nil {
		return err
	}

	b.entryOffsets = append(b.entryOffsets, uint32(b.end))
	b.end += h.Encode(b.data[b.end:])
	b.append(diffKey)
	b.append(value)

	return nil
}

func (b *block) append(src []byte) {
	b.end += copy(b.data[b.end:], src)
}
