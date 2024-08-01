/*
SPDX-License-Identifier: GPL-3.0-only
Copyright (c) 2021 - 2024 by Rafael Osipov <rafael.osipov@outlook.com>
*/

package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {

	fmt.Println(`WhatsAppJpegRepair version 2.2.1 Copyright (c) 2021 - 2024 by Rafael Osipov (rafael.osipov@outlook.com)
	Repairs jpeg images saved from WhatsApp application to prevent errors upon opening these images in the Adobe Photoshop.
	Project web-site, source code and documentation: https://github.com/cdefgah/whatsapp-jpeg-repair`)

	const appExitCodeRepairProcessSucceed int = 0
	const appExitCodeRepairProcessFailed int = -1

	var repairCompletedSuccessfully bool

	if isDirectModeEnabled() {
		repairCompletedSuccessfully = executeInDirectMode()
	} else {
		repairCompletedSuccessfully = executeInManagedMode()
	}

	if repairCompletedSuccessfully {
		os.Exit(appExitCodeRepairProcessSucceed)
	} else {
		os.Exit(appExitCodeRepairProcessFailed)
	}
}

/*
Runs the application in managed mode, taking in account command line parameters as keys and values.
Returns true if there were no errors upon files processing, false otherwise.
*/
func executeInManagedMode() bool {
	const sourceFilesPathParamKey string = "srcPath"
	const destinationFilesPathParamKey string = "destPath"
	const useCurrentModificationDateTimeParamKey = "useCurrentModificationDateTime"
	const deleteWhatsAppFilesParamKey = "deleteWhatsAppFiles"
	const dontWaitToCloseParamKey string = "dontWaitToClose"

	const predefinedSourceFilesFolder string = "whatsapp-files"
	const predefinedDestinationFilesFolder string = "repaired-files"

	currentWorkingFolder, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	var sourceFilesPath = filepath.Join(currentWorkingFolder, predefinedSourceFilesFolder)
	var destinationFilesPath = filepath.Join(currentWorkingFolder, predefinedDestinationFilesFolder)

	var sampleSourceFilesPathDeclarationPrefix string = "-" + sourceFilesPathParamKey + "="
	var sampleDestinationFilesPathDeclarationPrefix string = "-" + destinationFilesPathParamKey + "="
	var sampleSourceFilesPathDeclaration string
	var sampleDestinationFilesPathDeclaration string
	if runtime.GOOS == "windows" {
		sampleSourceFilesPathDeclaration = sampleSourceFilesPathDeclarationPrefix + "c:/Users/Username/Documents/brokenWhatsAppFiles\n"
		sampleDestinationFilesPathDeclaration = sampleDestinationFilesPathDeclarationPrefix + "c:/Users/Username/Documents/repairedImageFiles\n"
	} else {
		sampleSourceFilesPathDeclaration = sampleSourceFilesPathDeclarationPrefix + "/home/username/Documents/brokenWhatsAppFiles\n"
		sampleDestinationFilesPathDeclaration = sampleDestinationFilesPathDeclarationPrefix + "/home/username/Documents/repairedImageFiles\n"
	}

	sourcePathPtr := flag.String(sourceFilesPathParamKey,
		sourceFilesPath, "Path to folder with broken whatsapp files. Declaration example: "+sampleSourceFilesPathDeclaration)
	destPathPtr := flag.String(destinationFilesPathParamKey,
		destinationFilesPath, `Path to folder where repaired files will be stored.
		If folder does not exists, it will be created.
		Declaration example: `+sampleDestinationFilesPathDeclaration)

	useCurrentModificationDateTimePtr := flag.Bool(useCurrentModificationDateTimeParamKey, false, `If set to true, sets current date time as file modification time.
	 By default uses source file modification date time.`)

	deleteWhatsAppFilesPtr := flag.Bool(deleteWhatsAppFilesParamKey, false, "If set to true, every processed whatsapp file will be deleted. Only repaired files remain. By default is false.")
	dontWaitToClosePtr := flag.Bool(dontWaitToCloseParamKey, false, "If set to true, does not wait a key press until exits the application. By default is false.")

	// parsing and loading app params
	flag.Parse()
	sourceFilesPath = *sourcePathPtr
	destinationFilesPath = *destPathPtr
	var useCurrentModificationDateTime = *useCurrentModificationDateTimePtr
	var deleteWhatsAppFiles = *deleteWhatsAppFilesPtr
	var dontWaitToCloseApp bool = *dontWaitToClosePtr

	fmt.Println("\n----------------------------------------------------------------------------------------")
	flag.Usage()

	// creating destination folder if it does not exist
	if _, err := os.Stat(*destPathPtr); os.IsNotExist(err) {
		os.Mkdir(*destPathPtr, os.ModeDir)
	}

	// displaying effective parameter values
	fmt.Println("\n----------------------------------------------------------------------------------------")
	fmt.Println("Source folder path:", sourceFilesPath)
	fmt.Println("Destination folder path:", destinationFilesPath)
	fmt.Println("Use current modification date time for repaired files:", useCurrentModificationDateTime)
	fmt.Println("Delete processed whatsapp files:", deleteWhatsAppFiles)
	fmt.Println("Don't wait app to close:", dontWaitToCloseApp)
	fmt.Println("----------------------------------------------------------------------------------------")

	var repairCompletedSuccessfully bool = repairImageFilesInManagedMode(sourceFilesPath, destinationFilesPath, useCurrentModificationDateTime, deleteWhatsAppFiles)

	if !*dontWaitToClosePtr {
		fmt.Println("\nPress Enter to close the application")
		fmt.Scanln()
	}

	return repairCompletedSuccessfully
}

