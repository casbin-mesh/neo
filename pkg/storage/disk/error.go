package disk

import "errors"

var (
	ErrIOReadExceedFileSize = errors.New("I/O error: reading exceed file size")
)
