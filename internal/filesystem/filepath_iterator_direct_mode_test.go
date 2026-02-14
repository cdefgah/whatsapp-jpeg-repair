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
				{"01.jpg", false}, {"02.jpG", false}, {"03.jPg", false}, {"04.jPG", false},
				{"05.Jpg", false}, {"06.JpG", false}, {"07.JPg", false}, {"08.JPG", false},
			},
			expectedFiles: []string{
				"01.jpg", "02.jpG", "03.jPg", "04.jPG", "05.Jpg", "06.JpG", "07.JPg", "08.JPG",
			},
		},
		{
			name:      "JPG variations and invalid files",
			recursive: false,
			filesToCreate: []fileEntry{
				{"01.jpg", false}, {"02.jpG", false}, {"03.jPg", false}, {"04.jPG", false},
				{"05.Jpg", false}, {"06.JpG", false}, {"AA.png", false}, {"BB.bmp", false},
				{"CC.jpg.pdf", false}, {".jpg", false}, {"DD.", false}, {"EE.jpg.tmp", false},
				{"FF.jpg", true}, {"07.JPg", false}, {"08.image.JPG", false}, {"GG.tif", false},
				{"HH.gif", false}, {"II.apng", false}, {"JJ.avif", false}, {"KK.svg", false},
				{"LL.webp", false}, {"MM.ico", false}, {"NN.docx", false}, {"PP.txt", false},
			},
			expectedFiles: []string{
				"01.jpg", "02.jpG", "03.jPg", "04.jPG", "05.Jpg", "06.JpG", "07.JPg", "08.image.JPG",
			},
		},
		{
			name:      "Only valid JPEG variations",
			recursive: false,
			filesToCreate: []fileEntry{
				{"01.jpeg", false}, {"02.jpeG", false}, {"03.jpEg", false}, {"04.jpEG", false},
				{"05.jPeg", false}, {"06.jPeG", false}, {"07.jPEg", false}, {"08.jPEG", false},
				{"09.Jpeg", false}, {"10.JpeG", false}, {"11.JpEg", false}, {"12.JpEG", false},
				{"13.JPeg", false}, {"14.JPeG", false}, {"15.JPEg", false}, {"16.JPEG", false},
			},
			expectedFiles: []string{
				"01.jpeg", "02.jpeG", "03.jpEg", "04.jpEG",
				"05.jPeg", "06.jPeG", "07.jPEg", "08.jPEG",
				"09.Jpeg", "10.JpeG", "11.JpEg", "12.JpEG",
				"13.JPeg", "14.JPeG", "15.JPEg", "16.JPEG",
			},
		},
		{
			name:      "JPEG variations and invalid files",
			recursive: false,
			filesToCreate: []fileEntry{
				{"01.jpeg", false}, {"02.jpeG", false}, {"03.jpEg", false}, {"04.jpEG", false},
				{"05.jPeg", false}, {"06.jPeG", false}, {"07.jPEg", false}, {"08.jPEG", false},
				{"09.Jpeg", false}, {"10.JpeG", false}, {"11.JpEg", false}, {"12.JpEG", false},
				{"AA.png", false}, {"BB.bmp", false}, {"CC.jpeg.pdf", false}, {".jpeg", false},
				{"DD.", false}, {"EE.jpeg.tmp", false}, {"FF.jpEg", true}, {"GG.tif", false},
				{"HH.gif", false}, {"II.apng", false}, {"JJ.avif", false}, {"KK.svg", false},
				{"LL.webp", false}, {"MM.ico", false}, {"NN.docx", false}, {"PP.txt", false},
				{"13.JPeg", false}, {"14.JPeG", false}, {"15.image.JPEg", false}, {"16.JPEG", false},
			},
			expectedFiles: []string{
				"01.jpeg", "02.jpeG", "03.jpEg", "04.jpEG",
				"05.jPeg", "06.jPeG", "07.jPEg", "08.jPEG",
				"09.Jpeg", "10.JpeG", "11.JpEg", "12.JpEG",
				"13.JPeg", "14.JPeG", "15.image.JPEg", "16.JPEG",
			},
		},
		{
			name:      "Only valid JPE variations",
			recursive: false,
			filesToCreate: []fileEntry{
				{"01.jpe", false}, {"02.jpE", false}, {"03.jPe", false}, {"04.jPE", false},
				{"05.Jpe", false}, {"06.JpE", false}, {"07.JPe", false}, {"08.JPE", false},
			},
			expectedFiles: []string{
				"01.jpe", "02.jpE", "03.jPe", "04.jPE", "05.Jpe", "06.JpE", "07.JPe", "08.JPE",
			},
		},
		{
			name:      "JPE variations and invalid files",
			recursive: false,
			filesToCreate: []fileEntry{
				{"01.jpe", false}, {"02.jpE", false}, {"03.jPe", false}, {"04.jPE", false},
				{"05.Jpe", false}, {"06.JpE", false}, {"CC.jpe.pdf", false}, {".jpe", false},
				{"AA.png", false}, {"BB.bmp", false}, {"FF.jpE", true}, {"GG.tif", false},
				{"DD.", false}, {"EE.jpe.tmp", false}, {"JJ.avif", false}, {"KK.svg", false},
				{"HH.gif", false}, {"II.apng", false}, {"NN.docx", false}, {"PP.txt", false},
				{"LL.webp", false}, {"MM.ico", false}, {"07.image.JPe", false}, {"08.JPE", false},
			},
			expectedFiles: []string{
				"01.jpe", "02.jpE", "03.jPe", "04.jPE", "05.Jpe", "06.JpE", "07.image.JPe", "08.JPE",
			},
		},
		{
			name:      "Only valid JIF variations",
			recursive: false,
			filesToCreate: []fileEntry{
				{"01.jif", false}, {"02.jiF", false}, {"03.jIf", false}, {"04.jIF", false},
				{"05.Jif", false}, {"06.JiF", false}, {"07.JIf", false}, {"08.JIF", false},
			},
			expectedFiles: []string{
				"01.jif", "02.jiF", "03.jIf", "04.jIF", "05.Jif", "06.JiF", "07.JIf", "08.JIF",
			},
		},
		{
			name:      "JIF variations and invalid files",
			recursive: false,
			filesToCreate: []fileEntry{
				{"01.jif", false}, {"02.jiF", false}, {"03.jIf", false}, {"04.jIF", false},
				{"AA.png", false}, {"BB.bmp", false}, {"CC.jif.pdf", false}, {".jif", false},
				{"DD.", false}, {"EE.jif.tmp", false}, {"FF.jif", true}, {"GG.tif", false},
				{"HH.gif", false}, {"II.apng", false}, {"JJ.avif", false}, {"KK.svg", false},
				{"LL.webp", false}, {"MM.ico", false}, {"NN.docx", false}, {"PP.txt", false},
				{"05.Jif", false}, {"06.image.JiF", false}, {"07.JIf", false}, {"08.JIF", false},
			},
			expectedFiles: []string{
				"01.jif", "02.jiF", "03.jIf", "04.jIF", "05.Jif", "06.image.JiF", "07.JIf", "08.JIF",
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
