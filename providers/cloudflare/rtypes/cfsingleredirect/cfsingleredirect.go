package cfsingleredirect

import (
	"fmt"
	"strconv"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/rtypecontrol"
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

	if code != 301 && code != 302 {
		return fmt.Errorf("code (%03d) is not 301 or 302", code)
	}

	when, then = items[1].(string), items[2].(string)

	s := FromAPIData(when, then, code)

	rc.SetTarget(fmt.Sprintf("code=%03d when=(%v) then=(%v)", code, when, then))

	rc.CloudflareRedirect = s

	return nil
}

func FromAPIData(sm, sr string, code int) *models.CloudflareSingleRedirectConfig {
	r := &models.CloudflareSingleRedirectConfig{
		PRMatcher:     "UNKNOWABLE",
		PRReplacement: "UNKNOWABLE",
		Code:          code,
		SRMatcher:     sm,
		SRReplacement: sr,
	}
	return r
}
