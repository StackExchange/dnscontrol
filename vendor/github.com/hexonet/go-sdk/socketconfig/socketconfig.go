// Copyright (c) 2018 Kai Schwarz (HEXONET GmbH). All rights reserved.
//
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE.md file.

// Package socketconfig provides apiconnector client connection settings
package socketconfig

import (
	"net/url"
	"strings"
)

// SocketConfig is a struct representing connection settings used as POST data for http request against the insanely fast HEXONET backend API.
type SocketConfig struct {
	entity     string
	login      string
	otp        string
	pw         string
	remoteaddr string
	session    string
	user       string
}

// NewSocketConfig represents the constructor for struct SocketConfig.
func NewSocketConfig() *SocketConfig {
	sc := &SocketConfig{
		entity:     "",
		login:      "",
		otp:        "",
		pw:         "",
		remoteaddr: "",
		session:    "",
		user:       "",
	}
	return sc
}

// GetPOSTData method to return the struct data ready to submit within
// POST request of type "application/x-www-form-urlencoded"
func (s *SocketConfig) GetPOSTData() string {
	var tmp strings.Builder
	if len(s.entity) > 0 {
		tmp.WriteString(url.QueryEscape("s_entity"))
		tmp.WriteString("=")
		tmp.WriteString(url.QueryEscape(s.entity))
		tmp.WriteString("&")
	}
	if len(s.login) > 0 {
		tmp.WriteString(url.QueryEscape("s_login"))
		tmp.WriteString("=")
		tmp.WriteString(url.QueryEscape(s.login))
		tmp.WriteString("&")
	}
	if len(s.otp) > 0 {
		tmp.WriteString(url.QueryEscape("s_otp"))
		tmp.WriteString("=")
		tmp.WriteString(url.QueryEscape(s.otp))
		tmp.WriteString("&")
	}
	if len(s.pw) > 0 {
		tmp.WriteString(url.QueryEscape("s_pw"))
		tmp.WriteString("=")
		tmp.WriteString(url.QueryEscape(s.pw))
		tmp.WriteString("&")
	}
	if len(s.remoteaddr) > 0 {
		tmp.WriteString(url.QueryEscape("s_remoteaddr"))
		tmp.WriteString("=")
		tmp.WriteString(url.QueryEscape(s.remoteaddr))
		tmp.WriteString("&")
	}
	if len(s.session) > 0 {
		tmp.WriteString(url.QueryEscape("s_session"))
		tmp.WriteString("=")
		tmp.WriteString(url.QueryEscape(s.session))
		tmp.WriteString("&")
	}
	if len(s.user) > 0 {
		tmp.WriteString(url.QueryEscape("s_user"))
		tmp.WriteString("=")
		tmp.WriteString(url.QueryEscape(s.user))
		tmp.WriteString("&")
	}
	return tmp.String()
}

// GetSession method to return the session id currently in use.
func (s *SocketConfig) GetSession() string {
	return s.session
}

// GetSystemEntity method to return the API system entity currently in use.
func (s *SocketConfig) GetSystemEntity() string {
	return s.entity
}

// SetLogin method to set username to use for api communication
func (s *SocketConfig) SetLogin(value string) *SocketConfig {
	s.session = ""
	s.login = value
	return s
}

// SetOTP method to set one time password to use for api communication
func (s *SocketConfig) SetOTP(value string) *SocketConfig {
	s.session = ""
	s.otp = value
	return s
}

// SetPassword method to set password to use for api communication
func (s *SocketConfig) SetPassword(value string) *SocketConfig {
	s.session = ""
	s.pw = value
	return s
}

// SetRemoteAddress method to set remote ip address to be submitted to the HEXONET API.
// This ip address is being considered when you have ip filter settings activated.
// To reset this, simply provide an empty string as parameter.
func (s *SocketConfig) SetRemoteAddress(value string) *SocketConfig {
	s.remoteaddr = value
	return s
}

// SetSession method to set a API session id to use for api communication instead of credentials
// which is basically required in case you plan to use session based communication or if you want to use 2FA
func (s *SocketConfig) SetSession(sessionid string) *SocketConfig {
	s.login = ""
	s.pw = ""
	s.otp = ""
	s.session = sessionid
	return s
}

// SetSystemEntity method to set the system to use e.g. 1234 -> OT&E System, 54cd -> LIVE System
func (s *SocketConfig) SetSystemEntity(value string) *SocketConfig {
	s.entity = value
	return s
}

// SetUser method to set an user account (must be subuser account of your login user) to use for API communication
// use this if you want to make changes on that subuser account or if you want to have his data view
func (s *SocketConfig) SetUser(username string) *SocketConfig {
	s.user = username
	return s
}
