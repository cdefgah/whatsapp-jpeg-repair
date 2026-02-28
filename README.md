# WhatsApp Jpeg Image Repair

## version 3.0.0 (released on March XX, 2026)

[What's new](https://github.com/cdefgah/whatsapp-jpeg-repair/blob/master/CHANGELOG.md)

If you receive JPEG files via WhatsApp and then try to open them in Adobe Photoshop, you may get the following error message:

**Could not complete your request because a SOFn, DQT or DHT JPEG marker is missing before a JPEG SOS marker**.

In this case, users are usually advised to open the broken file in MS Paint (or a similar application on a Mac) and save it as a JPEG file. This usually solves the problem, but if you have many broken image files, opening and saving them one by one can be tedious.

The WhatsApp JPEG Image Repair application solves this problem by repairing multiple broken files at once.

**Breaking Changes in v3.0.0**

Application executable file and all command-line flags have been renamed in v3.0.0. See the [Migration Guide](#migration-guide) for details.

---

## Table of contents

- [How to Install](#how-to-install)
- [Quick Start & Usage Examples](#quick-start-and-usage)
- [How it Works: Direct vs Managed Mode](#how-it-works)
- [Managed Mode Parameters](#managed-mode-params)
- [Migration Guide](#migration-guide)
- [Frequently Asked Questions](#faq)
- [Building from Source](#building-from-source)

---

## <a name="how-to-instal">How to Install</a>

Download the application archive. Navigate to the [application releases](https://github.com/cdefgah/whatsapp-jpeg-repair/releases). Expand the **Assets** section, then download the application zip archive that is suitable for your operating system.

Here is the table showing which archive you should download according to your operating system and platform.

| Operating System | Platform                                                    | Which file should I download?              |
| ---------------- | ----------------------------------------------------------- | ------------------------------------------ |
| Windows 64-bit   | Intel/AMD PC and Laptops                                    | whatsapp_jpeg_repair_3_0_0_win_x64.zip     |
| Windows 32-bit   | Intel/AMD PC and Laptops                                    | whatsapp_jpeg_repair_3_0_0_win32.zip       |
| Windows          | New ARM Laptops (Microsoft Surface, Asus Zenbook A14, etc.) | whatsapp_jpeg_repair_3_0_0_win_arm64.zip   |
| macOS            | Apple Silicon (M1/M2/M3/etc)                                | whatsapp_jpeg_repair_3_0_0_macOS_arm64.zip |
| macOS            | Intel Processor                                             | whatsapp_jpeg_repair_3_0_0_macOS_x64.zip   |
| Linux 64-bit     | Intel/AMD PC and Laptops                                    | whatsapp_jpeg_repair_3_0_0_linux_x64.zip   |

Extract the application archive to a folder on your desktop or in your Documents folder. This ensures that there are no permission issues when running the application. If you choose to extract the archive to a different location, please ensure that the user running the application has write permissions to that folder and all its subfolders.

## <a name="quick-start-and-usage">Quick Start & Usage Examples</a>

Further instructions are provided below depending on the operating system you are using.

### Windows users

The simplest way is to put the files downloaded from WhatsApp into the **whatsapp-files** folder. Launch the **whatsapp-jpeg-repair.exe** application by double-clicking on it and wait until it has finished. Then get the repaired files from the repaired-files folder.

It is also possible to drag and drop an image file from WhatsApp to **whatsapp-jpeg-repair.exe**. In this case, the file will be repaired in situ.

If you want to use all the application's capabilities, refer to the **runme.bat** file. Open the file in Notepad and edit the following line:

```
whatsapp-jpeg-repair --dont-wait-to-close=false --use-current-modification-time=false --delete-whatsapp-files=false
```

This line shows the application launching with three random options selected. Feel free to use any other options or combinations that you think are necessary.

**whatsapp-jpeg-repair** is the application name; do not alter this text. However, you can add, remove or alter any parameters that follow the application name according to your requirements. More information on the available parameters is provided in the [Managed Mode Parameters](#managed-mode-params) section.

Once you have finished making changes to the **runme.bat** file, save it, close Notepad, then double-click on **runme.bat**. This will launch **whatsapp-jpeg-repair** with all the parameters you provided in the **runme.bat** file.

### macOS and Linux users

Before using the application, **it is important** to perform the following steps. Open terminal in the application folder. Then run the following command in the terminal:

```bash
chmod +x ./prepare.sh
```

This will set the **executable** attribute for the **prepare.sh** script file. Then run the following command:

```bash
./prepare.sh
```

It will set the required attributes for both the application file and the runme.sh file.

You can now put the files downloaded from WhatsApp into the **whatsapp-files** folder. Launch the **whatsapp-jpeg-repair** application by double-clicking on it. Then retrieve the repaired files from the **repaired-files** folder.

Mac OS users, please note that if you launch the application by double-clicking on it in the graphical user interface, the application window will open and you will need to close it manually when the text **[Process completed]** appears.

Please note that Linux users will not see an additional window when they double-click an executable file in the graphical user interface. This may be inconvenient. To control the application's output, open a Terminal window in the application's folder and enter the following command:

```bash
./whatsapp-jpeg-repair
```

You can use the **runme.sh** script to add an additional combination of supported options to the application command line. Open the file in a text editor and scroll down to the relevant line.

```bash
./whatsapp-jpeg-repair --dont-wait-to-close=false --use-current-modification-time=false --delete-whatsapp-files=false
```

This line shows the application being launched with three random options selected. Feel free to use any other options you think are necessary.

Do not alter the text **./whatsapp-jpeg-repair** as this is the name of the application itself. However, you can add, remove or alter any parameters that follow the application name according to your requirements. More information on the available parameters is provided in the [Managed Mode Parameters](#managed-mode-params) section.

Once you have finished editing the **runme.sh** file, save it and close the text editor. Then, open a terminal window in the application folder. And run the command:

```bash
./runme.sh
```

This will launch **whatsapp-jpeg-repair** with all the parameters you provided in the **runme.sh** file.

## <a name="how-it-works">How it Works: Direct vs Managed Mode</a>

The application supports two operating modes: **Direct Mode** and **Managed Mode**.

### Direct Mode

**Direct Mode** is used when an image file is dragged and dropped onto the whatsapp-jpeg-repair application file. In this case, the file will be repaired in situ.

This mode is also used when passing an arbitrary number of file paths via the command line, as demonstrated below.

For macOS/Linux environment:

```bash
./whatsapp-jpeg-repair /home/yourusername/Documents/photo126.jpeg /home/yourusername/Documents/Scans/photo18.jpeg /home/yourusername/Documents/Archive/photo154.jpeg
```

For Windows environment:

```
whatsapp-jpeg-repair c:\Users\yourusername\Documents\photo126.jpeg c:\Users\yourusername\Documents\Scans\photo18.jpeg c:\Users\yourusername\Documents\Archive\photo154.jpeg
```

All of these files: **photo126.jpeg**, **photo18.jpeg** and **photo154.jpeg** will be processed and saved in the same location.

### Managed Mode

**Managed mode** is used when the application is started without parameters or via a double-click of the mouse. This mode is also turned on if at least one managed mode parameter is included in the command-line arguments.

If the application is launched without parameters or via a double-click of the mouse, all managed options use their default values. To find the files to be repaired, the application looks in the **whatsapp-files** folder, which is located in the same folder as the application file. The repaired files are stored in the **repaired-files** folder, which is also located in the same folder as the application file.

For your convenience, you can use the **runme.bat** (for Windows users) or **runme.sh** (for macOS/Linux users) scripts, or create your own. If you are a macOS or Linux user creating new custom script files, remember to set the **executable** attribute for each new file.

```bash
chmod +x ./your-new-custom-script.sh
```

## <a name="managed-mode-params">Managed Mode Parameters</a>

## Usage Options

| Option                              | Shorthand | Description                                       | Default              |
| :---------------------------------- | :-------- | :------------------------------------------------ | :------------------- |
| **--src-path**                      | **-s**    | Path to the folder with broken WhatsApp files     | **./whatsapp-files** |
| **--dest-path**                     | **-d**    | Path to store repaired files (created if missing) | **./repaired-files** |
| **--use-current-modification-time** | **-t**    | Use current time for file modification date       | **false**            |
| **--delete-whatsapp-files**         | **-w**    | Delete source files after successful processing   | **false**            |
| **--process-nested-folders**        | **-n**    | Process files in subfolders recursively           | **false**            |
| **--dont-wait-to-close**            | **-c**    | Exit immediately after completion                 | **false**            |
| **--help**                          | **-h**    | Show all available options                        | -                    |

Note: The **-c**, **--dont-wait-to-close** flag is automatically ignored in non-interactive sessions. If the application detects that output is being redirected to a file (e.g., **whatsapp-jpeg-repair 2>log.txt**), it will exit immediately upon completion without waiting for a keypress.

## <a name="migration-guide">Migration Guide</a>

The name of the executable file has changed. The old name was **WhatsAppJpegRepair**, the new name is **whatsapp-jpeg-repair**. If you have any custom script files that call the executable, please update them accordingly.

The default behavior remains unchanged. When launched without parameters — either via terminal or by double-clicking the executable — the application searches for source files in the **whatsapp-files** folder located in the same directory as the executable. Repaired results are saved to the **repaired-files** folder in the same location.

Note: The old parameter format (camelCase with a single dash) is now deprecated. Please use the new kebab-case format or shorthands for future compatibility.

This table shows the correspondence between the old and new parameter names.

| Old parameter name                  | New parameter name                  | Shorthand |
| :---------------------------------- | :---------------------------------- | :-------- |
| **-srcPath**                        | **--src-path**                      | **-s**    |
| **-destPath**                       | **--dest-path**                     | **-d**    |
| **-dontWaitToClose**                | **--dont-wait-to-close**            | **-c**    |
| **-useCurrentModificationDateTime** | **--use-current-modification-time** | **-t**    |
| **-deleteWhatsAppFiles**            | **--delete-whatsapp-files**         | **-w**    |

The app now only processes JPEG-related image files with the following extensions: **.jpg**, **.jpeg**, **.jpe**, **.jif**, **.jfif** and **.jfi**, which are case insensitive. Files with other extensions will be ignored. This is because WhatsApp converts received images to JPEG format. The previous version of the application processed all files indiscriminately. It converted non-JPEG files that did not require repair into JPEG format while retaining their original extensions.

## <a name="faq">Frequently Asked Questions</a>

1. [Windows user: What is the purpose of the .bat files in the application folder?](#bat-purpose)
2. [macOS/Linux user: What is the purpose of the **.sh** files in the application folder?](#sh-purpose)
3. [How should parameters and their combinations be passed correctly when launching an application?](#param-rules)
4. [How can I specify a custom folder from which the application will take files for processing?](#custom-source-folder)
5. [How can I specify a custom folder in which the application can store the results of file processing?](#custom-dest-folder)
6. [How can the same modification time be applied to the repaired file as to the original file?](#mod-time)
7. [What steps should I take to ensure that the source files are deleted after processing?](#delete-source-files)
8. [What is the best course of action if I want all my source files, including those in subfolders, to be processed?](#recursive-processing)
9. [How can I stop the program from waiting for me to press Enter after it has finished running?](#dont-wait)
10. [How can I display information on all the possible operating modes and command line parameters?](#show-help)
11. [What if I want to process a file so that it is processed directly on the spot?](#single-file-in-direct-mode)
12. [How can the application be launched so that its output is redirected to a file?](#redirect-app-output)
13. [Why is nothing happening when I try to run a **.sh** or **.bat** file?](#cant-run-script-file)
14. [Why am I getting permission errors when I run the application?](#permission-problems)
15. [How should I go about reporting a bug or suggesting an improvement?](#how-to-report-bug)

### <a name="bat-purpose">1. Windows user: What is the purpose of the .bat files in the application folder?</a>

This optional file contains the application call command with various flags. This file is provided for convenience. You can edit the file, adding or removing any parameters supported by the application. Once saved, you can run the file by double-clicking on it.

### <a name="sh-purpose">2. macOS/Linux user: What is the purpose of the **.sh** files in the application folder?</a>

The **prepare.sh** file is very important and should not be edited. Simply set the **executable** attribute using the following terminal command:

```bash
chmod +x ./prepare.sh.
```

After that, you only need to run it once. You will not need to run it again for the installed version of the application. This sets the correct attributes for the application file and the auxiliary **runme.sh** file. The **runme.sh** file is designed for convenience. It contains a call to the application with an arbitrary set of parameters. You can edit this set of parameters to include only those you consider necessary. After that, you can execute this file in the terminal:

```bash
./runme.sh.
```

### <a name="param-rules">3. How should parameters and their combinations be passed correctly when launching an application?</a>

You can pass any number of supported parameters when launching the application. Parameters are separated by spaces.
For example, the following command launches the application with custom settings for the source and output files and specifies that the current time be used as the last modification time for all output files.

Example for Windows environment:

```
whatsapp-jpeg-repair --src-path=C:\BrokenFiles --dest-path=C:\Users\YourUserName\Documents\RepairedFiles --use-current-modification-time=true
```

The procedure is the same for Linux and macOS, except the executable file is specified as **./whatsapp-jpeg-repair** .

### <a name="custom-source-folder">4. How can I specify a custom folder from which the application will take files for processing?</a>

To specify the path to a custom folder from which the application will take files for processing, use the **--src-path** parameter or its shortcut, **-s**.

Examples:

Examples for Windows environment:

```
whatsapp-jpeg-repair --src-path=C:\BrokenFiles
```

or

```
whatsapp-jpeg-repair -s=C:\BrokenFiles
```

The procedure is the same for Linux and macOS, except the executable file is specified as **./whatsapp-jpeg-repair** .

Please note that we only specify the folder containing the source files here. All other parameters will take their default values. In particular, you will need to look for the results of the processing in the **repaired-files** folder, which is located in the same folder as the application executable file. To customise both the source and destination folders, specify the most suitable paths for your task in both the source and destination parameters.

Note: If the path contains spaces, it must be enclosed in double quotation marks.

### <a name="custom-dest-folder">5. How can I specify a custom folder in which the application can store the results of file processing?</a>

To specify the path to a custom folder from which the application will store result files, use the **--dest-path** parameter or its shortcut, **-d**.

Examples for Windows environment:

```
whatsapp-jpeg-repair --dest-path=C:\Users\YourUserName\Documents\RepairedFiles
```

or

```
whatsapp-jpeg-repair -d=C:\Users\YourUserName\Documents\RepairedFiles
```

The procedure is the same for Linux and macOS, except the executable file is specified as ./whatsapp-jpeg-repair.

Please note that we only specify the folder for saving the results of file processing here. All other parameters are set to their default values. In particular, the folder in which the application searches for files that need repairing is the **whatsApp-files** folder, which is located in the same folder as the application's executable file. To customise both the source and destination folders, specify the most suitable paths for your task in both the source and destination parameters.

Note: If the path contains spaces, it must be enclosed in double quotation marks.

### <a name="mod-time">6. How can the same modification time be applied to the repaired file as to the original file?</a>

By default, the application automatically sets the modification time of processed files to be the same as the original file's. To change this behavior and set the modification time of the processed file to the actual processing time, use the **--use-current-modification-time** or **-t** parameters.

Examples:

For Windows environment:

```
whatsapp-jpeg-repair --use-current-modification-time=true
```

or

```
whatsapp-jpeg-repair --use-current-modification-time
```

or

```
whatsapp-jpeg-repair -t=true
```

or

```
whatsapp-jpeg-repair -t
```

The procedure is the same for Linux and macOS, except the executable file is specified as **./whatsapp-jpeg-repair** .

Please note that setting a parameter with a logical type without assigning a value is equivalent to setting it to true.

### <a name="delete-source-files">6. What steps should I take to ensure that the source files are deleted after processing?</a>

To delete the source files after processing, use the **--delete-whatsapp-files** option, or the shortened version, **-w** .

For Windows environment:

```
whatsapp-jpeg-repair --delete-whatsapp-files=true
```

or

```
whatsapp-jpeg-repair --delete-whatsapp-files
```

or

```
whatsapp-jpeg-repair -w=true
```

or

```
whatsapp-jpeg-repair -w
```

The procedure is the same for Linux and macOS, except the executable file is specified as **./whatsapp-jpeg-repair** .

Please note that setting a parameter with a logical type without assigning a value is equivalent to setting it to true.

### <a name="recursive-processing">8. What is the best course of action if I want all my source files, including those in subfolders, to be processed?</a>

Let's assume the source files are located in **C:\Users\Username\Documents\ImageArchive**, and that there are many nested folders containing files that should also be processed. You want to process all the files in that folder, including all the nested folders, and place the resulting files in the folder: **C:\Users\Username\Documents\ProcessedFiles**

To complete this task, we must launch the application with the additional **--process-nested-folders** or **-n** parameter. Here's an example for Windows environment:

```
whatsapp-jpeg-repair --src-path=C:\Users\Username\Documents\ImageArchive --dest-path=C:\Users\Username\Documents\ProcessedFiles --process-nested-folders
```

The procedure is the same for Linux and macOS, except the executable file is specified as **./whatsapp-jpeg-repair** .

### <a name="dont-wait">9. How can I stop the program from waiting for me to press Enter after it has finished running?</a>

When started manually, the application waits for the Enter key to be pressed after processing the files to exit. This allows the user time to read the information in the window, in case the application is started from the graphical shell of the operating system. You can disable this wait time by passing the parameter **--dont-wait-to-close** or **-c** .

The wait will not be performed if the application is run with output redirected to a file, regardless of the parameter passed.

For example (Windows environment):

```
whatsapp-jpeg-repair 2>log.txt
```

The procedure is the same for Linux and macOS, except the executable file is specified as **./whatsapp-jpeg-repair** .

In this case, everything that would normally be displayed on the screen will be written to the **log.txt** file, and then the application will end.

### <a name="show-help">10. How can I display information on all the possible operating modes and command line parameters?</a>

To accomplish this, you must either call the application with an invalid parameter or use the **--help** or **-h** parameter.

Example for Windows environment:

```
whatsapp-jpeg-repair --help
```

or

```
whatsapp-jpeg-repair -h
```

The procedure is the same for Linux and macOS, except the executable file is specified as **./whatsapp-jpeg-repair** .

### <a name="single-file-in-direct-mode">11. What if I want to process a file so that it is processed directly on the spot?</a>

In the operating system's graphical shell, drag the file requiring processing onto the application's executable file icon.

### <a name="redirect-app-output">12. How can the application be launched so that its output is redirected to a file?</a>

Any information written by the application to the console is written to the **stderr** stream. Nothing is written to the **stdout** stream since the application only outputs progress information during operation. To redirect the application output, add a line like the following to the end of the command line parameters, separated by a space: **2>log.txt**, where **log.txt** is the name of the file to which the output will be redirected.

For example (Windows environment):

```
whatsapp-jpeg-repair 2>log.txt
```

The procedure is the same for Linux and macOS, except the executable file is specified as **./whatsapp-jpeg-repair** .

### <a name="cant-run-script-file">13. Why is nothing happening when I try to run a **.sh** file?</a>

If nothing happens when you run **./runme.sh**, make sure that you have set the executable attribute for **./prepare.sh** and that you have run it.

### <a name="permission-problems">14. Why am I getting permission errors when I run the application?</a>

If you encountered this problem after launching the application, then either you placed the application in a folder to which you do not have write access, or the application is attempting to write processed files to a folder to which you do not have write access.

### <a name="how-to-report-bug">15. How should I go about reporting a bug or suggesting an improvement?</a>

1. Open this page in the browser: [https://github.com/cdefgah/whatsapp-jpeg-repair](https://github.com/cdefgah/whatsapp-jpeg-repair)
2. Post your report in the "Issues" section at the top of the page.
3. If you don't have a GitHub account, email me at rafael.osipov (at) outlook.com

## <a name="building-from-source">Building from Source</a>

Step 1: Make sure you have the **Make** tool installed. This tool is usually preinstalled on Linux and Mac operating systems. Windows users can install Make via [Chocolatey](https://chocolatey.org/). To do this, run the following command in Terminal:

```
choco install make
```

Step 2: Clone the project source to your local hard drive.

Step 3: Open a terminal window in the cloned application folder where the **Makefile** is located.

Step 4: Run the command:

```bash
make build
```

Step 5: To access the built application files, navigate to the created **./dist** folder.
