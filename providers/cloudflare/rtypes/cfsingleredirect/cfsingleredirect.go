package cfsingleredirect

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/rtypecontrol"
)

// SINGLEREDIRECT is the string name for this rType.
const SINGLEREDIRECT = "CLOUDFLAREAPI_SINGLE_REDIRECT"

func init() {
	rtypecontrol.Register(SINGLEREDIRECT)
}

// FromRaw convert RecordConfig using data from a RawRecordConfig's parameters.
func FromRaw(rc *models.RecordConfig, items []any) error {

	// Validate types.
	if err := rtypecontrol.PaveArgs(items, "siss"); err != nil {
		return err
	}

	// Unpack the args:
	var name, when, then string
	var code uint16

	name = items[0].(string)
	code = items[1].(uint16)
	if code != 301 && code != 302 {
		return fmt.Errorf("code (%03d) is not 301 or 302", code)
	}
	when = items[2].(string)
	then = items[3].(string)

	makeSingleRedirectFromRawRec(rc, code, name, when, then)

	return nil
}
