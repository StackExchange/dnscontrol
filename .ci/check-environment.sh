#! /bin/bash

#### setup

# fpm (https://github.com/jordansissel/fpm) is currently not part of the 'ubuntu-latest' image;
# but because the latter contains ruby, we can install it using gem:
if ! type fpm 2>/dev/null 1>&1; then
    if type gem 2>/dev/null 1>&1; then
        sudo gem install fpm
    fi
fi

go version
go mod vendor


#### additional checks (if any)
