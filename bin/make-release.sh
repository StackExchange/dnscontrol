#!/bin/bash

echo -e

: "${1:?'Provide the version number as the first arg (v1.2.3, not 1.2.3).'}" ;

NEWVERSION="$1"
if [[ "$NEWVERSION" != v* ]]; then
  echo "Version should start with v: v1.2.3, not 1.2.3."
  exit 1
fi

git tag -d "$NEWVERSION"

PREVVERSION="$(git tag --list '[vV]*' --sort=v:refname |tail -1)"

echo '=========='
echo '== current version:' "$PREVVERSION"
echo '==     new version:' "$NEWVERSION"

git checkout -b release_"$NEWVERSION" || true

sed -i.bak -e 's/Version   = ".*"/Version   = "'"$NEWVERSION"'"/g' main.go

git commit -m'Release '"$NEWVERSION" main.go
git tag -f "$NEWVERSION"
git push --delete origin "$NEWVERSION"
git push origin tag "$NEWVERSION"

echo ======= Creating: draft-notes.txt
echo >draft-notes.txt '
This release includes many new providers (FILL IN), dozens
of bug fixes, and FILL IN.

Breaking changes:

* FILL IN

Major features:

* FILL IN

Provider-specific changes:

* FILL IN

Other changes and improvements:

* FILL IN


'
git log "$NEWVERSION"..."$PREVVERSION" >>draft-notes.txt

git push --set-upstream origin "release_$NEWVERSION"

echo "NEXT STEP:

1. Create a PR:
open \"https://github.com/StackExchange/dnscontrol/compare/master...release_$NEWVERSION\"

2. Edit draft-notes.txt into actual release notes.

3. Verify tests complete successfully.

4. Merge the PR when satisfied.

5. Promote the release.
"
