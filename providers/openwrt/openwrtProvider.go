package openwrt

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"strconv"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/diff2"
	"github.com/StackExchange/dnscontrol/v4/pkg/printer"
	"github.com/StackExchange/dnscontrol/v4/pkg/providers"
)

type openwrtProvider struct {
	auth string
	host string
}

type rewriteEntity struct {
	Section string `json:".name,omitempty"`
	Type    string `json:".type,omitempty"`

	// A
	Name string `json:"name,omitempty"`
	IP   string `json:"ip,omitempty"`

	// CNAME
	Cname  string `json:"cname,omitempty"`
	Target string `json:"target,omitempty"`

	// MX
	Domain string `json:"domain,omitempty"`
	Relay  string `json:"relay,omitempty"`
	Pref   string `json:"pref,omitempty"`

	// SRV
	Srv      string `json:"srv,omitempty"`
	Priority string `json:"class,omitempty"`
	Weight   string `json:"weight,omitempty"`
	Port     string `json:"port,omitempty"`
	// Target string `json:"target,omitempty"`
}

func newDsp(conf map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	return newOpenwrt(conf, metadata)
}

// newOpenwrt creates the provider.
func newOpenwrt(conf map[string]string, _ json.RawMessage) (*openwrtProvider, error) {
	if conf["username"] == "" {
		return nil, errors.New("missing openwrt username")
	}
	if conf["password"] == "" {
		return nil, errors.New("missing openwrt password")
	}
	if conf["host"] == "" {
		return nil, errors.New("missing openwrt host")
	}

	host := conf["host"]
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "http://" + host
	}

	auth, err := getAuthorization(conf["username"], conf["password"], host)
	if err != nil {
		return nil, fmt.Errorf("could not login: %w", err)
	}

	return &openwrtProvider{auth: auth, host: host}, nil
}

var features = providers.DocumentationNotes{
	providers.CanGetZones:            providers.Can(),
	providers.CanUseAlias:            providers.Cannot(),
	providers.CanUseSRV:              providers.Can(),
	providers.DocOfficiallySupported: providers.Cannot(),
}

