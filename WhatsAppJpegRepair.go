package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
)

func main() {
	const predefinedSourceFilesFolder string = "whatsapp-files"
	const predefinedDestinationFilesFolder string = "fixed-files"
	const consoleTextDelimiter = "---------------------------------------"

	fmt.Println("WhatsAppJpegRepair version 1.0 Copyright (c) 2021 by Rafael Osipov")
	fmt.Println("\nRepairs jpeg images saved from WhatsApp application to prevent errors upon opening these images in the Adobe Photoshop.")
	fmt.Println("Project web-site: https://github.com/cdefgah/whatsapp-jpeg-repair")
	fmt.Println(consoleTextDelimiter)

	fmt.Println("Usage:")
	fmt.Println("\nWhen launched without parameters, application uses predefined folders.")
	fmt.Println("For broken jpeg files the application uses internal folder: ", predefinedSourceFilesFolder)
	fmt.Println("Fixed files will be stored in the internal folder: ", predefinedDestinationFilesFolder)

	fmt.Println("\nAlso it is possible to specify custom source and destination folders. Use the following approach:")
	fmt.Println("\n\tWhatsAppJpegRepair <problematic-files-path> <processed-files-path>")
	fmt.Println("\nExamples (for Windows, and MacOS respectively):")
	fmt.Println("\n\tWhatsAppJpegRepair c:/temp/problem-files/ c:/temp/fixed-jpeg-files/")
	fmt.Println("\n\tWhatsAppJpegRepair /home/username/Documents/broken-files /home/username/Documents/correct-files")
	fmt.Println(consoleTextDelimiter)

	const paramsAmountForCustomFolders int = 3
	const paramsAmountForPredefinedFolders int = 1
	const codeRepairProcessSucceed int = 0
	const codeRepairProcessFailed int = -1

	var sourceFilesPath string
	var destinationFilesPath string

	switch args := os.Args; len(args) {
	case paramsAmountForCustomFolders:
		sourceFilesPath = args[1]
		destinationFilesPath = args[2]

	case paramsAmountForPredefinedFolders:
		currentWorkingFolder, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			log.Fatal(err)
		}

		sourceFilesPath = filepath.Join(currentWorkingFolder, predefinedSourceFilesFolder)
		destinationFilesPath = filepath.Join(currentWorkingFolder, predefinedDestinationFilesFolder)

	default:
		fmt.Println("Incorrect number of parameters. Either run this application without parameters or pass source and destination folder as described above.")
		os.Exit(codeRepairProcessFailed)
	}

	var repairCompletedSuccessfully bool = repairImageFiles(sourceFilesPath, destinationFilesPath)
	fmt.Println("\nPress Enter to close the application")
	fmt.Scanln()

	if repairCompletedSuccessfully {
		os.Exit(codeRepairProcessSucceed)
	} else {
		os.Exit(codeRepairProcessFailed)
	}
}

// Repairs broken jpeg image files
// Gets location of broken files in sourceFolderPath variable
// and the location of folder, where fixed files will be stored, in the destinationFolderPath variable
// Returns true, if there were no errors upon files processing, false otherwise.
func repairImageFiles(sourceFolderPath string, destinationFolderPath string) bool {
	var totalFilesCount int32
	var processedFilesCount int32

	fmt.Println("Source folder path: ", sourceFolderPath)
	fmt.Println("Destination folder path: ", destinationFolderPath)

	files, err := ioutil.ReadDir(sourceFolderPath)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		if !f.IsDir() {
			totalFilesCount++
			if processSingleImageFile(sourceFolderPath, f.Name(), destinationFolderPath) {
				processedFilesCount++
			}
		}
	}

	fmt.Println("\nTotal files count: ", totalFilesCount, "Processed files count: ", processedFilesCount)
	fmt.Println("\nDone!")

	return totalFilesCount == processedFilesCount
}

// Processes single jpeg image file.
// sourceFolderPath - path to folder, where broken image files are located.
// sourceFileNameWithExtension contains filename with extension of image file, that should be processed.
// destinationFilesFolderPath contains path to folder, where fixed image file will be stored.
// Returns true, if there were no errors upon file processing, false otherwise.
func processSingleImageFile(sourceFolderPath string, sourceFileNameWithExtension string, destinationFilesFolderPath string) bool {

	var sourceFilePath = filepath.Join(sourceFolderPath, sourceFileNameWithExtension)
	var sourceFileExtension = path.Ext(sourceFileNameWithExtension)
	var destinationFileWithExtension = sourceFileNameWithExtension[0:len(sourceFileNameWithExtension)-len(sourceFileExtension)] + ".jpg"
	var destinationFilePath = filepath.Join(destinationFilesFolderPath, destinationFileWithExtension)

	fmt.Print("Processing file: ", sourceFilePath, " ............... ")

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

	fmt.Println("OK!")

	return true
}

// Loads image from the file.
// filePath contains path to the image file.
// Returns object with saved image, or error, if something went wrong.
func getImageFromFilePath(filePath string) (image.Image, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	image, _, err := image.Decode(f)
	return image, err
}
