package cfsingleredirect

import (
	"fmt"
	"maps"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/rtypectl"
)

// SINGLEREDIRECT is the string name for this rType.
const SINGLEREDIRECT = "CF_SINGLE_REDIRECT"

//type CF_SINGLE_REDIRECT struct {
//}

func init() {
	rtypectl.Register("CF_SINGLE_REDIRECT", rtypectl.RegisterOpts{FromRaw: PopulateFromRawCFSINGLEREDIRECT})
	//fmt.Printf("DEBUG: REGISTERED CF_SINGLE_REDIRECT\n")
}

//// FromRaw convert RecordConfig using data from a RawRecordConfig's parameters.
//func FromRaw(rc *models.RecordConfig, items []any) error {
//
//	// Unpack the args:
//	var name, when, then string
//	var code uint16
//
//	name = items[0].(string)
//	code = items[1].(uint16)
//	if code != 301 && code != 302 {
//		return fmt.Errorf("code (%03d) is not 301 or 302", code)
//	}
//	when = items[2].(string)
//	then = items[3].(string)
//
//	makeSingleRedirectFromRawRec(rc, code, name, when, then)

// PopulateFromRawCFSINGLEREDIRECT updates rc to be a CF_SINGLE_REDIRECT record with contents from origin, rawfields and meta.
func PopulateFromRawCFSINGLEREDIRECT(rc *models.RecordConfig, origin string, rawfields []any, meta map[string]string) error {

	// Error checking

	if len(rawfields) <= 3 {
		return fmt.Errorf("rtype %q wants %d field(s), found %d: %+v", "CF_SINGLE_REDIRECT", 1, len(rawfields)-1, rawfields[1:])
	}

	// Create the struct

	//n := CF_SINGLE_REDIRECT{}
	//n := models.CloudflareSingleRedirectConfig{}
	var err error

	var name, when, then string
	var code uint16
	if name, err = rtypectl.ParseLabel(rawfields[0]); err != nil {
		return err
	}
	if code, err = rtypectl.ParseRedirectCode(rawfields[1]); err != nil {
		return err
	}
	if when, err = rtypectl.ParseString(rawfields[2]); err != nil {
		return err
	}
	if then, err = rtypectl.ParseString(rawfields[3]); err != nil {
		return err
	}
	MakeSingleRedirectFromRawRec(rc, code, name, when, then)

	// Update the record:
	//rc.SetLabel(rawfields[0].(string), origin) // Label
	maps.Copy(rc.Metadata, meta) // Add the metadata
	//rc.Comparable = fmt.Sprintf("%s", n.A)
	rc.Display = rc.CloudflareRedirect.SRDisplay

	return nil
}
