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
	"github.com/StackExchange/dnscontrol/v4/pkg/diff2"
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
	const providerName = "VERCEL"
	const providerMaintainer = "@SukkaW"
	fns := providers.DspFuncs{
		Initializer:   newProvider,
		RecordAuditor: AuditRecords,
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
// Though Vercel RESTful API supports getting "intendedNameServers", but it is not implemented in their Go SDK
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
	// Vercel is a "ByRecord" API.

	// Vercel enforces a minimum TTL of 60 seconds
	for _, record := range dc.Records {
		record.TTL = max(record.TTL, 60)
	}

	instructions, actualChangeCount, err := diff2.ByRecord(records, dc, nil)
	if err != nil {
		return nil, 0, err
	}

	var corrections []*models.Correction
	for _, inst := range instructions {
		switch inst.Type {
		case diff2.REPORT:
			corrections = append(corrections, &models.Correction{
				Msg: inst.MsgsJoined,
			})
		case diff2.CREATE:
			corrections = append(corrections, c.mkCreateCorrection(dc.Name, inst.New[0], inst.Msgs[0]))
		case diff2.CHANGE:
			corrections = append(corrections, c.mkChangeCorrection(dc.Name, inst.Old[0], inst.New[0], inst.Msgs[0]))
		case diff2.DELETE:
			corrections = append(corrections, c.mkDeleteCorrection(dc.Name, inst.Old[0], inst.Msgs[0]))
		default:
			panic(fmt.Sprintf("unhandled inst.Type %s", inst.Type))
		}
	}

	return corrections, actualChangeCount, nil
}

func (c *vercelProvider) createNewRecord(domain string, newRec *models.RecordConfig) error {
	ctx := context.Background()

	// Handle HTTPS records specially
	if newRec.Type == "HTTPS" {
		return c.createHTTPSRecord(ctx, domain, newRec)
	}

	// Use official SDK for other record types
	req := toVercelCreateRequest(domain, newRec)
	_, err := c.client.CreateDNSRecord(ctx, c.teamID, req)
	return err
}

func (c *vercelProvider) mkCreateCorrection(domain string, newRec *models.RecordConfig, msg string) *models.Correction {
	return &models.Correction{
		Msg: msg,
		F: func() error {
			return c.createNewRecord(domain, newRec)
		},
	}
}

func (c *vercelProvider) mkChangeCorrection(domain string, oldRec, newRec *models.RecordConfig, msg string) *models.Correction {
	return &models.Correction{
		Msg: msg,
		F: func() error {
			ctx := context.Background()
			existingID := oldRec.Original.(domainRecord).ID

			// UpdateDNSRecord doesn't support type changes
			// If record type changed, delete and re-create
			if oldRec.Type != newRec.Type {
				// Delete old record
				if err := c.client.DeleteDNSRecord(ctx, domain, existingID, c.teamID); err != nil {
					return err
				}

				return c.createNewRecord(domain, newRec)
			}

			// Handle HTTPS records specially
			if newRec.Type == "HTTPS" {
				return c.updateHTTPSRecord(ctx, existingID, newRec)
			}

			// Use official SDK for other record types
			req := toVercelUpdateRequest(newRec)
			_, err := c.client.UpdateDNSRecord(ctx, c.teamID, existingID, req)
			return err
		},
	}
}

func (c *vercelProvider) mkDeleteCorrection(domain string, oldRec *models.RecordConfig, msg string) *models.Correction {
	return &models.Correction{
		Msg: msg,
		F: func() error {
			ctx := context.Background()
			existingID := oldRec.Original.(domainRecord).ID
			return c.client.DeleteDNSRecord(ctx, domain, existingID, c.teamID)
		},
	}
}

// toVercelCreateRequest converts a RecordConfig to a Vercel CreateDNSRecordRequest.
func toVercelCreateRequest(domain string, rc *models.RecordConfig) vercelClient.CreateDNSRecordRequest {
	req := vercelClient.CreateDNSRecordRequest{
		Domain: domain,
		Name:   rc.Name,
		Type:   rc.Type,
		Value:  rc.GetTargetField(),
		TTL:    int64(rc.TTL),
	}

	switch rc.Type {
	case "MX":
		req.MXPriority = int64(rc.MxPreference)
	case "SRV":
		req.SRV = &vercelClient.SRV{
			Priority: int64(rc.SrvPriority),
			Weight:   int64(rc.SrvWeight),
			Port:     int64(rc.SrvPort),
			Target:   rc.GetTargetField(),
		}
		req.Value = "" // SRV uses the SRV struct, not Value
	case "TXT":
		req.Value = rc.GetTargetTXTJoined()
	}

	return req
}

// toVercelUpdateRequest converts a RecordConfig to a Vercel UpdateDNSRecordRequest.
func toVercelUpdateRequest(rc *models.RecordConfig) vercelClient.UpdateDNSRecordRequest {
	value := rc.GetTargetField()

	req := vercelClient.UpdateDNSRecordRequest{
		Name:    &rc.Name,
		Value:   &value,
		TTL:     ptrInt64(int64(rc.TTL)),
		Comment: "",
	}

	switch rc.Type {
	case "MX":
		req.MXPriority = ptrInt64(int64(rc.MxPreference))
	case "SRV":
		req.SRV = &vercelClient.SRVUpdate{
			Priority: ptrInt64(int64(rc.SrvPriority)),
			Weight:   ptrInt64(int64(rc.SrvWeight)),
			Port:     ptrInt64(int64(rc.SrvPort)),
			Target:   &value,
		}
		req.Value = nil // SRV uses the SRV struct, not Value
	case "TXT":
		txtValue := rc.GetTargetTXTJoined()
		req.Value = &txtValue
	}

	return req
}

// ptrInt64 returns a pointer to an int64
func ptrInt64(v int64) *int64 {
	return &v
}
