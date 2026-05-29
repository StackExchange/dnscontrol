#!bin/bash

DEST=$1 ; shift

DEST_UC=$( echo "${DEST}" | tr a-z A-Z )
DEST_LC=$( echo "${DEST}" | tr A-Z a-z )

sed -e 's@ALIAS@'${DEST_UC}'@g'  				< alias.go       > ${DEST_LC}.go
sed -e 's@ALIAS@'${DEST_UC}'@g' -e 's@Alias@'${DEST}'@g'	< alias_test.go  > ${DEST_LC}_test.go
sed -e 's@ALIAS@'${DEST_UC}'@g'   				< rdata/alias.go > rdata/${DEST_LC}.go

echo 'Remember to fix:'
echo 'models/fixhack.go'
echo 'integrationTest/integration_test.go'
num=$(echo 1 + $(grep -h  'const Type' *.go | awk '{ print $NF }'  |sort | tail -1) | bc)
echo "const Type"${DEST_UC}" = $num"
