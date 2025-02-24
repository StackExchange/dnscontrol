package models

import (
	"fmt"
	"strconv"

	"github.com/StackExchange/dnscontrol/v4/pkg/fieldtypes"
	"github.com/qdm12/reprint"
)

func init() {
	MustRegisterType("A", RegisterOpts{PopulateFromRaw: PopulateFromRawA})
	MustRegisterType("MX", RegisterOpts{PopulateFromRaw: PopulateFromRawMX})
	MustRegisterType("SRV", RegisterOpts{PopulateFromRaw: PopulateFromRawSRV})
	MustRegisterType("CNAME", RegisterOpts{PopulateFromRaw: PopulateFromRawCNAME})
	MustRegisterType("CF_SINGLE_REDIRECT", RegisterOpts{PopulateFromRaw: PopulateFromRawCFSINGLEREDIRECT})
	MustRegisterType("CAA", RegisterOpts{PopulateFromRaw: PopulateFromRawCAA})
	MustRegisterType("DS", RegisterOpts{PopulateFromRaw: PopulateFromRawDS})
	MustRegisterType("DNSKEY", RegisterOpts{PopulateFromRaw: PopulateFromRawDNSKEY})
}

// RecordType is a constraint for DNS records.
type RecordType interface {
	A | MX | SRV | CNAME | CFSINGLEREDIRECT | CAA | DS | DNSKEY
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
	case "DS":
		return RecordUpdateFields(rc,
			DS{KeyTag: rc.DsKeyTag, Algorithm: rc.DsAlgorithm, DigestType: rc.DsDigestType, Digest: rc.DsDigest},
			nil,
		)
	case "DNSKEY":
		return RecordUpdateFields(rc,
			DNSKEY{Flags: rc.DnskeyFlags, Protocol: rc.DnskeyProtocol, Algorithm: rc.DnskeyAlgorithm, PublicKey: rc.DnskeyPublicKey},
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
	case "DS":
		f := rc.Fields.(*DS)
		rc.DsKeyTag = f.KeyTag
		rc.DsAlgorithm = f.Algorithm
		rc.DsDigestType = f.DigestType
		rc.DsDigest = f.Digest

		rc.Comparable = fmt.Sprintf("%d %d %d %s", f.KeyTag, f.Algorithm, f.DigestType, f.Digest)
	case "DNSKEY":
		f := rc.Fields.(*DNSKEY)
		rc.DnskeyFlags = f.Flags
		rc.DnskeyProtocol = f.Protocol
		rc.DnskeyAlgorithm = f.Algorithm
		rc.DnskeyPublicKey = f.PublicKey

		rc.Comparable = fmt.Sprintf("%d %d %d %s", f.Flags, f.Protocol, f.Algorithm, f.PublicKey)
	default:
		return fmt.Errorf("unknown (Seal) rtype %q", rc.Type)
	}
	rc.Display = rc.Comparable

	return nil
}

// Copy returns a deep copy of a RecordConfig.
func (rc *RecordConfig) Copy() (*RecordConfig, error) {
	newR := &RecordConfig{}
	// Copy the exported fields.
	err := reprint.FromTo(rc, newR) // Deep copy
	// Copy each unexported field.
	newR.target = rc.target

	// Copy the fields to new memory so there is no aliasing.
	switch rc.Type {
	case "A":
		newR.Fields = &A{}
		newR.Fields = rc.Fields.(*A)
	case "MX":
		newR.Fields = &MX{}
		newR.Fields = rc.Fields.(*MX)
	case "SRV":
		newR.Fields = &SRV{}
		newR.Fields = rc.Fields.(*SRV)
	case "CNAME":
		newR.Fields = &CNAME{}
		newR.Fields = rc.Fields.(*CNAME)
	case "CFSINGLEREDIRECT":
		newR.Fields = &CFSINGLEREDIRECT{}
		newR.Fields = rc.Fields.(*CFSINGLEREDIRECT)
	case "CAA":
		newR.Fields = &CAA{}
		newR.Fields = rc.Fields.(*CAA)
	case "DS":
		newR.Fields = &DS{}
		newR.Fields = rc.Fields.(*DS)
	case "DNSKEY":
		newR.Fields = &DNSKEY{}
		newR.Fields = rc.Fields.(*DNSKEY)
	}
	//fmt.Printf("DEBUG: COPYING rc=%v new=%v\n", rc.Fields, newR.Fields)
	return newR, err
}

