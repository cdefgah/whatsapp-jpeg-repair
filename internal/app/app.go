package app

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/cdefgah/whatsapp-jpeg-repair/internal/filesystem"
	"github.com/cdefgah/whatsapp-jpeg-repair/internal/options"
	"github.com/cdefgah/whatsapp-jpeg-repair/internal/repair"
	"github.com/spf13/afero"
	"github.com/spf13/pflag"
)

/*
SPDX-License-Identifier: GPL-3.0-only
Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>
*/

// Describes single managed flag
type ManagedFlag struct {
	LongName  string
	ShortName string
	IsBool    bool
}

// List of all short and long keys for managed flags
var managedFlags = []ManagedFlag{
	{"source-files-path", "s", false},
	{"destination-files-path", "d", false},
	{"use-current-modification-time", "t", true},
	{"delete-whatsapp-files", "w", true},
	{"process-only-jpeg-files", "j", true},
	{"process-nested-folders", "n", true},
	{"dont-wait-to-close", "q", true},
}

func newManagedFlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("managed-detect", pflag.ContinueOnError)
	fs.SetInterspersed(true)

	for _, f := range managedFlags {
		if f.IsBool {
			fs.BoolP(f.LongName, f.ShortName, false, "")
		} else {
			fs.StringP(f.LongName, f.ShortName, "", "")
		}
	}

	return fs
}

func isManagedMode(args []string) (bool, error) {
	if len(args) == 0 {
		// Managed mode if no arguments provided
		return true, nil
	}

	fs := newManagedFlagSet()
	fs.SetOutput(nil) // suppressing usage output
	fs.ParseErrorsAllowlist.UnknownFlags = false

	err := fs.Parse(args)
	if err != nil {
		// Unknown key, raising error
		return false, err
	}

	// If at least one known key, then managed mode
	if fs.NFlag() > 0 {
		return true, nil
	}

	// Else - direct mode
	return false, nil
}

func RunAppInDirectMode(fs afero.Fs, options options.DirectModeOptions, writer io.Writer) error {
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

func RunAppInManagedMode(fs afero.Fs, options options.ManagedModeOptions, writer io.Writer) error {
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

func LaunchApp(logger *slog.Logger) error {

	/**
	1. options have been previously parsed
	2. GetBatchImageRepairer (for Managed or for Direct mode)
	3. Get filepath iterator for selected batch image repairer (fsi)
	4. foreach => fsi => filepath
		ir.ProcessSingleFile(filepath)
	5. Print report
	6. If mode == ManagedMode
		waitForEnterIfRequired()

	*/

	return nil
}
