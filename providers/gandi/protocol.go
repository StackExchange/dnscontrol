package gandi

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/providers/diff"
)

import (
	gandiclient "github.com/prasmussen/gandi-api/client"
	gandidomain "github.com/prasmussen/gandi-api/domain"
	gandizone "github.com/prasmussen/gandi-api/domain/zone"
	gandirecord "github.com/prasmussen/gandi-api/domain/zone/record"
	gandiversion "github.com/prasmussen/gandi-api/domain/zone/version"
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
		//	fmt.Println(err)
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

// getRecordsForDomain returns a list of records for a zone.
func (c *GandiApi) getZoneRecords(zoneid int64) ([]*gandirecord.RecordInfo, error) {
	gc := gandiclient.New(c.ApiKey, gandiclient.Production)
	record := gandirecord.New(gc)
	return record.List(zoneid, 0)
}

// listZones retrieves the list of zones.
func (c *GandiApi) listZones() ([]*gandizone.ZoneInfoBase, error) {
	gc := gandiclient.New(c.ApiKey, gandiclient.Production)
	zone := gandizone.New(gc)
	return zone.List()
}

// setZone assigns a particular zone to a domain.
func (c *GandiApi) setZones(domainname string, zone_id int64) (*gandidomain.DomainInfo, error) {
	gc := gandiclient.New(c.ApiKey, gandiclient.Production)
	zone := gandizone.New(gc)
	return zone.Set(domainname, zone_id)
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

// replaceZoneContents
func (c *GandiApi) replaceZoneContents(zone_id int64, version_id int64, records []diff.Record) error {
	return fmt.Errorf("replaceZoneContents unimplemented")
}

func (c *GandiApi) getEditableZone(domainname string, zoneinfo *gandizone.ZoneInfo) (int64, error) {
	var zone_id int64
	if zoneinfo.Domains < 2 {
		// If there is only on{ domain linked to this zone, use it.
		zone_id = zoneinfo.Id
		fmt.Printf("Using zone id=%d named %#v\n", zone_id, zoneinfo.Name)
		return zone_id, nil
	}

	// We can't use the zone_id given to us. Let's make/find a new one.
	zones, err := c.listZones()
	if err != nil {
		return 0, err
	}
	zonename := fmt.Sprintf("%s dnscontrol", domainname)
	for _, z := range zones {
		if z.Name == zonename {
			zone_id = z.Id
			fmt.Printf("Recycling zone id=%d named %#v\n", zone_id, z.Name)
			return zone_id, nil
		}
	}
	zoneinfo, err = c.createZone(zonename)
	if err != nil {
		return 0, err
	}
	zone_id = zoneinfo.Id
	fmt.Printf("Created zone id=%d named %#v\n", zone_id, zoneinfo.Name)
	return zone_id, nil
}

// makeEditableZone
func (c *GandiApi) makeEditableZone(zone_id int64) (int64, error) {
	gc := gandiclient.New(c.ApiKey, gandiclient.Production)
	version := gandiversion.New(gc)
	return version.New(zone_id, 0)
}

// setZoneRecords
func (c *GandiApi) setZoneRecords(zone_id, version_id int64, records []gandirecord.RecordSet) ([]*gandirecord.RecordInfo, error) {
	gc := gandiclient.New(c.ApiKey, gandiclient.Production)
	record := gandirecord.New(gc)
	return record.SetRecords(zone_id, version_id, records)
}

// activateVersion
func (c *GandiApi) activateVersion(zone_id, version_id int64) (bool, error) {
	gc := gandiclient.New(c.ApiKey, gandiclient.Production)
	version := gandiversion.New(gc)
	return version.Set(zone_id, version_id)
}

func (c *GandiApi) createGandiZone(domainname string, zone_id int64, records []gandirecord.RecordSet) error {

	// Get the zone_id of the zone we'll be updating.
	zoneinfo, err := c.getZoneInfo(zone_id)
	if err != nil {
		return err
	}
	//fmt.Println("ZONEINFO:", zoneinfo)
	zone_id, err = c.getEditableZone(domainname, zoneinfo)
	if err != nil {
		return err
	}

	// Get the version_id of the zone we're updating.
	version_id, err := c.makeEditableZone(zone_id)
	if err != nil {
		return err
	}

	// Update the new version.
	_, err = c.setZoneRecords(zone_id, version_id, records)
	if err != nil {
		return err
	}

	// Activate zone version
	_, err = c.activateVersion(zone_id, version_id)
	if err != nil {
		return err
	}
	_, err = c.setZones(domainname, zone_id)
	if err != nil {
		return err
	}

	return nil
}
