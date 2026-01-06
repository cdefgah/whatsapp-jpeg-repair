package app

/*
SPDX-License-Identifier: GPL-3.0-only
Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>
*/

func RunApp(options AppRuntimeOptions) {
	/**
	1. options have been previously parsed
	2. GetBatchImageRepairer (for Managed or for Direct mode)
	3. Get filepath iterator for selected batch image repairer (fsi)
	4. foreach => fsi => filepath
		bir.ProcessSingleFile(filepath)
	5. Print report
	6. If mode == ManagedMode
		waitForEnterIfRequired()

	*/
}
