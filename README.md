# WhatsApp Jpeg Image Repair

## version 2.2.1 (released on August 1, 2024)

[Changelog](./CHANGELOG.md)

When you are sent jpeg files via WhatsApp and then try to open received files in the Adobe Photoshop, there's a chance you'll get the following error in Photoshop:

`Could not complete your request because a SOFn, DQT, or DHT JPEG marker is missing before a JPEG SOS marker`

For this case users are usually advised to open the broken file in MS Paint (or something similar in MacOS) first and save it as jpeg file. Usually it helps, but when you have many broken image files, opening and saving them one by one may get a little tedious.

WhatsApp Jpeg Image Repair application solves this problem by repairing multiple broken files at once. 

## How to use the tool

Download application archive. Navigate to [the application releases](https://github.com/cdefgah/whatsapp-jpeg-repair/releases). Then expand `Assets` section and download application zip archive suitable for your operating system. Linux users are advised to download zip file, built for MacOS operating system.

Unzip application archive to the folder of your choosing, preferrably your home folder, e.g. `Documents` or `Desktop` folder, to prevent filesystem access control issues upon running the application.

Below, there are step-by-step instructions for different operating systems.

### Microsoft Windows users

For users' convenience I've added script file `runme.bat` file. It helps to use various options and switches to users who are not familiar with cmd-console/terminal window.

Just edit this file, change the relevant options and switches (listed in the `Options and switches` chapter below), add/remove options/switches you want, save the file and just run it via mouse doubleclick.

Please note, that `runme.bat` file is a plain text file. And you don't need a special application to edit this file. 

Just use a simple text editor of your choosing, like `Notepad++`, it can be downloaded here: https://notepad-plus-plus.org/downloads/

Install it, and then, just do the mouse right click on the `runme.bat` and select "Edit with Notepad++".

Next steps:

1. Place broken jpeg files from WhatsApp to the `whatsapp-files` directory, located in the application folder.
2. Run the application.
3. Go to the `repaired-files` folder to get repaired image files.

### Apple MacOS users

For users' convenience I've added script file `runme.sh` file. It helps to use various options and switches to users who are not familiar with terminal window. Available options are listed in the `Options and switches` chapter below.

To edit `runme.sh` file launch `TextEdit`, choose `File - Open` from menu and select `runme.sh` file: https://support.apple.com/guide/textedit/open-documents-txte51413d09/mac

Before using `runme.sh` open terminal in the application folder. Here is how: https://apple.stackexchange.com/questions/11323/how-can-i-open-a-terminal-window-directly-from-my-current-finder-location

And then type in the terminal the following command:

`chmod +x runme.sh`

And press Enter. After that just close the terminal window.

Please note that MacOS Gatekeeper blocks this application when you try to launch it. To run the downloaded application on MacOS your should either [disable MacOS Gatekeeper](https://www.google.com/search?q=disable+MacOS+Gatekeeper) or build the application from the source code as described below. And replace existing executable in downloaded zip archive with the newly-built file. Building from the source code is recommended way, because disabling Gatekeeper on different MacOS versions is not an easy task.

Source code file `WhatsAppJpegRepair.go` is included to the application archive for MacOS users. And in case you've decided to build the application from the source code, follow these steps:

1. Delete existing `WhatsAppJpegRepair` file, please don't confuse this file with `WhatsAppJpegRepair.go`.
2. Download and install Go language support from the [official web-site](https://golang.org/)
3. Open terminal in the unzipped application folder, and execute the following command:

```
go build WhatsAppJpegRepair.go
```
A file with the name `WhatsAppJpegRepair` (without any extension) will be generated. And now you can run the application on your MacOS without disabling Gatekeeper.

Next steps:

4. Place broken jpeg files from WhatsApp to the `whatsapp-files` directory, located in the application folder.
5. Run the application.
6. Go to the `repaired-files` folder to get repaired image files.

### Linux users

As you are a Linux user I suppose you are familiar with the terminal window and commands. Anyway for users' convenience I've added script file `runme.sh` file with some sample switches and options inside.
Just edit it using your text editor and add/remove options and switches of your choice.

Don't forget to assign `Executable` attribute to the `runme.sh` file via running the command:

`chmod +x runme.sh`

in the unzipped application folder.

Now, let's build the tool from the source code for your Linux operating system.

1. Install Go language support for Linux: https://golangdocs.com/install-go-linux
2. Download and unpack zip-file, built for MacOS users as advised above or clone this repository.
3. Delete existing `WhatsAppJpegRepair` file, please don't confuse this file with `WhatsAppJpegRepair.go`.
4. Open terminal in the unzipped application folder, and execute the following command:

```
go build WhatsAppJpegRepair.go
```
A file with the name `WhatsAppJpegRepair` (without any extension) will be generated.

Next steps:

5. Place broken jpeg files from WhatsApp to the `whatsapp-files` directory, located in the application folder.
6. Run the application.
7. Go to the `repaired-files` folder to get repaired image files.

### Operation modes, options and switches

Two operation modes are supported.

**Direct mode** - when you pass the path (or paths) to the image files that need fixing directly to the command line. In this case, the application processes files in their original location and overwrites them after processing. The application creates backup files that are deleted after the fix is applied. For example, for file `01.jpg`, the application creates a backup file named `01_wjr_backup_file.jpg` and deletes it when the operation is completed successfully. Use quotation marks if the path to the image file contains spaces.

For example:

`WhatsAppJpegRepair "d:\\my image files\\01.jpg"`

Multiple paths can be passed as space delimited arguments. For example:

`WhatsAppJpegRepair "d:\\my image files\\01.jpg" d:\\old-images\file.jpg "e:\\files and documents\\file.jpg"`

**Managed mode** - when you use command-line switches to control the application behavior.

#### Switches to run application in managed mode

`-srcPath` - contains path to the broken WhatsApp files
By default the application internal folder `whatsapp-files` is being used.
Currently this folder contains sample broken whatsapp jpeg images for demonstration purposes.

Example:
```
WhatsAppJpegRepair -srcPath=/home/username/Documents/Photos/WhatsAppFiles
```

this call will use `/home/username/Documents/Photos/WhatsAppFiles` folder as the source path to get broken whatsapp files.

`-destPath` - contains path to the folder, where repaired files will be stored.
By default the application internal folder `repaired-files` is being used.
If this folder does not exist, it will be created at runtime.

Example:
```
WhatsAppJpegRepair -srcPath=/home/username/Documents/Photos/WhatsAppFiles -destPath=/home/username/Documents/RepairedPhotos
```
this call will use `/home/username/Documents/Photos/WhatsAppFiles` folder to look for broken whatsapp files, and will use `/home/username/Documents/RepairedPhotos` folder to store repaired images.

`-dontWaitToClose` - if it is set to `true`, the application wil close when done, otherwise it will wait until user presses 'Enter'. Default value is `false`.

Example:
```
WhatsAppJpegRepair -srcPath=/home/username/Documents/Photos/WhatsAppFiles -dontWaitToClose=true
```
this call will use folder `/home/username/Documents/Photos/WhatsAppFiles` as a source files path, and application will be closed as it finished files processing. All repaired files will be stored to the default destination folder `repaired-files` (check `-destPath` option description above).

`-useCurrentModificationDateTime` - when set to `true`, this switch sets current date/time as repaired files' 'modified' attribute. By default it is set `false`: all repaired files retain the same file modification date/time as source (broken) image files.

Example:
```
WhatsAppJpegRepair -useCurrentModificationDateTime=true
```
this call will use default source and destination folders (check `-srcPath` and `-destPath` options above), the application will wait until user presses Enter to exit when all files are processed,
and current date/time will be set as file modification time for created repaired files.

`-deleteWhatsAppFiles` - when set to `true`, the application deletes every processed whatsapp file when done and only repaired files remain. By default it is `false`.

Example:
```
WhatsAppJpegRepair -deleteWhatsAppFiles=true
```
this call will use default source and destination folders (check `-srcPath` and `-destPath` options above), will preserve repaired file modification date/times (check `-useCurrentModificationDateTime` option above), will remove all processed source whatsapp files and will wait until user presses Enter to exit when all files are processed.

None of these options are mandatory. You can run the application without parameters, or set arbitrary set of parameters, default values will be applied for the rest.


## Building the application from the source

1. Download and install Go language support for your platform from the [official web-site](https://golang.org/)
2. In the application project folder run the following command:

```
go build WhatsAppJpegRepair.go
```
