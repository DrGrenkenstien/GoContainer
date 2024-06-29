#!/bin/bash

if [ "$#" -ne 0 ]; then
    echo "Ignoring passed arguments"
fi

echo "Showing stats for process id : $PPID"

./app "$PPID"
