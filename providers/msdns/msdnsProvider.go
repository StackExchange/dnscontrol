package msdns

import (
	"encoding/json"
	"runtime"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/printer"
	"github.com/StackExchange/dnscontrol/v4/providers"
)

// This is the struct that matches either (or both) of the Registrar and/or DNSProvider interfaces:
type msdnsProvider struct {
	dnsserver  string      // Which DNS Server to update
	pssession  string      // Remote machine to PSSession to
	psusername string      // Remote username for PSSession
	pspassword string      // Remote password for PSSession
	shell      DNSAccessor // Handle for
}

var features = providers.DocumentationNotes{
	// The default for unlisted capabilities is 'Cannot'.
	// See providers/capabilities.go for the entire list of capabilities.
	providers.CanGetZones:            providers.Can(),
	providers.CanConcur:              providers.Cannot(),
	providers.CanUseAlias:            providers.Cannot(),
	providers.CanUseCAA:              providers.Cannot(),
	providers.CanUseDS:               providers.Unimplemented(),
	providers.CanUseLOC:              providers.Cannot(),
	providers.CanUseNAPTR:            providers.Can(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseTLSA:             providers.Unimplemented(),
	providers.DocCreateDomains:       providers.Cannot("This provider assumes the zone already existing on the dns server"),
	providers.DocDualHost:            providers.Cannot("This driver does not manage NS records, so should not be used for dual-host scenarios"),
	providers.DocOfficiallySupported: providers.Can(),
}

// Register with the dnscontrol system.
//
//	This establishes the name (all caps), and the function to call to initialize it.
func init() {
	const providerName = "MSDNS"
	const providerMaintainer = "@tlimoncelli"
	fns := providers.DspFuncs{
		Initializer:   newDNS,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType(providerName, fns, features)
	providers.RegisterMaintainer(providerName, providerMaintainer)
}

func newDNS(config map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {

	if runtime.GOOS != "windows" {
		printer.Println("INFO: MSDNS deactivated. Required OS not detected.")
		return providers.None{}, nil
	}

	var err error

	p := &msdnsProvider{
		dnsserver:  config["dnsserver"],
		pssession:  config["pssession"],
		psusername: config["psusername"],
		pspassword: config["pspassword"],
	}
	p.shell, err = newPowerShell(config)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// Section 3: Domain Service Provider (DSP) related functions

// GetZoneRecords gathers the DNS records and converts them to
// dnscontrol's format.
func (client *msdnsProvider) GetZoneRecords(domain string, meta map[string]string) (models.Records, error) {

	// Get the existing DNS records in native format.
	nativeExistingRecords, err := client.shell.GetDNSZoneRecords(client.dnsserver, domain)
	if err != nil {
		return nil, err
	}
	// Convert them to DNScontrol's native format:
	existingRecords := make([]*models.RecordConfig, 0, len(nativeExistingRecords))
	for _, rr := range nativeExistingRecords {
		rc, err := nativeToRecords(rr, domain)
		if err != nil {
			return nil, err
		}
		if rc != nil {
			existingRecords = append(existingRecords, rc)
		}
	}

	return existingRecords, nil
}

// NB(tlim): If we want to implement a registrar, refer to
// http://go.microsoft.com/fwlink/?LinkId=288158
// (Get-DnsServerZoneDelegation) for hints about which PowerShell
// commands to use.
