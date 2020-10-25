// +build !windows

package activedir

func (c *activedirAPI) getRecords(domainname string) ([]byte, error) {
	if !c.fake {
		panic("Can not happen: PowerShell on non-windows")
	}
	return c.readZoneDump(domainname)
}

func (c *activedirAPI) powerShellDoCommand(command string, shouldLog bool) error {
	if !c.fake {
		panic("Can not happen: PowerShell on non-windows")
	}
	return c.powerShellRecord(command)
}
