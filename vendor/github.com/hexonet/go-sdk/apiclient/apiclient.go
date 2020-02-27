// Copyright (c) 2018 Kai Schwarz (HEXONET GmbH). All rights reserved.
//
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE.md file.

// Package apiclient contains all you need to communicate with the insanely fast HEXONET backend API.
package apiclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"runtime"
	"sort"
	"strings"
	"time"

	R "github.com/hexonet/go-sdk/response"
	RTM "github.com/hexonet/go-sdk/responsetemplatemanager"
	SC "github.com/hexonet/go-sdk/socketconfig"
)

var rtm = RTM.GetInstance()

// APIClient is the entry point class for communicating with the insanely fast HEXONET backend api.
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
type APIClient struct {
	socketTimeout time.Duration
	socketURL     string
	socketConfig  *SC.SocketConfig
	debugMode     bool
	ua            string
}

// NewAPIClient represents the constructor for struct APIClient.
func NewAPIClient() *APIClient {
	cl := &APIClient{
		debugMode:     false,
		socketTimeout: 300 * time.Second,
		socketURL:     "https://api.ispapi.net/api/call.cgi",
		socketConfig:  SC.NewSocketConfig(),
		ua:            "",
	}
	cl.UseLIVESystem()
	return cl
}

// EnableDebugMode method to enable Debug Output to logger
func (cl *APIClient) EnableDebugMode() *APIClient {
	cl.debugMode = true
	return cl
}

// DisableDebugMode method to disable Debug Output to logger
func (cl *APIClient) DisableDebugMode() *APIClient {
	cl.debugMode = false
	return cl
}

// GetPOSTData method to Serialize given command for POST request
// including connection configuration data
func (cl *APIClient) GetPOSTData(cmd map[string]string) string {
	data := cl.socketConfig.GetPOSTData()
	var tmp strings.Builder
	keys := []string{}
	for key := range cmd {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		val := cmd[key]
		tmp.WriteString(key)
		tmp.WriteString("=")
		val = strings.Replace(val, "\r", "", -1)
		val = strings.Replace(val, "\n", "", -1)
		tmp.WriteString(val)
		tmp.WriteString("\n")
	}
	str := tmp.String()
	str = str[:len(str)-1] //remove \n at end
	return strings.Join([]string{
		data,
		url.QueryEscape("s_command"),
		"=",
		url.QueryEscape(str),
	}, "")
}

// GetSession method to get the API Session that is currently set
func (cl *APIClient) GetSession() (string, error) {
	sessid := cl.socketConfig.GetSession()
	if len(sessid) == 0 {
		return "", errors.New("Could not find an active session")
	}
	return sessid, nil
}

// GetURL method to get the API connection url that is currently set
func (cl *APIClient) GetURL() string {
	return cl.socketURL
}

// SetUserAgent method to customize user-agent header (useful for tools that use our SDK)
func (cl *APIClient) SetUserAgent(str string, rv string) *APIClient {
	cl.ua = str + " (" + runtime.GOOS + "; " + runtime.GOARCH + "; rv:" + rv + ") go-sdk/" + cl.GetVersion() + " go/" + runtime.Version()
	return cl
}

// GetUserAgent method to return the user agent string
func (cl *APIClient) GetUserAgent() string {
	if len(cl.ua) == 0 {
		cl.ua = "GO-SDK (" + runtime.GOOS + "; " + runtime.GOARCH + "; rv:" + cl.GetVersion() + ") go/" + runtime.Version()
	}
	return cl.ua
}

// GetVersion method to get current module version
func (cl *APIClient) GetVersion() string {
	return "2.2.3"
}

// SaveSession method to apply data to a session for later reuse
// Please save/update that map into user session
func (cl *APIClient) SaveSession(sessionobj map[string]interface{}) *APIClient {
	sessionobj["socketcfg"] = map[string]string{
		"entity":  cl.socketConfig.GetSystemEntity(),
		"session": cl.socketConfig.GetSession(),
	}
	return cl
}

