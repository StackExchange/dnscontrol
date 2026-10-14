package misaka

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/DNSControl/dnscontrol/v5/models"
	"github.com/DNSControl/dnscontrol/v5/pkg/printer"
	"github.com/DNSControl/dnscontrol/v5/pkg/providers"
)

/*
Misaka.io DNS provider:

Info required in `creds.json`:
   - api_token

The v3 API addresses a record set ({name}/{type}) as a unit: a PUT replaces the
whole value list and the TTL belongs to the set. This provider therefore uses
diff2.ByRecordSet() and sends one request per changed record set.

The SOA record and the NS records at the zone apex are maintained by Misaka.io,
so they are left out of both the zone contents and the desired state. Record
sets that use features DNSControl does not model (GeoDNS filters, alias,
disabled or protected sets) make the zone read fail instead of being
overwritten. See records.go.
*/

type misakaProvider struct {
	client   *apiClient
	observer providers.ConversionObserver
}

func init() {
	providers.Register[*misakaProvider]("MISAKA", providers.Definition{
		FriendlyName:    "Misaka.io",
		VendorAPIDocURL: "https://docs.misaka.io/api",
		PortalURL:       "https://app.misaka.io/account/api",
		Notes:           "The token's team must have the DNS service enabled. The Misaka.io v3 API is labelled alpha by the vendor.",
		CredFields: []providers.CredsField{
			{
				Key:      "api_token",
				Label:    "API token",
				Help:     "A Misaka.io API token of a team that has the DNS service enabled.",
				Secret:   true,
				Required: true,
			},
		},
		Maintainer: "@ZeroClover",
		SupportedTypes: []string{
			"Basic8",
			"ALIAS",
			"DNAME",
			"PTR",
			"SSHFP",
			"TLSA",
		},
		CanAutoDNSSEC:          providers.Unimplemented("DNSSEC endpoints exist; not wired up yet"),
		CanConcur:              providers.Can(),
		DocDualHost:            providers.Cannot("The NS records at the zone apex are managed by Misaka.io"),
		DocOfficiallySupported: providers.Cannot(),
		// Features retains annotations for record types and interface-derived facts.
		Features: providers.DocumentationNotes{
			providers.CanUseSOA: providers.Cannot("The SOA record is managed by Misaka.io (zone settings)"),
		},
	})
}

// Initialize initializes a fresh provider instance.
func (p *misakaProvider) Initialize(settings map[string]string, _ json.RawMessage, options *providers.CreateOptions) error {
	token := settings["api_token"]
	if token == "" {
		return errors.New("missing MISAKA api_token")
	}

	*p = misakaProvider{client: newAPIClient(defaultAPIURL, token)}
	p.SetConversionObserver(options.WithDefaults().ConversionObserver)
	return nil
}

// SetConversionObserver lets the integration tests record conversions for the golden files.
func (p *misakaProvider) SetConversionObserver(observer providers.ConversionObserver) {
	p.observer = observer
}

// GetNameservers returns the nameservers of the zone, read from its apex NS
// record set.
//
// preview and push ask for the nameservers before they check whether a zone
// still has to be created, so a missing zone is a warning, not an error.
func (p *misakaProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	rs, err := p.client.getRecordSet(domain, apexLabel, "NS")
	if isNotFound(err) {
		printer.Warnf("%v (reporting no nameservers for %s)\n", err, domain)
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(rs.Records))
	for _, r := range rs.Records {
		// The API stores the value as it was entered, with or without the
		// trailing dot.
		names = append(names, strings.TrimSuffix(r.Value, "."))
	}
	return models.ToNameservers(names)
}
