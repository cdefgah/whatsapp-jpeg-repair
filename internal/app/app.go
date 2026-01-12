package app

import (
	"bufio"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/cdefgah/whatsapp-jpeg-repair/internal/filesystem"
	"github.com/spf13/afero"
)

/*
SPDX-License-Identifier: GPL-3.0-only
Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>
*/

const defaultFolderPermissions = 0755
const defaultFilePermissions = 0644

type AppRunner interface {
	RunAppInDirectMode(options DirectModeOptions, logger *slog.Logger) error
	RunAppInManagedMode(options ManagedModeOptions, logger *slog.Logger) error
}

// Declares a contract for file processor
// that works in managed and in direct mode.
type SingleFileProcessor interface {
	ProcessSingleFile(filePath string) error
	DisplayMessageOnFileProcessingStart(filePath string)
	RegisterFileProcessingError(filePath string, err error)
	RegisterFileProcessingSuccess(filePath string)
}

type FileError struct {
	FilePath string
	Error    error
}

// Stores statistics for batch image processing.
type RepairStats struct {
	Processed int
	Failed    int
	Errors    []FileError
}

// Represents batch image repairer base structure to
// process images in direct and in managed mode.
type ImageRepairerBase struct {
	fs     afero.Fs
	stats  *RepairStats
	logger *slog.Logger
}

// Represents batch image repairer for direct mode.
type ImageRepairerForDirectMode struct {
	ImageRepairerBase
	options DirectModeOptions
}

// Represents batch image repairer for managed mode.
type ImageRepairerForManagedMode struct {
	ImageRepairerBase
	options ManagedModeOptions
}

// Returns true if there's at least one error present in repair stats.
//
// # Returns
//
// true if there's at least one error present in repair stats.
func (bir *ImageRepairerBase) ErrorsPresent() bool {
	repairStats := bir.stats
	return len(repairStats.Errors) > 0
}

// Gets repair statistics as a text report.
//
// # Returns
//
// String with text report.
func (bir *ImageRepairerBase) ToString() string {
	actualStats := bir.stats
	var sb strings.Builder
	fmt.Fprintf(&sb, "Processed: %d file(s)\n", actualStats.Processed)
	if bir.ErrorsPresent() {
		fmt.Fprintf(&sb, "Failed: %d file(s)\n", actualStats.Failed)
		fmt.Fprintf(&sb, "Errors:\n")
		for _, fe := range actualStats.Errors {
			fmt.Fprintf(&sb, "\tFile path: %s, Error: %v\n", fe.FilePath, fe.Error)
		}
	}

	return sb.String()
}

// Loads image from the file.
//
// # Parameters
//
// fs - filesystem handler.
// filePath - path to the image file.
//
// # Returns
//
// object with loaded image or
// error if something went wrong.
func ReadImage(fs afero.Fs, filePath string) (image.Image, error) {
	file, err := fs.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	return img, err
}

// Writes repaired image to the file.
//
// # Parameters
//
// fs - filesystem handler.
// filePath - path to the image file.
// img - image obj to be saved.
//
// # Returns
//
// error - if something went wrong.
func WriteImage(fs afero.Fs, filePath string, img image.Image) error {
	file, err := fs.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return jpeg.Encode(file, img, nil)
}

// Creates new instance of batch image repairer for direct mode.
//
// # Parameters
//
// fs - filesystem reference.
// options - reference to the application runtime options for direct mode.
// logger - reference to actual logger.
//
// # Returns
//
// Reference to a new instance of batch image repairer for direct mode.
func NewImageRepairerForDirectMode(fs afero.Fs, options DirectModeOptions, logger *slog.Logger) *ImageRepairerForDirectMode {
	return &ImageRepairerForDirectMode{
		ImageRepairerBase: ImageRepairerBase{
			fs:     fs,
			stats:  &RepairStats{},
			logger: logger,
		},
		options: options,
	}
}

// Repairs single image mode in Direct mode.
//
// # Parameters
//
// fs - filesystem handler.
// sourceFilePath - path to the image file.
//
// # Returns
//
// error if something went wrong.
func (bir *ImageRepairerForDirectMode) ProcessSingleFile(sourceFilePath string) error {

	pathToBackupFile, err := createBackupFile(bir.fs, sourceFilePath)
	if err != nil {
		return err
	}

	img, err := ReadImage(bir.fs, sourceFilePath)
	if err != nil {
		return err
	}

	err = WriteImage(bir.fs, sourceFilePath, img)
	if err != nil {
		return err
	}

	err = deleteBackupFile(bir.fs, pathToBackupFile)
	if err != nil {
		return err
	}

	return nil
}

func (bir *ImageRepairerBase) RegisterFileProcessingError(filePath string, err error) {
	bir.stats.Failed++
	bir.stats.Errors = append(bir.stats.Errors, FileError{
		FilePath: filePath,
		Error:    err,
	})
	bir.logger.Error("Processing file ", filePath, " ....... ERROR!")
}

func (bir *ImageRepairerBase) RegisterFileProcessingSuccess(filePath string) {
	bir.stats.Processed++
	bir.logger.Info("Processing file ", filePath, " ....... OK")
}

