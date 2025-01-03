package models

import (
	"fmt"
	"maps"

	"github.com/StackExchange/dnscontrol/v4/pkg/fieldtypes"
)

// CNAME is the fields needed to store a DNS record of type CNAME
type CNAME struct {
	Target fieldtypes.HostnameDot
}

/*

GetAFields()
GetAStrings()
GetA()

PopulateARaw()
PopulateAStrings() (not needed at this time)
PopulateAFields()

*/

func init() {
	RegisterType("A", RegisterOpts{FromRaw: PopulateARaw})
	RegisterType("MX", RegisterOpts{FromRaw: PopulateMXRaw})
	//fmt.Printf("DEBUG: REGISTERED A\n")
}

//// A

// A is the fields needed to store a DNS record of type A
type A struct {
	A fieldtypes.IPv4
}

// PopulateFromRawA updates rc to be an A record with contents from rawfields, meta and origin.
func PopulateARaw(rc *RecordConfig, rawfields []string, meta map[string]string, origin string) error {
	var err error

	// Error checking

	if len(rawfields) <= 1 {
		return fmt.Errorf("rtype A wants %d field(s), found %d: %+v", 1, len(rawfields)-1, rawfields[1:])
	}

	// Convert each rawfield.

	rc.SetLabel3(rawfields[0], rc.SubDomain, origin) // Label

	var a fieldtypes.IPv4
	if a, err = fieldtypes.ParseIPv4(rawfields[1]); err != nil { // A
		return err
	}

	return rc.PopulateAFields(a, meta, origin)
}

// PopulateFromRawA updates rc to be an A record with contents from typed data, meta, and origin.
func (rc *RecordConfig) PopulateAFields(a fieldtypes.IPv4, meta map[string]string, origin string) error {
	// Create the struct if needed.
	if rc.Fields == nil {
		rc.Fields = &A{}
	}

	// Process each field:

	n := rc.Fields.(*A)

	n.A = a
	rc.SetTargetIP(n.A[:]) // Legacy

	// Update the RecordConfig:
	maps.Copy(rc.Metadata, meta) // Add the metadata
	rc.Comparable = fmt.Sprintf("%s", n.A)
	rc.Display = fmt.Sprintf("%s", n.A)

	return nil
}

// AsA returns rc.Fields as an A struct.
func (rc *RecordConfig) AsA() *A {
	return rc.Fields.(*A)
}

// GetAFields returns rc.Fields as individual typed values.
func (rc *RecordConfig) GetAFields() fieldtypes.IPv4 {
	n := rc.AsA()
	return n.A
}

// GetAFields returns rc.Fields as individual strings.
func (rc *RecordConfig) GetAStrings() string {
	n := rc.AsA()
	return n.A.String()
}

//// MX

// MX is the fields needed to store a DNS record of type MX
type MX struct {
	Preference fieldtypes.Uint16
	Mx         fieldtypes.HostnameDot
}

// PopulateFromRawMX updates rc to be an MX record with contents from rawfields, meta and origin.
func PopulateMXRaw(rc *RecordConfig, rawfields []string, meta map[string]string, origin string) error {
	var err error

	// Error checking

	if len(rawfields) <= 2 {
		return fmt.Errorf("rtype MX wants %d field(s), found %d: %+v", 1, len(rawfields)-1, rawfields[1:])
	}

	// Convert each rawfield.

	rc.SetLabel3(rawfields[0], rc.SubDomain, origin) // Label

	var preference fieldtypes.Uint16
	if preference, err = fieldtypes.ParseUint16(rawfields[1]); err != nil {
		return err
	}

	var mx fieldtypes.HostnameDot
	if mx, err = fieldtypes.ParseHostnameDot(rawfields[2], "", origin); err != nil {
		return err
	}

	return rc.PopulateMXFields(preference, mx, meta, origin)
}

// PopulateFromRawMX updates rc to be an MX record with contents from typed data, meta, and origin.
func (rc *RecordConfig) PopulateMXFields(preference fieldtypes.Uint16, mx fieldtypes.HostnameDot, meta map[string]string, origin string) error {
	// Create the struct if needed.
	if rc.Fields == nil {
		rc.Fields = &MX{}
	}

	// Process each field:

	n := rc.Fields.(*MX)
	n.Preference = preference
	n.Mx = mx
	rc.MxPreference = uint16(preference) // Legacy
	rc.SetTarget(string(mx))             // Legacy

	// Update the RecordConfig:
	maps.Copy(rc.Metadata, meta) // Add the metadata
	rc.Comparable = fmt.Sprintf("%s", n.Mx)
	rc.Display = fmt.Sprintf("%s", n.Mx)

	return nil
}

// AsMX returns rc.Fields as an MX struct.
func (rc *RecordConfig) AsMX() *MX {
	return rc.Fields.(*MX)
}

// GetMXFields returns rc.Fields as individual typed values.
func (rc *RecordConfig) GetMXFields() (fieldtypes.Uint16, fieldtypes.HostnameDot) {
	n := rc.AsMX()
	return n.Preference, n.Mx
}

// GetMXFields returns rc.Fields as individual strings.
func (rc *RecordConfig) GetMXStrings() [2]string {
	n := rc.AsMX()
	return [2]string{n.Preference.String(), n.Mx.String()}
}
