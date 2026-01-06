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

// Returns path to the next file or empty string "" if there are no more files left.
//
// # Returns
//
// path to the next file, returns empty string "" if there are no more files left.
func (it *FilePathsIteratorForDirectMode) NextFilePath() string {
	if it.index >= len(it.filePaths) {
		return ""
	}

	singleFilePath := it.filePaths[it.index]
	it.index++
	return singleFilePath
}
