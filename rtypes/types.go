package rtypes

import (
	"fmt"
	"maps"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/rtypectl"
)

// NS is the fields needed to store a DNS record of type NS
type NS struct {
	Ns string
}

// CNAME is the fields needed to store a DNS record of type CNAME
type CNAME struct {
	Target string
}

// A is the fields needed to store a DNS record of type A
type A struct {
	A [4]byte
}

func init() {
	rtypectl.Register("A", rtypectl.RegisterOpts{FromRaw: PopulateFromRawA})
	rtypectl.Register("MX", rtypectl.RegisterOpts{FromRaw: PopulateFromRawMX})
	//fmt.Printf("DEBUG: REGISTERED A\n")
}

// PopulateFromRawA updates rc to be an A record with contents from origin, rawfields and meta.
func PopulateFromRawA(rc *models.RecordConfig, origin string, rawfields []string, meta map[string]string) error {
	var err error

	// Error checking

	if len(rawfields) <= 1 {
		return fmt.Errorf("rtype %q wants %d field(s), found %d: %+v", "A", 1, len(rawfields)-1, rawfields[1:])
	}

	// Create the struct.
	n := A{}

	// Process each rawfield:

	rc.SetLabel3(rawfields[0], rc.SubDomain, origin) // Label

	if n.A, err = rtypectl.ParseIPv4(rawfields[1]); err != nil {
		return err
	}

	// Update legacy fields.
	rc.SetTargetIP(n.A[:])

	// Update the RecordConfig:
	maps.Copy(rc.Metadata, meta) // Add the metadata
	rc.Comparable = fmt.Sprintf("%s", n.A)
	rc.Display = fmt.Sprintf("%s", n.A)

	return nil
}

// MX is the fields needed to store a DNS record of type MX
type MX struct {
	Preference uint16
	Mx         string
}

// PopulateFromRawMX updates rc to be an MX record with contents from origin, rawfields and meta.
func PopulateFromRawMX(rc *models.RecordConfig, origin string, rawfields []string, meta map[string]string) error {
	var err error

	// Error checking

	if len(rawfields) <= 2 {
		return fmt.Errorf("rtype %q wants %d field(s), found %d: %+v", "MX", 2, len(rawfields)-1, rawfields[1:])
	}

	// Create the struct.
	n := MX{}
	rc.Fields = n

	// Process each rawfield:

	rc.SetLabel3(rawfields[0], rc.SubDomain, origin) // Label

	if n.Preference, err = rtypectl.ParseUint16(rawfields[1]); err != nil {
		return err
	}

	if n.Mx, err = rtypectl.ParseDottedHost(rawfields[2]); err != nil {
		return err
	}

	// Update legacy fields.
	rc.SetTargetMX(n.Preference, n.Mx)

	// Update the RecordConfig:
	maps.Copy(rc.Metadata, meta) // Add the metadata
	rc.Comparable = fmt.Sprintf("%s %s", n.Preference, n.Mx)
	rc.Display = rc.Comparable

	return nil
}
