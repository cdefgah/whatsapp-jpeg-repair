// SPDX-License-Identifier: GPL-3.0-only
// Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>

package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cdefgah/whatsapp-jpeg-repair/internal/filesystem"
	"github.com/cdefgah/whatsapp-jpeg-repair/internal/options"
	"github.com/cdefgah/whatsapp-jpeg-repair/internal/repair"
	"github.com/cdefgah/whatsapp-jpeg-repair/internal/testutil"
	"github.com/spf13/afero"
)

func TestEnv_runApp(t *testing.T) {
	fixedTime := time.Date(2026, 2, 20, 12, 0, 0, 0, time.UTC)

	const (
		fakeExePath = "/home/user/Documents/WJR/WhatsAppJpegRepair"
		fakeExeDir  = "/home/user/Documents/WJR"
	)

	tests := []struct {
		name          string
		args          []string
		setupFs       func(fs afero.Fs)
		expectedInErr string
		wantErr       bool
	}{
		{
			name: "Managed Mode: successfull procesing",
			args: []string{},
			setupFs: func(fs afero.Fs) {
				_ = fs.MkdirAll(filepath.Join(fakeExeDir, options.PredefinedSourceFilesFolder), filesystem.DefaultFolderPermissions)
			},
			expectedInErr: "managed mode",
			wantErr:       false,
		},
		{
			name: "Direct Mode: successfull procesing",
			args: []string{"test_image.jpg"},
			setupFs: func(fs afero.Fs) {
				testutil.CreateJpegFile(fs, "test_image.jpg")
			},
			expectedInErr: "direct mode",
			wantErr:       false,
		},
		{
			name:          "Displaying help",
			args:          []string{"--help"},
			setupFs:       func(fs afero.Fs) {},
			expectedInErr: "Usage:",
			wantErr:       false,
		},
		{
			name:          "Incorrect combination of managed mode flags with positional arguments",
			args:          []string{"--" + options.FlagSrcPath, "/some/dir", "extra-file.jpg"},
			setupFs:       func(fs afero.Fs) {},
			expectedInErr: "Usage:",
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			if tt.setupFs != nil {
				tt.setupFs(fs)
			}

			var stderr bytes.Buffer
			stdin := strings.NewReader("")
			mClock := &testutil.MockClock{FixedTime: fixedTime}

			env := &Env{
				fs:          fs,
				stderr:      &stderr,
				stdin:       stdin,
				exeFilePath: fakeExePath,
				args:        tt.args,
				clock:       mClock,
			}

			err := env.runApp()

			if (err != nil) != tt.wantErr {
				t.Errorf("runApp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.expectedInErr != "" {
				gotOutput := stderr.String()
				if !strings.Contains(gotOutput, tt.expectedInErr) {
					t.Errorf("Output does not contain expected string %q.\nGot:\n%s", tt.expectedInErr, gotOutput)
				}
			}
		})
	}
}

func TestNewEnv(t *testing.T) {
	var stderr bytes.Buffer
	stdin := strings.NewReader("")
	exePath := "/home/user/Documents/WJR/WhatsAppJpegRepair"
	args := []string{"arg1", "arg2"}

	env := NewEnv(&stderr, stdin, exePath, args)

	if env.stderr != &stderr {
		t.Error("stderr is not initialized properly")
	}

	if env.stdin != stdin {
		t.Error("stdin is not initialized properly")
	}

	if env.exeFilePath != exePath {
		t.Errorf("want exeFilePath %q, got %q", exePath, env.exeFilePath)
	}

	if len(env.args) != len(args) {
		t.Errorf("want %d arguments, got %d", len(args), len(env.args))
	} else {
		for i, v := range args {
			if env.args[i] != v {
				t.Errorf("Argument on position %d: want %q, got %q", i, v, env.args[i])
			}
		}
	}

	if _, ok := env.fs.(*afero.OsFs); !ok {
		t.Errorf("want type fs *afero.OsFs, got %T", env.fs)
	}

	if _, ok := env.clock.(repair.RealClock); !ok {
		t.Errorf("want type repair.RealClock, got %T", env.clock)
	}
}
