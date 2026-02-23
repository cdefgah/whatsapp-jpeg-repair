# WhatsApp Jpeg Image Repair

## version 3.0.0 (released on March XX, 2026)

[Changelog](./CHANGELOG.md)

If you receive JPEG files via WhatsApp and then try to open them in Adobe Photoshop, you may get the following error message:

`Could not complete your request because a SOFn, DQT or DHT JPEG marker is missing before a JPEG SOS marker`.

In this case, users are usually advised to open the broken file in MS Paint (or a similar application on a Mac) and save it as a JPEG file. This usually solves the problem, but if you have many broken image files, opening and saving them one by one can be tedious.

The WhatsApp JPEG Image Repair application solves this problem by repairing multiple broken files at once.

**Breaking Changes in v3.0.0**

All command-line flags have been renamed in v3.0.0. See the [Migration Guide](#migration-guide) for details.

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

Download the application archive. Navigate to the [application releases](https://github.com/cdefgah/whatsapp-jpeg-repair/releases). Expand the `Assets` section, then download the application zip archive that is suitable for your operating system.

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

The simplest way is to put the files downloaded from WhatsApp into the `whatsapp-files` folder. Launch the `WhatsAppJpegRepair.exe` application by double-clicking on it and wait until it has finished. Then get the repaired files from the repaired-files folder.

It is also possible to drag and drop an image file from WhatsApp to `WhatsAppJpegRepair.exe`. In this case, the file will be repaired in situ.

If you want to use all the application's capabilities, refer to the `runme.bat` file. Open the file in Notepad and edit the following line:

```
WhatsAppJpegRepair --dont-wait-to-close=false --use-current-modification-time=false --delete-whatsapp-files=false
```

This line shows the application being launched with three random options selected. Feel free to use any other options you think are necessary.

`WhatsAppJpegRepair` is the application name; do not alter this text. However, you can add, remove or alter any parameters that follow the application name according to your requirements. More information on the available parameters is provided in the [Managed Mode Parameters](#managed-mode-params) section.

Once you have finished making changes to the `runme.bat` file, save it, close Notepad, then double-click on `runme.bat`. This will launch `WhatsAppJpegRepair` with all the parameters you provided in the `runme.bat` file.

### macOS and Linux users

You need to perform the following steps once before using the application. Open terminal in the application folder. Then run the following command in the terminal:

```bash
chmod +x ./prepare.sh
```

This will set the `executable` attribute for the `prepare.sh` script file. Then run the following command:

```bash
./prepare.sh
```

It will set the required attributes for both the application file and the runme.sh file.

You can now put the files downloaded from WhatsApp into the `whatsapp-files` folder. Launch the `WhatsAppJpegRepair` application by double-clicking on it. Then retrieve the repaired files from the `repaired-files` folder.

Mac OS users should note that they will need to close the displayed window manually when the text `[Process completed]` appears.

Please note that Linux users will not see an additional window, which may be inconvenient.

To control the application's output, open a Terminal window in the application's folder and enter the following command:

```bash
./WhatsAppJpegRepair
```

You can use the `runme.sh` script to add extra options to the application command line. Open the file with a text editor and scroll down to the line.

```bash
./WhatsAppJpegRepair --dont-wait-to-close=false --use-current-modification-time=false --delete-whatsapp-files=false
```

This line shows the application being launched with three random options selected. Feel free to use any other options you think are necessary.

Do not alter the text `./WhatsAppJpegRepair` as this is the name of the application itself. However, you can add, remove or alter any parameters that follow the application name according to your requirements. More information on the available parameters is provided in the [Managed Mode Parameters](#managed-mode-params) section.

Once you have finished editing the `runme.sh` file, save it and close the text editor. Then, open a terminal window in the application folder. And run the command:

```bash
./runme.sh
```

This will launch `WhatsAppJpegRepair` with all the parameters you provided in the `runme.sh` file.

## <a name="how-it-works">How it Works: Direct vs Managed Mode</a>

The application supports two operating modes: `Direct Mode` and `Managed Mode`.

### Direct Mode

`Direct Mode` is used when an image file is dragged and dropped onto the WhatsAppJpegRepair application file. In this case, the file will be repaired in situ.

This mode is also used when passing an arbitrary number of file paths via the command line, as demonstrated below.

For macOS/Linux environment:

```bash
./WhatsAppJpegRepair /home/yourusername/Documents/photo126.jpeg /home/yourusername/Documents/Scans/photo18.jpeg /home/yourusername/Documents/Archive/photo154.jpeg
```

For Windows environment:

```console
WhatsAppJpegRepair c:\Users\yourusername\Documents\photo126.jpeg c:\Users\yourusername\Documents\Scans\photo18.jpeg c:\Users\yourusername\Documents\Archive\photo154.jpeg
```

All of these files: `photo126.jpeg`, `photo18.jpeg` and `photo154.jpeg` will be processed and saved in the same location.

### Managed Mode

`Managed mode` is used when the application is started without parameters or via a double-click of the mouse. This mode is also turned on if at least one managed mode parameter is included in the command-line arguments.

If the application is launched without parameters or via a double-click of the mouse, all managed options use their default values. To find the files to be repaired, the application looks in the `whatsApp-files` folder, which is located in the same folder as the application file. The repaired files are stored in the `repaired-files` folder, which is also located in the same folder as the application file.

For your convenience, you can use the `runme.bat` (for Windows users) or `runme.sh` (for macOS/Linux users) scripts, or create your own. If you are a macOS/Linux user creating new script files, remember to set the `executable` attribute for the new script file.

```bash
chmod +x ./your-new-script.sh
```

## <a name="managed-mode-params">Managed Mode Parameters</a>

Таблица всех доступных новых флагов.

Про интерактивный режим и ожидание в конце работы порграммы. Ожидание всегда отключено при запуске с пайпом.

## <a name="migration-guide">Migration Guide</a>

Таблица соответствия старых флагов новым.

## <a name="faq">Frequently Asked Questions</a>

Здесь можно описать всякое такое, как открыть консоль, что делать при ошибке прав на Mac/Linux и т.д.

## <a name="building-from-source">Building from Source</a>

Тут написать про порядок сборки.
