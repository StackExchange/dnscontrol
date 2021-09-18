#! /bin/bash

#### setup

# fpm (https://github.com/jordansissel/fpm) is currently not part of the 'ubuntu-latest' image;
# but because the latter contains the ruby package, we can install it using gem:
if ! type fpm 2>/dev/null 1>&2; then
    if type gem 2>/dev/null 1>&2; then
        sudo gem install fpm
        # note that fpm depends on rpmbuild in order to create .rpm archives
        # (which is included in the above image as part of the rpm package)
    fi
fi

go version
go mod vendor


#### additional checks (if any)
