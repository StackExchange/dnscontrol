package models

import (
	"fmt"
	"strconv"

	"github.com/StackExchange/dnscontrol/v4/pkg/fieldtypes"
)

func init() {
	MustRegisterType("A", RegisterOpts{PopulateFromRaw: PopulateFromRawA})
	MustRegisterType("MX", RegisterOpts{PopulateFromRaw: PopulateFromRawMX})
	MustRegisterType("SRV", RegisterOpts{PopulateFromRaw: PopulateFromRawSRV})
	MustRegisterType("CNAME", RegisterOpts{PopulateFromRaw: PopulateFromRawCNAME})
	MustRegisterType("CF_SINGLE_REDIRECT", RegisterOpts{PopulateFromRaw: PopulateFromRawCFSINGLEREDIRECT})
	MustRegisterType("CAA", RegisterOpts{PopulateFromRaw: PopulateFromRawCAA})
}

// RecordType is a constraint for DNS records.
type RecordType interface {
	A | MX | SRV | CNAME | CFSINGLEREDIRECT | CAA
}

// ImportFromLegacy copies the legacy fields (MxPreference, SrvPort, etc.) to
// the .Fields structure.  It is the reverse of Seal*().
func (rc *RecordConfig) ImportFromLegacy(origin string) error {

	if IsTypeLegacy(rc.Type) {
		// Nothing to convert!
		return nil
	}

	switch rc.Type {
	case "A":
		ip, err := fieldtypes.ParseIPv4(rc.target)
		if err != nil {
			return err
		}
		return RecordUpdateFields(rc, A{A: ip}, nil)
	case "MX":
		return RecordUpdateFields(rc,
			MX{Preference: rc.MxPreference, Mx: rc.target},
			nil,
		)
	case "SRV":
		return RecordUpdateFields(rc,
			SRV{Priority: rc.SrvPriority, Weight: rc.SrvWeight, Port: rc.SrvPort, Target: rc.target},
			nil,
		)
	case "CNAME":
		return RecordUpdateFields(rc,
			CNAME{Target: rc.target},
			nil,
		)
	case "CAA":
		return RecordUpdateFields(rc,
			CAA{Flag: rc.CaaFlag, Tag: rc.CaaTag, Value: rc.target},
			nil,
		)
	}
	panic("Should not happen")
}

func (rc *RecordConfig) Seal() error {
	rc.Type = GetTypeName(rc.Fields)

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
	case "CNAME":
		f := rc.Fields.(*CNAME)
		rc.target = f.Target

		rc.Comparable = f.Target
	case "CF_SINGLE_REDIRECT":
		f := rc.Fields.(*CFSINGLEREDIRECT)
		rc.target = f.SRDisplay

		rc.Comparable = fmt.Sprintf("%q %d %q %q", f.SRName, f.Code, f.SRWhen, f.SRThen)
	case "CAA":
		f := rc.Fields.(*CAA)
		rc.CaaFlag = f.Flag
		rc.CaaTag = f.Tag
		rc.target = f.Value

		rc.Comparable = fmt.Sprintf("%d %s %q", f.Flag, f.Tag, f.Value)
	default:
		return fmt.Errorf("unknown (Seal) rtype %q", rc.Type)
	}
	rc.Display = rc.Comparable

	return nil
}

// GetTargetField returns the target. There may be other fields, but they are
// not included. For example, the .MxPreference field of an MX record isn't included.
func (rc *RecordConfig) GetTargetField() string {
	switch rc.Type { // #rtype_variations
	case "A":
		return rc.AsA().A.String()
	case "MX":
		return rc.AsMX().Mx
	case "SRV":
		return rc.AsSRV().Target
	case "CNAME":
		return rc.AsCNAME().Target
	case "CFSINGLEREDIRECT":
		return rc.AsCFSINGLEREDIRECT().SRDisplay
	case "CAA":
		return rc.AsCAA().Value
	}
	return rc.target
}

//// A

// A is the fields needed to store a DNS record of type A.
type A struct {
	A fieldtypes.IPv4 `dns:"a"`
}

