 param (
    [string]$SHA = ""
 )

if ($SHA -eq ""){
    $SHA = (git rev-parse HEAD) | Out-String
    $SHA = $SHA.Replace([System.Environment]::NewLine,"")
}


$PKG = "github.com/StackExchange/dnscontrol"
$DATE = [int][double]::Parse((Get-Date -UFormat %s))
$FLAGS="-mod=readonly -s -w -X main.SHA=$SHA -X main.BuildTime=$DATE"
Write-Host $FLAGS

$OrigGOOS = $env:GOOS

$env:GO111MODULE = "on"

Write-Host 'Building Linux'
$env:GOOS = "linux"
go build -o dnscontrol-Linux -ldflags "$FLAGS" $PKG

Write-Host 'Building Windows'
$env:GOOS = "windows"
go build -o dnscontrol.exe -ldflags "$FLAGS" $PKG

Write-Host 'Building Darwin'
$env:GOOS = "darwin"
go build -o dnscontrol-Darwin -ldflags "$FLAGS" $PKG

$env:GOOS = $OrigGOOS

#No compression if building on windows
<#
if [ "$COMPRESS" = "1" ]
then
    echo 'Compressing executables'
    upx dnscontrol.exe
    upx dnscontrol-Linux
    upx dnscontrol-Darwin
fi
#>