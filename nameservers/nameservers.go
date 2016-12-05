//Package nameservers provides logic for dynamically finding nameservers for a domain, and configuring NS records for them.
package nameservers

import (
	"fmt"
	"strconv"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers"
	"github.com/miekg/dns/dnsutil"
)

//DetermineNameservers will find all nameservers we should use for a domain. It follows the following rules:
//1. All explicitly defined NAMESERVER records will be used.
//2. If any NAMESERVERS_FROM providers are present, those dsps will be queried for their nameservers.
//3. If neither 1 or 2 are true, dsps will be queried in order, and the first to respond with valid nameservers will be considered the authority to point at via the registrar.
func DetermineNameservers(dc *models.DomainConfig, maxNS int, dsps map[string]providers.DNSServiceProvider) ([]*models.Nameserver, error) {
	//always take explicit
	ns := dc.Nameservers
	//nameservers_from
	getNs := func(dsp string) ([]*models.Nameserver, error) {
		fmt.Printf("----- Getting nameservers from: %s\n", dsp)
		p, ok := dsps[dsp]
		if !ok {
			return nil, fmt.Errorf("DNS provider %s not declared", dsp)
		}
		return p.GetNameservers(dc.Name)
	}
	if len(dc.NameserversFrom) > 0 {
		perProvider, err := nsCountPerProvider(maxNS, dc)
		if err != nil {
			return nil, err
		}
		for _, nsf := range dc.NameserversFrom {
			nservers, err := getNs(nsf)
			if err != nil {
				return nil, err
			}
			for i, nserver := range nservers {
				if perProvider != 0 && i >= perProvider {
					break
				}
				ns = append(ns, nserver)
			}
		}
	}
	if len(ns) > 0 {
		return ns, nil
	}
	for _, dsp := range dc.Dsps {
		nservers, err := getNs(dsp)
		if err != nil {
			return nil, err
		}
		if len(nservers) > 0 {
			return nservers, nil
		}
	}
	return ns, nil
}

//nsCountPerProvider calculates the number of nameservers we want from each NAMESERVERS_FROM provider.
//zero returned means no limit
func nsCountPerProvider(max int, dc *models.DomainConfig) (int, error) {
	explicitCount := len(dc.Nameservers)
	providerCount := len(dc.NameserversFrom)
	var err error
	if countStr := dc.Metadata["NameserverCount"]; countStr != "" {
		max, err = strconv.Atoi(countStr)
		if err != nil {
			return 0, err
		}
	}
	if max == 0 {
		return 0, nil
	}
	needed := max - explicitCount
	if needed < providerCount {
		return 0, fmt.Errorf("Cannot pull nameservers from dns providers as limit would be exceeded. Set NameserverCount metadata to fix")
	}
	if needed%providerCount != 0 {
		return 0, fmt.Errorf("nameserver count error. Need to find %d nameservers from %d providers. Not possible to do evenly. Set NameserverCount metadata to fix", needed, providerCount)
	}
	return needed / providerCount, nil
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
