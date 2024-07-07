package cfsingleredirect

import (
	"fmt"
	"strconv"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/rtypecontrol"
	"github.com/StackExchange/dnscontrol/v4/providers/cloudflare/singleredirect"
)

func init() {
	rtypecontrol.Register("CF_SINGLE_REDIRECT")
}

func FromRaw(rc *models.RecordConfig, items []any) error {
	var err error

	// Unpack the args:

	var when, then string
	var code int

	if err := rtypecontrol.CheckArgTypes(items, "iss"); err != nil {
		return err
	}

	ucode := items[0]
	switch v := ucode.(type) {
	case int:
		code = v
	case float64:
		code = int(v)
	case string:
		code, err = strconv.Atoi(v)
		if err != nil {
			return err
		}
	default:
		fmt.Printf("code %q unexpected type %T", ucode, v)
	}

	when, then = items[1].(string), items[2].(string)

	s := singleredirect.FromAPIData(when, then, code)

	rc.SetTarget(fmt.Sprintf("code=%03d when=(%v) then=(%v)", code, when, then))

	rc.CloudflareRedirect = s

	return nil
}
