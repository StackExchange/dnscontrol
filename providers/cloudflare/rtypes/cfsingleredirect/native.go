package cfsingleredirect

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/models"
)

func FromAPIData(sm, sr string, code uint16) *models.CloudflareSingleRedirectConfig {
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
