package fortigate

import (
	"encoding/json"
	"errors"

	"github.com/StackExchange/dnscontrol/v4/providers"
)

// ---- Features --------------------------------------------------------
var features = providers.DocumentationNotes{
	providers.CanGetZones:            providers.Can(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanConcur:              providers.Unimplemented(),
	providers.DocOfficiallySupported: providers.Cannot(),
	providers.CanUseLOC:              providers.Cannot(),
}

// ---- Register Provider ---------------------------------------------------------
func init() {
	const name = "FORTIGATE"
	providers.RegisterDomainServiceProviderType(name, providers.DspFuncs{
		Initializer:   NewFortiGate,
	}, features)
}

// ---- Provider‑Struct -------------------------------------------------------
type fortigateProvider struct {
	vdom     string
	host     string
	insecure bool
	apiKey   string

	client    *apiClient
}

// ---- Constructor -----------------------------------------------------------
func NewFortiGate(m map[string]string, _ json.RawMessage) (providers.DNSServiceProvider, error) {
	p := &fortigateProvider{
		vdom:      m["vdom"],
		host:      m["host"],
		apiKey:    m["apikey"],
		insecure:  m["insecure_tls"] == "true",
	}
	if p.vdom == "" || p.host == "" || p.apiKey == "" {
		return nil, errors.New("Fortigate Provider: need host, vdom and apikey")
	}
	p.client = newClient(p.host, p.vdom, p.apiKey, p.insecure)
	return p, nil
}
