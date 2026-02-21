// SPDX-License-Identifier: GPL-3.0-only
// Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>

package repair

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/cdefgah/whatsapp-jpeg-repair/internal/filesystem"
	"github.com/cdefgah/whatsapp-jpeg-repair/internal/options"
	"github.com/spf13/afero"
)

// Represents image repairer for managed mode.
type ImageRepairerForManagedMode struct {
	ImageRepairerBase
	options options.ManagedModeOptions
}

// NewImageRepairerForManagedMode creates new instance of image repairer for managed mode.
func NewImageRepairerForManagedMode(fs afero.Fs, options options.ManagedModeOptions, stderr io.Writer) *ImageRepairerForManagedMode {
	return &ImageRepairerForManagedMode{
		ImageRepairerBase: ImageRepairerBase{
			fs:     fs,
			stats:  &Stats{},
			stderr: stderr,
		},
		options: options,
	}
}

// makeFolderIfMissing creates folder if it does not exist.
func (ir *ImageRepairerForManagedMode) makeFolderIfMissing(pathToFolder string) error {
	info, err := ir.fs.Stat(pathToFolder)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return ir.fs.MkdirAll(pathToFolder, filesystem.DefaultFolderPermissions)
		}

		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("path %q already exists and is not a directory", pathToFolder)
	}

	return nil
}

// ProcessSingleFile performs single image file repair.
func (ir *ImageRepairerForManagedMode) ProcessSingleFile(ctx context.Context, srcFilePath string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	destFilePath, err := ir.prepareDestFilePath(srcFilePath)
	if err != nil {
		return fmt.Errorf("error upon preparing destination file path: %w", err)
	}

	err = ir.createBackupIfFileExists(ctx, destFilePath)
	if err != nil {
		return fmt.Errorf("error upon backing up existing file: %w", err)
	}

	img, err := ir.readImage(ctx, srcFilePath)
	if err != nil {
		return err
	}

	err = ir.writeImage(ctx, destFilePath, img)
	if err != nil {
		return err
	}

	if !ir.options.UseCurrentModificationTime {
		if err := ir.setSrcFileModTimeToDestFile(srcFilePath, destFilePath); err != nil {
			return err
		}
	}

	// Checking if process interrupted by Ctrl+C
	if err := ctx.Err(); err != nil {
		return err
	}

	if ir.options.DeleteWhatsAppFiles {
		return ir.fs.Remove(srcFilePath)
	}

	return nil
}

// createBackupIfFileExists checks if file exists, creates its backup.
func (ir *ImageRepairerForManagedMode) createBackupIfFileExists(ctx context.Context, filePath string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	fileExists, err := afero.Exists(ir.fs, filePath)
	if err != nil {
		return err
	}

	if !fileExists {
		return nil
	}

	_, err = ir.createBackupFile(ctx, filePath)
	return err
}

// ensureDestFolderPath ensures that particular destination path exist, creates it when necessary.
func (ir *ImageRepairerForManagedMode) ensureDestFolderPath(srcFilePath string) (string, error) {

	srcBase := ir.options.SourceFolderPath
	srcDir := filepath.Dir(srcFilePath)

	relPath, err := filepath.Rel(srcBase, srcDir)
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(relPath, "..") {
		return "", fmt.Errorf("path %q is outside of source folder", srcFilePath)
	}

	dstBase := ir.options.DestinationFolderPath
	dstDir := filepath.Join(dstBase, relPath)

	if err := ir.makeFolderIfMissing(dstDir); err != nil {
		return "", err
	}

	return dstDir, nil
}

// setSrcFileModTimeToDestFile sets source file modification time to destingation file.
func (ir *ImageRepairerForManagedMode) setSrcFileModTimeToDestFile(srcFilePath, destFilePath string) error {
	stats, err := ir.fs.Stat(srcFilePath)
	if err != nil {
		return err
	}

	modTime := stats.ModTime()

	return ir.fs.Chtimes(destFilePath, modTime, modTime)
}

// prepareDestFilePath prepares destination folder to store the result file.
func (ir *ImageRepairerForManagedMode) prepareDestFilePath(srcFilePath string) (string, error) {
	srcFileName := filepath.Base(srcFilePath)
	destFolderPath, err := ir.ensureDestFolderPath(srcFilePath)
	if err != nil {
		return "", fmt.Errorf("ensure destination folder: %w", err)
	}

	return filepath.Join(destFolderPath, srcFileName), nil
}
