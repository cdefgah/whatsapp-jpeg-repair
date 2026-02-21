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

type Env struct {
	fs          afero.Fs
	stderr      io.Writer
	stdin       io.Reader
	exeFilePath string
	args        []string
	clock       repair.Clock
}

func main() {
	appOutput := os.Stderr
	stdin := os.Stdin

	fmt.Fprintln(appOutput, "WhatsAppJpegRepair version 3.0.0 Copyright (c) 2021 by Rafael Osipov (rafael.osipov@outlook.com)")
	fmt.Fprintln(appOutput, "The application repairs JPEG images saved from the WhatsApp app to prevent errors when opening them in Adobe Photoshop.")
	fmt.Fprintln(appOutput, "\nProject web-site, source code and documentation: https://github.com/cdefgah/whatsapp-jpeg-repair")
	fmt.Fprintln(appOutput)

	exeFilePath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(appOutput, "failed to get executable file path: %v\n", err)
		os.Exit(1)
	}

	env := NewEnv(appOutput, stdin, exeFilePath, os.Args[1:])
	if err := env.runApp(); err != nil {
		fmt.Fprintf(appOutput, "error: %v\n", err)
		os.Exit(1)
	}
}

func NewEnv(stderr io.Writer, stdin io.Reader, exeFilePath string, args []string) *Env {
	return &Env{
		fs:          afero.NewOsFs(),
		stdin:       stdin,
		stderr:      stderr,
		exeFilePath: exeFilePath,
		args:        args,
		clock:       repair.RealClock{},
	}
}

func (env *Env) runApp() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	appRunner := app.NewAppRunner(env.fs, env.stderr, env.clock)

	exeFolderPath := filepath.Dir(env.exeFilePath)
	globalParams := app.NewGlobalProcessParams(env.stdin, exeFolderPath, env.args)

	err := appRunner.ProcessCommandLineArguments(ctx, *globalParams)
	if err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			return nil
		}

		if errors.Is(err, context.Canceled) {
			fmt.Fprintln(env.stderr, "\ninterrupted by user")
			return nil
		}

		return err
	}

	return nil
}
