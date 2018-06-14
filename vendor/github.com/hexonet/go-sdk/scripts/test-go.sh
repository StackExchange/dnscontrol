#!/usr/bin/env bash

set -euo pipefail

echo
echo "==> Running automated tests <=="
cd test
go test
cd .. || exit
exit
