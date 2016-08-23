// +build !windows

package activedir

func (c *adProvider) getRecords(domainname string) ([]byte, error) {
	if !*flagFakePowerShell {
		panic("Can not happen: PowerShell on non-windows")
	}
	return c.readZoneDump(domainname)
}

func powerShellDoCommand(command string) error {
	if !*flagFakePowerShell {
		panic("Can not happen: PowerShell on non-windows")
	}
	return powerShellRecord(command)
}
