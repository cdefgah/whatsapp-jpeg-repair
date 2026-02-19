// SPDX-License-Identifier: GPL-3.0-only
// Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>

package repair

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"image"
	"image/jpeg"

	"github.com/cdefgah/whatsapp-jpeg-repair/internal/filesystem"
	"github.com/spf13/afero"
	"golang.org/x/term"
)

// FileError associates a specific file path with the error that occurred during its processing.
type FileError struct {
	FilePath string
	Err      error
}

// Error implements the error interface.
func (fe FileError) Error() string {
	return fmt.Sprintf("%s: %v", fe.FilePath, fe.Err)
}

// Stats holds the results of a batch image repair operation.
type Stats struct {
	Errors   []FileError
	Total    int
	Repaired int
	Failed   int
}

// ImageRepairerBase provides a foundation for repairing images
// in both direct and managed modes.
type ImageRepairerBase struct {
	fs     afero.Fs
	stats  *Stats
	stderr io.Writer
}

// SingleFileProcessor defines the contract for processing individual files
// and reporting the results of those operations.
type SingleFileProcessor interface {
	ProcessSingleFile(ctx context.Context, path string) error
	RegisterStart(path string)
	RegisterError(path string, err error)
	RegisterSuccess()
}

// readImage opens and decodes an image from the specified path.
func (ir *ImageRepairerBase) readImage(ctx context.Context, path string) (image.Image, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	file, err := ir.fs.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open image file: %w", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("decode image %s: %w", path, err)
	}

	return img, nil
}

// writeImage saves the image in JPEG format to the specified path.
func (ir *ImageRepairerBase) writeImage(ctx context.Context, filePath string, img image.Image) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	file, err := ir.fs.Create(filePath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	bw := bufio.NewWriter(file)
	errEncode := jpeg.Encode(bw, img, nil)

	if errEncode != nil {
		return fmt.Errorf("encode %s: %w", filePath, errEncode)
	}

	if err := bw.Flush(); err != nil {
		return fmt.Errorf("flush buffer: %w", err)
	}

	return ctx.Err()
}

// HasErrors returns true if there's at least one error present in repair stats.
func (ir *ImageRepairerBase) HasErrors() bool {
	return len(ir.stats.Errors) > 0
}

// TextReport returns repair statistics as a formatted string report.
func (ir *ImageRepairerBase) TextReport() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "\nTotal: %d file(s).", ir.stats.Total)
	fmt.Fprintf(&sb, "\nRepaired: %d file(s).", ir.stats.Repaired)

	if ir.HasErrors() {
		fmt.Fprintf(&sb, "\nFailed: %d file(s). Error details:\n", ir.stats.Failed)
		for _, fe := range ir.stats.Errors {
			fmt.Fprintf(&sb, "  - %v\n", fe)
		}
	}

	return sb.String()
}

// RegisterError registers file processing error and outputs it to the writer.
func (ir *ImageRepairerBase) RegisterError(filePath string, err error) {
	if errors.Is(err, context.Canceled) {
		fmt.Fprintln(ir.stderr, "CANCELED!")
		return
	}

	ir.stats.Failed++

	ir.stats.Errors = append(ir.stats.Errors, FileError{
		FilePath: filePath,
		Err:      err,
	})

	fmt.Fprintln(ir.stderr, "ERROR!")
}

// RegisterSuccess registers that file processing succeeded.
func (ir *ImageRepairerBase) RegisterSuccess() {
	ir.stats.Repaired++

	fmt.Fprintln(ir.stderr, "OK")
}

// RegisterStart registers start repairing of a file
func (ir *ImageRepairerBase) RegisterStart(filePath string) {
	ir.stats.Total++
	fmt.Fprintf(ir.stderr, "Processing file %s .......................... ", filePath)
}

// ProcessAllFiles processes all files using the provided iterator and processor.
// It respects context cancellation (e.g., Ctrl+C) at both the file iteration and processing levels.
func ProcessAllFiles(ctx context.Context, it filesystem.FilePathIterator, p SingleFileProcessor) {
	for path := range it.All(ctx) {
		if err := ctx.Err(); err != nil {
			p.RegisterError("", fmt.Errorf("process interrupted: %w", err))
			return
		}

		p.RegisterStart(path)

		if err := p.ProcessSingleFile(ctx, path); err != nil {
			p.RegisterError(path, err)

			// if Ctrl+C pressed inside of ProcessSingleFile
			// stopping the processing loop
			if ctx.Err() != nil {
				return
			}
			continue
		}

		p.RegisterSuccess()
	}
}

// isInteractive returns true if app is running in interactive mode
func isInteractive(in io.Reader, out io.Writer) bool {
	fIn, okIn := in.(*os.File)
	if !okIn || !term.IsTerminal(int(fIn.Fd())) {
		return false
	}

	fOut, okOut := out.(*os.File)
	if !okOut || !term.IsTerminal(int(fOut.Fd())) {
		return false
	}

	return true
}

// RunAndWaitForExit awaits for "Enter" key press or context cancellation.
func RunAndWaitForExit(ctx context.Context, stdin io.Reader, stderr io.Writer, dontWait bool) {
	if !isInteractive(stdin, stderr) || dontWait || ctx.Err() != nil {
		return
	}

	fmt.Fprintln(stderr, "\nProcessing is complete. Press Enter to exit.")

	// Creating a channel to receive a signal when required key is pressed
	signalIsReceived := make(chan struct{})

	scanner := bufio.NewScanner(stdin)

	go func() {
		_ = scanner.Scan()
		close(signalIsReceived)
	}()

	select {
	case <-ctx.Done():
		// If pressed Ctrl+C sequence
		// Please note that we are not closing the input stream here.
		// The goroutine above will leak, but since the application is about to exit,
		// the operating system will reclaim and free all resources immediately.
		return
	case <-signalIsReceived:
		// If pressed Enter key
		return
	}
}
