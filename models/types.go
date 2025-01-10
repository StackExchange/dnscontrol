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

func init() {
	RegisterType("A", RegisterOpts{PopulateFromRaw: PopulateFromRawA})
	RegisterType("MX", RegisterOpts{PopulateFromRaw: PopulateFromRawMX})
	RegisterType("SRV", RegisterOpts{PopulateFromRaw: PopulateFromRawSRV})
}

//// A

// A is the fields needed to store a DNS record of type A
type A struct {
	A fieldtypes.IPv4
}

// NewFromRawA creates a new RecordConfig of type A from rawfields, meta, and origin.
func NewFromRawA(rawfields []string, meta map[string]string, origin string) (*RecordConfig, error) {
	rc := &RecordConfig{
		Metadata: map[string]string{},
	}
	if err := PopulateFromRawA(rc, rawfields, meta, origin); err != nil {
		return nil, err
	}
	return rc, nil
}

// PopulateFromRawA updates rc to be an A record with contents from rawfields, meta and origin.
func PopulateFromRawA(rc *RecordConfig, rawfields []string, meta map[string]string, origin string) error {
	var err error

	// Error checking

	if len(rawfields) <= 1 {
		return fmt.Errorf("rtype A wants %d field(s), found %d: %+v", 1, len(rawfields)-1, rawfields[1:])
	}

	// Convert each rawfield.

	if origin != "" { //  If we don't know the origin, don't muck with the label.
		rc.SetLabel3(rawfields[0], rc.SubDomain, origin) // Label
	}

	var a fieldtypes.IPv4
	if a, err = fieldtypes.ParseIPv4(rawfields[1]); err != nil { // A
		return err
	}

	return rc.PopulateFromFieldsA(a, meta, origin)
}

// PopulateFromFieldsA updates rc to be an A record with contents from typed data, meta, and origin.
func (rc *RecordConfig) PopulateFromFieldsA(a fieldtypes.IPv4, meta map[string]string, origin string) error {
	// Create the struct if needed.
	if rc.Fields == nil {
		rc.Fields = &A{}
	}

	// Update the RecordConfig:
	maps.Copy(rc.Metadata, meta) // Add the metadata

	// Process each field:

	f := rc.Fields.(*A)
	f.A = a

	return rc.SealA()
}

func (rc *RecordConfig) SealA() error {
	if rc.Type == "" {
		rc.Type = "A"
	}
	if rc.Type != "A" {
		panic("assertion failed: SealA called when .Type is not A")
	}

	f := rc.Fields.(*A)

	// Pre-compute useful things
	rc.Comparable = f.A.String()
	rc.Display = rc.Comparable

	// Copy the fields to the legacy fields:
	rc.target = f.A.String()

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

// NewFromRawMX creates a new RecordConfig of type MX from rawfields, meta, and origin.
func NewFromRawMX(rawfields []string, meta map[string]string, origin string) (*RecordConfig, error) {
	rc := &RecordConfig{}
	if err := PopulateFromRawMX(rc, rawfields, meta, origin); err != nil {
		return nil, err
	}
	return rc, nil
}

// PopulateFromRawMX updates rc to be an MX record with contents from rawfields, meta and origin.
func PopulateFromRawMX(rc *RecordConfig, rawfields []string, meta map[string]string, origin string) error {
	var err error

	// Error checking

	if len(rawfields) <= 2 {
		return fmt.Errorf("rtype MX wants %d field(s), found %d: %+v", 1, len(rawfields)-1, rawfields[1:])
	}

	// Convert each rawfield.

	if origin != "" { //  If we don't know the origin, don't muck with the label.
		rc.SetLabel3(rawfields[0], rc.SubDomain, origin) // Label
	}

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

	// Update the RecordConfig:
	maps.Copy(rc.Metadata, meta) // Add the metadata

	// Process each field:

	n := rc.Fields.(*MX)
	n.Preference = preference
	n.Mx = mx

	return rc.SealMX()
}

// SealMX updates rc to be an MX record with contents from typed data, meta, and origin.
func (rc *RecordConfig) SealMX() error {
	if rc.Type == "" {
		rc.Type = "MX"
	}
	if rc.Type != "MX" {
		panic("assertion failed: SealMX called when .Type is not MX")
	}

	f := rc.Fields.(*MX)

	// Pre-compute useful things
	rc.Comparable = fmt.Sprintf("%d %s", f.Preference, f.Mx)
	rc.Display = rc.Comparable

	// Copy the fields to the legacy fields:
	rc.MxPreference = f.Preference
	rc.target = f.Mx

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

// NewFromRawSRV creates a new RecordConfig of type SRV from rawfields, meta, and origin.
func NewFromRawSRV(rawfields []string, meta map[string]string, origin string) (*RecordConfig, error) {
	rc := &RecordConfig{}
	if err := PopulateFromRawSRV(rc, rawfields, meta, origin); err != nil {
		return nil, err
	}
	return rc, nil
}

// PopulateFromRawSRV updates rc to be an SRV record with contents from rawfields, meta and origin.
func PopulateFromRawSRV(rc *RecordConfig, rawfields []string, meta map[string]string, origin string) error {
	var err error

	// Error checking

	if len(rawfields) <= 4 {
		return fmt.Errorf("rtype SRV wants %d field(s), found %d: %+v", 4, len(rawfields)-1, rawfields[1:])
	}

	// Convert each rawfield.

	if origin != "" { //  If we don't know the origin, don't muck with the label.
		rc.SetLabel3(rawfields[0], rc.SubDomain, origin) // Label
	}

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

	return rc.PopulateFromFieldsSRV(priority, weight, port, target, meta, origin)
}

// PopulateFromFieldsSRV updates rc to be an SRV record with contents from typed data, meta, and origin.
func (rc *RecordConfig) PopulateFromFieldsSRV(priority, weight, port uint16, target string, meta map[string]string, origin string) error {
	// Create the struct if needed.
	if rc.Fields == nil {
		rc.Fields = &SRV{}
	}

	// Update the RecordConfig:
	maps.Copy(rc.Metadata, meta) // Add the metadata

	// Process each field:

	n := rc.Fields.(*SRV)
	n.Priority = priority
	n.Weight = weight
	n.Port = port
	n.Target = target

	return rc.SealSRV()
}

func (rc *RecordConfig) SealSRV() error {
	if rc.Type == "" {
		rc.Type = "SRV"
	}
	if rc.Type != "SRV" {
		panic("assertion failed: SetTargetSRV called when .Type is not SRV")
	}

	f := rc.Fields.(*SRV)

	// Pre-compute useful things
	rc.Comparable = fmt.Sprintf("%d %d %d %s", f.Priority, f.Weight, f.Port, f.Target)
	rc.Display = rc.Comparable

	// Copy the fields to the legacy fields:
	rc.SrvPriority = f.Priority
	rc.SrvWeight = f.Weight
	rc.SrvPort = f.Port
	rc.target = f.Target

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
