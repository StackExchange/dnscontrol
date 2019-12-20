package gandi

// Domain represents a DNS domain
type Domain struct {
	FQDN              string `json:"fqdn,omitempty"`
	DomainHref        string `json:"domain_href,omitempty"`
	DomainKeysHref    string `json:"domain_keys_href,omitempty"`
	DomainRecordsHref string `json:"domain_records_href,omitempty"`
	ZoneUUID          string `json:"zone_uuid,omitempty"`
	ZoneHref          string `json:"zone_href,omitempty"`
	ZoneRecordsHref   string `json:"zone_records_href,omitempty"`
}

// SigningKey holds data about a DNSSEC signing key
type SigningKey struct {
	Status        string `json:"status,omitempty"`
	UUID          string `json:"uuid,omitempty"`
	Algorithm     int    `json:"algorithm,omitempty"`
	Deleted       *bool  `json:"deleted"`
	AlgorithmName string `json:"algorithm_name,omitempty"`
	FQDN          string `json:"fqdn,omitempty"`
	Flags         int    `json:"flags,omitempty"`
	DS            string `json:"ds,omitempty"`
	KeyHref       string `json:"key_href,omitempty"`
}

// ListDomains lists all domains
func (g *Gandi) ListDomains() (domains []Domain, err error) {
	_, err = g.askGandi(mGET, "domains", nil, &domains)
	return
}

// AddDomainToZone adds a domain to a zone
// It is equivalent to AttachDomainToZone, the only difference is the entry point in the LiveDNS API.
func (g *Gandi) AddDomainToZone(fqdn, uuid string) (response StandardResponse, err error) {
	_, err = g.askGandi(mPOST, "domains", Domain{FQDN: fqdn, ZoneUUID: uuid}, &response)
	return
}

// GetDomain returns a domain
func (g *Gandi) GetDomain(fqdn string) (domain Domain, err error) {
	_, err = g.askGandi(mGET, "domains/"+fqdn, nil, &domain)
	return
}

// ChangeAssociatedZone changes the zone associated to a domain
func (g *Gandi) ChangeAssociatedZone(fqdn, uuid string) (response StandardResponse, err error) {
	_, err = g.askGandi(mPATCH, "domains/"+fqdn, Domain{ZoneUUID: uuid}, &response)
	return
}

// DetachDomain detaches a domain from the zone it is attached to
func (g *Gandi) DetachDomain(fqdn string) (err error) {
	_, err = g.askGandi(mDELETE, "domains/"+fqdn, nil, nil)
	return
}

// SignDomain creates a DNSKEY and asks Gandi servers to automatically sign the domain
func (g *Gandi) SignDomain(fqdn string) (response StandardResponse, err error) {
	f := SigningKey{Flags: 257}
	_, err = g.askGandi(mPOST, "domains/"+fqdn+"/keys", f, &response)
	return
}

// GetDomainKeys returns data about the signing keys created for a domain
func (g *Gandi) GetDomainKeys(fqdn string) (keys []SigningKey, err error) {
	_, err = g.askGandi(mGET, "domains/"+fqdn+"/keys", nil, &keys)
	return
}

// DeleteDomainKey deletes a signing key from a domain
func (g *Gandi) DeleteDomainKey(fqdn, uuid string) (err error) {
	_, err = g.askGandi(mDELETE, "domains/"+fqdn+"/keys/"+uuid, nil, nil)
	return
}

// UpdateDomainKey updates a signing key for a domain (only the deleted status, actually...)
func (g *Gandi) UpdateDomainKey(fqdn, uuid string, deleted bool) (err error) {
	_, err = g.askGandi(mPUT, "domains/"+fqdn+"/keys/"+uuid, SigningKey{Deleted: &deleted}, nil)
	return
}

// GetDomainNS returns the list of the nameservers for a domain
func (g *Gandi) GetDomainNS(fqdn string) (ns []string, err error) {
	_, err = g.askGandiFromBytes(mGET, "nameservers/"+fqdn, nil, &ns)
	return
}