// ReuseSession method to reuse given configuration out of a user session
// to rebuild and reuse connection settings
func (cl *APIClient) ReuseSession(sessionobj map[string]interface{}) *APIClient {
	cfg := sessionobj["socketcfg"].(map[string]string)
	cl.socketConfig.SetSystemEntity(cfg["entity"])
	cl.SetSession(cfg["session"])
	return cl
}

// SetURL method to set another connection url to be used for API communication
func (cl *APIClient) SetURL(value string) *APIClient {
	cl.socketURL = value
	return cl
}

// SetOTP method to set one time password to be used for API communication
func (cl *APIClient) SetOTP(value string) *APIClient {
	cl.socketConfig.SetOTP(value)
	return cl
}

// SetSession method to set an API session id to be used for API communication
func (cl *APIClient) SetSession(value string) *APIClient {
	cl.socketConfig.SetSession(value)
	return cl
}

// SetRemoteIPAddress method to set an Remote IP Address to be used for API communication
func (cl *APIClient) SetRemoteIPAddress(value string) *APIClient {
	cl.socketConfig.SetRemoteAddress(value)
	return cl
}

// SetCredentials method to set Credentials to be used for API communication
func (cl *APIClient) SetCredentials(uid string, pw string) *APIClient {
	cl.socketConfig.SetLogin(uid)
	cl.socketConfig.SetPassword(pw)
	return cl
}

// SetRoleCredentials method to set Role User Credentials to be used for API communication
func (cl *APIClient) SetRoleCredentials(uid string, role string, pw string) *APIClient {
	if len(role) > 0 {
		return cl.SetCredentials(uid+"!"+role, pw)
	}
	return cl.SetCredentials(uid, pw)
}

// Login method to perform API login to start session-based communication
// 1st parameter: one time password
func (cl *APIClient) Login(params ...string) *R.Response {
	otp := ""
	if len(params) > 0 {
		otp = params[0]
	}
	cl.SetOTP(otp)
	rr := cl.Request(map[string]string{"COMMAND": "StartSession"})
	if rr.IsSuccess() {
		col := rr.GetColumn("SESSION")
		if col != nil {
			cl.SetSession(col.GetData()[0])
		} else {
			cl.SetSession("")
		}
	}
	return rr
}

// LoginExtended method to perform API login to start session-based communication.
// 1st parameter: map of additional command parameters
// 2nd parameter: one time password
func (cl *APIClient) LoginExtended(params ...interface{}) *R.Response {
	otp := ""
	parameters := map[string]string{}
	if len(params) == 2 {
		otp = params[1].(string)
	}
	cl.SetOTP(otp)
	if len(params) > 0 {
		parameters = params[0].(map[string]string)
	}
	cmd := map[string]string{
		"COMMAND": "StartSession",
	}
	for k, v := range parameters {
		cmd[k] = v
	}
	rr := cl.Request(cmd)
	if rr.IsSuccess() {
		col := rr.GetColumn("SESSION")
		if col != nil {
			cl.SetSession(col.GetData()[0])
		} else {
			cl.SetSession("")
		}
	}
	return rr
}

// Logout method to perform API logout to close API session in use
func (cl *APIClient) Logout() *R.Response {
	rr := cl.Request(map[string]string{
		"COMMAND": "EndSession",
	})
	if rr.IsSuccess() {
		cl.SetSession("")
	}
	return rr
}

