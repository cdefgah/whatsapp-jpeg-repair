# WhatsApp Jpeg Image Repair
### version 2.1.1 (released on June 05, 2021)

[![](https://github.com/cdefgah/whatsapp-jpeg-repair/workflows/build/badge.svg)](https://github.com/cdefgah/whatsapp-jpeg-repair/actions)

When you are sent jpeg files via WhatsApp and then try to open received files in the Adobe Photoshop, there's a chance you'll get the following error in Photoshop:

`Could not complete your request because a SOFn, DQT, or DHT JPEG marker is missing before a JPEG SOS marker`

For this case users are usually advised to open the broken file in MS Paint (or something similar in MacOS) first and save it as jpeg file. Usually it helps, but when you have many broken image files, opening and saving them one by one may get a little tedious.

WhatsApp Jpeg Image Repair application solves this problem by repairing multiple broken files at once.

Follow these steps:
1. Download application archive. Navigate to [the application releases](https://github.com/cdefgah/whatsapp-jpeg-repair/releases). Then expand `Assets` section and download application zip archive suitable for your operating system.
2. Unzip application archive to the folder of your choosing, preferrably your home folder, e.g. `Documents` or `Desktop` folder, to prevent filesystem access control issues upon running the application. If you use MacOS, please note that MacOS Gatekeeper blocks this application when you try to launch it. To run downloaded application on MacOS your should either [disable MacOS Gatekeeper](https://www.google.com/search?q=disable+MacOS+Gatekeeper) or build the application from the source code as described below, and replace existing executable in downloaded zip archive with the newly-built file.
3. Open application folder `WhatsAppJpegRepair`.
4. Place broken jpeg files from WhatsApp to the `whatsapp-files` directory, located in the application folder.
5. Run the application.
6. Go to the `repaired-files` folder to get repaired image files.

Options and switches:

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

For users' convenience I've added script files `runme.bat` (for Windows) and `runme.sh` (for MacOS). These scripts launch the application with some parameters set to default values. Just edit these files, change the relevant property values, and/or add/remove properties you want, according to the instruction, save the script file and just run it.

MacOS users, before using `runme.sh` open terminal in the application folder.
Here is how: https://apple.stackexchange.com/questions/11323/how-can-i-open-a-terminal-window-directly-from-my-current-finder-location

And then type in the terminal the following command:

`chmod +x runme.sh`

And press Enter. After that just close the terminal window.

Please note, that `runme.bat` and `runme.sh` files are plain text files. And you don't need a special application to edit these files. 

Just use a simple text editor of your choice, like `Notepad++` for Windows, or `TextEdit` on MacOS.

For Windows users, install Notepad++: https://notepad-plus-plus.org/downloads/
Then, just do the mouse right click on the `runme.bat` and select "Edit with Notepad++".

For MacOS users, launch `TextEdit`, choose `File - Open` from menu and select `runme.sh` file:
https://support.apple.com/guide/textedit/open-documents-txte51413d09/mac


## Building the application from the source

1. Download and install Go language support: https://golang.org/
2. In the application project folder run the following command:
```
go build WhatsAppJpegRepair.go
```
