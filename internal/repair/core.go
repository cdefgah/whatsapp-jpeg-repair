package repair

import (
	"bufio"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log/slog"
	"strings"

	"github.com/cdefgah/whatsapp-jpeg-repair/internal/filesystem"
	"github.com/spf13/afero"
)

const defaultFolderPermissions = 0755
const defaultFilePermissions = 0644

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

// Declares a contract for file processor
// that works in managed and in direct mode.
type SingleFileProcessor interface {
	ProcessSingleFile(filePath string) error
	DisplayMessageOnFileProcessingStart(filePath string)
	RegisterFileProcessingError(filePath string, err error)
	RegisterFileProcessingSuccess(filePath string)
}

// Loads image from the file.
//
// # Parameters
//
// filePath - path to the image file.
//
// # Returns
//
// object with loaded image or
// error if something went wrong.
func (ir *ImageRepairerBase) readImage(filePath string) (image.Image, error) {
	file, err := ir.fs.Open(filePath)
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
// filePath - path to the image file.
// img - image obj to be saved.
//
// # Returns
//
// error - if something went wrong.
func (ir *ImageRepairerBase) writeImage(filePath string, img image.Image) error {
	file, err := ir.fs.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return jpeg.Encode(file, img, nil)
}

// Returns true if there's at least one error present in repair stats.
//
// # Returns
//
// true if there's at least one error present in repair stats.
func (ir *ImageRepairerBase) ErrorsPresent() bool {
	return len(ir.stats.Errors) > 0
}

// Gets repair statistics as a text report.
//
// # Returns
//
// String with text report.
func (ir *ImageRepairerBase) GetTextReport() string {
	actualStats := ir.stats
	var sb strings.Builder
	fmt.Fprintf(&sb, "Processed: %d file(s)\n", actualStats.Processed)
	if ir.ErrorsPresent() {
		fmt.Fprintf(&sb, "Failed: %d file(s)\n", actualStats.Failed)
		fmt.Fprintf(&sb, "Errors:\n")
		for _, fe := range actualStats.Errors {
			fmt.Fprintf(&sb, "\tFile path: %s, Error: %v\n", fe.FilePath, fe.Error)
		}
	}

	return sb.String()
}

func (ir *ImageRepairerBase) RegisterFileProcessingError(filePath string, err error) {
	ir.stats.Failed++
	ir.stats.Errors = append(ir.stats.Errors, FileError{
		FilePath: filePath,
		Error:    err,
	})
	ir.logger.Error("Processing file ", filePath, " ....... ERROR!")
}

func (ir *ImageRepairerBase) RegisterFileProcessingSuccess(filePath string) {
	ir.stats.Processed++
	ir.logger.Info("Processing file ", filePath, " ....... OK")
}

func (ir *ImageRepairerBase) DisplayMessageOnFileProcessingStart(filePath string) {
	ir.logger.Info("Processing file ", filePath, " ....... ")
}

func ProcessAllFiles(filePathIterator filesystem.FilePathIterator, singleFileProcessor SingleFileProcessor) {
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
