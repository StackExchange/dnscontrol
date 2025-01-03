package models

import (
	"fmt"
	"maps"
	"strconv"

	"github.com/StackExchange/dnscontrol/v4/pkg/fieldtypes"
)

// CNAME is the fields needed to store a DNS record of type CNAME
type CNAME struct {
	Target string
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

// PopulateARaw updates rc to be an A record with contents from rawfields, meta and origin.
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

// PopulateAFields updates rc to be an A record with contents from typed data, meta, and origin.
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
	rc.Display = rc.Comparable

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

// GetAStrings returns rc.Fields as individual strings.
func (rc *RecordConfig) GetAStrings() string {
	n := rc.AsA()
	return n.A.String()
}

//// MX

// MX is the fields needed to store a DNS record of type MX
type MX struct {
	Preference uint16
	Mx         string
}

// PopulateMXRaw updates rc to be an MX record with contents from rawfields, meta and origin.
func PopulateMXRaw(rc *RecordConfig, rawfields []string, meta map[string]string, origin string) error {
	var err error

	// Error checking

	if len(rawfields) <= 2 {
		return fmt.Errorf("rtype MX wants %d field(s), found %d: %+v", 1, len(rawfields)-1, rawfields[1:])
	}

	// Convert each rawfield.

	rc.SetLabel3(rawfields[0], rc.SubDomain, origin) // Label

	var preference uint16
	if preference, err = fieldtypes.ParseUint16(rawfields[1]); err != nil {
		return err
	}

	var mx string
	if mx, err = fieldtypes.ParseHostnameDot(rawfields[2], "", origin); err != nil {
		return err
	}

	return rc.PopulateMXFields(preference, mx, meta, origin)
}

// PopulateMXFields updates rc to be an MX record with contents from typed data, meta, and origin.
func (rc *RecordConfig) PopulateMXFields(preference uint16, mx string, meta map[string]string, origin string) error {
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
	rc.Comparable = fmt.Sprintf("%d %s", preference, mx)
	rc.Display = rc.Comparable

	return nil
}

// AsMX returns rc.Fields as an MX struct.
func (rc *RecordConfig) AsMX() *MX {
	return rc.Fields.(*MX)
}

// GetMXFields returns rc.Fields as individual typed values.
func (rc *RecordConfig) GetMXFields() (uint16, string) {
	n := rc.AsMX()
	return n.Preference, n.Mx
}

// GetMXStrings returns rc.Fields as individual strings.
func (rc *RecordConfig) GetMXStrings() [2]string {
	n := rc.AsMX()
	return [2]string{strconv.Itoa(int(n.Preference)), n.Mx}
}

//// SRV

// SRV is the fields needed to store a DNS record of type SRV
type SRV struct {
	Priority uint16 `json:"priority"`
	Weight   uint16 `json:"weight"`
	Port     uint16 `json:"port"`
	Target   string `json:"target"`
}

// PopulateSRVRaw updates rc to be an SRV record with contents from rawfields, meta and origin.
func PopulateSRVRaw(rc *RecordConfig, rawfields []string, meta map[string]string, origin string) error {
	var err error

	// Error checking

	if len(rawfields) <= 4 {
		return fmt.Errorf("rtype SRV wants %d field(s), found %d: %+v", 1, len(rawfields)-1, rawfields[1:])
	}

	// Convert each rawfield.

	rc.SetLabel3(rawfields[0], rc.SubDomain, origin) // Label

	var priority uint16
	if priority, err = fieldtypes.ParseUint16(rawfields[1]); err != nil {
		return err
	}
	var weight uint16
	if weight, err = fieldtypes.ParseUint16(rawfields[2]); err != nil {
		return err
	}
	var port uint16
	if port, err = fieldtypes.ParseUint16(rawfields[3]); err != nil {
		return err
	}
	var target string
	if target, err = fieldtypes.ParseHostnameDot(rawfields[4], "", origin); err != nil {
		return err
	}

	return rc.PopulateSRVFields(priority, weight, port, target, meta, origin)
}

// PopulateSRVFields updates rc to be an SRV record with contents from typed data, meta, and origin.
func (rc *RecordConfig) PopulateSRVFields(priority, weight, port uint16, target string, meta map[string]string, origin string) error {
	// Create the struct if needed.
	if rc.Fields == nil {
		rc.Fields = &SRV{}
	}

	// Process each field:

	n := rc.Fields.(*SRV)
	n.Priority = priority
	n.Weight = weight
	n.Port = port
	n.Target = target

	rc.SrvPriority = uint16(priority) // Legacy
	rc.SrvWeight = uint16(weight)     // Legacy
	rc.SrvPort = uint16(port)         // Legacy
	rc.SetTarget(string(target))      // Legacy

	// Update the RecordConfig:
	maps.Copy(rc.Metadata, meta) // Add the metadata
	rc.Comparable = fmt.Sprintf("%d %d %d %s", priority, weight, port, target)
	rc.Display = rc.Comparable

	return nil
}

// AsSRV returns rc.Fields as an SRV struct.
func (rc *RecordConfig) AsSRV() *SRV {
	return rc.Fields.(*SRV)
}

// GetSRVFields returns rc.Fields as individual typed values.
func (rc *RecordConfig) GetSRVFields() (uint16, uint16, uint16, string) {
	n := rc.AsSRV()
	return n.Priority, n.Weight, n.Port, n.Target
}

// GetSRVStrings returns rc.Fields as individual strings.
func (rc *RecordConfig) GetSRVStrings() [4]string {
	n := rc.AsSRV()
	return [4]string{strconv.Itoa(int(n.Priority)), strconv.Itoa(int(n.Weight)), strconv.Itoa(int(n.Port)), n.Target}
}
