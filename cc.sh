#!/bin/sh

GOOS=windows GOARCH=amd64 go build -o watchdog_win64.exe
GOOS=windows GOARCH=386 go build -o watchdog_win32.exe 

GOOS=linux GOARCH=amd64 go build -o watchdog_linux64
GOOS=linux GOARCH=386 go build  -o watchdog_linux32