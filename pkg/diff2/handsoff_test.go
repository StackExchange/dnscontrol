package diff2

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/js"
	"github.com/StackExchange/dnscontrol/v3/pkg/prettyzone"
	"github.com/miekg/dns"
	testifyrequire "github.com/stretchr/testify/require"
)

// ParseZoneContents is copied verbatium from providers/bind/bindProvider.go
// because import cycles and... tests shouldn't depend on huge modules.
func ParseZoneContents(content string, zoneName string, zonefileName string) (models.Records, error) {
	zp := dns.NewZoneParser(strings.NewReader(content), zoneName, zonefileName)

	foundRecords := models.Records{}
	for rr, ok := zp.Next(); ok; rr, ok = zp.Next() {
		rec, err := models.RRtoRC(rr, zoneName)
		if err != nil {
			return nil, err
		}
		foundRecords = append(foundRecords, &rec)
	}

	if err := zp.Err(); err != nil {
		return nil, fmt.Errorf("error while parsing '%v': %w", zonefileName, err)
	}
	return foundRecords, nil
}

func handsoffHelper(t *testing.T, existingZone, desiredJs string, noPurge bool, wantedZone string) {
	t.Helper()

	existing, err := ParseZoneContents(existingZone, "f.com", "no_file_name")
	if err != nil {
		panic(err)
	}

	dnsconfig, err := js.ExecuteJavascriptString([]byte(desiredJs), false, nil)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	err = prettyzone.WriteZoneFileRC(&buf, dnsconfig.FindDomain("f.com").Records, "f.com", 300, nil)
	if err != nil {
		t.Fatal(err)
	}
	actualZone := strings.TrimSpace(buf.String())

	wantedZone = strings.TrimSpace(wantedZone)

	ignored, purged := ignoreOrNoPurge(existing, existing, ensureAbsent, unmanagedConfigs, noPurge)

	if wantedZone != actualZone {
		testifyrequire.Equal(t, wantedZone, actualZone, "EXPECTING =\n```\n%s```", actualZone)

	}
}

func Test_normal(t *testing.T) {
	existingZone := `
foo1 IN A 1.1.1.1
foo2 IN A 2.2.2.2
`
	desiredJs := `
D("f.com", "none",
	A("foo1", "1.1.1.1"),
	A("foo2", "2.2.2.2"),
{})
`

	handsoffHelper(t, existingZone, desiredJs, false, `
$TTL 300
foo1             IN A     1.1.1.1
foo2             IN A     2.2.2.2
`)

	handsoffHelper(t, existingZone, desiredJs, true, `
foo1 IN A 1.1.1.1
foo2 IN A 2.2.2.2
`)

}
