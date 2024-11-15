package cnr

// EnsureZoneExists returns an error
// * if access to dnszone is not allowed (not authorized) or
// * if it doesn't exist and creating it fails
func (n *CNRClient) EnsureZoneExists(domain string) error {
	r := n.client.Request(map[string]interface{}{
		"COMMAND": "StatusDNSZone",
		"DNSZONE": domain,
	})
	code := r.GetCode()
	if code == 545 {
		command := map[string]interface{}{
			"COMMAND": "AddDNSZone",
			"DNSZONE": domain,
		}
		if n.APIEntity == "OTE" {
			command["SOATTL"] = "33200"
			command["SOASERIAL"] = "0000000000"
		}
		// Create the zone
		r = n.client.Request(command)
		if !r.IsSuccess() {
			return n.GetCNRApiError("Failed to create not existing zone ", domain, r)
		}
	} else if code == 531 {
		return n.GetCNRApiError("Not authorized to manage dnszone", domain, r)
	} else if r.IsError() || r.IsTmpError() {
		return n.GetCNRApiError("Error while checking status of dnszone", domain, r)
	}
	return nil
}

// ListZones lists all the
func (n *CNRClient) ListZones() ([]string, error) {
	var zones []string

	// Basic

	rs := n.client.RequestAllResponsePages(map[string]string{
		"COMMAND": "QueryDNSZoneList",
	})
	for _, r := range rs {
		if r.IsError() {
			return nil, n.GetCNRApiError("Error while QueryDNSZoneList", "Basic", &r)
		}
		zoneColumn := r.GetColumn("DNSZONE")
		if zoneColumn != nil {
			//return nil, fmt.Errorf("failed getting DNSZONE BASIC column")
			zones = append(zones, zoneColumn.GetData()...)
		}
	}

	return zones, nil
}
