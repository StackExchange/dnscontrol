package models

import (
	"fmt"
	"maps"
	"strconv"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/pkg/fieldtypes"
)

/*

GetAFields()
GetAStrings()
GetA()

PopulateARaw()
PopulateAStrings() (not needed at this time)
PopulateAFields()

*/

/*

- Create from RawRecord
                models.CreateRecordFromRaw(rc, rawstrings, meta, origin) (error)
- Fully create an RC for Test purposes:
                models.MustCreateRecord[models.A](T, Meta, TTL, Origin) (*models.RecordConfig)
- Populate type=X from typed-fields
                models.RecordUpdate[T](T, meta, origin) (error)
- Populate type=X from strings
                rc.RecordUpdateFromStrings([]string, meta, origin) (error)

 -ParseA([]string) (A, error)
 -ParseMX([]string) (MX, error)
 -ParseSRV[]string) (SRV, error)

 - Get the fields:
 				models.GetFields[T]() (*T)
*/

// RecordType is a constraint for DNS records.
type RecordType interface {
	A | MX | SRV | CFSINGLEREDIRECT
}

func init() {
	RegisterType("A", RegisterOpts{PopulateFromRaw: PopulateFromRawA})
	RegisterType("MX", RegisterOpts{PopulateFromRaw: PopulateFromRawMX})
	RegisterType("SRV", RegisterOpts{PopulateFromRaw: PopulateFromRawSRV})
}

func RecordUpdateFields[T RecordType](rc *RecordConfig, rdata T, meta map[string]string) error {
	rc.Fields = &rdata

	// Update the RecordConfig:
	maps.Copy(rc.Metadata, meta) // Add the metadata

	return rc.Seal()
}

func (rc *RecordConfig) Seal() error {
	if rc.Type == "" {
		switch rc.Fields.(type) {
		case A:
			rc.Type = "A"
		case MX:
			rc.Type = "MX"
		case SRV:
			rc.Type = "SRV"
		case CFSINGLEREDIRECT:
			rc.Type = "CF_SINGLE_REDIRECT"
		}
	}

	// Copy the fields to the legacy fields:
	// Pre-compute useful things
	switch rc.Type {
	case "A":
		f := rc.Fields.(*A)
		rc.target = f.A.String()

		rc.Comparable = fmt.Sprintf("%d.%d.%d.%d", f.A[0], f.A[1], f.A[2], f.A[3])
	case "MX":
		f := rc.Fields.(*MX)
		rc.MxPreference = f.Preference
		rc.target = f.Mx
		rc.Comparable = fmt.Sprintf("%d %s", f.Preference, f.Mx)
	case "SRV":
		f := rc.Fields.(*SRV)
		rc.SrvPriority = f.Priority
		rc.SrvWeight = f.Weight
		rc.SrvPort = f.Port
		rc.target = f.Target
		rc.Comparable = fmt.Sprintf("%d %d %d %s", f.Priority, f.Weight, f.Port, f.Target)
	case "CF_SINGLE_REDIRECT":
		// Legacy fields have been eliminated.
		f := rc.Fields.(*CFSINGLEREDIRECT)
		rc.Comparable = fmt.Sprintf("%q %d %q %q", f.SRName, f.Code, f.SRWhen, f.SRThen)
	default:
		return fmt.Errorf("unknown (Seal) rtype %q", rc.Type)
	}
	rc.Display = rc.Comparable

	return nil
}

func MustCreateRecord[T RecordType](label string, rdata T, meta map[string]string, ttl uint32, origin string) *RecordConfig {
	rc := &RecordConfig{
		Type: strings.Split(fmt.Sprintf("%T", rdata), ".")[1],
		TTL:  ttl,
	}
	rc.SetLabel3(label, "", origin) // Label
	if err := RecordUpdateFields(rc, rdata, meta); err != nil {
		panic(err)
	}
	return rc
}

//// A

// A is the fields needed to store a DNS record of type A
type A struct {
	A fieldtypes.IPv4
}

