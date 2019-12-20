package gandi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strings"
)

const (
	gandiEndpoint = "https://dns.api.gandi.net/api/v5/"
	// HTTP Methods
	mPATCH  = http.MethodPatch
	mGET    = http.MethodGet
	mPOST   = http.MethodPost
	mDELETE = http.MethodDelete
	mPUT    = http.MethodPut
)

// Gandi makes it easier to interact with Gandi LiveDNS
type Gandi struct {
	apikey     string
	sharing_id string
	debug      bool
}

// New instantiates a new Gandi instance
func New(apikey string, sharing_id string) *Gandi {
	return &Gandi{apikey: apikey, sharing_id: sharing_id}
}

func (g *Gandi) askGandi(method, path string, params, recipient interface{}) (http.Header, error) {
	marshalledParams, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	resp, err := g.doAskGandi(method, path, marshalledParams, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	decoder.Decode(recipient)
	return resp.Header, nil
}

func (g *Gandi) askGandiToBytes(method, path string, params interface{}) (http.Header, []byte, error) {
	headers := [][2]string{
		[2]string{"Accept", "text/plain"},
	}
	marshalledParams, err := json.Marshal(params)
	if err != nil {
		return nil, nil, err
	}
	resp, err := g.doAskGandi(method, path, marshalledParams, headers)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	return resp.Header, content, err
}

func (g *Gandi) askGandiFromBytes(method, path string, params []byte, recipient interface{}) (http.Header, error) {
	resp, err := g.doAskGandi(method, path, params, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	decoder.Decode(recipient)
	return resp.Header, nil
}

func (g *Gandi) doAskGandi(method, path string, params []byte, extraHeaders [][2]string) (*http.Response, error) {
	var (
		err error
		req *http.Request
	)
	client := &http.Client{}
	suffix := ""
	if len(g.sharing_id) != 0 {
		suffix += "?sharing_id=" + g.sharing_id
	}
	if params != nil && string(params) != "null" {
		req, err = http.NewRequest(method, gandiEndpoint+path+suffix, bytes.NewReader(params))
	} else {
		req, err = http.NewRequest(method, gandiEndpoint+path+suffix, nil)
	}
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-Api-Key", g.apikey)
	req.Header.Add("Content-Type", "application/json")
	for _, header := range extraHeaders {
		req.Header.Add(header[0], header[1])
	}
	if g.debug {
		dump, _ := httputil.DumpRequestOut(req, true)
		fmt.Println("=======================================\nREQUEST:")
		fmt.Println(string(dump))
	}
	resp, err := client.Do(req)
	if err != nil {
		return resp, err
	}
	if g.debug {
		dump, _ := httputil.DumpResponse(resp, true)
		fmt.Println("=======================================\nREQUEST:")
		fmt.Println(string(dump))
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		defer resp.Body.Close()
		var message StandardResponse
		defer resp.Body.Close()
		decoder := json.NewDecoder(resp.Body)
		decoder.Decode(&message)
		if message.Message != "" {
			err = fmt.Errorf("%d: %s", resp.StatusCode, message.Message)
		} else if len(message.Errors) > 0 {
			var errors []string
			for _, oneError := range message.Errors {
				errors = append(errors, fmt.Sprintf("%s: %s", oneError.Name, oneError.Description))
			}
			err = fmt.Errorf(strings.Join(errors, ", "))
		} else {
			err = fmt.Errorf("%d", resp.StatusCode)

		}
	}
	return resp, err
}

// StandardResponse is a standard response
type StandardResponse struct {
	Code    int             `json:"code,omitempty"`
	Message string          `json:"message,omitempty"`
	UUID    string          `json:"uuid,omitempty"`
	Object  string          `json:"object,omitempty"`
	Cause   string          `json:"cause,omitempty"`
	Status  string          `json:"status,omitempty"`
	Errors  []StandardError `json:"errors,omitempty"`
}

// StandardError is embedded in a standard error
type StandardError struct {
	Location    string `json:"location"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
