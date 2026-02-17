// SPDX-License-Identifier: GPL-3.0-only
// Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>

package options

import (
	"bytes"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestNewDirectOptions(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name:     "Multiple file paths",
			args:     []string{"/images/photo01.jpg", "/archive/document8.jpg", "/scans/receipt003.jpg"},
			expected: []string{"/images/photo01.jpg", "/archive/document8.jpg", "/scans/receipt003.jpg"},
		},
		{
			name:     "Empty slice",
			args:     []string{},
			expected: []string{},
		},
		{
			name:     "Nil input",
			args:     nil,
			expected: nil,
		},
		{
			name:     "Single path",
			args:     []string{"/archive/scans/document01.jpg"},
			expected: []string{"/archive/scans/document01.jpg"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			opts := NewDirectOptions(tc.args)

			if !slices.Equal(opts.FilePaths, tc.expected) {
				t.Errorf("NewDirectOptions(%v).FilePaths = %v; want %v",
					tc.args, opts.FilePaths, tc.expected)
			}
		})
	}
}

func TestNewDefaultManagedModeOptions(t *testing.T) {

	tests := []struct {
		name string
		cwd  string
	}{
		{
			name: "Absolute path",
			cwd:  "/home/user/wjr",
		},
		{
			name: "Relative path",
			cwd:  "./wjr",
		},
		{
			name: "Empty string",
			cwd:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := NewDefaultManagedModeOptions(tc.cwd)

			// Проверяем пути
			expectedSrc := filepath.Join(tc.cwd, predefinedSourceFilesFolder)
			if got.SourceFolderPath != expectedSrc {
				t.Errorf("SourceFolderPath mismatch: got %q, want %q", got.SourceFolderPath, expectedSrc)
			}

			expectedDest := filepath.Join(tc.cwd, predefinedDestinationFilesFolder)
			if got.DestinationFolderPath != expectedDest {
				t.Errorf("DestinationFolderPath mismatch: got %q, want %q", got.DestinationFolderPath, expectedDest)
			}

			if got.UseCurrentModificationTime || got.DeleteWhatsAppFiles ||
				got.ProcessNestedFolders || got.DontWaitToClose {
				t.Errorf("Expected all boolean flags to be false, but got: %+v", got)
			}
		})
	}
}

func TestManagedModeOptions_String(t *testing.T) {
	tests := []struct {
		name     string
		opts     ManagedModeOptions
		mustHave map[string]string
	}{
		{
			name: "All values not set",
			opts: ManagedModeOptions{},
			mustHave: map[string]string{
				"Source folder path:":            "",
				"Destination folder path:":       "",
				"Use current modification time:": "false",
				"Delete WhatsApp files:":         "false",
				"Process nested folders:":        "false",
				"Don't wait to close:":           "false",
			},
		},
		{
			name: "All values set",
			opts: ManagedModeOptions{
				SourceFolderPath:           "/input/dir",
				DestinationFolderPath:      "/output/dir",
				UseCurrentModificationTime: true,
				DeleteWhatsAppFiles:        true,
				ProcessNestedFolders:       true,
				DontWaitToClose:            true,
			},
			mustHave: map[string]string{
				"Source folder path:":            "/input/dir",
				"Destination folder path:":       "/output/dir",
				"Use current modification time:": "true",
				"Delete WhatsApp files:":         "true",
				"Process nested folders:":        "true",
				"Don't wait to close:":           "true",
			},
		},
	}

	const expectedNumberOfOutputLines = 6

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.opts.String()
			lines := strings.Split(strings.TrimSpace(got), "\n")

			if len(lines) != expectedNumberOfOutputLines {
				t.Fatalf("Expected %d lines in output, got %d", expectedNumberOfOutputLines, len(lines))
			}

			for key, expectedValue := range tc.mustHave {
				found := false
				for _, line := range lines {
					if strings.Contains(line, key) && strings.Contains(line, expectedValue) {
						found = true
						break
					}
				}

				if !found {
					t.Errorf("Could not find line with key %q and value %q.\nOutput:\n%s", key, expectedValue, got)
				}
			}
		})
	}
}

