package activedir

import (
	"encoding/json"
	"flag"
	"fmt"
	"runtime"

	"github.com/StackExchange/dnscontrol/providers"
)

var flagFakePowerShell = flag.Bool("fakeps", false, "ACTIVEDIR: Do not run PowerShell. Open adzonedump.*.json files for input, and write to -psout any PS1 commands that make changes.")
var flagPsFuture = flag.String("psout", "dns_update_commands.ps1", "ACTIVEDIR: Where to write PS1 commands for future execution.")
var flagPsLog = flag.String("pslog", "powershell.log", "ACTIVEDIR: filename of PS1 command log.")

// This is the struct that matches either (or both) of the Registrar and/or DNSProvider interfaces:
type adProvider struct {
	adServer string
}

// Register with the dnscontrol system.
//   This establishes the name (all caps), and the function to call to initialize it.
func init() {
	providers.RegisterDomainServiceProviderType("ACTIVEDIRECTORY_PS", newDNS)
}

func newDNS(config map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	if runtime.GOOS == "windows" || *flagFakePowerShell {
		srv := config["ADServer"]
		if srv == "" {
			return nil, fmt.Errorf("ADServer required for Active Directory provider")
		}
		return &adProvider{adServer: srv}, nil
	}
	fmt.Printf("WARNING: PowerShell not available. ActiveDirectory will not be updated.\n")
	return providers.None{}, nil
}
