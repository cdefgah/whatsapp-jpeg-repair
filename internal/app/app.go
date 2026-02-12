// SPDX-License-Identifier: GPL-3.0-only
// Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>

package app

import (
	"context"
	"fmt"
	"io"
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

// CliProcessParams encapsulates processing command line params for the application.
type CliProcessParams struct {
	Stdin              io.Reader
	ExeFolderPath      string
	ArgsWithoutAppName []string
}

// NewGlobalProcessParams creates new instance of GlobalProcessParams structure.
func NewGlobalProcessParams(stdin io.Reader, exeFolderPath string, argsWithoutAppName []string) *CliProcessParams {
	return &CliProcessParams{
		Stdin:              stdin,
		ExeFolderPath:      exeFolderPath,
		ArgsWithoutAppName: argsWithoutAppName,
	}
}

// ProcessCommandLineArguments is entry point to the repair process, handles command line arguments and acts accordingly.
func (r *AppRunner) ProcessCommandLineArguments(ctx context.Context, params CliProcessParams) error {
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

		return r.runAppInManagedMode(ctx, params.Stdin, *managedOptions)
	}

	directOptions := options.NewDirectOptions(flagSet.Args())
	return r.runAppInDirectMode(ctx, directOptions)
}

// runAppInDirectMode runs application in direct mode, repairs files whose paths were specified in the command-line parameters.
func (r *AppRunner) runAppInDirectMode(ctx context.Context, options options.DirectModeOptions) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	fmt.Fprintln(r.stderr, "Now the application runs in direct mode, processing file paths that are passed in the command line.")

	imageRepairer := repair.NewImageRepairerForDirectMode(r.fs, options, r.stderr)
	filePathIterator := filesystem.NewFilePathsIteratorForDirectMode(options.FilePaths)

	repair.ProcessAllFiles(ctx, filePathIterator, imageRepairer)
	fmt.Fprintln(r.stderr, imageRepairer.TextReport())

	if imageRepairer.HasErrors() {
		return fmt.Errorf("The processing of image files in direct mode has failed.")
	} else {
		return nil
	}
}

// runAppInManagedMode runs application in managed mode, according to the parameters passed in the command line.
func (r *AppRunner) runAppInManagedMode(ctx context.Context, stdin io.Reader, options options.ManagedModeOptions) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	fmt.Fprintln(r.stderr, "Now the application runs in managed mode with the following parameters:")
	fmt.Fprintln(r.stderr, options.String())

	filePathIterator, err :=
		filesystem.NewFilePathsIteratorForManagedMode(r.fs,
			options.SourceFolderPath,
			options.ProcessNestedFolders)

	if err != nil {
		return err
	}

	imageRepairer := repair.NewImageRepairerForManagedMode(r.fs, options, r.stderr)

	repair.ProcessAllFiles(ctx, filePathIterator, imageRepairer)
	fmt.Fprintln(r.stderr, imageRepairer.TextReport())

	repair.RunAndWaitForExit(ctx, stdin, r.stderr, options.DontWaitToClose)

	if imageRepairer.HasErrors() {
		return fmt.Errorf("The processing of image files in managed mode has failed.")
	} else {
		return nil
	}
}
