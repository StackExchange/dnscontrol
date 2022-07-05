package hostingde

import (
	"encoding/json"
	"fmt"
	"github.com/StackExchange/dnscontrol/v3/internal/dnscontrol"
	"sort"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/providers"
)

var defaultNameservers = []string{"ns1.hosting.de.", "ns2.hosting.de.", "ns3.hosting.de."}

var features = providers.DocumentationNotes{
	providers.CanAutoDNSSEC:          providers.Unimplemented("Supported but not implemented yet."),
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
	providers.DocCreateDomains:       providers.Can(),
	providers.DocDualHost:            providers.Can(),
	providers.DocOfficiallySupported: providers.Cannot(),
}

func init() {
	providers.RegisterRegistrarType("HOSTINGDE", newHostingdeReg)
	fns := providers.DspFuncs{
		Initializer:   newHostingdeDsp,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType("HOSTINGDE", fns, features)
}

type providerMeta struct {
	DefaultNS []string `json:"default_ns"`
}

func newHostingde(m map[string]string, providermeta json.RawMessage) (*hostingdeProvider, error) {
	authToken, ownerAccountID, baseURL := m["authToken"], m["ownerAccountId"], m["baseURL"]

	if authToken == "" {
		return nil, fmt.Errorf("hosting.de: authtoken must be provided")
	}

	if baseURL == "" {
		baseURL = "https://secure.hosting.de"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	hp := &hostingdeProvider{
		authToken:      authToken,
		ownerAccountID: ownerAccountID,
		baseURL:        baseURL,
		nameservers:    defaultNameservers,
	}

	if len(providermeta) > 0 {
		var pm providerMeta
		if err := json.Unmarshal(providermeta, &pm); err != nil {
			return nil, fmt.Errorf("hosting.de: could not parse providermeta: %w", err)
		}

		if len(pm.DefaultNS) > 0 {
			hp.nameservers = pm.DefaultNS
		}
	}

	return hp, nil
}

func newHostingdeDsp(m map[string]string, providermeta json.RawMessage) (providers.DNSServiceProvider, error) {
	return newHostingde(m, providermeta)
}

func newHostingdeReg(m map[string]string) (providers.Registrar, error) {
	return newHostingde(m, json.RawMessage{})
}

func (hp *hostingdeProvider) GetNameservers(_ dnscontrol.Context, domain string) ([]*models.Nameserver, error) {
	return models.ToNameserversStripTD(hp.nameservers)
}

func (hp *hostingdeProvider) GetZoneRecords(_ dnscontrol.Context, domain string) (models.Records, error) {
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

func (hp *hostingdeProvider) GetDomainCorrections(ctx dnscontrol.Context, dc *models.DomainConfig) ([]*models.Correction, error) {
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

	records, err := hp.GetZoneRecords(ctx, dc.Name)
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

func (hp *hostingdeProvider) GetRegistrarCorrections(_ dnscontrol.Context, dc *models.DomainConfig) ([]*models.Correction, error) {
	err := dc.Punycode()
	if err != nil {
		return nil, err
	}

	found, err := hp.getNameservers(dc.Name)
	if err != nil {
		return nil, fmt.Errorf("error getting nameservers: %w", err)
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

func (hp *hostingdeProvider) EnsureDomainExists(_ dnscontrol.Context, domain string) error {
	_, err := hp.getZoneConfig(domain)
	if err == errZoneNotFound {
		if err := hp.createZone(domain); err != nil {
			return err
		}
	}
	return nil
}