func init() {
	const providerName = "OPENWRT"
	const providerMaintainer = "@huskyistaken"
	fns := providers.DspFuncs{
		Initializer:   newDsp,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType(providerName, fns, features)
	providers.RegisterMaintainer(providerName, providerMaintainer)
}

// GetNameservers returns the nameservers for a domain.
func (c *openwrtProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	return []*models.Nameserver{}, nil
}

// GetZoneRecordsCorrections returns a list of corrections that will turn existing records into dc.Records.
func (c *openwrtProvider) GetZoneRecordsCorrections(dc *models.DomainConfig, existingRecords models.Records) ([]*models.Correction, int, error) {
	// TTLs don't matter in OPENWRT and
	// we use the default value of 300
	for _, record := range dc.Records {
		record.TTL = 300
	}

	var corrections []*models.Correction

	changes, actualChangeCount, err := diff2.ByRecord(existingRecords, dc,
		func(rec *models.RecordConfig) string { return "" },
	)
	if err != nil {
		return nil, 0, err
	}
	for _, change := range changes {
		var corr *models.Correction
		switch change.Type {
		case diff2.REPORT:
			printer.Warnf("diff2 report message\n")
			corr = &models.Correction{Msg: change.MsgsJoined}

		case diff2.CREATE:
			var recordType string
			switch change.New[0].Type {
			case "A", "AAAA":
				recordType = "domain"
			case "CNAME":
				recordType = "cname"
			case "SRV":
				recordType = "srvhost"
			case "MX":
				recordType = "mxhost"
			}
			re, err := toRewriteEntry(change.New[0])
			if err != nil {
				return nil, 0, err
			}

			corr = &models.Correction{
				Msg: change.Msgs[0],
				F: func() error {
					_, err := c.uciSection(recordType, re)
					return err
				},
			}

		case diff2.DELETE:
			section := change.Old[0].Original.(rewriteEntity).Section
			corr = &models.Correction{
				Msg: change.Msgs[0],
				F: func() error {
					fmt.Println(section)
					_, err := c.uciDelete(section)
					return err
				},
			}

		case diff2.CHANGE:
			section := change.Old[0].Original.(rewriteEntity).Section
			re, err := toRewriteEntry(change.New[0])
			if err != nil {
				return nil, 0, err
			}
			corr = &models.Correction{
				Msg: change.Msgs[0],
				F: func() error {
					_, err := c.uciTset(section, re)
					return err
				},
			}

		default:
			panic(fmt.Sprintf("unhandled change.Type %s", change.Type))
		}

		corrections = append(corrections, corr)
	}

	// Apply changes last, changes cannot be applied incrementally
	// because doing so shifts the section names, making deleting
	// records unreliable
	if actualChangeCount > 0 {
		corrections = append(corrections, &models.Correction{
			Msg: "Applying changes",
			F: func() error {
				_, err := c.uciApply()
				return err
			},
		})
	}

	return corrections, actualChangeCount, nil
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (c *openwrtProvider) GetZoneRecords(dc *models.DomainConfig) (models.Records, error) {
	domain := dc.Name

	records, err := c.getRecords(domain)
	if err != nil {
		return nil, err
	}

	existingRecords := make([]*models.RecordConfig, 0)
	for _, r := range records {
		rc, err := toRc(domain, r)
		if err != nil {
			return nil, err
		}
		existingRecords = append(existingRecords, rc)
	}

	return existingRecords, nil
}

func toRc(domain string, r rewriteEntity) (*models.RecordConfig, error) {
	rc := &models.RecordConfig{
		TTL:      300,
		Original: r,
	}
	var recDomain string

	switch r.Type {
	case "domain":
		recDomain = r.Name
		addr, err := netip.ParseAddr(r.IP)
		if err != nil {
			return nil, err
		}

		rc.SetTargetIP(addr)
		switch {
		case addr.Is4():
			rc.Type = "A"
		case addr.Is6():
			rc.Type = "AAAA"
		}

	case "cname":
		recDomain = r.Cname
		rc.Type = "CNAME"
		rc.SetTarget(r.Target)

	case "mxhost":
		recDomain = r.Domain
		rc.Type = "MX"
		pref, err := strconv.ParseUint(r.Pref, 10, 16)
		if err != nil {
			return nil, err
		}
		rc.SetTargetMX(uint16(pref), r.Relay)

	case "srvhost":
		recDomain = r.Srv
		rc.Type = "SRV"
		priority, err := strconv.ParseUint(r.Priority, 10, 16)
		if err != nil {
			return nil, err
		}
		weight, err := strconv.ParseUint(r.Weight, 10, 16)
		if err != nil {
			return nil, err
		}
		port, err := strconv.ParseUint(r.Port, 10, 16)
		if err != nil {
			return nil, err
		}
		rc.SetTargetSRV(uint16(priority), uint16(weight), uint16(port), r.Target)

	default:
		return nil, fmt.Errorf("unhandled record type: %s", r.Type)
	}

	rc.SetLabelFromFQDN(recDomain, domain)

	return rc, nil
}

func toRewriteEntry(rc *models.RecordConfig) (rewriteEntity, error) {
	var newRecordEntry rewriteEntity

	// omits .type and .name
	switch rc.Type {
	case "A", "AAAA":
		newRecordEntry.Name = rc.NameFQDN
		newRecordEntry.IP = rc.GetTargetIP().String()

	case "CNAME":
		newRecordEntry.Cname = rc.NameFQDN
		newRecordEntry.Target = rc.GetTargetField()

	case "SRV":
		newRecordEntry.Srv = rc.NameFQDN
		newRecordEntry.Priority = string(strconv.Itoa(int(rc.SrvPriority)))
		newRecordEntry.Weight = strconv.Itoa(int(rc.SrvWeight))
		newRecordEntry.Port = strconv.Itoa(int(rc.SrvPort))
		newRecordEntry.Target = rc.GetTargetField()

	case "MX":
		newRecordEntry.Domain = rc.NameFQDN
		newRecordEntry.Pref = strconv.Itoa(int(rc.MxPreference))
		newRecordEntry.Relay = rc.GetTargetField()

	default:
		return rewriteEntity{}, fmt.Errorf("unhandled record type: %s", rc.Type)
	}

	return newRecordEntry, nil
}
