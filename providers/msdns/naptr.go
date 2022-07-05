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

func generatePSCreateNaptr(dnsServerName, domain string, rec *models.RecordConfig) string {

	var computername string
	if dnsServerName != "" {
		computername = escapePS(dnsServerName) + " "
	}

	var b bytes.Buffer
	fmt.Fprintf(&b, `$zoneName    = %s ; `, escapePS(domain))
	fmt.Fprintf(&b, `$rrName      = %s ; `, escapePS(rec.Name))
	fmt.Fprintf(&b, `$Order       = %d ; `, rec.NaptrOrder)
	fmt.Fprintf(&b, `$Preference  = %d ; `, rec.NaptrPreference)
	fmt.Fprintf(&b, `$Flags       = %s ; `, escapePS(rec.NaptrFlags))
	fmt.Fprintf(&b, `$Service     = %s ; `, escapePS(rec.NaptrService))
	fmt.Fprintf(&b, `$Regex       = %s ; `, escapePS(rec.NaptrRegexp))
	fmt.Fprintf(&b, `$Replacement = %s ; `, escapePS(rec.GetTargetField()))
	fmt.Fprintf(&b, `dnscmd %s/recordadd $zoneName $rrName %d naptr $Order $Preference $Flags $Service $Regex $Replacement ; `, computername, rec.TTL)
	return b.String()
}

func generatePSDeleteNaptr(dnsServerName, domain string, rec *models.RecordConfig) string {
	target := rec.GetTargetField()
	if target == "" {
		target = "."
	}

	var computername string
	if dnsServerName != "" {
		computername = escapePS(dnsServerName) + " "
	}

	var b bytes.Buffer
	fmt.Fprintf(&b, `$zoneName    = %s ; `, escapePS(domain))
	fmt.Fprintf(&b, `$rrName      = %s ; `, escapePS(rec.Name))
	fmt.Fprintf(&b, `$Order       = %d ; `, rec.NaptrOrder)
	fmt.Fprintf(&b, `$Preference  = %d ; `, rec.NaptrPreference)
	fmt.Fprintf(&b, `$Flags       = %s ; `, escapePS(rec.NaptrFlags))
	fmt.Fprintf(&b, `$Service     = %s ; `, escapePS(rec.NaptrService))
	fmt.Fprintf(&b, `$Regex       = %s ; `, escapePS(rec.NaptrRegexp))
	fmt.Fprintf(&b, `$Replacement = %s ; `, escapePS(target))
	fmt.Fprintf(&b, `dnscmd %s/recorddelete $zoneName $rrName naptr $Order $Preference $Flags $Service $Regex $Replacement /f ; `, computername)
	return b.String()
}

// decoding

func decodeRecordDataNaptr(s string) models.RecordConfig {
	// These strings look like this:
	// C8AFB0B30153075349502B4432540474657374165F7369702E5F7463702E6578616D706C652E6F72672E
	// The first 2 groups of 16 bits (4 hex digits) are uinet16.
	// The rest are 4 length-prefixed strings.
	//	The string should be entirely consumed.
	rc := models.RecordConfig{}

	s, rc.NaptrOrder = eatUint16(s)
	s, rc.NaptrPreference = eatUint16(s)
	s, rc.NaptrFlags = eatString(s)
	s, rc.NaptrService = eatString(s)
	s, rc.NaptrRegexp = eatString(s)
	s, targ := eatString(s)
	rc.SetTarget(targ)

	// At this point we should have consumed the entire string.
	if s != "" {
		ctx.Log.Printf("WARNING: REMAINDER:=%q\n", s)
	}

	return rc
}

// eatUint16 consumes the first 16 bits of the string, returns it as a
// uint16, and returns the remaining bytes of the string.
func eatUint16(s string) (string, uint16) {
	value, err := strconv.ParseUint(s[2:4]+s[0:2], 16, 16)
	if err != nil {
		log.Fatal(err)
	}
	return s[4:], uint16(value)
}

// eatString consumes an encoded string (8-bit length byte, then the string).
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

	return s[last:], string(ret)
}
