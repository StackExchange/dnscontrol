#!/bin/sh

# fix_js_parse_tests.sh -- Fix up the pkg/js/parse_tests "want" files.
#
# DO NOT use this without carefully checking the output. You could
# accidentally codify bad test data and commit it to the repo.
#
# Useful bash/zsh alias:
# alias fixjsparse='"$(git rev-parse --show-toplevel)/bin/fix_js_parse_tests.sh"'

set -e

cd $(git rev-parse --show-toplevel)
cd pkg/js
find . -type f -name \*.ACTUAL -print -delete
go test -count=1 ./... || true
cd parse_tests
fmtjson *.json *.json.ACTUAL
for i in *.ACTUAL ; do f=$(basename $i .ACTUAL) ; mv $i $f ; done
