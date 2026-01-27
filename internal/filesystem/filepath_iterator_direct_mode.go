package filesystem

/*
SPDX-License-Identifier: GPL-3.0-only
Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>
*/

// Contains filesystem iterator data for processing files in direct mode.
type FilePathsIteratorForDirectMode struct {
	filePaths []string
	index     int
}

// Creates new instance of file path iterator for direct mode.
//
// # Parameters
//
// filePaths - slice with list of file paths to be processed.
//
// # Returns
//
// Reference to a new instance of file path iterator for direct mode.
func NewFilePathsIteratorForDirectMode(filePaths []string) *FilePathsIteratorForDirectMode {
	return &FilePathsIteratorForDirectMode{
		filePaths: filePaths,
		index:     0,
	}
}

// Returns path to the next file or empty string "" if there are no more files left.
//
// # Returns
//
// path to the next file, returns empty string "" if there are no more files left.
func (it *FilePathsIteratorForDirectMode) Next() (string, bool) {
	if it.index >= len(it.filePaths) {
		return "", false
	}

	singleFilePath := it.filePaths[it.index]
	it.index++
	return singleFilePath, true
}
