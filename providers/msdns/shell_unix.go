//go:build !windows
// +build !windows

package msdns

import "fmt"

type psHandle struct {
	shell *shellHandle
}

func newPowerShell(config map[string]string) (*psHandle, error) {

	psh := &psHandle{
		shell: newShellHandle(),
	}
	return psh, nil
}

type shellHandle struct {
}

func newShellHandle() *shellHandle {
	return &shellHandle{}
}

func (*shellHandle) Execute(s string) (string, string, error) {
	// NOT IMPLEMENTED
	return "", "", fmt.Errorf("Not implemented")
}

func (*shellHandle) Exit() {
	// NOT IMPLEMENTED
}