/*
Repairs image files in direct mode.
Gets every image file path from the command line, one command line parameter recognized as one image file path.
Returns true, if there were no errors upon files processing, false otherwise.
*/
func executeInDirectMode() bool {
	var totalFilesCount int32
	var processedFilesCount int32

	fmt.Println("Processing files in direct mode: ")

	for _, arg := range os.Args[1:] {
		totalFilesCount++
		if repairSingleImageFileInDirectMode(arg) {
			processedFilesCount++
		}
	}

	fmt.Println("\nTotal files count: ", totalFilesCount, "Processed files count: ", processedFilesCount)
	fmt.Println("\nDone!")

	return totalFilesCount == processedFilesCount
}

/*
Processes a single image file in direct mode.
sourceFileFullPath contains full path to the image file that needs to be fixed.
returns true if there were no errors, false otherwise.
*/
func repairSingleImageFileInDirectMode(sourceFileFullPath string) bool {
	fmt.Println("Processing file: " + sourceFileFullPath)

	var sourceFolderOnlyPath = filepath.Dir(sourceFileFullPath)
	var sourceFileNameWithExtension = filepath.Base(sourceFileFullPath)
	var sourceFileExtension = path.Ext(sourceFileNameWithExtension)
	var sourceFileNameOnly = strings.TrimSuffix(sourceFileNameWithExtension, sourceFileExtension)

	var backupFileNameWithExtension = sourceFileNameOnly + "_wjr_backup_file" + sourceFileExtension
	var backupFileFullPath = filepath.Join(sourceFolderOnlyPath, backupFileNameWithExtension)

	fmt.Println("Reading image from file: " + sourceFileFullPath)
	image, err := getImageFromFilePath(sourceFileFullPath)
	if err != nil {
		fmt.Println(err)
		return false
	}

	fmt.Println("Creating backup file (will be deleted later): " + backupFileFullPath)
	os.Rename(sourceFileFullPath, backupFileFullPath)

	fmt.Println("Writing fixed file: " + sourceFileFullPath)
	f, err := os.Create(sourceFileFullPath)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer f.Close()
	jpeg.Encode(f, image, nil)

	defer os.Remove(backupFileFullPath)

	fmt.Println("------------------------------------")

	return true
}

