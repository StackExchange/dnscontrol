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

// FromRawArgs update a models.RecordConfig using the args (from a
// models.RawRecord.Args). In other words, use the data from dnsconfig.js's
// rawrecordBuilder to create (actually... update) a models.RecordConfig.
func FromRawArgs(rc *models.RecordConfig, items []any) error {

	// Pave the arguments.
	if err := rtypecontrol.PaveArgs(items, "siss"); err != nil {
		return err
	}

	// Unpack the arguments:
	var name = items[0].(string)
	var code = items[1].(uint16)
	if code != 301 && code != 302 {
		return fmt.Errorf("code (%03d) is not 301 or 302", code)
	}
	var when = items[2].(string)
	var then = items[3].(string)

	// Use the arguments to perfect the record:
	makeSingleRedirectFromRawRec(rc, code, name, when, then)

	return nil
}
