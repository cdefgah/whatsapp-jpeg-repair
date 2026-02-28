# WhatsAppJpegRepair changelog

## Version 3.0.0 (March 1, 2026)

- Breaking changes: The application executable and command line parameter names have been renamed;
- A new option has been added to process nested folders;
- The app now only processes the following JPEG-related image file extensions: **.jpg**, **.jpeg**, **.jpe**, **.jif**, **.jfif**, and **.jfi**, which are case-insensitive. Files with different extensions will be ignored. This is because WhatsApp converts received images to JPEG format. Previous versions of the application processed all files indiscriminately. It converted non-JPEG files that did not require repair into JPEG format while retaining their original extensions;
- In managed mode, if a file with the same name and extension as the resulting file exists in the target folder, a backup copy of that file will be created;
- The wait at the end of the program will not be executed if the application is not running in interactive mode (e.g., in a Docker container) or if its output is redirected to a file;
- macOS users can now launch the application without any hassle with Gatekeeper. Please refer to the README instructions for more information.

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
