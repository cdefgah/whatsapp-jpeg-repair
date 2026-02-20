// SPDX-License-Identifier: GPL-3.0-only
// Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>

package repair

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"path/filepath"
	"testing"
	"time"

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

func TestImageRepairerForManagedMode_MakeFolderIfMissing(t *testing.T) {
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

func TestImageRepairerForManagedMode_EnsureDestFolderPath(t *testing.T) {
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
			name:         "Successfull processing",
			srcBase:      "/data/source",
			dstBase:      "/data/destination",
			srcFilePath:  "/data/source/vacation/photo.jpg",
			expectedPath: filepath.Clean("/data/destination/vacation/photo.jpg"),
			wantErr:      false,
		},
		{
			name:         "File in the root of the source folder",
			srcBase:      "/data/source",
			dstBase:      "/data/destination",
			srcFilePath:  "/data/source/root-image.png",
			expectedPath: filepath.Clean("/data/destination/root-image.png"),
			wantErr:      false,
		},
		{
			name:         "Deeply nested folders",
			srcBase:      "/src",
			dstBase:      "/dst",
			srcFilePath:  "/src/2023/reports/january/file.pdf",
			expectedPath: filepath.Clean("/dst/2023/reports/january/file.pdf"),
			wantErr:      false,
		},
		{
			name:         "Error: file is outside of the source folder",
			srcBase:      "/data/source",
			dstBase:      "/data/destination",
			srcFilePath:  "/data/other/intruder.jpg",
			expectedPath: "",
			wantErr:      true,
		},
		{
			name:         "Empty path to file",
			srcBase:      "/data/source",
			dstBase:      "/data/destination",
			srcFilePath:  "",
			expectedPath: "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			got, err := ir.prepareDestFilePath(tt.srcFilePath)

			if (err != nil) != tt.wantErr {
				t.Errorf("prepareDestFilePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.expectedPath {
				t.Errorf("prepareDestFilePath() got = %v, want %v", got, tt.expectedPath)
			}

			if !tt.wantErr {
				exists, _ := afero.DirExists(fs, filepath.Dir(got))
				if !exists {
					t.Errorf("expected destination directory %q was not created", filepath.Dir(got))
				}
			}
		})
	}
}

func TestImageRepairerForManagedMode_SetSrcFileModTimeToDestFile(t *testing.T) {
	testTime := time.Date(2026, time.January, 3, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		srcPath     string
		dstPath     string
		setupFs     func(fs afero.Fs)
		wantErr     bool
		expectedErr error
	}{
		{
			name:    "Successful operation",
			srcPath: "/source.jpg",
			dstPath: "/dest.jpg",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "/source.jpg", []byte("data"), filesystem.DefaultFilePermissions)
				_ = fs.Chtimes("/source.jpg", testTime, testTime)
				_ = afero.WriteFile(fs, "/dest.jpg", []byte("data"), filesystem.DefaultFilePermissions)
				_ = fs.Chtimes("/dest.jpg", testTime.Add(time.Hour), testTime.Add(time.Hour))
			},
			wantErr: false,
		},
		{
			name:    "Error: source file dos not exist",
			srcPath: "/missing.jpg",
			dstPath: "/dest.jpg",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "/dest.jpg", []byte("data"), filesystem.DefaultFilePermissions)
			},
			wantErr: true,
		},
		{
			name:    "Error: target file does not exist",
			srcPath: "/source.jpg",
			dstPath: "/ghost.jpg",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "/source.jpg", []byte("data"), filesystem.DefaultFilePermissions)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			if tt.setupFs != nil {
				tt.setupFs(fs)
			}

			ir := &ImageRepairerForManagedMode{
				ImageRepairerBase: ImageRepairerBase{fs: fs},
			}

			err := ir.setSrcFileModTimeToDestFile(tt.srcPath, tt.dstPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("setSrcFileModTimeToDestFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				dstStat, _ := fs.Stat(tt.dstPath)
				if !dstStat.ModTime().Equal(testTime) {
					t.Errorf("ModTime mismatch: got %v, want %v", dstStat.ModTime(), testTime)
				}
			}
		})
	}
}

