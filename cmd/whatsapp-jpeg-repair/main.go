package main

import (
	"fmt"
	"os"

	"github.com/cdefgah/whatsapp-jpeg-repair/internal/app"
	"github.com/spf13/afero"
)

/*
SPDX-License-Identifier: GPL-3.0-only
Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>
*/

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	fs := afero.NewOsFs()
	args := os.Args
	if err := app.LaunchApp(fs, cwd, args, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
