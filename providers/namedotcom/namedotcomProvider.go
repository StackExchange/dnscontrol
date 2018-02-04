// Package namedotcom implements a registrar that uses the name.com api to set name servers. It will self register it's providers when imported.
package namedotcom

import (
	"encoding/json"

	"github.com/StackExchange/dnscontrol/providers"
	"github.com/namedotcom/go/namecom"
	"github.com/pkg/errors"
)

const defaultAPIBase = "api.name.com"

// NameCom describes a connection to the NDC API.
type NameCom struct {
	APIUrl  string `json:"apiurl"`
	APIUser string `json:"apiuser"`
	APIKey  string `json:"apikey"`
	client  *namecom.NameCom
}

var features = providers.DocumentationNotes{
	providers.CanUseAlias:            providers.Can(),
	providers.CanUsePTR:              providers.Cannot("PTR records are not supported (See Link)", "https://www.name.com/support/articles/205188508-Reverse-DNS-records"),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseTXTMulti:         providers.Can(),
	providers.DocCreateDomains:       providers.Cannot("New domains require registration"),
	providers.DocDualHost:            providers.Cannot("Apex NS records not editable"),
	providers.DocOfficiallySupported: providers.Can(),
}

func newReg(conf map[string]string) (providers.Registrar, error) {
	return newProvider(conf)
}

func newDsp(conf map[string]string, meta json.RawMessage) (providers.DNSServiceProvider, error) {
	return newProvider(conf)
}

func newProvider(conf map[string]string) (*NameCom, error) {
	api := &NameCom{
		client: namecom.New(conf["apiuser"], conf["apikey"]),
	}
	api.client.Server = conf["apiurl"]
	api.APIUser, api.APIKey, api.APIUrl = conf["apiuser"], conf["apikey"], conf["apiurl"]
	if api.APIKey == "" || api.APIUser == "" {
		return nil, errors.Errorf("missing Name.com apikey or apiuser")
	}
	if api.APIUrl == "" {
		api.APIUrl = defaultAPIBase
	}
	return api, nil
}

func init() {
	providers.RegisterRegistrarType("NAMEDOTCOM", newReg)
	providers.RegisterDomainServiceProviderType("NAMEDOTCOM", newDsp, features)
}
