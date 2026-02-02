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

// Creates path to folder if it does not exist.
//
// # Parameters
//
// pathToFolder - path to be checked and in case path to folder does not exist, it will be created.
//
// # Returns
//
// error if something went wrong.
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

// Creates new instance of image repairer for managed mode.
//
// # Parameters
//
// fs - filesystem reference.
// options - reference to the application runtime options for managed mode.
// writer - reference to actual writer to print output.
//
// # Returns
//
// Reference to a new instance of batch image repairer for managed mode.
func NewImageRepairerForManagedMode(fs afero.Fs, options options.ManagedModeOptions, out io.Writer, errOut io.Writer) *ImageRepairerForManagedMode {
	return &ImageRepairerForManagedMode{
		ImageRepairerBase: ImageRepairerBase{
			fs:     fs,
			stats:  &RepairStats{},
			out:    out,
			errOut: errOut,
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
func (ir *ImageRepairerForManagedMode) ProcessSingleFile(ctx context.Context, sourceFilePath string) error {
	// Checking if process interrupted by Ctrl+C
	if err := ctx.Err(); err != nil {
		return err
	}

	destinationFilePath, err := ir.prepareDestinationFilePath(sourceFilePath)
	if err != nil {
		return fmt.Errorf("Error upon preparing destination file path: %w", err)
	}

	img, format, err := ir.readImage(ctx, sourceFilePath)
	if err != nil {
		return err
	}

	if !ir.options.PreserveImageFormat {
		const jpegFormatName = "jpeg"
		format = jpegFormatName
	}

	// Checking if process interrupted by Ctrl+C
	if err := ctx.Err(); err != nil {
		return err
	}

	err = ir.writeImage(ctx, destinationFilePath, img, format)
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
func (ir *ImageRepairerForManagedMode) setSourceFileModificationTimeToDestFile(sourceFilePath string, destinationFilePath string) error {
	sourceFileStats, err := ir.fs.Stat(sourceFilePath)
	if err != nil {
		return err
	}

	modTime := sourceFileStats.ModTime()

	return ir.fs.Chtimes(destinationFilePath, modTime, modTime)
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
func (ir *ImageRepairerForManagedMode) prepareDestinationFilePath(sourceFilePath string) (string, error) {
	sourceFileName := filepath.Base(sourceFilePath)
	destinationFolderPath, err := ir.ensureParticularDestinationFolderPath(sourceFilePath)
	if err != nil {
		return "", fmt.Errorf("Error upon ensuring particular destination folder path: %w", err)
	}

	return filepath.Join(destinationFolderPath, sourceFileName), nil
}
