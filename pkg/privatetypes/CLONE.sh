#!bin/bash

SRC="Adguardhome_AAAA_Passthrough"
SRC_UC=$( echo "${SRC}" | tr a-z A-Z )
SRC_LC=$( echo "${SRC}" | tr A-Z a-z )

DEST=$1 ; shift

DEST_UC=$( echo "${DEST}" | tr a-z A-Z )
DEST_LC=$( echo "${DEST}" | tr A-Z a-z )

sed -e 's@'${SRC_UC}'@'${DEST_UC}'@g'  				   < "${SRC_LC}.go"       > "${DEST_LC}.go"
sed -e 's@'${SRC_UC}'@'${DEST_UC}'@g' -e 's@'${SRC}'@'${DEST}'@g'  < "${SRC_LC}_test.go"  > "${DEST_LC}_test.go"
sed -e 's@'${SRC_UC}'@'${DEST_UC}'@g'   			   < "rdata/${SRC_LC}.go" > "rdata/${DEST_LC}.go"

echo 'Remember to fix:'
echo 'models/fixhack.go'
echo 'integrationTest/integration_test.go'
num=$(echo 1 + $(grep -h  'const Type' *.go | awk '{ print $NF }'  |sort | tail -1) | bc)
echo "const Type"${DEST_UC}" = $num"