func TestImageRepairerForManagedMode_ProcessSingleFile(t *testing.T) {
	createValidImage := func(fs afero.Fs, path string) error {
		t.Helper() // to ensure clean logs if test fails

		img := image.NewRGBA(image.Rect(0, 0, 1, 1))
		img.Set(0, 0, color.White)

		f, err := fs.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()
		return jpeg.Encode(f, img, nil)
	}

	const (
		srcDir   = "/source"
		dstDir   = "/dest"
		fileName = "test.jpg"
	)

	tests := []struct {
		name        string
		options     options.ManagedModeOptions
		setupCtx    func() (context.Context, context.CancelFunc)
		setupFs     func(fs afero.Fs)
		srcPath     string
		wantErr     bool
		checkResult func(t *testing.T, fs afero.Fs)
	}{
		{
			name: "Successfull processing without deleting source file",
			options: options.ManagedModeOptions{
				SourceFolderPath:           srcDir,
				DestinationFolderPath:      dstDir,
				UseCurrentModificationTime: true,
				DeleteWhatsAppFiles:        false,
			},
			setupFs: func(fs afero.Fs) {
				_ = fs.MkdirAll(srcDir, filesystem.DefaultFolderPermissions)
				_ = createValidImage(fs, filepath.Join(srcDir, fileName))
			},
			srcPath: filepath.Join(srcDir, fileName),
			wantErr: false,
			checkResult: func(t *testing.T, fs afero.Fs) {
				exists, _ := afero.Exists(fs, filepath.Join(dstDir, fileName))
				if !exists {
					t.Error("target file is not found")
				}
				exists, _ = afero.Exists(fs, filepath.Join(srcDir, fileName))
				if !exists {
					t.Error("the source file must not be deleted")
				}
			},
		},
		{
			name: "Successfull processing with deleting source file",
			options: options.ManagedModeOptions{
				SourceFolderPath:      srcDir,
				DestinationFolderPath: dstDir,
				DeleteWhatsAppFiles:   true,
			},
			setupFs: func(fs afero.Fs) {
				_ = fs.MkdirAll(srcDir, filesystem.DefaultFolderPermissions)
				_ = createValidImage(fs, filepath.Join(srcDir, fileName))
			},
			srcPath: filepath.Join(srcDir, fileName),
			wantErr: false,
			checkResult: func(t *testing.T, fs afero.Fs) {
				exists, _ := afero.Exists(fs, filepath.Join(srcDir, fileName))
				if exists {
					t.Error("source file must be deleted")
				}
			},
		},
		{
			name: "Copying source file modification time",
			options: options.ManagedModeOptions{
				SourceFolderPath:           srcDir,
				DestinationFolderPath:      dstDir,
				UseCurrentModificationTime: false,
			},
			setupFs: func(fs afero.Fs) {
				p := filepath.Join(srcDir, fileName)
				_ = fs.MkdirAll(srcDir, filesystem.DefaultFolderPermissions)
				_ = createValidImage(fs, p)
				past := time.Now().Add(-24 * time.Hour).Truncate(time.Second)
				_ = fs.Chtimes(p, past, past)
			},
			srcPath: filepath.Join(srcDir, fileName),
			wantErr: false,
			checkResult: func(t *testing.T, fs afero.Fs) {
				sStat, _ := fs.Stat(filepath.Join(srcDir, fileName))
				dStat, _ := fs.Stat(filepath.Join(dstDir, fileName))
				if !sStat.ModTime().Equal(dStat.ModTime()) {
					t.Errorf("file modification time does not match: src=%v, dst=%v", sStat.ModTime(), dStat.ModTime())
				}
			},
		},
		{
			name: "Error: cancelling context before start",
			setupCtx: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx, cancel
			},
			srcPath: filepath.Join(srcDir, fileName),
			wantErr: true,
		},
		{
			name: "Error: file is outside of source folder",
			options: options.ManagedModeOptions{
				SourceFolderPath:      srcDir,
				DestinationFolderPath: dstDir,
			},
			srcPath: "/etc/passwd", // path is outside of srcDir
			wantErr: true,
		},
		{
			name: "Error: corrupted image file",
			options: options.ManagedModeOptions{
				SourceFolderPath:      srcDir,
				DestinationFolderPath: dstDir,
			},
			setupFs: func(fs afero.Fs) {
				_ = fs.MkdirAll(srcDir, filesystem.DefaultFolderPermissions)
				_ = afero.WriteFile(fs, filepath.Join(srcDir, fileName), []byte("non-image data"), filesystem.DefaultFilePermissions)
			},
			srcPath: filepath.Join(srcDir, fileName),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			if tt.setupFs != nil {
				tt.setupFs(fs)
			}

			ctx := context.Background()
			if tt.setupCtx != nil {
				var cancel context.CancelFunc
				ctx, cancel = tt.setupCtx()
				defer cancel()
			}

			ir := &ImageRepairerForManagedMode{
				ImageRepairerBase: ImageRepairerBase{
					fs:    fs,
					stats: &Stats{},
				},
				options: tt.options,
			}

			err := ir.ProcessSingleFile(ctx, tt.srcPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessSingleFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkResult != nil {
				tt.checkResult(t, fs)
			}
		})
	}
}
