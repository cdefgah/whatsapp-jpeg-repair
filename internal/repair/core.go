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
	"path/filepath"
	"strings"
	"time"

	"image"
	"image/jpeg"

	"github.com/cdefgah/whatsapp-jpeg-repair/internal/filesystem"
	"github.com/spf13/afero"
	"golang.org/x/term"
)

// Clock interface is used to help to inject clock implementation for production and for testing environments.
type Clock interface {
	Now() time.Time
}

// RealClock is used in the actual operation of the application.
type RealClock struct{}

// Now returns current time.
func (RealClock) Now() time.Time { return time.Now() }

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
	clock  Clock
}

// SingleFileProcessor defines the contract for processing individual files
// and reporting the results of those operations.
type SingleFileProcessor interface {
	ProcessSingleFile(ctx context.Context, path string) error
	RegisterStart(path string)
	RegisterError(path string, err error)
	RegisterSuccess()
}

// createBackupFile creates a copy in the same directory as the source.
func (ir *ImageRepairerBase) createBackupFile(ctx context.Context, sourceFilePath string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	fileInfo, err := ir.fs.Stat(sourceFilePath)
	if err != nil {
		return "", err
	}

	if !fileInfo.Mode().IsRegular() {
		return "", fmt.Errorf("source file is not a regular file")
	}

	// Format constant: 2006(6) 01(1) 02(2) _ 15(3) 04(4) 05(5)
	const timeFormatLayout = "20060102_150405"

	dir := filepath.Dir(sourceFilePath)
	ext := filepath.Ext(sourceFilePath)
	nameOnly := strings.TrimSuffix(filepath.Base(sourceFilePath), ext)

	timestamp := ir.clock.Now().Format(timeFormatLayout)
	backupName := fmt.Sprintf("%s_%s_backup%s", nameOnly, timestamp, ext)
	backupPath := filepath.Join(dir, backupName)

	src, err := ir.fs.Open(sourceFilePath)
	if err != nil {
		return "", fmt.Errorf("open source file: %w", err)
	}
	defer src.Close()

	dst, err := ir.fs.OpenFile(backupPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, filesystem.DefaultFilePermissions)
	if err != nil {
		return "", fmt.Errorf("create backup file: %w", err)
	}

	// We use the defer close() call to close the file in the event of an error when calling io.Copy().
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("copy data to backup: %w", err)
	}

	// Calling the close() method explicitly to identify any issues when writing data to disk.
	if err := dst.Close(); err != nil {
		return "", fmt.Errorf("close backup file: %w", err)
	}

	return backupPath, nil
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
		return nil, fmt.Errorf("decode image %q: %w", path, err)
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

	options := jpeg.Options{
		Quality: 100, // lossless compression
	}

	bw := bufio.NewWriter(file)
	errEncode := jpeg.Encode(bw, img, &options)

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
	// nolint:gosec // G115: fd is a small integer for standard streams
	if !okIn || !term.IsTerminal(int(fIn.Fd())) {
		return false
	}

	fOut, okOut := out.(*os.File)
	// nolint:gosec // G115: fd is a small integer for standard streams
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
