// SPDX-License-Identifier: GPL-3.0-only
// Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>

package filesystem

import (
	"context"
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/afero"
)

// folderEntry contains state for a single directory level in the stack.
type folderEntry struct {
	pathToFolder      string
	folderObjectsList []os.FileInfo
	index             int
}

// FilePathsIteratorForManagedMode contains filesystem iterator data for processing files in managed mode.
type FilePathsIteratorForManagedMode struct {
	filesystem afero.Fs
	stack      []folderEntry
	recursive  bool
}

// NewFilePathsIteratorForManagedMode creates a new file path iterator for managed mode based on the provided parameters.
func NewFilePathsIteratorForManagedMode(fs afero.Fs, root string, recursive bool) (*FilePathsIteratorForManagedMode, error) {
	info, err := fs.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("stat root directory: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %q", root)
	}

	entries, err := afero.ReadDir(fs, root)
	if err != nil {
		return nil, fmt.Errorf("read root directory: %w", err)
	}

	return &FilePathsIteratorForManagedMode{
		filesystem: fs,
		recursive:  recursive,
		stack: []folderEntry{
			{
				pathToFolder:      root,
				folderObjectsList: entries,
				index:             0,
			},
		},
	}, nil
}

// All returns an iterator over all file paths in the traverser.
func (it *FilePathsIteratorForManagedMode) All(ctx context.Context) iter.Seq[string] {
	return func(yield func(string) bool) {
		for len(it.stack) > 0 {
			if err := ctx.Err(); err != nil {
				return
			}

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

			if entry.IsDir() && it.recursive {
				entries, err := afero.ReadDir(it.filesystem, fullPath)

				// Skipping empty folders and folders that can't be read (with permission errors for example)
				if err != nil || len(entries) == 0 {
					continue
				}

				it.stack = append(it.stack, folderEntry{
					pathToFolder:      fullPath,
					folderObjectsList: entries,
					index:             0,
				})
				continue
			}

			if entry.Mode().IsRegular() {
				if !isJpegFileExtension(fullPath) {
					continue
				}

				if !yield(fullPath) {
					return
				}
			}
		}
	}
}

// isJpegFileExtension returns true if the filename has a known JPEG extension.
func isJpegFileExtension(filename string) bool {
	normalizedFullname := strings.ToLower(filename)
	ext := filepath.Ext(normalizedFullname)

	name := strings.TrimSuffix(filepath.Base(normalizedFullname), ext)
	if name == "" || name == "." {
		return false
	}

	valid := []string{".jpg", ".jpeg", ".jpe", ".jif", ".jfif", ".jfi"}
	return slices.Contains(valid, ext)
}
