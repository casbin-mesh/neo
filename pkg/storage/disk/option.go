package disk

import "os"

var (
	DefaultOptions = Options{Flag: os.O_RDWR | os.O_CREATE, Perm: 0755}
)

type Options struct {
	Flag int
	Perm os.FileMode
}

// Clone for write-on-copy
func (opts Options) Clone() Options {
	return opts
}

type Option func(*Options) error
