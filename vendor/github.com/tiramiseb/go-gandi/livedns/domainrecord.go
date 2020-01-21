package livedns

import "github.com/tiramiseb/go-gandi/internal/client"

// DomainRecord represents a DNS Record
type DomainRecord struct {
	RrsetType   string   `json:"rrset_type,omitempty"`
	RrsetTTL    int      `json:"rrset_ttl,omitempty"`
	RrsetName   string   `json:"rrset_name,omitempty"`
	RrsetHref   string   `json:"rrset_href,omitempty"`
	RrsetValues []string `json:"rrset_values,omitempty"`
}

// ListDomainRecords lists all records in the zone associated with a domain
func (g *LiveDNS) ListDomainRecords(fqdn string) (records []DomainRecord, err error) {
	_, err = g.client.Get("domains/"+fqdn+"/records", nil, &records)
	return
}

// ListDomainRecordsAsText lists all records in a zone and returns them as a text file
// ... and by text, I mean a slice of bytes
func (g *LiveDNS) ListDomainRecordsAsText(uuid string) ([]byte, error) {
	_, content, err := g.client.GetBytes("domains/"+uuid+"/records", nil)
	return content, err
}

// ListDomainRecordsWithName lists all records with a specific name in a zone
func (g *LiveDNS) ListDomainRecordsWithName(fqdn, name string) (records []DomainRecord, err error) {
	_, err = g.client.Get("domains/"+fqdn+"/records/"+name, nil, &records)
	return
}

// GetDomainRecordWithNameAndType gets the record with specific name and type in the zone attached to the domain
func (g *LiveDNS) GetDomainRecordWithNameAndType(fqdn, name, recordtype string) (record DomainRecord, err error) {
	_, err = g.client.Get("domains/"+fqdn+"/records/"+name+"/"+recordtype, nil, &record)
	return
}

// CreateDomainRecord creates a record in the zone attached to a domain
func (g *LiveDNS) CreateDomainRecord(fqdn, name, recordtype string, ttl int, values []string) (response client.StandardResponse, err error) {
	_, err = g.client.Post("domains/"+fqdn+"/records",
		DomainRecord{
			RrsetType:   recordtype,
			RrsetTTL:    ttl,
			RrsetName:   name,
			RrsetValues: values,
		},
		&response)
	return
}

type itemsPrefixForZoneRecords struct {
	Items []DomainRecord `json:"items"`
}

// ChangeDomainRecords changes all records in the zone attached to a domain
func (g *LiveDNS) ChangeDomainRecords(fqdn string, records []DomainRecord) (response client.StandardResponse, err error) {
	prefixedRecords := itemsPrefixForZoneRecords{Items: records}
	_, err = g.client.Put("domains/"+fqdn+"/records", prefixedRecords, &response)
	return
}

// ChangeDomainRecordsWithName changes all records with the given name in the zone attached to the domain
func (g *LiveDNS) ChangeDomainRecordsWithName(fqdn, name string, records []DomainRecord) (response client.StandardResponse, err error) {
	prefixedRecords := itemsPrefixForZoneRecords{Items: records}
	_, err = g.client.Put("domains/"+fqdn+"/records/"+name, prefixedRecords, &response)
	return
}

// ChangeDomainRecordWithNameAndType changes the record with the given name and the given type in the zone attached to a domain
func (g *LiveDNS) ChangeDomainRecordWithNameAndType(fqdn, name, recordtype string, ttl int, values []string) (response client.StandardResponse, err error) {
	_, err = g.client.Put("domains/"+fqdn+"/records/"+name+"/"+recordtype,
		DomainRecord{
			RrsetTTL:    ttl,
			RrsetValues: values,
		},
		&response)
	return
}

// DeleteAllDomainRecords deletes all records in the zone attached to a domain
func (g *LiveDNS) DeleteAllDomainRecords(fqdn string) (err error) {
	_, err = g.client.Delete("domains/"+fqdn+"/records", nil, nil)
	return
}

// DeleteDomainRecords deletes all records with the given name in the zone attached to a domain
func (g *LiveDNS) DeleteDomainRecords(fqdn, name string) (err error) {
	_, err = g.client.Delete("domains/"+fqdn+"/records/"+name, nil, nil)
	return
}

// DeleteDomainRecord deletes the record with the given name and the given type in the zone attached to a domain
func (g *LiveDNS) DeleteDomainRecord(fqdn, name, recordtype string) (err error) {
	_, err = g.client.Delete("domains/"+fqdn+"/records/"+name+"/"+recordtype, nil, nil)
	return
}
