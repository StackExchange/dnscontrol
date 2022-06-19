package powerdns

import (
	"encoding/json"
	"fmt"
	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/providers"
	pdns "github.com/mittwald/go-powerdns"
)

var features = providers.DocumentationNotes{
	providers.CanAutoDNSSEC:          providers.Can(),
	providers.CanGetZones:            providers.Can(),
	providers.CanUseAlias:            providers.Can("Needs to be enabled in PowerDNS first", "https://doc.powerdns.com/authoritative/guides/alias.html"),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseDS:               providers.Can(),
	providers.CanUseNAPTR:            providers.Can(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseSSHFP:            providers.Can(),
	providers.CanUseTLSA:             providers.Can(),
	providers.DocCreateDomains:       providers.Can(),
	providers.DocDualHost:            providers.Can(),
	providers.DocOfficiallySupported: providers.Cannot(),
}

func init() {
	fns := providers.DspFuncs{
		Initializer:   NewProvider,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType("POWERDNS", fns, features)
}

// powerdnsProvider represents the powerdnsProvider DNSServiceProvider.
type powerdnsProvider struct {
	client         pdns.Client
	APIKey         string
	APIUrl         string
	ServerName     string
	DefaultNS      []string `json:"default_ns"`
	DNSSecOnCreate bool     `json:"dnssec_on_create"`

	nameservers []*models.Nameserver
}

// NewProvider initializes a PowerDNS DNSServiceProvider.
func NewProvider(m map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	api := &powerdnsProvider{}

	api.APIKey = m["apiKey"]
	if api.APIKey == "" {
		return nil, fmt.Errorf("PowerDNS API Key is required")
	}

	api.APIUrl = m["apiUrl"]
	if api.APIUrl == "" {
		return nil, fmt.Errorf("PowerDNS API URL is required")
	}

	api.ServerName = m["serverName"]
	if api.ServerName == "" {
		return nil, fmt.Errorf("PowerDNS server name is required")
	}

	// load js config
	if len(metadata) != 0 {
		err := json.Unmarshal(metadata, api)
		if err != nil {
			return nil, err
		}
	}
	var nss []string
	for _, ns := range api.DefaultNS {
		nss = append(nss, ns[0:len(ns)-1])
	}
	var err error
	api.nameservers, err = models.ToNameservers(nss)
	if err != nil {
		return api, err
	}

	var clientErr error
	api.client, clientErr = pdns.New(
		pdns.WithBaseURL(api.APIUrl),
		pdns.WithAPIKeyAuthentication(api.APIKey),
	)
	return api, clientErr
}
