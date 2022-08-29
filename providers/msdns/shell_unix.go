//go:build !windows
// +build !windows

package msdns

import "fmt"

type psHandle struct {
	shell *shellHandle
}

func newPowerShell(config map[string]string) (*psHandle, error) {

	remoteHost := "remotehost" // Pull this out of config

	psh := &psHandle{
		shell: newShellHandle(remoteHost),
	}
	return psh, nil
}

type shellHandle struct {
	remoteHost string
	// Add fields for anything needed in the session.
}

func newShellHandle(remoteHost string) *shellHandle {
	return &shellHandle{
		remoteHost: remoteHost,
	}
}

func (sh *shellHandle) Execute(s string) (string, string, error) {
	// NOT IMPLEMENTED
	// Run the command on sh.remoteHost
	return "", "", fmt.Errorf("Not implemented")
}

func (sh *shellHandle) Exit() {
	// NOT IMPLEMENTED
}
