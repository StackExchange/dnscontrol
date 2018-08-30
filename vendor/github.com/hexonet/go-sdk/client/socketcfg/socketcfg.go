// Copyright (c) 2018 Kai Schwarz (1API GmbH). All rights reserved.
//
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE.md file.

// Package socketcfg provides apiconnector client connection settings
package socketcfg

import (
	"net/url"
	"strings"
)

// Socketcfg is a struct representing connection settings used as POST data for http request against the insanely fast 1API backend API.
type Socketcfg struct {
	login      string
	pw         string
	remoteaddr string
	entity     string
	session    string
	user       string
	otp        string
}

// SetIPAddress method to set remote ip address to be submitted to the HEXONET API.
// This ip address is being considered when you have ip filter settings activated.
// To reset this, simply provide an empty string as parameter.
func (s *Socketcfg) SetIPAddress(ip string) {
	s.remoteaddr = ip
}

// SetCredentials method to set username and password to use for api communication
func (s *Socketcfg) SetCredentials(username string, password string, otpcode string) {
	s.login = username
	s.pw = password
	s.otp = otpcode
}

// SetEntity method to set the system entity id used to communicate with
// "1234" -> OT&E system, "54cd" -> LIVE system
func (s *Socketcfg) SetEntity(entityid string) {
	s.entity = entityid
}

// SetSession method to set a API session id to use for api communication instead of credentials
// which is basically required in case you plan to use session based communication or if you want to use 2FA
func (s *Socketcfg) SetSession(sessionid string) {
	s.login = ""
	s.pw = ""
	s.otp = ""
	s.session = sessionid
}

// SetUser method to set an user account (must be subuser account of your login user) to use for API communication
// use this if you want to make changes on that subuser account or if you want to have his data view
func (s *Socketcfg) SetUser(username string) {
	s.user = username
}

// EncodeData method to return the struct data ready to submit within POST request of type "application/x-www-form-urlencoded"
func (s *Socketcfg) EncodeData() string {
	var tmp strings.Builder
	if len(s.login) > 0 {
		tmp.WriteString(url.QueryEscape("s_login"))
		tmp.WriteString("=")
		tmp.WriteString(url.QueryEscape(s.login))
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
	if len(s.entity) > 0 {
		tmp.WriteString(url.QueryEscape("s_entity"))
		tmp.WriteString("=")
		tmp.WriteString(url.QueryEscape(s.entity))
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
	if len(s.otp) > 0 {
		tmp.WriteString(url.QueryEscape("s_otp"))
		tmp.WriteString("=")
		tmp.WriteString(url.QueryEscape(s.otp))
		tmp.WriteString("&")
	}
	return tmp.String()
}
