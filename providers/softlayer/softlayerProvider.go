package softlayer

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/filter"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/session"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/providers"
)

// softlayerProvider is the protocol handle for this provider.
type softlayerProvider struct {
	Session *session.Session
}

var features = providers.DocumentationNotes{
	providers.CanUseSRV:   providers.Can(),
	providers.CanGetZones: providers.Unimplemented(),
}

func init() {
	providers.RegisterDomainServiceProviderType("SOFTLAYER", newReg, features)
}

func newReg(conf map[string]string, _ json.RawMessage) (providers.DNSServiceProvider, error) {
	s := session.New(conf["username"], conf["api_key"], conf["endpoint_url"], conf["timeout"])

	if len(s.UserName) == 0 || len(s.APIKey) == 0 {
		return nil, fmt.Errorf("SoftLayer UserName and APIKey must be provided")
	}

	// s.Debug = true

	api := &softlayerProvider{
		Session: s,
	}

	return api, nil
}

// GetNameservers returns the nameservers for a domain.
func (s *softlayerProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	// Always use the same nameservers for softlayer
	return models.ToNameservers([]string{"ns1.softlayer.com", "ns2.softlayer.com"})
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (s *softlayerProvider) GetZoneRecords(domain string) (models.Records, error) {
	return nil, fmt.Errorf("not implemented")
	// This enables the get-zones subcommand.
	// Implement this by extracting the code from GetDomainCorrections into
	// a single function.  For most providers this should be relatively easy.
}

// GetDomainCorrections returns corrections to update a domain.
func (s *softlayerProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	corrections := []*models.Correction{}

	domain, err := s.getDomain(&dc.Name)

	if err != nil {
		return nil, err
	}

	actual, err := s.getExistingRecords(domain)

	if err != nil {
		return nil, err
	}

	_, create, delete, modify, err := diff.New(dc).IncrementalDiff(actual)
	if err != nil {
		return nil, err
	}

	for _, del := range delete {
		existing := del.Existing.Original.(datatypes.Dns_Domain_ResourceRecord)
		corrections = append(corrections, &models.Correction{
			Msg: del.String(),
			F:   s.deleteRecordFunc(*existing.Id),
		})
	}

	for _, cre := range create {
		corrections = append(corrections, &models.Correction{
			Msg: cre.String(),
			F:   s.createRecordFunc(cre.Desired, domain),
		})
	}

	for _, mod := range modify {
		existing := mod.Existing.Original.(datatypes.Dns_Domain_ResourceRecord)
		corrections = append(corrections, &models.Correction{
			Msg: mod.String(),
			F:   s.updateRecordFunc(&existing, mod.Desired),
		})
	}

	return corrections, nil
}

func (s *softlayerProvider) getDomain(name *string) (*datatypes.Dns_Domain, error) {
	domains, err := services.GetAccountService(s.Session).
		Filter(filter.Path("domains.name").Eq(name).Build()).
		Mask("resourceRecords").
		GetDomains()

	if err != nil {
		return nil, err
	}

	if len(domains) == 0 {
		return nil, fmt.Errorf("didn't find a domain matching %s", *name)
	} else if len(domains) > 1 {
		return nil, fmt.Errorf("found %d domains matching %s", len(domains), *name)
	}

	return &domains[0], nil
}

func (s *softlayerProvider) getExistingRecords(domain *datatypes.Dns_Domain) ([]*models.RecordConfig, error) {
	actual := []*models.RecordConfig{}

	for _, record := range domain.ResourceRecords {
		recType := strings.ToUpper(*record.Type)

		if recType == "SOA" {
			continue
		}

		recConfig := &models.RecordConfig{
			Type:     recType,
			TTL:      uint32(*record.Ttl),
			Original: record,
		}
		recConfig.SetTarget(*record.Data)

		switch recType {
		case "SRV":
			var service, protocol string = "", "_tcp"

			if record.Weight != nil {
				recConfig.SrvWeight = uint16(*record.Weight)
			}
			if record.Port != nil {
				recConfig.SrvPort = uint16(*record.Port)
			}
			if record.Priority != nil {
				recConfig.SrvPriority = uint16(*record.Priority)
			}
			if record.Protocol != nil {
				protocol = *record.Protocol
			}
			if record.Service != nil {
				service = *record.Service
			}
			recConfig.SetLabel(fmt.Sprintf("%s.%s", service, strings.ToLower(protocol)), *domain.Name)
		case "TXT":
			recConfig.TxtStrings = append(recConfig.TxtStrings, *record.Data)
			fallthrough
		case "MX":
			if record.MxPriority != nil {
				recConfig.MxPreference = uint16(*record.MxPriority)
			}
			fallthrough
		default:
			recConfig.SetLabel(*record.Host, *domain.Name)
		}

		actual = append(actual, recConfig)
	}

	// Normalize
	models.PostProcessRecords(actual)

	return actual, nil
}

func (s *softlayerProvider) createRecordFunc(desired *models.RecordConfig, domain *datatypes.Dns_Domain) func() error {
	var ttl, preference, domainID int = verifyMinTTL(int(desired.TTL)), int(desired.MxPreference), *domain.Id
	var weight, priority, port int = int(desired.SrvWeight), int(desired.SrvPriority), int(desired.SrvPort)
	var host, data, newType string = desired.GetLabel(), desired.GetTargetField(), desired.Type
	var err error

	srvRegexp := regexp.MustCompile(`^_(?P<Service>\w+)\.\_(?P<Protocol>\w+)$`)

	return func() error {
		newRecord := datatypes.Dns_Domain_ResourceRecord{
			DomainId: &domainID,
			Ttl:      &ttl,
			Type:     &newType,
			Data:     &data,
			Host:     &host,
		}

		switch newType {
		case "MX":
			service := services.GetDnsDomainResourceRecordMxTypeService(s.Session)

			newRecord.MxPriority = &preference

			newMx := datatypes.Dns_Domain_ResourceRecord_MxType{
				Dns_Domain_ResourceRecord: newRecord,
			}

			_, err = service.CreateObject(&newMx)

		case "SRV":
			service := services.GetDnsDomainResourceRecordSrvTypeService(s.Session)
			result := srvRegexp.FindStringSubmatch(host)

			if len(result) != 3 {
				return fmt.Errorf("SRV Record must match format \"_service._protocol\" not %s", host)
			}

			var serviceName, protocol string = result[1], strings.ToLower(result[2])

			newSrv := datatypes.Dns_Domain_ResourceRecord_SrvType{
				Dns_Domain_ResourceRecord: newRecord,
				Service:                   &serviceName,
				Port:                      &port,
				Priority:                  &priority,
				Protocol:                  &protocol,
				Weight:                    &weight,
			}

			_, err = service.CreateObject(&newSrv)

		default:
			service := services.GetDnsDomainResourceRecordService(s.Session)
			_, err = service.CreateObject(&newRecord)
		}

		return err
	}
}

func (s *softlayerProvider) deleteRecordFunc(resID int) func() error {
	// seems to be no problem deleting MX and SRV records via common interface
	return func() error {
		_, err := services.GetDnsDomainResourceRecordService(s.Session).
			Id(resID).
			DeleteObject()

		return err
	}
}

func (s *softlayerProvider) updateRecordFunc(existing *datatypes.Dns_Domain_ResourceRecord, desired *models.RecordConfig) func() error {
	var ttl, preference int = verifyMinTTL(int(desired.TTL)), int(desired.MxPreference)
	var priority, weight, port int = int(desired.SrvPriority), int(desired.SrvWeight), int(desired.SrvPort)

	return func() error {
		var changes = false
		var err error

		switch desired.Type {
		case "MX":
			service := services.GetDnsDomainResourceRecordMxTypeService(s.Session)
			updated := datatypes.Dns_Domain_ResourceRecord_MxType{}

			label := desired.GetLabel()
			if label != *existing.Host {
				updated.Host = &label
				changes = true
			}

			target := desired.GetTargetField()
			if target != *existing.Data {
				updated.Data = &target
				changes = true
			}

			if ttl != *existing.Ttl {
				updated.Ttl = &ttl
				changes = true
			}

			if preference != *existing.MxPriority {
				updated.MxPriority = &preference
				changes = true
			}

			if !changes {
				return fmt.Errorf("didn't find changes when I expect some")
			}

			_, err = service.Id(*existing.Id).EditObject(&updated)

		case "SRV":
			service := services.GetDnsDomainResourceRecordSrvTypeService(s.Session)
			updated := datatypes.Dns_Domain_ResourceRecord_SrvType{}

			label := desired.GetLabel()
			if label != *existing.Host {
				updated.Host = &label
				changes = true
			}

			target := desired.GetTargetField()
			if target != *existing.Data {
				updated.Data = &target
				changes = true
			}

			if ttl != *existing.Ttl {
				updated.Ttl = &ttl
				changes = true
			}

			if priority != *existing.Priority {
				updated.Priority = &priority
				changes = true
			}

			if weight != *existing.Weight {
				updated.Weight = &weight
				changes = true
			}

			if port != *existing.Port {
				updated.Port = &port
				changes = true
			}

			// TODO: handle service & protocol - or does that just result in a
			// delete and recreate?

			if !changes {
				return fmt.Errorf("didn't find changes when I expect some")
			}

			_, err = service.Id(*existing.Id).EditObject(&updated)

		default:
			service := services.GetDnsDomainResourceRecordService(s.Session)
			updated := datatypes.Dns_Domain_ResourceRecord{}

			label := desired.GetLabel()
			if label != *existing.Host {
				updated.Host = &label
				changes = true
			}

			target := desired.GetTargetField()
			if target != *existing.Data {
				updated.Data = &target
				changes = true
			}

			if ttl != *existing.Ttl {
				updated.Ttl = &ttl
				changes = true
			}

			if !changes {
				return fmt.Errorf("didn't find changes when I expect some")
			}

			_, err = service.Id(*existing.Id).EditObject(&updated)
		}

		return err
	}
}

func verifyMinTTL(ttl int) int {
	const minTTL = 60
	if ttl < minTTL {
		fmt.Printf("\nMODIFY TTL to Min supported TTL value: (ttl=%d) -> (ttl=%d)\n", ttl, minTTL)
		return minTTL
	}
	return ttl
}
