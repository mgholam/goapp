#!/bin/bash

ANAME="goapp"
# -------------------------------------------------------------
echo "building linux native..."

GOOS=linux \
CGO_ENABLED=1 \
CC="zig cc -target native-native-musl" \
go build -o output/${ANAME} -ldflags "-w -s" ./
strip output/${ANAME}
# -------------------------------------------------------------
echo "building linux x64..."

GOOS=linux \
GOARCH=amd64 \
CGO_ENABLED=1 \
CC="zig cc -target x86_64-linux-musl" \
go build -o output/${ANAME}-linux -ldflags "-w -s" ./
strip output/${ANAME}-linux

# -------------------------------------------------------------
echo "building windows..."

CGO_ENABLED=1 \
GOOS=windows \
GOARCH=amd64 \
CC="zig cc -target x86_64-windows-gnu" \
go build -o output/${ANAME}64.exe -ldflags "-w -s" ./

CGO_ENABLED=1 \
GOOS=windows \
GOARCH=386 \
CC="zig cc -target i386-windows-gnu" \
go build -o output/${ANAME}386-2.exe -ldflags "-w -s" ./

# -------------------------------------------------------------
# echo "building darwin..."

# CGO_ENABLED=1 \
# GOOS=darwin \
# GOARCH=arm64 \
# CC="zig cc -target aarch64-macos-gnu" \
# go build -o output/${ANAME}-darwin -ldflags "-w -s" ./

# -------------------------------------------------------------
echo "building pi zero w..."

GOARM=5 \
GOOS=linux \
GOARCH=arm \
CGO_ENABLED=1 \
CC="zig cc -target arm-linux-musleabihf -march=arm1176jzf_s" \
go build -o output/${ANAME}-pizw -ldflags "-w -s" ./
