//go:build windows
// +build windows

package msdns

import (
	"github.com/StackExchange/dnscontrol/v3/pkg/printer"
	ps "github.com/bhendo/go-powershell"
	"github.com/bhendo/go-powershell/backend"
	"github.com/bhendo/go-powershell/middleware"
)

type psHandle struct {
	shell ps.Shell
}

func newPowerShell(config map[string]string) (*psHandle, error) {

	back := &backend.Local{}
	sh, err := ps.New(back)
	if err != nil {
		return nil, err
	}
	shell := sh

	pssession := config["pssession"]
	if pssession != "" {
		printer.Printf("INFO: PowerShell commands will run on %q\n", pssession)
		// create a remote shell by wrapping the existing one in the session middleware
		mconfig := middleware.NewSessionConfig()
		mconfig.ComputerName = pssession

		cred := &middleware.UserPasswordCredential{
			Username: config["psusername"],
			Password: config["pspassword"],
		}
		if cred.Password != "" && cred.Username != "" {
			mconfig.Credential = cred
		}

		session, err := middleware.NewSession(sh, mconfig)
		if err != nil {
			panic(err)
		}
		shell = session
	}

	psh := &psHandle{
		shell: shell,
	}
	return psh, nil
}
