package infomaniak

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/diff2"
	"github.com/StackExchange/dnscontrol/v4/providers"
)

// infomaniakProvider is the handle for operations.
type infomaniakProvider struct {
	apiToken string // the account access token
}

var features = providers.DocumentationNotes{
	// The default for unlisted capabilities is 'Cannot'.
	// See providers/capabilities.go for the entire list of capabilities.
	providers.CanGetZones: providers.Can(),
	providers.CanUseCAA:   providers.Can(),
	providers.CanUseDNAME: providers.Can(),
	providers.CanUseDS:    providers.Can(),
	providers.CanUseSSHFP: providers.Can(),
	providers.CanUseTLSA:  providers.Can(),
	providers.CanUseSRV:   providers.Can(),
	// providers.DocCreateDomains: providers.Can(),
}

func newInfomaniak(m map[string]string, message json.RawMessage) (providers.DNSServiceProvider, error) {
	api := &infomaniakProvider{}
	api.apiToken = m["token"]
	if api.apiToken == "" {
		return nil, errors.New("missing Infomaniak personal access token")
	}

	return api, nil
}

func init() {
	const providerName = "INFOMANIAK"
	const providerMaintainer = "@jbelien"
	fns := providers.DspFuncs{
		Initializer:   newInfomaniak,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType(providerName, fns, features)
	providers.RegisterMaintainer(providerName, providerMaintainer)
}

func (p *infomaniakProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	zone, err := p.getDNSZone(domain)
	if err != nil {
		return nil, err
	}

	return models.ToNameservers(zone.Nameservers)
}

func (p *infomaniakProvider) GetZoneRecords(domain string, meta map[string]string) (models.Records, error) {
	records, err := p.getDNSRecords(domain)
	if err != nil {
		return nil, err
	}

	cleanRecords := make(models.Records, 0)

	for _, r := range records {
		recConfig := &models.RecordConfig{
			Original: r,
			TTL:      uint32(r.TTL),
			Type:     r.Type,
		}
		recConfig.SetLabelFromFQDN(r.Source, domain)
		recConfig.SetTarget(r.Target)

		cleanRecords = append(cleanRecords, recConfig)
	}

	return cleanRecords, nil
}

func (p *infomaniakProvider) GetZoneRecordsCorrections(dc *models.DomainConfig, existingRecords models.Records) ([]*models.Correction, int, error) {
	var corrections []*models.Correction

	changes, actualChangeCount, err := diff2.ByRecord(existingRecords, dc, nil)
	if err != nil {
		return nil, 0, err
	}

	for _, change := range changes {
		switch change.Type {
		case diff2.REPORT:
			corrections = append(corrections, &models.Correction{Msg: change.MsgsJoined})
		case diff2.CHANGE:
			fmt.Printf("CHANGE: %+v\n", change.New)
			// corrections = append(corrections, &models.Correction{
			// 	Msg: change.Msgs[0],
			// 	F: func() error {
			// 		return p.updateRecord(change.Old[0].Original.(dnsRecord), change.New[0], dc.Name)
			// 	},
			// })
		case diff2.CREATE:
			fmt.Printf("CREATE: %+v\n", change.New)
			// corrections = append(corrections, &models.Correction{
			// 	Msg: change.Msgs[0],
			// 	F: func() error {
			// 		_, err := p.createDNSRecord(dc.Name, change.New[0])
			// 		return err
			// 	},
			// })
		case diff2.DELETE:
			rec := change.Old[0].Original.(dnsRecord)
			corrections = append(corrections, &models.Correction{
				Msg: change.Msgs[0],
				F: func() error {
					return p.deleteDNSRecord(dc.Name, fmt.Sprintf("%v", rec.ID))
				},
			})
		}
	}

	return corrections, actualChangeCount, nil
}
