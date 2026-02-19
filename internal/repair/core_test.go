// SPDX-License-Identifier: GPL-3.0-only
// Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>

package repair

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/jpeg"
	"io"
	"iter"
	"os"
	"strings"
	"testing"

	"github.com/cdefgah/whatsapp-jpeg-repair/internal/filesystem"
	"github.com/spf13/afero"
	"golang.org/x/term"
)

func TestImageRepairerBase_readImage(t *testing.T) {
	createTestJPEG := func() []byte {
		img := image.NewRGBA(image.Rect(0, 0, 10, 10))
		var buf bytes.Buffer
		jpeg.Encode(&buf, img, nil)
		return buf.Bytes()
	}

	tests := []struct {
		name       string
		setup      func(fs afero.Fs)
		path       string
		ctx        context.Context
		wantErr    bool
		errContain string
	}{
		{
			name: "success",
			setup: func(fs afero.Fs) {
				afero.WriteFile(fs, "/test.jpg", createTestJPEG(), filesystem.DefaultFilePermissions)
			},
			path:    "/test.jpg",
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:       "context cancelled",
			setup:      func(fs afero.Fs) {},
			path:       "/test.jpg",
			ctx:        cancelledContext(),
			wantErr:    true,
			errContain: "context canceled",
		},
		{
			name:       "file not found",
			setup:      func(fs afero.Fs) {},
			path:       "/nonexistent.jpg",
			ctx:        context.Background(),
			wantErr:    true,
			errContain: "open image file",
		},
		{
			name: "invalid image data",
			setup: func(fs afero.Fs) {
				afero.WriteFile(fs, "/invalid.jpg", []byte("not an image"), filesystem.DefaultFilePermissions)
			},
			path:       "/invalid.jpg",
			ctx:        context.Background(),
			wantErr:    true,
			errContain: "decode image",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			tt.setup(fs)

			ir := &ImageRepairerBase{fs: fs}
			img, err := ir.readImage(tt.ctx, tt.path)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.errContain != "" && !bytes.Contains([]byte(err.Error()), []byte(tt.errContain)) {
					t.Errorf("error = %v, want containing %q", err, tt.errContain)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if img == nil {
				t.Error("expected image, got nil")
			}
		})
	}
}

func TestImageRepairerBase_writeImage(t *testing.T) {
	const readonlyFilePermissions = 0o444

	testImg := image.NewRGBA(image.Rect(0, 0, 10, 10))

	verifyJPEG := func(fs afero.Fs, path string) bool {
		data, err := afero.ReadFile(fs, path)
		if err != nil {
			return false
		}
		_, err = jpeg.Decode(bytes.NewReader(data))
		return err == nil
	}

	tests := []struct {
		name       string
		setup      func(fs afero.Fs) afero.Fs
		path       string
		ctx        context.Context
		wantErr    bool
		errContain string
	}{
		{
			name:    "success",
			setup:   func(fs afero.Fs) afero.Fs { return fs },
			path:    "/output.jpg",
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:       "context cancelled",
			setup:      func(fs afero.Fs) afero.Fs { return fs },
			path:       "/output.jpg",
			ctx:        cancelledContext(),
			wantErr:    true,
			errContain: "context canceled",
		},
		{
			name: "create file error",
			setup: func(fs afero.Fs) afero.Fs {
				fs.Mkdir("/readonly", filesystem.DefaultFolderPermissions)
				afero.WriteFile(fs, "/readonly/file.jpg", []byte{}, readonlyFilePermissions)
				return afero.NewReadOnlyFs(fs)
			},
			path:       "/readonly/file.jpg",
			ctx:        context.Background(),
			wantErr:    true,
			errContain: "create file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseFs := afero.NewMemMapFs()
			fs := tt.setup(baseFs)

			ir := &ImageRepairerBase{fs: fs}
			err := ir.writeImage(tt.ctx, tt.path, testImg)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.errContain != "" && !bytes.Contains([]byte(err.Error()), []byte(tt.errContain)) {
					t.Errorf("error = %v, want containing %q", err, tt.errContain)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !verifyJPEG(baseFs, tt.path) {
				t.Error("written file is not valid JPEG")
			}
		})
	}
}

func cancelledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func TestImageRepairerBase_HasErrors(t *testing.T) {
	tests := []struct {
		name  string
		stats *Stats
		want  bool
	}{
		{
			name:  "no errors",
			stats: &Stats{Errors: []FileError{}},
			want:  false,
		},
		{
			name:  "nil errors slice",
			stats: &Stats{Errors: nil},
			want:  false,
		},
		{
			name: "has errors",
			stats: &Stats{Errors: []FileError{
				{FilePath: "/file1.jpg", Err: errors.New("error1")},
				{FilePath: "/file2.jpg", Err: errors.New("error2")},
			}},
			want: true,
		},
		{
			name: "single error",
			stats: &Stats{Errors: []FileError{
				{FilePath: "/file.jpg", Err: errors.New("error")},
			}},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ir := &ImageRepairerBase{stats: tt.stats}
			got := ir.HasErrors()
			if got != tt.want {
				t.Errorf("HasErrors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImageRepairerBase_TextReport(t *testing.T) {
	tests := []struct {
		name     string
		stats    *Stats
		contains []string
	}{
		{
			name: "no errors",
			stats: &Stats{
				Total:    10,
				Repaired: 8,
				Failed:   0,
				Errors:   nil,
			},
			contains: []string{
				"Total: 10 file(s).",
				"Repaired: 8 file(s).",
			},
		},
		{
			name: "with errors",
			stats: &Stats{
				Total:    5,
				Repaired: 3,
				Failed:   2,
				Errors: []FileError{
					{FilePath: "/img1.jpg", Err: errors.New("decode failed")},
					{FilePath: "/img2.jpg", Err: errors.New("write failed")},
				},
			},
			contains: []string{
				"Total: 5 file(s).",
				"Repaired: 3 file(s).",
				"Failed: 2 file(s).",
				"/img1.jpg",
				"/img2.jpg",
				"decode failed",
				"write failed",
			},
		},
		{
			name: "empty stats",
			stats: &Stats{
				Total:    0,
				Repaired: 0,
				Failed:   0,
				Errors:   []FileError{},
			},
			contains: []string{
				"Total: 0 file(s).",
				"Repaired: 0 file(s).",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ir := &ImageRepairerBase{stats: tt.stats}
			report := ir.TextReport()

			for _, want := range tt.contains {
				if !strings.Contains(report, want) {
					t.Errorf("report missing %q\nGot:\n%s", want, report)
				}
			}
		})
	}
}

func TestImageRepairerBase_RegisterError(t *testing.T) {
	tests := []struct {
		name         string
		initialStats *Stats
		filePath     string
		err          error
		wantFailed   int
		wantErrors   int
		wantStderr   string
	}{
		{
			name:         "regular error",
			initialStats: &Stats{Failed: 0, Errors: []FileError{}},
			filePath:     "/img.jpg",
			err:          errors.New("decode failed"),
			wantFailed:   1,
			wantErrors:   1,
			wantStderr:   "ERROR!\n",
		},
		{
			name:         "context canceled",
			initialStats: &Stats{Failed: 0, Errors: []FileError{}},
			filePath:     "/img.jpg",
			err:          context.Canceled,
			wantFailed:   0,
			wantErrors:   0,
			wantStderr:   "CANCELED!\n",
		},
		{
			name:         "multiple errors append",
			initialStats: &Stats{Failed: 2, Errors: []FileError{{FilePath: "/a.jpg", Err: errors.New("err1")}}},
			filePath:     "/b.jpg",
			err:          errors.New("err2"),
			wantFailed:   3,
			wantErrors:   2,
			wantStderr:   "ERROR!\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stderr bytes.Buffer
			ir := &ImageRepairerBase{
				stats:  tt.initialStats,
				stderr: &stderr,
			}

			ir.RegisterError(tt.filePath, tt.err)

			if ir.stats.Failed != tt.wantFailed {
				t.Errorf("Failed = %d, want %d", ir.stats.Failed, tt.wantFailed)
			}
			if len(ir.stats.Errors) != tt.wantErrors {
				t.Errorf("Errors count = %d, want %d", len(ir.stats.Errors), tt.wantErrors)
			}
			if stderr.String() != tt.wantStderr {
				t.Errorf("stderr = %q, want %q", stderr.String(), tt.wantStderr)
			}
			if tt.wantErrors > 0 {
				lastErr := ir.stats.Errors[len(ir.stats.Errors)-1]
				if lastErr.FilePath != tt.filePath {
					t.Errorf("FilePath = %q, want %q", lastErr.FilePath, tt.filePath)
				}
				if lastErr.Err.Error() != tt.err.Error() {
					t.Errorf("Err = %v, want %v", lastErr.Err, tt.err)
				}
			}
		})
	}
}

func TestImageRepairerBase_RegisterSuccess(t *testing.T) {
	tests := []struct {
		name         string
		initialCount int
		wantRepaired int
		wantStderr   string
	}{
		{
			name:         "first success",
			initialCount: 0,
			wantRepaired: 1,
			wantStderr:   "OK\n",
		},
		{
			name:         "increment existing",
			initialCount: 5,
			wantRepaired: 6,
			wantStderr:   "OK\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stderr bytes.Buffer
			ir := &ImageRepairerBase{
				stats:  &Stats{Repaired: tt.initialCount},
				stderr: &stderr,
			}

			ir.RegisterSuccess()

			if ir.stats.Repaired != tt.wantRepaired {
				t.Errorf("Repaired = %d, want %d", ir.stats.Repaired, tt.wantRepaired)
			}
			if stderr.String() != tt.wantStderr {
				t.Errorf("stderr = %q, want %q", stderr.String(), tt.wantStderr)
			}
		})
	}
}

func TestImageRepairerBase_RegisterStart(t *testing.T) {
	tests := []struct {
		name         string
		initialTotal int
		filePath     string
		wantTotal    int
		wantStderr   string
	}{
		{
			name:         "first file",
			initialTotal: 0,
			filePath:     "/img1.jpg",
			wantTotal:    1,
			wantStderr:   "Processing file /img1.jpg .......................... ",
		},
		{
			name:         "increment existing",
			initialTotal: 3,
			filePath:     "/img2.jpg",
			wantTotal:    4,
			wantStderr:   "Processing file /img2.jpg .......................... ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stderr bytes.Buffer
			ir := &ImageRepairerBase{
				stats:  &Stats{Total: tt.initialTotal},
				stderr: &stderr,
			}

			ir.RegisterStart(tt.filePath)

			if ir.stats.Total != tt.wantTotal {
				t.Errorf("Total = %d, want %d", ir.stats.Total, tt.wantTotal)
			}
			if stderr.String() != tt.wantStderr {
				t.Errorf("stderr = %q, want %q", stderr.String(), tt.wantStderr)
			}
		})
	}
}

// mockIterator implements FilePathIterator
type mockIterator struct {
	paths []string
}

func (m *mockIterator) All(_ context.Context) iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, p := range m.paths {
			if !yield(p) {
				return
			}
		}
	}
}

