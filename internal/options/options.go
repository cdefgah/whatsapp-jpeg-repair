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
	sourceFilesPathParamKey                     = "src-path"
	sourceFilesPathShorthandParamKey            = "s"
	destinationFilesPathParamKey                = "dest-path"
	destinationFilesPathShorthandParamKey       = "d"
	useCurrentModificationTimeParamKey          = "use-current-modification-time"
	useCurrentModificationTimeShorthandParamKey = "t"
	deleteWhatsAppFilesParamKey                 = "delete-whatsapp-files"
	deleteWhatsAppFilesShorthandParamKey        = "w"
	dontWaitToCloseParamKey                     = "dont-wait-to-close"
	dontWaitToCloseShorthandParamKey            = "c"
	processNestedSourceFoldersParamKey          = "process-nested-folders"
	processNestedSourceFoldersShorthandParamKey = "n"
	processOnlyJpegFilesParamKey                = "process-only-jpeg-files"
	processOnlyJpegFilesShorthandParamKey       = "j"
	preserveImageFormatParamKey                 = "preserve-image-format"
	preserveImageFormatShorthandParamKey        = "p"
	displayHelpParamKey                         = "help"
	displayHelpShorthandParamKey                = "h"
)

// Contains options for the direct processing mode.
// The application running in direct mode processes all passed files "in place",
// source file will be overwritten by result file.
type DirectModeOptions struct {
	FilePaths []string
}

// Contains options for the managed processing mode.
// The application running in managed processing mode reads the command line control
// parameters passed and acts accordingly.
type ManagedModeOptions struct {
	SourceFolderPath           string
	DestinationFolderPath      string
	PreserveImageFormat        bool
	UseCurrentModificationTime bool
	DeleteWhatsAppFiles        bool
	ProcessOnlyJpegFiles       bool
	ProcessNestedFolders       bool
	DontWaitToClose            bool
}

// Generates text representation of managed mode options.
//
// # Returns
//
// text representation of managed mode options
func (mmo ManagedModeOptions) String() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "Source folder path:            %s\n", mmo.SourceFolderPath)
	fmt.Fprintf(&sb, "Destination folder path:       %s\n", mmo.DestinationFolderPath)
	fmt.Fprintf(&sb, "Preserve image format:	     %t\n", mmo.PreserveImageFormat)
	fmt.Fprintf(&sb, "Use current modification time: %t\n", mmo.UseCurrentModificationTime)
	fmt.Fprintf(&sb, "Delete WhatsApp files:         %t\n", mmo.DeleteWhatsAppFiles)
	fmt.Fprintf(&sb, "Process only JPEG files:       %t\n", mmo.ProcessOnlyJpegFiles)
	fmt.Fprintf(&sb, "Process nested folders:        %t\n", mmo.ProcessNestedFolders)
	fmt.Fprintf(&sb, "Don't wait to close:           %t\n", mmo.DontWaitToClose)

	return sb.String()
}

// Creates and populates structure for managed processing mode options with default values.
//
// # Parameters
//
// currentWorkingFolderPath - Full path to the current working folder.
//
// # Returns
//
// Structure with managed mode options.
func CreateAndGetDefaultManagedModeOptions(currentWorkingFolder string) ManagedModeOptions {
	const (
		predefinedSourceFilesFolder      = "whatsapp-files"
		predefinedDestinationFilesFolder = "repaired-files"
	)

	pathToRootFolderWithSourceFiles := filepath.Join(currentWorkingFolder, predefinedSourceFilesFolder)
	pathToRootDestinationFolder := filepath.Join(currentWorkingFolder, predefinedDestinationFilesFolder)

	return ManagedModeOptions{
		SourceFolderPath:           pathToRootFolderWithSourceFiles,
		DestinationFolderPath:      pathToRootDestinationFolder,
		PreserveImageFormat:        true,
		UseCurrentModificationTime: false,
		DeleteWhatsAppFiles:        false,
		ProcessNestedFolders:       false,
		ProcessOnlyJpegFiles:       false,
		DontWaitToClose:            false,
	}
}

