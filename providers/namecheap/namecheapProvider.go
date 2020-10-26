package namecheap

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	nc "github.com/billputer/go-namecheap"
	"golang.org/x/net/publicsuffix"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/pkg/printer"
	"github.com/StackExchange/dnscontrol/v3/providers"
)

// NamecheapDefaultNs lists the default nameservers for this provider.
var NamecheapDefaultNs = []string{"dns1.registrar-servers.com", "dns2.registrar-servers.com"}

// namecheapProvider is the handle for this provider.
type namecheapProvider struct {
	APIKEY  string
	APIUser string
	client  *nc.Client
}

var features = providers.DocumentationNotes{
	providers.CanUseAlias:            providers.Cannot(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUsePTR:              providers.Cannot(),
	providers.CanUseSRV:              providers.Cannot("The namecheap web console allows you to make SRV records, but their api does not let you read or set them"),
	providers.CanUseTLSA:             providers.Cannot(),
	providers.CantUseNOPURGE:         providers.Cannot(),
	providers.DocCreateDomains:       providers.Cannot("Requires domain registered through their service"),
	providers.DocDualHost:            providers.Cannot("Doesn't allow control of apex NS records"),
	providers.DocOfficiallySupported: providers.Cannot(),
	providers.CanGetZones:            providers.Unimplemented(),
}

func init() {
	providers.RegisterRegistrarType("NAMECHEAP", newReg)
	providers.RegisterDomainServiceProviderType("NAMECHEAP", newDsp, features)
	providers.RegisterCustomRecordType("URL", "NAMECHEAP", "")
	providers.RegisterCustomRecordType("URL301", "NAMECHEAP", "")
	providers.RegisterCustomRecordType("FRAME", "NAMECHEAP", "")
}

func newDsp(conf map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	return newProvider(conf, metadata)
}

func newReg(conf map[string]string) (providers.Registrar, error) {
	return newProvider(conf, nil)
}

func newProvider(m map[string]string, metadata json.RawMessage) (*namecheapProvider, error) {
	api := &namecheapProvider{}
	api.APIUser, api.APIKEY = m["apiuser"], m["apikey"]
	if api.APIKEY == "" || api.APIUser == "" {
		return nil, fmt.Errorf("missing Namecheap apikey and apiuser")
	}
	api.client = nc.NewClient(api.APIUser, api.APIKEY, api.APIUser)
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
// from support in SEP-2017:
//    "The limits for the API calls will be 20/Min, 700/Hour and 8000/Day for one user.
//     If you can limit the requests within these it should be fine."
// this helper performs some api action, checks for rate limited response, and if so, enters a retry loop until it resolves
// if you are consistently hitting this, you may have success asking their support to increase your account's limits.
func doWithRetry(f func() error) {
	// sleep 5 seconds at a time, up to 23 times (1 minute, 15 seconds)
	const maxRetries = 23
	const sleepTime = 5 * time.Second
	var currentRetry int
	for {
		err := f()
		if err == nil {
			return
		}
		if strings.Contains(err.Error(), "Error 500000: Too many requests") {
			currentRetry++
			if currentRetry >= maxRetries {
				return
			}
			printer.Printf("Namecheap rate limit exceeded. Waiting %s to retry.\n", sleepTime)
			time.Sleep(sleepTime)
		} else {
			return
		}
	}
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (n *namecheapProvider) GetZoneRecords(domain string) (models.Records, error) {
	return nil, fmt.Errorf("not implemented")
	// This enables the get-zones subcommand.
	// Implement this by extracting the code from GetDomainCorrections into
	// a single function.  For most providers this should be relatively easy.
}

// GetDomainCorrections returns the corrections for the domain.
func (n *namecheapProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	dc.Punycode()
	sld, tld := splitDomain(dc.Name)
	var records *nc.DomainDNSGetHostsResult
	var err error
	doWithRetry(func() error {
		records, err = n.client.DomainsDNSGetHosts(sld, tld)
		return err
	})
	if err != nil {
		return nil, err
	}

	var actual []*models.RecordConfig

	// namecheap does not allow setting @ NS with basic DNS
	dc.Filter(func(r *models.RecordConfig) bool {
		if r.Type == "NS" && r.GetLabel() == "@" {
			if !strings.HasSuffix(r.GetTargetField(), "registrar-servers.com.") {
				fmt.Println("\n", r.GetTargetField(), "Namecheap does not support changing apex NS records. Skipping.")
			}
			return false
		}
		return true
	})

	// namecheap has this really annoying feature where they add some parking records if you have no records.
	// This causes a few problems for our purposes, specifically the integration tests.
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
			Type:         r.Type,
			TTL:          uint32(r.TTL),
			MxPreference: uint16(r.MXPref),
			Original:     r,
		}
		rec.SetLabel(r.Name, dc.Name)
		switch rtype := r.Type; rtype { // #rtype_variations
		case "TXT":
			rec.SetTargetTXT(r.Address)
		case "CAA":
			rec.SetTargetCAAString(r.Address)
		default:
			rec.SetTarget(r.Address)
		}
		actual = append(actual, rec)
	}

	// Normalize
	models.PostProcessRecords(actual)

	differ := diff.New(dc)
	_, create, delete, modify, err := differ.IncrementalDiff(actual)
	if err != nil {
		return nil, err
	}

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
					return n.generateRecords(dc)
				},
			})
	}

	return corrections, nil
}

func (n *namecheapProvider) generateRecords(dc *models.DomainConfig) error {

	var recs []nc.DomainDNSHost

	id := 1
	for _, r := range dc.Records {
		var value string
		switch rtype := r.Type; rtype { // #rtype_variations
		case "CAA":
			value = r.GetTargetCombined()
		default:
			value = r.GetTargetField()
		}

		rec := nc.DomainDNSHost{
			ID:      id,
			Name:    r.GetLabel(),
			Type:    r.Type,
			Address: value,
			MXPref:  int(r.MxPreference),
			TTL:     int(r.TTL),
		}
		recs = append(recs, rec)
		id++
	}
	sld, tld := splitDomain(dc.Name)
	var err error
	doWithRetry(func() error {
		_, err = n.client.DomainDNSSetHosts(sld, tld, recs)
		return err
	})
	return err
}

// GetNameservers returns the nameservers for a domain.
func (n *namecheapProvider) GetNameservers(domainName string) ([]*models.Nameserver, error) {
	// return default namecheap nameservers
	return models.ToNameservers(NamecheapDefaultNs)
}

// GetRegistrarCorrections returns corrections to update nameservers.
func (n *namecheapProvider) GetRegistrarCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	var info *nc.DomainInfo
	var err error
	doWithRetry(func() error {
		info, err = n.client.DomainGetInfo(dc.Name)
		return err
	})
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
				F: func() (err error) {
					doWithRetry(func() error {
						_, err = n.client.DomainDNSSetCustom(sld, tld, desired)
						return err
					})
					return
				}},
		}, nil
	}
	return nil, nil
}
