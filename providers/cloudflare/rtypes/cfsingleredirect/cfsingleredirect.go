package cfsingleredirect

import (
	"fmt"
	"strconv"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/rtypecontrol"
)

func init() {
	rtypecontrol.Register("CLOUDFLAREAPI_SINGLE_REDIRECT")
}

func FromRaw(rc *models.RecordConfig, items []any) error {
	fmt.Printf("DEBUG: FromRaw: items=%+v\n", items)

	if len(items) != 4 {
		return fmt.Errorf("expected 4 items: %v", items)
	}
	var err error

	// Validate types.
	if err := rtypecontrol.CheckArgTypes(items, "siss"); err != nil {
		return err
	}

	// Unpack the args:
	var name, when, then string
	var code int

	name = items[0].(string)

	ucode := items[1]
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
		return fmt.Errorf("code %q unexpected type %T", ucode, v)
	}

	if code != 301 && code != 302 {
		return fmt.Errorf("code (%03d) is not 301 or 302", code)
	}

	when, then = items[2].(string), items[3].(string)

	rc.Name = name
	rc.SetTarget(fmt.Sprintf("code=%03d when=(%v) then=(%v)", code, when, then))
	rc.CloudflareRedirect = FromAPIData(when, then, code)
	fmt.Printf("DEBUG: FromRaw: result cr=%+v\n", rc.CloudflareRedirect)

	return nil
}
