package gandi

// Tsig contains tsig data (no kidding!)
type Tsig struct {
	KeyName       string      `json:"key_name, omitempty"`
	Secret        string      `json:"secret,omitempty"`
	UUID          string      `json:"uuid,omitempty"`
	AxfrTsigURL   string      `json:"axfr_tsig_url,omitempty"`
	ConfigSamples interface{} `json:"config_samples,omitempty"`
}

// ListTsigs lists all tsigs
func (g *Gandi) ListTsigs() (tsigs []Tsig, err error) {
	_, err = g.askGandi(mGET, "axfr/tsig", nil, &tsigs)
	return
}

// GetTsig lists more tsig details
func (g *Gandi) GetTsig(uuid string) (tsig Tsig, err error) {
	_, err = g.askGandi(mGET, "axfr/tsig/"+uuid, nil, &tsig)
	return
}

// GetTsigBIND shows a BIND nameserver config, and includes the nameservers available for zone transfers
func (g *Gandi) GetTsigBIND(uuid string) ([]byte, error) {
	_, content, err := g.askGandiToBytes(mGET, "axfr/tsig/"+uuid+"/config/bind", nil)
	return content, err
}

// GetTsigPowerDNS shows a PowerDNS nameserver config, and includes the nameservers available for zone transfers
func (g *Gandi) GetTsigPowerDNS(uuid string) ([]byte, error) {
	_, content, err := g.askGandiToBytes(mGET, "axfr/tsig/"+uuid+"/config/powerdns", nil)
	return content, err
}

// GetTsigNSD shows a NSD nameserver config, and includes the nameservers available for zone transfers
func (g *Gandi) GetTsigNSD(uuid string) ([]byte, error) {
	_, content, err := g.askGandiToBytes(mGET, "axfr/tsig/"+uuid+"/config/nsd", nil)
	return content, err
}

// GetTsigKnot shows a Knot nameserver config, and includes the nameservers available for zone transfers
func (g *Gandi) GetTsigKnot(uuid string) ([]byte, error) {
	_, content, err := g.askGandiToBytes(mGET, "axfr/tsig/"+uuid+"/config/knot", nil)
	return content, err
}

// CreateTsig creates a tsig
func (g *Gandi) CreateTsig() (tsig Tsig, err error) {
	_, err = g.askGandi(mPOST, "axfr/tsig", nil, &tsig)
	return
}

// AddTsigToDomain adds a tsig to a domain
func (g *Gandi) AddTsigToDomain(fqdn, uuid string) (err error) {
	_, err = g.askGandi(mPUT, "domains/"+fqdn+"/axfr/tsig/"+uuid, nil, nil)
	return
}

// AddSlaveToDomain adds a slave to a domain
func (g *Gandi) AddSlaveToDomain(fqdn, host string) (err error) {
	_, err = g.askGandi(mPUT, "domains/"+fqdn+"/axfr/slaves/"+host, nil, nil)
	return
}

// ListSlavesInDomain lists slaves in a domain
func (g *Gandi) ListSlavesInDomain(fqdn string) (slaves []string, err error) {
	_, err = g.askGandi(mGET, "domains/"+fqdn+"/axfr/slaves", nil, &slaves)
	return
}

// DelSlaveFromDomain removes a slave from a domain
func (g *Gandi) DelSlaveFromDomain(fqdn, host string) (err error) {
	_, err = g.askGandi(mDELETE, "domains/"+fqdn+"/axfr/slaves/"+host, nil, nil)
	return
}
