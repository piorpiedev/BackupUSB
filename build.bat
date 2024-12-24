@echo off

echo Compiling for windows/arm64
go build -ldflags "-w" -o bin/backup.exe .

rem echo Compiling for linux/arm64
rem set "GOOS=linux" & go build -ldflags "-w" -o bin/backup .

echo Stripping username
py stripBinaries.py
