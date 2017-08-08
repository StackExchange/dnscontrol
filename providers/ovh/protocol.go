package ovh

import (
	"fmt"
	"github.com/StackExchange/dnscontrol/models"
	"github.com/miekg/dns/dnsutil"
)

type Void struct {
}

// fetchDomainList gets list of zones for account
func (c *ovhProvider) fetchZones() error {
	if c.zones != nil {
		return nil
	}
	c.zones = map[string]bool{}

	var response []string

	err := c.client.Call("GET", "/domain/zone", nil, &response)

	if err != nil {
		return err
	}

	for _, d := range response {
		c.zones[d] = true
	}
	return nil
}

type Zone struct {
	LastUpdate      string   `json:"lastUpdate,omitempty"`
	HasDNSAnycast   bool     `json:"hasDNSAnycast,omitempty"`
	NameServers     []string `json:"nameServers"`
	DNSSecSupported bool     `json:"dnssecSupported"`
}

// get info about a zone.
func (c *ovhProvider) fetchZone(fqdn string) (*Zone, error) {
	var response Zone

	err := c.client.Call("GET", "/domain/zone/"+fqdn, nil, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

type Record struct {
	Target    string `json:"target,omitempty"`
	Zone      string `json:"zone,omitempty"`
	TTL       uint32 `json:"ttl,omitempty"`
	FieldType string `json:"fieldType,omitempty"`
	Id        int64  `json:"id,omitempty"`
	SubDomain string `json:"subDomain,omitempty"`
}

type records struct {
	recordsId []int
}

func (c *ovhProvider) fetchRecords(fqdn string) ([]*Record, error) {
	var recordIds []int

	err := c.client.Call("GET", "/domain/zone/"+fqdn+"/record", nil, &recordIds)
	if err != nil {
		return nil, err
	}

	records := make([]*Record, len(recordIds))
	for i, id := range recordIds {
		r, err := c.fecthRecord(fqdn, id)
		if err != nil {
			return nil, err
		}
		records[i] = r
	}

	return records, nil
}

func (c *ovhProvider) fecthRecord(fqdn string, id int) (*Record, error) {
	var response Record

	err := c.client.Call("GET", fmt.Sprintf("/domain/zone/%s/record/%d", fqdn, id), nil, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// Returns a function that can be invoked to delete a record in a zone.
func (c *ovhProvider) deleteRecordFunc(id int64, fqdn string) func() error {
	return func() error {
		err := c.client.Call("DELETE", fmt.Sprintf("/domain/zone/%s/record/%d", fqdn, id), nil, nil)
		if err != nil {
			return err
		}
		return nil
	}
}

// Returns a function that can be invoked to create a record in a zone.
func (c *ovhProvider) createRecordFunc(rc *models.RecordConfig, fqdn string) func() error {
	return func() error {
		record := Record{
			SubDomain: dnsutil.TrimDomainName(rc.NameFQDN, fqdn),
			FieldType: rc.Type,
			Target:    rc.Content(),
			TTL:       rc.TTL,
		}
		if record.SubDomain == "@" {
			record.SubDomain = ""
		}
		var response Record
		err := c.client.Call("POST", fmt.Sprintf("/domain/zone/%s/record", fqdn), &record, &response)
		return err
	}
}

// Returns a function that can be invoked to update a record in a zone.
func (c *ovhProvider) updateRecordFunc(old *Record, rc *models.RecordConfig, fqdn string) func() error {
	return func() error {
		record := Record{
			SubDomain: dnsutil.TrimDomainName(rc.NameFQDN, fqdn),
			FieldType: rc.Type,
			Target:    rc.Content(),
			TTL:       rc.TTL,
			Zone:      fqdn,
			Id:        old.Id,
		}
		if record.SubDomain == "@" {
			record.SubDomain = ""
		}

		return c.client.Call("PUT", fmt.Sprintf("/domain/zone/%s/record/%d", fqdn, old.Id), &record, &Void{})
	}
}

func (c *ovhProvider) refreshZone(fqdn string) error {
	return c.client.Call("POST", fmt.Sprintf("/domain/zone/%s/refresh", fqdn), nil, &Void{})
}

// fetch the NS OVH attributed to this zone (which is distinct from fetchRealNS which
// get the exact NS stored at the registrar
func (c *ovhProvider) fetchNS(fqdn string) ([]string, error) {
	zone, err := c.fetchZone(fqdn)
	if err != nil {
		return nil, err
	}

	return zone.NameServers, nil
}