/*
Repairs broken jpeg image files in managed mode.
Gets location of broken files in sourceFolderPath variable
and the location of folder, where repaired files will be stored, in the destinationFolderPath variable
set useCurrentModificationDateTime as the relevant parameter value.
deleteWhatsAppFiles if true, we delete processed whatsapp files.
Returns true, if there were no errors upon files processing, false otherwise.
*/
func repairImageFilesInManagedMode(sourceFolderPath string, destinationFolderPath string, useCurrentModificationDateTime bool, deleteWhatsAppFiles bool) bool {
	var totalFilesCount int32
	var processedFilesCount int32

	f, err := os.Open(sourceFolderPath)
	if err != nil {
		log.Fatal(err)
	}

	files, err := f.Readdir(0)
	if err != nil {
		log.Fatal(err)
	}

	for _, singleFileHandler := range files {
		if singleFileHandler.IsDir() {
			continue
		}

		totalFilesCount++
		if processSingleImageFileInManagedMode(sourceFolderPath, singleFileHandler.Name(), destinationFolderPath, useCurrentModificationDateTime, deleteWhatsAppFiles) {
			processedFilesCount++
		}
	}

	fmt.Println("\nTotal files count: ", totalFilesCount, "Processed files count: ", processedFilesCount)

	fmt.Println("All repaired files are located in folder: ", destinationFolderPath)
	fmt.Println("\nDone!")

	return totalFilesCount == processedFilesCount
}

/*
Processes single jpeg image file in managed mode.
sourceFolderPath - path to folder, where broken image files are located.
sourceFileNameWithExtension contains filename with extension of image file, that should be processed.
destinationFilesFolderPath contains path to folder, where repaired image file will be stored.
useCurrentModificationDateTime if true, setting current date/time as file modification time for generated file.
otherwise preserves source file modification date/time.
deleteWhatsAppFiles if true, we delete processed whatsapp files.
Returns true, if there were no errors upon file processing, false otherwise.
*/
func processSingleImageFileInManagedMode(sourceFolderPath string, sourceFileNameWithExtension string, destinationFilesFolderPath string, useCurrentModificationDateTime bool, deleteWhatsAppFiles bool) bool {

	var sourceFilePath = filepath.Join(sourceFolderPath, sourceFileNameWithExtension)
	var sourceFileExtension = path.Ext(sourceFileNameWithExtension)
	var destinationFileWithExtension = sourceFileNameWithExtension[0:len(sourceFileNameWithExtension)-len(sourceFileExtension)] + ".jpg"
	var destinationFilePath = filepath.Join(destinationFilesFolderPath, destinationFileWithExtension)

	fmt.Print("Processing file: ", sourceFilePath, " ........................... ")

	image, err := getImageFromFilePath(sourceFilePath)
	if err != nil {
		fmt.Println(err)
		return false
	}

	f, err := os.Create(destinationFilePath)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer f.Close()
	jpeg.Encode(f, image, nil)

	// setting modification time as in source file if necessary
	if !useCurrentModificationDateTime {
		sourceFileStats, err := os.Stat(sourceFilePath)
		if err != nil {
			log.Fatal(err)
		}

		var sourceFileModificationDateTime = sourceFileStats.ModTime()
		err = os.Chtimes(destinationFilePath, sourceFileModificationDateTime, sourceFileModificationDateTime)
		if err != nil {
			fmt.Println(err)
		}
	}

	fmt.Println("OK!")

	if deleteWhatsAppFiles {
		fmt.Print("Deleting whatsapp file: ", sourceFilePath, ".....................")
		err := os.Remove(sourceFilePath)
		if err != nil {
			fmt.Println("Unable to delete whatsapp file: ", sourceFilePath, err)
		} else {
			fmt.Println("OK!")
		}
	}

	return true
}

/*
Loads image from the file.
filePath contains path to the image file.
Returns object with saved image, or error, if something went wrong.
*/
func getImageFromFilePath(filePath string) (image.Image, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	image, _, err := image.Decode(f)
	return image, err
}

/*
Checks command line whether command line keys are present.
If at least one key present, it is not direct mode.
Direct mode means that only filenames (with paths) are provided as parameters.
*/
func isDirectModeEnabled() bool {
	for _, arg := range os.Args[1:] {
		if !strings.HasPrefix(arg, "-") {
			return true
		}
	}

	return false
}
