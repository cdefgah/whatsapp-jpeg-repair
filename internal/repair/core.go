package repair

/*
SPDX-License-Identifier: GPL-3.0-only
Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>
*/

import (
	"fmt"
	"image"
	"image/jpeg"
	"log/slog"
	"strings"

	"github.com/spf13/afero"
)

const DefaultFolderPermissions = 0755
const DefaultFilePermissions = 0644

// Declares a contract for file processor
// that works in managed and in direct mode.
type SingleFileProcessor interface {
	ProcessSingleFile(filePath string)
}

// Stores information on file processing error.
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
type BatchImageRepairerBase struct {
	fs     afero.Fs
	stats  *RepairStats
	logger *slog.Logger
}

// Returns true if there's at least one error present in repair stats.
//
// # Returns
//
// true if there's at least one error present in repair stats.
func (bir *BatchImageRepairerBase) ErrorsPresent() bool {
	repairStats := bir.stats
	return len(repairStats.Errors) > 0
}

// Gets repair statistics as a text report.
//
// # Returns
//
// String with text report.
func (bir *BatchImageRepairerBase) GetReport() string {
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
