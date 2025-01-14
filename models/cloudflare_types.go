package models

import (
	"fmt"
	"strconv"

	"github.com/StackExchange/dnscontrol/v4/pkg/fieldtypes"
)

func init() {
	RegisterType("CF_SINGLE_REDIRECT", RegisterOpts{PopulateFromRaw: PopulateFromRawCFSINGLEREDIRECT})
}

//// CFSINGLEREDIRECT

// CFSINGLEREDIRECT contains info about a Cloudflare Single Redirect.
//
//	When these are used, .target is set to a human-readable version (only to be used for display purposes).
type CFSINGLEREDIRECT struct {
	//
	// PR == PageRule
	PRWhen     string `dns:"skip" json:"pr_when,omitempty"`
	PRThen     string `dns:"skip" json:"pr_then,omitempty"`
	PRPriority int    `dns:"skip" json:"pr_priority,omitempty"` // Really an identifier for the rule.
	PRDisplay  string `dns:"skip" json:"pr_display,omitempty"`  // How is this displayed to the user (SetTarget) for CF_REDIRECT/CF_TEMP_REDIRECT
	//
	// SR == SingleRedirect
	SRName           string `json:"sr_name,omitempty"` // How is this displayed to the user
	Code             uint16 `json:"code,omitempty"`    // 301 or 302
	SRWhen           string `json:"sr_when,omitempty"`
	SRThen           string `json:"sr_then,omitempty"`
	SRRRulesetID     string `dns:"skip" json:"sr_rulesetid,omitempty"`
	SRRRulesetRuleID string `dns:"skip" json:"sr_rulesetruleid,omitempty"`
	SRDisplay        string `dns:"skip" json:"sr_display,omitempty"` // How is this displayed to the user (SetTarget) for CF_SINGLE_REDIRECT
}

func ParseCFSINGLEREDIRECT(rawfields []string, origin string) (CFSINGLEREDIRECT, error) {
	var srname, srwhen, srthen string
	var code uint16
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

	return CFSINGLEREDIRECT{
		PRWhen:     "UNKNOWABLE",
		PRThen:     "UNKNOWABLE",
		PRPriority: 0,
		PRDisplay:  "UNKNOWABLE",

		SRName:           srname,
		Code:             code,
		SRWhen:           srwhen,
		SRThen:           srthen,
		SRRRulesetID:     "",
		SRRRulesetRuleID: "",
		SRDisplay:        cfSingleRedirecttargetFromRaw(srname, code, srwhen, srthen),
	}, nil
}

func NewFromRawCFSINGLEREDIRECT(rawfields []string, meta map[string]string, origin string, ttl uint32) (*RecordConfig, error) {
	rc := &RecordConfig{TTL: ttl}
	if err := PopulateFromRawCFSINGLEREDIRECT(rc, rawfields, meta, origin); err != nil {
		return nil, err
	}
	return rc, nil
}

// PopulateFromRawCFSINGLEREDIRECT updates rc to be an CFSINGLEREDIRECT record with contents from rawfields, meta and origin.
func PopulateFromRawCFSINGLEREDIRECT(rc *RecordConfig, rawfields []string, meta map[string]string, origin string) error {
	rc.Type = "CF_SINGLE_REDIRECT"
	rc.TTL = 1

	// Error checking
	if len(rawfields) <= 3 {
		return fmt.Errorf("rtype CFSINGLEREDIRECT wants %d field(s), found %d: %+v", 1, len(rawfields)-1, rawfields[1:])
	}

	// First rawfield is the label.
	if origin != "" { //  If we don't know the origin, don't muck with the label.
		if err := rc.SetLabel3(rawfields[0], rc.SubDomain, origin); err != nil {
			return err
		}
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

// GetCFSINGLEREDIRECTFields returns rc.Fields as individual typed values.
func (rc *RecordConfig) GetCFSINGLEREDIRECTFields() (string, uint16, string, string) {
	n := rc.AsCFSINGLEREDIRECT()
	return n.SRName, n.Code, (n.SRWhen), (n.SRThen)
}

// GetCFSINGLEREDIRECTStrings returns rc.Fields as individual strings.
func (rc *RecordConfig) GetCFSINGLEREDIRECTStrings() [4]string {
	n := rc.AsCFSINGLEREDIRECT()
	return [4]string{n.SRName, strconv.Itoa(int(n.Code)), n.SRWhen, n.SRThen}
}
