package schema

type StoreOption struct {
	path string
	// TODO(noneback): add more option
}

type Store interface {
	Read(key []byte) ([]byte, error)
	Append(key []byte, value []byte) error
	Close()
	// TODO(noneback): add other necessary method
}