func NewManagedFlagSet(
	writer io.Writer,
	managedOptions *ManagedModeOptions,
) (flagSet *pflag.FlagSet, displayHelp *bool) {

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

	examplePath := func(subpath string) string {
		if runtime.GOOS == "windows" {
			return fmt.Sprintf("c:/Users/YourUsername/Documents/%s", subpath)
		}
		return fmt.Sprintf("/home/yourusername/Documents/%s", subpath)
	}

	sampleSourcePath := fmt.Sprintf("--%s=%s", sourceFilesPathParamKey, examplePath("brokenWhatsAppFiles"))
	sampleDestPath := fmt.Sprintf("--%s=%s", destinationFilesPathParamKey, examplePath("repairedImageFiles"))

	flagSet.StringVarP(
		&managedOptions.SourceFolderPath,
		sourceFilesPathParamKey,
		sourceFilesPathShorthandParamKey,
		managedOptions.SourceFolderPath,
		fmt.Sprintf("Path to the folder containing the broken WhatsApp files.\nExample: %s.", sampleSourcePath),
	)

	flagSet.StringVarP(
		&managedOptions.DestinationFolderPath,
		destinationFilesPathParamKey,
		destinationFilesPathShorthandParamKey,
		managedOptions.DestinationFolderPath,
		fmt.Sprintf("This is the path to the folder where the repaired files will be stored. If the folder does not exist, it will be created.\nExample: %s.", sampleDestPath),
	)

	flagSet.BoolVarP(
		&managedOptions.PreserveImageFormat,
		preserveImageFormatParamKey,
		preserveImageFormatShorthandParamKey,
		managedOptions.PreserveImageFormat,
		"If this is set to true, the application will attempt to preserve the format of the source image when writing the resulting file. "+
			"Otherwise, the file contents will be converted to JPEG format. "+
			"Supported formats: JPEG, PNG, GIF, BMP, TIFF. "+
			"If you need to process image files in an unsupported format, select 'false' for this option. However, the resulting file will contain a JPEG image.",
	)

	flagSet.BoolVarP(
		&managedOptions.UseCurrentModificationTime,
		useCurrentModificationTimeParamKey,
		useCurrentModificationTimeShorthandParamKey,
		managedOptions.UseCurrentModificationTime,
		"If this is true, the current time will be set as the file's modification time. The default is the modification time of the source file.",
	)

	flagSet.BoolVarP(
		&managedOptions.DeleteWhatsAppFiles,
		deleteWhatsAppFilesParamKey,
		deleteWhatsAppFilesShorthandParamKey,
		managedOptions.DeleteWhatsAppFiles,
		fmt.Sprintf("If it is true, the processed WhatsApp files will be deleted. Default: %v.", managedOptions.DeleteWhatsAppFiles),
	)

	flagSet.BoolVarP(
		&managedOptions.ProcessNestedFolders,
		processNestedSourceFoldersParamKey,
		processNestedSourceFoldersShorthandParamKey,
		managedOptions.ProcessNestedFolders,
		fmt.Sprintf("If it is true, then the application processes files in nested folders recursively. Default: %v.", managedOptions.ProcessNestedFolders),
	)

	flagSet.BoolVarP(
		&managedOptions.ProcessOnlyJpegFiles,
		processOnlyJpegFilesParamKey,
		processOnlyJpegFilesShorthandParamKey,
		managedOptions.ProcessOnlyJpegFiles,
		fmt.Sprintf("If it is true, the application only processes JPEG files with the extensions 'jpg', 'jpeg', 'jpe', 'jif', 'jfif', or 'jfi' (case insensitive). Default: %v.", managedOptions.ProcessOnlyJpegFiles),
	)

	flagSet.BoolVarP(
		&managedOptions.DontWaitToClose,
		dontWaitToCloseParamKey,
		dontWaitToCloseShorthandParamKey,
		managedOptions.DontWaitToClose,
		fmt.Sprintf("If this is true, the application will exit immediately once processing is complete. Default: %v.", managedOptions.DontWaitToClose),
	)

	displayHelp = flagSet.BoolP(
		displayHelpParamKey,
		displayHelpShorthandParamKey,
		false,
		"Show this help message and exit.",
	)

	return flagSet, displayHelp
}

func IsManagedMode(argsWithoutAppName []string, fs *pflag.FlagSet) bool {
	managedModeFlagUsed := false

	fs.Visit(func(flag *pflag.Flag) {
		if flag.Name == displayHelpParamKey || flag.Name == displayHelpShorthandParamKey {
			return
		}
		managedModeFlagUsed = true
	})

	return len(argsWithoutAppName) == 0 || managedModeFlagUsed
}

func NewDirectOptions(args []string) DirectModeOptions {
	return DirectModeOptions{
		FilePaths: args,
	}
}
