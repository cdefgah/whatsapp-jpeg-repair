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
	const shorthandPrefix = "-"
	const fullNamePrefix = "--"

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
			initialOptions: ManagedModeOptions{},
			args:           []string{shorthandPrefix + flagUseCurrentModTimeShort},
			wantOptions: ManagedModeOptions{
				UseCurrentModificationTime: true,
			},
			wantHelp: false,
		},
		{
			name:           "Set use-current-modification-time full name",
			initialOptions: ManagedModeOptions{},
			args:           []string{fullNamePrefix + flagUseCurrentModTime},
			wantOptions: ManagedModeOptions{
				UseCurrentModificationTime: true,
			},
			wantHelp: false,
		},
		{
			name:           "Set delete-whatsapp-files shorthand",
			initialOptions: ManagedModeOptions{},
			args:           []string{shorthandPrefix + flagDeleteWhatsAppFilesShort},
			wantOptions: ManagedModeOptions{
				DeleteWhatsAppFiles: true,
			},
			wantHelp: false,
		},
		{
			name:           "Set delete-whatsapp-files full name",
			initialOptions: ManagedModeOptions{},
			args:           []string{fullNamePrefix + flagDeleteWhatsAppFiles},
			wantOptions: ManagedModeOptions{
				DeleteWhatsAppFiles: true,
			},
			wantHelp: false,
		},
		{
			name:           "Set process-nested-folders shorthand",
			initialOptions: ManagedModeOptions{},
			args:           []string{shorthandPrefix + flagPrcsNestedFoldersShort},
			wantOptions: ManagedModeOptions{
				ProcessNestedFolders: true,
			},
			wantHelp: false,
		},
		{
			name:           "Set process-nested-folders full name",
			initialOptions: ManagedModeOptions{},
			args:           []string{fullNamePrefix + flagPrcsNestedFolders},
			wantOptions: ManagedModeOptions{
				ProcessNestedFolders: true,
			},
			wantHelp: false,
		},
		{
			name:           "Set dont-wait-to-close shorthand",
			initialOptions: ManagedModeOptions{},
			args:           []string{shorthandPrefix + flagDontWaitToCloseShort},
			wantOptions: ManagedModeOptions{
				DontWaitToClose: true,
			},
			wantHelp: false,
		},
		{
			name:           "Set dont-wait-to-close full name",
			initialOptions: ManagedModeOptions{},
			args:           []string{fullNamePrefix + flagDontWaitToClose},
			wantOptions: ManagedModeOptions{
				DontWaitToClose: true,
			},
			wantHelp: false,
		},
		{
			name:           "Help flag triggered via shorthand",
			initialOptions: ManagedModeOptions{},
			args:           []string{shorthandPrefix + flagDisplayHelpShort},
			wantOptions:    ManagedModeOptions{},
			wantHelp:       true,
		},
		{
			name:           "Help flag triggered via fullname",
			initialOptions: ManagedModeOptions{},
			args:           []string{fullNamePrefix + flagDisplayHelp},
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

func TestIsManagedMode(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		skipParseCall bool // to intentionally cause situations when flags are not parsed yet
		isManagedMode bool
		wantErr       bool
	}{
		{
			name:          "Error: Flags not parsed",
			args:          []string{},
			skipParseCall: true,
			isManagedMode: false,
			wantErr:       true,
		},
		{
			name:          "Managed Mode: No arguments provided",
			args:          []string{},
			skipParseCall: false,
			isManagedMode: true,
			wantErr:       false,
		},
		{
			name:          "Managed Mode: At least one managed flag is present",
			args:          []string{"--" + flagSrcPath, "/home/user/Documents/brokenFiles"},
			skipParseCall: false,
			isManagedMode: true,
			wantErr:       false,
		},
		{
			name:          "Direct Mode: Only positional args (files)",
			args:          []string{"/home/user/Documents/brokenFiles/file1.jpg", "/home/user/Documents/archivedFiles/file002.jpg"},
			skipParseCall: false,
			isManagedMode: false,
			wantErr:       false,
		},
		{
			name:          "Not managed if only help flag provided",
			args:          []string{flagDisplayHelp},
			skipParseCall: false,
			isManagedMode: false,
			wantErr:       false,
		},
		{
			name:          "Managed Mode: mixed flags and positional args",
			args:          []string{"--" + flagSrcPath, "/home/user/Documents/brokenFiles", "file05.jpg"},
			skipParseCall: false,
			isManagedMode: true,
			wantErr:       false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			opts := ManagedModeOptions{}

			fs, _ := NewManagedFlagSet(&buf, &opts)

			if !tc.skipParseCall {
				err := fs.Parse(tc.args)
				if err != nil {
					t.Fatalf("Unexpected setup error in fs.Parse: %v", err)
				}
			}

			isManagedMode, err := IsManagedMode(tc.args, fs)

			if tc.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}

				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if isManagedMode != tc.isManagedMode {
				t.Errorf("IsManagedMode() = %v, want %v. Args: %v", isManagedMode, tc.isManagedMode, tc.args)
			}
		})
	}
}
