#! /bin/bash
FPM_LOGLEVEL=error
if [[ "${1,,}" == "--debug" || "${DRONE_BUILD_DEBUG,,}" == "true" ]]; then shift; set -x; FPM_LOGLEVEL=debug; fi
set -e

# the below is supposed to handle both tags, branches when specified as argument/environment variable:
PACKAGE_VERSION="$1"
PACKAGE_VERSION="${PACKAGE_VERSION:-${DRONE_SEMVER}}"
PACKAGE_VERSION="${PACKAGE_VERSION:-${DRONE_SOURCE_BRANCH}}"
PACKAGE_VERSION="${PACKAGE_VERSION:-v0.0.0}"
PACKAGE_VERSION="${PACKAGE_VERSION#v}"
PACKAGE_VERSION="${PACKAGE_VERSION##*/}"

# metadata
FPM_OPTIONS=(
    --name dnscontrol
    --version "${PACKAGE_VERSION}"
    --license "The MIT License (MIT)"
    --url "https://dnscontrol.org/"
    --description "DNSControl: Infrastructure as Code for DNS Zones"
)
# list of files to be packaged and their respective locations (path names are subject to os-/archive-specific adjustment)
# TODO: maybe include additional documentation/examples?
DNSCONTROL_FILES=(
    dnscontrol=/bin/
    LICENSE=/share/doc/dnscontrol/
)

rm -Rf .ci/build/ 2>/dev/null
mkdir -p .ci/build

# taken from build/build.go
MAIN_SHA="$(git rev-parse HEAD)"
MAIN_BUILDTIME="$(date +%s)"

# TODO: check whether to include armel/armhf builds for .deb/.rpm/.txz (NB we might need to map 'arm' to 'armXX' in this case)
for BUILD_OS_ARCH in freebsd/386 freebsd/amd64 freebsd/arm64 linux/386 linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64; do
    BUILD_OS="${BUILD_OS_ARCH%%/*}"
    BUILD_ARCH1="${BUILD_OS_ARCH##*/}"
    BUILD_ARCH2="${BUILD_ARCH1}"
    [[ "${BUILD_ARCH2}" == "386" ]] && BUILD_ARCH2="i386"
    BUILD_ARCH3="${BUILD_ARCH2}"
    [[ "${BUILD_ARCH3}" == "arm64" ]] && BUILD_ARCH3="aarch64"
    BUILD_ARCH4="${BUILD_ARCH3}"
    [[ "${BUILD_ARCH4}" == "amd64" ]] && BUILD_ARCH4="x86_64"
    BUILD_OPTS=""
    [[ "${BUILD_OS}" == "linux" ]] && BUILD_OPTS="${BUILD_OPTS} CGO_ENABLED=0"
    SUFFIX=""
    [[ "${BUILD_OS}" == "windows" ]] && SUFFIX=".exe"
    go clean
    echo "**** Executing 'env${BUILD_OPTS} GOOS=\"${BUILD_OS}\" GOARCH=\"${BUILD_ARCH1}\" go build -mod vendor -ldflags=\"-s -w -X main.SHA=\"${MAIN_SHA}\" -X main.BuildTime=${MAIN_BUILDTIME}\"'"
    #shellcheck disable=SC2086
    env${BUILD_OPTS} GOOS="${BUILD_OS}" GOARCH="${BUILD_ARCH1}" go build -mod vendor -ldflags="-s -w -X main.SHA=\"${MAIN_SHA}\" -X main.BuildTime=${MAIN_BUILDTIME}"
    if [[ -f "dnscontrol${SUFFIX}" ]]; then
        if [[ "${BUILD_OS}" == "linux" ]]; then
            # create rpm, deb archives using fpm (if available)
            if type fpm 2>/dev/null 1>&2; then
                fpm -a "${BUILD_ARCH2}" --log "${FPM_LOGLEVEL}" -p .ci/build --prefix /usr -s dir -t deb "${FPM_OPTIONS[@]}" "${DNSCONTROL_FILES[@]}"
                fpm -a "${BUILD_ARCH4}" --log "${FPM_LOGLEVEL}" -p .ci/build --prefix /usr -s dir -t rpm "${FPM_OPTIONS[@]}" "${DNSCONTROL_FILES[@]}"
            fi
        elif [[ "${BUILD_OS}" == "freebsd" ]]; then
            # create txz archive using fpm (if available)
            if type fpm 2>/dev/null 1>&2; then
                rm -Rf ./*.txz 2>/dev/null
                fpm -a "${BUILD_ARCH3}" --log "${FPM_LOGLEVEL}" --prefix /usr/local -s dir -t freebsd "${FPM_OPTIONS[@]}" "${DNSCONTROL_FILES[@]}"
                TXZNAME="$(ls ./*.txz 2>/dev/null)"
                if [[ -n "${TXZNAME}" ]]; then
                    # FIXUP: fpm 3.13.1 (and older?) creates invalid txz archives lacking a leading '/' for non-metadata files
                    # see https://github.com/jordansissel/fpm/issues/1832
                    if tar -tf "${TXZNAME}" 2>/dev/null | grep -qE "^[a-z]"; then
                        FTMPDIR="$(mktemp -d -p .)"
                        if [[ -d "${FTMPDIR}" ]]; then
                            tar -C "${FTMPDIR}" -xf "${TXZNAME}"
                            #shellcheck disable=SC2046
                            tar -cJf "${TXZNAME}" $(find "${FTMPDIR}" -type f | sort) --transform "s|${FTMPDIR}||" --transform 's|/+|+|'
                            rm -Rf "./${FTMPDIR}" 2>/dev/null
                        fi
                    fi
                    mv "${TXZNAME}" ".ci/build/${TXZNAME/\.txz/_${BUILD_ARCH3}.txz}" 2>/dev/null
                fi
            fi
        fi
        # create zip archives containing ${DNSCONTROL_FILES[@]} with stripped paths and accounting for the executable's ${SUFFIX}
        DNSCONTROL_ZFILES=("${DNSCONTROL_FILES[@]}")
        #shellcheck disable=SC2068
        for idx in ${!DNSCONTROL_ZFILES[@]}; do
            BASENAME="${DNSCONTROL_ZFILES[$idx]}"
            BASENAME="${BASENAME%=*}"
            [[ "${BASENAME}" == "dnscontrol" ]] && BASENAME="dnscontrol${SUFFIX}"
            DNSCONTROL_ZFILES[$idx]="${BASENAME}"
        done
        zip -X -9 -o ".ci/build/dnscontrol_${PACKAGE_VERSION}_${BUILD_OS}-${BUILD_ARCH1}.zip" "${DNSCONTROL_ZFILES[@]}"
    fi
done

set +x
echo "===================="
ls -l .ci/build/*
echo "===================="
