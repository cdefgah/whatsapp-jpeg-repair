// SPDX-License-Identifier: GPL-3.0-only
// Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/cdefgah/whatsapp-jpeg-repair/internal/app"
	"github.com/cdefgah/whatsapp-jpeg-repair/internal/repair"
	"github.com/spf13/afero"
	"github.com/spf13/pflag"
)

func main() {
	appOutput := os.Stderr
	stdin := os.Stdin

	fmt.Fprintln(appOutput, "WhatsAppJpegRepair version 3.0.0 Copyright (c) 2021 by Rafael Osipov (rafael.osipov@outlook.com)")
	fmt.Fprintln(appOutput, "The application repairs JPEG images saved from the WhatsApp app to prevent errors when opening them in Adobe Photoshop.")
	fmt.Fprintln(appOutput, "\nProject web-site, source code and documentation: https://github.com/cdefgah/whatsapp-jpeg-repair")
	fmt.Fprintln(appOutput)

	var clock repair.RealClock
	if err := runApp(stdin, appOutput, clock); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runApp(stdin io.Reader, stderr io.Writer, clock repair.Clock) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	exeFilePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable file path: %w", err)
	}

	exeFolderPath := filepath.Dir(exeFilePath)

	filesystem := afero.NewOsFs()
	argsWithoutAppName := os.Args[1:]

	appRunner := app.NewAppRunner(filesystem, stderr, clock)
	globalParams := app.NewGlobalProcessParams(stdin, exeFolderPath, argsWithoutAppName)

	err = appRunner.ProcessCommandLineArguments(ctx, *globalParams)
	if err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			return nil
		}

		if errors.Is(err, context.Canceled) {
			fmt.Fprintln(stderr, "\ninterrupted by user")
			return nil
		}

		return err
	}

	return nil
}
