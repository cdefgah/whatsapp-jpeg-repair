// SPDX-License-Identifier: GPL-3.0-only
// Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>

package app

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/cdefgah/whatsapp-jpeg-repair/internal/filesystem"
	"github.com/cdefgah/whatsapp-jpeg-repair/internal/options"
	"github.com/cdefgah/whatsapp-jpeg-repair/internal/repair"
	"github.com/cdefgah/whatsapp-jpeg-repair/internal/testutil"
	"github.com/spf13/afero"
)

func TestNewAppRunner(t *testing.T) {
	memFs := afero.NewMemMapFs()
	var buf bytes.Buffer
	fixedTime := time.Date(2026, 12, 25, 17, 18, 19, 0, time.UTC)
	mClock := &testutil.MockClock{FixedTime: fixedTime}

	tests := []struct {
		name   string
		fs     afero.Fs
		stderr io.Writer
		clock  repair.Clock
	}{
		{
			name:   "Initialization with all deps",
			fs:     memFs,
			stderr: &buf,
			clock:  mClock,
		},
		{
			name:   "Initialization with nil",
			fs:     nil,
			stderr: nil,
			clock:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := NewAppRunner(tt.fs, tt.stderr, tt.clock)

			if runner == nil {
				t.Fatal("NewAppRunner returned nil")
			}

			if runner.fs != tt.fs {
				t.Errorf("fs: want %v, got %v", tt.fs, runner.fs)
			}

			if runner.stderr != tt.stderr {
				t.Errorf("stderr: want %v, got %v", tt.stderr, runner.stderr)
			}

			if runner.clock != tt.clock {
				t.Errorf("clock: want %v, got %v", tt.clock, runner.clock)
			}
		})
	}
}

func TestNewGlobalProcessParams(t *testing.T) {
	inputContent := "test input"
	testStdin := strings.NewReader(inputContent)
	testExePath := "/usr/bin/app"
	testArgs := []string{"image1.jpg", "image2.jpg", "image3.jpg"}

	tests := []struct {
		name          string
		stdin         io.Reader
		exeFolderPath string
		args          []string
	}{
		{
			name:          "Standard initialization",
			stdin:         testStdin,
			exeFolderPath: testExePath,
			args:          testArgs,
		},
		{
			name:          "Empty arguments",
			stdin:         nil,
			exeFolderPath: "",
			args:          []string{},
		},
		{
			name:          "Passing Nil for args",
			stdin:         testStdin,
			exeFolderPath: ".",
			args:          nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewGlobalProcessParams(tt.stdin, tt.exeFolderPath, tt.args)

			if got == nil {
				t.Fatal("NewGlobalProcessParams() returned nil")
			}

			if got.Stdin != tt.stdin {
				t.Errorf("Stdin: got %v, want %v", got.Stdin, tt.stdin)
			}

			if got.ExeFolderPath != tt.exeFolderPath {
				t.Errorf("ExeFolderPath: got %q, want %q", got.ExeFolderPath, tt.exeFolderPath)
			}

			if !slices.Equal(got.ArgsWithoutAppName, tt.args) {
				t.Errorf("ArgsWithoutAppName: got %v, want %v", got.ArgsWithoutAppName, tt.args)
			}
		})
	}
}

func createTestImage(fs afero.Fs, path string) {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})
	f, _ := fs.Create(path)
	defer f.Close()
	_ = jpeg.Encode(f, img, nil)
}

