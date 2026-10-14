package bunnydns

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/DNSControl/dnscontrol/v5/models"
	"github.com/DNSControl/dnscontrol/v5/pkg/providers"
)

type bunnydnsProvider struct {
	apiKey string
	// zonesMu guards zones, which is filled on first use and shared by every zone being processed.
	zonesMu  sync.Mutex
	zones    map[string]*zone
	observer providers.ConversionObserver
}

func (b *bunnydnsProvider) SetConversionObserver(observer providers.ConversionObserver) {
	b.observer = observer
}

func init() {
	providers.Register[*bunnydnsProvider]("BUNNY_DNS", providers.Definition{
		FriendlyName: "Bunny DNS",
		Maintainer:   "@ppmathis",
		SupportedTypes: []string{
			"Basic8",
			"ALIAS:Can:Bunny flattens CNAME records into A/AAAA records dynamically",
			"BUNNY_DNS_PZ",
			"BUNNY_DNS_RDR",
			"HTTPS",
			"PTR",
			"SVCB",
			"TLSA",
		},
		OfficiallySupported: providers.Cannot(),
		CanAutoDNSSEC:       providers.Can(),
		CanConcur:           providers.Can(),
		CanDualHost:         providers.Cannot(),
		CanUseDSForChildren: providers.Cannot(),
		CredFields: []providers.CredsField{
			{
				Key:      "api_key",
				Label:    "API key",
				Help:     "Bunny.net account API key.",
				Secret:   true,
				Required: true,
			},
		},
		DocsURL:   "https://docs.dnscontrol.org/provider/bunnydns",
		PortalURL: "https://dash.bunny.net/account/api-key",
	})
}

// Initialize initializes a fresh provider instance.
func (b *bunnydnsProvider) Initialize(settings map[string]string, _ json.RawMessage, options *providers.CreateOptions) error {
	apiKey := settings["api_key"]
	if apiKey == "" {
		return errors.New("missing BUNNY_DNS api_key")
	}

	*b = bunnydnsProvider{
		apiKey: apiKey,
	}
	b.SetConversionObserver(options.WithDefaults().ConversionObserver)
	return nil
}

// GetNameservers returns the nameservers Bunny DNS serves the zone from, so that a registrar can be pointed at them.
// DNSControl also turns them into apex NS records, which Bunny DNS does not permit; removeApexNS drops those again.
func (b *bunnydnsProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	zone, err := b.findZoneByDomain(domain)
	if err != nil {
		return nil, err
	}

	return models.ToNameservers(zone.Nameservers())
}
