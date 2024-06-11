#!/bin/bash

# build/checkowner.sh: Make sure that every provider is mentioned in
# OWNERS.  If any are missing (from the list of providers, or the
# OWNERS file) report them and

DIRS="$(mktemp /tmp/ownercheck.dirs.XXXXXXXXXX)" || { echo "Failed to create temp file"; exit 1; }
OWNR="$(mktemp /tmp/ownercheck.ownr.XXXXXXXXXX)" || { echo "Failed to create temp file"; exit 1; }
OWNS="$(mktemp /tmp/ownercheck.owns.XXXXXXXXXX)" || { echo "Failed to create temp file"; exit 1; }
MISS="$(mktemp /tmp/ownercheck.miss.XXXXXXXXXX)" || { echo "Failed to create temp file"; exit 1; }

# What directories are in the filesystem?
( cd providers/ && find * -type d -maxdepth 0 -print | grep -v '^_all' | sort >"$DIRS" )

# What directories are in the OWNERS file?
grep '^providers/' OWNERS | awk -F/ '{ print $2 }' | awk '{ print $1 }' >"$OWNR"
grep '^# providers/.*NEEDS VOLUNTEER' OWNERS | awk -F/ '{ print $2 }' | awk '{ print $1 }' >>"$OWNR"
sort <"$OWNR" >"$OWNS"

# Are they the same?
comm -3 "$DIRS" "$OWNS" >"$MISS"

# Report results:
if [[ -s "$MISS" ]]; then
  echo ======= ALERT: PROVIDERS MISSING FROM OWNERS or FROM THE PROVIDERS DIRECTORY:
  cat "$MISS"
  echo ======= 
  echo FAILURE!
  cleanup
  exit 1
else
  echo SUCCESS!
  cleanup
  exit 0
fi

function cleanup() {
  rm -f "$DIRS" "$OWNR" "$OWNS" "$MISS"
}
