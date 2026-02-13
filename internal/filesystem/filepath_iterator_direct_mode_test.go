/*
SPDX-License-Identifier: GPL-3.0-only
Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>
*/

package filesystem

import (
	"context"
	"path/filepath"
	"slices"
	"testing"

	"github.com/spf13/afero"
)

func TestFileSystemIteratorForManagedMode_All(t *testing.T) {
	type fileEntry struct {
		path  string
		isDir bool
	}

	tests := []struct {
		name          string
		recursive     bool
		filesToCreate []fileEntry
		expectedFiles []string
	}{
		{
			name:          "Empty directory, non-recursive",
			recursive:     false,
			filesToCreate: []fileEntry{},
			expectedFiles: []string{},
		},
		{
			name:          "Empty directory, recursive",
			recursive:     true,
			filesToCreate: []fileEntry{},
			expectedFiles: []string{},
		},
		{
			name:      "Only valid JPG variations",
			recursive: false,
			filesToCreate: []fileEntry{
				{"01.jpg", false}, {"02.jPg", false}, {"03.jpG", false}, {"04.jPG", false},
				{"05.Jpg", false}, {"06.JPg", false}, {"07.JpG", false},
			},
			expectedFiles: []string{
				"01.jpg", "02.jPg", "03.jpG", "04.jPG", "05.Jpg", "06.JPg", "07.JpG",
			},
		},
		{
			name:      "JPG variations and invalid files",
			recursive: false,
			filesToCreate: []fileEntry{
				{"01.jpg", false}, {"02.jPg", false}, {"03.jpG", false},
				{"04.jPG", false}, {"05.Jpg", false}, {"06.JPg", false},
				{"07.png", false}, {"08.bmp", false}, {"09.JpG", false},
				{"10.jpg.pdf", false}, {".jpg", false}, {"11.", false},
				{"12.jpg.tmp", false}, {"13.image.jpg", false}, {"14.jpg", true},
			},
			expectedFiles: []string{
				"01.jpg", "02.jPg", "03.jpG", "04.jPG", "05.Jpg", "06.JPg", "09.JpG", "13.image.jpg",
			},
		},
		{
			name:      "Only valid JPEG variations",
			recursive: false,
			filesToCreate: []fileEntry{
				{"01.jpeg", false}, {"02.jpeG", false}, {"03.jpEg", false}, {"04.jpEG", false},
				{"05.JPeG", false}, {"06.JPeg", false}, {"07.JpeG", false}, {"08.jPeg", false},
				{"09.jPeg", false}, {"10.jPEg", false}, {"11.jPEG", false}, {"11.Jpeg", false},
				{"12.JpEG", false}, {"13.JpEg", false}, {"14.JPEg", false}, {"15.JPEG", false},
			},
			expectedFiles: []string{
				"01.jpeg", "02.jpeG", "03.jpEg", "04.jpEG",
				"05.JPeG", "06.JPeg", "07.JpeG", "08.jPeg",
				"09.jPeg", "10.jPEg", "11.jPEG", "11.Jpeg",
				"12.JpEG", "13.JpEg", "14.JPEg", "15.JPEG",
			},
		},
		{
			name:      "JPEG variations with invalid files",
			recursive: false,
			filesToCreate: []fileEntry{
				{"01.jpeg", false}, {"02.jpeG", false}, {"03.jpEg", false}, {"04.jpEG", false},
				{"05.JPeG", false}, {"06.JPeg", false}, {"07.JpeG", false}, {"08.jPeg", false},
				{"09.pdf", false}, {"10.png", false}, {"11.BMP", false}, {"12.TiF", false},
				{"13.jPeg", false}, {"14.jPEg", false}, {"15.jPEG", false}, {"16.Jpeg", false},
				{"17.jpEg.pdf", false}, {".jpEG", false}, {"18.", false}, {"19.JPEG.tmp", false},
				{"20.image.JpEG", false}, {"21.JpEg", true}, {"22.JpEg", false}, {"23.JPEg", false},
				{"24.JPEG", false},
			},
			expectedFiles: []string{
				"01.jpeg", "02.jpeG", "03.jpEg", "04.jpEG",
				"05.JPeG", "06.JPeg", "07.JpeG", "08.jPeg",
				"13.jPeg", "14.jPEg", "15.jPEG", "16.Jpeg",
				"20.image.JpEG", "22.JpEg", "23.JPEg", "24.JPEG",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()

			// We're creating files and folders structure here
			for _, entry := range tt.filesToCreate {
				if entry.isDir {
					if err := fs.MkdirAll(entry.path, DefaultFolderPermissions); err != nil {
						t.Fatalf("setup failed: mkdir %q: %v", entry.path, err)
					}
				} else {
					// Creating parent folder
					dir := filepath.Dir(entry.path)
					if err := fs.MkdirAll(dir, DefaultFolderPermissions); err != nil {
						t.Fatalf("setup failed: mkdir parent %q: %v", dir, err)
					}
					// Creating file
					if _, err := fs.Create(entry.path); err != nil {
						t.Fatalf("setup failed: create file %q: %v", entry.path, err)
					}
				}
			}

			iter, err := NewFilePathsIteratorForManagedMode(fs, ".", tt.recursive)
			if err != nil {
				t.Fatalf("Failed to create iterator: %v", err)
			}

			var foundFiles []string
			for path := range iter.All(context.Background()) {
				foundFiles = append(foundFiles, filepath.Clean(path))
			}

			slices.Sort(foundFiles)
			slices.Sort(tt.expectedFiles)

			if !slices.Equal(foundFiles, tt.expectedFiles) {
				t.Errorf("\nExpected: %v\nGot: %v", tt.expectedFiles, foundFiles)
			}
		})
	}
}
