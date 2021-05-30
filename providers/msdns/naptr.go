package msdns

// NAPTR records are not supported by the PowerShell module.
// Until this bug is fixed we use old-school commands instead.

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"

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

func decodeRecordDataNaptr(s string) models.RecordConfig {
	// C8AFB0B30153075349502B4432540474657374165F7369702E5F7463702E6578616D706C652E6F72672E
	rc := models.RecordConfig{}

	s, rc.NaptrOrder = eatUint16(s)
	s, rc.NaptrPreference = eatUint16(s)
	s, rc.NaptrFlags = eatString(s)
	s, rc.NaptrService = eatString(s)
	s, rc.NaptrRegexp = eatString(s)
	s, targ := eatString(s)
	rc.SetTarget(targ)
	if s != "" {
		fmt.Printf("WARNING: REMAINDER:=%q\n", s)
	}

	return rc
}

func eatUint16(s string) (string, uint16) {
	value, err := strconv.ParseUint(s[2:4]+s[0:2], 16, 64)
	if err != nil {
		log.Fatal(err)
	}
	return s[4:], uint16(value)
}

func eatString(s string) (string, string) {
	sl, err := strconv.ParseUint(s[:2], 16, 64)
	if err != nil {
		log.Fatal(err)
	}
	last := 2 + sl*2
	hexcoded := s[2:last]
	ret, err := hex.DecodeString(hexcoded)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Decoding: %s\n", s)
	fmt.Printf("      sl: %d\n", sl)
	fmt.Printf("    last: %d\n", last)
	fmt.Printf("     ret: %q\n", ret)
	return s[last:], string(ret)
}
