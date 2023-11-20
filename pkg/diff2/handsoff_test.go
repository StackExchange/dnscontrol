package diff2

import (
	"fmt"
	"strings"
	"testing"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/js"
	"github.com/miekg/dns"
	testifyrequire "github.com/stretchr/testify/require"
)

// parseZoneContents is copied verbatium from providers/bind/bindProvider.go
// because import cycles and... tests shouldn't depend on huge modules.
func parseZoneContents(content string, zoneName string, zonefileName string) (models.Records, error) {
	zp := dns.NewZoneParser(strings.NewReader(content), zoneName, zonefileName)

	foundRecords := models.Records{}
	for rr, ok := zp.Next(); ok; rr, ok = zp.Next() {
		rec, err := models.RRtoRCTxtBug(rr, zoneName)
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

func showRecs(recs models.Records) string {
	result := ""
	for _, rec := range recs {
		result += (rec.GetLabel() +
			" " + rec.Type +
			" " + rec.GetTargetCombined() +
			"\n")
	}
	return result
}

func handsoffHelper(t *testing.T, existingZone, desiredJs string, noPurge bool, resultWanted string) {
	t.Helper()

	existing, err := parseZoneContents(existingZone, "f.com", "no_file_name")
	if err != nil {
		panic(err)
	}

	dnsconfig, err := js.ExecuteJavascriptString([]byte(desiredJs), false, nil)
	if err != nil {
		panic(err)
	}
	dc := dnsconfig.FindDomain("f.com")
	desired := dc.Records
	absences := dc.EnsureAbsent
	unmanagedConfigs := dc.Unmanaged
	// BUG(tlim): For some reason ExecuteJavascriptString() isn't setting the NameFQDN on records.
	//            This fixes up the records. It is a crass workaround. We should find the real
	//            cause and fix it.
	for i, j := range desired {
		desired[i].SetLabel(j.GetLabel(), "f.com")
	}
	for i, j := range absences {
		absences[i].SetLabel(j.GetLabel(), "f.com")
	}

	ignored, purged := processIgnoreAndNoPurge(
		"f.com",
		existing, desired,
		absences,
		unmanagedConfigs,
		noPurge,
	)

	ignoredRecs := showRecs(ignored)
	purgedRecs := showRecs(purged)
	resultActual := "IGNORED:\n" + ignoredRecs + "FOREIGN:\n" + purgedRecs
	resultWanted = strings.TrimSpace(resultWanted) + "\n"
	resultActual = strings.TrimSpace(resultActual) + "\n"

	existingTxt := showRecs(existing)
	desiredTxt := showRecs(desired)
	debugTxt := "EXISTING:\n" + existingTxt + "DESIRED:\n" + desiredTxt

	if resultWanted != resultActual {
		testifyrequire.Equal(t,
			resultWanted,
			resultActual,
			"GOT =\n```\n%s```\nWANT=\n```%s```\nINPUTS=\n```\n%s\n```\n",
			resultActual,
			resultWanted,
			debugTxt)
	}
}

func Test_purge_empty(t *testing.T) {
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
IGNORED:
FOREIGN:
	`)
}

func Test_purge_1(t *testing.T) {
	existingZone := `
foo1 IN A 1.1.1.1
foo2 IN A 2.2.2.2
foo3 IN A 2.2.2.2
`
	desiredJs := `
D("f.com", "none",
	A("foo1", "1.1.1.1"),
	A("foo2", "2.2.2.2"),
{})
`
	handsoffHelper(t, existingZone, desiredJs, false, `
IGNORED:
FOREIGN:
	`)
}

func Test_nopurge_1(t *testing.T) {
	existingZone := `
foo1 IN A 1.1.1.1
foo2 IN A 2.2.2.2
foo3 IN A 3.3.3.3
`
	desiredJs := `
D("f.com", "none",
	A("foo1", "1.1.1.1"),
	A("foo2", "2.2.2.2"),
{})
`
	handsoffHelper(t, existingZone, desiredJs, true, `
IGNORED:
FOREIGN:
foo3 A 3.3.3.3
	`)
}

func Test_absent_1(t *testing.T) {
	existingZone := `
foo1 IN A 1.1.1.1
foo2 IN A 2.2.2.2
foo3 IN A 3.3.3.3
`
	desiredJs := `
D("f.com", "none",
	A("foo1", "1.1.1.1"),
	A("foo2", "2.2.2.2"),
	A("foo3", "3.3.3.3", ENSURE_ABSENT_REC()),
{})
`
	handsoffHelper(t, existingZone, desiredJs, false, `
IGNORED:
FOREIGN:
	`)
}

func Test_ignore_lab(t *testing.T) {
	existingZone := `
foo1 IN A 1.1.1.1
foo2 IN A 2.2.2.2
foo3 IN A 3.3.3.3
foo3 IN MX 10 mymx.example.com.
`
	desiredJs := `
D("f.com", "none",
	A("foo1", "1.1.1.1"),
	A("foo2", "2.2.2.2"),
	IGNORE_NAME("foo3"),
{})
`
	handsoffHelper(t, existingZone, desiredJs, false, `
IGNORED:
foo3 A 3.3.3.3
foo3 MX 10 mymx.example.com.
FOREIGN:
	`)
}

func Test_ignore_labAndType(t *testing.T) {
	existingZone := `
foo1 IN A 1.1.1.1
foo2 IN A 2.2.2.2
foo3 IN A 3.3.3.3
foo3 IN MX 10 mymx.example.com.
`
	desiredJs := `
D("f.com", "none",
	A("foo1", "1.1.1.1"),
	A("foo2", "2.2.2.2"),
	A("foo3", "3.3.3.3"),
	IGNORE_NAME("foo3", "MX"),
{})
`
	handsoffHelper(t, existingZone, desiredJs, false, `
IGNORED:
foo3 MX 10 mymx.example.com.
FOREIGN:
	`)
}

func Test_ignore_target(t *testing.T) {
	existingZone := `
foo1 IN A 1.1.1.1
foo2 IN A 2.2.2.2
_2222222222222222.cr IN CNAME _333333.nnn.acm-validations.aws.
`
	desiredJs := `
D("f.com", "none",
	A("foo1", "1.1.1.1"),
	A("foo2", "2.2.2.2"),
	MX("foo3", 10, "mymx.example.com."),
	IGNORE_TARGET('**.acm-validations.aws.', 'CNAME'),
{})
`
	handsoffHelper(t, existingZone, desiredJs, false, `
IGNORED:
_2222222222222222.cr CNAME _333333.nnn.acm-validations.aws.
FOREIGN:
	`)
}
