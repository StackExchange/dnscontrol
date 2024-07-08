package cfsingleredirect

import "github.com/StackExchange/dnscontrol/v4/models"

func FromAPIData(sm, sr string, code int) *models.CloudflareSingleRedirectConfig {
	r := &models.CloudflareSingleRedirectConfig{
		PRWhen: "UNKNOWABLE",
		PRThen: "UNKNOWABLE",
		Code:   code,
		SRWhen: sm,
		SRThen: sr,
	}
	return r
}
