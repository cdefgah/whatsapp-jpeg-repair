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
	args := os.Args
	if len(args) != 3 {
		fmt.Println("WhatsAppImageFixer version 1.0 Copyright (c) 2021 by Rafael Osipov")
		fmt.Println("Fixes images saved from WhatsApp (tm) application to prevent errors upon opening these images in the Adobe Photoshop (tm).")
		fmt.Println("All trademarks belong to their respective owners.")
		fmt.Println("\nUsage:")
		fmt.Println("\tWhatsAppImageFixer <problematic-files-path> <processed-files-path>")
		fmt.Println("\nExample:")
		fmt.Println("\tWhatsAppImageFixer c:/temp/problem-files/ c:/temp/fixed-files/")
		return
	}

	var sourceFilesPath string
	var destinationFilesPath string
	var totalFilesCount int32
	var processedFilesCount int32

	sourceFilesPath = args[1]
	destinationFilesPath = args[2]

	fmt.Println("Source file path: " + sourceFilesPath)
	fmt.Println("Destination file path: " + destinationFilesPath)

	files, err := ioutil.ReadDir(sourceFilesPath)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		if !f.IsDir() {
			totalFilesCount++
			if processImageFile(sourceFilesPath, f.Name(), destinationFilesPath) {
				processedFilesCount++
			}
		}
	}

	fmt.Println("\nTotal files count: ", totalFilesCount, "Processed files count: ", processedFilesCount)

	if totalFilesCount == processedFilesCount {
		os.Exit(0) // OKAY
	} else {
		os.Exit(-1) // ERROR
	}
}

func processImageFile(sourceFolderPath string, sourceFileNameWithExtension string, destinationFilesFolderPath string) bool {

	var sourceFilePath = filepath.Join(sourceFolderPath, sourceFileNameWithExtension)
	var sourceFileExtension = path.Ext(sourceFileNameWithExtension)
	var destinationFileWithExtension = sourceFileNameWithExtension[0:len(sourceFileNameWithExtension)-len(sourceFileExtension)] + ".jpg"
	var destinationFilePath = filepath.Join(destinationFilesFolderPath, destinationFileWithExtension)

	fmt.Print("Processing file: " + sourceFilePath + " ............... ")

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

func getImageFromFilePath(filePath string) (image.Image, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	image, _, err := image.Decode(f)
	return image, err
}
