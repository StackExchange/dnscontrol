// Copyright (c) 2018 Kai Schwarz (1API GmbH). All rights reserved.
//
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE.md file.

// Package client contains all you need to communicate with the insanely fast 1API backend API.
package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/hexonet/go-sdk/client/socketcfg"
	"github.com/hexonet/go-sdk/response/hashresponse"
	"github.com/hexonet/go-sdk/response/listresponse"
)

// Client is the entry point class for communicating with the insanely fast 1API backend api.
// It allows two ways of communication:
// * session based communication
// * sessionless communication
//
// A session based communication makes sense in case you use it to
// build your own frontend on top. It allows also to use 2FA
// (2 Factor Auth) by providing "otp" in the config parameter of
// the login method.
// A sessionless communication makes sense in case you do not need
// to care about the above and you have just to request some commands.
//
// Possible commands can be found at https://github.com/hexonet/hexonet-api-documentation/tree/master/API
type Client struct {
	debugMode     bool
	socketTimeout int
	apiurl        string
	socketcfg.Socketcfg
}

// NewClient represents the constructor for struct Client.
// The client is by default set to communicate with the LIVE system. Use method UseOTESystem to switch to the OT&E system instance.
func NewClient() *Client {
	cl := &Client{
		debugMode:     false,
		socketTimeout: 300000,
		apiurl:        "https://coreapi.1api.net/api/call.cgi",
		Socketcfg:     socketcfg.Socketcfg{},
	}
	cl.UseLiveSystem()
	return cl
}

// EncodeData method to use to encode provided data (socket configuration and api command) before sending it to the API server
// It returns the encoded data ready to use within POST request of type "application/x-www-form-urlencoded"
func (c *Client) EncodeData(cfg *socketcfg.Socketcfg, cmd map[string]string) string {
	var tmp, data strings.Builder
	tmp.WriteString(cfg.EncodeData())
	tmp.WriteString(url.QueryEscape("s_command"))
	tmp.WriteString("=")

	for k, v := range cmd {
		re := regexp.MustCompile(`\r?\n`)
		v = re.ReplaceAllString(v, "")
		if len(v) > 0 {
			data.WriteString(k)
			data.WriteString("=")
			data.WriteString(v)
			data.WriteString("\n")
		}
	}
	tmp.WriteString(url.QueryEscape(data.String()))
	return tmp.String()
}

// Getapiurl is the getter method for apiurl property
func (c *Client) Getapiurl() string {
	return c.apiurl
}

// Setapiurl is the setter method for apiurl
func (c *Client) Setapiurl(url string) {
	c.apiurl = url
}

// SetCredentials method to set username and password and otp code to use for api communication
// set otp code to empty string, if you do not use 2FA
func (c *Client) SetCredentials(username string, password string, otpcode string) {
	c.Socketcfg.SetCredentials(username, password, otpcode)
}

// SetIPAddress method to set api client to submit this ip address in api communication
func (c *Client) SetIPAddress(ip string) {
	c.Socketcfg.SetIPAddress(ip)
}

// SetSubuserView method to activate the use of a subuser account as data view
func (c *Client) SetSubuserView(username string) {
	c.Socketcfg.SetUser(username)
}

// ResetSubuserView method to deactivate the use of a subuser account as data view
func (c *Client) ResetSubuserView() {
	c.Socketcfg.SetUser("")
}

// UseLiveSystem method to set api client to communicate with the LIVE backend API
func (c *Client) UseLiveSystem() {
	c.Socketcfg.SetEntity("54cd")
}

// UseOTESystem method to set api client to communicate with the OT&E backend API
func (c *Client) UseOTESystem() {
	c.Socketcfg.SetEntity("1234")
}

// EnableDebugMode method to enable debugMode for debug output
func (c *Client) EnableDebugMode() {
	c.debugMode = true
}

// DisableDebugMode method to disable debugMode for debug output
func (c *Client) DisableDebugMode() {
	c.debugMode = false
}

// Request method requests the given command to the api server and returns the response as ListResponse.
func (c *Client) Request(cmd map[string]string) *listresponse.ListResponse {
	if c.Socketcfg == (socketcfg.Socketcfg{}) {
		return listresponse.NewListResponse(hashresponse.NewTemplates().Get("expired"))
	}
	return c.dorequest(cmd, &c.Socketcfg)
}

