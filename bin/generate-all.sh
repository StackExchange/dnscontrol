#!/bin/sh

go fmt ./...

if [[ -x node_modules/.bin/prettier ]]; then
  node_modules/.bin/prettier --write pkg/js/helpers.js
fi

# JSON
bin/fmtjson $(find . -path ./.vscode -prune -o -type f -name "*.json" ! -name "package-lock.json" -print)

# dnsconfig.js-compatible files:
for i in pkg/js/parse_tests/*.js ; do dnscontrol fmt -i $i -o $i ; done

go generate ./...

go mod tidy
