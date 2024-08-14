package rtypesingleredirect

import (
	"encoding/json"
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/pkg/rtypecontrol"
)

// Name is the string name for this rType.
const Name = "CLOUDFLAREAPI_SINGLE_REDIRECT"

func init() {
	rtypecontrol.Register(rtypecontrol.RegisterTypeOpts{
		Name: Name,
	})
}

// SingleRedirect contains info about a Cloudflare Single Redirect.
type SingleRedirect struct {
	//
	Code uint16 `json:"code,omitempty"` // 301 or 302
	// PR == PageRule
	PRWhen     string `json:"pr_when,omitempty"`
	PRThen     string `json:"pr_then,omitempty"`
	PRPriority int    `json:"pr_priority,omitempty"` // Really an identifier for the rule.
	PRDisplay  string `json:"pr_display,omitempty"`  // How is this displayed to the user (SetTarget) for CF_REDIRECT/CF_TEMP_REDIRECT
	//
	// SR == SingleRedirect
	SRName           string `json:"sr_name,omitempty"` // How is this displayed to the user
	SRWhen           string `json:"sr_when,omitempty"`
	SRThen           string `json:"sr_then,omitempty"`
	SRRRulesetID     string `json:"sr_rulesetid,omitempty"`
	SRRRulesetRuleID string `json:"sr_rulesetruleid,omitempty"`
	SRDisplay        string `json:"sr_display,omitempty"` // How is this displayed to the user (SetTarget) for CF_SINGLE_REDIRECT
}

func (rdata *SingleRedirect) Name() string {
	return Name
}

func (rdata *SingleRedirect) ComputeTarget() string {
	// The closest equivalent to a target "hostname" is the rule name.
	return rdata.SRName
}

func (rdata *SingleRedirect) ComputeComparableMini() string {
	// The differencing engine uses this.
	return rdata.SRDisplay
}

func (rdata *SingleRedirect) MarshalJSON() ([]byte, error) {
	return json.Marshal(*rdata)
}

// FromRawArgs creates a Rdata...
// update a RecordConfig using the args (from a
// RawRecord.Args). In other words, use the data from dnsconfig.js's
// rawrecordBuilder to create (actually... update) a models.RecordConfig.
func FromRawArgs(items []any, name string) (*SingleRedirect, error) {

	// Pave the arguments.
	if err := rtypecontrol.PaveArgs(items, "iss"); err != nil {
		return nil, err
	}

	// Unpack the arguments:
	var code = items[0].(uint16)
	if code != 301 && code != 302 {
		return nil, fmt.Errorf("code (%03d) is not 301 or 302", code)
	}
	var when = items[1].(string)
	var then = items[2].(string)

	// Use the arguments to perfect the record:
	return makeSingleRedirect(code, name, when, then)
}

// makeSingleRedirect
func makeSingleRedirect(code uint16, name, when, then string) (*SingleRedirect, error) {
	target := targetFromRaw(name, code, when, then)

	rdata := &SingleRedirect{
		Code: code,
		//
		PRWhen:     "UNKNOWABLE",
		PRThen:     "UNKNOWABLE",
		PRPriority: 0,
		PRDisplay:  "UNKNOWABLE",
		//
		SRName:    name,
		SRWhen:    when,
		SRThen:    then,
		SRDisplay: target,
	}
	return rdata, nil
}

// targetFromRaw create the display text used for a normal Redirect.
func targetFromRaw(name string, code uint16, when, then string) string {
	return fmt.Sprintf("%s code=(%03d) when=(%s) then=(%s)",
		name,
		code,
		when,
		then,
	)
}
