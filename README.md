# WhatsApp Jpeg Image Repair

When you send jpeg files via WhatsApp and afterwards try to open received jpeg files in the Adobe Photoshop, you get the following error:

`Could not complete your request because a SOFn, DQT, or DHT JPEG marker is missing before a JPEG SOS marker`

In this cases users advised to open broken file in MS Paint (or something similar on MacOS) and save as jpeg file to solve this issue. But when you have many broken files, opening and saving every file is a kind of tedious work.

WhatsApp Jpeg Image Repair application solves this problem and can repair many broken files at once.

Follow these steps:
1. Download archive. Use [this link](https://github.com/cdefgah/whatsapp-jpeg-repair/releases). Expand `Assets` section and download either `WhatsAppJpegRepair-1.0-Windows.zip` or `WhatsApp Jpeg Image Repair (Apple MacOS version)` according to your operating system.
2. Extract application to an arbitrary folder. I recommend to use your home folder, for example `Documents` or `Desktop` folder, to prevent filesystem access control issues upon running the application.
3. Open application folder `WhatsAppJpegRepair`.
4. Place broken jpeg files from WhatsApp to the `whatsapp-files`.
5. Run the application.
6. Go to the `fixed-files` folder to get fixed image files.

It is possible to specify custom source and destination folder for the application. Just specify source and destination folders as application parameters, for example:
```
WhatsAppJpegRepair c:/Users/username/Documents/whats-app-files/ c:/Users/username/Documents/correct files/
```

If you are building the application from the source code, before using the built application, please delete `.gitkeep` files from the internal application folders.
