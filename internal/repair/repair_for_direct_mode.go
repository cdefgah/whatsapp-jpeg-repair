// SPDX-License-Identifier: GPL-3.0-only
// Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>

package repair

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cdefgah/whatsapp-jpeg-repair/internal/options"
	"github.com/spf13/afero"
)

// ImageRepairerForDirectMode provides functionality to repair images specifically in direct mode.
type ImageRepairerForDirectMode struct {
	ImageRepairerBase
	options options.DirectModeOptions
}

// NewImageRepairerForDirectMode creates and initializes a new ImageRepairerForDirectMode.
// It sets up the base repairer with the provided filesystem, writer, and fresh statistics.
func NewImageRepairerForDirectMode(fs afero.Fs, opts options.DirectModeOptions, stderr io.Writer) *ImageRepairerForDirectMode {
	return &ImageRepairerForDirectMode{
		ImageRepairerBase: ImageRepairerBase{
			fs:     fs,
			stats:  &Stats{},
			stderr: stderr,
		},
		options: opts,
	}
}

// ProcessSingleFile repairs a single image in Direct mode.
// It creates a temporary backup, performs the repair, and removes the backup upon success.
func (ir *ImageRepairerForDirectMode) ProcessSingleFile(ctx context.Context, sourceFilePath string) error {
	// Checking if process interrupted by Ctrl+C
	if err := ctx.Err(); err != nil {
		return err
	}

	backupFilePath, err := ir.createBackupFile(ctx, sourceFilePath)
	if err != nil {
		return fmt.Errorf("create backup: %w", err)
	}

	img, err := ir.readImage(ctx, sourceFilePath)
	if err != nil {
		return fmt.Errorf("read image for repair: %w", err)
	}

	if err := ir.writeImage(ctx, sourceFilePath, img); err != nil {
		return fmt.Errorf("write repaired image: %w", err)
	}

	if err := ir.deleteBackupFile(ctx, backupFilePath); err != nil {
		return fmt.Errorf("remove backup after successful repair: %w", err)
	}

	return nil
}

// createBackupFile creates a copy in the same directory as the source.
// The backup file is expected to be cleaned up later by the caller or a cleanup function.
func (ir *ImageRepairerForDirectMode) createBackupFile(ctx context.Context, sourceFilePath string) (string, error) {
	// Checking if process interrupted by Ctrl+C
	if err := ctx.Err(); err != nil {
		return "", err
	}

	// Format constant: 2006(6) 01(1) 02(2) _ 15(3) 04(4) 05(5)
	const timeFormatLayout = "20060102_150405"

	dir := filepath.Dir(sourceFilePath)
	ext := filepath.Ext(sourceFilePath)
	nameOnly := strings.TrimSuffix(filepath.Base(sourceFilePath), ext)

	timestamp := time.Now().Format(timeFormatLayout)
	backupName := fmt.Sprintf("%s_%s_backup%s", nameOnly, timestamp, ext)
	backupPath := filepath.Join(dir, backupName)

	src, err := ir.fs.Open(sourceFilePath)
	if err != nil {
		return "", fmt.Errorf("open source file: %w", err)
	}
	defer src.Close()

	dst, err := ir.fs.OpenFile(backupPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, defaultFilePermissions)
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

// deleteBackupFile removes the backup file. It returns an error if the file
// does not exist or cannot be removed, as the backup's presence is expected.
func (ir *ImageRepairerForDirectMode) deleteBackupFile(ctx context.Context, path string) error {
	// Checking if process interrupted by Ctrl+C
	if err := ctx.Err(); err != nil {
		return err
	}

	if path == "" {
		return nil
	}

	err := ir.fs.Remove(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("backup file vanished unexpectedly: %s", path)
		}
		return fmt.Errorf("remove backup file: %w", err)
	}

	return nil
}
