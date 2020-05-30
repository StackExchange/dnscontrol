package servers

// Server models a PowerDNS server.
//
// More information: https://doc.powerdns.com/authoritative/http-api/server.html#server
type Server struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	DaemonType string `json:"daemon_type"`
	Version    string `json:"version"`
	URL        string `json:"url,omitempty"`
	ConfigURL  string `json:"config_url,omitempty"`
	ZonesURL   string `json:"zones_url,omitempty"`
}
