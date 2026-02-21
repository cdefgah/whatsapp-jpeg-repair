# WhatsAppJpegRepair changelog

## Version 3.0.0 (TBA)

Breaking changes. Introduced POSIX-compliant command line options instead of old ones.
Please check README for more details.
Added new option to process nested folders.
Now app processes only JPEG-related image files (with extensions: ".jpg", ".jpeg", ".jpe", ".jif", ".jfif", ".jfi" case insensitive). Files with other extensions will be ignored. It is because WhatsApp converts received images to JPEG format.
Mac users can now launch the application hassle-free with Gatekeeper (please refer to the README instructions).
Если в managed-режиме в target-папке попадётся файл, который рискует быть перезаписанным, то будет создана его резервная копия.

## Version 2.2.1 (Aug 1, 2024)

- Refactored deprecated function calls, updated supported Golang version;

## Version 2.2.0 (Apr 29, 2023)

- Added support for direct mode, where the file image path can be passed directly to the application as a command line argument. Multiple paths can be passed as space-delimited arguments.

## Version 2.1.1 (Nov 2, 2021)

- Added `runme.bat` (for Windows users) and `runme.sh` (for MacOS users).
  Both files contain command to call to the application with some parameters, assigned to default values. It should be convenient for users not familiar with the terminal.
  Just edit `.sh` or `.bat` file, relevant to your operating system and set the parameter value according to your needs or add/remove necessary/unwanted parameters.

## Version 2.1.0 (May 26, 2021)

- Renamed default folder "fixed-files" to "repaired-files"
- Introduced new option "-deleteWhatsAppFiles", when it is set to true, the application deletes source whatsapp files upon processing and only repaired files remain. By default it is false.

## Version 2.0 (Feb 12, 2021)

- Breaking changes. Introduced new command line options to control source and destination path, file modification time for fixed files, and user interface control option.

## Version 1.0 (Jan 14, 2021)

- The first release of the application. Plain and basic :)