func TestNewManagedFlagSet(t *testing.T) {
	// const (
	// 	currentFolder = "/home/user/Documents/wjr"
	// )

	// defaultSourceFolderPath := filepath.Join(currentFolder, predefinedSourceFilesFolder)
	// defaultDestinationFolderPath := filepath.Join(currentFolder, predefinedDestinationFilesFolder)

	type testCase struct {
		name           string
		initialOptions ManagedModeOptions
		args           []string
		wantOptions    ManagedModeOptions
		wantHelp       bool
		wantErr        bool
	}

	tests := []testCase{
		{
			name: "Default values (no args)",
			initialOptions: ManagedModeOptions{
				SourceFolderPath:      "home/user/Documents/BrokenFiles",
				DestinationFolderPath: "home/user/Documents/FixedFiles",
			},
			args: []string{},
			wantOptions: ManagedModeOptions{
				SourceFolderPath:      "home/user/Documents/BrokenFiles",
				DestinationFolderPath: "home/user/Documents/FixedFiles",
			},
			wantHelp: false,
		},
		{
			name:           "Set use-current-modification-time shorthand",
			initialOptions: ManagedModeOptions{DeleteWhatsAppFiles: false},
			args:           []string{"-t"},
			wantOptions: ManagedModeOptions{
				UseCurrentModificationTime: true,
			},
			wantHelp: false,
		},
		{
			name:           "Set use-current-modification-time full name",
			initialOptions: ManagedModeOptions{DeleteWhatsAppFiles: false},
			args:           []string{"--use-current-modification-time"},
			wantOptions: ManagedModeOptions{
				UseCurrentModificationTime: true,
			},
			wantHelp: false,
		},
		{
			name:           "Set delete-whatsapp-files shorthand",
			initialOptions: ManagedModeOptions{DeleteWhatsAppFiles: false},
			args:           []string{"-w"},
			wantOptions: ManagedModeOptions{
				DeleteWhatsAppFiles: true,
			},
			wantHelp: false,
		},
		{
			name:           "Set delete-whatsapp-files full name",
			initialOptions: ManagedModeOptions{DeleteWhatsAppFiles: false},
			args:           []string{"--delete-whatsapp-files"},
			wantOptions: ManagedModeOptions{
				DeleteWhatsAppFiles: true,
			},
			wantHelp: false,
		},
		{
			name:           "Set delete-whatsapp-files shorthand",
			initialOptions: ManagedModeOptions{DeleteWhatsAppFiles: false},
			args:           []string{"-w"},
			wantOptions: ManagedModeOptions{
				DeleteWhatsAppFiles: true,
			},
			wantHelp: false,
		},
		{
			name:           "Set delete-whatsapp-files full name",
			initialOptions: ManagedModeOptions{DeleteWhatsAppFiles: false},
			args:           []string{"--delete-whatsapp-files"},
			wantOptions: ManagedModeOptions{
				DeleteWhatsAppFiles: true,
			},
			wantHelp: false,
		},
		{
			name:           "Set process-nested-folders shorthand",
			initialOptions: ManagedModeOptions{DeleteWhatsAppFiles: false},
			args:           []string{"-n"},
			wantOptions: ManagedModeOptions{
				ProcessNestedFolders: true,
			},
			wantHelp: false,
		},
		{
			name:           "Set process-nested-folders full name",
			initialOptions: ManagedModeOptions{DeleteWhatsAppFiles: false},
			args:           []string{"--process-nested-folders"},
			wantOptions: ManagedModeOptions{
				ProcessNestedFolders: true,
			},
			wantHelp: false,
		},
		{
			name:           "Set dont-wait-to-close shorthand",
			initialOptions: ManagedModeOptions{DeleteWhatsAppFiles: false},
			args:           []string{"-w"},
			wantOptions: ManagedModeOptions{
				DontWaitToClose: true,
			},
			wantHelp: false,
		},
		{
			name:           "Set dont-wait-to-close full name",
			initialOptions: ManagedModeOptions{DeleteWhatsAppFiles: false},
			args:           []string{"--dont-wait-to-close"},
			wantOptions: ManagedModeOptions{
				DontWaitToClose: true,
			},
			wantHelp: false,
		},
		{
			name:           "Help flag triggered via shorthand",
			initialOptions: ManagedModeOptions{},
			args:           []string{"-h"},
			wantOptions:    ManagedModeOptions{},
			wantHelp:       true,
		},
		{
			name:           "Help flag triggered via fullname",
			initialOptions: ManagedModeOptions{},
			args:           []string{"--help"},
			wantOptions:    ManagedModeOptions{},
			wantHelp:       true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			opts := tc.initialOptions

			fs, displayHelp := NewManagedFlagSet(&buf, &opts)
			err := fs.Parse(tc.args)

			if (err != nil) != tc.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tc.wantErr)
			}

			if *displayHelp != tc.wantHelp {
				t.Errorf("displayHelp = %v, want %v", *displayHelp, tc.wantHelp)
			}

			if opts != tc.wantOptions {
				t.Errorf("Resulting options = %+v, want %+v", opts, tc.wantOptions)
			}
		})
	}
}