func (bir *ImageRepairerBase) DisplayMessageOnFileProcessingStart(filePath string) {
	bir.logger.Info("Processing file ", filePath, " ....... ")
}

func processAllFiles(filePathIterator filesystem.FilePathIterator, singleFileProcessor SingleFileProcessor) {
	for {
		filePath := filePathIterator.NextFilePath()
		if filePath == "" {
			break // iterator returned empty string, no more files
		}

		singleFileProcessor.DisplayMessageOnFileProcessingStart(filePath)
		if err := singleFileProcessor.ProcessSingleFile(filePath); err != nil {
			singleFileProcessor.RegisterFileProcessingError(filePath, err)

			// continue processing...
			continue
		}

		singleFileProcessor.RegisterFileProcessingSuccess(filePath)
	}
}

// Creates backup for a file.
//
// # Parameters
//
// fs - filesystem handler.
// sourceFilePath - path to the image file.
//
// # Returns
//
// path to backup file or
// error if something went wrong.
func createBackupFile(fs afero.Fs, sourceFilePath string) (string, error) {
	var sourceFolderOnlyPath = filepath.Dir(sourceFilePath)
	var sourceFileNameWithExtension = filepath.Base(sourceFilePath)
	var sourceFileExtension = path.Ext(sourceFileNameWithExtension)
	var sourceFileNameOnly = strings.TrimSuffix(sourceFileNameWithExtension, sourceFileExtension)

	var backupFileNameWithExtension = sourceFileNameOnly + "_wjr_backup_file" + sourceFileExtension
	var backupFilePath = filepath.Join(sourceFolderOnlyPath, backupFileNameWithExtension)

	// Check if source file exists
	exists, err := afero.Exists(fs, sourceFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to check source file presence: %w", err)
	}
	if !exists {
		return "", fmt.Errorf("source file does not exist: %s", sourceFilePath)
	}

	// Copy source file to backup location
	sourceData, err := afero.ReadFile(fs, sourceFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read source file: %w", err)
	}

	err = afero.WriteFile(fs, backupFilePath, sourceData, defaultFilePermissions)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %w", err)
	}

	return backupFilePath, nil
}

// Deletes backup file.
//
// # Parameters
//
// fs - filesystem handler.
// sourceFilePath - path to the backup file.
//
// # Returns
//
// error if something went wrong.
func deleteBackupFile(fs afero.Fs, backupFilePath string) error {
	if backupFilePath == "" {
		return nil // Nothing to cleanup
	}

	// Check if backup file exists
	exists, err := afero.Exists(fs, backupFilePath)
	if err != nil {
		return fmt.Errorf("failed to check backup file existence: %w", err)
	}

	if !exists {
		return fmt.Errorf("unable to find backup file to delete it: %w", err)
	}

	// Remove the backup file
	err = fs.Remove(backupFilePath)
	if err != nil {
		return fmt.Errorf("failed to remove backup file: %w", err)
	}

	return nil
}

// Creates new instance of batch image repairer for managed mode.
//
// # Parameters
//
// fs - filesystem reference.
// options - reference to the application runtime options for managed mode.
// logger - reference to actual logger.
//
// # Returns
//
// Reference to a new instance of batch image repairer for managed mode.
func NewBatchImageRepairerForManagedMode(fs afero.Fs, options ManagedModeOptions, logger *slog.Logger) *ImageRepairerForManagedMode {
	return &ImageRepairerForManagedMode{
		ImageRepairerBase: ImageRepairerBase{
			fs:     fs,
			stats:  &RepairStats{},
			logger: logger,
		},
		options: options,
	}
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
func (bir *ImageRepairerForManagedMode) ProcessSingleFile(sourceFilePath string) error {
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

	// перенести содержимое finalize сюда
	return bir.finalizeManagedModeOperation(sourceFilePath, destinationFilePath)
}

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
func (bir *ImageRepairerForManagedMode) ensureParticularDestinationFolderPath(sourceFilePath string) (string, error) {

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
func (bir *ImageRepairerForManagedMode) createFolderIfItDoesNotExist(pathToFolder string) error {
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
		err = bir.fs.MkdirAll(pathToFolder, defaultFolderPermissions)
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
func (bir *ImageRepairerForManagedMode) setSourceFileModificationTimeToDestFile(sourceFilePath string, destinationFilePath string) error {
	sourceFileStats, err := bir.fs.Stat(sourceFilePath)
	if err != nil {
		return err
	}

	modTime := sourceFileStats.ModTime()

	return bir.fs.Chtimes(destinationFilePath, modTime, modTime)
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
func (bir *ImageRepairerForManagedMode) prepareDestinationFilePath(sourceFilePath string) (string, error) {
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
func (bir *ImageRepairerForManagedMode) finalizeManagedModeOperation(sourceFilePath string, destinationFilePath string) error {
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

func LaunchApp(logger *slog.Logger) error {

	/**
	1. options have been previously parsed
	2. GetBatchImageRepairer (for Managed or for Direct mode)
	3. Get filepath iterator for selected batch image repairer (fsi)
	4. foreach => fsi => filepath
		bir.ProcessSingleFile(filepath)
	5. Print report
	6. If mode == ManagedMode
		waitForEnterIfRequired()

	*/

	return nil
}
