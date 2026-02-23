// SPDX-License-Identifier: GPL-3.0-only
// Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>

package options

import (
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/pflag"
)

// Command line parameter keys
const (
	FlagSrcPath                  = "src-path"
	FlagSrcPathShort             = "s"
	FlagDestPath                 = "dest-path"
	FlagDestPathShort            = "d"
	FlagUseCurrentModTime        = "use-current-modification-time"
	FlagUseCurrentModTimeShort   = "t"
	FlagDeleteWhatsAppFiles      = "delete-whatsapp-files"
	FlagDeleteWhatsAppFilesShort = "w"
	FlagDontWaitToClose          = "dont-wait-to-close"
	FlagDontWaitToCloseShort     = "c"
	FlagPrcsNestedFolders        = "process-nested-folders"
	FlagPrcsNestedFoldersShort   = "n"
	FlagDisplayHelp              = "help"
	FlagDisplayHelpShort         = "h"
)

// predefined source and dest folder names
const (
	PredefinedSourceFilesFolder      = "whatsapp-files"
	PredefinedDestinationFilesFolder = "repaired-files"
)

// DirectModeOptions contains options for the direct processing mode.
// The application running in direct mode processes all passed files "in place",
// source file will be overwritten by result file.
type DirectModeOptions struct {
	FilePaths []string
}

// NewDirectOptions returns new instance of DirectModeOptions.
func NewDirectOptions(args []string) DirectModeOptions {
	return DirectModeOptions{
		FilePaths: args,
	}
}

// ManagedModeOptions contains options for the managed processing mode.
// The application running in managed processing mode reads the command line control
// parameters passed and acts accordingly.
type ManagedModeOptions struct {
	SourceFolderPath           string
	DestinationFolderPath      string
	UseCurrentModificationTime bool
	DeleteWhatsAppFiles        bool
	ProcessNestedFolders       bool
	DontWaitToClose            bool
}

// NewDefaultManagedModeOptions creates instance of ManagedModeOptions with default values.
func NewDefaultManagedModeOptions(currentWorkingFolder string) *ManagedModeOptions {
	return &ManagedModeOptions{
		SourceFolderPath:      filepath.Join(currentWorkingFolder, PredefinedSourceFilesFolder),
		DestinationFolderPath: filepath.Join(currentWorkingFolder, PredefinedDestinationFilesFolder),
	}
}

// String generates text representation of managed mode options.
func (mmo *ManagedModeOptions) String() string {
	var sb strings.Builder

	// Using fixed width to output flag names (for example, 30 symbols) to display aligned text block
	fmt.Fprintf(&sb, "%-30s %s\n", "Source folder path:", mmo.SourceFolderPath)
	fmt.Fprintf(&sb, "%-30s %s\n", "Destination folder path:", mmo.DestinationFolderPath)
	fmt.Fprintf(&sb, "%-30s %t\n", "Use current modification time:", mmo.UseCurrentModificationTime)
	fmt.Fprintf(&sb, "%-30s %t\n", "Delete WhatsApp files:", mmo.DeleteWhatsAppFiles)
	fmt.Fprintf(&sb, "%-30s %t\n", "Process nested folders:", mmo.ProcessNestedFolders)
	fmt.Fprintf(&sb, "%-30s %t\n", "Don't wait to close:", mmo.DontWaitToClose)

	return sb.String()
}

