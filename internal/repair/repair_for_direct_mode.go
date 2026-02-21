// SPDX-License-Identifier: GPL-3.0-only
// Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>

package repair

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

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
func NewImageRepairerForDirectMode(fs afero.Fs, opts options.DirectModeOptions, stderr io.Writer, clock Clock) *ImageRepairerForDirectMode {
	return &ImageRepairerForDirectMode{
		ImageRepairerBase: ImageRepairerBase{
			fs:     fs,
			stats:  &Stats{},
			stderr: stderr,
			clock:  clock,
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
