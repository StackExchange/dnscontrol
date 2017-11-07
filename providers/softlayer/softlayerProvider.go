package softlayer

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers"
	"github.com/StackExchange/dnscontrol/providers/diff"
	"github.com/miekg/dns/dnsutil"

	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/filter"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/session"
)

type SoftLayer struct {
	Session *session.Session
}

func init() {
	providers.RegisterDomainServiceProviderType("SOFTLAYER", newReg, providers.CanUseSRV)
}

func newReg(conf map[string]string, _ json.RawMessage) (providers.DNSServiceProvider, error) {
	s := session.New(conf["username"], conf["api_key"], conf["endpoint_url"], conf["timeout"])

	if len(s.UserName) == 0 || len(s.APIKey) == 0 {
		return nil, fmt.Errorf("SoftLayer UserName and APIKey must be provided")
	}

	//s.Debug = true

	api := &SoftLayer{
		Session: s,
	}

	return api, nil
}

func (s *SoftLayer) GetNameservers(domain string) ([]*models.Nameserver, error) {
	// Always use the same nameservers for softlayer
	nservers := []string{"ns1.softlayer.com", "ns2.softlayer.com"}
	return models.StringsToNameservers(nservers), nil
}

func (s *SoftLayer) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	corrections := []*models.Correction{}

	domain, err := s.getDomain(&dc.Name)

	if err != nil {
		return nil, err
	}

	actual, err := s.getExistingRecords(domain)

	if err != nil {
		return nil, err
	}

	_, create, delete, modify := diff.New(dc).IncrementalDiff(actual)

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

func (s *SoftLayer) getDomain(name *string) (*datatypes.Dns_Domain, error) {
	domains, err := services.GetAccountService(s.Session).
		Filter(filter.Path("domains.name").Eq(name).Build()).
		Mask("resourceRecords").
		GetDomains()

	if err != nil {
		return nil, err
	}

	if len(domains) == 0 {
		return nil, fmt.Errorf("Didn't find a domain matching %s", *name)
	} else if len(domains) > 1 {
		return nil, fmt.Errorf("Found %d domains matching %s", len(domains), *name)
	}

	return &domains[0], nil
}

func (s *SoftLayer) getExistingRecords(domain *datatypes.Dns_Domain) ([]*models.RecordConfig, error) {
	actual := []*models.RecordConfig{}

	for _, record := range domain.ResourceRecords {
		recType := strings.ToUpper(*record.Type)

		if recType == "SOA" {
			continue
		}

		recConfig := &models.RecordConfig{
			Type:     recType,
			Target:   *record.Data,
			TTL:      uint32(*record.Ttl),
			Original: record,
		}

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

			recConfig.Name = fmt.Sprintf("%s.%s", service, strings.ToLower(protocol))

		case "MX":
			if record.MxPriority != nil {
				recConfig.MxPreference = uint16(*record.MxPriority)
			}

			fallthrough

		default:
			recConfig.Name = *record.Host
		}

		recConfig.NameFQDN = dnsutil.AddOrigin(recConfig.Name, *domain.Name)
		actual = append(actual, recConfig)
	}

	return actual, nil
}

func (s *SoftLayer) createRecordFunc(desired *models.RecordConfig, domain *datatypes.Dns_Domain) func() error {
	var ttl, preference, domainId int = int(desired.TTL), int(desired.MxPreference), *domain.Id
	var weight, priority, port int = int(desired.SrvWeight), int(desired.SrvPriority), int(desired.SrvPort)
	var host, data, newType string = desired.Name, desired.Target, desired.Type
	var err error = nil

	srvRegexp := regexp.MustCompile(`^_(?P<Service>\w+)\.\_(?P<Protocol>\w+)$`)

	return func() error {
		newRecord := datatypes.Dns_Domain_ResourceRecord{
			DomainId: &domainId,
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

func (s *SoftLayer) deleteRecordFunc(resId int) func() error {
	// seems to be no problem deleting MX and SRV records via common interface
	return func() error {
		_, err := services.GetDnsDomainResourceRecordService(s.Session).
			Id(resId).
			DeleteObject()

		return err
	}
}

func (s *SoftLayer) updateRecordFunc(existing *datatypes.Dns_Domain_ResourceRecord, desired *models.RecordConfig) func() error {
	var ttl, preference int = int(desired.TTL), int(desired.MxPreference)
	var priority, weight, port int = int(desired.SrvPriority), int(desired.SrvWeight), int(desired.SrvPort)

	return func() error {
		var changes bool = false
		var err error = nil

		switch desired.Type {
		case "MX":
			service := services.GetDnsDomainResourceRecordMxTypeService(s.Session)
			updated := datatypes.Dns_Domain_ResourceRecord_MxType{}

			if desired.Name != *existing.Host {
				updated.Host = &desired.Name
				changes = true
			}

			if desired.Target != *existing.Data {
				updated.Data = &desired.Target
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
				return fmt.Errorf("Error: Didn't find changes when I expect some.")
			}

			_, err = service.Id(*existing.Id).EditObject(&updated)

		case "SRV":
			service := services.GetDnsDomainResourceRecordSrvTypeService(s.Session)
			updated := datatypes.Dns_Domain_ResourceRecord_SrvType{}

			if desired.Name != *existing.Host {
				updated.Host = &desired.Name
				changes = true
			}

			if desired.Target != *existing.Data {
				updated.Data = &desired.Target
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
				return fmt.Errorf("Error: Didn't find changes when I expect some.")
			}

			_, err = service.Id(*existing.Id).EditObject(&updated)

		default:
			service := services.GetDnsDomainResourceRecordService(s.Session)
			updated := datatypes.Dns_Domain_ResourceRecord{}

			if desired.Name != *existing.Host {
				updated.Host = &desired.Name
				changes = true
			}

			if desired.Target != *existing.Data {
				updated.Data = &desired.Target
				changes = true
			}

			if ttl != *existing.Ttl {
				updated.Ttl = &ttl
				changes = true
			}

			if !changes {
				return fmt.Errorf("Error: Didn't find changes when I expect some.")
			}

			_, err = service.Id(*existing.Id).EditObject(&updated)
		}

		return err
	}
}
