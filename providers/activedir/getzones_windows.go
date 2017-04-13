package activedir

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

var checkPS sync.Once
var psAvailible = false

func (c *adProvider) getRecords(domainname string) ([]byte, error) {

	// If we are using PowerShell, make sure it is enabled
	// and then run the PS1 command to generate the adzonedump file.

	if !*flagFakePowerShell {
		checkPS.Do(func() {
			psAvailible = isPowerShellReady()
			if !psAvailible {
				fmt.Printf("\n\n\n")
				fmt.Printf("***********************************************\n")
				fmt.Printf("PowerShell DnsServer module not installed.\n")
				fmt.Printf("See http://social.technet.microsoft.com/wiki/contents/articles/2202.remote-server-administration-tools-rsat-for-windows-client-and-windows-server-dsforum2wiki.aspx\n")
				fmt.Printf("***********************************************\n")
				fmt.Printf("\n\n\n")
			}
		})
		if !psAvailible {
			return nil, fmt.Errorf("powershell module DnsServer not installed")
		}

		_, err := powerShellExec(c.generatePowerShellZoneDump(domainname), true)
		if err != nil {
			return []byte{}, err
		}
	}
	// Return the contents of zone.*.json file instead.
	return c.readZoneDump(domainname)
}

func isPowerShellReady() bool {
	query, _ := powerShellExec(`(Get-Module -ListAvailable DnsServer) -ne $null`, true)
	q, err := strconv.ParseBool(strings.TrimSpace(string(query)))
	if err != nil {
		return false
	}
	return q
}

func powerShellDoCommand(command string, shouldLog bool) error {
	if *flagFakePowerShell {
		// If fake, just record the command.
		return powerShellRecord(command)
	}
	_, err := powerShellExec(command, shouldLog)
	return err
}

func powerShellExec(command string, shouldLog bool) ([]byte, error) {
	// log it.
	err := logCommand(command)
	if err != nil {
		return nil, err
	}

	// Run it.
	out, err := exec.Command("powershell", "-NoProfile", command).CombinedOutput()
	if err != nil {
		// If there was an error, log it.
		logErr(err)
	}
	if shouldLog {
		err = logOutput(string(out))
		if err != nil {
			return []byte{}, err
		}
	}

	// Return the result.
	return out, err
}
