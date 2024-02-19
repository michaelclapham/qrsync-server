#!/bin/bash

# Kill process thats already using port 4010
PROC_USING_PORT=$(lsof -i tcp:4010 | awk 'NR!=1 {print $2}')

if [ ! -z $PROC_USING_PORT ]
then
    echo "Killing process $PROC_USING_PORT that was already using port 4010"
    kill $PROC_USING_PORT
fi

echo "Starting QR sync server"

./qrsync-server