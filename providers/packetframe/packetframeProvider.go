package packetframe

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/diff"
	"github.com/StackExchange/dnscontrol/v4/providers"
)

// packetframeProvider is the handle for this provider.
type packetframeProvider struct {
	client      *http.Client
	baseURL     *url.URL
	token       string
	domainIndex map[string]zoneInfo
}

// newPacketframe creates the provider.
func newPacketframe(m map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	if m["token"] == "" {
		return nil, fmt.Errorf("missing Packetframe token")
	}

	baseURL, err := url.Parse(defaultBaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL for Packetframe")
	}
	client := http.Client{}

	api := &packetframeProvider{client: &client, baseURL: baseURL, token: m["token"]}

	return api, nil
}

var features = providers.DocumentationNotes{
	// The default for unlisted capabilities is 'Cannot'.
	// See providers/capabilities.go for the entire list of capabilities.
	providers.CanGetZones:            providers.Unimplemented(),
	providers.CanConcur:              providers.Cannot(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseSRV:              providers.Can(),
	providers.DocDualHost:            providers.Cannot(),
	providers.DocOfficiallySupported: providers.Cannot(),
}

func init() {
	const providerName = "PACKETFRAME"
	const providerMaintainer = "@hamptonmoore"
	fns := providers.DspFuncs{
		Initializer:   newPacketframe,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType(providerName, fns, features)
	providers.RegisterMaintainer(providerName, providerMaintainer)
}

// GetNameservers returns the nameservers for a domain.
func (api *packetframeProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	return models.ToNameservers(defaultNameServerNames)
}

func (api *packetframeProvider) getZone(domain string) (*zoneInfo, error) {
	if api.domainIndex == nil {
		if err := api.fetchDomainList(); err != nil {
			return nil, err
		}
	}
	z, ok := api.domainIndex[domain+"."]
	if !ok {
		return nil, fmt.Errorf("%q not a zone in Packetframe account", domain)
	}

	return &z, nil
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (api *packetframeProvider) GetZoneRecords(domain string, meta map[string]string) (models.Records, error) {

	zone, err := api.getZone(domain)
	if err != nil {
		return nil, fmt.Errorf("no such zone %q in Packetframe account", domain)
	}

	records, err := api.getRecords(zone.ID)
	if err != nil {
		return nil, fmt.Errorf("could not load records for domain %q", domain)
	}

	existingRecords := make([]*models.RecordConfig, len(records))

	dc := models.DomainConfig{
		Name: domain,
	}

	for i := range records {
		existingRecords[i] = toRc(&dc, &records[i])
	}

	return existingRecords, nil
}

// GetZoneRecordsCorrections returns a list of corrections that will turn existing records into dc.Records.
func (api *packetframeProvider) GetZoneRecordsCorrections(dc *models.DomainConfig, existingRecords models.Records) ([]*models.Correction, int, error) {
	zone, err := api.getZone(dc.Name)
	if err != nil {
		return nil, 0, fmt.Errorf("no such zone %q in Packetframe account", dc.Name)
	}

	toReport, create, dels, modify, actualChangeCount, err := diff.NewCompat(dc).IncrementalDiff(existingRecords)
	if err != nil {
		return nil, 0, err
	}
	// Start corrections with the reports
	corrections := diff.GenerateMessageCorrections(toReport)

	for _, m := range create {
		req, err := toReq(zone.ID, m.Desired)
		if err != nil {
			return nil, 0, err
		}
		corr := &models.Correction{
			Msg: m.String(),
			F: func() error {
				_, err := api.createRecord(req)
				return err
			},
		}
		corrections = append(corrections, corr)
	}

	for _, m := range dels {
		original := m.Existing.Original.(*domainRecord)
		if original.ID == "0" { // Skip the default nameservers
			continue
		}

		corr := &models.Correction{
			Msg: m.String(),
			F: func() error {
				err := api.deleteRecord(zone.ID, original.ID)
				return err
			},
		}
		corrections = append(corrections, corr)
	}

	for _, m := range modify {
		original := m.Existing.Original.(*domainRecord)
		if original.ID == "0" { // Skip the default nameservers
			continue
		}

		req, _ := toReq(zone.ID, m.Desired)
		req.ID = original.ID
		corr := &models.Correction{
			Msg: m.String(),
			F: func() error {
				err := api.modifyRecord(req)
				return err
			},
		}
		corrections = append(corrections, corr)
	}

	return corrections, actualChangeCount, nil
}

func toReq(zoneID string, rc *models.RecordConfig) (*domainRecord, error) {
	req := &domainRecord{
		Type:  rc.Type,
		TTL:   int(rc.TTL),
		Label: rc.GetLabel(),
		Zone:  zoneID,
	}

	switch rc.Type { // #rtype_variations
	case "A", "AAAA", "PTR", "TXT", "CNAME", "NS":
		req.Value = rc.GetTargetField()
	case "MX":
		req.Value = fmt.Sprintf("%d %s", rc.MxPreference, rc.GetTargetField())
	case "SRV":
		req.Value = fmt.Sprintf("%d %d %d %s", rc.SrvPriority, rc.SrvWeight, rc.SrvPort, rc.GetTargetField())
	default:
		return nil, fmt.Errorf("packetframe.toReq rtype %q unimplemented", rc.Type)
	}

	return req, nil
}

func toRc(dc *models.DomainConfig, r *domainRecord) *models.RecordConfig {
	rc := &models.RecordConfig{
		Type:     r.Type,
		TTL:      uint32(r.TTL),
		Original: r,
	}

	label := strings.TrimSuffix(r.Label, dc.Name+".")
	label = strings.TrimSuffix(label, ".")
	if label == "" {
		label = "@"
	}
	rc.SetLabel(label, dc.Name)

	switch rtype := r.Type; rtype { // #rtype_variations
	case "TXT":
		rc.SetTargetTXT(r.Value)
	case "SRV":
		spl := strings.Split(r.Value, " ")
		prio, _ := strconv.ParseUint(spl[0], 10, 16)
		weight, _ := strconv.ParseUint(spl[1], 10, 16)
		port, _ := strconv.ParseUint(spl[2], 10, 16)
		rc.SetTargetSRV(uint16(prio), uint16(weight), uint16(port), spl[3])
	case "MX":
		spl := strings.Split(r.Value, " ")
		prio, _ := strconv.ParseUint(spl[0], 10, 16)
		rc.SetTargetMX(uint16(prio), spl[1])
	default:
		rc.SetTarget(r.Value)
	}

	return rc
}
