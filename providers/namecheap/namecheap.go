package namecheap

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers"
	"github.com/StackExchange/dnscontrol/providers/diff"
	nc "github.com/billputer/go-namecheap"
	"github.com/miekg/dns/dnsutil"
	ps "golang.org/x/net/publicsuffix"
)

var NamecheapDefaultNs = []string{"dns1.registrar-servers.com", "dns2.registrar-servers.com"}

type Namecheap struct {
	ApiKey  string
	ApiUser string
	client  *nc.Client
}

func init() {
	providers.RegisterRegistrarType("NAMECHEAP", newReg)
	providers.RegisterDomainServiceProviderType("NAMECHEAP", newDsp, providers.CanUseURL)
}

func newDsp(conf map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	return newProvider(conf, metadata)
}

func newReg(conf map[string]string) (providers.Registrar, error) {
	return newProvider(conf, nil)
}

func splitDomain(domain string) (sld string, tld string) {
	tld, _ = ps.PublicSuffix(domain)
	d, _ := ps.EffectiveTLDPlusOne(domain)
	sld = strings.Split(d, ".")[0]
	return sld, tld
}

func (n *Namecheap) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {

	sld, tld := splitDomain(dc.Name)
	records, err := n.client.DomainsDNSGetHosts(sld, tld)
	if err != nil {
		return nil, err
	}

	var actual []*models.RecordConfig

	// namecheap does not allow setting @ NS with basic DNS
	dc.Filter(func(r *models.RecordConfig) bool {
		if r.Type == "NS" && r.Name == "@" {
			//check to make sure it looks like a namecheap server and print a warning if it does not.
			nsIsDefault := false
			for _, n := range NamecheapDefaultNs {
				if n+"." == r.Target {
					nsIsDefault = true
				}

			}
			if !nsIsDefault {
				fmt.Println("\n", r.Target, "is not a default NS, Namecheap will reject this")
			}
			return false
		}
		return true
	})

	for _, r := range records.Hosts {
		if r.Type == "SOA" {
			continue
		}
		rec := &models.RecordConfig{
			NameFQDN: dnsutil.AddOrigin(r.Name, dc.Name),
			Type:     r.Type,
			Target:   r.Address,
			TTL:      uint32(r.TTL),
			Priority: uint16(r.MXPref),
			Original: r,
		}
		actual = append(actual, rec)

	}

	differ := diff.New(dc)
	_, create, delete, modify := differ.IncrementalDiff(actual)

	// because namecheap doesn't have selective create, delete, modify,
	// we bundle them all up to send at once.  We *do* want to see the
	// changes though

	var desc []string
	for _, i := range create {
		desc = append(desc, "\n"+i.String())
	}
	for _, i := range delete {
		desc = append(desc, "\n"+i.String())
	}
	for _, i := range modify {
		desc = append(desc, "\n"+i.String())
	}

	msg := fmt.Sprintf("GENERATE_ZONE: %s (%d records)%s", dc.Name, len(dc.Records), desc)
	corrections := []*models.Correction{}

	// only create corrections if there are changes
	if len(desc) > 0 {
		corrections = append(corrections,
			&models.Correction{
				Msg: msg,
				F: func() error {
					fmt.Printf("RECREATING ZONE: %v\n", dc.Name)
					return n.UpdateRecords(dc)
				},
			})
	}

	return corrections, nil
}

func (n *Namecheap) GetNameservers(domainName string) ([]*models.Nameserver, error) {
	// return default namecheap nameservers
	ns := NamecheapDefaultNs

	return models.StringsToNameservers(ns), nil
}

// UpdateRecords bundles up the expected zone and sends it wholesale to
// namecheap.
func (n *Namecheap) UpdateRecords(dc *models.DomainConfig) error {

	var recs []nc.DomainDNSHost

	id := 1
	for _, r := range dc.Records {
		name := dnsutil.TrimDomainName(r.NameFQDN, dc.Name)

		rec := nc.DomainDNSHost{
			ID:      id,
			Name:    name,
			Type:    r.Type,
			Address: r.Target,
			MXPref:  int(r.Priority),
			TTL:     int(r.TTL),
		}
		recs = append(recs, rec)
		id++
	}

	sld, tld := splitDomain(dc.Name)
	_, err := n.client.DomainDNSSetHosts(sld, tld, recs)

	return err
}

func newProvider(m map[string]string, metadata json.RawMessage) (*Namecheap, error) {
	api := &Namecheap{}
	api.ApiUser, api.ApiKey = m["apiuser"], m["apikey"]
	if api.ApiKey == "" || api.ApiUser == "" {
		return nil, fmt.Errorf("Namecheap apikey and apiuser must be provided.")
	}
	api.client = nc.NewClient(api.ApiUser, api.ApiKey, api.ApiUser)

	// if BaseURL is specified in creds, use that url
	BaseURL, ok := m["BaseURL"]
	if ok {
		api.client.BaseURL = BaseURL
	}

	return api, nil
}

func (n *Namecheap) GetRegistrarCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	info, err := n.client.DomainGetInfo(dc.Name)
	if err != nil {
		return nil, err
	}
	sort.Strings(info.DNSDetails.Nameservers)
	found := strings.Join(info.DNSDetails.Nameservers, ",")
	desiredNs := []string{}
	for _, d := range dc.Nameservers {
		desiredNs = append(desiredNs, d.Name)
	}
	sort.Strings(desiredNs)
	desired := strings.Join(desiredNs, ",")
	if found != desired {
		parts := strings.SplitN(dc.Name, ".", 2)
		sld, tld := parts[0], parts[1]
		return []*models.Correction{
			{
				Msg: fmt.Sprintf("Change Nameservers from '%s' to '%s'", found, desired),
				F: func() error {
					_, err := n.client.DomainDNSSetCustom(sld, tld, desired)
					if err != nil {
						return err
					}
					return nil
				}},
		}, nil
	}
	return nil, nil
}
