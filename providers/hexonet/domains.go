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

	// Basic

	r := n.client.Request(map[string]interface{}{
		"COMMAND": "QueryDNSZoneList",
	})
	if r.IsError() {
		return nil, n.GetHXApiError("Error while QueryDNSZoneList", "Basic", r)
	}
	basicColumn := r.GetColumn("DNSZONE")
	if basicColumn == nil {
		return nil, fmt.Errorf("failed getting DNSZONE BASIC column")
	}
	basic := basicColumn.GetData()

	// Premium

	r = n.client.Request(map[string]interface{}{
		"COMMAND":         "QueryDNSZoneList",
		"PROPERTIES":      "PREMIUMDNS",
		"PREMIUMDNSCLASS": "*",
	})
	if r.IsError() {
		return nil, n.GetHXApiError("Error while QueryDNSZoneList", "Basic", r)
	}
	premiumColumn := r.GetColumn("DNSZONE")
	if premiumColumn == nil {
		return nil, fmt.Errorf("failed getting DNSZONE PREMIUM column")
	}
	premium := premiumColumn.GetData()

	return append(basic, premium...), nil
}
