package msdns

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/txtutil"
	"github.com/StackExchange/dnscontrol/v3/providers"
)

// This is the struct that matches either (or both) of the Registrar and/or DNSProvider interfaces:
type msdnsProvider struct {
	dnsserver string      // Which DNS Server to update
	pssession string      // Remote machine to PSSession to
	shell     DNSAccessor // Handle for
}

var features = providers.DocumentationNotes{
	providers.CanGetZones:            providers.Can(),
	providers.CanUseAlias:            providers.Cannot(),
	providers.CanUseCAA:              providers.Cannot(),
	providers.CanUseDS:               providers.Unimplemented(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseTLSA:             providers.Unimplemented(),
	providers.DocCreateDomains:       providers.Cannot("This provider assumes the zone already existing on the dns server"),
	providers.DocDualHost:            providers.Cannot("This driver does not manage NS records, so should not be used for dual-host scenarios"),
	providers.DocOfficiallySupported: providers.Can(),
}

// Register with the dnscontrol system.
//   This establishes the name (all caps), and the function to call to initialize it.
func init() {
	fns := providers.DspFuncs{
		Initializer:    newDNS,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType("MSDNS", fns, features)
}

func newDNS(config map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {

	if runtime.GOOS != "windows" {
		fmt.Println("INFO: PowerShell not available. Disabling Active Directory provider.")
		return providers.None{}, nil
	}

	var err error

	p := &msdnsProvider{
		dnsserver: config["dnsserver"],
	}
	p.shell, err = newPowerShell(config)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// Section 3: Domain Service Provider (DSP) related functions

// NB(tal): To future-proof your code, all new providers should
// implement GetDomainCorrections exactly as you see here
// (byte-for-byte the same). In 3.0
// we plan on using just the individual calls to GetZoneRecords,
// PostProcessRecords, and so on.
//
// Currently every provider does things differently, which prevents
// us from doing things like using GetZoneRecords() of a provider
// to make convertzone work with all providers.

// GetDomainCorrections get the current and existing records,
// post-process them, and generate corrections.
func (client *msdnsProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	existing, err := client.GetZoneRecords(dc.Name)
	if err != nil {
		return nil, err
	}
	models.PostProcessRecords(existing)
	txtutil.SplitSingleLongTxt(dc.Records) // Autosplit long TXT records

	clean := PrepFoundRecords(existing)
	PrepDesiredRecords(dc)
	return client.GenerateDomainCorrections(dc, clean)
}

// GetZoneRecords gathers the DNS records and converts them to
// dnscontrol's format.
func (client *msdnsProvider) GetZoneRecords(domain string) (models.Records, error) {

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

// PrepFoundRecords munges any records to make them compatible with
// this provider. Usually this is a no-op.
func PrepFoundRecords(recs models.Records) models.Records {
	// If there are records that need to be modified, removed, etc. we
	// do it here.  Usually this is a no-op.
	return recs
}

// PrepDesiredRecords munges any records to best suit this provider.
func PrepDesiredRecords(dc *models.DomainConfig) {
	// Sort through the dc.Records, eliminate any that can't be
	// supported; modify any that need adjustments to work with the
	// provider.  We try to do minimal changes otherwise it gets
	// confusing.

	dc.Punycode()
}

// NB(tlim): If we want to implement a registrar, refer to
// http://go.microsoft.com/fwlink/?LinkId=288158
// (Get-DnsServerZoneDelegation) for hints about which PowerShell
// commands to use.
