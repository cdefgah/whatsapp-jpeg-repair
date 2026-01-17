package repair

import (
	"fmt"
	"io"
	"path"
	"path/filepath"
	"strings"

	"github.com/cdefgah/whatsapp-jpeg-repair/internal/options"
	"github.com/spf13/afero"
)

// Represents image repairer for direct mode.
type ImageRepairerForDirectMode struct {
	ImageRepairerBase
	options options.DirectModeOptions
}

// Creates new instance of batch image repairer for direct mode.
//
// # Parameters
//
// fs - filesystem reference.
// options - reference to the application runtime options for direct mode.
// writer - reference to actual writer to print output.
//
// # Returns
//
// Reference to a new instance of batch image repairer for direct mode.
func NewImageRepairerForDirectMode(fs afero.Fs, options options.DirectModeOptions, writer io.Writer) *ImageRepairerForDirectMode {
	return &ImageRepairerForDirectMode{
		ImageRepairerBase: ImageRepairerBase{
			fs:     fs,
			stats:  &RepairStats{},
			writer: writer,
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
func (ir *ImageRepairerForDirectMode) ProcessSingleFile(sourceFilePath string) error {

	pathToBackupFile, err := ir.createBackupFile(sourceFilePath)
	if err != nil {
		return err
	}

	img, err := ir.readImage(sourceFilePath)
	if err != nil {
		return err
	}

	err = ir.writeImage(sourceFilePath, img)
	if err != nil {
		return err
	}

	err = ir.deleteBackupFile(pathToBackupFile)
	if err != nil {
		return err
	}

	return nil
}

// Creates backup for a file.
//
// # Parameters
//
// sourceFilePath - path to the image file.
//
// # Returns
//
// path to backup file or
// error if something went wrong.
func (ir *ImageRepairerForDirectMode) createBackupFile(sourceFilePath string) (string, error) {
	var sourceFolderOnlyPath = filepath.Dir(sourceFilePath)
	var sourceFileNameWithExtension = filepath.Base(sourceFilePath)
	var sourceFileExtension = path.Ext(sourceFileNameWithExtension)
	var sourceFileNameOnly = strings.TrimSuffix(sourceFileNameWithExtension, sourceFileExtension)

	var backupFileNameWithExtension = sourceFileNameOnly + "_wjr_backup_file" + sourceFileExtension
	var backupFilePath = filepath.Join(sourceFolderOnlyPath, backupFileNameWithExtension)

	// Check if source file exists
	exists, err := afero.Exists(ir.fs, sourceFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to check source file presence: %w", err)
	}
	if !exists {
		return "", fmt.Errorf("source file does not exist: %s", sourceFilePath)
	}

	// Copy source file to backup location
	sourceData, err := afero.ReadFile(ir.fs, sourceFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read source file: %w", err)
	}

	err = afero.WriteFile(ir.fs, backupFilePath, sourceData, defaultFilePermissions)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %w", err)
	}

	return backupFilePath, nil
}

// Deletes backup file.
//
// # Parameters
//
// sourceFilePath - path to the backup file.
//
// # Returns
//
// error if something went wrong.
func (ir *ImageRepairerForDirectMode) deleteBackupFile(backupFilePath string) error {
	if backupFilePath == "" {
		return nil // Nothing to cleanup
	}

	// Check if backup file exists
	exists, err := afero.Exists(ir.fs, backupFilePath)
	if err != nil {
		return fmt.Errorf("failed to check backup file existence: %w", err)
	}

	if !exists {
		return fmt.Errorf("unable to find backup file to delete it: %w", err)
	}

	// Remove the backup file
	err = ir.fs.Remove(backupFilePath)
	if err != nil {
		return fmt.Errorf("failed to remove backup file: %w", err)
	}

	return nil
}
