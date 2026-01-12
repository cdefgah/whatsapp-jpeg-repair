package filesystem

/*
SPDX-License-Identifier: GPL-3.0-only
Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>
*/

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

// Contains filesystem iterator data for processing files in managed mode.
type FilePathsIteratorForManagedMode struct {
	filesystem           afero.Fs
	recursive            bool
	processOnlyJpegFiles bool
	stack                []folderEntry
}

// Contains one folder entry element for the stack.
type folderEntry struct {
	pathToFolder      string
	folderObjectsList []os.FileInfo
	index             int
}

// Creates a new file system iterator to process files in managed mode.
//
// # Parameters
//
// fs - Reference to the filesystem object.
// root - path to root folder.
// recursive - if we process nested folder, then this parameter is set to true, false otherwise.
// processOnlyJpegFiles - if we want to process only jpeg files, we set this parameter to true, false otherwise.
//
// # Returns
//
// Pointer to the FilePathsIteratorForManagedMode structure.
// error if something went wrong.
func NewFilePathsIteratorForManagedMode(fs afero.Fs, root string, recursive bool, processOnlyJpegFiles bool) (*FilePathsIteratorForManagedMode, error) {
	info, err := fs.Stat(root)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("Path is not a directory: %s", root)
	}

	// populating stack with root folder entries
	entries, err := afero.ReadDir(fs, root)
	if err != nil {
		return nil, err
	}

	it := &FilePathsIteratorForManagedMode{
		filesystem:           fs,
		recursive:            recursive,
		processOnlyJpegFiles: processOnlyJpegFiles,
		stack: []folderEntry{
			{
				pathToFolder:      root,
				folderObjectsList: entries,
				index:             0,
			},
		},
	}

	return it, nil
}

// Returns path to the next file or empty string "" if there are no more files left.
//
// # Returns
//
// path to the next file, returns empty string "" if there are no more files left.
func (it *FilePathsIteratorForManagedMode) NextFilePath() string {
	for len(it.stack) > 0 {
		// Taking element from the stack top
		topIndex := len(it.stack) - 1
		currentFolder := &it.stack[topIndex]

		// if all elements in the current folder are processed, removing the folder from the stack and continue the loop
		if currentFolder.index >= len(currentFolder.folderObjectsList) {
			it.stack = it.stack[:topIndex] // pop top element
			continue
		}

		// Else, processing the next element
		entry := currentFolder.folderObjectsList[currentFolder.index]
		currentFolder.index++

		// Ignoring symlinks
		if entry.Mode()&os.ModeSymlink != 0 {
			continue
		}

		fullPath := filepath.Join(currentFolder.pathToFolder, entry.Name())

		if entry.IsDir() {
			if it.recursive {
				entries, err := afero.ReadDir(it.filesystem, fullPath)
				if err != nil || len(entries) == 0 {
					continue
				}
				it.stack = append(it.stack, folderEntry{
					pathToFolder:      fullPath,
					folderObjectsList: entries,
					index:             0,
				})
			}
			continue
		}

		// Processing regular files
		if entry.Mode().IsRegular() {

			if it.processOnlyJpegFiles {
				if isJpegFileExtension(fullPath) {
					return fullPath // returning path to jpeg-file
				}
				continue // skipping non-JPEG files
			}

			// if we process all files, just returning path
			return fullPath
		}
	}

	// if stack is empty, then there are no files left
	return ""
}

// Checks if a file name is jpeg file or not.
//
// # Parameters
//
// filename - filename with extension.
//
// # Returns
//
// true, if file is a JPEG-file, false otherwise.
func isJpegFileExtension(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".jpg", ".jpeg", ".jpe", ".jif", ".jfif", ".jfi":
		return true
	}
	return false
}
