package filesystem

/*
SPDX-License-Identifier: GPL-3.0-only
Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>
*/

// FilePathsIteratorForDirectMode contains filesystem iterator data for processing files in direct mode.
type FilePathsIteratorForDirectMode struct {
	filePaths []string
	index     int
}

// NewFilePathsIteratorForDirectMode creates a new file path iterator for direct mode based on the provided parameters.
func NewFilePathsIteratorForDirectMode(filePaths []string) *FilePathsIteratorForDirectMode {
	return &FilePathsIteratorForDirectMode{
		filePaths: filePaths,
		index:     0,
	}
}

// Next returns the path to the next file and true.
// If no files are left, it returns an empty string and false.
func (it *FilePathsIteratorForDirectMode) Next() (string, bool) {
	if it.index >= len(it.filePaths) {
		return "", false
	}

	singleFilePath := it.filePaths[it.index]
	it.index++
	return singleFilePath, true
}
