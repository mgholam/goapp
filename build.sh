ANAME="goapp"
# -------------------------------------------------------------
echo "building linux..."

GOOS=linux \
CGO_ENABLED=1 \
CC="zig cc -target native-native-musl" \
CXX="zig cc -target native-native-musl" \
go build -o output/${ANAME} -ldflags "-w -s" ./
strip output/${ANAME}


# -------------------------------------------------------------
echo "building windows..."

# # gcc
# GOOS=windows \
# GOARCH=386 \
# CGO_ENABLED=1 \
# CC=i686-w64-mingw32-g++ \
# CC=i686-w64-mingw32-gcc \
# go build -o output/${ANAME}386.exe -ldflags "-w -s" ./

# zig
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
echo "building pi..."

GOARM=5 \
GOOS=linux \
GOARCH=arm \
CGO_ENABLED=1
CC="zig cc -v -target arm-linux-musleabihf" \
go build -o output/${ANAME}-arm -ldflags "-w -s" ./

#GOOS=linux GOARCH=arm GOARM=6 CGO_ENABLED=1 CC=arm-linux-gnu-gcc go build --tags "libsqlite3 linux" -v -o output/gifibtest-arm -ldflags="-w -s -extld=$CC"

#--- works with "github.com/glebarez/sqlite" driver
# pure go sqlite
#GOOS=linux GOARCH=arm GOARM=5 go build -o output/${ANAME}-arm -ldflags "-w -s" ./
