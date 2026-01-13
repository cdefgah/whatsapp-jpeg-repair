package app

import (
	"log/slog"

	"github.com/cdefgah/whatsapp-jpeg-repair/internal/options"
)

/*
SPDX-License-Identifier: GPL-3.0-only
Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>
*/

type AppRunner interface {
	RunAppInDirectMode(options options.DirectModeOptions, logger *slog.Logger) error
	RunAppInManagedMode(options options.ManagedModeOptions, logger *slog.Logger) error
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
