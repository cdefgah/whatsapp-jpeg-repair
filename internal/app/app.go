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

// AppRunner encapsulates core params for the application.
type AppRunner struct {
	fs     afero.Fs
	stderr io.Writer
}

// NewAppRunner create new instance of AppRunner structure.
func NewAppRunner(fs afero.Fs, stderr io.Writer) *AppRunner {
	return &AppRunner{
		fs:     fs,
		stderr: stderr,
	}
}

// GlobalProcessParams encapsulates processing params for the application.
type GlobalProcessParams struct {
	ExeFolderPath      string
	ArgsWithoutAppName []string
}

// NewGlobalProcessParams creates new instance of GlobalProcessParams structure.
func NewGlobalProcessParams(exeFolderPath string, argsWithoutAppName []string) *GlobalProcessParams {
	return &GlobalProcessParams{
		ExeFolderPath:      exeFolderPath,
		ArgsWithoutAppName: argsWithoutAppName,
	}
}

// ProcessCommandLineArguments is entry point to the repair process, handles command line arguments and acts accordingly.
func (r *AppRunner) ProcessCommandLineArguments(
	ctx context.Context,
	params GlobalProcessParams,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	managedOptions := options.NewDefaultManagedModeOptions(params.ExeFolderPath)

	flagSet, displayHelp := options.NewManagedFlagSet(r.stderr, managedOptions)
	if err := flagSet.Parse(params.ArgsWithoutAppName); err != nil || *displayHelp {
		flagSet.Usage()
		return nil
	}

	useManagedMode := options.IsManagedMode(params.ArgsWithoutAppName, flagSet)

	if useManagedMode {
		// check if non-managed mode arguments present for managed mode
		positionalArgsPresent := len(flagSet.Args()) > 0

		if positionalArgsPresent {
			flagSet.Usage()
			return nil
		}

		managedOptions.SourceFolderPath = filepath.Clean(managedOptions.SourceFolderPath)
		managedOptions.DestinationFolderPath = filepath.Clean(managedOptions.DestinationFolderPath)

		return runAppInManagedMode(ctx, r.fs, *managedOptions, r.stderr)
	}

	directOptions := options.NewDirectOptions(flagSet.Args())
	return runAppInDirectMode(ctx, r.fs, directOptions, r.stderr)
}

// runAppInDirectMode runs application in direct mode, repairs files whose paths were specified in the command-line parameters.
func runAppInDirectMode(ctx context.Context, fs afero.Fs, options options.DirectModeOptions, stderr io.Writer) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	fmt.Fprintln(stderr, "Now the application runs in direct mode, processing file paths that are passed in the command line.")

	imageRepairer := repair.NewImageRepairerForDirectMode(fs, options, stderr)
	filePathIterator := filesystem.NewFilePathsIteratorForDirectMode(options.FilePaths)

	repair.ProcessAllFiles(ctx, filePathIterator, imageRepairer)
	fmt.Fprintln(stderr, imageRepairer.TextReport())

	if imageRepairer.HasErrors() {
		return fmt.Errorf("processing of image files in direct mode has failed!")
	} else {
		return nil
	}
}

// runAppInManagedMode runs application in managed mode, according to the parameters passed in the command line.
func runAppInManagedMode(ctx context.Context, fs afero.Fs, options options.ManagedModeOptions, stderr io.Writer) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	fmt.Fprintln(stderr, "Now the application runs in managed mode with the following parameters:")
	fmt.Fprintln(stderr, options.String())

	filePathIterator, err :=
		filesystem.NewFilePathsIteratorForManagedMode(fs,
			options.SourceFolderPath,
			options.ProcessNestedFolders)

	if err != nil {
		return err
	}

	imageRepairer := repair.NewImageRepairerForManagedMode(fs, options, stderr)

	repair.ProcessAllFiles(ctx, filePathIterator, imageRepairer)
	fmt.Fprintln(stderr, imageRepairer.TextReport())

	repair.RunAndWaitForExit(ctx, os.Stdin, stderr, options.DontWaitToClose)

	if imageRepairer.HasErrors() {
		return fmt.Errorf("processing of image files in managed mode has failed!")
	} else {
		return nil
	}
}