func ParseA(rawfields []string, origin string) (A, error) {
	var a fieldtypes.IPv4
	var err error
	if a, err = fieldtypes.ParseIPv4(rawfields[0]); err != nil {
		return A{}, err
	}
	return A{A: a}, nil
}

// PopulateFromRawA updates rc to be an A record with contents from rawfields, meta and origin.
func PopulateFromRawA(rc *RecordConfig, rawfields []string, meta map[string]string, origin string) error {
	rc.Type = "A"

	// Error checking
	if len(rawfields) <= 1 {
		return fmt.Errorf("rtype A wants %d field(s), found %d: %+v", 1, len(rawfields)-1, rawfields[1:])
	}

	// First rawfield is the label.
	if origin != "" { //  If we don't know the origin, don't muck with the label.
		rc.SetLabel3(rawfields[0], rc.SubDomain, origin) // Label
	}

	// Parse the remaining fields.
	rdata, err := ParseA(rawfields[1:], origin)
	if err != nil {
		return err
	}

	return RecordUpdateFields(rc, rdata, meta)
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

func ParseMX(rawfields []string, origin string) (MX, error) {
	var preference uint16
	var mx string
	var err error
	if preference, err = fieldtypes.ParseUint16(rawfields[0]); err != nil {
		return MX{}, err
	}
	if mx, err = fieldtypes.ParseHostnameDot(rawfields[1], "", origin); err != nil {
		return MX{}, err
	}
	return MX{Preference: preference, Mx: mx}, nil
}

// PopulateFromRawMX updates rc to be an MX record with contents from rawfields, meta and origin.
func PopulateFromRawMX(rc *RecordConfig, rawfields []string, meta map[string]string, origin string) error {
	rc.Type = "MX"
	var err error

	// Error checking
	if len(rawfields) <= 2 {
		return fmt.Errorf("rtype MX wants %d field(s), found %d: %+v", 1, len(rawfields)-1, rawfields[1:])
	}

	// First rawfield is the label.
	if origin != "" { //  If we don't know the origin, don't muck with the label.
		rc.SetLabel3(rawfields[0], rc.SubDomain, origin) // Label
	}

	// Parse the remaining fields.
	rdata, err := ParseMX(rawfields[1:], origin)
	if err != nil {
		return err
	}

	//return rc.PopulateFieldsMX(preference, mx, meta, origin)
	return RecordUpdateFields(rc, rdata, meta)
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

func ParseSRV(rawfields []string, origin string) (SRV, error) {
	var priority, weight, port uint16
	var target string
	var err error
	if priority, err = fieldtypes.ParseUint16(rawfields[0]); err != nil {
		return SRV{}, err
	}
	if weight, err = fieldtypes.ParseUint16(rawfields[1]); err != nil {
		return SRV{}, err
	}
	if port, err = fieldtypes.ParseUint16(rawfields[2]); err != nil {
		return SRV{}, err
	}
	if target, err = fieldtypes.ParseHostnameDot(rawfields[3], "", origin); err != nil {
		return SRV{}, err
	}
	return SRV{Priority: priority, Weight: weight, Port: port, Target: target}, nil
}

// PopulateFromRawSRV updates rc to be an SRV record with contents from rawfields, meta and origin.
func PopulateFromRawSRV(rc *RecordConfig, rawfields []string, meta map[string]string, origin string) error {
	rc.Type = "SRV"
	var err error

	// Error checking
	if len(rawfields) <= 4 {
		return fmt.Errorf("rtype SRV wants %d field(s), found %d: %+v", 4, len(rawfields)-1, rawfields[1:])
	}

	// First rawfield is the label.
	if origin != "" { //  If we don't know the origin, don't muck with the label.
		rc.SetLabel3(rawfields[0], rc.SubDomain, origin) // Label
	}

	// Parse the remaining fields.
	rdata, err := ParseSRV(rawfields[1:], origin)
	if err != nil {
		return err
	}

	return RecordUpdateFields(rc, rdata, meta)

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
