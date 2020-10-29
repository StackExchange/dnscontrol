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

func (c *activedirProvider) getRecords(domainname string) ([]byte, error) {

	// If we are using PowerShell, make sure it is enabled
	// and then run the PS1 command to generate the adzonedump file.

	if !c.fake {
		checkPS.Do(func() {
			psAvailible = c.isPowerShellReady()
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

		_, err := c.powerShellExec(c.generatePowerShellZoneDump(domainname), true)
		if err != nil {
			return []byte{}, err
		}
	}
	// Return the contents of zone.*.json file instead.
	return c.readZoneDump(domainname)
}

func (c *activedirProvider) isPowerShellReady() bool {
	query, _ := c.powerShellExec(`(Get-Module -ListAvailable DnsServer) -ne $null`, true)
	q, err := strconv.ParseBool(strings.TrimSpace(string(query)))
	if err != nil {
		return false
	}
	return q
}

func (c *activedirProvider) powerShellDoCommand(command string, shouldLog bool) error {
	if c.fake {
		// If fake, just record the command.
		return c.powerShellRecord(command)
	}
	_, err := c.powerShellExec(command, shouldLog)
	return err
}

func (c *activedirProvider) powerShellExec(command string, shouldLog bool) ([]byte, error) {
	// log it.
	err := c.logCommand(command)
	if err != nil {
		return nil, err
	}

	// Run it.
	out, err := exec.Command("powershell", "-NoProfile", command).CombinedOutput()
	if err != nil {
		// If there was an error, log it.
		c.logErr(err)
	}
	if shouldLog {
		err = c.logOutput(string(out))
		if err != nil {
			return []byte{}, err
		}
	}

	// Return the result.
	return out, err
}