func PopulateFromFields(rc *RecordConfig, rtype string, fields []string, origin string) error {
	switch rtype {
	case "A":
		if rdata, err := ParseA(fields, origin); err == nil {
			return RecordUpdateFields(rc, rdata, nil)
		}
	case "MX":
		if rdata, err := ParseMX(fields, origin); err == nil {
			return RecordUpdateFields(rc, rdata, nil)
		}
	case "SRV":
		if rdata, err := ParseSRV(fields, origin); err == nil {
			return RecordUpdateFields(rc, rdata, nil)
		}
	case "CNAME":
		if rdata, err := ParseCNAME(fields, origin); err == nil {
			return RecordUpdateFields(rc, rdata, nil)
		}
	case "CFSINGLEREDIRECT":
		if rdata, err := ParseCFSINGLEREDIRECT(fields, origin); err == nil {
			return RecordUpdateFields(rc, rdata, nil)
		}
	case "CAA":
		if rdata, err := ParseCAA(fields, origin); err == nil {
			return RecordUpdateFields(rc, rdata, nil)
		}
	case "DS":
		if rdata, err := ParseDS(fields, origin); err == nil {
			return RecordUpdateFields(rc, rdata, nil)
		}
	case "DNSKEY":
		if rdata, err := ParseDNSKEY(fields, origin); err == nil {
			return RecordUpdateFields(rc, rdata, nil)
		}
	}
	return fmt.Errorf("rtype %q not found (%v)", rtype, fields)
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
	case "DS":
		return rc.AsDS().Digest
	case "DNSKEY":
		return rc.AsDNSKEY().PublicKey
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

// SetTargetA sets the A fields.
func (rc *RecordConfig) SetTargetA(a string) error {
	rc.Type = "A"
	rdata, err := ParseA([]string{a}, "")
	if err != nil {
		return err
	}
	return RecordUpdateFields(rc, rdata, nil)
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

// SetTargetMX sets the MX fields.
func (rc *RecordConfig) SetTargetMX(preference uint16, mx string) error {
	rc.Type = "MX"
	return RecordUpdateFields(rc, MX{Preference: preference, Mx: mx}, nil)
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

// SetTargetSRV sets the SRV fields.
func (rc *RecordConfig) SetTargetSRV(priority uint16, weight uint16, port uint16, target string) error {
	rc.Type = "SRV"
	return RecordUpdateFields(rc, SRV{Priority: priority, Weight: weight, Port: port, Target: target}, nil)
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

// SetTargetCNAME sets the CNAME fields.
func (rc *RecordConfig) SetTargetCNAME(target string) error {
	rc.Type = "CNAME"
	return RecordUpdateFields(rc, CNAME{Target: target}, nil)
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

// SetTargetCFSINGLEREDIRECT sets the CFSINGLEREDIRECT fields.
func (rc *RecordConfig) SetTargetCFSINGLEREDIRECT(srname string, code uint16, srwhen string, srthen string) error {
	rc.Type = "CFSINGLEREDIRECT"
	return RecordUpdateFields(rc, CFSINGLEREDIRECT{SRName: srname, Code: code, SRWhen: srwhen, SRThen: srthen, SRDisplay: cfSingleRedirecttargetFromRaw(srname, code, srwhen, srthen), PRWhen: "UNKNOWABLE", PRThen: "UNKNOWABLE", PRDisplay: "UNKNOWABLE"}, nil)
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
	rawfields, meta, err := BuilderCAA(rawfields, meta, origin)
	if err != nil {
		return err
	}

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

// SetTargetCAA sets the CAA fields.
func (rc *RecordConfig) SetTargetCAA(flag uint8, tag string, value string) error {
	rc.Type = "CAA"
	return RecordUpdateFields(rc, CAA{Flag: flag, Tag: tag, Value: value}, nil)
}

//// DS

// DS is the fields needed to store a DNS record of type DS.
type DS struct {
	KeyTag     uint16
	Algorithm  uint8
	DigestType uint8
	Digest     string `dnscontrol:"_,target,allcaps"`
}

func ParseDS(rawfields []string, origin string) (DS, error) {

	// Error checking
	if errorCheckFieldCount(rawfields, 4) {
		return DS{}, fmt.Errorf("rtype DS wants %d field(s), found %d: %+v", 4, len(rawfields)-1, rawfields[1:])
	}
	var keytag uint16
	var algorithm uint8
	var digesttype uint8
	var digest string
	var err error
	if keytag, err = fieldtypes.ParseUint16(rawfields[0]); err != nil {
		return DS{}, err
	}
	if algorithm, err = fieldtypes.ParseUint8(rawfields[1]); err != nil {
		return DS{}, err
	}
	if digesttype, err = fieldtypes.ParseUint8(rawfields[2]); err != nil {
		return DS{}, err
	}
	if digest, err = fieldtypes.ParseStringTrimmedAllCaps(rawfields[3]); err != nil {
		return DS{}, err
	}

	return DS{KeyTag: keytag, Algorithm: algorithm, DigestType: digesttype, Digest: digest}, nil
}

// PopulateFromRawDS updates rc to be an DS record with contents from rawfields, meta and origin.
func PopulateFromRawDS(rc *RecordConfig, rawfields []string, meta map[string]string, origin string) error {
	rc.Type = "DS"

	// First rawfield is the label.
	if err := rc.SetLabel3(rawfields[0], rc.SubDomain, origin); err != nil {
		return err
	}

	// Parse the remaining fields.
	rdata, err := ParseDS(rawfields[1:], origin)
	if err != nil {
		return err
	}

	return RecordUpdateFields(rc, rdata, meta)
}

// AsDS returns rc.Fields as an DS struct.
func (rc *RecordConfig) AsDS() *DS {
	return rc.Fields.(*DS)
}

// GetFieldsDS returns rc.Fields as individual typed values.
func (rc *RecordConfig) GetFieldsDS() (uint16, uint8, uint8, string) {
	n := rc.AsDS()
	return n.KeyTag, n.Algorithm, n.DigestType, n.Digest
}

// GetFieldsAsStringsDS returns rc.Fields as individual strings.
func (rc *RecordConfig) GetFieldsAsStringsDS() [4]string {
	n := rc.AsDS()
	return [4]string{strconv.Itoa(int(n.KeyTag)), strconv.Itoa(int(n.Algorithm)), strconv.Itoa(int(n.DigestType)), n.Digest}
}

// SetTargetDS sets the DS fields.
func (rc *RecordConfig) SetTargetDS(keytag uint16, algorithm uint8, digesttype uint8, digest string) error {
	rc.Type = "DS"
	return RecordUpdateFields(rc, DS{KeyTag: keytag, Algorithm: algorithm, DigestType: digesttype, Digest: digest}, nil)
}

//// DNSKEY

// DNSKEY is the fields needed to store a DNS record of type DNSKEY.
type DNSKEY struct {
	Flags     uint16
	Protocol  uint8
	Algorithm uint8
	PublicKey string `dns:"base64"`
}

func ParseDNSKEY(rawfields []string, origin string) (DNSKEY, error) {

	// Error checking
	if errorCheckFieldCount(rawfields, 4) {
		return DNSKEY{}, fmt.Errorf("rtype DNSKEY wants %d field(s), found %d: %+v", 4, len(rawfields)-1, rawfields[1:])
	}
	var flags uint16
	var protocol uint8
	var algorithm uint8
	var publickey string
	var err error
	if flags, err = fieldtypes.ParseUint16(rawfields[0]); err != nil {
		return DNSKEY{}, err
	}
	if protocol, err = fieldtypes.ParseUint8(rawfields[1]); err != nil {
		return DNSKEY{}, err
	}
	if algorithm, err = fieldtypes.ParseUint8(rawfields[2]); err != nil {
		return DNSKEY{}, err
	}
	if publickey, err = fieldtypes.ParseStringTrimmed(rawfields[3]); err != nil {
		return DNSKEY{}, err
	}

	return DNSKEY{Flags: flags, Protocol: protocol, Algorithm: algorithm, PublicKey: publickey}, nil
}

// PopulateFromRawDNSKEY updates rc to be an DNSKEY record with contents from rawfields, meta and origin.
func PopulateFromRawDNSKEY(rc *RecordConfig, rawfields []string, meta map[string]string, origin string) error {
	rc.Type = "DNSKEY"

	// First rawfield is the label.
	if err := rc.SetLabel3(rawfields[0], rc.SubDomain, origin); err != nil {
		return err
	}

	// Parse the remaining fields.
	rdata, err := ParseDNSKEY(rawfields[1:], origin)
	if err != nil {
		return err
	}

	return RecordUpdateFields(rc, rdata, meta)
}

// AsDNSKEY returns rc.Fields as an DNSKEY struct.
func (rc *RecordConfig) AsDNSKEY() *DNSKEY {
	return rc.Fields.(*DNSKEY)
}

// GetFieldsDNSKEY returns rc.Fields as individual typed values.
func (rc *RecordConfig) GetFieldsDNSKEY() (uint16, uint8, uint8, string) {
	n := rc.AsDNSKEY()
	return n.Flags, n.Protocol, n.Algorithm, n.PublicKey
}

// GetFieldsAsStringsDNSKEY returns rc.Fields as individual strings.
func (rc *RecordConfig) GetFieldsAsStringsDNSKEY() [4]string {
	n := rc.AsDNSKEY()
	return [4]string{strconv.Itoa(int(n.Flags)), strconv.Itoa(int(n.Protocol)), strconv.Itoa(int(n.Algorithm)), n.PublicKey}
}

// SetTargetDNSKEY sets the DNSKEY fields.
func (rc *RecordConfig) SetTargetDNSKEY(flags uint16, protocol uint8, algorithm uint8, publickey string) error {
	rc.Type = "DNSKEY"
	return RecordUpdateFields(rc, DNSKEY{Flags: flags, Protocol: protocol, Algorithm: algorithm, PublicKey: publickey}, nil)
}