// NewManagedFlagSet creates new set of flags to process command line arguments.
func NewManagedFlagSet(writer io.Writer, managedOptions *ManagedModeOptions) (flagSet *pflag.FlagSet, displayHelp *bool) {

	flagSet = pflag.NewFlagSet("available command-line switches", pflag.ContinueOnError)
	flagSet.SetOutput(writer)
	flagSet.SortFlags = false

	flagSet.Usage = func() {
		fmt.Fprintln(writer, "Usage: ")
		fmt.Fprintln(writer, "\tWhatsAppJpegRepair [managed mode options]")
		fmt.Fprintln(writer, "\tWhatsAppJpegRepair <file>...")

		fmt.Fprintln(writer, "\nDescription:")
		fmt.Fprintln(writer, "\tThe application operates in one of two modes, depending on the arguments provided.")
		fmt.Fprintln(writer, "\n\tManaged mode is used when no arguments are provided or at least one managed option is specified. All managed options are optional and have default values.")
		fmt.Fprintln(writer, "\n\tDirect mode is used when only positional arguments are provided and no known managed options are present. In this mode, the positional arguments are treated as paths to files and processed in place.")

		fmt.Fprintln(writer, "\nA list of the available managed options is shown below.")
		fmt.Fprintln(writer)
		flagSet.PrintDefaults()
	}

	exampleHomeDocsFolder := "/home/yourusername/Documents"
	if runtime.GOOS == "windows" {
		exampleHomeDocsFolder = "c:\\Users\\YourUsername\\Documents"
	}

	sampleSourcePath := fmt.Sprintf("--%s=%s", FlagSrcPath, filepath.Join(exampleHomeDocsFolder, "brokenWhatsAppFiles"))
	sampleDestPath := fmt.Sprintf("--%s=%s", FlagDestPath, filepath.Join(exampleHomeDocsFolder, "repairedImageFiles"))

	flagSet.StringVarP(
		&managedOptions.SourceFolderPath,
		FlagSrcPath,
		FlagSrcPathShort,
		managedOptions.SourceFolderPath,
		fmt.Sprintf("Path to the folder containing the broken WhatsApp files.\nExample: %s.", sampleSourcePath),
	)

	flagSet.StringVarP(
		&managedOptions.DestinationFolderPath,
		FlagDestPath,
		FlagDestPathShort,
		managedOptions.DestinationFolderPath,
		fmt.Sprintf("This is the path to the folder where the repaired files will be stored.\n"+
			"If the folder does not exist, it will be created.\nExample: %s.", sampleDestPath),
	)

	flagSet.BoolVarP(
		&managedOptions.UseCurrentModificationTime,
		FlagUseCurrentModTime,
		FlagUseCurrentModTimeShort,
		managedOptions.UseCurrentModificationTime,
		"If this is true, the current time will be used to set the file's modification time. "+
			"The default setting is false, meaning that the repaired file will have "+
			"the same modification time as the source file.",
	)

	flagSet.BoolVarP(
		&managedOptions.DeleteWhatsAppFiles,
		FlagDeleteWhatsAppFiles,
		FlagDeleteWhatsAppFilesShort,
		managedOptions.DeleteWhatsAppFiles,
		fmt.Sprintf("If it is true, the successfully processed source WhatsApp files will be deleted. Default: %v.", managedOptions.DeleteWhatsAppFiles),
	)

	flagSet.BoolVarP(
		&managedOptions.ProcessNestedFolders,
		FlagPrcsNestedFolders,
		FlagPrcsNestedFoldersShort,
		managedOptions.ProcessNestedFolders,
		fmt.Sprintf("If it is true, then the application processes files in nested folders recursively. Default: %v.", managedOptions.ProcessNestedFolders),
	)

	flagSet.BoolVarP(
		&managedOptions.DontWaitToClose,
		FlagDontWaitToClose,
		FlagDontWaitToCloseShort,
		managedOptions.DontWaitToClose,
		fmt.Sprintf("If this is true, the application will exit immediately once processing is complete. Default: %v.", managedOptions.DontWaitToClose),
	)

	displayHelp = flagSet.BoolP(FlagDisplayHelp, FlagDisplayHelpShort, false, "Show this help message and exit.")

	return flagSet, displayHelp
}

// IsManagedMode returns true if managed mode selected.
// Function assumes that Parse() method was already called. Otherwise function will fail.
func IsManagedMode(argsWithoutAppName []string, fs *pflag.FlagSet) (bool, error) {
	if !fs.Parsed() {
		return false, fmt.Errorf("flags must be parsed before calling IsManagedMode")
	}

	managedModeFlagUsed := false

	fs.Visit(func(flag *pflag.Flag) {
		// Even the help shorthand passed, here will be the long flag name present
		if flag.Name == FlagDisplayHelp {
			return
		}
		managedModeFlagUsed = true
	})

	return len(argsWithoutAppName) == 0 || managedModeFlagUsed, nil
}
