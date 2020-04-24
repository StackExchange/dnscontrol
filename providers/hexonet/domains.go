package hexonet

//EnsureDomainExists returns an error
// * if access to dnszone is not allowed (not authorized) or
// * if it doesn't exist and creating it fails
func (n *HXClient) EnsureDomainExists(domain string) error {
	r := n.client.Request(map[string]string{
		"COMMAND": "StatusDNSZone",
		"DNSZONE": domain + ".",
	})
	code := r.GetCode()
	if code == 545 {
		r = n.client.Request(map[string]string{
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
