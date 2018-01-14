package gandi

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers"
	"github.com/StackExchange/dnscontrol/providers/diff"
	"github.com/google/uuid"
	"github.com/miekg/dns/dnsutil"
	"github.com/pkg/errors"
	gandiclient "github.com/prasmussen/gandi-api/client"
	gandilivedomain "github.com/prasmussen/gandi-api/live_dns/domain"
	gandiliverecord "github.com/prasmussen/gandi-api/live_dns/record"
	gandilivezone "github.com/prasmussen/gandi-api/live_dns/zone"
)

var liveFeatures = providers.DocumentationNotes{
	providers.CanUseCAA:              providers.Can(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseSRV:              providers.Can(),
	providers.CantUseNOPURGE:         providers.Cannot(),
	providers.DocCreateDomains:       providers.Cannot("Can only manage domains registered through their service"),
	providers.DocOfficiallySupported: providers.Cannot(),
}

func init() {
	providers.RegisterDomainServiceProviderType("GANDI-LIVEDNS", newLiveDsp, liveFeatures)
}

func newLiveDsp(m map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	APIKey := m["apikey"]
	if APIKey == "" {
		return nil, fmt.Errorf("missing Gandi apikey")
	}

	return newLiveClient(APIKey), nil
}

type domainManager interface {
	Info(string) (*gandilivedomain.Info, error)
	Records(string) gandiliverecord.Manager
}

type zoneManager interface {
	InfoByUUID(uuid.UUID) (*gandilivezone.Info, error)
	Create(gandilivezone.Info) (*gandilivezone.CreateStatus, error)
	Set(string, gandilivezone.Info) (*gandilivezone.Status, error)
	Records(gandilivezone.Info) gandiliverecord.Manager
}

type liveClient struct {
	client        *gandiclient.Client
	zoneManager   zoneManager
	domainManager domainManager
}

func newLiveClient(APIKey string) *liveClient {
	cl := gandiclient.New(APIKey, gandiclient.LiveDNS)
	return &liveClient{
		client:        cl,
		zoneManager:   gandilivezone.New(cl),
		domainManager: gandilivedomain.New(cl),
	}
}

// GetNameservers returns the list of gandi name servers for a given domain
func (c *liveClient) GetNameservers(domain string) ([]*models.Nameserver, error) {
	domains := []string{}
	response, err := c.client.Get("/nameservers/"+domain, &domains)
	if err != nil {
		return nil, fmt.Errorf("failed to get nameservers for domain %s", domain)
	}
	defer response.Body.Close()

	ns := []*models.Nameserver{}
	for _, domain := range domains {
		ns = append(ns, &models.Nameserver{Name: domain})
	}
	return ns, nil
}

// GetDomainCorrections returns a list of corrections recommended for this domain.
func (c *liveClient) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	dc.Punycode()
	dc.CombineSRVs()
	dc.CombineCAAs()
	dc.CombineMXs()
	records, err := c.domainManager.Records(dc.Name).List()
	if err != nil {
		return nil, err
	}
	foundRecords := c.recordConfigFromInfo(records, dc.Name)
	models.PostProcessRecords(foundRecords)

	differ := diff.New(dc)
	_, create, del, mod := differ.IncrementalDiff(foundRecords)
	if len(create)+len(del)+len(mod) > 0 {
		message := fmt.Sprintf("Setting dns records for %s:", dc.Name)
		for _, record := range dc.Records {
			message += "\n" + record.String()
		}
		return []*models.Correction{
			{
				Msg: message,
				F: func() error {
					records, err := c.recordsToInfo(dc.Records)
					if err != nil {
						return err
					}
					c.createZone(dc.Name, records)
					return nil
				},
			},
		}, nil
	}
	return []*models.Correction{}, nil
}

// createZone creates a new empty zone for the domain, populates it with the record infos and associates the domain to it
func (c *liveClient) createZone(domainname string, records []*gandiliverecord.Info) error {
	domainInfo, err := c.domainManager.Info(domainname)
	infos, err := c.zoneManager.InfoByUUID(*domainInfo.ZoneUUID)
	if err != nil {
		return err
	}
	// duplicate zone Infos
	status, err := c.zoneManager.Create(*infos)
	if err != nil {
		return err
	}
	zoneInfos, err := c.zoneManager.InfoByUUID(*status.UUID)
	if err != nil {
		return err
	}
	recordManager := c.zoneManager.Records(*zoneInfos)
	for _, record := range records {
		_, err := recordManager.Create(*record)
		if err != nil {
			return err
		}
	}
	_, err = c.zoneManager.Set(domainname, *zoneInfos)
	if err != nil {
		return err
	}

	return nil
}

