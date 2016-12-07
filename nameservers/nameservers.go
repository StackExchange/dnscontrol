//Package nameservers provides logic for dynamically finding nameservers for a domain, and configuring NS records for them.
package nameservers

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers"
	"github.com/miekg/dns/dnsutil"
)

//DetermineNameservers will find all nameservers we should use for a domain. It follows the following rules:
//1. All explicitly defined NAMESERVER records will be used.
//2. Each DSP declares how many nameservers to use. Default is all. 0 indicates to use none.
func DetermineNameservers(dc *models.DomainConfig, maxNS int, dsps map[string]providers.DNSServiceProvider) ([]*models.Nameserver, error) {
	//always take explicit
	ns := dc.Nameservers
	for dsp, n := range dc.Dsps {
		if n == 0 {
			continue
		}
		fmt.Printf("----- Getting nameservers from: %s\n", dsp)
		p, ok := dsps[dsp]
		if !ok {
			return nil, fmt.Errorf("DNS provider %s not declared", dsp)
		}
		nss, err := p.GetNameservers(dc.Name)
		if err != nil {
			return nil, err
		}
		take := len(nss)
		if n > 0 && n < take {
			take = n
		}
		for i := 0; i < take; i++ {
			ns = append(ns, nss[i])
		}
	}
	return ns, nil
}

//AddNSRecords creates NS records on a domain corresponding to the nameservers specified.
func AddNSRecords(dc *models.DomainConfig) {
	for _, ns := range dc.Nameservers {
		rc := &models.RecordConfig{
			Type:     "NS",
			Name:     "@",
			Target:   ns.Name + ".",
			Metadata: map[string]string{},
		}
		rc.NameFQDN = dnsutil.AddOrigin(rc.Name, dc.Name)
		dc.Records = append(dc.Records, rc)
	}
}
