if [ ! -z $1 ] 
then 
    SHA=$1
else
    SHA=`git rev-parse HEAD`
fi


PKG=github.com/StackExchange/dnscontrol
DATE=`date +%s`
FLAGS="-s -w -X main.SHA=$SHA -X main.BuildTime=$DATE"
echo $FLAGS
set +e
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