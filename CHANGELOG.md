# Changelog

## [3.0.0] - 2026-03-01

### Added

- Option to process nested folders.
- Backup creation in managed mode if a file with the same name and extension already exists in the target folder.
- Seamless launch on macOS without Gatekeeper workarounds.

### Changed

- Renamed the application executable and command-line parameter names.
- The application now only processes specific JPEG-related extensions (`.jpg`, `.jpeg`, `.jpe`, `.jif`, `.jfif`, `.jfi`, case-insensitive). Non-JPEG files are now ignored.
- Disabled the wait prompt at the end of the program when running in non-interactive mode (e.g., Docker) or when output is redirected to a file.

## [2.2.1] - 2024-08-01

### Changed

- Refactored deprecated function calls.
- Updated supported Golang version.

## [2.2.0] - 2023-04-29

### Added

- Direct mode support to pass file image paths directly as space-delimited command-line arguments.

## [2.1.1] - 2021-11-02

### Added

- `runme.bat` (Windows) and `runme.sh` (macOS) helper scripts with default parameters for easier execution.

## [2.1.0] - 2021-05-26

### Added

- `-deleteWhatsAppFiles` option to delete source files after processing (defaults to false).

### Changed

- Renamed default folder from "fixed-files" to "repaired-files".

## [2.0.0] - 2021-02-12

### Added

- Command-line options to control source and destination paths.
- Option to control file modification time for fixed files.
- User interface control option.

## [1.0.0] - 2021-01-14

### Added

- Initial release of WhatsAppJpegRepair.
