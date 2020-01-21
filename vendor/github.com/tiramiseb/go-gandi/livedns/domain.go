package livedns

import "github.com/tiramiseb/go-gandi/internal/client"

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

type zone struct {
	TTL int `json:"ttl"`
}

type createDomainRequest struct {
	FQDN string `json:"fqdn"`
	Zone zone `json:"zone,omitempty"`
}

type UpdateDomainRequest struct {
	AutomaticSnapshots *bool `json:"automatic_snapshots,omitempty"`
}

// ListDomains lists all domains
func (g *LiveDNS) ListDomains() (domains []Domain, err error) {
	_, err = g.client.Get("domains", nil, &domains)
	return
}

// CreateDomain adds a domain to a zone
func (g *LiveDNS) CreateDomain(fqdn string, ttl int) (response client.StandardResponse, err error) {
	_, err = g.client.Post("domains", createDomainRequest{FQDN: fqdn, Zone: zone{TTL: ttl}}, &response)
	return
}

// GetDomain returns a domain
func (g *LiveDNS) GetDomain(fqdn string) (domain Domain, err error) {
	_, err = g.client.Get("domains/"+fqdn, nil, &domain)
	return
}

// ChangeAssociatedZone changes the zone associated to a domain
func (g *LiveDNS) UpdateDomain(fqdn string, details UpdateDomainRequest) (response client.StandardResponse, err error) {
	_, err = g.client.Patch("domains/"+fqdn, details, &response)
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
