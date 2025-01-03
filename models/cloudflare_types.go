package models

import (
	"fmt"
	"maps"
	"strconv"

	"github.com/StackExchange/dnscontrol/v4/pkg/fieldtypes"
)

func init() {
	RegisterType("CF_SINGLE_REDIRECT", RegisterOpts{FromRaw: PopulateCFSINGLEREDIRECTRaw})
	//fmt.Printf("DEBUG: REGISTERED CF_SINGLE_REDIRECT\n")
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

// PopulateCFSINGLEREDIRECTRaw updates rc to be an CFSINGLEREDIRECT record with contents from rawfields, meta and origin.
func PopulateCFSINGLEREDIRECTRaw(rc *RecordConfig, rawfields []string, meta map[string]string, origin string) error {
	var err error

	// Error checking

	if len(rawfields) <= 3 {
		return fmt.Errorf("rtype CFSINGLEREDIRECT wants %d field(s), found %d: %+v", 1, len(rawfields)-1, rawfields[1:])
	}

	// Convert each rawfield.

	rc.SetLabel(rawfields[0], origin) // Label

	var srname string
	if srname, err = fieldtypes.ParseStringTrimmed(rawfields[0]); err != nil {
		return err
	}
	var code uint16
	if code, err = fieldtypes.ParseUint16(rawfields[1]); err != nil {
		return err
	}
	var srwhen string
	if srwhen, err = fieldtypes.ParseStringTrimmed(rawfields[2]); err != nil {
		return err
	}
	var srthen string
	if srthen, err = fieldtypes.ParseStringTrimmed(rawfields[3]); err != nil {
		return err
	}

	return rc.PopulateCFSINGLEREDIRECTFields(srname, code, srwhen, srthen, meta, origin)
}

// PopulateCFSINGLEREDIRECTFields updates rc to be an CFSINGLEREDIRECT record with contents from typed data, meta, and origin.
func (rc *RecordConfig) PopulateCFSINGLEREDIRECTFields(srname string, code uint16,
	srwhen, srthen string, meta map[string]string, origin string) error {
	// Create the struct if needed.
	if rc.Fields == nil {
		rc.Fields = &CFSINGLEREDIRECT{}
	}

	// Process each field:

	n := rc.Fields.(*CFSINGLEREDIRECT)
	n.SRName = string(srname)
	n.Code = uint16(code)
	n.SRWhen = string(srwhen)
	n.SRThen = string(srthen)

	// Update legacy fields.
	MakeSingleRedirectFromRawRec(rc, code, srname, srwhen, srthen)

	// Update the RecordConfig:
	maps.Copy(rc.Metadata, meta) // Add the metadata
	rc.Comparable = fmt.Sprintf("%q %d %q %q", srname, code, srwhen, srthen)
	rc.Display = rc.Comparable

	return nil
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
