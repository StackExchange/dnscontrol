// Package hexonet implements a registrar that uses the hexonet api to set name servers. It will self register it's providers when imported.
package hexonet

import (
	"encoding/json"
	"fmt"

	"github.com/StackExchange/dnscontrol/v3/pkg/version"
	"github.com/StackExchange/dnscontrol/v3/providers"
	hxcl "github.com/hexonet/go-sdk/apiclient"
)

// HXClient describes a connection to the hexonet API.
type HXClient struct {
	APILogin    string
	APIPassword string
	APIEntity   string
	client      *hxcl.APIClient
}

var features = providers.DocumentationNotes{
	providers.CanUseAlias:            providers.Cannot("Using ALIAS is possible through our extended DNS (X-DNS) service. Feel free to get in touch with us."),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseRoute53Alias:     providers.Cannot("Using ALIAS is possible through our extended DNS (X-DNS) service. Feel free to get in touch with us."),
	providers.CanUseSRV:              providers.Can("SRV records with empty targets are not supported"),
	providers.CanUseTLSA:             providers.Can(),
	providers.CanUseTXTMulti:         providers.Can(),
	providers.CantUseNOPURGE:         providers.Can(),
	providers.DocCreateDomains:       providers.Can(),
	providers.DocDualHost:            providers.Can(),
	providers.DocOfficiallySupported: providers.Cannot("Actively maintained provider module."),
	providers.CanGetZones:            providers.Unimplemented(),
}

func newProvider(conf map[string]string) (*HXClient, error) {
	api := &HXClient{
		client: hxcl.NewAPIClient(),
	}
	api.client.SetUserAgent("DNSControl", version.Banner())
	api.APILogin, api.APIPassword, api.APIEntity = conf["apilogin"], conf["apipassword"], conf["apientity"]
	if conf["debugmode"] == "1" {
		api.client.EnableDebugMode()
	}
	if len(conf["ipaddress"]) > 0 {
		api.client.SetRemoteIPAddress(conf["ipaddress"])
	}
	if api.APIEntity != "OTE" && api.APIEntity != "LIVE" {
		return nil, fmt.Errorf("wrong api system entity used. use \"OTE\" for OT&E system or \"LIVE\" for Live system")
	}
	if api.APIEntity == "OTE" {
		api.client.UseOTESystem()
	}
	if api.APILogin == "" || api.APIPassword == "" {
		return nil, fmt.Errorf("missing login credentials apilogin or apipassword")
	}
	api.client.SetCredentials(api.APILogin, api.APIPassword)
	return api, nil
}

func newReg(conf map[string]string) (providers.Registrar, error) {
	return newProvider(conf)
}

func newDsp(conf map[string]string, meta json.RawMessage) (providers.DNSServiceProvider, error) {
	return newProvider(conf)
}

func init() {
	providers.RegisterRegistrarType("HEXONET", newReg)
	providers.RegisterDomainServiceProviderType("HEXONET", newDsp, features)
}
