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

// FilePathsIteratorForManagedMode contains filesystem iterator data for processing files in managed mode.
type FilePathsIteratorForManagedMode struct {
	filesystem           afero.Fs
	stack                []folderEntry
	recursive            bool
	processOnlyJpegFiles bool
}

// folderEntry contains state for a single directory level in the stack.
type folderEntry struct {
	pathToFolder      string
	folderObjectsList []os.FileInfo
	index             int
}

// NewFilePathsIteratorForManagedMode creates a new file path iterator for managed mode based on the provided parameters.
func NewFilePathsIteratorForManagedMode(fs afero.Fs, root string, recursive bool, processOnlyJpegFiles bool) (*FilePathsIteratorForManagedMode, error) {
	info, err := fs.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("stat root directory: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", root)
	}

	entries, err := afero.ReadDir(fs, root)
	if err != nil {
		return nil, fmt.Errorf("read root directory: %w", err)
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

// Next returns the path to the next file and true.
// If no files are left, it returns an empty string and false.
func (it *FilePathsIteratorForManagedMode) Next() (string, bool) {
	for len(it.stack) > 0 {
		topIndex := len(it.stack) - 1
		currentFolder := &it.stack[topIndex]

		// if all elements in the current folder are processed,
		// removing the folder from the stack and continue the loop
		if currentFolder.index >= len(currentFolder.folderObjectsList) {
			it.stack = it.stack[:topIndex]
			continue
		}

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
				// Skipping empty folders and folders with permission errors
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
			if it.processOnlyJpegFiles && !isJpegFileExtension(fullPath) {
				continue
			}

			return fullPath, true
		}
	}

	return "", false
}

// isJpegFileExtension returns true if the filename has a known JPEG extension.
func isJpegFileExtension(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".jpe", ".jif", ".jfif", ".jfi":
		return true
	default:
		return false
	}
}
