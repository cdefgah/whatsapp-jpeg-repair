// SPDX-License-Identifier: GPL-3.0-only
// Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>

package repair

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

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
			stats:  &RepairStats{},
			stderr: stderr,
		},
		options: options,
	}
}

// createFolderIfItDoesNotExist creates folder if it does not exist.
func (ir *ImageRepairerForManagedMode) createFolderIfItDoesNotExist(pathToFolder string) error {
	dirExists, err := afero.DirExists(ir.fs, pathToFolder)
	if err != nil {
		return err
	}

	if !dirExists {
		// Safe to create directory
		err = ir.fs.MkdirAll(pathToFolder, defaultFolderPermissions)
		if err != nil {
			return err
		}
		return nil
	}

	// Don't need to create folder, return no error
	return nil
}

// ProcessSingleFile performs single image file repair.
func (ir *ImageRepairerForManagedMode) ProcessSingleFile(ctx context.Context, sourceFilePath string) error {
	// Checking if process interrupted by Ctrl+C
	if err := ctx.Err(); err != nil {
		return err
	}

	destinationFilePath, err := ir.prepareDestinationFilePath(sourceFilePath)
	if err != nil {
		return fmt.Errorf("Error upon preparing destination file path: %w", err)
	}

	img, err := ir.readImage(ctx, sourceFilePath)
	if err != nil {
		return err
	}

	// Checking if process interrupted by Ctrl+C
	if err := ctx.Err(); err != nil {
		return err
	}

	err = ir.writeImage(ctx, destinationFilePath, img)
	if err != nil {
		return err
	}

	if !ir.options.UseCurrentModificationTime {
		if err := ir.setSourceFileModificationTimeToDestFile(sourceFilePath, destinationFilePath); err != nil {
			return err
		}
	}

	// Checking if process interrupted by Ctrl+C
	if err := ctx.Err(); err != nil {
		return err
	}

	if ir.options.DeleteWhatsAppFiles {
		return ir.fs.Remove(sourceFilePath)
	}

	return nil
}

// ensureParticularDestinationFolderPath ensures that particular destination path exist, creates it when necessary.
func (ir *ImageRepairerForManagedMode) ensureParticularDestinationFolderPath(sourceFilePath string) (string, error) {

	initialSourceFolderPath := ir.options.SourceFolderPath
	processingSourceFolderPath := filepath.Dir(sourceFilePath)
	relativeSourceFolderPath, err := filepath.Rel(initialSourceFolderPath, processingSourceFolderPath)
	if err != nil {
		return "", err
	}

	initialDestFolderPath := ir.options.DestinationFolderPath
	processingDestFolderPath := filepath.Join(initialDestFolderPath, relativeSourceFolderPath)

	destinationFolderCreationError := ir.createFolderIfItDoesNotExist(processingDestFolderPath)
	if destinationFolderCreationError != nil {
		return "", destinationFolderCreationError
	}

	return processingDestFolderPath, nil
}

// setSourceFileModificationTimeToDestFile sets source file modification time to destingation file.
func (ir *ImageRepairerForManagedMode) setSourceFileModificationTimeToDestFile(sourceFilePath string, destinationFilePath string) error {
	sourceFileStats, err := ir.fs.Stat(sourceFilePath)
	if err != nil {
		return err
	}

	modTime := sourceFileStats.ModTime()

	return ir.fs.Chtimes(destinationFilePath, modTime, modTime)
}

// prepareDestinationFilePath prepares destination folder to store the result file.
func (ir *ImageRepairerForManagedMode) prepareDestinationFilePath(sourceFilePath string) (string, error) {
	sourceFileName := filepath.Base(sourceFilePath)
	destinationFolderPath, err := ir.ensureParticularDestinationFolderPath(sourceFilePath)
	if err != nil {
		return "", fmt.Errorf("Error upon ensuring particular destination folder path: %w", err)
	}

	return filepath.Join(destinationFolderPath, sourceFileName), nil
}
