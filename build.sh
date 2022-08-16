#!/bin/bash

ANAME="goapp"
ZIG="zig cc -static "
# -------------------------------------------------------------
echo "building linux native..."

ON=${ANAME}
GOOS=linux \
CGO_ENABLED=1 \
CC="${ZIG} -target native-native-musl" \
go build -o output/${ON} -ldflags "-w -s" ./
strip output/${ON}

# -------------------------------------------------------------
echo "building linux x64..."

ON=${ANAME}-linux
GOOS=linux \
GOARCH=amd64 \
CGO_ENABLED=1 \
CC="${ZIG} -target x86_64-linux-musl" \
go build -o output/${ON} -ldflags "-w -s" ./
strip output/${ON}

# -------------------------------------------------------------
echo "building android x64..."

ON=${ANAME}-android
GOOS=linux \
GOARCH=arm64 \
CGO_ENABLED=1 \
CC="${ZIG} -static -target aarch64-linux-musl" \
go build -o output/${ON} -ldflags "-w -s" ./
strip output/${ON}

# -------------------------------------------------------------
echo "building windows..."

ON=${ANAME}64.exe
CGO_ENABLED=1 \
GOOS=windows \
GOARCH=amd64 \
CC="${ZIG} -static -target x86_64-windows-gnu" \
go build -o output/${ON} -ldflags "-w -s" ./

ON=${ANAME}386.exe
CGO_ENABLED=1 \
GOOS=windows \
GOARCH=386 \
CC="${ZIG} -static -target i386-windows-gnu" \
go build -o output/${ON} -ldflags "-w -s" ./

# -------------------------------------------------------------
echo "building pi zero w..."

ON=${ANAME}-pizw
GOARM=5 \
GOOS=linux \
GOARCH=arm \
CGO_ENABLED=1 \
CC="${ZIG} -static -target arm-linux-musleabihf -march=arm1176jzf_s" \
go build -o output/${ON} -ldflags "-w -s" ./
strip output/${ON}

# -------------------------------------------------------------
# echo "building darwin..."

# CGO_ENABLED=1 \
# GOOS=darwin \
# GOARCH=arm64 \
# CC="zig cc -target aarch64-macos-gnu" \
# go build -o output/${ANAME}-darwin -ldflags "-w -s" ./