func TestRunner_RunAppInDirectMode(t *testing.T) {
	fixedTime := time.Date(2026, 2, 20, 14, 0, 0, 0, time.UTC)
	mClock := &testutil.MockClock{FixedTime: fixedTime}

	tests := []struct {
		name          string
		filePaths     []string
		setupFs       func(fs afero.Fs)
		wantErr       bool
		expectedStats []string
	}{
		{
			name:      "Successful repair: backup created and deleted, file overwritten",
			filePaths: []string{"/photo.jpg"},
			setupFs: func(fs afero.Fs) {
				createTestImage(fs, "/photo.jpg")
			},
			wantErr: false,
			expectedStats: []string{
				"Total: 1 file(s)",
				"Repaired: 1 file(s)",
			},
		},
		{
			name:      "Error: File is not an image (Decode error)",
			filePaths: []string{"/bad.jpg"},
			setupFs: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, "/bad.jpg", []byte("this is not a jpeg"), filesystem.DefaultFilePermissions)
			},
			wantErr: true,
			expectedStats: []string{
				"Total: 1 file(s)",
				"Failed: 1 file(s)",
				"decode image",
			},
		},
		{
			name:      "Error: file does not exist",
			filePaths: []string{"/missing.jpg"},
			setupFs:   func(fs afero.Fs) {}, // empty filesystem
			wantErr:   true,
			expectedStats: []string{
				"Total: 1 file(s)",
				"Failed: 1 file(s)",
				"file does not exist",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			if tt.setupFs != nil {
				tt.setupFs(fs)
			}

			var stderr bytes.Buffer
			r := &Runner{
				fs:     fs,
				stderr: &stderr,
			}

			opts := options.DirectModeOptions{
				FilePaths: tt.filePaths,
			}

			err := r.runAppInDirectMode(context.Background(), opts, mClock)

			if (err != nil) != tt.wantErr {
				t.Errorf("runAppInDirectMode() error = %v, wantErr %v", err, tt.wantErr)
			}

			output := stderr.String()
			for _, s := range tt.expectedStats {
				if !strings.Contains(output, s) {
					t.Errorf("Report does not contain expected string %q. Full report:\n%s", s, output)
				}
			}

			if !tt.wantErr && len(tt.filePaths) > 0 {
				files, _ := afero.ReadDir(fs, "/")
				for _, f := range files {
					if strings.Contains(f.Name(), "_backup") {
						t.Errorf("temporary backup file %s was not deleted!", f.Name())
					}
				}
			}
		})
	}
}

