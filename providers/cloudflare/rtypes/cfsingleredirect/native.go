package cfsingleredirect

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/models"
)

func FromAPIData(sn, sm, sr string, code uint16) *models.CloudflareSingleRedirectConfig {
	r := &models.CloudflareSingleRedirectConfig{
		Code:    code,
		Display: fmt.Sprintf("name=(%s) code=%03d when=(%v) then=(%v)", sn, code, sm, sr),
		//
		PRWhen: "UNKNOWABLE",
		PRThen: "UNKNOWABLE",
		SRName: sn,
		SRWhen: sm,
		SRThen: sr,
	}
	return r
}
