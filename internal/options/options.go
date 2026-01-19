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
	UseCurrentModificationTime bool
	DeleteWhatsAppFiles        bool
	ProcessOnlyJpegFiles       bool
	ProcessNestedFolders       bool
	DontWaitToClose            bool
}

func IsManagedMode(allCliArguments []string) bool {
	argsWithoutAppName := allCliArguments[1:]

	if len(argsWithoutAppName) == 0 {
		return true
	}

	if noManagedModeFlagsPassed(argsWithoutAppName) {
		return false
	}

	return true
}

// Generates text representation of managed mode options.
//
// # Returns
//
// text representation of managed mode options
func (mmo ManagedModeOptions) ToString() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "Source folder path:            %s\n", mmo.SourceFolderPath)
	fmt.Fprintf(&sb, "Destination folder path:       %s\n", mmo.DestinationFolderPath)
	fmt.Fprintf(&sb, "Use current modification time: %t\n", mmo.UseCurrentModificationTime)
	fmt.Fprintf(&sb, "Delete WhatsApp files:         %t\n", mmo.DeleteWhatsAppFiles)
	fmt.Fprintf(&sb, "Process only JPEG files:       %t\n", mmo.ProcessOnlyJpegFiles)
	fmt.Fprintf(&sb, "Process nested folders:        %t\n", mmo.ProcessNestedFolders)
	fmt.Fprintf(&sb, "Dont wait to close:            %t\n", mmo.DontWaitToClose)

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
// Pointer to structure with managed mode options.
func createAndGetDefaultManagedModeOptions(currentWorkingFolder string) *ManagedModeOptions {
	const (
		predefinedSourceFilesFolder      = "whatsapp-files"
		predefinedDestinationFilesFolder = "repaired-files"
	)

	pathToRootFolderWithSourceFiles := filepath.Join(currentWorkingFolder, predefinedSourceFilesFolder)
	pathToRootDestinationFolder := filepath.Join(currentWorkingFolder, predefinedDestinationFilesFolder)

	return &ManagedModeOptions{
		SourceFolderPath:           pathToRootFolderWithSourceFiles,
		DestinationFolderPath:      pathToRootDestinationFolder,
		UseCurrentModificationTime: false,
		DeleteWhatsAppFiles:        false,
		ProcessNestedFolders:       false,
		ProcessOnlyJpegFiles:       false,
		DontWaitToClose:            false,
	}
}