// debugRequest method used to trigger debug output in case debugMode is activated
func (c *Client) debugRequest(cmd map[string]string, data string, r *listresponse.ListResponse) {
	if c.debugMode {
		j, _ := json.Marshal(cmd)
		fmt.Printf("%s\n", j)
		fmt.Println("POST: " + data)
		fmt.Println(strconv.Itoa(r.Code()) + " " + r.Description() + "\n")
	}
}

// RequestAll method requests ALL entries matching the request criteria by the given command from api server.
// So useful for client-side lists. Finally it returns the response as ListResponse.
func (c *Client) RequestAll(cmd map[string]string) *listresponse.ListResponse {
	if c.Socketcfg == (socketcfg.Socketcfg{}) {
		return listresponse.NewListResponse(hashresponse.NewTemplates().Get("expired"))
	}
	cmd["LIMIT"] = "1"
	cmd["FIRST"] = "0"
	r := c.dorequest(cmd, &c.Socketcfg)
	if r.IsSuccess() {
		cmd["LIMIT"] = strconv.Itoa(r.Total())
		cmd["FIRST"] = "0"
		r = c.dorequest(cmd, &c.Socketcfg)
	}
	return r
}

// request the given command to the api server by using the provided socket configuration and return the response as ListResponse.
func (c *Client) dorequest(cmd map[string]string, cfg *socketcfg.Socketcfg) *listresponse.ListResponse {
	data := c.EncodeData(cfg, cmd)
	client := &http.Client{}
	req, err := http.NewRequest("POST", c.apiurl, strings.NewReader(data))
	if err != nil {
		tpl := hashresponse.NewTemplates().Get("commonerror")
		tpl = strings.Replace(tpl, "####ERRMSG####", err.Error(), 1)
		r := listresponse.NewListResponse(tpl)
		c.debugRequest(cmd, data, r)
		return r
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Expect", "")
	resp, err2 := client.Do(req)
	if err2 != nil {
		tpl := hashresponse.NewTemplates().Get("commonerror")
		tpl = strings.Replace(tpl, "####ERRMSG####", err2.Error(), 1)
		r := listresponse.NewListResponse(tpl)
		c.debugRequest(cmd, data, r)
		return r
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		response, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			tpl := hashresponse.NewTemplates().Get("commonerror")
			tpl = strings.Replace(tpl, "####ERRMSG####", err.Error(), 1)
			r := listresponse.NewListResponse(tpl)
			c.debugRequest(cmd, data, r)
			return r
		}
		r := listresponse.NewListResponse(string(response))
		c.debugRequest(cmd, data, r)
		return r
	}
	tpl := hashresponse.NewTemplates().Get("commonerror")
	tpl = strings.Replace(tpl, "####ERRMSG####", string(resp.StatusCode)+resp.Status, 1)
	r := listresponse.NewListResponse(tpl)
	c.debugRequest(cmd, data, r)
	return r
}

// Login method to use as entry point for session based communication.
// Response is returned as ListResponse.
func (c *Client) Login() *listresponse.ListResponse {
	return c.dologin(map[string]string{"COMMAND": "StartSession"})
}

// LoginExtended method to use as entry point for session based communication.
// This method allows to provide further command parameters for startsession command.
// Response is returned as ListResponse.
func (c *Client) LoginExtended(cmdparams map[string]string) *listresponse.ListResponse {
	cmd := map[string]string{"COMMAND": "StartSession"}
	for k, v := range cmdparams {
		cmd[k] = v
	}
	return c.dologin(cmd)
}

// dologin method used internally to perform a login using the given command.
// Response is returned as ListResponse.
func (c *Client) dologin(cmd map[string]string) *listresponse.ListResponse {
	r := c.dorequest(cmd, &c.Socketcfg)
	if r.Code() == 200 {
		sessid, _ := r.GetColumnIndex("SESSION", 0)
		c.Socketcfg.SetSession(sessid)
	}
	return r
}

// Logout method to use for session based communication.
// This method logs you out and destroys the api session.
// Response is returned as ListResponse.
func (c *Client) Logout() *listresponse.ListResponse {
	cmd := map[string]string{"COMMAND": "EndSession"}
	return c.dorequest(cmd, &c.Socketcfg)
}
