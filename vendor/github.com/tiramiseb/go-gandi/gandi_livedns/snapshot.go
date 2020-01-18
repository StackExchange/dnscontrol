package gandi_livedns

import "github.com/tiramiseb/go-gandi/internal/client"

// Snapshot represents a zone snapshot
type Snapshot struct {
	UUID        string         `json:"uuid,omitempty"`
	DateCreated string         `json:"date_created,omitempty"`
	ZoneUUID    string         `json:"zone_uuid,omitempty"`
	ZoneData    []DomainRecord `json:"zone_data,omitempty"`
}

// ListSnapshots lists all domains
func (g *LiveDNS) ListSnapshots(fqdn string) (snapshots []Snapshot, err error) {
	_, err = g.client.Get("domains/"+fqdn+"/snapshots", nil, &snapshots)
	return
}

// CreateSnapshot creates a domain
func (g *LiveDNS) CreateSnapshot(fqdn string) (response client.StandardResponse, err error) {
	_, err = g.client.Post("domains/"+fqdn+"/snapshots", nil, &response)
	return
}

// GetSnapshot returns a domain
func (g *LiveDNS) GetSnapshot(fqdn, snapUUID string) (snapshot Snapshot, err error) {
	_, err = g.client.Get("domains/"+fqdn+"/snapshots/"+snapUUID, nil, &snapshot)
	return
}
