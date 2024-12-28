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

// MX is the fields needed to store a DNS record of type MX
type MX struct {
	Preference uint16
	Mx         string
}

// A is the fields needed to store a DNS record of type A
type A struct {
	A [4]byte
}

func init() {
	rtypectl.Register("A", rtypectl.RegisterOpts{FromRaw: PopulateFromRawA})
	fmt.Printf("DEBUG: REGISTERED A\n")
}

func PopulateFromRawA(rc *models.RecordConfig, origin string, rawfields []any, meta map[string]string) error {

	// Error checking

	if len(rawfields) <= 1 {
		return fmt.Errorf("rtype %q wants %d field(s), found %d: %+v", "A", 1, len(rawfields)-1, rawfields[1:])
	}

	// Create the struct

	n := A{}
	var err error
	if n.A, err = rtypectl.ParseIPv4(rawfields[1]); err != nil {
		return err
	}
	rc.SetTargetIP(n.A[:])

	// Update the record:
	//rc.SetLabel(rawfields[0].(string), origin) // Label
	maps.Copy(rc.Metadata, meta) // Add the metadata
	rc.Comparable = fmt.Sprintf("%s", n.A)
	rc.Display = fmt.Sprintf("%s", n.A)

	return nil
}
