/*
SPDX-License-Identifier: GPL-3.0-only
Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>
*/

package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
)

func main() {

	fmt.Println(`WhatsAppJpegRepair version 2.1.1 Copyright (c) 2021 by Rafael Osipov (rafael.osipov@outlook.com)
	Repairs jpeg images saved from WhatsApp application to prevent errors upon opening these images in the Adobe Photoshop.
	Project web-site, source code and documentation: https://github.com/cdefgah/whatsapp-jpeg-repair`)

	const sourceFilesPathParamKey string = "srcPath"
	const destinationFilesPathParamKey string = "destPath"
	const useCurrentModificationDateTimeParamKey = "useCurrentModificationDateTime"
	const deleteWhatsAppFilesParamKey = "deleteWhatsAppFiles"
	const dontWaitToCloseParamKey string = "dontWaitToClose"

	const predefinedSourceFilesFolder string = "whatsapp-files"
	const predefinedDestinationFilesFolder string = "repaired-files"

	const appExitCodeRepairProcessSucceed int = 0
	const appExitCodeRepairProcessFailed int = -1

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

	var repairCompletedSuccessfully bool = repairImageFiles(sourceFilesPath, destinationFilesPath, useCurrentModificationDateTime, deleteWhatsAppFiles)

	if !*dontWaitToClosePtr {
		fmt.Println("\nPress Enter to close the application")
		fmt.Scanln()
	}

	if repairCompletedSuccessfully {
		os.Exit(appExitCodeRepairProcessSucceed)
	} else {
		os.Exit(appExitCodeRepairProcessFailed)
	}
}

/*
Repairs broken jpeg image files
Gets location of broken files in sourceFolderPath variable
and the location of folder, where repaired files will be stored, in the destinationFolderPath variable
set useCurrentModificationDateTime as the relevant parameter value.
deleteWhatsAppFiles if true, we delete processed whatsapp files.
Returns true, if there were no errors upon files processing, false otherwise.
*/
func repairImageFiles(sourceFolderPath string, destinationFolderPath string, useCurrentModificationDateTime bool, deleteWhatsAppFiles bool) bool {
	var totalFilesCount int32
	var processedFilesCount int32

	files, err := ioutil.ReadDir(sourceFolderPath)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		if !f.IsDir() {
			totalFilesCount++
			if processSingleImageFile(sourceFolderPath, f.Name(), destinationFolderPath, useCurrentModificationDateTime, deleteWhatsAppFiles) {
				processedFilesCount++
			}
		}
	}

	fmt.Println("\nTotal files count: ", totalFilesCount, "Processed files count: ", processedFilesCount)

	fmt.Println("All repaired files are located in folder: ", destinationFolderPath)
	fmt.Println("\nDone!")

	return totalFilesCount == processedFilesCount
}

/*
Processes single jpeg image file.
sourceFolderPath - path to folder, where broken image files are located.
sourceFileNameWithExtension contains filename with extension of image file, that should be processed.
destinationFilesFolderPath contains path to folder, where repaired image file will be stored.
useCurrentModificationDateTime if true, setting current date/time as file modification time for generated file.
otherwise preserves source file modification date/time.
deleteWhatsAppFiles if true, we delete processed whatsapp files.
Returns true, if there were no errors upon file processing, false otherwise.
*/
func processSingleImageFile(sourceFolderPath string, sourceFileNameWithExtension string, destinationFilesFolderPath string, useCurrentModificationDateTime bool, deleteWhatsAppFiles bool) bool {

	var sourceFilePath = filepath.Join(sourceFolderPath, sourceFileNameWithExtension)
	var sourceFileExtension = path.Ext(sourceFileNameWithExtension)
	var destinationFileWithExtension = sourceFileNameWithExtension[0:len(sourceFileNameWithExtension)-len(sourceFileExtension)] + ".jpg"
	var destinationFilePath = filepath.Join(destinationFilesFolderPath, destinationFileWithExtension)

	fmt.Print("Processing file: ", sourceFilePath, " .......................... ")

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
