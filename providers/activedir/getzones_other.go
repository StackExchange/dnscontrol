// +build !windows

package activedir

func (c *activedirProvider) getRecords(domainname string) ([]byte, error) {
	if !c.fake {
		panic("Can not happen: PowerShell on non-windows")
	}
	return c.readZoneDump(domainname)
}

func (c *activedirProvider) powerShellDoCommand(command string, shouldLog bool) error {
	if !c.fake {
		panic("Can not happen: PowerShell on non-windows")
	}
	return c.powerShellRecord(command)
}
