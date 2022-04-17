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

import "errors"

const (
	pageSize          = 4096
	Deleted           = uint16(0)
	MaxInlineDataSize = (1 << 16) - 1
	OffsetOfSlotCount = 2
	OffsetOfWOffset   = OffsetOfSlotCount + 2
	OffsetOfFree      = OffsetOfWOffset + 2
	OffsetOfHeader    = 8
	SizeOfSlot        = 4
)

// +-----------------------------------------+
//             Slotted Page layout
// +--------+------------+------------+------+
// | Header | Data Chunk | Free Space | Slot |
// +--------+------------+------------+------+
// *******************************************
// *         This is NOT thread-safe         *
// *******************************************

// +-----------------------------------------+
//                   Header
// +-----------------------------------------+
// - SlotCount (uint16) The number of occupied slots.
// - WOffset (uint16) It points to the end of Data Chunk.
// - Free (uint16) The Size of free.
// +-----------------------------------------+

// +-----------------------------------------+
//                    Slot
// +-----------------------------------------+
// - DOffset (uint16) It points to the start of Data Chunk.
// - DSize (uint16) The size of Data Chunk.
// +-----------------------------------------+

type slotted [pageSize]byte

var (
	ErrZeroSizeData = errors.New("couldn't insert 0 size data")
	ErrOutOfSpace   = errors.New("page run out of space")
)

func (s *slotted) Insert(data []byte) (slot uint16, err error) {
	ds := len(data)
	if ds > MaxInlineDataSize {
		//TODO: insert to a overflow page
	}
	if ds == 0 {
		return 0, ErrZeroSizeData
	}
	idx := s.getSlotCount() + 1
	woff := s.getWOffset()
	free := s.getFree()
	if free < (uint16(ds) + uint16(SizeOfSlot)) {
		return 0, ErrOutOfSpace
	}
	copy(s[woff:], data)

	// updates slots info
	s.setSlot(idx, uint16(ds), woff)
	// updates writing offset
	s.setWOffset(woff + uint16(ds))
	// updates idx
	s.setSlotCount(idx)
	// updates free
	s.setFree(free - uint16(ds) - SizeOfSlot)
	return idx, err
}

// Delete mask target slot data deleted.
// We set the first bit of size to 1 to identify current data as unreadable.
// In the future, when the page needs to be compacted, we can easily figure out the size of each tuple.
func (s *slotted) Delete(slot uint16) error {
	start := pageSize - slot*SizeOfSlot
	mask := 0b1 << 7 // 10000000
	// set first bit to 1
	s[start] = s[start] | byte(mask)
	return nil
}

func isDeleted(size uint16) bool {
	mask := 0b1 << 7 // 10000000
	return byte(size>>8)&byte(mask) == byte(mask)
}

func (s *slotted) Get(slot uint16) ([]byte, bool, error) {
	size, offset, _ := s.getSlot(slot)
	if isDeleted(size) { // deleted
		return nil, true, nil
	}
	return s[offset : offset+size], false, nil
}

func NewSlotted() *slotted {
	s := slotted{}
	s.setFree(pageSize - OffsetOfHeader)
	s.setWOffset(OffsetOfHeader)
	return &s
}

func (s *slotted) setSlot(idx uint16, size uint16, offset uint16) {
	start := pageSize - idx*SizeOfSlot
	// set size
	s[start] = byte(size >> 8)
	s[start+1] = byte(size)
	// set offset
	s[start+2] = byte(offset >> 8)
	s[start+3] = byte(offset)
}

func (s *slotted) getSlot(idx uint16) (size uint16, offset uint16, err error) {
	start := pageSize - idx*SizeOfSlot
	return uint16(s[start+1]) | uint16(s[start])<<8,
		uint16(s[start+3]) | uint16(s[start+2])<<8,
		nil
}

func (s *slotted) setFree(n uint16) {
	// big endian
	s[OffsetOfFree] = byte(n >> 8)
	s[OffsetOfFree+1] = byte(n)
}

func (s *slotted) getFree() uint16 {
	// big endian
	return uint16(s[OffsetOfFree+1]) | uint16(s[OffsetOfFree])<<8
}

func (s *slotted) setWOffset(n uint16) {
	// big endian
	s[OffsetOfWOffset] = byte(n >> 8)
	s[OffsetOfWOffset+1] = byte(n)
}

func (s *slotted) getWOffset() uint16 {
	// big endian
	return uint16(s[OffsetOfWOffset+1]) | uint16(s[OffsetOfWOffset])<<8
}

func (s *slotted) setSlotCount(n uint16) {
	// big endian
	s[OffsetOfSlotCount] = byte(n >> 8)
	s[OffsetOfSlotCount+1] = byte(n)
}

func (s *slotted) getSlotCount() (n uint16) {
	// big endian
	return uint16(s[OffsetOfSlotCount+1]) | uint16(s[OffsetOfSlotCount])<<8
}

func (s *slotted) addSlotCount() {
	s.setSlotCount(s.getSlotCount() + 1)
}
