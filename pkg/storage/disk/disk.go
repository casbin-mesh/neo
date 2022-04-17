package disk

import (
	"unsafe"
)

// BasicManager Reads pages from disk, Writes pages to disk.
// Max Mapping Size: ( 2^64 - 1 ) * 4KB (Page Size)
type BasicManager interface {
	Open(filename string, opts ...Option) error
	ShutDown() error
	WritePage(pageId uint64, p unsafe.Pointer) error
	ReadPage(pageId uint64, p unsafe.Pointer) error
}
