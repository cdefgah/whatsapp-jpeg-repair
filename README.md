# WhatsApp Jpeg Image Repair

[![](https://github.com/cdefgah/whatsapp-jpeg-repair/workflows/build/badge.svg)](https://github.com/cdefgah/whatsapp-jpeg-repair/actions)

When you send jpeg files via WhatsApp and afterwards try to open received jpeg files in the Adobe Photoshop, there is a chance to get the following error in Photoshop:

`Could not complete your request because a SOFn, DQT, or DHT JPEG marker is missing before a JPEG SOS marker`

For such cases users advised to open the broken file in MS Paint (or something similar on MacOS) and save it as jpeg file. Usually it helps, but when you have many broken image files, opening and saving every file is a kind of tedious work.

WhatsApp Jpeg Image Repair application solves this problem and can repair many broken files at once.

Follow these steps:
1. Download application archive. Navigate to [the application releases](https://github.com/cdefgah/whatsapp-jpeg-repair/releases). Then expand `Assets` section and download application zip archive relevant to your operating system.
2. Unzip application archive to an arbitrary folder. I recommend to use your home folder, for example `Documents` or `Desktop` folder, to prevent filesystem access control issues upon running the application. If you use MacOS, please note that MacOS Gatekeeper blocks this application when you try to launch it. To run downloaded application on MacOS your should either [disable MacOS Gatekeeper](https://www.google.com/search?q=disable+MacOS+Gatekeeper) or build the application from the source code as described hereinafter, and replace existing executable in downloaded archive with new executable file you have built from the source code.
3. Open application folder `WhatsAppJpegRepair`.
4. Place broken jpeg files from WhatsApp to the `whatsapp-files` directory, located in the application folder.
5. Run the application.
6. Go to the `fixed-files` folder to get fixed image files.

There are following option available:

`-srcPath` - contains path to the broken WhatsApp files
By default the application internal folder `whatsapp-files` is being used.
Currently this folder contains sample broken whatsapp jpeg images for demonstration purposes.

Example:
```
WhatsAppJpegRepair -srcPath=/home/Documents/Photos/WhatsAppFiles
```

this call will use `/home/Documents/Photos/WhatsAppFiles` folder as the source path to get broken whatsapp files.

`-destPath` - contains path to the folder, where fixed files will be stored.
By default the application internal folder `fixed-files` is being used.
If this folder does not exist, it will be created at runtime.

Example:
```
WhatsAppJpegRepair -srcPath=/home/Documents/Photos/WhatsAppFiles -destPath=/home/Documents/FixedPhotos
```
this call will use `/home/Documents/Photos/WhatsAppFiles` folder to look for broken whatsapp files, and will use `/home/Documents/FixedPhotos` folder to store fixed images.

`-dontWaitToClose` - if set to true, will close application just as it finished processing, otherwise it will wait until user presses 'Enter' key. By default its value is `false`.

Example:
```
WhatsAppJpegRepair -srcPath=/home/Documents/Photos/WhatsAppFiles -dontWaitToClose=true
```
this call will use folder `/home/Documents/Photos/WhatsAppFiles` as source file path, and application will be closes as it finished files processing. All fixed files will be stored to the default destination folder `fixed-files` (check `-destPath` option description above).

`-useCurrentModificationDateTime` - if it set to true, then created fixed files will get current date/time as file modification time. By default it is `false`, and all created fixed files get the same file modification date/time as source (broken) image files.

```
WhatsAppJpegRepair -useCurrentModificationDateTime=true
```
this call will use default source and destination folders (check `-srcPath` and `-destPath` options above), will wait until user presses Enter when the application completed the file processing,
and will set current date/time as file modification time for created fixed files.

There are no mandatory options provided. You can run the application without parameters, or set arbitrary set of parameters, for the rest of parameters default values will be applied.

## Building the application from the source

1. Download and install Go language support: https://golang.org/
2. In the application project folder run the following command:
```
go build WhatsAppJpegRepair.go
```
