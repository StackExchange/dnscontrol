package powerdns

import (
	"encoding/json"
	"fmt"

	"github.com/mittwald/go-powerdns/apis/zones"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/providers"
	pdns "github.com/mittwald/go-powerdns"
)

var features = providers.DocumentationNotes{
	// The default for unlisted capabilities is 'Cannot'.
	// See providers/capabilities.go for the entire list of capabilities.
	providers.CanAutoDNSSEC:          providers.Can(),
	providers.CanGetZones:            providers.Can(),
	providers.CanConcur:              providers.Cannot(),
	providers.CanUseAlias:            providers.Can("Needs to be enabled in PowerDNS first", "https://doc.powerdns.com/authoritative/guides/alias.html"),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseDS:               providers.Can(),
	providers.CanUseDHCID:            providers.Can(),
	providers.CanUseLOC:              providers.Unimplemented("Normalization within the PowerDNS API seems to be buggy, so disabled", "https://github.com/PowerDNS/pdns/issues/10558"),
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
		Initializer:   newDSP,
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
	DefaultNS      []string       `json:"default_ns"`
	DNSSecOnCreate bool           `json:"dnssec_on_create"`
	ZoneKind       zones.ZoneKind `json:"zone_kind"`
	SOAEditAPI     string         `json:"soa_edit_api,omitempty"`

	nameservers []*models.Nameserver
}

// newDSP initializes a PowerDNS DNSServiceProvider.
func newDSP(m map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	dsp := &powerdnsProvider{}

	dsp.APIKey = m["apiKey"]
	if dsp.APIKey == "" {
		return nil, fmt.Errorf("PowerDNS API Key is required")
	}

	dsp.APIUrl = m["apiUrl"]
	if dsp.APIUrl == "" {
		return nil, fmt.Errorf("PowerDNS API URL is required")
	}

	dsp.ServerName = m["serverName"]
	if dsp.ServerName == "" {
		return nil, fmt.Errorf("PowerDNS server name is required")
	}

	// load js config
	if len(metadata) != 0 {
		err := json.Unmarshal(metadata, dsp)
		if err != nil {
			return nil, err
		}
	}
	var nss []string
	for _, ns := range dsp.DefaultNS {
		nss = append(nss, ns[0:len(ns)-1])
	}
	var err error
	dsp.nameservers, err = models.ToNameservers(nss)
	if err != nil {
		return dsp, err
	}

	var clientErr error
	dsp.client, clientErr = pdns.New(
		pdns.WithBaseURL(dsp.APIUrl),
		pdns.WithAPIKeyAuthentication(dsp.APIKey),
	)
	return dsp, clientErr
}
