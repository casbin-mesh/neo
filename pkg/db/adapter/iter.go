package adapter

type IteratorOptions struct {
	// PrefetchSize is the number of KV pairs to prefetch while iterating.
	// Valid only if PrefetchValues is true.
	PrefetchSize int
	// PrefetchValues Indicates whether we should prefetch values during
	// iteration and store them.
	PrefetchValues bool
	Reverse        bool // Direction of iteration. False is forward, true is backward.
	AllVersions    bool // Fetch all valid versions of the same key.
	InternalAccess bool // Used to allow internal access to badger keys.

	// The following option is used to narrow down the SSTables that iterator
	// picks up. If Prefix is specified, only tables which could have this
	// prefix are picked based on their range of keys.
	prefixIsKey bool   // If set, use the prefix for bloom filter lookup.
	Prefix      []byte // Only iterate over this given prefix.
	SinceTs     uint64 // Only read data that has version > SinceTs.
}

// DefaultIteratorOptions contains default options when iterating over Badger key-value stores.
var DefaultIteratorOptions = IteratorOptions{
	PrefetchValues: false,
	PrefetchSize:   100,
	Reverse:        false,
	AllVersions:    false,
}
