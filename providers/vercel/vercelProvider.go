package vercel

/*
Vercel DNS provider (vercel.com)

Info required in `creds.json`:
	- account_id
	- api_token
*/

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/providers"
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
	providers.DocDualHost:            providers.Cannot(),
	providers.DocOfficiallySupported: providers.Cannot(),
}

// hednsProvider stores login credentials and represents and API connection
type vercelProvider struct {
	client vercelClient.Client
}

func init() {
	const providerName = "Vercel"
	const providerMaintainer = "@SukkaW"
	fns := providers.DspFuncs{
		Initializer:   newProvider,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType(providerName, fns, providers.CanUseSRV, docNotes)
	providers.RegisterMaintainer(providerName, providerMaintainer)
}

func newProvider(creds map[string]string, meta json.RawMessage) (providers.DNSServiceProvider, error) {
	if creds["account_id"] == "" || creds["api_token"] == "" {
		return nil, errors.New("api_token required for ns1")
	}

	// Enable Sleep API Rate limit strategy - it will sleep until new tokens are available
	// see https://help.ns1.com/hc/en-us/articles/360020250573-About-API-rate-limiting
	// this strategy would imply the least sleep time for non-parallel client requests
	c := vercelClient.New(
		creds["api_token"],
	)

	ctx := context.Background()

	team, err := c.Team(ctx, creds["account_id"])
	if err != nil {
		return nil, err
	}

	c = c.WithTeam(team)
	return &vercelProvider{client: c}, nil
}

// GetNameservers returns the default Vercel nameservers.
// Though Vercel RESTful API supports getting "intendedNameServers", but it is not implemented in the Go SDK
// Let's hard-coded this for now
func (c *vercelProvider) GetNameservers(_ string) ([]*models.Nameserver, error) {
	return models.ToNameservers(defaultNameservers)
}
