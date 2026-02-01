package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cdefgah/whatsapp-jpeg-repair/internal/filesystem"
	"github.com/cdefgah/whatsapp-jpeg-repair/internal/options"
	"github.com/cdefgah/whatsapp-jpeg-repair/internal/repair"
	"github.com/spf13/afero"
)

/*
SPDX-License-Identifier: GPL-3.0-only
Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>
*/

func ProcessCommandLineArguments(
	ctx context.Context,
	fs afero.Fs,
	cwd string,
	argsWithoutAppName []string,
	writer io.Writer,
) error {
	// Checking if process interrupted by Ctrl+C
	if err := ctx.Err(); err != nil {
		return err
	}

	managedOptions := options.NewDefaultManagedModeOptions(cwd)

	flagSet, displayHelp := options.NewManagedFlagSet(writer, managedOptions)
	if err := flagSet.Parse(argsWithoutAppName); err != nil || *displayHelp {
		flagSet.Usage()
		return nil
	}

	useManagedMode := options.IsManagedMode(argsWithoutAppName, flagSet)

	if useManagedMode {
		// check if non-managed mode arguments present for managed mode
		positionalArgsPresent := len(flagSet.Args()) > 0

		if positionalArgsPresent {
			flagSet.Usage()
			return nil
		}

		managedOptions.SourceFolderPath = filepath.Clean(managedOptions.SourceFolderPath)
		managedOptions.DestinationFolderPath = filepath.Clean(managedOptions.DestinationFolderPath)

		return runAppInManagedMode(ctx, fs, *managedOptions, writer)
	}

	directOptions := options.NewDirectOptions(flagSet.Args())
	return runAppInDirectMode(ctx, fs, directOptions, writer)
}

func runAppInDirectMode(ctx context.Context, fs afero.Fs, options options.DirectModeOptions, writer io.Writer) error {
	// Checking if process interrupted by Ctrl+C
	if err := ctx.Err(); err != nil {
		return err
	}

	fmt.Fprintln(writer, "Now the application runs in direct mode, processing file paths that are passed in the command line.")

	imageRepairer := repair.NewImageRepairerForDirectMode(fs, options, writer)
	filePathIterator := filesystem.NewFilePathsIteratorForDirectMode(options.FilePaths)

	repair.ProcessAllFiles(ctx, filePathIterator, imageRepairer)
	fmt.Fprintln(writer, imageRepairer.TextReport())

	if imageRepairer.HasErrors() {
		return fmt.Errorf("Image files processing in direct mode failed!")
	} else {
		return nil
	}
}

func runAppInManagedMode(ctx context.Context, fs afero.Fs, options options.ManagedModeOptions, writer io.Writer) error {
	// Checking if process interrupted by Ctrl+C
	if err := ctx.Err(); err != nil {
		return err
	}

	fmt.Fprintln(writer, "Now the application runs in managed mode with the following parameters:")
	fmt.Fprintln(writer, options.String())

	filePathIterator, err :=
		filesystem.NewFilePathsIteratorForManagedMode(fs,
			options.SourceFolderPath,
			options.ProcessNestedFolders,
			options.ProcessOnlyJpegFiles)

	if err != nil {
		return err
	}

	imageRepairer := repair.NewImageRepairerForManagedMode(fs, options, writer)

	repair.ProcessAllFiles(ctx, filePathIterator, imageRepairer)
	fmt.Fprintln(writer, imageRepairer.TextReport())

	repair.RunAndWaitForExit(ctx, options.DontWaitToClose, os.Stdin, os.Stdout)

	if imageRepairer.HasErrors() {
		return fmt.Errorf("Image files processing in managed mode failed!")
	} else {
		return nil
	}
}
