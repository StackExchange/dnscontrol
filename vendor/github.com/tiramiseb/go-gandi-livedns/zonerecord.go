package gandi

// ZoneRecord represents a DNS Record
type ZoneRecord struct {
	RrsetType   string   `json:"rrset_type,omitempty"`
	RrsetTTL    int      `json:"rrset_ttl,omitempty"`
	RrsetName   string   `json:"rrset_name,omitempty"`
	RrsetHref   string   `json:"rrset_href,omitempty"`
	RrsetValues []string `json:"rrset_values,omitempty"`
}

// ListZoneRecords lists all records in a zone
func (g *Gandi) ListZoneRecords(uuid string) (records []ZoneRecord, err error) {
	_, err = g.askGandi(mGET, "zones/"+uuid+"/records", nil, &records)
	return
}

// ListDomainRecords lists all records in the zone associated with a domain
func (g *Gandi) ListDomainRecords(fqdn string) (records []ZoneRecord, err error) {
	_, err = g.askGandi(mGET, "domains/"+fqdn+"/records", nil, &records)
	return
}

// ListZoneRecordsAsText lists all records in a zone and returns them as a text file
// ... and by text, I mean a slice of bytes
func (g *Gandi) ListZoneRecordsAsText(uuid string) ([]byte, error) {
	_, content, err := g.askGandiToBytes(mGET, "zones/"+uuid+"/records", nil)
	return content, err
}

// ListZoneRecordsWithName lists all records with a specific name in a zone
func (g *Gandi) ListZoneRecordsWithName(uuid, name string) (records []ZoneRecord, err error) {
	_, err = g.askGandi(mGET, "zones/"+uuid+"/records/"+name, nil, &records)
	return
}

// ListDomainRecordsWithName lists all records with a specific name in a zone
func (g *Gandi) ListDomainRecordsWithName(fqdn, name string) (records []ZoneRecord, err error) {
	_, err = g.askGandi(mGET, "domains/"+fqdn+"/records/"+name, nil, &records)
	return
}

// GetZoneRecordWithNameAndType gets the record with specific name and type in a zone
func (g *Gandi) GetZoneRecordWithNameAndType(uuid, name, recordtype string) (record ZoneRecord, err error) {
	_, err = g.askGandi(mGET, "zones/"+uuid+"/records/"+name+"/"+recordtype, nil, &record)
	return
}

// GetDomainRecordWithNameAndType gets the record with specific name and type in the zone attached to the domain
func (g *Gandi) GetDomainRecordWithNameAndType(fqdn, name, recordtype string) (record ZoneRecord, err error) {
	_, err = g.askGandi(mGET, "domains/"+fqdn+"/records/"+name+"/"+recordtype, nil, &record)
	return
}

// CreateZoneRecord creates a record in a zone
func (g *Gandi) CreateZoneRecord(uuid, name, recordtype string, ttl int, values []string) (response StandardResponse, err error) {
	_, err = g.askGandi(mPOST, "zones/"+uuid+"/records",
		ZoneRecord{
			RrsetType:   recordtype,
			RrsetTTL:    ttl,
			RrsetName:   name,
			RrsetValues: values,
		},
		&response)
	return
}

// CreateDomainRecord creates a record in the zone attached to a domain
func (g *Gandi) CreateDomainRecord(fqdn, name, recordtype string, ttl int, values []string) (response StandardResponse, err error) {
	_, err = g.askGandi(mPOST, "domains/"+fqdn+"/records",
		ZoneRecord{
			RrsetType:   recordtype,
			RrsetTTL:    ttl,
			RrsetName:   name,
			RrsetValues: values,
		},
		&response)
	return
}

type itemsPrefixForZoneRecords struct {
	Items []ZoneRecord `json:"items"`
}

// ChangeZoneRecords changes all records in a zone
func (g *Gandi) ChangeZoneRecords(uuid string, records []ZoneRecord) (response StandardResponse, err error) {
	prefixedRecords := itemsPrefixForZoneRecords{Items: records}
	_, err = g.askGandi(mPUT, "zones/"+uuid+"/records", prefixedRecords, &response)
	return
}

