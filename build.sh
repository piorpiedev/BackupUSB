#!/bin/sh

echo "Compiling for windows/arm64"
GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o bin/backup.exe .

echo Compiling for linux/arm64
go build -ldflags "-s -w" -o bin/backup .

echo Stripping username
python3 stripBinaries.py backup $USER
