package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/cdefgah/whatsapp-jpeg-repair/internal/app"
	"github.com/spf13/afero"
	"github.com/spf13/pflag"
)

/*
SPDX-License-Identifier: GPL-3.0-only
Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>
*/

func main() {
	fmt.Println("WhatsAppJpegRepair version 3.0.0 Copyright (c) 2021 by Rafael Osipov (rafael.osipov@outlook.com)")
	fmt.Println("The application repairs JPEG images saved from the WhatsApp app to prevent errors when opening them in Adobe Photoshop.")
	fmt.Println("\nProject web-site, source code and documentation: https://github.com/cdefgah/whatsapp-jpeg-repair")
	fmt.Println()

	// Creating context and listening to interrupt and terminate signals.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	// Calling deferred stop() to restore system calls behaviour handling
	defer stop()

	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get current working directory: %v\n", err)
		os.Exit(1)
	}

	filesystem := afero.NewOsFs()
	argumentsWithoutAppName := os.Args[1:]
	writer := os.Stdout

	err = app.ProcessCommandLineArguments(ctx, filesystem, currentWorkingDirectory, argumentsWithoutAppName, writer)
	if err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			return
		}

		// Checking if error is related to the termination request
		if errors.Is(err, context.Canceled) {
			fmt.Fprintln(os.Stderr, "\nExecution cancelled by user.")
			os.Exit(0)
		}

		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
