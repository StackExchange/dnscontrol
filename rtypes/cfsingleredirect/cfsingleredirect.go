package cfsingleredirect

import (
	"fmt"
	"strconv"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/providers/cloudflare/singleredirect"
)

func FromRaw(rc *models.RecordConfig, items []any) error {
	var err error

	rc.Type = "CF_SINGLE_REDIRECT"
	var name, when, then string
	var code int

	name = items[0].(string)

	ucode := items[1]
	switch v := ucode.(type) {
	case int:
		code = v
	case string:
		code, err = strconv.Atoi(v)
		if err != nil {
			return err
		}
	default:
		fmt.Printf("code %q unexpected type %T", ucode, v)
	}

	when, then = items[2].(string), items[3].(string)

	s := singleredirect.FromAPIData(when, then, code)

	rc.Name = name
	rc.CloudflareRedirect = s

	return nil
}
