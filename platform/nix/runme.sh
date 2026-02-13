#!/bin/bash

# Important! Please remember to set the eXecutable attribute for the runme.sh file. This can be done via the command: chmod +x ./runme.sh.
# If you have run the prepare.sh script, this attribute will already be set.
# 
# Description:
#         The application operates in one of two modes, depending on the arguments provided.
# 
#         Managed mode is used when no arguments are provided or at least one managed option is specified. All managed options are optional and have default values.
# 
#         Direct mode is used when only positional arguments are provided and no known managed options are present. 
#         In this mode, the positional arguments are treated as paths to files and processed in place.
# 
# A list of the available managed options is shown below.
# 
#   -s, --src-path string                 Path to the folder containing the broken WhatsApp files.
#                                         Example: --src-path=/home/yourusername/Documents/brokenWhatsAppFiles.
#
#   -d, --dest-path string                This is the path to the folder where the repaired files will be stored.
#                                         If the folder does not exist, it will be created.
#                                         Example: --dest-path=/home/yourusername/Documents/repairedImageFiles.
#
#   -t, --use-current-modification-time   If this is true, the current time will be set as the file's modification time. The default is the modification time of the source file.
#
#   -w, --delete-whatsapp-files           If it is true, the successfully processed source WhatsApp files will be deleted. Default: false.
#
#   -n, --process-nested-folders          If it is true, then the application processes files in nested folders recursively. Default: false.
#
#   -c, --dont-wait-to-close              If this is true, the application will exit immediately once processing is complete. Default: false.
#
#   -h, --help                            Show this help message and exit.

WhatsAppJpegRepair --dont-wait-to-close=false --use-current-modification-time=false --delete-whatsapp-files=false

