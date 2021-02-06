package httpnet

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/providers"
)

var defaultNameservers = []string{"ns1.routing.net", "ns2.routing.net", "ns3.routing.net"}

var features = providers.DocumentationNotes{
	providers.CanAutoDNSSEC:          providers.Unimplemented("Supported by http.net but not implemented yet."),
	providers.CanGetZones:            providers.Can(),
	providers.CanUseAlias:            providers.Can(),
	providers.CanUseAzureAlias:       providers.Cannot(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseDS:               providers.Can(),
	providers.CanUseNAPTR:            providers.Cannot(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseRoute53Alias:     providers.Cannot(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseSSHFP:            providers.Can(),
	providers.CanUseTLSA:             providers.Can(),
	providers.CanUseTXTMulti:         providers.Can(),
	providers.DocCreateDomains:       providers.Can(),
	providers.DocDualHost:            providers.Can(),
	providers.DocOfficiallySupported: providers.Cannot(),
}

func init() {
	providers.RegisterRegistrarType("HTTPNET", newHttpnetReg)
	providers.RegisterDomainServiceProviderType("HTTPNET", newHttpnetDsp, features)
}

func newHttpnet(m map[string]string) (*httpnetProvider, error) {
	authToken, ownerAccountID := m["authToken"], m["ownerAccountID"]

	if authToken == "" {
		return nil, fmt.Errorf("http.net: authtoken must be provided")
	}

	hp := &httpnetProvider{
		authToken:      authToken,
		ownerAccountID: ownerAccountID,
	}

	return hp, nil
}

func newHttpnetDsp(m map[string]string, raw json.RawMessage) (providers.DNSServiceProvider, error) {
	return newHttpnet(m)
}

func newHttpnetReg(m map[string]string) (providers.Registrar, error) {
	return newHttpnet(m)
}

func (hp *httpnetProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	return models.ToNameservers(defaultNameservers)
}

func (hp *httpnetProvider) GetZoneRecords(domain string) (models.Records, error) {
	src, err := hp.getRecords(domain)
	if err != nil {
		return nil, err
	}

	records := []*models.RecordConfig{}
	for _, r := range src {
		if r.Type == "SOA" {
			continue
		}
		records = append(records, r.nativeToRecord(domain))
	}

	return records, nil
}

func (hp *httpnetProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	err := dc.Punycode()
	if err != nil {
		return nil, err
	}

	// TTL must be between (inclusive) 1m and 1y (in fact, a little bit more)
	for _, r := range dc.Records {
		if r.TTL < 60 {
			r.TTL = 60
		}
		if r.TTL > 31556926 {
			r.TTL = 31556926
		}
	}

	records, err := hp.GetZoneRecords(dc.Name)
	if err != nil {
		return nil, err
	}

	differ := diff.New(dc)
	_, create, del, mod, err := differ.IncrementalDiff(records)
	if err != nil {
		return nil, err
	}

	// NOPURGE
	if dc.KeepUnknown {
		del = []diff.Correlation{}
	}

	msg := []string{}
	for _, c := range append(del, append(create, mod...)...) {
		msg = append(msg, c.String())
	}

	if len(create) == 0 && len(del) == 0 && len(mod) == 0 {
		return nil, nil
	}

	corrections := []*models.Correction{
		{
			Msg: fmt.Sprintf("\n%s", strings.Join(msg, "\n")),
			F: func() error {
				return hp.updateRecords(dc.Name, create, del, mod)
			},
		},
	}

	return corrections, nil
}

func (hp *httpnetProvider) GetRegistrarCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	err := dc.Punycode()
	if err != nil {
		return nil, err
	}

	found, err := hp.getNameservers(dc.Name)
	if err != nil {
		return nil, fmt.Errorf("error getting nameservers: %v", err)
	}
	sort.Strings(found)
	foundNameservers := strings.Join(found, ",")

	expected := []string{}
	for _, ns := range dc.Nameservers {
		expected = append(expected, ns.Name)
	}
	sort.Strings(expected)
	expectedNameservers := strings.Join(expected, ",")

	// We don't care about glued records because we disallowed them
	if foundNameservers != expectedNameservers {
		return []*models.Correction{
			{
				Msg: fmt.Sprintf("Update nameservers %s -> %s", foundNameservers, expectedNameservers),
				F:   hp.updateNameservers(expected, dc.Name),
			},
		}, nil
	}

	return nil, nil

	// TODO: Handle AutoDNSSEC
}

func (hp *httpnetProvider) EnsureDomainExists(domain string) error {
	_, err := hp.getZoneConfig(domain)
	if err == errZoneNotFound {
		if err := hp.createZone(domain); err != nil {
			return err
		}
	}
	return err
}
