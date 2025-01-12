@echo off

echo Compiling for windows/arm64
go build -ldflags "-s -w" -o bin/backup.exe .

echo Compiling for linux/arm64
set "GOOS=linux" & set "GOARCH=amd64" & go build -ldflags "-s -w" -o bin/backup .

echo Stripping username
py stripBinaries.py backup
