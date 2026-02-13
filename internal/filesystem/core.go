// SPDX-License-Identifier: GPL-3.0-only
// Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>

package filesystem

import (
	"context"
	"iter"
)

const DefaultFolderPermissions = 0o755
const DefaultFilePermissions = 0o644

// FilePathIterator provides a way to iterate over a sequence of file paths.
type FilePathIterator interface {
	// All returns an iterator over all files discovered by the iterator.
	// It is intended to be used with a for-range loop.
	All(context context.Context) iter.Seq[string]
}
