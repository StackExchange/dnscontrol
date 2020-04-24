// Package nameservers provides logic for dynamically finding nameservers for a domain, and configuring NS records for them.
package nameservers

import (
	"fmt"
	"strings"

	"strconv"

	"github.com/StackExchange/dnscontrol/v3/models"
)

// DetermineNameservers will find all nameservers we should use for a domain. It follows the following rules:
// 1. All explicitly defined NAMESERVER records will be used.
// 2. Each DSP declares how many nameservers to use. Default is all. 0 indicates to use none.
func DetermineNameservers(dc *models.DomainConfig) ([]*models.Nameserver, error) {
	// always take explicit
	ns := dc.Nameservers
	for _, dnsProvider := range dc.DNSProviderInstances {
		n := dnsProvider.NumberOfNameservers
		if n == 0 {
			continue
		}
		fmt.Printf("----- Getting nameservers from: %s\n", dnsProvider.Name)
		nss, err := dnsProvider.Driver.GetNameservers(dc.Name)
		if err != nil {
			return nil, err
		}
		// Clean up the nameservers due to
		// https://github.com/StackExchange/dnscontrol/issues/491
		// In the far future, this warning will become a fatal error.
		for i := range nss {
			if strings.HasSuffix(nss[i].Name, ".") {
				models.WarnNameserverDot(dnsProvider.Name, fmt.Sprintf("DetermineNameservers (%s) (%s)", dc.Name, nss[i].Name))
				nss[i].Name = strings.TrimSuffix(nss[i].Name, ".")
			}
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

// AddNSRecords creates NS records on a domain corresponding to the nameservers specified.
func AddNSRecords(dc *models.DomainConfig) {
	ttl := uint32(300)
	if ttls, ok := dc.Metadata["ns_ttl"]; ok {
		t, err := strconv.ParseUint(ttls, 10, 32)
		if err != nil {
			fmt.Printf("WARNING: ns_ttl for %s (%s) is not a valid int", dc.Name, ttls)
		} else {
			ttl = uint32(t)
		}
	}
	for _, ns := range dc.Nameservers {
		rc := &models.RecordConfig{
			Type:     "NS",
			Metadata: map[string]string{},
			TTL:      ttl,
		}
		rc.SetLabel("@", dc.Name)
		t := ns.Name
		if !strings.HasSuffix(t, ".") {
			t += "."
		}
		rc.SetTarget(t)

		dc.Records = append(dc.Records, rc)
	}
}
