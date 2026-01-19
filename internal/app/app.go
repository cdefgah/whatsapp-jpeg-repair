package app

import (
	"fmt"
	"io"
	"os"

	"github.com/cdefgah/whatsapp-jpeg-repair/internal/filesystem"
	"github.com/cdefgah/whatsapp-jpeg-repair/internal/options"
	"github.com/cdefgah/whatsapp-jpeg-repair/internal/repair"
	"github.com/spf13/afero"
)

/*
SPDX-License-Identifier: GPL-3.0-only
Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>
*/

func LaunchApp(fs afero.Fs, currentWorkingFolderPath string, allCliArguments []string, writer io.Writer) error {
	var err error = nil
	if options.IsManagedMode(allCliArguments) {
		managedModeOptions, err := options.ParseManagedModeOptions(currentWorkingFolderPath, allCliArguments, writer)
		if err != nil {
			return err
		}

		err = runAppInManagedMode(fs, *managedModeOptions, writer)

	} else {
		directModeOptions := options.ParseDirectModeOptions(allCliArguments)
		err = runAppInDirectMode(fs, *directModeOptions, writer)
	}

	return err
}

func runAppInDirectMode(fs afero.Fs, options options.DirectModeOptions, writer io.Writer) error {
	imageRepairer := repair.NewImageRepairerForDirectMode(fs, options, writer)
	filePathIterator := filesystem.NewFilePathsIteratorForDirectMode(options.FilePaths)

	repair.ProcessAllFiles(filePathIterator, imageRepairer)
	fmt.Println(writer, imageRepairer.GetTextReport())

	if imageRepairer.ErrorsPresent() {
		return fmt.Errorf("Image files processing in direct mode failed!")
	} else {
		return nil
	}
}

func runAppInManagedMode(fs afero.Fs, options options.ManagedModeOptions, writer io.Writer) error {
	filePathIterator, err :=
		filesystem.NewFilePathsIteratorForManagedMode(fs,
			options.SourceFolderPath,
			options.ProcessNestedFolders,
			options.ProcessOnlyJpegFiles)

	if err != nil {
		return err
	}

	imageRepairer := repair.NewImageRepairerForManagedMode(fs, options, writer)

	fmt.Println(writer, options.ToString())

	repair.ProcessAllFiles(filePathIterator, imageRepairer)
	fmt.Println(writer, imageRepairer.GetTextReport())

	repair.RunAndWaitForExit(options.DontWaitToClose, os.Stdin, os.Stdout)

	if imageRepairer.ErrorsPresent() {
		return fmt.Errorf("Image files processing in managed mode failed!")
	} else {
		return nil
	}
}
