package packetframe

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/providers"
)

// packetframeProvider is the handle for this provider.
type packetframeProvider struct {
	client      *http.Client
	baseURL     *url.URL
	token       string
	domainIndex map[string]zone
}

var defaultNameServerNames = []string{
	"ns1v4.packetframe.com",
	"ns2v4.packetframe.com",
}

// newPacketframe creates the provider.
func newPacketframe(m map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	if m["apikey"] == "" {
		return nil, fmt.Errorf("missing Packetframe token")
	}

	baseURL, err := url.Parse(defaultBaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL for Packetframe")
	}
	client := http.Client{}

	api := &packetframeProvider{client: &client, baseURL: baseURL, token: m["apikey"]}

	return api, nil
}

var features = providers.DocumentationNotes{
	providers.DocDualHost:            providers.Cannot(),
	providers.DocOfficiallySupported: providers.Cannot(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanGetZones:            providers.Unimplemented(),
}

func init() {
	fns := providers.DspFuncs{
		Initializer:   newPacketframe,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType("PACKETFRAME", fns, features)
}

// GetNameservers returns the nameservers for a domain.
func (api *packetframeProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	return models.ToNameservers(defaultNameServerNames)
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (api *packetframeProvider) GetZoneRecords(domain string) (models.Records, error) {

	if api.domainIndex == nil {
		if err := api.fetchDomainList(); err != nil {
			return nil, err
		}
	}
	zone, ok := api.domainIndex[domain+"."]
	if !ok {
		return nil, fmt.Errorf("'%s' not a zone in Packetframe account", domain)
	}

	records, err := api.getRecords(zone.ID)
	if err != nil {
		return nil, fmt.Errorf("could not load records for '%s'", domain)
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

// GetDomainCorrections returns the corrections for a domain.
func (api *packetframeProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	dc, err := dc.Copy()
	if err != nil {
		return nil, err
	}

	dc.Punycode()

	if api.domainIndex == nil {
		if err := api.fetchDomainList(); err != nil {
			return nil, err
		}
	}
	zone, ok := api.domainIndex[dc.Name+"."]
	if !ok {
		return nil, fmt.Errorf("'%s' not a zone in Packetframe account", dc.Name)
	}

	records, err := api.getRecords(zone.ID)
	if err != nil {
		return nil, fmt.Errorf("could not load records for '%s'", dc.Name)
	}

	existingRecords := make([]*models.RecordConfig, len(records))

	for i := range records {
		existingRecords[i] = toRc(dc, &records[i])
	}

	// Normalize
	models.PostProcessRecords(existingRecords)

	differ := diff.New(dc)
	_, create, delete, modify, err := differ.IncrementalDiff(existingRecords)
	if err != nil {
		return nil, err
	}

	var corrections []*models.Correction

	for _, m := range create {
		req, err := toReq(zone.ID, dc, m.Desired)
		if err != nil {
			return nil, err
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

	for _, m := range delete {
		original := m.Existing.Original.(*domainRecord)
		corr := &models.Correction{
			Msg: fmt.Sprintf("Deleting record %s from %s", original.ID, zone.Zone),
			F: func() error {
				err := api.deleteRecord(zone.ID, original.ID)
				return err
			},
		}
		corrections = append(corrections, corr)
	}

	for _, m := range modify {
		original := m.Existing.Original.(*domainRecord)
		req, _ := toReq(zone.ID, dc, m.Desired)
		req.ID = original.ID
		corr := &models.Correction{
			Msg: fmt.Sprintf("Modifying record %s from %s", original.ID, zone.Zone),
			F: func() error {
				err := api.modifyRecord(req)
				return err
			},
		}
		corrections = append(corrections, corr)
	}

	return corrections, nil
}

func toReq(zoneID string, dc *models.DomainConfig, rc *models.RecordConfig) (*domainRecord, error) {
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
		rc.SetTargetTXTString(r.Value)
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
