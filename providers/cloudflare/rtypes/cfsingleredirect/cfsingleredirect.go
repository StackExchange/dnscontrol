package cfsingleredirect

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/rtypecontrol"
)

func init() {
	rtypecontrol.Register("CLOUDFLAREAPI_SINGLE_REDIRECT")
}

// FromRawArgs update a models.RecordConfig using the args (from a
// models.RawRecord.Args). In other words, use the data from dnsconfig.js's
// rawrecordBuilder to create (actually... update) a models.RecordConfig.
func FromRawArgs(rc *models.RecordConfig, items []any) error {

	// Validate types.
	if err := rtypecontrol.PaveArgs(items, "siss"); err != nil {
		return err
	}

	// Unpack the args:
	var name = items[0].(string)
	var code = items[1].(uint16)
	if code != 301 && code != 302 {
		return fmt.Errorf("code (%03d) is not 301 or 302", code)
	}
	var when = items[2].(string)
	var then = items[3].(string)

	rc.Name = name // FIXME: Parent should set?  normalize?  Maybe not normalize.  Normalizing happens later and might not be wanted (as in CF_SINGLE_REDIRECT)
	rc.CloudflareRedirect = MakeRdata(when, then, code)
	rc.SetTarget(rc.CloudflareRedirect.SRDisplay)

	return nil
}

func MakeRdata(sm, sr string, code uint16) *models.CloudflareSingleRedirectConfig {
	r := &models.CloudflareSingleRedirectConfig{
		PRWhen:    "UNKNOWABLE",
		PRThen:    "UNKNOWABLE",
		Code:      code,
		SRDisplay: fmt.Sprintf("code=%03d when=(%v) then=(%v)", code, sm, sr),
		SRWhen:    sm,
		SRThen:    sr,
	}
	return r
}
