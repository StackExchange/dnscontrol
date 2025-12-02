#!/bin/sh

set -e

cd $(git rev-parse --show-toplevel)
cd pkg/js
find . -type f -name \*.ACTUAL -print -delete
go test -count=1 ./... || true
cd parse_tests
fmtjson *.json *.json.ACTUAL
for i in *.ACTUAL ; do f=$(basename $i .ACTUAL) ; mv $i $f ; done
