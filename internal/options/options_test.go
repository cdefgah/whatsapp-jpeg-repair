// SPDX-License-Identifier: GPL-3.0-only
// Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>

package options

import (
	"bytes"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

func TestNewDirectOptions(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name:     "Multiple file paths",
			args:     []string{"/images/photo01.jpg", "/images/photo02.jpg", "/images/photo03.jpg"},
			expected: []string{"/images/photo01.jpg", "/images/photo02.jpg", "/images/photo03.jpg"},
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
	assertFlag := func(fs *pflag.FlagSet, name string, wantDefault, wantActual any) {
		t.Helper() // Calling t.Helper() to have more clear logs if test failed

		if !fs.Parsed() {
			t.Fatal("flag set must be parsed before asserting a flag")
			return
		}

		// Checking flag presence
		f := fs.Lookup(name)
		if f == nil {
			t.Fatalf("flag %q not found in FlagSet", name)
			return
		}

		// Determine the type of expected value and perform the appropriate checks
		switch expectedDef := wantDefault.(type) {

		case string:
			// Checking default values for strings
			if f.DefValue != expectedDef {
				t.Fatalf("flag %q default value mismatch.\nGot: %q\nWant: %q", name, f.DefValue, expectedDef)
			}

			// Checking that actual value is also string (for cases when different types passed for default and actual values)
			expectedAct, ok := wantActual.(string)
			if !ok {
				t.Fatalf("type mismatch for flag %q: wantDefault is string, but wantActual is %T", name, wantActual)
			}

			// Checking actual flag value for string
			if val := f.Value.String(); val != expectedAct {
				t.Fatalf("flag %q actual value mismatch.\nGot: %q\nWant: %q", name, val, expectedAct)
			}

		case bool:
			// pflag stores DefValue as string ("true"/"false")
			parsedDef, err := strconv.ParseBool(f.DefValue)
			if err != nil {
				t.Fatalf("flag %q has invalid boolean default value string: %q", name, f.DefValue)
			}

			// Checking default values here
			if parsedDef != expectedDef {
				t.Fatalf("flag %q default value mismatch.\nGot: %v\nWant: %v", name, parsedDef, expectedDef)
			}

			// Checking that actual value is also bool (for cases when different types passed for default and actual values)
			expectedAct, ok := wantActual.(bool)
			if !ok {
				t.Fatalf("type mismatch for flag %q: wantDefault is bool, but wantActual is %T", name, wantActual)
			}

			// Checking actual flag value via GetBool to handle "true", "True", "1", etc.
			val, err := fs.GetBool(name)
			if err != nil {
				t.Fatalf("failed to get bool value for flag %q: %v", name, err)
			}

			if val != expectedAct {
				t.Fatalf("flag %q actual value mismatch.\nGot: %v\nWant: %v", name, val, expectedAct)
			}

		default:
			t.Fatalf("unsupported type for flag check: %T. Only string and bool are supported.", wantDefault)
		}
	}

	const currentFolder = "/home/user/Documents/wjr"
	defaultSourceFolderPath := filepath.Join(currentFolder, predefinedSourceFilesFolder)
	defaultDestinationFolderPath := filepath.Join(currentFolder, predefinedDestinationFilesFolder)

	t.Run("DefaultValues", func(t *testing.T) {
		opts := NewDefaultManagedModeOptions(currentFolder)

		out := new(bytes.Buffer)
		fs, _ := NewManagedFlagSet(out, opts)

		args := []string{} // no passed arguments
		fs.Parse(args)

		assertFlag(fs, flagSrcPath, defaultSourceFolderPath, defaultSourceFolderPath)
		assertFlag(fs, flagDestPath, defaultDestinationFolderPath, defaultDestinationFolderPath)
		assertFlag(fs, flagUseCurrentModTime, false, false)
		assertFlag(fs, flagDeleteWhatsAppFiles, false, false)
		assertFlag(fs, flagDontWaitToClose, false, false)
		assertFlag(fs, flagPrcsNestedSrcFolders, false, false)
		assertFlag(fs, flagDisplayHelp, false, false)
	})
}
