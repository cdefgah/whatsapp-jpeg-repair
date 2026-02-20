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

func TestMakeFolderIfMissing(t *testing.T) {
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

			err := ir.makeFolderIfMissing(tt.pathToFolder)

			if (err != nil) != tt.wantErr {
				t.Errorf("makeFolderIfMissing() error = %v, wantErr %v", err, tt.wantErr)
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

func TestImageRepairerForManagedMode_PrepareDestFilePath(t *testing.T) {
	tests := []struct {
		name         string
		srcBase      string // ir.options.SourceFolderPath
		dstBase      string // ir.options.DestinationFolderPath
		srcFilePath  string
		expectedPath string
		wantErr      bool
	}{
		{
			name:         "Успешное сопоставление в подпапке",
			srcBase:      "/data/source",
			dstBase:      "/data/destination",
			srcFilePath:  "/data/source/vacation/photo.jpg",
			expectedPath: filepath.Clean("/data/destination/vacation/photo.jpg"),
			wantErr:      false,
		},
		{
			name:         "Файл в корне исходной папки",
			srcBase:      "/data/source",
			dstBase:      "/data/destination",
			srcFilePath:  "/data/source/root-image.png",
			expectedPath: filepath.Clean("/data/destination/root-image.png"),
			wantErr:      false,
		},
		{
			name:         "Глубокая вложенность",
			srcBase:      "/src",
			dstBase:      "/dst",
			srcFilePath:  "/src/2023/reports/january/file.pdf",
			expectedPath: filepath.Clean("/dst/2023/reports/january/file.pdf"),
			wantErr:      false,
		},
		{
			name:         "Ошибка: файл вне исходной папки",
			srcBase:      "/data/source",
			dstBase:      "/data/destination",
			srcFilePath:  "/data/other/intruder.jpg",
			expectedPath: "",
			wantErr:      true,
		},
		{
			name:         "Пустой путь к файлу",
			srcBase:      "/data/source",
			dstBase:      "/data/destination",
			srcFilePath:  "",
			expectedPath: "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Используем In-Memory файловую систему, чтобы не мусорить на диске
			fs := afero.NewMemMapFs()

			ir := &ImageRepairerForManagedMode{
				ImageRepairerBase: ImageRepairerBase{
					fs: fs,
				},
				options: options.ManagedModeOptions{
					SourceFolderPath:      tt.srcBase,
					DestinationFolderPath: tt.dstBase,
				},
			}

			// Вызываем тестируемую функцию
			got, err := ir.prepareDestFilePath(tt.srcFilePath)

			// Проверка на наличие ошибки
			if (err != nil) != tt.wantErr {
				t.Errorf("prepareDestFilePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Проверка результата
			if got != tt.expectedPath {
				t.Errorf("prepareDestFilePath() got = %v, want %v", got, tt.expectedPath)
			}

			// Если ошибки не должно быть, проверим, создалась ли папка в MemMapFs
			if !tt.wantErr {
				exists, _ := afero.DirExists(fs, filepath.Dir(got))
				if !exists {
					t.Errorf("expected destination directory %q was not created", filepath.Dir(got))
				}
			}
		})
	}
}
