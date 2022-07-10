echo "building linux..."
ANAME="goapp"
go build -o output/${ANAME} -ldflags "-w -s" ./


echo "building windows..."
# cgo sqlite
GOOS=windows GOARCH=386 CGO_ENABLED=1 CC=i686-w64-mingw32-g++ CC=i686-w64-mingw32-gcc go build -o output/${ANAME}386.exe -ldflags "-w -s" ./

# pure go sqlite
#GOOS=windows GOARCH=386  go build -o output/gofibtest386.exe -ldflags "-w -s" ./


echo "building pi..."
GOOS=linux GOARCH=arm GOARM=5 CGO_ENABLED=1 CC=arm-linux-gnu-gcc go build -o output/${ANAME}-arm -ldflags "-w -s" ./
#GOOS=linux GOARCH=arm GOARM=6 CGO_ENABLED=1 CC=arm-linux-gnu-gcc go build --tags "libsqlite3 linux" -v -o output/gifibtest-arm -ldflags="-w -s -extld=$CC"

#--- works with "github.com/glebarez/sqlite" driver
# pure go sqlite
#GOOS=linux GOARCH=arm GOARM=5 go build -o output/${ANAME}-arm -ldflags "-w -s" ./
