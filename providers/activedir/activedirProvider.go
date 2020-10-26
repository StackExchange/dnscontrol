package activedir

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/StackExchange/dnscontrol/v3/providers"
)

// This is the struct that matches either (or both) of the Registrar and/or DNSProvider interfaces:
type activedirProvider struct {
	adServer string
	fake     bool
	psOut    string
	psLog    string
}

var features = providers.DocumentationNotes{
	providers.CanUseAlias:            providers.Cannot(),
	providers.CanUseCAA:              providers.Cannot(),
	providers.CanUsePTR:              providers.Cannot(),
	providers.CanUseSRV:              providers.Cannot(),
	providers.DocCreateDomains:       providers.Cannot("AD depends on the zone already existing on the dns server"),
	providers.DocDualHost:            providers.Cannot("This driver does not manage NS records, so should not be used for dual-host scenarios"),
	providers.DocOfficiallySupported: providers.Can(),
	providers.CanGetZones:            providers.Unimplemented(),
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
