package cfsingleredirect

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/rtypecontrol"
)

func init() {
	rtypecontrol.Register(&SingleRedirectConfig{})
}

// SingleRedirectConfig contains info about a Cloudflare Single Redirect.
type SingleRedirectConfig struct {
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

// Name returns the text (all caps) name of the rtype.
func (handle *SingleRedirectConfig) Name() string {
	return "CLOUDFLAREAPI_SINGLE_REDIRECT"
}

func (handle *SingleRedirectConfig) FromArgs(rec *models.RecordConfig, args []any) error {
	// Pave the args to be the expected types.
	if err := rtypecontrol.PaveArgs(args, "siss"); err != nil {
		return err
	}

	// Unpack the args:
	var name, when, then string
	var code uint16

	name = args[0].(string)
	code = args[1].(uint16)
	if code != 301 && code != 302 && code != 303 && code != 307 && code != 308 {
		return fmt.Errorf("%s: code (%03d) is not 301,302,303,307,308", rec.FilePos, code)
	}
	when = args[2].(string)
	then = args[3].(string)
	display := targetFromRaw(name, code, when, then)

	rec.F = &SingleRedirectConfig{
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
		SRDisplay: display,
	}

	rec.Comparable = display
	rec.ZonefilePartial = display

	return nil
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

func (handle *SingleRedirectConfig) AsRFC1038String(*models.RecordConfig) string {
	return handle.SRDisplay
}

func (handle *SingleRedirectConfig) CopyToLegacyFields(rec *models.RecordConfig) {
	rec.SetTarget(handle.SRDisplay)
}

//func (handle *SingleRedirectConfig) IDNFields(argsRaw) (argsIDN, argsUnicode, error) {}
//func (handle *SingleRedirectConfig) CopyFromLegacyFields(*models.RecordConfig)       {}