// mockProcessor implements SingleFileProcessor
type mockProcessor struct {
	processErr  map[string]error
	processHook func(path string) error
	started     []string
	errors      []struct {
		path string
		err  error
	}
	successCount int
}

func (m *mockProcessor) ProcessSingleFile(_ context.Context, path string) error {
	if m.processHook != nil {
		return m.processHook(path)
	}
	return m.processErr[path]
}

func (m *mockProcessor) RegisterStart(path string) {
	m.started = append(m.started, path)
}

func (m *mockProcessor) RegisterError(path string, err error) {
	m.errors = append(m.errors, struct {
		path string
		err  error
	}{path, err})
}

func (m *mockProcessor) RegisterSuccess() {
	m.successCount++
}

func TestProcessAllFiles(t *testing.T) {
	tests := []struct {
		name         string
		paths        []string
		processErr   map[string]error
		cancelBefore int // cancel context after N files started
		wantStarted  int
		wantErrors   int
		wantSuccess  int
	}{
		{
			name:        "all success",
			paths:       []string{"/a.jpg", "/b.jpg", "/c.jpg"},
			processErr:  map[string]error{},
			wantStarted: 3,
			wantErrors:  0,
			wantSuccess: 3,
		},
		{
			name:        "some failures",
			paths:       []string{"/a.jpg", "/b.jpg", "/c.jpg"},
			processErr:  map[string]error{"/b.jpg": errors.New("decode failed")},
			wantStarted: 3,
			wantErrors:  1,
			wantSuccess: 2,
		},
		{
			name:         "context canceled before iteration",
			paths:        []string{"/a.jpg", "/b.jpg"},
			processErr:   map[string]error{},
			cancelBefore: 0,
			wantStarted:  0,
			wantErrors:   1, // interrupted error
			wantSuccess:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			if tt.cancelBefore == 0 && tt.name == "context canceled before iteration" {
				cancel()
			} else {
				defer cancel()
			}

			it := &mockIterator{paths: tt.paths}
			p := &mockProcessor{processErr: tt.processErr}

			// Cancel after N files started
			if tt.cancelBefore > 0 {
				go func() {
					for len(p.started) < tt.cancelBefore {
					}
					cancel()
				}()
			}

			ProcessAllFiles(ctx, it, p)

			if len(p.started) != tt.wantStarted {
				t.Errorf("started = %d, want %d", len(p.started), tt.wantStarted)
			}
			if len(p.errors) != tt.wantErrors {
				t.Errorf("errors = %d, want %d", len(p.errors), tt.wantErrors)
			}
			if p.successCount != tt.wantSuccess {
				t.Errorf("success = %d, want %d", p.successCount, tt.wantSuccess)
			}
		})
	}
}
func TestProcessAllFiles_ContextCanceledDuringProcessing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	it := &mockIterator{paths: []string{"/a.jpg", "/b.jpg", "/c.jpg"}}

	p := &mockProcessor{
		processHook: func(path string) error {
			if path == "/b.jpg" {
				cancel()
				return errors.New("failed")
			}
			return nil
		},
	}

	ProcessAllFiles(ctx, it, p)

	if len(p.started) != 2 {
		t.Errorf("started = %d, want 2", len(p.started))
	}
	if p.successCount != 1 {
		t.Errorf("success = %d, want 1", p.successCount)
	}
	if len(p.errors) != 1 {
		t.Errorf("errors = %d, want 1", len(p.errors))
	}
}

func TestIsInteractive(t *testing.T) {
	// Creating pipes: these are real *os.File objects, but not terminals
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	defer r.Close()
	defer w.Close()

	buf := new(bytes.Buffer)

	// Detecting is that actual test run interactive
	stdinIsTerm := term.IsTerminal(int(os.Stdin.Fd()))
	stdoutIsTerm := term.IsTerminal(int(os.Stdout.Fd()))
	isCurrentlyInteractive := stdinIsTerm && stdoutIsTerm

	tests := []struct {
		name string
		in   io.Reader
		out  io.Writer
		want bool
	}{
		{"not os files", buf, buf, false},
		{"IN is pipe, OUT is buffer", r, buf, false},
		{"IN is buffer, OUT is pipe", buf, w, false},
		{"IN and OUT are_pipes, not terminals", r, w, false},
		{"Using real stdin and stdout", os.Stdin, os.Stdout, isCurrentlyInteractive},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isInteractive(tt.in, tt.out); got != tt.want {
				t.Errorf("isInteractive() = %v, want %v", got, tt.want)
			}
		})
	}
}
