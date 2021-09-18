#! /bin/bash

#the below is supposed to handle both tags, branches when specified as argument:
PACKAGE_VERSION="${1:-v0.0.0}"
PACKAGE_VERSION="${PACKAGE_VERSION#v}"
PACKAGE_VERSION="${PACKAGE_VERSION##*/}"

rm -Rf .ci/build/ 2>/dev/null
mkdir -p .ci/build

#taken from build/build.go
MAIN_SHA="$(git rev-parse HEAD)"
MAIN_BUILDTIME="$(date +%s)"

#TODO: check whether to include armel/armhf builds for .deb/.rpm (NB we might need to map 'arm' to 'armXX' in this case)
for BUILD_OS_ARCH in darwin/amd64 darwin/arm64 freebsd/386 freebsd/arm freebsd/amd64 linux/386 linux/amd64 linux/arm64 windows/amd64; do
    BUILD_OS="${BUILD_OS_ARCH%%/*}"
    BUILD_ARCH1="${BUILD_OS_ARCH##*/}"
    BUILD_ARCH2="${BUILD_ARCH1}"
    [[ "${BUILD_ARCH2}" == "386" ]] && BUILD_ARCH2="i386"
    BUILD_ARCH3="${BUILD_ARCH2}"
    [[ "${BUILD_ARCH3}" == "arm64" ]] && BUILD_ARCH3="aarch64"
    BUILD_OPTS=""
    [[ "${BUILD_OS}" == "linux" ]] && BUILD_OPTS="${BUILD_OPTS} CGO_ENABLED=0"
    SUFFIX=""
    [[ "${BUILD_OS}" == "windows" ]] && SUFFIX=".exe"
    go clean
    echo "**** Executing 'env${BUILD_OPTS} GOOS=\"${BUILD_OS}\" GOARCH=\"${BUILD_ARCH1}\" go build -mod vendor -ldflags=\"-s -w -X main.SHA=\"${MAIN_SHA}\" -X main.BuildTime=${MAIN_BUILDTIME}\"'"
    # shellcheck disable=SC2086
    env${BUILD_OPTS} GOOS="${BUILD_OS}" GOARCH="${BUILD_ARCH1}" go build -mod vendor -ldflags="-s -w -X main.SHA=\"${MAIN_SHA}\" -X main.BuildTime=${MAIN_BUILDTIME}"
    if [[ -f "dnscontrol${SUFFIX}" ]]; then
        if [[ "${BUILD_OS}" == "linux" ]]; then
            if type fpm 2>/dev/null 1>&2; then
                # create rpm, deb archives using fpm (if available)
                rm -Rf ./*.deb ./*.rpm usr/ 2>/dev/null
                mkdir -p usr/bin usr/share/doc/dnscontrol
                cp -a dnscontrol usr/bin/
                cp -a LICENSE usr/share/doc/dnscontrol
                fpm -n dnscontrol -t deb -s dir -a "${BUILD_ARCH2}" -v "${PACKAGE_VERSION}" --license "The MIT License (MIT)" --url "https://dnscontrol.org/" --description "DNSControl: Infrastructure as Code for DNS Zones" usr/
                fpm -n dnscontrol -t rpm -s dir -a "${BUILD_ARCH3}" -v "${PACKAGE_VERSION}" --license "The MIT License (MIT)" --url "https://dnscontrol.org/" --description "DNSControl: Infrastructure as Code for DNS Zones" usr/
                mv ./*.deb ./*.rpm .ci/build/ 2>/dev/null
            fi
        elif [[ "${BUILD_OS}" == "freebsd" ]]; then
            if type fpm 2>/dev/null 1>&2; then
                echo "FIXME: fpm -n dnscontrol -t freebsd -s dir -a ..."
            fi
        fi
        # create zip archives containing LICENSE *and* binary (TODO: maybe include additional documentation/examples?)
        zip -X -9 -o ".ci/build/dnscontrol_${PACKAGE_VERSION}_${BUILD_OS}-${BUILD_ARCH1}.zip" LICENSE "dnscontrol${SUFFIX}"
    fi
done

echo "----------"
ls -l .ci/build/*
echo "----------"