// ChangeDomainRecords changes all records in the zone attached to a domain
func (g *Gandi) ChangeDomainRecords(fqdn string, records []ZoneRecord) (response StandardResponse, err error) {
	prefixedRecords := itemsPrefixForZoneRecords{Items: records}
	_, err = g.askGandi(mPUT, "domains/"+fqdn+"/records", prefixedRecords, &response)
	return
}

// ChangeZoneRecordsAsText changes all zone records, taking them as text
// ... and by text, I mean a slice of bytes
func (g *Gandi) ChangeZoneRecordsAsText(uuid string, records []byte) (response StandardResponse, err error) {
	_, err = g.askGandiFromBytes(mPUT, "zones/"+uuid+"/records", records, &response)
	return
}

// ChangeZoneRecordsWithName changes all zone records with the given name
func (g *Gandi) ChangeZoneRecordsWithName(uuid, name string, records []ZoneRecord) (response StandardResponse, err error) {
	prefixedRecords := itemsPrefixForZoneRecords{Items: records}
	_, err = g.askGandi(mPUT, "zones/"+uuid+"/records/"+name, prefixedRecords, &response)
	return
}

// ChangeDomainRecordsWithName changes all records with the given name in the zone attached to the domain
func (g *Gandi) ChangeDomainRecordsWithName(fqdn, name string, records []ZoneRecord) (response StandardResponse, err error) {
	prefixedRecords := itemsPrefixForZoneRecords{Items: records}
	_, err = g.askGandi(mPUT, "domains/"+fqdn+"/records/"+name, prefixedRecords, &response)
	return
}

// ChangeZoneRecordWithNameAndType changes the zone record with the given name and the given type
func (g *Gandi) ChangeZoneRecordWithNameAndType(uuid, name, recordtype string, ttl int, values []string) (response StandardResponse, err error) {
	_, err = g.askGandi(mPUT, "zones/"+uuid+"/records/"+name+"/"+recordtype,
		ZoneRecord{
			RrsetTTL:    ttl,
			RrsetValues: values,
		},
		&response)
	return
}

// ChangeDomainRecordWithNameAndType changes the record with the given name and the given type in the zone attached to a domain
func (g *Gandi) ChangeDomainRecordWithNameAndType(fqdn, name, recordtype string, ttl int, values []string) (response StandardResponse, err error) {
	_, err = g.askGandi(mPUT, "domains/"+fqdn+"/records/"+name+"/"+recordtype,
		ZoneRecord{
			RrsetTTL:    ttl,
			RrsetValues: values,
		},
		&response)
	return
}

// DeleteAllZoneRecords deletes all records in a zone
func (g *Gandi) DeleteAllZoneRecords(uuid string) (err error) {
	_, err = g.askGandi(mDELETE, "zones/"+uuid+"/records", nil, nil)
	return
}

// DeleteAllDomainRecords deletes all records in the zone attached to a domain
func (g *Gandi) DeleteAllDomainRecords(fqdn string) (err error) {
	_, err = g.askGandi(mDELETE, "domains/"+fqdn+"/records", nil, nil)
	return
}

// DeleteZoneRecords deletes all records with the given name in a zone
func (g *Gandi) DeleteZoneRecords(uuid, name string) (err error) {
	_, err = g.askGandi(mDELETE, "zones/"+uuid+"/records/"+name, nil, nil)
	return
}

// DeleteDomainRecords deletes all records with the given name in the zone attached to a domain
func (g *Gandi) DeleteDomainRecords(fqdn, name string) (err error) {
	_, err = g.askGandi(mDELETE, "domains/"+fqdn+"/records/"+name, nil, nil)
	return
}

// DeleteZoneRecord deletes the record with the given name and the given type in a zone
func (g *Gandi) DeleteZoneRecord(uuid, name, recordtype string) (err error) {
	_, err = g.askGandi(mDELETE, "zones/"+uuid+"/records/"+name+"/"+recordtype, nil, nil)
	return
}

// DeleteDomainRecord deletes the record with the given name and the given type in the zone attached to a domain
func (g *Gandi) DeleteDomainRecord(fqdn, name, recordtype string) (err error) {
	_, err = g.askGandi(mDELETE, "domains/"+fqdn+"/records/"+name+"/"+recordtype, nil, nil)
	return
}
