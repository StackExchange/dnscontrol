#! /bin/bash
if [[ "${1,,}" == "--debug" || "${DRONE_BUILD_DEBUG,,}" == "true" ]]; then shift; set -x; env; fi
set -e

#### setup

APTUPDATE=n

# the DockerHub 'golang' image does not have sudo, but commands are already executed by root
SUDO=""
if type sudo 2>/dev/null 1>&2; then
    SUDO="sudo"
fi

# syntax: install_pkg [<differing-name-of-package>:]<name-of-binary>
function install_pkg() {
    if ! type "${1#*:}" 2>/dev/null 1>&2; then
        if [[ "$APTUPDATE" == "n" ]]; then
            ${SUDO} apt-get -qq update
            APTUPDATE=y
        fi
        ${SUDO} apt-get -q install -y "${1%:*}"
    fi
}

# fpm (https://github.com/jordansissel/fpm) [, ruby]
# (1) ruby is part of the GitHub 'ubuntu-latest' image, but not part of the DockerHub 'golang' image
# (2) fpm isn't part of the forementioned images, but can be installed using gem (comes with ruby)
# note that fpm depends on rpmbuild (comes with rpm) in order to create .rpm archives
# note that fpm depends on xz (comes with xz-utils) in order to create .txz archives
if ! type fpm 2>/dev/null 1>&2; then
    install_pkg ruby:gem
    if type gem 2>/dev/null 1>&2; then
        ${SUDO} gem install fpm
    fi
fi

# xz[, xz-utils], rpmbuild[, rpm], tar, zip
for PKG in xz-utils:xz rpm:rpmbuild tar zip; do
    install_pkg $PKG
done


#### additional checks (if any)

go version
go mod vendor
