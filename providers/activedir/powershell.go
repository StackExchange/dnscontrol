package activedir

import (
	"encoding/json"
	"fmt"

	ps "github.com/bhendo/go-powershell"
	"github.com/bhendo/go-powershell/backend"
)

type psHandle struct {
	shell ps.Shell
}

func newPowerShell() (*psHandle, error) {

	back := &backend.Local{}

	sh, err := ps.New(back)
	if err != nil {
		return nil, err
	}
	//defer sh.Exit()

	psh := &psHandle{
		shell: sh,
	}
	return psh, nil

}

func (psh *psHandle) Exit() {
	psh.shell.Exit()
}

type dnsZone map[string]interface{}
type dnsZones []dnsZone

func (psh *psHandle) GetDNSServerZoneAll() ([]string, error) {
	stdout, stderr, err := psh.shell.Execute(`Get-DnsServerZone | ConvertTo-Json`)
	if err != nil {
		return nil, err
	}
	if stderr != "" {
		fmt.Printf("STDERROR = %q\n", stderr)
		return nil, fmt.Errorf("unexpected stderr from Get-DnsServerZones: %q", stderr)
	}

	var zones dnsZones
	json.Unmarshal([]byte(stdout), &zones)

	var result []string
	for _, z := range zones {
		zonename := z["ZoneName"].(string)
		result = append(result, zonename)
	}

	return result, nil
}
