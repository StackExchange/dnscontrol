# Run build.sh in a docker container that is guaranteed to have all of the appropriate tools we need
PKG=github.com/StackExchange/dnscontrol
docker run -v `pwd`:/go/src/$PKG -w /go/src/github.com/StackExchange/dnscontrol captncraig/golang-build /bin/sh /go/src/github.com/StackExchange/dnscontrol/build.sh
