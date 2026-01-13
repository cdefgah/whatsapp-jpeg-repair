package app

import (
	"log/slog"

	"github.com/cdefgah/whatsapp-jpeg-repair/internal/filesystem"
	"github.com/cdefgah/whatsapp-jpeg-repair/internal/options"
	"github.com/cdefgah/whatsapp-jpeg-repair/internal/repair"
	"github.com/spf13/afero"
)

/*
SPDX-License-Identifier: GPL-3.0-only
Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>
*/

func RunAppInDirectMode(fs afero.Fs, options options.DirectModeOptions, logger *slog.Logger) error {
	imageRepairer := repair.NewImageRepairerForDirectMode(fs, options, logger)
	filePathIterator := filesystem.NewFilePathsIteratorForDirectMode(options.FilePaths)

	repair.ProcessAllFiles(filePathIterator, imageRepairer)

	return nil
}

func RunAppInManagedMode(fs afero.Fs, options options.ManagedModeOptions, logger *slog.Logger) error {
	filePathIterator, err :=
		filesystem.NewFilePathsIteratorForManagedMode(fs,
			options.SourceFolderPath,
			options.ProcessNestedFolders,
			options.ProcessOnlyJpegFiles)

	if err != nil {
		return err
	}

	imageRepairer := repair.NewImageRepairerForManagedMode(fs, options, logger)

	repair.ProcessAllFiles(filePathIterator, imageRepairer)

	return nil
}

func LaunchApp(logger *slog.Logger) error {

	/**
	1. options have been previously parsed
	2. GetBatchImageRepairer (for Managed or for Direct mode)
	3. Get filepath iterator for selected batch image repairer (fsi)
	4. foreach => fsi => filepath
		ir.ProcessSingleFile(filepath)
	5. Print report
	6. If mode == ManagedMode
		waitForEnterIfRequired()

	*/

	return nil
}
