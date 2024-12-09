// Package CNR implements a registrar that uses the CNR api to set name servers. It will self register it's providers when imported.
package cnr

import (
	"encoding/json"
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/providers"
	cnrcl "github.com/centralnicgroup-opensource/rtldev-middleware-go-sdk/v5/apiclient"
)

// GoReleaser: version
var (
	version = "dev"
)

// CNRClient describes a connection to the CNR API.
type CNRClient struct {
	conf        map[string]string
	APILogin    string
	APIPassword string
	APIEntity   string
	client      *cnrcl.APIClient
}

var features = providers.DocumentationNotes{
	// The default for unlisted capabilities is 'Cannot'.
	// See providers/capabilities.go for the entire list of capabilities.
	providers.CanGetZones:            providers.Can(),
	providers.CanConcur:              providers.Can(),
	providers.CanUseAlias:            providers.Cannot("Not supported. You may use CNAME records instead. An Alternative solution is planned."),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseLOC:              providers.Unimplemented(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseSRV:              providers.Can("SRV records with empty targets are not supported"),
	providers.CanUseTLSA:             providers.Can(),
	providers.DocCreateDomains:       providers.Can(),
	providers.DocDualHost:            providers.Can(),
	providers.DocOfficiallySupported: providers.Cannot("Actively maintained provider module."),
}

func newProvider(conf map[string]string) (*CNRClient, error) {
	api := &CNRClient{
		conf:   conf,
		client: cnrcl.NewAPIClient(),
	}
	api.client.SetUserAgent("DNSControl", version)
	api.APILogin, api.APIPassword, api.APIEntity = conf["apilogin"], conf["apipassword"], conf["apientity"]
	if conf["debugmode"] == "1" {
		api.client.EnableDebugMode()
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
	const providerName = "CNR"
	const providerMaintainer = "@KaiSchwarz-cnic"
	fns := providers.DspFuncs{
		Initializer:   newDsp,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterRegistrarType(providerName, newReg)
	providers.RegisterDomainServiceProviderType(providerName, fns, features)
	providers.RegisterMaintainer(providerName, providerMaintainer)
}
