// SPDX-License-Identifier: GPL-3.0-only
// Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>

package repair

import (
	"bytes"
	"io"
	"path/filepath"
	"testing"

	"github.com/cdefgah/whatsapp-jpeg-repair/internal/filesystem"
	"github.com/cdefgah/whatsapp-jpeg-repair/internal/options"
	"github.com/spf13/afero"
)

func TestNewImageRepairerForManagedMode(t *testing.T) {
	memFS := afero.NewMemMapFs()
	var buf bytes.Buffer

	type args struct {
		fs     afero.Fs
		opts   options.ManagedModeOptions
		stderr io.Writer
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "Successful initialization with default options",
			args: args{
				fs:     memFS,
				stderr: &buf,
				opts: options.ManagedModeOptions{
					SourceFolderPath:      "/src",
					DestinationFolderPath: "/dst",
				},
			},
		},
		{
			name: "Initialization with empty options",
			args: args{
				fs:     nil,
				stderr: nil,
				opts:   options.ManagedModeOptions{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewImageRepairerForManagedMode(tt.args.fs, tt.args.opts, tt.args.stderr)

			if got == nil {
				t.Fatal("expected non-nil ImageRepairerForManagedMode")
			}

			if got.fs != tt.args.fs {
				t.Errorf("fs: got %v, want %v", got.fs, tt.args.fs)
			}

			if got.stderr != tt.args.stderr {
				t.Errorf("stderr: got %v, want %v", got.stderr, tt.args.stderr)
			}

			if got.stats == nil {
				t.Error("expected stats to be initialized, got nil")
			} else if got.stats.Total != 0 || len(got.stats.Errors) != 0 {
				t.Errorf("expected empty stats, got %+v", got.stats)
			}

			if got.options != tt.args.opts {
				t.Errorf("options: got %+v, want %+v", got.options, tt.args.opts)
			}
		})
	}
}

func TestCreateFolderIfItDoesNotExist(t *testing.T) {
	tests := []struct {
		name           string
		setupFS        func(fs afero.Fs)
		pathToFolder   string
		wantErr        bool
		checkCondition func(t *testing.T, fs afero.Fs, path string)
	}{
		{
			name: "Folder already exists",
			setupFS: func(fs afero.Fs) {
				_ = fs.MkdirAll("/already/exists", filesystem.DefaultFolderPermissions)
			},
			pathToFolder: "/already/exists",
			wantErr:      false,
			checkCondition: func(t *testing.T, fs afero.Fs, path string) {
				exists, _ := afero.DirExists(fs, path)
				if !exists {
					t.Errorf("folder %q should still exist", path)
				}
			},
		},
		{
			name:         "Folder does not exist and should be created",
			setupFS:      func(fs afero.Fs) {}, // empty filesystem
			pathToFolder: "/new/folder",
			wantErr:      false,
			checkCondition: func(t *testing.T, fs afero.Fs, path string) {
				exists, _ := afero.DirExists(fs, path)
				if !exists {
					t.Errorf("folder %q was not created", path)
				}
			},
		},
		{
			name: "Path is an existing file (should return error)",
			setupFS: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "/path/is/a/file.jpg", []byte("data"), filesystem.DefaultFilePermissions)
			},
			pathToFolder: "/path/is/a/file.jpg",
			wantErr:      true,
			checkCondition: func(t *testing.T, fs afero.Fs, path string) {
				isDir, _ := afero.IsDir(fs, path)
				if isDir {
					t.Errorf("path %q should not be a directory", path)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			if tt.setupFS != nil {
				tt.setupFS(fs)
			}

			ir := &ImageRepairerForManagedMode{
				ImageRepairerBase: ImageRepairerBase{
					fs: fs,
				},
			}

			err := ir.createFolderIfItDoesNotExist(tt.pathToFolder)

			if (err != nil) != tt.wantErr {
				t.Errorf("createFolderIfItDoesNotExist() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.checkCondition != nil {
				tt.checkCondition(t, fs, tt.pathToFolder)
			}
		})
	}
}

func TestEnsureDestFolderPath(t *testing.T) {
	tests := []struct {
		name           string
		opts           options.ManagedModeOptions
		sourceFilePath string
		setupFS        func(fs afero.Fs)
		wantPath       string
		wantErr        bool
	}{
		{
			name: "Basic mapping: file in root source",
			opts: options.ManagedModeOptions{
				SourceFolderPath:      "/src",
				DestinationFolderPath: "/dst",
			},
			sourceFilePath: "/src/image.jpg",
			wantPath:       filepath.Join("/dst", "."),
			wantErr:        false,
		},
		{
			name: "Nested mapping: file in subfolder",
			opts: options.ManagedModeOptions{
				SourceFolderPath:      "/src",
				DestinationFolderPath: "/dst",
			},
			sourceFilePath: "/src/vacation/2023/photo.png",
			wantPath:       filepath.Join("/dst", "vacation/2023"),
			wantErr:        false,
			setupFS:        func(fs afero.Fs) {},
		},
		{
			name: "Path outside of source folder",
			opts: options.ManagedModeOptions{
				SourceFolderPath:      "/src/images",
				DestinationFolderPath: "/dst",
			},
			sourceFilePath: "/other/random_file.jpg",
			wantErr:        true, // filepath.Rel should return error for this case
		},
		{
			name: "Conflict: destination path is a file",
			opts: options.ManagedModeOptions{
				SourceFolderPath:      "/src",
				DestinationFolderPath: "/dst",
			},
			sourceFilePath: "/src/folder/img.jpg",
			setupFS: func(fs afero.Fs) {
				_ = fs.MkdirAll("/dst", filesystem.DefaultFolderPermissions)
				_ = afero.WriteFile(fs, "/dst/folder", []byte("I am a file"), filesystem.DefaultFilePermissions)
			},
			wantErr: true, // ensureDestFolderPath should return error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			if tt.setupFS != nil {
				tt.setupFS(fs)
			}

			ir := &ImageRepairerForManagedMode{
				ImageRepairerBase: ImageRepairerBase{
					fs: fs,
				},
				options: tt.opts,
			}

			gotPath, err := ir.ensureDestFolderPath(tt.sourceFilePath)

			if (err != nil) != tt.wantErr {
				t.Errorf("ensureDestFolderPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if gotPath != tt.wantPath {
					t.Errorf("got path %q, want %q", gotPath, tt.wantPath)
				}

				exists, _ := afero.DirExists(fs, gotPath)
				if !exists {
					t.Errorf("directory %q was not actually created in FS", gotPath)
				}
			}
		})
	}
}
