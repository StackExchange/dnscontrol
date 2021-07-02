package transip

import (
	"encoding/json"
	"fmt"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/providers"
	"github.com/transip/gotransip/v6"
	"github.com/transip/gotransip/v6/domain"
	"github.com/transip/gotransip/v6/repository"
)

type transipProvider struct {
	client  *repository.Client
	domains *domain.Repository
}

var features = providers.DocumentationNotes{
	providers.DocCreateDomains:       providers.Cannot(),
	providers.DocOfficiallySupported: providers.Cannot(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseCAA:              providers.Can("Semicolons not supported in issue/issuewild fields.", "https://www.digitalocean.com/docs/networking/dns/how-to/create-caa-records"),
	providers.CanGetZones:            providers.Can(),
}

func NewTransip(m map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	if m["AccessToken"] == "" {
		return nil, fmt.Errorf("no TransIP token provided")
	}

	client, err := gotransip.NewClient(gotransip.ClientConfiguration{
		Token: m["AccessToken"],
	})

	if err != nil {
		return nil, fmt.Errorf("TransIP client fail %s", err.Error())
	}

	api := &transipProvider{}
	api.client = &client
	api.domains = &domain.Repository{Client: client}

	return api, nil
}

func init() {
	fns := providers.DspFuncs{
		Initializer:   NewTransip,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType("TRANSIP", fns, features)
}

func (n *transipProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	return nil, fmt.Errorf("not implemented corrections")
}

func (n *transipProvider) GetZoneRecords(domainName string) (models.Records, error) {

	entries, err := n.domains.GetDNSEntries(domainName)
	if err != nil {
		return nil, err
	}

	var existingRecords = []*models.RecordConfig{}
	for _, entry := range entries {
		rts, err := nativeToRecord(entry, domainName)
		if err != nil {
			return nil, err
		}
		existingRecords = append(existingRecords, rts)
	}

	return existingRecords, nil
}

func (n *transipProvider) GetNameservers(domainName string) ([]*models.Nameserver, error) {
	return nil, fmt.Errorf("not implements ns")
}

func nativeToRecord(entry domain.DNSEntry, origin string) (*models.RecordConfig, error) {
	rc := &models.RecordConfig{TTL: uint32(*&entry.Expire)}
	rc.SetLabelFromFQDN(entry.Name, origin)
	if err := rc.PopulateFromString(entry.Type, entry.Content, origin); err != nil {
		return nil, fmt.Errorf("unparsable record received from R53: %w", err)
	}

	return rc, nil
}
