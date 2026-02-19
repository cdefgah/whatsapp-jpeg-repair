// SPDX-License-Identifier: GPL-3.0-only
// Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>

package repair

import (
	"bytes"
	"context"
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

func TestDeleteBackupFile(t *testing.T) {
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
				_ = afero.WriteFile(fs, "readonly.bak", []byte("data"), 0644)
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
					t.Errorf("File %s should have been deleted, but it still exists", tt.path)
				}
			}
		})
	}
}

func TestCreateBackupFile(t *testing.T) {
	fixedTime := time.Date(2026, 12, 25, 17, 59, 47, 0, time.UTC)
	mockClock := MockClock{FixedTime: fixedTime}

	fs := afero.NewMemMapFs()
	sourcePath := "image.jfif"
	_ = afero.WriteFile(fs, sourcePath, []byte("data"), filesystem.DefaultFilePermissions)

	ir := &ImageRepairerForDirectMode{
		ImageRepairerBase: ImageRepairerBase{
			fs: fs,
		},

		clock: mockClock,
	}

	backupPath, err := ir.createBackupFile(context.Background(), sourcePath)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedName := "image_20261225_175947_backup.jfif"
	if filepath.Base(backupPath) != expectedName {
		t.Errorf("Expected filename %q, got %q", expectedName, filepath.Base(backupPath))
	}
}
