// SPDX-License-Identifier: GPL-3.0-only
// Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>

package repair

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	"io"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cdefgah/whatsapp-jpeg-repair/internal/filesystem"
	"github.com/cdefgah/whatsapp-jpeg-repair/internal/options"
	"github.com/spf13/afero"
)

// MockClock всегда возвращает одно и то же время
type MockClock struct {
	FixedTime time.Time
}

func (m MockClock) Now() time.Time { return m.FixedTime }

func TestNewImageRepairerForDirectMode(t *testing.T) {
	tests := []struct {
		name          string
		fs            afero.Fs
		opts          options.DirectModeOptions
		stderr        io.Writer
		expectedFiles int
	}{
		{
			name:   "Basic initialization with two files",
			fs:     afero.NewMemMapFs(),
			stderr: &bytes.Buffer{},
			opts: options.DirectModeOptions{
				FilePaths: []string{"img1.jpg", "img2.png"},
			},
			expectedFiles: 2,
		},
		{
			name:   "Initialization with empty file list",
			fs:     afero.NewMemMapFs(),
			stderr: &bytes.Buffer{},
			opts: options.DirectModeOptions{
				FilePaths: []string{},
			},
			expectedFiles: 0,
		},
		{
			name:   "Initialization with nil stderr",
			fs:     afero.NewMemMapFs(),
			stderr: nil,
			opts: options.DirectModeOptions{
				FilePaths: []string{"single.webp"},
			},
			expectedFiles: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockClock MockClock
			repairer := NewImageRepairerForDirectMode(tt.fs, tt.opts, tt.stderr, mockClock)

			if repairer == nil {
				t.Fatal("Resulting repairer is nil")
			}

			if repairer.fs != tt.fs {
				t.Errorf("FS mismatch: expected %p, got %p", tt.fs, repairer.fs)
			}
			if repairer.stderr != tt.stderr {
				t.Errorf("Stderr mismatch: expected %p, got %p", tt.stderr, repairer.stderr)
			}

			if repairer.stats == nil {
				t.Error("Stats should not be nil")
			} else if repairer.stats.Total != 0 {
				t.Errorf("Expected initial Total 0, got %d", repairer.stats.Total)
			}

			if !reflect.DeepEqual(repairer.options, tt.opts) {
				t.Errorf("Options mismatch.\nWant: %+v\nGot:  %+v", tt.opts, repairer.options)
			}

			if len(repairer.options.FilePaths) != tt.expectedFiles {
				t.Errorf("File count mismatch: expected %d, got %d", tt.expectedFiles, len(repairer.options.FilePaths))
			}
		})
	}
}

func TestImageRepairerForDirectMode_DeleteBackupFile(t *testing.T) {
	tests := []struct {
		name             string
		wrapFsAsReadonly bool
		setupFs          func(fs afero.Fs)
		path             string
		cancelCtx        bool
		wantErr          bool
		expectedError    string
	}{
		{
			name: "Success: file exists and is removed",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "backup.jpg.bak", []byte("data"), filesystem.DefaultFilePermissions)
			},
			path:    "backup.jpg.bak",
			wantErr: false,
		},
		{
			name:    "Success: empty path returns nil",
			path:    "",
			wantErr: false,
		},
		{
			name:          "Error: file does not exist (unexpectedly vanished)",
			setupFs:       func(fs afero.Fs) {}, // No file created here
			path:          "missing.bak",
			wantErr:       true,
			expectedError: "backup file vanished unexpectedly: missing.bak",
		},
		{
			name: "Error: context is canceled",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "interrupted.bak", []byte("data"), filesystem.DefaultFilePermissions)
			},
			path:          "interrupted.bak",
			cancelCtx:     true,
			wantErr:       true,
			expectedError: context.Canceled.Error(),
		},
		{
			name:             "Error: filesystem is read-only",
			wrapFsAsReadonly: true,
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "readonly.bak", []byte("data"), filesystem.DefaultFilePermissions)
			},
			path:    "readonly.bak",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			if tt.setupFs != nil {
				tt.setupFs(fs)
			}

			currentFs := fs
			if tt.wrapFsAsReadonly {
				currentFs = afero.NewReadOnlyFs(fs)
			}

			ir := &ImageRepairerForDirectMode{
				ImageRepairerBase: ImageRepairerBase{
					fs: currentFs,
				},
			}

			ctx := context.Background()
			if tt.cancelCtx {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(context.Background())
				cancel()
			}

			err := ir.deleteBackupFile(ctx, tt.path)

			if (err != nil) != tt.wantErr {
				t.Fatalf("deleteBackupFile() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && tt.expectedError != "" {
				if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error to contain %q, got %q", tt.expectedError, err.Error())
				}
			}

			if !tt.wantErr && tt.path != "" {
				exists, _ := afero.Exists(currentFs, tt.path)
				if exists {
					t.Errorf("File %q should have been deleted, but it still exists", tt.path)
				}
			}
		})
	}
}

