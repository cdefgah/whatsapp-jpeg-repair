package options

/*
SPDX-License-Identifier: GPL-3.0-only
Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>
*/

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
	flagDisplayHelp      = "help"
	flagDisplayHelpShort = "h"
)

// DirectModeOptions contains options for the direct processing mode.
// The application running in direct mode processes all passed files "in place",
// source file will be overwritten by result file.
type DirectModeOptions struct {
	FilePaths []string
}

// ManagedModeOptions contains options for the managed processing mode.
// The application running in managed processing mode reads the command line control
// parameters passed and acts accordingly.
type ManagedModeOptions struct {
	SourceFolderPath           string
	DestinationFolderPath      string
	PreserveImageFormat        bool
	DontShowProgress           bool
	UseCurrentModificationTime bool
	DeleteWhatsAppFiles        bool
	ProcessOnlyJpegFiles       bool
	ProcessNestedFolders       bool
	DontWaitToClose            bool
}

// String generates text representation of managed mode options.
func (mmo *ManagedModeOptions) String() string {
	var sb strings.Builder

	// Using fixed width to output flag names (for example, 30 symbols) to display aligned text block
	fmt.Fprintf(&sb, "%-30s %s\n", "Source folder path:", mmo.SourceFolderPath)
	fmt.Fprintf(&sb, "%-30s %s\n", "Destination folder path:", mmo.DestinationFolderPath)
	fmt.Fprintf(&sb, "%-30s %t\n", "Preserve image format:", mmo.PreserveImageFormat)
	fmt.Fprintf(&sb, "%-30s %t\n", "Use current modification time:", mmo.UseCurrentModificationTime)
	fmt.Fprintf(&sb, "%-30s %t\n", "Delete WhatsApp files:", mmo.DeleteWhatsAppFiles)
	fmt.Fprintf(&sb, "%-30s %t\n", "Process only JPEG files:", mmo.ProcessOnlyJpegFiles)
	fmt.Fprintf(&sb, "%-30s %t\n", "Process nested folders:", mmo.ProcessNestedFolders)
	fmt.Fprintf(&sb, "%-30s %t\n", "Don't wait to close:", mmo.DontWaitToClose)

	return sb.String()
}

// NewDefaultManagedModeOptions creates instance of ManagedModeOptions with default values.
func NewDefaultManagedModeOptions(currentWorkingFolder string) *ManagedModeOptions {
	const (
		predefinedSourceFilesFolder      = "whatsapp-files"
		predefinedDestinationFilesFolder = "repaired-files"
	)

	return &ManagedModeOptions{
		SourceFolderPath:      filepath.Join(currentWorkingFolder, predefinedSourceFilesFolder),
		DestinationFolderPath: filepath.Join(currentWorkingFolder, predefinedDestinationFilesFolder),
		PreserveImageFormat:   true,
	}
}

