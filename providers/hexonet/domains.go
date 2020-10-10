package hexonet

import "fmt"

//EnsureDomainExists returns an error
// * if access to dnszone is not allowed (not authorized) or
// * if it doesn't exist and creating it fails
func (n *HXClient) EnsureDomainExists(domain string) error {
	r := n.client.Request(map[string]interface{}{
		"COMMAND": "StatusDNSZone",
		"DNSZONE": domain + ".",
	})
	code := r.GetCode()
	if code == 545 {
		r = n.client.Request(map[string]interface{}{
			"COMMAND": "CreateDNSZone",
			"DNSZONE": domain + ".",
		})
		if !r.IsSuccess() {
			return n.GetHXApiError("Failed to create not existing zone for domain", domain, r)
		}
	} else if code == 531 {
		return n.GetHXApiError("Not authorized to manage dnszone", domain, r)
	} else if r.IsError() || r.IsError() {
		return n.GetHXApiError("Error while checking status of dnszone", domain, r)
	}
	return nil
}

// ListZones lists all the
func (n *HXClient) ListZones() ([]string, error) {
	var zones []string

	// Basic

	rs := n.client.RequestAllResponsePages(map[string]string{
		"COMMAND": "QueryDNSZoneList",
	})
	for _, r := range rs {
		if r.IsError() {
			return nil, n.GetHXApiError("Error while QueryDNSZoneList", "Basic", &r)
		}
		zoneColumn := r.GetColumn("DNSZONE")
		if zoneColumn == nil {
			return nil, fmt.Errorf("failed getting DNSZONE BASIC column")
		}
		zones = append(zones, zoneColumn.GetData()...)
	}

	// Premium

	rs = n.client.RequestAllResponsePages(map[string]string{
		"COMMAND":         "QueryDNSZoneList",
		"PROPERTIES":      "PREMIUMDNS",
		"PREMIUMDNSCLASS": "*",
	})
	for _, r := range rs {
		if r.IsError() {
			return nil, n.GetHXApiError("Error while QueryDNSZoneList", "Basic", &r)
		}
		zoneColumn := r.GetColumn("DNSZONE")
		if zoneColumn == nil {
			return nil, fmt.Errorf("failed getting DNSZONE PREMIUM column")
		}
		zones = append(zones, zoneColumn.GetData()...)
	}

	return zones, nil
}
