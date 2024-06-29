#!/bin/bash

# Check if an argument is provided
if [ $# -eq 0 ]; then
    echo "Please provide the rootfs path as an argument."
    echo "Usage: $0 <rootfs_path>"
    exit 1
fi

# Store the rootfs path from the argument
rootfs="$1"

# Create the rootfs directory
mkdir -p "$rootfs"

# Array of directories and files to sync
dirs=("/bin" "/lib" "/lib64" "/usr/bin" "/usr/lib" "/usr/lib64" "/etc" "./app" "./monitor_res.sh")

# Sync each item in the array
for item in "${dirs[@]}"; do
    echo "Syncing $item to $rootfs"
    if [ -f "$item" ]; then
        # If it's a file, copy it directly to $rootfs
        rsync -av --ignore-existing "$item" "$rootfs/"
    else
        # If it's a directory, create the necessary parent directories and sync
        mkdir -p "$rootfs/$item"
        rsync -av --ignore-existing "$item/" "$rootfs/$item/"
    fi
done

echo "All specified directories and files synced to $rootfs"
