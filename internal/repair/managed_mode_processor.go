package repair

/*
SPDX-License-Identifier: GPL-3.0-only
Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>
*/

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/cdefgah/whatsapp-jpeg-repair/internal/app"
	"github.com/spf13/afero"
)

// Represents batch image repairer for managed mode.
type BatchImageRepairerForManagedMode struct {
	BatchImageRepairerBase
	options app.ManagedModeOptions
}

// Creates new instance of batch image repairer.
//
// # Parameters
//
// fs - filesystem reference.
// options - reference to the application runtime options for managed mode.
// logger - reference to actual logger.
//
// # Returns
//
// Reference to a new instance of batch image repairer.
func NewBatchImageRepairerForManagedMode(fs afero.Fs, options app.ManagedModeOptions, logger *slog.Logger) *BatchImageRepairerForManagedMode {
	return &BatchImageRepairerForManagedMode{
		BatchImageRepairerBase: BatchImageRepairerBase{
			fs:     fs,
			stats:  &RepairStats{},
			logger: logger,
		},
		options: options,
	}
}

// Performs batch repair (in managed mode) of all image files provided by iterator.
//
// # Parameters
//
// iterator - reference to FileSystemIterator instance.
// func (bir *BatchImageRepairer) RepairAllFilesInManagedMode(iterator *filesystem.FileSystemIterator) {
// 	for {
// 		filePath := iterator.Next()
// 		if filePath == "" {
// 			break // iterator returned empty string, no more files
// 		}

// 		bir.logger.Info("Processing file ", filePath, " ....... ")
// 		if err := bir.repairSingleFileInManagedMode(filePath); err != nil {
// 			bir.stats.Failed++
// 			bir.stats.Errors = append(bir.stats.Errors, FileError{
// 				FilePath: filePath,
// 				Error:    err,
// 			})

// 			// logging error and continue processing...
// 			bir.logger.Error("Processing file ", filePath, " ....... ERROR!")
// 			continue
// 		}

// 		bir.logger.Info("Processing file ", filePath, " ....... OK")
// 		bir.stats.Processed++
// 	}
// }

// Launches and awaits for "Enter" key if dontWaitToClose is false.
// Otherwise just completes its execution.
//
// # Parameters
//
// dontWaitToClose - if false, function awaits for "Enter" key press.
// input - I/O reader handler.
// output - I/O writer handler.
func RunAndWaitForExit(dontWaitToClose bool, input io.Reader, output io.Writer) {
	if !dontWaitToClose {
		const newLine = '\n'
		fmt.Fprintln(output, "Press Enter to exit")
		bufio.NewReader(input).ReadString(newLine)
	}
}

// Ensures that particular destination path exist.
//
// # Parameters
//
// sourceFilePath - path to the image file.
//
// # Returns
//
// path to destination folder for result file related to sourceFilePath or
// error if something went wrong.
func (bir *BatchImageRepairerForManagedMode) ensureParticularDestinationFolderPath(sourceFilePath string) (string, error) {

	initialSourceFolderPath := bir.options.SourceFolderPath
	processingSourceFolderPath := filepath.Dir(sourceFilePath)
	relativeSourceFolderPath, err := filepath.Rel(initialSourceFolderPath, processingSourceFolderPath)
	if err != nil {
		return "", err
	}

	initialDestFolderPath := bir.options.DestinationFolderPath
	processingDestFolderPath := filepath.Join(initialDestFolderPath, relativeSourceFolderPath)

	destinationFolderCreationError := bir.createFolderIfItDoesNotExist(processingDestFolderPath)
	if destinationFolderCreationError != nil {
		return "", destinationFolderCreationError
	}

	return processingDestFolderPath, nil
}

// Creates path to folder if it does not exist.
//
// # Parameters
//
// pathToFolder - path to be checked and in case path to folder does not exist, it will be created.
//
// # Returns
//
// error if something went wrong.
func (bir *BatchImageRepairerForManagedMode) createFolderIfItDoesNotExist(pathToFolder string) error {
	info, err := bir.fs.Stat(pathToFolder)
	if err == nil {
		// Path exists — check if it's a directory
		if !info.IsDir() {
			return fmt.Errorf("path %s exists and is not a directory", pathToFolder)
		}
		// It's already a directory, nothing to do
		return nil
	}

	if errors.Is(err, os.ErrNotExist) {
		// Safe to create directory
		err = bir.fs.MkdirAll(pathToFolder, DefaultFolderPermissions)
		if err != nil {
			return err
		}
		return nil
	}

	// Some other error from Stat
	return err
}

// Sets modification time for destination file equal to the modification time of the source file.
//
// # Parameters
//
// sourceFilePath - path to source file.
// destinationFilePath - path to destination file.
//
// # Returns
//
// error if something went wrong.
func (bir *BatchImageRepairerForManagedMode) setSourceFileModificationTimeToDestFile(sourceFilePath string, destinationFilePath string) error {
	sourceFileStats, err := bir.fs.Stat(sourceFilePath)
	if err != nil {
		return err
	}

	modTime := sourceFileStats.ModTime()

	return bir.fs.Chtimes(destinationFilePath, modTime, modTime)
}

// Performs single image file repair.
//
// # Parameters
//
// sourceFilePath - path to image file that needs to be repaired.
//
// # Returns
//
// error if something went wrong.
func (bir *BatchImageRepairerForManagedMode) ProcessSingleFile(sourceFilePath string) error {
	destinationFilePath, err := bir.prepareDestinationFilePath(sourceFilePath)
	if err != nil {
		return fmt.Errorf("Error upon preparing destination file path: %w", err)
	}

	img, err := ReadImage(bir.fs, sourceFilePath)
	if err != nil {
		return err
	}

	err = WriteImage(bir.fs, destinationFilePath, img)
	if err != nil {
		return err
	}

	return bir.finalizeManagedModeOperation(sourceFilePath, destinationFilePath)
}

// Prepares destination folder to store the result file.
//
// # Parameters
//
// sourceFilePath - path to image file that needs to be repaired.
//
// # Returns
//
// destination file path if all things are ok, or error if something went wrong.
func (bir *BatchImageRepairerForManagedMode) prepareDestinationFilePath(sourceFilePath string) (string, error) {
	sourceFileName := filepath.Base(sourceFilePath)
	destinationFolderPath, err := bir.ensureParticularDestinationFolderPath(sourceFilePath)
	if err != nil {
		return "", fmt.Errorf("Error upon ensuring particular destination folder path: %w", err)
	}

	return filepath.Join(destinationFolderPath, sourceFileName), nil
}

// Finalizes the repair operation, updates result file date/time when necessary,
// and removes source file when necessary.
//
// # Parameters
//
// sourceFilePath - path to image file that needs to be repaired.
// destinationFilePath - path, where the repaired file should be saved.
//
// # Returns
//
// error if something went wrong.
func (bir *BatchImageRepairerForManagedMode) finalizeManagedModeOperation(sourceFilePath string, destinationFilePath string) error {
	if !bir.options.UseCurrentModificationTime {
		if err := bir.setSourceFileModificationTimeToDestFile(sourceFilePath, destinationFilePath); err != nil {
			return err
		}
	}

	if bir.options.DeleteWhatsAppFiles {
		return bir.fs.Remove(sourceFilePath)
	}

	return nil
}