func TestRunner_RunAppInManagedMode(t *testing.T) {
	srcDir := "/src"
	dstDir := "/dst"

	tests := []struct {
		name          string
		options       options.ManagedModeOptions
		setupFs       func(fs afero.Fs)
		wantErr       bool
		errSubstring  string
		expectedStats []string
	}{
		{
			name: "Success: recursive find and repair",
			options: options.ManagedModeOptions{
				SourceFolderPath:      srcDir,
				DestinationFolderPath: dstDir,
				ProcessNestedFolders:  true,
				DontWaitToClose:       true,
			},
			setupFs: func(fs afero.Fs) {
				_ = fs.MkdirAll(srcDir, filesystem.DefaultFolderPermissions)
				createTestImage(fs, filepath.Join(srcDir, "root.jpg"))
			},
			wantErr:       false,
			expectedStats: []string{"Total: 1", "Repaired: 1"},
		},
		{
			name: "Processing finished with errors in stats",
			options: options.ManagedModeOptions{
				SourceFolderPath:      srcDir,
				DestinationFolderPath: dstDir,
				DontWaitToClose:       true,
			},
			setupFs: func(fs afero.Fs) {
				_ = fs.MkdirAll(srcDir, filesystem.DefaultFolderPermissions)
				_ = afero.WriteFile(fs, filepath.Join(srcDir, "corrupted.jpg"), []byte("invalid jpeg content"), filesystem.DefaultFilePermissions)
			},
			wantErr:      true,
			errSubstring: "the processing of image files in managed mode has failed",
			expectedStats: []string{
				"Total: 1",
				"Failed: 1",
			},
		},
		{
			name: "Error: source folder path not found",
			options: options.ManagedModeOptions{
				SourceFolderPath: "/non_existent",
				DontWaitToClose:  true,
			},
			setupFs: func(fs afero.Fs) {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			fs := afero.NewMemMapFs()
			if tt.setupFs != nil {
				tt.setupFs(fs)
			}

			var stderr bytes.Buffer
			stdin := strings.NewReader("\n")
			r := &Runner{fs: fs, stderr: &stderr}

			err := r.runAppInManagedMode(ctx, stdin, tt.options)

			if (err != nil) != tt.wantErr {
				t.Errorf("runAppInManagedMode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errSubstring != "" {
				if !strings.Contains(err.Error(), tt.errSubstring) {
					t.Errorf("expected error that contains %q, but get %q", tt.errSubstring, err.Error())
				}
			}

			if tt.expectedStats != nil {
				output := stderr.String()
				for _, s := range tt.expectedStats {
					if !strings.Contains(output, s) {
						t.Errorf("report does not contain expected substring: %q", s)
					}
				}
			}
		})
	}
}

func TestRunner_ProcessCommandLineArguments(t *testing.T) {
	tests := []struct {
		name           string
		params         cliProcessParams
		setupFs        func(fs afero.Fs)
		wantErr        bool
		expectedOutput string
	}{
		{
			name: "Managed mode: by default, without cli arguments",
			params: cliProcessParams{
				ExeFolderPath:      "/app",
				ArgsWithoutAppName: []string{},
			},
			setupFs: func(fs afero.Fs) {
				_ = fs.MkdirAll("/app/"+options.PredefinedSourceFilesFolder, filesystem.DefaultFolderPermissions)
			},
			wantErr:        false,
			expectedOutput: "Now the application runs in managed mode",
		},
		{
			name: "Managed mode: source path flag specified",
			params: cliProcessParams{
				ExeFolderPath:      "/app",
				ArgsWithoutAppName: []string{"--" + options.FlagSrcPath, "/custom/source"},
			},
			setupFs: func(fs afero.Fs) {
				_ = fs.MkdirAll("/custom/source", filesystem.DefaultFolderPermissions)
			},
			wantErr:        false,
			expectedOutput: filepath.Clean("/custom/source"),
		},
		{
			name: "Direct mode: only positional arguments passed",
			params: cliProcessParams{
				ArgsWithoutAppName: []string{"file1.jpg", "file2.jpg"},
			},
			setupFs: func(fs afero.Fs) {
				createTestImage(fs, "file1.jpg")
				createTestImage(fs, "file2.jpg")
			},
			wantErr:        false,
			expectedOutput: "application runs in direct mode",
		},
		{
			name: "Displaying help",
			params: cliProcessParams{
				ArgsWithoutAppName: []string{"--" + options.FlagDisplayHelp},
			},
			wantErr:        false,
			expectedOutput: "Usage:",
		},
		{
			name: "Mixed managed flags with positional arguments",
			params: cliProcessParams{
				ArgsWithoutAppName: []string{"--" + options.FlagSrcPath, "/src", "extra-file.jpg"},
			},
			wantErr:        false, // no error, but Usage must be printed
			expectedOutput: "Usage:",
		},
		{
			name: "Incorrect managed mode flag",
			params: cliProcessParams{
				ArgsWithoutAppName: []string{"--unknown-flag"},
			},
			wantErr:        false,
			expectedOutput: "Usage:",
		},
	}

	fixedTime := time.Date(2026, 12, 25, 17, 18, 19, 0, time.UTC)
	mClock := &testutil.MockClock{FixedTime: fixedTime}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			fs := afero.NewMemMapFs()
			if tt.setupFs != nil {
				tt.setupFs(fs)
			}

			var stderr bytes.Buffer

			r := &Runner{
				fs:     fs,
				stderr: &stderr,
				clock:  mClock,
			}

			err := r.ProcessCommandLineArguments(ctx, tt.params)

			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessCommandLineArguments() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			output := stderr.String()
			if tt.expectedOutput != "" && !strings.Contains(output, tt.expectedOutput) {
				t.Errorf("Output mismatch.\nExpected substring: %q\nActual output: %q", tt.expectedOutput, output)
			}
		})
	}
}
