package gandi

import "strings"

// Zone represents a DNS Zone
type Zone struct {
	Retry           int    `json:"retry,omitempty"`
	UUID            string `json:"uuid,omitempty"`
	ZoneHref        string `json:"zone_href,omitempty"`
	Minimum         int    `json:"minimum,omitempty"`
	DomainsHref     string `json:"domains_href,omitempty"`
	Refresh         int    `json:"refresh,omitempty"`
	ZoneRecordsHref string `json:"zone_records_href,omitempty"`
	Expire          int    `json:"expire,omitempty"`
	SharingID       string `json:"sharing_id,omitempty"`
	Serial          int    `json:"serial,omitempty"`
	Email           string `json:"email,omitempty"`
	PrimaryNS       string `json:"primary_ns,omitempty"`
	Name            string `json:"name,omitempty"`
}

// ListZones lists all zones
func (g *Gandi) ListZones() (zones []Zone, err error) {
	_, err = g.askGandi(mGET, "livedns/zones", nil, &zones)
	return
}

// CreateZone creates a zone
func (g *Gandi) CreateZone(name string) (response StandardResponse, err error) {
	headers, err := g.askGandi(mPOST, "livedns/zones", Zone{Name: name}, &response)
	spLoc := strings.Split(headers.Get("Location"), "/")
	response.UUID = spLoc[len(spLoc)-1]
	return
}

// GetZone returns a zone
func (g *Gandi) GetZone(uuid string) (zone Zone, err error) {
	_, err = g.askGandi(mGET, "livedns/zones/"+uuid, nil, &zone)
	return
}

// UpdateZone updates a zone (only its name, actually...)
func (g *Gandi) UpdateZone(uuid, name string) (response StandardResponse, err error) {
	headers, err := g.askGandi(mPATCH, "livedns/zones/"+uuid, Zone{Name: name}, &response)
	spLoc := strings.Split(headers.Get("Location"), "/")
	response.UUID = spLoc[len(spLoc)-1]
	return
}

// DeleteZone deletes a zone
func (g *Gandi) DeleteZone(uuid string) (err error) {
	_, err = g.askGandi(mDELETE, "livedns/zones/"+uuid, nil, nil)
	return
}

// GetZoneDomains returns domains attached to a zone
func (g *Gandi) GetZoneDomains(uuid string) (domains []Domain, err error) {
	_, err = g.askGandi(mGET, "livedns/zones/"+uuid+"/domains", nil, &domains)
	return
}

// AttachDomainToZone attaches a domain to a zone
func (g *Gandi) AttachDomainToZone(uuid, fqdn string) (response StandardResponse, err error) {
	_, err = g.askGandi(mPOST, "livedns/zones/"+uuid+"/domains/"+fqdn, nil, &response)
	return
}
