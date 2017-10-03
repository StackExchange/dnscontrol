//Package nameservers provides logic for dynamically finding nameservers for a domain, and configuring NS records for them.
package nameservers

import (
	"fmt"
	"strings"

	"strconv"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/miekg/dns/dnsutil"
)

//DetermineNameservers will find all nameservers we should use for a domain. It follows the following rules:
//1. All explicitly defined NAMESERVER records will be used.
//2. Each DSP declares how many nameservers to use. Default is all. 0 indicates to use none.
func DetermineNameservers(dc *models.DomainConfig) ([]*models.Nameserver, error) {
	//always take explicit
	ns := dc.Nameservers
	for _, provider := range dc.DNSProviders {
		n := provider.NumberOfNameserversToUse()
		if n == 0 {
			continue
		}
		fmt.Printf("----- Getting nameservers from: %s\n", provider.Name())
		nss, err := provider.GetNameservers(dc.Name)
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
	ttl := uint32(300)
	if ttls, ok := dc.Metadata["ns_ttl"]; ok {
		t, err := strconv.ParseUint(ttls, 10, 32)
		if err != nil {
			fmt.Printf("WARNING: ns_ttl fpr %s (%s) is not a valid int", dc.Name, ttls)
		} else {
			ttl = uint32(t)
		}
	}
	for _, ns := range dc.Nameservers {
		rc := &models.RecordConfig{
			Type:     "NS",
			Name:     "@",
			Target:   ns.Name,
			Metadata: map[string]string{},
			TTL:      ttl,
		}
		if !strings.HasSuffix(rc.Target, ".") {
			rc.Target += "."
		}
		rc.NameFQDN = dnsutil.AddOrigin(rc.Name, dc.Name)
		dc.Records = append(dc.Records, rc)
	}
}
