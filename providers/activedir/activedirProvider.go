package activedir

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/providers"
)

// This is the struct that matches either (or both) of the Registrar and/or DNSProvider interfaces:
type activedirProvider struct {
	adServer string
	fake     bool
	psOut    string
	psLog    string
	// new fields here:
	shell DNSAccessor
}

var features = providers.DocumentationNotes{
	providers.CanUseAlias:            providers.Cannot(),
	providers.CanUseCAA:              providers.Cannot(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseSRV:              providers.Cannot(),
	providers.DocCreateDomains:       providers.Cannot("AD depends on the zone already existing on the dns server"),
	providers.DocDualHost:            providers.Cannot("This driver does not manage NS records, so should not be used for dual-host scenarios"),
	providers.DocOfficiallySupported: providers.Can(),
	providers.CanGetZones:            providers.Can(),
}

// Register with the dnscontrol system.
//   This establishes the name (all caps), and the function to call to initialize it.
func init() {
	providers.RegisterDomainServiceProviderType("ACTIVEDIRECTORY_PS", newDNS, features)
}

func newDNS(config map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {

	fake := false
	if fVal := config["fakeps"]; fVal == "true" {
		fake = true
	} else if fVal != "" && fVal != "false" {
		return nil, fmt.Errorf("fakeps value must be 'true' or 'false'")
	}

	psOut, psLog := config["psout"], config["pslog"]
	if psOut == "" {
		psOut = "dns_update_commands.ps1"
	}
	if psLog == "" {
		psLog = "powershell.log"
	}

	p := &activedirProvider{psLog: psLog, psOut: psOut, fake: fake}
	var err error
	p.shell, err = newPowerShell()
	if err != nil {
		return nil, err
	}

	if fake {
		return p, nil
	}
	if runtime.GOOS == "windows" {
		srv := config["ADServer"]
		if srv == "" {
			return nil, fmt.Errorf("ADServer required for Active Directory provider")
		}
		p.adServer = srv
		return p, nil
	}
	fmt.Printf("WARNING: PowerShell not available. Active Directory will not be updated.\n")
	return providers.None{}, nil
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
func (client *activedirProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	existing, err := client.GetZoneRecords(dc.Name)
	if err != nil {
		return nil, err
	}
	models.PostProcessRecords(existing)
	clean := PrepFoundRecords(existing)
	PrepDesiredRecords(dc)
	return client.GenerateDomainCorrections(dc, clean)
}

// GetZoneRecords gathers the DNS records and converts them to
// dnscontrol's format.
func (client *activedirProvider) GetZoneRecords(domain string) (models.Records, error) {

	// Get the existing DNS records in native format.
	nativeExistingRecords, err := client.shell.GetDNSZoneRecords(domain)
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

	// 	recordsToKeep := make([]*models.RecordConfig, 0, len(dc.Records))
	// 	for _, rec := range dc.Records {
	// 		if rec.Type == "ALIAS" && rec.Name != "@" {
	// 			// GANDI only permits aliases on a naked domain.
	// 			// Therefore, we change this to a CNAME.
	// 			rec.Type = "CNAME"
	// 		}
	// 		if rec.TTL < 300 {
	// 			printer.Warnf("Gandi does not support ttls < 300. Setting %s from %d to 300\n", rec.GetLabelFQDN(), rec.TTL)
	// 			rec.TTL = 300
	// 		}
	// 		if rec.TTL > 2592000 {
	// 			printer.Warnf("Gandi does not support ttls > 30 days. Setting %s from %d to 2592000\n", rec.GetLabelFQDN(), rec.TTL)
	// 			rec.TTL = 2592000
	// 		}
	// 		if rec.Type == "TXT" {
	// 			rec.SetTarget("\"" + rec.GetTargetField() + "\"") // FIXME(tlim): Should do proper quoting.
	// 		}
	// 		if rec.Type == "NS" && rec.GetLabel() == "@" {
	// 			if !strings.HasSuffix(rec.GetTargetField(), ".gandi.net.") {
	// 				printer.Warnf("Gandi does not support changing apex NS records. Ignoring %s\n", rec.GetTargetField())
	// 			}
	// 			continue
	// 		}
	// 		recordsToKeep = append(recordsToKeep, rec)
	// 	}
	// 	dc.Records = recordsToKeep
}
