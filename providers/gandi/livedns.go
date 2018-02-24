package gandi

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers"
	"github.com/StackExchange/dnscontrol/providers/diff"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	gandiclient "github.com/prasmussen/gandi-api/client"
	gandilivedomain "github.com/prasmussen/gandi-api/live_dns/domain"
	gandiliverecord "github.com/prasmussen/gandi-api/live_dns/record"
	gandilivezone "github.com/prasmussen/gandi-api/live_dns/zone"
)

// Enable/disable debug output:
const debug = false

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
		return nil, errors.Errorf("missing Gandi apikey")
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
		return nil, errors.Errorf("failed to get nameservers for domain %s", domain)
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
	records, err := c.domainManager.Records(dc.Name).List()
	if err != nil {
		return nil, err
	}
	foundRecords := c.recordConfigFromInfo(records, dc.Name)
	recordsToKeep, records, err := c.recordsToInfo(dc.Records)
	if err != nil {
		return nil, err
	}
	dc.Records = recordsToKeep

	// Normalize
	models.PostProcessRecords(foundRecords)

	differ := diff.New(dc)

	_, create, del, mod := differ.IncrementalDiff(foundRecords)
	if len(create)+len(del)+len(mod) > 0 {
		message := fmt.Sprintf("Setting dns records for %s:", dc.Name)
		for _, record := range dc.Records {
			message += "\n" + record.GetTargetCombined()
		}
		return []*models.Correction{
			{
				Msg: message,
				F: func() error {
					return c.createZone(dc.Name, records)
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
	infos.Name = fmt.Sprintf("zone created by dnscontrol for %s on %s", domainname, time.Now().Format(time.RFC3339))
	if debug {
		fmt.Printf("DEBUG: createZone SharingID=%v\n", infos.SharingID)
	}
	// duplicate zone Infos
	status, err := c.zoneManager.Create(*infos)
	if err != nil {
		return err
	}
	zoneInfos, err := c.zoneManager.InfoByUUID(*status.UUID)
	if err != nil {
		// gandi might take some time to make the new zone available
		for i := 0; i < 10; i++ {
			log.Printf("INFO: zone info not yet available. Delay and retry: %s", err.Error())
			time.Sleep(100 * time.Millisecond)
			zoneInfos, err = c.zoneManager.InfoByUUID(*status.UUID)
			if err == nil {
				break
			}
		}
	}
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
		// TXT records might have multiple values. In that case,
		// they are all for the TXT record at that label.
		if info.Type == "TXT" {
			rc := &models.RecordConfig{
				Type:     info.Type,
				Original: info,
				TTL:      uint32(info.TTL),
			}
			rc.SetLabel(info.Name, origin)
			var parsed []string
			for _, txt := range info.Values {
				parsed = append(parsed, models.StripQuotes(txt))
			}
			err := rc.SetTargetTXTs(parsed)
			if err != nil {
				panic(errors.Wrapf(err, "recordConfigFromInfo=TXT failed"))
			}
			rcs = append(rcs, rc)
		} else {
			// All other record types might have multiple values, but that means
			// we should create one Recordconfig for each one.
			for _, value := range info.Values {
				rc := &models.RecordConfig{
					Type:     info.Type,
					Original: info,
					TTL:      uint32(info.TTL),
				}
				rc.SetLabel(info.Name, origin)
				switch rtype := info.Type; rtype {
				default:
					err := rc.PopulateFromString(rtype, value, origin)
					if err != nil {
						panic(errors.Wrapf(err, "recordConfigFromInfo failed"))
					}
				}
				rcs = append(rcs, rc)
			}
		}
	}
	return rcs
}

// recordsToInfo generates gandi record sets and filters incompatible entries from native records format
func (c *liveClient) recordsToInfo(records models.Records) (models.Records, []*gandiliverecord.Info, error) {
	recordSets := map[string]map[string]*gandiliverecord.Info{}
	recordInfos := []*gandiliverecord.Info{}
	recordToKeep := models.Records{}

	for _, rec := range records {
		if rec.TTL < 300 {
			log.Printf("WARNING: Gandi does not support ttls < 300. %s will not be set to %d.", rec.NameFQDN, rec.TTL)
			rec.TTL = 300
		}
		if rec.TTL > 2592000 {
			return nil, nil, errors.Errorf("ERROR: Gandi does not support TTLs > 30 days (TTL=%d)", rec.TTL)
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
		recordToKeep = append(recordToKeep, rec)
		if rec.Type == "TXT" {
			for _, t := range rec.TxtStrings {
				r.Values = append(r.Values, "\""+t+"\"") // FIXME(tlim): Should do proper quoting.
			}
		} else {
			r.Values = append(r.Values, rec.GetTargetCombined())
		}
	}
	return recordToKeep, recordInfos, nil
}
