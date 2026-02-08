// SPDX-License-Identifier: GPL-3.0-only
// Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>

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

// ProcessCommandLineArguments is entry point to the repair process, handles command line arguments and acts accordingly.
func ProcessCommandLineArguments(
	ctx context.Context,
	fs afero.Fs,
	exeFolderPath string,
	argsWithoutAppName []string,
	out io.Writer,
	errOut io.Writer,
) error {
	// Checking if process interrupted by Ctrl+C
	if err := ctx.Err(); err != nil {
		return err
	}

	managedOptions := options.NewDefaultManagedModeOptions(exeFolderPath)

	flagSet, displayHelp := options.NewManagedFlagSet(out, managedOptions)
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

		return runAppInManagedMode(ctx, fs, *managedOptions, out, errOut)
	}

	directOptions := options.NewDirectOptions(flagSet.Args())
	return runAppInDirectMode(ctx, fs, directOptions, out, errOut)
}

// runAppInDirectMode runs application in direct mode, repairs files whose paths were specified in the command-line parameters.
func runAppInDirectMode(ctx context.Context, fs afero.Fs, options options.DirectModeOptions, out io.Writer, errOut io.Writer) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	fmt.Fprintln(out, "Now the application runs in direct mode, processing file paths that are passed in the command line.")

	imageRepairer := repair.NewImageRepairerForDirectMode(fs, options, out, errOut)
	filePathIterator := filesystem.NewFilePathsIteratorForDirectMode(options.FilePaths)

	repair.ProcessAllFiles(ctx, filePathIterator, imageRepairer)
	fmt.Fprintln(out, imageRepairer.TextReport())

	if imageRepairer.HasErrors() {
		return fmt.Errorf("processing of image files in direct mode has failed!")
	} else {
		return nil
	}
}

// runAppInManagedMode runs application in managed mode, according to the parameters passed in the command line.
func runAppInManagedMode(ctx context.Context, fs afero.Fs, options options.ManagedModeOptions, out io.Writer, errOut io.Writer) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	fmt.Fprintln(out, "Now the application runs in managed mode with the following parameters:")
	fmt.Fprintln(out, options.String())

	filePathIterator, err :=
		filesystem.NewFilePathsIteratorForManagedMode(fs,
			options.SourceFolderPath,
			options.ProcessNestedFolders)

	if err != nil {
		return err
	}

	imageRepairer := repair.NewImageRepairerForManagedMode(fs, options, out, errOut)

	repair.ProcessAllFiles(ctx, filePathIterator, imageRepairer)
	fmt.Fprintln(out, imageRepairer.TextReport())

	if err := ctx.Err(); err != nil {
		return err
	}

	repair.RunAndWaitForExit(ctx, os.Stdin, os.Stdout, options.DontWaitToClose)

	if imageRepairer.HasErrors() {
		return fmt.Errorf("processing of image files in managed mode has failed!")
	} else {
		return nil
	}
}