// NewManagedFlagSet creates new set of flags to process command line arguments.
func NewManagedFlagSet(
	writer io.Writer,
	managedOptions *ManagedModeOptions,
) (flagSet *pflag.FlagSet, displayHelp *bool) {

	const (
		flagSrcPath                   = "src-path"
		flagSrcPathShort              = "s"
		flagDestPath                  = "dest-path"
		flagDestPathShort             = "d"
		flagUseCurrentModTime         = "use-current-modification-time"
		flagUseCurrentModTimeShort    = "t"
		flagDeleteWhatsAppFiles       = "delete-whatsapp-files"
		flagDeleteWhatsAppFilesShort  = "w"
		flagDontWaitToClose           = "dont-wait-to-close"
		flagDontWaitToCloseShort      = "c"
		flagPrcsNestedSrcFolders      = "process-nested-folders"
		flagPrcsNestedSrcFoldersShort = "n"
		flagPrcsOnlyJpegFiles         = "process-only-jpeg-files"
		flagPrcsOnlyJpegFilesShort    = "j"
		flagPresImageFormat           = "preserve-image-format"
		flagPresImageFormatShort      = "p"
		flagDontShowProgress          = "quiet"
		flagDontShowProgressShort     = "q"
	)

	flagSet = pflag.NewFlagSet("available command-line switches", pflag.ContinueOnError)
	flagSet.SetOutput(writer)
	flagSet.SortFlags = false

	flagSet.Usage = func() {
		fmt.Fprintln(writer, "Usage: ")
		fmt.Fprintln(writer, "\twhatsapp-jpeg-repair [managed mode options]")
		fmt.Fprintln(writer, "\twhatsapp-jpeg-repair <file>...")

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

	sampleSourcePath := fmt.Sprintf("--%s=%s", flagSrcPath, filepath.Join(exampleHomeDocsFolder, "brokenWhatsAppFiles"))
	sampleDestPath := fmt.Sprintf("--%s=%s", flagDestPath, filepath.Join(exampleHomeDocsFolder, "repairedImageFiles"))

	flagSet.StringVarP(
		&managedOptions.SourceFolderPath,
		flagSrcPath,
		flagSrcPathShort,
		managedOptions.SourceFolderPath,
		fmt.Sprintf("Path to the folder containing the broken WhatsApp files.\nExample: %s.", sampleSourcePath),
	)

	flagSet.StringVarP(
		&managedOptions.DestinationFolderPath,
		flagDestPath,
		flagDestPathShort,
		managedOptions.DestinationFolderPath,
		fmt.Sprintf("This is the path to the folder where the repaired files will be stored. If the folder does not exist, it will be created.\nExample: %s.", sampleDestPath),
	)

	flagSet.BoolVarP(
		&managedOptions.PreserveImageFormat,
		flagPresImageFormat,
		flagPresImageFormatShort,
		managedOptions.PreserveImageFormat,
		"If this is set to true, the application will attempt to preserve the format of the source image when writing the resulting file. "+
			"Otherwise, the file contents will be converted to JPEG format. "+
			"Supported formats: JPEG, PNG, GIF, BMP, TIFF. Default value is: false."+
			"If you need to process image files in an unsupported format, select 'false' for this option. However, the resulting file will contain a JPEG image.",
	)

	flagSet.BoolVarP(
		&managedOptions.DontShowProgress,
		flagDontShowProgress,
		flagDontShowProgressShort,
		managedOptions.DontShowProgress,
		"Setting this value to true will stop the program from displaying progress information while it is running. "+
			"The program will run in quiet mode, meaning that if errors occur, stderr will only contain error information. "+
			"This mode is useful if you want to check the error log after the programme has finished running. Default: false.",
	)

	flagSet.BoolVarP(
		&managedOptions.UseCurrentModificationTime,
		flagUseCurrentModTime,
		flagUseCurrentModTimeShort,
		managedOptions.UseCurrentModificationTime,
		"If this is true, the current time will be set as the file's modification time. The default is the modification time of the source file.",
	)

	flagSet.BoolVarP(
		&managedOptions.DeleteWhatsAppFiles,
		flagDeleteWhatsAppFiles,
		flagDeleteWhatsAppFilesShort,
		managedOptions.DeleteWhatsAppFiles,
		fmt.Sprintf("If it is true, the processed WhatsApp files will be deleted. Default: %v.", managedOptions.DeleteWhatsAppFiles),
	)

	flagSet.BoolVarP(
		&managedOptions.ProcessNestedFolders,
		flagPrcsNestedSrcFolders,
		flagPrcsNestedSrcFoldersShort,
		managedOptions.ProcessNestedFolders,
		fmt.Sprintf("If it is true, then the application processes files in nested folders recursively. Default: %v.", managedOptions.ProcessNestedFolders),
	)

	flagSet.BoolVarP(
		&managedOptions.ProcessOnlyJpegFiles,
		flagPrcsOnlyJpegFiles,
		flagPrcsOnlyJpegFilesShort,
		managedOptions.ProcessOnlyJpegFiles,
		fmt.Sprintf("If it is true, the application only processes JPEG files with the extensions 'jpg', 'jpeg', 'jpe', 'jif', 'jfif', or 'jfi' (case insensitive). Default: %v.", managedOptions.ProcessOnlyJpegFiles),
	)

	flagSet.BoolVarP(
		&managedOptions.DontWaitToClose,
		flagDontWaitToClose,
		flagDontWaitToCloseShort,
		managedOptions.DontWaitToClose,
		fmt.Sprintf("If this is true, the application will exit immediately once processing is complete. Default: %v.", managedOptions.DontWaitToClose),
	)

	displayHelp = flagSet.BoolP(
		flagDisplayHelp,
		flagDisplayHelpShort,
		false,
		"Show this help message and exit.",
	)

	return flagSet, displayHelp
}

// IsManagedMode returns true if managed mode selected.
// Function assumes that Parse() method was already called.
// Otherwise function won't work properly.
func IsManagedMode(argsWithoutAppName []string, fs *pflag.FlagSet) bool {
	managedModeFlagUsed := false

	fs.Visit(func(flag *pflag.Flag) {
		// Even the help shorthand passed, here will be the long flag name present
		if flag.Name == flagDisplayHelp {
			return
		}
		managedModeFlagUsed = true
	})

	return len(argsWithoutAppName) == 0 || managedModeFlagUsed
}

// NewDirectOptions returns new instance of DirectModeOptions.
func NewDirectOptions(args []string) DirectModeOptions {
	return DirectModeOptions{
		FilePaths: args,
	}
}
