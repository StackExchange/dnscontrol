package doh

import (
	"github.com/StackExchange/dnscontrol/v3/pkg/printer"
	"sort"

	"github.com/babolivier/go-doh-client"
)

type dohProvider struct {
	host string
}

func (c *dohProvider) getNameservers(domain string) ([]string, error) {
	resolver := doh.Resolver{
		Host:  c.host,
		Class: doh.IN,
	}

	// Perform a NS lookup
	nss, _, err := resolver.LookupNS(domain)
	if err != nil {
		return nil, printer.Errorf("failed fetching nameservers list (DNS-over-HTTPS): %s", err)
	}

	ns := []string{}
	for _, res := range nss {
		ns = append(ns, res.Host)
	}
	sort.Strings(ns)
	return ns, nil
}

func (c *dohProvider) updateNameservers(ns []string, domain string) error {
	return printer.Errorf("DNS-over-HTTPS 'Registrar' is read only, changes must be applied to %s manually", domain)
}
