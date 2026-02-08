// SPDX-License-Identifier: GPL-3.0-only
// Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>

package filesystem

import (
	"context"
	"iter"
)

// FilePathsIteratorForDirectMode contains filesystem iterator data for processing files in direct mode.
type FilePathsIteratorForDirectMode struct {
	filePaths []string
}

// NewFilePathsIteratorForDirectMode creates a new file path iterator for direct mode based on the provided parameters.
func NewFilePathsIteratorForDirectMode(filePaths []string) *FilePathsIteratorForDirectMode {
	return &FilePathsIteratorForDirectMode{
		filePaths: filePaths,
	}
}

// All returns file paths iterator.
func (it *FilePathsIteratorForDirectMode) All(ctx context.Context) iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, path := range it.filePaths {
			if ctx.Err() != nil {
				return
			}

			if !yield(path) {
				return
			}
		}
	}
}
