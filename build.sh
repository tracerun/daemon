#!/bin/bash

mkdir -p ./builds

program="tracerun"
winprogram="tracerun.exe"
tag="$1"

# build Mac 64bit program
env GOOS=darwin GOARCH=amd64 go build -o $program main.go
maczip=$(printf "%s_%s_darwin_amd64.zip" "$program" "$tag")
zip -r ./builds/$maczip $program

# build Windows 32bit program
env GOOS=windows GOARCH=386 go build -o $winprogram main.go
winzip32=$(printf "%s_%s_windows_386.zip" "$program" "$tag")
zip -r ./builds/$winzip32 $winprogram

# build Windows 64bit program
env GOOS=windows GOARCH=amd64 go build -o $winprogram main.go
winzip64=$(printf "%s_%s_windows_amd64.zip" "$program" "$tag")
zip -r ./builds/$winzip64 $winprogram

# build Linux 32bit program
env GOOS=linux GOARCH=386 go build -o $program main.go
linux32=$(printf "%s_%s_linux_386.tar.gz" "$program" "$tag")
tar -cvzf ./builds/$linux32 $program

# build Linux 64bit program
env GOOS=linux GOARCH=amd64 go build -o $program main.go
linux64=$(printf "%s_%s_linux_amd64.tar.gz" "$program" "$tag")
tar -cvzf ./builds/$linux64 $program

# build Linux 64bit arm program
env GOOS=linux GOARCH=arm64 go build -o $program main.go
linux64arm=$(printf "%s_%s_linux_arm64.tar.gz" "$program" "$tag")
tar -cvzf ./builds/$linux64arm $program

# build FreeBSD 64bit program
env GOOS=freebsd GOARCH=amd64 go build -o $program main.go
freebsd64=$(printf "%s_%s_freebsd_amd64.tar.gz" "$program" "$tag")
tar -cvzf ./builds/$freebsd64 $program