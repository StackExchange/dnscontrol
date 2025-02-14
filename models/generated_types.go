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
}

// RecordType is a constraint for DNS records.
type RecordType interface {
	A | MX | SRV | CNAME | CFSINGLEREDIRECT
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
	fmt.Printf("DEBUG: ParseCNAME(%q, %q)\n", rawfields, origin)

	// Error checking
	if errorCheckFieldCount(rawfields, 1) {
		return CNAME{}, fmt.Errorf("rtype CNAME wants %d field(s), found %d: %+v", 1, len(rawfields)-1, rawfields[1:])
	}
	var target string
	var err error
	if target, err = fieldtypes.ParseHostnameDot(rawfields[0], "", origin); err != nil {
		return CNAME{}, err
	}
	fmt.Printf("DEBUG: CNAME target: %s %s\n", rawfields[0], target)

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
	SRName           string `json:"sr_name,omitempty"`
	Code             uint16 `json:"code,omitempty" dnscontrol:"_,redirectcode"`
	SRWhen           string `json:"sr_when,omitempty"`
	SRThen           string `json:"sr_then,omitempty"`
	SRRRulesetID     string `json:"sr_rulesetid,omitempty" dnscontrol:"_,noraw,noparsereturn"`
	SRRRulesetRuleID string `json:"sr_rulesetruleid,omitempty" dnscontrol:"_,noraw,noparsereturn"`
	SRDisplay        string `json:"sr_display,omitempty" dnscontrol:"_,srdisplay,noraw,noparsereturn"`
	PRWhen           string `json:"pr_when,omitempty" dnscontrol:"_,noraw,parsereturnunknowable"`
	PRThen           string `json:"pr_then,omitempty" dnscontrol:"_,noraw,parsereturnunknowable"`
	PRPriority       int    `json:"pr_priority,omitempty" dnscontrol:"_,noraw,noparsereturn"`
	PRDisplay        string `json:"pr_display" dnscontrol:"_,noraw,parsereturnunknowable,noparsereturn"`
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
