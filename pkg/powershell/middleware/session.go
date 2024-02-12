// Copyright (c) 2017 Gorillalabs. All rights reserved.

package middleware

import (
	"fmt"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/pkg/powershell/utils"
	"github.com/juju/errors"
)

type session struct {
	upstream Middleware
	name     string
}

func NewSession(upstream Middleware, config *SessionConfig) (Middleware, error) {
	asserted, ok := config.Credential.(credential)
	if ok {
		credentialParamValue, err := asserted.prepare(upstream)
		if err != nil {
			return nil, errors.Annotate(err, "Could not setup credentials")
		}

		config.Credential = credentialParamValue
	}

	name := "goSess" + utils.CreateRandomString(8)
	args := strings.Join(config.ToArgs(), " ")

	_, _, err := upstream.Execute(fmt.Sprintf("$%s = New-PSSession %s", name, args))
	if err != nil {
		return nil, errors.Annotate(err, "Could not create new PSSession")
	}

	return &session{upstream, name}, nil
}

func (s *session) Execute(cmd string) (string, string, error) {
	return s.upstream.Execute(fmt.Sprintf("Invoke-Command -Session $%s -Script {%s}", s.name, cmd))
}

func (s *session) Exit() {
	s.upstream.Execute(fmt.Sprintf("Disconnect-PSSession -Session $%s", s.name))
	s.upstream.Exit()
}
