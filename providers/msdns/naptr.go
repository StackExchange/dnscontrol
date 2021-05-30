package msdns

// NAPTR records are not supported by the PowerShell module.
// Until this bug is fixed we use old-school commands instead.

import (
	"bytes"
	"fmt"

	"github.com/StackExchange/dnscontrol/v3/models"
)

func generatePSCCreateNaptr(domain string, rec *models.RecordConfig) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, `$zoneName = "%s"\n`, domain)
	fmt.Fprintf(&b, `$rrName = "%s"\n`, rec.Name)
	fmt.Fprintf(&b, `$Order       = %d\n`, rec.NaptrOrder)
	fmt.Fprintf(&b, `$Preference  = %d\n`, rec.NaptrPreference)
	fmt.Fprintf(&b, `$Flags       = "%s"\n`, rec.NaptrFlags)
	fmt.Fprintf(&b, `$Service     = "%s"\n`, rec.NaptrService)
	fmt.Fprintf(&b, `$Regex       = "%s"\n`, rec.NaptrRegexp)
	fmt.Fprintf(&b, `$Replacement = '%s'\n`, rec.GetTargetField())
	fmt.Fprintf(&b, `dnscmd /recordadd $zoneName $rrName naptr $Order $Preference $Flags $Service $Regex $Replacement\n`)
	return b.String()
}
