package livedns

import "github.com/tiramiseb/go-gandi-livedns/client"

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

// Nameservers represents a list of nameservers
type Nameservers struct {
	Nameservers []string `json:"nameservers,omitempty"`
}

// ListDomains lists all domains
func (g *LiveDNS) ListDomains() (domains []Domain, err error) {
	_, err = g.client.Get("domains", nil, &domains)
	return
}

// AddDomainToZone adds a domain to a zone
// It is equivalent to AttachDomainToZone, the only difference is the entry point in the LiveDNS API.
func (g *LiveDNS) AddDomainToZone(fqdn, uuid string) (response client.StandardResponse, err error) {
	_, err = g.client.Post("domains", Domain{FQDN: fqdn, ZoneUUID: uuid}, &response)
	return
}

// GetDomain returns a domain
func (g *LiveDNS) GetDomain(fqdn string) (domain Domain, err error) {
	_, err = g.client.Get("domains/"+fqdn, nil, &domain)
	return
}

// ChangeAssociatedZone changes the zone associated to a domain
func (g *LiveDNS) ChangeAssociatedZone(fqdn, uuid string) (response client.StandardResponse, err error) {
	_, err = g.client.Patch("domains/"+fqdn, Domain{ZoneUUID: uuid}, &response)
	return
}

// DetachDomain detaches a domain from the zone it is attached to
func (g *LiveDNS) DetachDomain(fqdn string) (err error) {
	_, err = g.client.Delete("domains/"+fqdn, nil, nil)
	return
}

// SignDomain creates a DNSKEY and asks Gandi servers to automatically sign the domain
func (g *LiveDNS) SignDomain(fqdn string) (response client.StandardResponse, err error) {
	f := SigningKey{Flags: 257}
	_, err = g.client.Post("domains/"+fqdn+"/keys", f, &response)
	return
}

// GetDomainKeys returns data about the signing keys created for a domain
func (g *LiveDNS) GetDomainKeys(fqdn string) (keys []SigningKey, err error) {
	_, err = g.client.Get("domains/"+fqdn+"/keys", nil, &keys)
	return
}

// DeleteDomainKey deletes a signing key from a domain
func (g *LiveDNS) DeleteDomainKey(fqdn, uuid string) (err error) {
	_, err = g.client.Delete("domains/"+fqdn+"/keys/"+uuid, nil, nil)
	return
}

// UpdateDomainKey updates a signing key for a domain (only the deleted status, actually...)
func (g *LiveDNS) UpdateDomainKey(fqdn, uuid string, deleted bool) (err error) {
	_, err = g.client.Put("domains/"+fqdn+"/keys/"+uuid, SigningKey{Deleted: &deleted}, nil)
	return
}

// GetDomainNS returns the list of the nameservers for a domain
func (g *LiveDNS) GetDomainNS(fqdn string) (ns []string, err error) {
	_, err = g.client.Get("nameservers/"+fqdn, nil, &ns)
	return
}

// UpdateDomainNS returns the list of the nameservers for a domain
func (g *LiveDNS) UpdateDomainNS(fqdn string, ns []string) (err error) {
	_, err = g.client.Put("domain/domains/"+fqdn+"/nameservers", Nameservers{ns}, nil)
	return
}