// Parses all command line arguments and creates structure for managed processing mode options.
//
// # Parameters
//
// currentWorkingFolderPath - Full path to the current working folder.
// allCliArguments - All command line arguments, including the the first argument with application executable file.
// writer - Reference to writer to print all output.
//
// # Returns
//
// Pointer to structure with managed mode options if application arguments processed successfully.
// Error object on error.
func ParseManagedModeOptions(currentWorkingFolderPath string, allCliArguments []string, writer io.Writer) (*ManagedModeOptions, error) {
	options := createAndGetDefaultManagedModeOptions(currentWorkingFolderPath)

	commandLineFlags := pflag.NewFlagSet("available command-line switches", pflag.ContinueOnError)

	examplePath := func(subpath string) string {
		if runtime.GOOS == "windows" {
			return fmt.Sprintf("c:/Users/Username/Documents/%s", subpath)
		}
		return fmt.Sprintf("/home/username/Documents/%s", subpath)
	}

	sampleSourcePath := fmt.Sprintf("--%s=%s", sourceFilesPathParamKey, examplePath("brokenWhatsAppFiles"))
	sampleDestPath := fmt.Sprintf("--%s=%s", destinationFilesPathParamKey, examplePath("repairedImageFiles"))

	sourcePath := commandLineFlags.StringP(sourceFilesPathParamKey,
		sourceFilesPathShorthandParamKey,
		options.SourceFolderPath,
		fmt.Sprintf("Path to folder containing broken WhatsApp files.\nExample: %s", sampleSourcePath))

	destPath := commandLineFlags.StringP(destinationFilesPathParamKey,
		destinationFilesPathShorthandParamKey,
		options.DestinationFolderPath,
		fmt.Sprintf("This is the path to the folder where the repaired files will be stored. If the folder does not exist, it will be created.\nExample: %s", sampleDestPath))

	useCurrentModificationTime := commandLineFlags.BoolP(useCurrentModificationTimeParamKey,
		useCurrentModificationTimeShorthandParamKey,
		options.UseCurrentModificationTime,
		`If true, sets current time as file modification time. Default: source file's modification time.`)

	deleteWhatsAppFiles := commandLineFlags.BoolP(deleteWhatsAppFilesParamKey,
		deleteWhatsAppFilesShorthandParamKey,
		options.DeleteWhatsAppFiles,
		fmt.Sprintf("If true, processed WhatsApp files will be deleted. Default: %v", options.DeleteWhatsAppFiles))

	processNestedFolders := commandLineFlags.BoolP(processNestedSourceFoldersParamKey,
		processNestedSourceFoldersShorthandParamKey,
		options.ProcessNestedFolders,
		fmt.Sprintf("If true, processes files in nested folders recursively. Default: %v", options.ProcessNestedFolders))

	processOnlyJpegFiles := commandLineFlags.BoolP(processOnlyJpegFilesParamKey,
		processOnlyJpegFilesShorthandParamKey,
		options.ProcessOnlyJpegFiles,
		fmt.Sprintf("If true, processes only jpeg files (with 'jpg', 'jpeg', 'jpe', 'jif', 'jfif' and 'jfi' extensions, case insensitive). Default: %v", options.ProcessOnlyJpegFiles))

	dontWaitToClose := commandLineFlags.BoolP(dontWaitToCloseParamKey,
		dontWaitToCloseShorthandParamKey,
		options.DontWaitToClose,
		fmt.Sprintf("If true, the application exits immediately. Default: %v", options.DontWaitToClose))

	argsWithoutAppName := allCliArguments[1:]
	if err := commandLineFlags.Parse(argsWithoutAppName); err != nil {
		return nil, err
	}

	commandLineFlags.SetOutput(writer)
	commandLineFlags.Usage()

	options.SourceFolderPath = filepath.Clean(*sourcePath)
	options.DestinationFolderPath = filepath.Clean(*destPath)
	options.UseCurrentModificationTime = *useCurrentModificationTime
	options.DeleteWhatsAppFiles = *deleteWhatsAppFiles
	options.ProcessNestedFolders = *processNestedFolders
	options.ProcessOnlyJpegFiles = *processOnlyJpegFiles
	options.DontWaitToClose = *dontWaitToClose

	fmt.Println(writer, "Actual parameters:")
	fmt.Println(writer, options.ToString())

	return options, nil
}

// Parses all command line arguments and creates structure for direct processing mode.
//
// # Parameters
//
// allCliArguments - All command line arguments, including the the first argument with application executable file.
//
// # Returns
//
// Pointer to structure with application options if application arguments processed successfully.
// Error object on error.
func ParseDirectModeOptions(allCliArguments []string) *DirectModeOptions {
	return &DirectModeOptions{
		FilePaths: allCliArguments[1:],
	}
}

func noManagedModeFlagsPassed(argsWithoutAppName []string) bool {
	fs := pflag.NewFlagSet("probe", pflag.ContinueOnError)
	fs.ParseErrorsAllowlist.UnknownFlags = true

	// just registering flags here to control its presence in args
	fs.StringP(sourceFilesPathParamKey, sourceFilesPathShorthandParamKey, "", "")
	fs.StringP(destinationFilesPathParamKey, destinationFilesPathShorthandParamKey, "", "")
	fs.BoolP(useCurrentModificationTimeParamKey, useCurrentModificationTimeShorthandParamKey, false, "")
	fs.BoolP(deleteWhatsAppFilesParamKey, deleteWhatsAppFilesShorthandParamKey, false, "")
	fs.BoolP(dontWaitToCloseParamKey, dontWaitToCloseShorthandParamKey, false, "")
	fs.BoolP(processNestedSourceFoldersParamKey, processNestedSourceFoldersShorthandParamKey, false, "")
	fs.BoolP(processOnlyJpegFilesParamKey, processOnlyJpegFilesShorthandParamKey, false, "")

	fs.SetOutput(nil) // suppressing usage output
	_ = fs.Parse(argsWithoutAppName)

	found := false
	fs.Visit(func(_ *pflag.Flag) {
		found = true
	})

	return !found
}
