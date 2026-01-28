package repair

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"image"
	"image/gif"
	"image/jpeg"
	"image/png"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"

	"github.com/cdefgah/whatsapp-jpeg-repair/internal/filesystem"
	"github.com/spf13/afero"
)

const defaultFolderPermissions = 0755
const defaultFilePermissions = 0644

// FileError associates a specific file path with the error that occurred during its processing.
type FileError struct {
	FilePath string
	Err      error
}

// Error implements the error interface.
func (fe FileError) Error() string {
	return fmt.Sprintf("%s: %v", fe.FilePath, fe.Err)
}

// RepairStats holds the results of a batch image repair operation.
type RepairStats struct {
	Errors    []FileError
	Processed int
	Failed    int
}

// ImageRepairerBase provides a foundation for repairing images
// in both direct and managed modes.
type ImageRepairerBase struct {
	fs     afero.Fs
	stats  *RepairStats
	writer io.Writer
}

// SingleFileProcessor defines the contract for processing individual files
// and reporting the results of those operations.
type SingleFileProcessor interface {
	ProcessSingleFile(path string) error
	DisplayStart(path string)
	RegisterError(path string, err error)
	RegisterSuccess()
}

// readImage opens and decodes an image from the specified path.
func (ir *ImageRepairerBase) readImage(path string) (image.Image, string, error) {
	file, err := ir.fs.Open(path)
	if err != nil {
		return nil, "", fmt.Errorf("open image file: %w", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	img, format, err := image.Decode(reader)
	if err != nil {
		return nil, "", fmt.Errorf("decode image %s: %w", path, err)
	}

	return img, format, nil
}

// writeImage saves the image to the specified path using the provided format.
// It supports jpeg, png, gif, bmp, and tiff.
func (ir *ImageRepairerBase) writeImage(filePath string, img image.Image, format string) error {
	file, err := ir.fs.Create(filePath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	bw := bufio.NewWriter(file)

	var errEncode error
	switch format {
	case "jpeg":
		errEncode = jpeg.Encode(bw, img, nil)
	case "png":
		errEncode = png.Encode(bw, img)
	case "gif":
		errEncode = gif.Encode(bw, img, nil)
	case "bmp":
		errEncode = bmp.Encode(bw, img)
	case "tiff":
		errEncode = tiff.Encode(bw, img, nil)
	default:
		return fmt.Errorf("unsupported image format for encoding: %s", format)
	}

	if errEncode != nil {
		return fmt.Errorf("encode %s: %w", format, errEncode)
	}

	if err := bw.Flush(); err != nil {
		return fmt.Errorf("flush buffer: %w", err)
	}

	return nil
}

// HasErrors returns true if there's at least one error present in repair stats.
func (ir *ImageRepairerBase) HasErrors() bool {
	return len(ir.stats.Errors) > 0
}

// TextReport returns repair statistics as a formatted string report.
func (ir *ImageRepairerBase) TextReport() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "Processed: %d file(s)\n", ir.stats.Processed)

	if ir.HasErrors() {
		fmt.Fprintf(&sb, "Failed: %d file(s)\n", ir.stats.Failed)
		sb.WriteString("Errors:\n")

		for _, fe := range ir.stats.Errors {
			fmt.Fprintf(&sb, "  - %v\n", fe)
		}
	}

	return sb.String()
}

// RegisterError registers file processing error and outputs it to the writer.
func (ir *ImageRepairerBase) RegisterError(filePath string, err error) {
	ir.stats.Failed++
	ir.stats.Processed++ // Считаем как общую попытку обработки
	ir.stats.Errors = append(ir.stats.Errors, FileError{
		FilePath: filePath,
		Err:      err,
	})

	fmt.Fprintf(ir.writer, "ERROR!\n")
}

// RegisterSuccess registers that file processing succeeded.
func (ir *ImageRepairerBase) RegisterSuccess() {
	ir.stats.Processed++
	fmt.Fprintf(ir.writer, "OK\n")
}

// DisplayStart outputs information that the file processing started.
func (ir *ImageRepairerBase) DisplayStart(filePath string) {
	fmt.Fprintf(ir.writer, "Processing file %s .......................... ", filePath)
}

// ProcessAllFiles processes all files using the provided iterator and processor.
func ProcessAllFiles(iterator filesystem.FilePathIterator, p SingleFileProcessor) {
	// TODO add context processing to handle Ctrl+C upon file processing

	for {
		path, moreFiles := iterator.Next()
		if !moreFiles {
			break
		}

		p.DisplayStart(path)
		if err := p.ProcessSingleFile(path); err != nil {
			p.RegisterError(path, err)
			continue
		}
		p.RegisterSuccess()
	}
}

// RunAndWaitForExit awaits for "Enter" key press if dontWaitToClose is false.
func RunAndWaitForExit(dontWait bool, input io.Reader, output io.Writer) {
	if dontWait {
		return
	}

	fmt.Fprintln(output, "Press Enter to exit")
	_, _ = bufio.NewReader(input).ReadString('\n')
}
