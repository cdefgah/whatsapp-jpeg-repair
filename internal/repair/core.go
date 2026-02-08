// SPDX-License-Identifier: GPL-3.0-only
// Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>

package repair

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"image"
	"image/jpeg"

	"github.com/cdefgah/whatsapp-jpeg-repair/internal/filesystem"
	"github.com/spf13/afero"
)

const defaultFolderPermissions = 0755
const defaultFilePermissions = 0644

// FileError associates a specific file path with the error that occurred during its processing.
type FileError struct {
	FilePath string
	Err      error
}

// Error implements the error interface.
func (fe FileError) Error() string {
	return fmt.Sprintf("%s: %v", fe.FilePath, fe.Err)
}

// RepairStats holds the results of a batch image repair operation.
type RepairStats struct {
	Errors    []FileError
	Processed int
	Failed    int
}

// ImageRepairerBase provides a foundation for repairing images
// in both direct and managed modes.
type ImageRepairerBase struct {
	fs     afero.Fs
	stats  *RepairStats
	out    io.Writer
	errOut io.Writer
}

// SingleFileProcessor defines the contract for processing individual files
// and reporting the results of those operations.
type SingleFileProcessor interface {
	ProcessSingleFile(ctx context.Context, path string) error
	DisplayStart(p SingleFileProcessor, path string)
	RegisterError(p SingleFileProcessor, path string, err error)
	RegisterSuccess(p SingleFileProcessor)
	DontShowProgress() bool
}

// readImage opens and decodes an image from the specified path.
func (ir *ImageRepairerBase) readImage(ctx context.Context, path string) (image.Image, error) {
	// Checking if process interrupted by Ctrl+C
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	file, err := ir.fs.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open image file: %w", err)
	}
	defer file.Close()

	// Checking if process interrupted by Ctrl+C
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	reader := bufio.NewReader(file)

	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("decode image %s: %w", path, err)
	}

	return img, nil
}

// writeImage saves the image in JPEG format to the specified path.
func (ir *ImageRepairerBase) writeImage(ctx context.Context, filePath string, img image.Image) error {
	// Checking if process interrupted by Ctrl+C
	if err := ctx.Err(); err != nil {
		return err
	}

	file, err := ir.fs.Create(filePath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	// Checking if process interrupted by Ctrl+C
	if err := ctx.Err(); err != nil {
		return err
	}

	bw := bufio.NewWriter(file)
	errEncode := jpeg.Encode(bw, img, nil)

	// Checking if process interrupted by Ctrl+C
	if err := ctx.Err(); err != nil {
		return err
	}

	if errEncode != nil {
		return fmt.Errorf("encode %s: %w", filePath, errEncode)
	}

	if err := bw.Flush(); err != nil {
		return fmt.Errorf("flush buffer: %w", err)
	}

	return nil
}

// HasErrors returns true if there's at least one error present in repair stats.
func (ir *ImageRepairerBase) HasErrors() bool {
	return len(ir.stats.Errors) > 0
}

// TextReport returns repair statistics as a formatted string report.
func (ir *ImageRepairerBase) TextReport() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "\nProcessed: %d file(s)\n", ir.stats.Processed)

	if ir.HasErrors() {
		fmt.Fprintf(&sb, "\nFailed: %d file(s)\n", ir.stats.Failed)
		sb.WriteString("Errors:\n")

		for _, fe := range ir.stats.Errors {
			fmt.Fprintf(&sb, "  - %v\n", fe)
		}
	}

	return sb.String()
}

// prints progress message if it is allowed by actual options set.
func (ir *ImageRepairerBase) printProgressMessage(p SingleFileProcessor, message string) {
	if p.DontShowProgress() {
		return
	}

	fmt.Fprint(ir.errOut, message)
}

// RegisterError registers file processing error and outputs it to the writer.
func (ir *ImageRepairerBase) RegisterError(p SingleFileProcessor, filePath string, err error) {
	if errors.Is(err, context.Canceled) {
		ir.printProgressMessage(p, "CANCELED!\n")
		return
	}

	ir.stats.Failed++
	ir.stats.Processed++

	ir.stats.Errors = append(ir.stats.Errors, FileError{
		FilePath: filePath,
		Err:      err,
	})

	ir.printProgressMessage(p, "ERROR!\n")
}

// RegisterSuccess registers that file processing succeeded.
func (ir *ImageRepairerBase) RegisterSuccess(p SingleFileProcessor) {
	ir.stats.Processed++

	ir.printProgressMessage(p, "OK\n")
}

// DisplayStart outputs information that the file processing started.
func (ir *ImageRepairerBase) DisplayStart(p SingleFileProcessor, filePath string) {
	ir.printProgressMessage(p, "Processing file "+filePath+" .......................... ")
}

// ProcessAllFiles processes all files using the provided iterator and processor.
// It respects context cancellation (e.g., Ctrl+C) at both the iteration and processing levels.
func ProcessAllFiles(ctx context.Context, it filesystem.FilePathIterator, p SingleFileProcessor) {
	for path := range it.All(ctx) {
		// Checking if process interrupted by Ctrl+C
		if err := ctx.Err(); err != nil {
			p.RegisterError(p, "", fmt.Errorf("process interrupted: %w", err))
			return
		}

		p.DisplayStart(p, path)

		if err := p.ProcessSingleFile(ctx, path); err != nil {
			p.RegisterError(p, path, err)

			// if Ctrl+C pressed inside of ProcessSingleFile
			// stopping the processing loop
			if ctx.Err() != nil {
				return
			}
			continue
		}

		p.RegisterSuccess(p)
	}
}

// RunAndWaitForExit awaits for "Enter" key press or context cancellation.
func RunAndWaitForExit(ctx context.Context, in io.Reader, out io.Writer, dontWait bool) {
	if dontWait || ctx.Err() != nil {
		return
	}

	fmt.Fprintln(out, "Processing is complete. Press Enter to exit.")

	// Creating a channel to receive a signal when required key is pressed
	done := make(chan struct{})

	keyboardReader := bufio.NewReader(in)

	go func() {
		_, _ = keyboardReader.ReadString('\n')
		close(done)
	}()

	select {
	case <-ctx.Done():
		// If pressed Ctrl+C sequence
		// Please note that we are not closing the input stream here.
		// The goroutine above will leak, but since the application is about to exit,
		// the operating system will reclaim and free all resources immediately.
		return
	case <-done:
		// If pressed Enter key
		return
	}
}
