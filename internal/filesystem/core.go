package filesystem

import (
	"context"
	"iter"
)

// FilePathIterator provides a way to iterate over a sequence of file paths.
type FilePathIterator interface {
	// All returns an iterator over all files discovered by the iterator.
	// It is intended to be used with a for-range loop.
	All(context context.Context) iter.Seq[string]
}
