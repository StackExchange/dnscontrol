package cloudflare

// MagicTransitTunnelHealthcheck contains information about a tunnel health check.
type MagicTransitTunnelHealthcheck struct {
	Enabled   bool   `json:"enabled"`
	Target    string `json:"target,omitempty"`
	Type      string `json:"type,omitempty"`
	Rate      string `json:"rate,omitempty"`
	Direction string `json:"direction,omitempty"`
}