// recordConfigFromInfo takes a DNS record from Gandi liveDNS and returns our native RecordConfig format.
func (c *liveClient) recordConfigFromInfo(infos []*gandiliverecord.Info, origin string) []*models.RecordConfig {
	rcs := []*models.RecordConfig{}
	for _, info := range infos {
		for i, value := range info.Values {
			rc := &models.RecordConfig{
				NameFQDN: dnsutil.AddOrigin(info.Name, origin),
				Name:     info.Name,
				Type:     info.Type,
				Original: info,
				Target:   value,
				TTL:      uint32(info.TTL),
			}
			switch info.Type {
			case "A", "AAAA", "NS", "CNAME", "PTR":
				// no-op
			case "TXT":
				value = strings.Join(info.Values, " ")
				rc.SetTxtParse(value)
				rc.Target = value
				if i > 0 {
					continue
				}
			case "CAA":
				var err error
				rc.CaaTag, rc.CaaFlag, rc.Target, err = models.SplitCombinedCaaValue(value)
				if err != nil {
					panic(fmt.Sprintf("gandi.convert bad caa value format: %#v (%s)", value, err))
				}
			case "SRV":
				var err error
				rc.SrvPriority, rc.SrvWeight, rc.SrvPort, rc.Target, err = models.SplitCombinedSrvValue(value)
				if err != nil {
					panic(fmt.Sprintf("gandi-livedns.convert bad srv value format: %#v (%s)", value, err))
				}
			case "MX":
				var err error
				rc.MxPreference, rc.Target, err = models.SplitCombinedMxValue(value)
				if err != nil {
					panic(fmt.Sprintf("gandi-livedns.convert bad mx value format: %#v", value))
				}
			default:
				panic(fmt.Sprintf("gandi-livedns.convert unimplemented rtype %v", info.Type))
				// We panic so that we quickly find any switch statements
				// that have not been updated for a new RR type.
			}
			rcs = append(rcs, rc)
		}
	}
	return rcs
}

// recordsToInfo generates gandi record sets and filters incompatible entries from native records format
func (c *liveClient) recordsToInfo(records models.Records) ([]*gandiliverecord.Info, error) {
	recordSets := map[string]map[string]*gandiliverecord.Info{}
	recordInfos := []*gandiliverecord.Info{}

	for _, rec := range records {
		if rec.TTL < 300 {
			log.Printf("WARNING: Gandi does not support ttls < 300. %s will not be set to %d.", rec.NameFQDN, rec.TTL)
			rec.TTL = 300
		}
		if rec.TTL > 2592000 {
			return nil, errors.Errorf("ERROR: Gandi does not support TTLs > 30 days (TTL=%d)", rec.TTL)
		}
		if rec.Type == "NS" && rec.Name == "@" {
			if !strings.HasSuffix(rec.Target, ".gandi.net.") {
				log.Printf("WARNING: Gandi does not support changing apex NS records. %s will not be added.", rec.Target)
			}
			continue
		}
		r, ok := recordSets[rec.Name][rec.Type]
		if !ok {
			_, ok := recordSets[rec.Name]
			if !ok {
				recordSets[rec.Name] = map[string]*gandiliverecord.Info{}
			}
			r = &gandiliverecord.Info{
				Type: rec.Type,
				Name: rec.Name,
				TTL:  int64(rec.TTL),
			}
			recordInfos = append(recordInfos, r)
			recordSets[rec.Name][rec.Type] = r
		} else {
			if r.TTL != int64(rec.TTL) {
				log.Printf(
					"WARNING: Gandi liveDNS API does not support different TTL for the couple fqdn/type. Will use TTL of %d for %s %s",
					r.TTL,
					r.Type,
					r.Name,
				)
			}
		}
		if rec.Type == "TXT" {
			for _, t := range rec.TxtStrings {
				r.Values = append(r.Values, "\""+t+"\"") // FIXME(tlim): Should do proper quoting.
			}
		} else {
			r.Values = append(r.Values, rec.Content())
		}
	}
	return recordInfos, nil
}
