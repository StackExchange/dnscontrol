package gandi

import (
	"fmt"
	"log"
	"strings"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/miekg/dns/dnsutil"
	"github.com/pkg/errors"
	gandiclient "github.com/prasmussen/gandi-api/client"
	gandilivedomain "github.com/prasmussen/gandi-api/live_dns/domain"
	gandiliverecord "github.com/prasmussen/gandi-api/live_dns/record"
	gandilivezone "github.com/prasmussen/gandi-api/live_dns/zone"
)

// liveDNSGetNameservers returns the nameservers for domain.
func (c *GandiApi) liveDNSGetNameservers(domain string) ([]*models.Nameserver, error) {
	client := gandilivedomain.New(gandiclient.New(c.ApiKey, gandiclient.LiveDNS))
	domains, err := client.List()
	if err != nil {
		return nil, err
	}
	ns := []*models.Nameserver{}
	for _, domain := range domains {
		ns = append(ns, &models.Nameserver{Name: domain.Fqdn})
	}
	return ns, nil
}

// liveDNSGetZoneRecords returns a list of records for a zone.
func (c *GandiApi) liveDNSGetDomainRecords(origin string) ([]*models.RecordConfig, error) {
	client := gandilivedomain.New(gandiclient.New(c.ApiKey, gandiclient.LiveDNS))
	records, err := client.Records(origin).List()
	if err != nil {
		return nil, err
	}
	rcs := make([]*models.RecordConfig, 0)
	for _, record := range records {
		rcs = append(rcs, liveDNSConvert(record, origin))
	}
	return rcs, nil
}

func (c *GandiApi) liveDNSCreateGandiZone(domainname string, records []gandiliverecord.Info) error {
	domainInfo, err := gandilivedomain.New(gandiclient.New(c.ApiKey, gandiclient.LiveDNS)).Info(domainname)
	client := gandilivezone.New(gandiclient.New(c.ApiKey, gandiclient.LiveDNS))
	infos, err := client.InfoByUUID(*domainInfo.ZoneUUID)
	if err != nil {
		return err
	}
	// duplicate zone Infos
	status, err := client.Create(*infos)
	if err != nil {
		return err
	}
	zoneInfos, err := client.InfoByUUID(*status.UUID)
	if err != nil {
		return err
	}
	recordManager := client.Records(*zoneInfos)
	for _, record := range records {
		_, err := recordManager.Create(record)
		if err != nil {
			return err
		}
	}
	_, err = client.Set(domainname, *zoneInfos)
	if err != nil {
		return err
	}

	return nil
}

// liveDNSGetRecordUpdates generates gandi record sets and filters incompatible entries from native records format
func liveDNSGetRecordUpdates(records models.Records) ([]gandiliverecord.Info, []*models.RecordConfig, error) {
	expectedRecordSets := make([]gandiliverecord.Info, 0, len(records))
	recordsToKeep := make([]*models.RecordConfig, 0, len(records))
	for _, rec := range records {
		if rec.TTL < 300 {
			log.Printf("WARNING: Gandi does not support ttls < 300. %s will not be set to %d.", rec.NameFQDN, rec.TTL)
			rec.TTL = 300
		}
		if rec.TTL > 2592000 {
			return nil, nil, errors.Errorf("ERROR: Gandi does not support TTLs > 30 days (TTL=%d)", rec.TTL)
		}
		if rec.Type == "TXT" {
			rec.Target = "\"" + rec.Target + "\"" // FIXME(tlim): Should do proper quoting.
		}
		if rec.Type == "NS" && rec.Name == "@" {
			if !strings.HasSuffix(rec.Target, ".gandi.net.") {
				log.Printf("WARNING: Gandi does not support changing apex NS records. %s will not be added.", rec.Target)
			}
			continue
		}
		rs := gandiliverecord.Info{
			Type:   rec.Type,
			Name:   rec.Name,
			Values: strings.Split(rec.Target, ","),
			TTL:    int64(rec.TTL),
		}
		expectedRecordSets = append(expectedRecordSets, rs)
		recordsToKeep = append(recordsToKeep, rec)
	}
	return expectedRecordSets, recordsToKeep, nil
}

// liveDNSConvert takes a DNS record from Gandi liveDNS and returns our native RecordConfig format.
func liveDNSConvert(r *gandiliverecord.Info, origin string) *models.RecordConfig {
	value := strings.Join(r.Values, ",")
	rc := &models.RecordConfig{
		NameFQDN: dnsutil.AddOrigin(r.Name, origin),
		Name:     r.Name,
		Type:     r.Type,
		Original: r,
		Target:   value,
		TTL:      uint32(r.TTL),
	}
	switch r.Type {
	case "A", "AAAA", "NS", "CNAME", "PTR":
		// no-op
	case "TXT":
		rc.SetTxtParse(value)
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
			panic(fmt.Sprintf("gandi.convert bad srv value format: %#v (%s)", value, err))
		}
	case "MX":
		var err error
		rc.MxPreference, rc.Target, err = models.SplitCombinedMxValue(value)
		if err != nil {
			panic(fmt.Sprintf("gandi.convert bad mx value format: %#v", value))
		}
	default:
		panic(fmt.Sprintf("gandi.convert unimplemented rtype %v", r.Type))
		// We panic so that we quickly find any switch statements
		// that have not been updated for a new RR type.
	}
	return rc
}
