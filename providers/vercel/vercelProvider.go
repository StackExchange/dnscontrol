package vercel

/*
Vercel DNS provider (vercel.com)

Info required in `creds.json`:
	- team_id
	- api_token
*/

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/providers"
	"github.com/miekg/dns"
	vercelClient "github.com/vercel/terraform-provider-vercel/client"
)

var defaultNameservers = []string{
	"ns1.vercel-dns.com",
	"ns2.vercel-dns.com",
}

var features = providers.DocumentationNotes{
	// The default for unlisted capabilities is 'Cannot'.
	// See providers/capabilities.go for the entire list of capabilities.
	providers.CanAutoDNSSEC:          providers.Cannot(),
	providers.CanGetZones:            providers.Cannot(),
	providers.CanConcur:              providers.Unimplemented(),
	providers.CanUseAlias:            providers.Can(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseDHCID:            providers.Cannot(),
	providers.CanUseDS:               providers.Cannot(),
	providers.CanUseDSForChildren:    providers.Cannot(),
	providers.CanUseLOC:              providers.Cannot(),
	providers.CanUseNAPTR:            providers.Cannot(),
	providers.CanUsePTR:              providers.Cannot(),
	providers.CanUseSOA:              providers.Cannot(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseSSHFP:            providers.Cannot(),
	providers.CanUseTLSA:             providers.Cannot(),
	providers.DocCreateDomains:       providers.Unimplemented(),
	providers.DocDualHost:            providers.Cannot("Vercel does not allow sufficient control over the apex NS records"),
	providers.DocOfficiallySupported: providers.Cannot(),
}

// hednsProvider stores login credentials and represents and API connection
type vercelProvider struct {
	client   vercelClient.Client
	apiToken string
	teamID   string
}

// uint16Zero converts value to uint16 or returns 0.
func uint16Zero(value interface{}) uint16 {
	switch v := value.(type) {
	case float64:
		return uint16(v)
	case uint16:
		return v
	case nil:
	}
	return 0
}

func init() {
	const providerName = "Vercel"
	const providerMaintainer = "@SukkaW"
	fns := providers.DspFuncs{
		Initializer: newProvider,
	}
	providers.RegisterDomainServiceProviderType(providerName, fns, providers.CanUseSRV, features)
	providers.RegisterMaintainer(providerName, providerMaintainer)
}

func newProvider(creds map[string]string, meta json.RawMessage) (providers.DNSServiceProvider, error) {
	if creds["team_id"] == "" || creds["api_token"] == "" {
		return nil, errors.New("api_token required for ns1")
	}

	c := vercelClient.New(
		creds["api_token"],
	)

	ctx := context.Background()

	team, err := c.Team(ctx, creds["team_id"])
	if err != nil {
		return nil, err
	}

	c = c.WithTeam(team)
	return &vercelProvider{
		client:   *c,
		apiToken: creds["api_token"],
		// store this information so that we can access this anywhere we want
		teamID: creds["team_id"],
	}, nil
}

// GetNameservers returns the default Vercel nameservers.
// Though Vercel RESTful API supports getting "intendedNameServers", but it is not implemented in the Go SDK
// Let's hard-coded this for now
func (c *vercelProvider) GetNameservers(_ string) ([]*models.Nameserver, error) {
	return models.ToNameservers(defaultNameservers)
}

func (c *vercelProvider) GetZoneRecords(domain string, meta map[string]string) (models.Records, error) {
	var zoneRecords []*models.RecordConfig

	records, err := c.listDNSRecords(domain)
	if err != nil {
		return nil, err
	}

	for _, r := range records {
		rc := &models.RecordConfig{
			TTL:      uint32(r.TTL),
			Original: r,
		}
		rc.SetLabel(r.Name, domain)

		if r.Type == "CNAME" || r.Type == "MX" {
			r.Value = dns.CanonicalName(r.Value)
		}

		switch rtype := r.RecordType; rtype {
		case "MX":
			if err := rc.SetTargetMX(uint16Zero(r.MXPriority), r.Value); err != nil {
				return nil, fmt.Errorf("unparsable MX record: %w", err)
			}
		case "SRV":
			// Vercel's API doesn't always return SRV as an SRV object.
			// It might return priority in the json field, and the srv as a big string `[weight] [port] [domain]` in json 'value' field.
			// We have to create our own string before passing in.
			// Fallback to parsing from string if SRV object is missing
			// r.Value is "weight port target", we need "priority weight port target"
			if err := rc.PopulateFromString(
				rtype,
				fmt.Sprintf("%d %s", uint16Zero(r.Priority), r.Value),
				domain,
			); err != nil {
				return nil, fmt.Errorf("unparsable SRV record from value: %w", err)
			}
		case "HTTPS":
			// Vercel returns priority in a separate field, and value contains "target params".
			// We need to combine them for PopulateFromString.
			if err := rc.PopulateFromString(
				rtype,
				fmt.Sprintf("%d %s", uint16Zero(r.Priority), r.Value),
				domain,
			); err != nil {
				return nil, fmt.Errorf("unparsable HTTPS record: %w", err)
			}
		case "TXT":
			err := rc.SetTargetTXT(r.Value)
			if err != nil {
				return nil, fmt.Errorf("unparsable TXT record: %w", err)
			}
		default:
			if err := rc.PopulateFromString(rtype, r.Value, domain); err != nil {
				return nil, fmt.Errorf("unparsable record received from vercel: %w", err)
			}
		}

		zoneRecords = append(zoneRecords, rc)
	}

	return zoneRecords, nil
}

func (c *vercelProvider) GetZoneRecordsCorrections(dc *models.DomainConfig, records models.Records) ([]*models.Correction, int, error) {
	return nil, 0, nil
}
