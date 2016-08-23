package activedir

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func (c *adProvider) getRecords(domainname string) ([]byte, error) {

	if !*flagFakePowerShell {
		// If we are using PowerShell, make sure it is enabled
		// and then run the PS1 command to generate the adzonedump file.

		if !isPowerShellReady() {
			fmt.Printf("\n\n\n")
			fmt.Printf("***********************************************\n")
			fmt.Printf("PowerShell DnsServer module not installed.\n")
			fmt.Printf("See http://social.technet.microsoft.com/wiki/contents/articles/2202.remote-server-administration-tools-rsat-for-windows-client-and-windows-server-dsforum2wiki.aspx\n")
			fmt.Printf("***********************************************\n")
			fmt.Printf("\n\n\n")
			return nil, fmt.Errorf("PowerShell module DnsServer not installed.")
		}

		_, err := powerShellExecCombined(c.generatePowerShellZoneDump(domainname))
		if err != nil {
			return []byte{}, err
		}
	}

	// Return the contents of zone.*.json file instead.
	return c.readZoneDump(domainname)
}

func isPowerShellReady() bool {
	query, _ := powerShellExec(`(Get-Module -ListAvailable DnsServer) -ne $null`)
	q, err := strconv.ParseBool(strings.TrimSpace(string(query)))
	if err != nil {
		return false
	}
	return q
}

func powerShellDoCommand(command string) error {
	if *flagFakePowerShell {
		// If fake, just record the command.
		return powerShellRecord(command)
	}
	_, err := powerShellExec(command)
	return err
}

func powerShellExec(command string) ([]byte, error) {
	// log it.
	err := powerShellLogCommand(command)
	if err != nil {
		return []byte{}, err
	}

	// Run it.
	out, err := exec.Command("powershell", "-NoProfile", command).CombinedOutput()
	if err != nil {
		// If there was an error, log it.
		powerShellLogErr(err)
	}
	// Return the result.
	return out, err
}

// powerShellExecCombined runs a PS1 command and logs the output. This is useful when the output should be none or very small.
func powerShellExecCombined(command string) ([]byte, error) {
	// log it.
	err := powerShellLogCommand(command)
	if err != nil {
		return []byte{}, err
	}

	// Run it.
	out, err := exec.Command("powershell", "-NoProfile", command).CombinedOutput()
	if err != nil {
		// If there was an error, log it.
		powerShellLogErr(err)
		return out, err
	}

	// Log output.
	err = powerShellLogOutput(string(out))
	if err != nil {
		return []byte{}, err
	}

	// Return the result.
	return out, err
}
