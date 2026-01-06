package repair

/*
SPDX-License-Identifier: GPL-3.0-only
Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>
*/

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/cdefgah/whatsapp-jpeg-repair/internal/config"
	"github.com/spf13/afero"
)

// Represents batch image repairer for direct mode.
type BatchImageRepairerForDirectMode struct {
	BatchImageRepairerBase
	options config.DirectModeOptions
}

// Performs batch repair (in direct mode) of all image files provided by iterator.
// func (bir *BatchImageRepairerForDirectMode) __repairAllFilesInDirectMode() {
// 	for _, singleFilePath := range bir.options.FilePaths {

// 		bir.logger.Info("Processing file ", singleFilePath, " ....... ")
// 		if err := bir.repairSingleImage(singleFilePath); err != nil {
// 			bir.stats.Failed++
// 			bir.stats.Errors = append(bir.stats.Errors, FileError{
// 				FilePath: singleFilePath,
// 				Error:    err,
// 			})

// 			bir.logger.Error("Processing file ", singleFilePath, " ....... ERROR!")
// 			// logging error and continue processing...
// 			continue
// 		}

// 		bir.logger.Info("Processing file ", singleFilePath, " ....... OK")
// 		bir.stats.Processed++
// 	}
// }

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

	err = afero.WriteFile(fs, backupFilePath, sourceData, DefaultFilePermissions)
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
func (bir *BatchImageRepairerForDirectMode) ProcessSingleFile(sourceFilePath string) error {

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
