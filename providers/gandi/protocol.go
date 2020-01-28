package gandi

import (
	"fmt"

	gandiclient "github.com/prasmussen/gandi-api/client"
	gandidomain "github.com/prasmussen/gandi-api/domain"
	gandinameservers "github.com/prasmussen/gandi-api/domain/nameservers"
	gandizone "github.com/prasmussen/gandi-api/domain/zone"
	gandirecord "github.com/prasmussen/gandi-api/domain/zone/record"
	gandiversion "github.com/prasmussen/gandi-api/domain/zone/version"
	gandioperation "github.com/prasmussen/gandi-api/operation"

	"github.com/StackExchange/dnscontrol/v2/models"
)

// fetchDomainList gets list of domains for account. Cache ids for easy lookup.
func (c *GandiApi) fetchDomainList() error {
	if c.domainIndex != nil {
		return nil
	}
	c.domainIndex = map[string]int64{}
	gc := gandiclient.New(c.ApiKey, gandiclient.Production)
	domain := gandidomain.New(gc)
	domains, err := domain.List()
	if err != nil {
		// fmt.Println(err)
		return err
	}
	for _, d := range domains {
		c.domainIndex[d.Fqdn] = d.Id
	}
	return nil
}

// fetchDomainInfo gets information about a domain.
func (c *GandiApi) fetchDomainInfo(fqdn string) (*gandidomain.DomainInfo, error) {
	gc := gandiclient.New(c.ApiKey, gandiclient.Production)
	domain := gandidomain.New(gc)
	return domain.Info(fqdn)
}

// setDomainNameservers updates the nameservers of a domain.
func (c *GandiApi) setDomainNameservers(fqdn string, nameservers []string) (*gandioperation.OperationInfo, error) {
	gc := gandiclient.New(c.ApiKey, gandiclient.Production)
	nameserversapi := gandinameservers.New(gc)
	return nameserversapi.Set(fqdn, nameservers)
}

// getRecordsForDomain returns a list of records for a zone.
func (c *GandiApi) getZoneRecords(zoneid int64, origin string) ([]*models.RecordConfig, error) {
	gc := gandiclient.New(c.ApiKey, gandiclient.Production)
	record := gandirecord.New(gc)
	recs, err := record.List(zoneid, 0)
	if err != nil {
		return nil, err
	}
	rcs := make([]*models.RecordConfig, 0, len(recs))
	for _, r := range recs {
		rcs = append(rcs, nativeToRecord(r, origin))
	}
	return rcs, nil
}

// convert takes a DNS record from Gandi and returns our native RecordConfig format.
func nativeToRecord(r *gandirecord.RecordInfo, origin string) *models.RecordConfig {

	rc := &models.RecordConfig{
		//NameFQDN: dnsutil.AddOrigin(r.Name, origin),
		//Name:     r.Name,
		//Type:     r.Type,
		TTL:      uint32(r.Ttl),
		Original: r,
		//Target:   r.Value,
	}
	rc.SetLabel(r.Name, origin)
	switch rtype := r.Type; rtype {
	default: //  "A", "AAAA", "CAA", "NS", "CNAME", "MX", "PTR", "SRV", "TXT"
		if err := rc.PopulateFromString(rtype, r.Value, origin); err != nil {
			panic(fmt.Errorf("unparsable record received from gandi: %w", err))
		}
	}
	return rc
}

// listZones retrieves the list of zones.
func (c *GandiApi) listZones() ([]*gandizone.ZoneInfoBase, error) {
	gc := gandiclient.New(c.ApiKey, gandiclient.Production)
	zone := gandizone.New(gc)
	return zone.List()
}

// setZone assigns a particular zone to a domain.
func (c *GandiApi) setZones(domainname string, zoneID int64) (*gandidomain.DomainInfo, error) {
	gc := gandiclient.New(c.ApiKey, gandiclient.Production)
	zone := gandizone.New(gc)
	return zone.Set(domainname, zoneID)
}

// getZoneInfo gets ZoneInfo about a zone.
func (c *GandiApi) getZoneInfo(zoneid int64) (*gandizone.ZoneInfo, error) {
	gc := gandiclient.New(c.ApiKey, gandiclient.Production)
	zone := gandizone.New(gc)
	return zone.Info(zoneid)
}

// createZone creates an entirely new zone.
func (c *GandiApi) createZone(name string) (*gandizone.ZoneInfo, error) {
	gc := gandiclient.New(c.ApiKey, gandiclient.Production)
	zone := gandizone.New(gc)
	return zone.Create(name)
}

func (c *GandiApi) getEditableZone(domainname string, zoneinfo *gandizone.ZoneInfo) (int64, error) {
	var zoneID int64
	if zoneinfo.Domains < 2 {
		// If there is only on{ domain linked to this zone, use it.
		zoneID = zoneinfo.Id
		fmt.Printf("Using zone id=%d named %#v\n", zoneID, zoneinfo.Name)
		return zoneID, nil
	}

	// We can't use the zone_id given to us. Let's make/find a new one.
	zones, err := c.listZones()
	if err != nil {
		return 0, err
	}
	zonename := fmt.Sprintf("%s dnscontrol", domainname)
	for _, z := range zones {
		if z.Name == zonename {
			zoneID = z.Id
			fmt.Printf("Recycling zone id=%d named %#v\n", zoneID, z.Name)
			return zoneID, nil
		}
	}
	zoneinfo, err = c.createZone(zonename)
	if err != nil {
		return 0, err
	}
	zoneID = zoneinfo.Id
	fmt.Printf("Created zone id=%d named %#v\n", zoneID, zoneinfo.Name)
	return zoneID, nil
}

// makeEditableZone
func (c *GandiApi) makeEditableZone(zoneID int64) (int64, error) {
	gc := gandiclient.New(c.ApiKey, gandiclient.Production)
	version := gandiversion.New(gc)
	return version.New(zoneID, 0)
}

// setZoneRecords
func (c *GandiApi) setZoneRecords(zoneID, versionID int64, records []gandirecord.RecordSet) ([]*gandirecord.RecordInfo, error) {
	gc := gandiclient.New(c.ApiKey, gandiclient.Production)
	record := gandirecord.New(gc)
	return record.SetRecords(zoneID, versionID, records)
}

// activateVersion
func (c *GandiApi) activateVersion(zoneID, versionID int64) (bool, error) {
	gc := gandiclient.New(c.ApiKey, gandiclient.Production)
	version := gandiversion.New(gc)
	return version.Set(zoneID, versionID)
}

func (c *GandiApi) createGandiZone(domainname string, zoneID int64, records []gandirecord.RecordSet) error {

	// Get the zone_id of the zone we'll be updating.
	zoneinfo, err := c.getZoneInfo(zoneID)
	if err != nil {
		return err
	}
	zoneID, err = c.getEditableZone(domainname, zoneinfo)
	if err != nil {
		return err
	}

	// Get the versionID of the zone we're updating.
	versionID, err := c.makeEditableZone(zoneID)
	if err != nil {
		return err
	}

	// Update the new version.
	_, err = c.setZoneRecords(zoneID, versionID, records)
	if err != nil {
		return err
	}

	// Activate zone version
	_, err = c.activateVersion(zoneID, versionID)
	if err != nil {
		return err
	}
	_, err = c.setZones(domainname, zoneID)
	if err != nil {
		return err
	}

	return nil
}
