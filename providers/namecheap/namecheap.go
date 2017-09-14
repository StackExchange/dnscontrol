package namecheap

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers"
	"github.com/StackExchange/dnscontrol/providers/diff"
	nc "github.com/billputer/go-namecheap"
	"github.com/miekg/dns/dnsutil"
)

var NamecheapDefaultNs = []string{"dns1.registrar-servers.com", "dns2.registrar-servers.com"}

type Namecheap struct {
	ApiKey  string
	ApiUser string
	client  *nc.Client
}

var docNotes = providers.DocumentationNotes{
	providers.DocCreateDomains:       providers.Cannot("Requires domain registered through their service"),
	providers.DocOfficiallySupported: providers.Cannot(),
}

func init() {
	providers.RegisterRegistrarType("NAMECHEAP", newReg, docNotes)
	providers.RegisterDomainServiceProviderType("NAMECHEAP", newDsp, providers.CantUseNOPURGE)
}

func newDsp(conf map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	return newProvider(conf, metadata)
}

func newReg(conf map[string]string) (providers.Registrar, error) {
	return newProvider(conf, nil)
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

func splitDomain(domain string) (sld string, tld string) {
	tld, _ = publicsuffix.PublicSuffix(domain)
	d, _ := publicsuffix.EffectiveTLDPlusOne(domain)
	sld = strings.Split(d, ".")[0]
	return sld, tld
}

// namecheap has request limiting at unpublished limits
// this channel acts as a global rate limiter
// read from it before every request
var throttle = time.NewTicker(time.Second * 5).C

func (n *Namecheap) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	dc.Punycode()
	sld, tld := splitDomain(dc.Name)
	<-throttle
	records, err := n.client.DomainsDNSGetHosts(sld, tld)
	if err != nil {
		return nil, err
	}

	var actual []*models.RecordConfig

	// namecheap does not allow setting @ NS with basic DNS
	dc.Filter(func(r *models.RecordConfig) bool {
		if r.Type == "NS" && r.Name == "@" {
			if !strings.HasSuffix(r.Target, "registrar-servers.com.") {
				fmt.Println("\n", r.Target, "Namecheap does not support changing apex NS records. Skipping.")
			}
			return false
		}
		return true
	})

	// namecheap has this really crappy feature where they add some parking records if you have no records.
	// This is really crappy for our purposes, specifically the integration tests.
	// lets detect that one case and pretend it is a no-op.
	if len(dc.Records) == 0 && len(records.Hosts) == 2 {
		if records.Hosts[0].Type == "CNAME" &&
			strings.Contains(records.Hosts[0].Address, "parkingpage") &&
			records.Hosts[1].Type == "URL" {
			return nil, nil
		}
	}

	for _, r := range records.Hosts {
		if r.Type == "SOA" {
			continue
		}
		rec := &models.RecordConfig{
			NameFQDN:     dnsutil.AddOrigin(r.Name, dc.Name),
			Type:         r.Type,
			Target:       r.Address,
			TTL:          uint32(r.TTL),
			MxPreference: uint16(r.MXPref),
			Original:     r,
		}
		actual = append(actual, rec)
	}

	differ := diff.New(dc)
	_, create, delete, modify := differ.IncrementalDiff(actual)

	// // because namecheap doesn't have selective create, delete, modify,
	// // we bundle them all up to send at once.  We *do* want to see the
	// // changes though

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
					return n.UpdateRecords(dc)
				},
			})
	}

	return corrections, nil
}

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
			MXPref:  int(r.MxPreference),
			TTL:     int(r.TTL),
		}
		recs = append(recs, rec)
		id++
	}

	sld, tld := splitDomain(dc.Name)
	<-throttle
	_, err := n.client.DomainDNSSetHosts(sld, tld, recs)

	return err
}

func (n *Namecheap) GetNameservers(domainName string) ([]*models.Nameserver, error) {
	// return default namecheap nameservers
	ns := NamecheapDefaultNs

	return models.StringsToNameservers(ns), nil
}

func (n *Namecheap) GetRegistrarCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	<-throttle
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
					<-throttle
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
