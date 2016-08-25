
PKG=github.com/StackExchange/dnscontrol
FLAGS="-s -w"

echo 'Building Linux'
go build -o dnscontrol-Linux -ldflags "$FLAGS" $PKG

echo 'Building Windows'
export GOOS=windows
go build -o dnscontrol.exe -ldflags "$FLAGS" $PKG

echo 'Building Darwin'
export GOOS=darwin
go build -o dnscontrol-Darwin -ldflags "$FLAGS" $PKG

if [ "$COMPRESS" = "1" ]
then
    echo 'Compressing executables'
    upx dnscontrol.exe
    upx dnscontrol-Linux
    upx dnscontrol-Darwin
fi