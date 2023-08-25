// Package mythicbeasts provides a provider for managing zones in Mythic Beasts.
//
// This package uses the Primary DNS API v2, as described in https://www.mythic-beasts.com/support/api/dnsv2
package mythicbeasts

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/diff2"
	"github.com/StackExchange/dnscontrol/v4/providers"
	"github.com/miekg/dns"
)

// mythicBeastsDefaultNS lists the default nameservers, per https://www.mythic-beasts.com/support/domains/nameservers.
var mythicBeastsDefaultNS = []string{
	"ns1.mythic-beasts.com",
	"ns2.mythic-beasts.com",
}

// mythicBeastsProvider is the handle for this provider.
type mythicBeastsProvider struct {
	secret string
	keyID  string
	client *http.Client
}

var features = providers.DocumentationNotes{
	providers.CanGetZones:            providers.Can(),
	providers.CanUseAlias:            providers.Cannot(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseLOC:              providers.Cannot(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseTLSA:             providers.Can(),
	providers.DocCreateDomains:       providers.Cannot("Requires domain registered through Web UI"),
	providers.DocDualHost:            providers.Can(),
	providers.DocOfficiallySupported: providers.Cannot(),
}

func init() {
	fns := providers.DspFuncs{
		Initializer:   newDsp,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType("MYTHICBEASTS", fns, features)
}

func newDsp(conf map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	if conf["keyID"] == "" {
		return nil, fmt.Errorf("missing Mythic Beasts auth keyID")
	}
	if conf["secret"] == "" {
		return nil, fmt.Errorf("missing Mythic Beasts auth secret")
	}
	return &mythicBeastsProvider{
		keyID:  conf["keyID"],
		secret: conf["secret"],
		client: &http.Client{},
	}, nil
}

func (n *mythicBeastsProvider) httpRequest(method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, "https://api.mythic-beasts.com/dns/v2"+url, body)
	if err != nil {
		return nil, err
	}
	// https://www.mythic-beasts.com/support/api/auth
	req.SetBasicAuth(n.keyID, n.secret)
	req.Header.Add("Content-Type", "text/dns")
	req.Header.Add("Accept", "text/dns")
	return n.client.Do(req)
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (n *mythicBeastsProvider) GetZoneRecords(domain string, meta map[string]string) (models.Records, error) {
	resp, err := n.httpRequest("GET", "/zones/"+domain+"/records", nil)
	if err != nil {
		return nil, err
	}
	if got, want := resp.StatusCode, 200; got != want {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("got HTTP %v, want %v: %v", got, want, string(body))
	}
	return zoneFileToRecords(resp.Body, domain)
}

func zoneFileToRecords(r io.Reader, origin string) (models.Records, error) {
	zp := dns.NewZoneParser(r, origin, origin)
	var records []*models.RecordConfig
	for rr, ok := zp.Next(); ok; rr, ok = zp.Next() {
		rec, err := models.RRtoRC(rr, origin)
		if err != nil {
			return nil, err
		}
		records = append(records, &rec)
	}

	if err := zp.Err(); err != nil {
		return nil, fmt.Errorf("parsing zone for %v: %w", origin, err)
	}
	return records, nil
}

// GetZoneRecordsCorrections returns a list of corrections that will turn existing records into dc.Records.
func (n *mythicBeastsProvider) GetZoneRecordsCorrections(dc *models.DomainConfig, actual models.Records) ([]*models.Correction, error) {
	msgs, changes, err := diff2.ByZone(actual, dc, nil)
	if err != nil {
		return nil, err
	}

	var corrections []*models.Correction
	if changes {
		corrections = append(corrections,
			&models.Correction{
				Msg: strings.Join(msgs, "\n"),
				F: func() error {
					var b strings.Builder
					for _, rr := range dc.Records {
						fmt.Fprintf(&b, "%v\n", rr.ToRR().String())
					}

					resp, err := n.httpRequest("PUT", "/zones/"+dc.Name+"/records", strings.NewReader(b.String()))
					if err != nil {
						return err
					}
					if got, want := resp.StatusCode, 200; got != want {
						body, _ := ioutil.ReadAll(resp.Body)
						return fmt.Errorf("got HTTP %v, want %v: %v", got, want, string(body))
					}
					return nil
				},
			})
	}

	return corrections, nil
}

// GetNameservers returns the nameservers for a domain.
func (n *mythicBeastsProvider) GetNameservers(domainName string) ([]*models.Nameserver, error) {
	return models.ToNameservers(mythicBeastsDefaultNS)
}