func TestImageRepairerForDirectMode_CreateBackupFile(t *testing.T) {
	fixedTime := time.Date(2026, 12, 25, 17, 18, 19, 0, time.UTC)
	const expectedTimestamp = "20261225_171819"

	tests := []struct {
		name             string
		sourcePath       string
		content          string
		wrapFsAsReadonly bool
		setupFs          func(fs afero.Fs)
		cancelCtx        bool
		wantErr          bool
		expectedPath     string
		expectedErrSub   string
	}{
		{
			name:       "Success: regular copy",
			sourcePath: "photos/vacation.jpg",
			content:    "fake-image-bytes",
			setupFs: func(fs afero.Fs) {
				_ = fs.MkdirAll("photos", filesystem.DefaultFolderPermissions)
				_ = afero.WriteFile(fs, "photos/vacation.jpg", []byte("fake-image-bytes"), filesystem.DefaultFilePermissions)
			},
			wantErr:      false,
			expectedPath: filepath.Join("photos", "vacation_"+expectedTimestamp+"_backup.jpg"),
		},
		{
			name:           "Error: context already cancelled",
			sourcePath:     "any.jpg",
			setupFs:        func(fs afero.Fs) {},
			cancelCtx:      true,
			wantErr:        true,
			expectedErrSub: context.Canceled.Error(),
		},
		{
			name:           "Error: source file not found",
			sourcePath:     "non_existent.jpg",
			setupFs:        func(fs afero.Fs) {}, // Empty filesystem
			wantErr:        true,
			expectedErrSub: "file does not exist",
		},
		{
			name:       "Error: source is a directory",
			sourcePath: "my_dir",
			setupFs: func(fs afero.Fs) {
				_ = fs.MkdirAll("my_dir", filesystem.DefaultFolderPermissions)
			},
			wantErr:        true,
			expectedErrSub: "source file is not a regular file",
		},
		{
			name:             "Error: destination not writable (ReadOnly filesystem)",
			sourcePath:       "writable.jpg",
			wrapFsAsReadonly: true,
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "writable.jpg", []byte("data"), filesystem.DefaultFilePermissions)
			},
			wantErr:        true,
			expectedErrSub: "create backup file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memFs := afero.NewMemMapFs()
			if tt.setupFs != nil {
				tt.setupFs(memFs)
			}

			var finalFs afero.Fs = memFs
			if tt.wrapFsAsReadonly {
				finalFs = afero.NewReadOnlyFs(memFs)
			}

			ir := &ImageRepairerForDirectMode{
				ImageRepairerBase: ImageRepairerBase{
					fs: finalFs,
				},
				clock: MockClock{fixedTime},
			}

			ctx := context.Background()
			if tt.cancelCtx {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			backupPath, err := ir.createBackupFile(ctx, tt.sourcePath)

			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr is %v, but got error: %v", tt.wantErr, err)
			}

			if tt.wantErr {
				if tt.expectedErrSub != "" && !strings.Contains(err.Error(), tt.expectedErrSub) {
					t.Errorf("expected error to contain %q, but got %q", tt.expectedErrSub, err.Error())
				}
				return
			}

			if backupPath != tt.expectedPath {
				t.Errorf("path mismatch: expected %q, got %q", tt.expectedPath, backupPath)
			}

			exists, _ := afero.Exists(finalFs, backupPath)
			if !exists {
				t.Error("backup file was not created")
			}

			data, _ := afero.ReadFile(finalFs, backupPath)
			if string(data) != tt.content {
				t.Errorf("content corruption: expected %q, got %q", tt.content, string(data))
			}
		})
	}
}

func TestImageRepairerForDirectMode_ProcessSingleFile(t *testing.T) {
	createTestJPEG := func() []byte {
		img := image.NewRGBA(image.Rect(0, 0, 10, 10))
		var buf bytes.Buffer
		jpeg.Encode(&buf, img, nil)
		return buf.Bytes()
	}

	fixedTime := time.Date(2026, 12, 25, 17, 18, 19, 0, time.UTC)
	const expectedTimestamp = "20261225_171819"

	tests := []struct {
		name           string
		sourcePath     string
		setupFs        func(fs afero.Fs)
		cancelCtx      bool
		wantErr        bool
		expectedErrSub string
		verify         func(t *testing.T, fs afero.Fs) // Additional checks
	}{
		{
			name:       "Success: full cycle",
			sourcePath: "image.jpg",
			setupFs: func(fs afero.Fs) {
				afero.WriteFile(fs, "image.jpg", createTestJPEG(), filesystem.DefaultFilePermissions)
			},
			wantErr: false,
			verify: func(t *testing.T, fs afero.Fs) {
				backupName := "image_" + expectedTimestamp + "_backup.jpg"
				exists, _ := afero.Exists(fs, backupName)
				if exists {
					t.Error("Backup file should have been deleted after success")
				}
			},
		},
		{
			name:           "Error: backup creation fails",
			sourcePath:     "missing.jpg",
			setupFs:        func(fs afero.Fs) {}, // Empty FS triggers error in createBackupFile
			wantErr:        true,
			expectedErrSub: "create backup",
		},
		{
			name:       "Error: context canceled before start",
			sourcePath: "image.jpg",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "image.jpg", createTestJPEG(), filesystem.DefaultFilePermissions)
			},
			cancelCtx:      true,
			wantErr:        true,
			expectedErrSub: context.Canceled.Error(),
		},
		{
			name:       "Error: repair fails (read image)",
			sourcePath: "corrupt.jpg",
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "corrupt.jpg", []byte("invalid"), filesystem.DefaultFilePermissions)
			},
			wantErr:        true,
			expectedErrSub: "unknown format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			if tt.setupFs != nil {
				tt.setupFs(fs)
			}

			ir := &ImageRepairerForDirectMode{
				ImageRepairerBase: ImageRepairerBase{
					fs: fs,
				},

				clock: MockClock{fixedTime},
			}

			ctx := context.Background()
			if tt.cancelCtx {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			err := ir.ProcessSingleFile(ctx, tt.sourcePath)

			if (err != nil) != tt.wantErr {
				t.Fatalf("ProcessSingleFile() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && tt.expectedErrSub != "" {
				if !strings.Contains(err.Error(), tt.expectedErrSub) {
					t.Errorf("Expected error to contain %q, got %q", tt.expectedErrSub, err.Error())
				}
			}

			if !tt.wantErr && tt.verify != nil {
				tt.verify(t, fs)
			}
		})
	}
}
