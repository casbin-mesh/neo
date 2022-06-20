package hybrid

const (
	// xxxxxxx0 in memory, xxxxxxx1 evicted, xxxxxx10 cooling , xxxxxx00 hot

	evicted_bit  = uint64(1)
	evicted_mask = ^(evicted_bit)
	cool_bit     = uint64(2)
	cool_mask    = ^(cool_bit) | evicted_bit
)

// Swip swizzling pointer
type Swip struct {
	pid uint64
	bf  *BufferFrame
}

func NewSwip(pid uint64, frame *BufferFrame) *Swip {
	return &Swip{
		pid: pid << 2,
		bf:  frame,
	}
}

func (s *Swip) isHot() bool {
	return s.pid&(evicted_bit|cool_bit) == 0
}

func (s *Swip) isCool() bool {
	return s.pid&cool_bit == cool_bit
}

func (s *Swip) isEvicted() bool {
	return s.pid&evicted_bit == evicted_bit
}

func (s *Swip) asPageId() uint64 {
	// 2 bit for state
	return s.pid >> 2
}

func (s *Swip) asPtr() *BufferFrame {
	return s.bf
}

func (s *Swip) warm(bf *BufferFrame) {
	// assert is cool
	s.bf = bf
	s.pid &= ^cool_bit
}

func (s *Swip) cool() {
	s.pid |= cool_bit
}

func (s *Swip) evict(pid uint64) {
	s.pid = (pid << 2) | evicted_bit
}