func ParseA(rawfields []string, origin string) (A, error) {

	// Error checking
	if errorCheckFieldCount(rawfields, 1) {
		return A{}, fmt.Errorf("rtype A wants %d field(s), found %d: %+v", 1, len(rawfields)-1, rawfields[1:])
	}
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

	// First rawfield is the label.
	if err := rc.SetLabel3(rawfields[0], rc.SubDomain, origin); err != nil {
		return err
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

// GetFieldsA returns rc.Fields as individual typed values.
func (rc *RecordConfig) GetFieldsA() fieldtypes.IPv4 {
	n := rc.AsA()
	return n.A
}

// GetFieldsAsStringsA returns rc.Fields as individual strings.
func (rc *RecordConfig) GetFieldsAsStringsA() [1]string {
	n := rc.AsA()
	return [1]string{n.A.String()}
}

//// MX

// MX is the fields needed to store a DNS record of type MX.
type MX struct {
	Preference uint16
	Mx         string `dns:"cdomain-name"`
}

func ParseMX(rawfields []string, origin string) (MX, error) {

	// Error checking
	if errorCheckFieldCount(rawfields, 2) {
		return MX{}, fmt.Errorf("rtype MX wants %d field(s), found %d: %+v", 2, len(rawfields)-1, rawfields[1:])
	}
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

	// First rawfield is the label.
	if err := rc.SetLabel3(rawfields[0], rc.SubDomain, origin); err != nil {
		return err
	}

	// Parse the remaining fields.
	rdata, err := ParseMX(rawfields[1:], origin)
	if err != nil {
		return err
	}

	return RecordUpdateFields(rc, rdata, meta)
}

// AsMX returns rc.Fields as an MX struct.
func (rc *RecordConfig) AsMX() *MX {
	return rc.Fields.(*MX)
}

// GetFieldsMX returns rc.Fields as individual typed values.
func (rc *RecordConfig) GetFieldsMX() (uint16, string) {
	n := rc.AsMX()
	return n.Preference, n.Mx
}

// GetFieldsAsStringsMX returns rc.Fields as individual strings.
func (rc *RecordConfig) GetFieldsAsStringsMX() [2]string {
	n := rc.AsMX()
	return [2]string{strconv.Itoa(int(n.Preference)), n.Mx}
}

//// SRV

// SRV is the fields needed to store a DNS record of type SRV.
type SRV struct {
	Priority uint16 `json:"priority"`
	Weight   uint16 `json:"weight"`
	Port     uint16 `json:"port"`
	Target   string `json:"target" dns:"domain-name"`
}

func ParseSRV(rawfields []string, origin string) (SRV, error) {

	// Error checking
	if errorCheckFieldCount(rawfields, 4) {
		return SRV{}, fmt.Errorf("rtype SRV wants %d field(s), found %d: %+v", 4, len(rawfields)-1, rawfields[1:])
	}
	var priority uint16
	var weight uint16
	var port uint16
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

	// First rawfield is the label.
	if err := rc.SetLabel3(rawfields[0], rc.SubDomain, origin); err != nil {
		return err
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

// GetFieldsSRV returns rc.Fields as individual typed values.
func (rc *RecordConfig) GetFieldsSRV() (uint16, uint16, uint16, string) {
	n := rc.AsSRV()
	return n.Priority, n.Weight, n.Port, n.Target
}

// GetFieldsAsStringsSRV returns rc.Fields as individual strings.
func (rc *RecordConfig) GetFieldsAsStringsSRV() [4]string {
	n := rc.AsSRV()
	return [4]string{strconv.Itoa(int(n.Priority)), strconv.Itoa(int(n.Weight)), strconv.Itoa(int(n.Port)), n.Target}
}

//// CNAME

// CNAME is the fields needed to store a DNS record of type CNAME.
type CNAME struct {
	Target string `dns:"cdomain-name"`
}

func ParseCNAME(rawfields []string, origin string) (CNAME, error) {

	// Error checking
	if errorCheckFieldCount(rawfields, 1) {
		return CNAME{}, fmt.Errorf("rtype CNAME wants %d field(s), found %d: %+v", 1, len(rawfields)-1, rawfields[1:])
	}
	var target string
	var err error
	if target, err = fieldtypes.ParseHostnameDot(rawfields[0], "", origin); err != nil {
		return CNAME{}, err
	}

	return CNAME{Target: target}, nil
}

// PopulateFromRawCNAME updates rc to be an CNAME record with contents from rawfields, meta and origin.
func PopulateFromRawCNAME(rc *RecordConfig, rawfields []string, meta map[string]string, origin string) error {
	rc.Type = "CNAME"

	// First rawfield is the label.
	if err := rc.SetLabel3(rawfields[0], rc.SubDomain, origin); err != nil {
		return err
	}

	// Parse the remaining fields.
	rdata, err := ParseCNAME(rawfields[1:], origin)
	if err != nil {
		return err
	}

	return RecordUpdateFields(rc, rdata, meta)
}

// AsCNAME returns rc.Fields as an CNAME struct.
func (rc *RecordConfig) AsCNAME() *CNAME {
	return rc.Fields.(*CNAME)
}

// GetFieldsCNAME returns rc.Fields as individual typed values.
func (rc *RecordConfig) GetFieldsCNAME() string {
	n := rc.AsCNAME()
	return n.Target
}

// GetFieldsAsStringsCNAME returns rc.Fields as individual strings.
func (rc *RecordConfig) GetFieldsAsStringsCNAME() [1]string {
	n := rc.AsCNAME()
	return [1]string{n.Target}
}

//// CFSINGLEREDIRECT

// CFSINGLEREDIRECT is the fields needed to store a DNS record of type CFSINGLEREDIRECT.
type CFSINGLEREDIRECT struct {
	SRName           string `json:"sr_name,omitempty" dnscontrol:"_,label,anyascii"`
	Code             uint16 `json:"code,omitempty" dnscontrol:"_,redirectcode"`
	SRWhen           string `json:"sr_when,omitempty" dnscontrol:"_,anyascii"`
	SRThen           string `json:"sr_then,omitempty" dnscontrol:"_,anyascii"`
	SRRRulesetID     string `json:"sr_rulesetid,omitempty" dnscontrol:"_,noraw,noinput"`
	SRRRulesetRuleID string `json:"sr_rulesetruleid,omitempty" dnscontrol:"_,noraw,noinput"`
	SRDisplay        string `json:"sr_display,omitempty" dnscontrol:"_,srdisplay,noraw,noinput"`
	PRWhen           string `json:"pr_when,omitempty" dnscontrol:"_,noraw,parsereturnunknowable,noinput"`
	PRThen           string `json:"pr_then,omitempty" dnscontrol:"_,noraw,parsereturnunknowable,noinput"`
	PRPriority       int    `json:"pr_priority,omitempty" dnscontrol:"_,noraw,noinput"`
	PRDisplay        string `json:"pr_display" dnscontrol:"_,noraw,parsereturnunknowable,noinput"`
}

func ParseCFSINGLEREDIRECT(rawfields []string, origin string) (CFSINGLEREDIRECT, error) {

	// Error checking
	if errorCheckFieldCount(rawfields, 4) {
		return CFSINGLEREDIRECT{}, fmt.Errorf("rtype CFSINGLEREDIRECT wants %d field(s), found %d: %+v", 4, len(rawfields)-1, rawfields[1:])
	}
	var srname string
	var code uint16
	var srwhen string
	var srthen string
	var err error
	if srname, err = fieldtypes.ParseStringTrimmed(rawfields[0]); err != nil {
		return CFSINGLEREDIRECT{}, err
	}
	if code, err = fieldtypes.ParseUint16(rawfields[1]); err != nil {
		return CFSINGLEREDIRECT{}, err
	}
	if srwhen, err = fieldtypes.ParseStringTrimmed(rawfields[2]); err != nil {
		return CFSINGLEREDIRECT{}, err
	}
	if srthen, err = fieldtypes.ParseStringTrimmed(rawfields[3]); err != nil {
		return CFSINGLEREDIRECT{}, err
	}

	return CFSINGLEREDIRECT{SRName: srname, Code: code, SRWhen: srwhen, SRThen: srthen, SRDisplay: cfSingleRedirecttargetFromRaw(srname, code, srwhen, srthen), PRWhen: "UNKNOWABLE", PRThen: "UNKNOWABLE", PRDisplay: "UNKNOWABLE"}, nil
}

// PopulateFromRawCFSINGLEREDIRECT updates rc to be an CFSINGLEREDIRECT record with contents from rawfields, meta and origin.
func PopulateFromRawCFSINGLEREDIRECT(rc *RecordConfig, rawfields []string, meta map[string]string, origin string) error {
	rc.Type = "CF_SINGLE_REDIRECT"
	rc.TTL = 1

	// First rawfield is the label.
	if err := rc.SetLabel3(rawfields[0], rc.SubDomain, origin); err != nil {
		return err
	}

	// Parse the remaining fields.
	rdata, err := ParseCFSINGLEREDIRECT(rawfields, origin)
	if err != nil {
		return err
	}

	return RecordUpdateFields(rc, rdata, meta)
}

// AsCFSINGLEREDIRECT returns rc.Fields as an CFSINGLEREDIRECT struct.
func (rc *RecordConfig) AsCFSINGLEREDIRECT() *CFSINGLEREDIRECT {
	return rc.Fields.(*CFSINGLEREDIRECT)
}

// GetFieldsCFSINGLEREDIRECT returns rc.Fields as individual typed values.
func (rc *RecordConfig) GetFieldsCFSINGLEREDIRECT() (string, uint16, string, string) {
	n := rc.AsCFSINGLEREDIRECT()
	return n.SRName, n.Code, n.SRWhen, n.SRThen
}

// GetFieldsAsStringsCFSINGLEREDIRECT returns rc.Fields as individual strings.
func (rc *RecordConfig) GetFieldsAsStringsCFSINGLEREDIRECT() [4]string {
	n := rc.AsCFSINGLEREDIRECT()
	return [4]string{n.SRName, strconv.Itoa(int(n.Code)), n.SRWhen, n.SRThen}
}

//// CAA

// CAA is the fields needed to store a DNS record of type CAA.
type CAA struct {
	Flag  uint8
	Tag   string
	Value string `dnscontrol:"_,anyascii"`
}

func ParseCAA(rawfields []string, origin string) (CAA, error) {

	// Error checking
	if errorCheckFieldCount(rawfields, 3) {
		return CAA{}, fmt.Errorf("rtype CAA wants %d field(s), found %d: %+v", 3, len(rawfields)-1, rawfields[1:])
	}
	var flag uint8
	var tag string
	var value string
	var err error
	if flag, err = fieldtypes.ParseUint8(rawfields[0]); err != nil {
		return CAA{}, err
	}
	if tag, err = fieldtypes.ParseStringTrimmed(rawfields[1]); err != nil {
		return CAA{}, err
	}
	if value, err = fieldtypes.ParseStringTrimmed(rawfields[2]); err != nil {
		return CAA{}, err
	}

	return CAA{Flag: flag, Tag: tag, Value: value}, nil
}

// PopulateFromRawCAA updates rc to be an CAA record with contents from rawfields, meta and origin.
func PopulateFromRawCAA(rc *RecordConfig, rawfields []string, meta map[string]string, origin string) error {
	rc.Type = "CAA"

	// First rawfield is the label.
	if err := rc.SetLabel3(rawfields[0], rc.SubDomain, origin); err != nil {
		return err
	}

	// Parse the remaining fields.
	rdata, err := ParseCAA(rawfields[1:], origin)
	if err != nil {
		return err
	}

	return RecordUpdateFields(rc, rdata, meta)
}

// AsCAA returns rc.Fields as an CAA struct.
func (rc *RecordConfig) AsCAA() *CAA {
	return rc.Fields.(*CAA)
}

// GetFieldsCAA returns rc.Fields as individual typed values.
func (rc *RecordConfig) GetFieldsCAA() (uint8, string, string) {
	n := rc.AsCAA()
	return n.Flag, n.Tag, n.Value
}

// GetFieldsAsStringsCAA returns rc.Fields as individual strings.
func (rc *RecordConfig) GetFieldsAsStringsCAA() [3]string {
	n := rc.AsCAA()
	return [3]string{strconv.Itoa(int(n.Flag)), n.Tag, n.Value}
}