// Request method to perform API request using the given command
func (cl *APIClient) Request(cmd map[string]string) *R.Response {
	data := cl.GetPOSTData(cmd)

	client := &http.Client{
		Timeout: cl.socketTimeout,
	}
	req, err := http.NewRequest("POST", cl.socketURL, strings.NewReader(data))
	if err != nil {
		tpl := rtm.GetTemplate("httperror").GetPlain()
		r := R.NewResponse(tpl, cmd)
		if cl.debugMode {
			j, _ := json.Marshal(cmd)
			fmt.Printf("%s\n", j)
			fmt.Println("POST: " + data)
			fmt.Println("HTTP communication failed: " + err.Error())
			fmt.Println(r.GetPlain())
		}
		return r
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Expect", "")
	req.Header.Add("User-Agent", cl.GetUserAgent())
	resp, err2 := client.Do(req)
	if err2 != nil {
		tpl := rtm.GetTemplate("httperror").GetPlain()
		r := R.NewResponse(tpl, cmd)
		if cl.debugMode {
			j, _ := json.Marshal(cmd)
			fmt.Printf("%s\n", j)
			fmt.Println("POST: " + data)
			fmt.Println("HTTP communication failed: " + err2.Error())
			fmt.Println(r.GetPlain())
		}
		return r
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		response, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			tpl := rtm.GetTemplate("httperror").GetPlain()
			r := R.NewResponse(tpl, cmd)
			if cl.debugMode {
				j, _ := json.Marshal(cmd)
				fmt.Printf("%s\n", j)
				fmt.Println("POST: " + data)
				fmt.Println("HTTP communication failed: " + err.Error())
				fmt.Println(r.GetPlain())
			}
			return r
		}
		r := R.NewResponse(string(response), cmd)
		if cl.debugMode {
			j, _ := json.Marshal(cmd)
			fmt.Printf("%s\n", j)
			fmt.Println("POST: " + data)
			fmt.Println(r.GetPlain())
		}
		return r
	}
	tpl := rtm.GetTemplate("httperror").GetPlain()
	r := R.NewResponse(tpl, cmd)
	if cl.debugMode {
		j, _ := json.Marshal(cmd)
		fmt.Printf("%s\n", j)
		fmt.Println("POST: " + data)
		fmt.Println(r.GetPlain())
	}
	return r
}

// RequestNextResponsePage method to request the next page of list entries for the current list query
// Useful for lists
func (cl *APIClient) RequestNextResponsePage(rr *R.Response) (*R.Response, error) {
	mycmd := cl.toUpperCaseKeys(rr.GetCommand())
	if _, ok := mycmd["LAST"]; ok {
		return nil, errors.New("Parameter LAST in use. Please remove it to avoid issues in requestNextPage")
	}
	first := 0
	if v, ok := mycmd["FIRST"]; ok {
		first, _ = fmt.Sscan("%s", v)
	}
	total := rr.GetRecordsTotalCount()
	limit := rr.GetRecordsLimitation()
	first += limit
	if first < total {
		mycmd["FIRST"] = fmt.Sprintf("%d", first)
		mycmd["LIMIT"] = fmt.Sprintf("%d", limit)
		return cl.Request(mycmd), nil
	}
	return nil, errors.New("Could not find further existing pages")
}

// RequestAllResponsePages method to request all pages/entries for the given query command
// Use this method with caution as it requests all list data until done.
func (cl *APIClient) RequestAllResponsePages(cmd map[string]string) []R.Response {
	var err error
	responses := []R.Response{}
	mycmd := map[string]string{
		"FIRST": "0",
	}
	for k, v := range cmd {
		mycmd[k] = v
	}
	rr := cl.Request(mycmd)
	tmp := rr
	for {
		responses = append(responses, *tmp)
		tmp, err = cl.RequestNextResponsePage(tmp)
		if err != nil {
			break
		}
	}
	return responses
}

// SetUserView method to set a data view to a given subuser
func (cl *APIClient) SetUserView(uid string) *APIClient {
	cl.socketConfig.SetUser(uid)
	return cl
}

// ResetUserView method to reset data view back from subuser to user
func (cl *APIClient) ResetUserView() *APIClient {
	cl.socketConfig.SetUser("")
	return cl
}

// UseOTESystem method to set OT&E System for API communication
func (cl *APIClient) UseOTESystem() *APIClient {
	cl.socketConfig.SetSystemEntity("1234")
	return cl
}

// UseLIVESystem method to set LIVE System for API communication
// Usage of LIVE System is active by default.
func (cl *APIClient) UseLIVESystem() *APIClient {
	cl.socketConfig.SetSystemEntity("54cd")
	return cl
}

// toUpperCaseKeys method to translate all command parameter names to uppercase
func (cl *APIClient) toUpperCaseKeys(cmd map[string]string) map[string]string {
	newcmd := map[string]string{}
	for k, v := range cmd {
		newcmd[strings.ToUpper(k)] = v
	}
	return newcmd
}
