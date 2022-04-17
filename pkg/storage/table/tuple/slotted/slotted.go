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
		//TODO(insert to a overflow page)
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

func (s *slotted) Delete(slot uint16) error {
	start := pageSize - slot*SizeOfSlot
	s[start] = byte(Deleted >> 8)
	s[start+1] = byte(Deleted)
	return nil
}

func (s *slotted) Get(slot uint16) ([]byte, bool, error) {
	size, offset, _ := s.getSlot(slot)
	if size == 0 { // deleted
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
