package powerdns

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/providers"
	"github.com/miekg/dns/dnsutil"
	pdns "github.com/mittwald/go-powerdns"
	"github.com/mittwald/go-powerdns/apis/zones"
	"github.com/mittwald/go-powerdns/pdnshttp"
)

var features = providers.DocumentationNotes{
	providers.CanUseAlias:            providers.Can(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseTLSA:             providers.Can(),
	providers.CanUseSSHFP:            providers.Can(),
	providers.CanAutoDNSSEC:          providers.Can(),
	providers.DocCreateDomains:       providers.Can(),
	providers.DocOfficiallySupported: providers.Cannot(),
	providers.CanGetZones:            providers.Can(),
}

func init() {
	providers.RegisterDomainServiceProviderType("POWERDNS", NewProvider, features)
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

// GetNameservers returns the nameservers for a domain.
func (api *powerdnsProvider) GetNameservers(string) ([]*models.Nameserver, error) {
	var r []string
	for _, j := range api.nameservers {
		r = append(r, j.Name)
	}
	return models.ToNameservers(r)
}

// ListZones returns all the zones in an account
func (api *powerdnsProvider) ListZones() ([]string, error) {
	var result []string
	zones, err := api.client.Zones().ListZones(context.Background(), api.ServerName)
	if err != nil {
		return result, err
	}
	for _, zone := range zones {
		result = append(result, zone.Name)
	}
	return result, nil
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (api *powerdnsProvider) GetZoneRecords(domain string) (models.Records, error) {
	zone, err := api.client.Zones().GetZone(context.Background(), api.ServerName, domain)
	if err != nil {
		return nil, err
	}

	curRecords := models.Records{}
	// loop over grouped records by type, called RRSet
	for _, rrset := range zone.ResourceRecordSets {
		if rrset.Type == "SOA" {
			continue
		}
		// loop over single records of this group and create records
		for _, pdnsRecord := range rrset.Records {
			r, err := toRecordConfig(domain, pdnsRecord, rrset.TTL, rrset.Name, rrset.Type)
			if err != nil {
				return nil, err
			}
			curRecords = append(curRecords, r)
		}
	}

	return curRecords, nil
}

// GetDomainCorrections returns a list of corrections to update a domain.
func (api *powerdnsProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	var corrections []*models.Correction

	// record corrections
	curRecords, err := api.GetZoneRecords(dc.Name)
	if err != nil {
		return nil, err
	}

	// post-process records
	dc.Punycode()
	models.PostProcessRecords(curRecords)

	// create record diff by group
	keysToUpdate, err := (diff.New(dc)).ChangedGroups(curRecords)
	if err != nil {
		return nil, err
	}
	desiredRecords := dc.Records.GroupedByKey()

	// create corrections by group
	for label, msgs := range keysToUpdate {
		labelName := label.NameFQDN + "."
		labelType := label.Type

		if _, ok := desiredRecords[label]; !ok {
			// nothing found, must be a delete
			corrections = append(corrections, &models.Correction{
				Msg: strings.Join(msgs, "\n   "),
				F: func() error {
					return api.client.Zones().RemoveRecordSetFromZone(context.Background(), api.ServerName, dc.Name, labelName, labelType)
				},
			})
		} else {
			ttl := desiredRecords[label][0].TTL
			records := []zones.Record{}
			for _, recordContent := range desiredRecords[label] {
				records = append(records, zones.Record{
					Content: recordContent.GetTargetCombined(),
				})
			}
			corrections = append(corrections, &models.Correction{
				Msg: strings.Join(msgs, "\n   "),
				F: func() error {
					return api.client.Zones().AddRecordSetToZone(context.Background(), api.ServerName, dc.Name, zones.ResourceRecordSet{
						Name:    labelName,
						Type:    labelType,
						TTL:     int(ttl),
						Records: records,
					})
				},
			})
		}
	}

	// DNSSec corrections
	dnssecCorrections, err := api.getDNSSECCorrections(dc)
	if err != nil {
		return nil, err
	}
	corrections = append(corrections, dnssecCorrections...)

	return corrections, nil
}

// EnsureDomainExists adds a domain to the DNS service if it does not exist
func (api *powerdnsProvider) EnsureDomainExists(domain string) error {
	if _, err := api.client.Zones().GetZone(context.Background(), api.ServerName, domain+"."); err != nil {
		if e, ok := err.(pdnshttp.ErrUnexpectedStatus); ok {
			if e.StatusCode != http.StatusNotFound {
				return err
			}
		}
	} else { // domain seems to be there
		return nil
	}

	_, err := api.client.Zones().CreateZone(context.Background(), api.ServerName, zones.Zone{
		Name:        domain + ".",
		Type:        zones.ZoneTypeZone,
		DNSSec:      api.DNSSecOnCreate,
		Nameservers: api.DefaultNS,
	})
	return err
}

// toRecordConfig converts a PowerDNS DNSRecord to a RecordConfig. #rtype_variations
func toRecordConfig(domain string, r zones.Record, ttl int, name string, rtype string) (*models.RecordConfig, error) {
	// trimming trailing dot and domain from name
	name = strings.TrimSuffix(name, domain+".")
	name = strings.TrimSuffix(name, ".")

	rc := &models.RecordConfig{
		TTL:      uint32(ttl),
		Original: r,
		Type:     rtype,
	}
	rc.SetLabel(name, domain)

	content := r.Content
	switch rtype {
	case "CNAME", "NS":
		return rc, rc.SetTarget(dnsutil.AddOrigin(content, domain))
	case "CAA":
		return rc, rc.SetTargetCAAString(content)
	case "MX":
		return rc, rc.SetTargetMXString(content)
	case "SRV":
		return rc, rc.SetTargetSRVString(content)
	case "TXT":
		// Remove quotes if it is a TXT record.
		if !strings.HasPrefix(content, `"`) || !strings.HasSuffix(content, `"`) {
			return nil, errors.New("unexpected lack of quotes in TXT record from PowerDNS")
		}
		return rc, rc.SetTargetTXT(content[1 : len(content)-1])
	default:
		return rc, rc.PopulateFromString(rtype, content, domain)
	}
}
