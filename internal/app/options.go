package app

/*
SPDX-License-Identifier: GPL-3.0-only
Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>
*/

import (
	"fmt"
	"strings"
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

// Generates text representation of managed mode options.
//
// # Returns
//
// text representation of managed mode options
func (mmo ManagedModeOptions) GetReport() string {
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
