package gandi

// Snapshot represents a zone snapshot
type Snapshot struct {
	UUID        string       `json:"uuid,omitempty"`
	DateCreated string       `json:"date_created,omitempty"`
	ZoneUUID    string       `json:"zone_uuid,omitempty"`
	ZoneData    []ZoneRecord `json:"zone_data,omitempty"`
}

// ListSnapshots lists all zones
func (g *Gandi) ListSnapshots(uuid string) (snapshots []Snapshot, err error) {
	_, err = g.askGandi(mGET, "zones/"+uuid+"/snapshots", nil, &snapshots)
	return
}

// CreateSnapshot creates a zone
func (g *Gandi) CreateSnapshot(uuid string) (response StandardResponse, err error) {
	_, err = g.askGandi(mPOST, "zones/"+uuid+"/snapshots", nil, &response)
	return
}

// GetSnapshot returns a zone
func (g *Gandi) GetSnapshot(uuid, snapUUID string) (snapshot Snapshot, err error) {
	_, err = g.askGandi(mGET, "zones/"+uuid+"/snapshots/"+snapUUID, nil, &snapshot)
	return
}
