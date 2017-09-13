// +build !windows

package activedir

func (c *adProvider) getRecords(domainname string) ([]byte, error) {
	if !c.fake {
		panic("Can not happen: PowerShell on non-windows")
	}
	return c.readZoneDump(domainname)
}

func (c *adProvider) powerShellDoCommand(command string, shouldLog bool) error {
	if !c.fake {
		panic("Can not happen: PowerShell on non-windows")
	}
	return c.powerShellRecord(command)
}
