package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/kolo/xmlrpc"
)

const (
	// Production is the SystemType to provide to New to use the production XML API
	Production SystemType = iota
	// Testing is the SystemType to provide to New to use the test XML API
	Testing
	// LiveDNS is the SystemType to provide to New to use the Live DNS REST API
	// Full documentation of the API is available here: http://doc.livedns.gandi.net/
	LiveDNS
)

// SystemType is the type used to resolve gandi API address
type SystemType int

// Url returns the actual gandi API base URL
func (s SystemType) Url() string {
	if s == Production {
		return "https://rpc.gandi.net/xmlrpc/"
	}
	if s == LiveDNS {
		return "https://dns.api.gandi.net/api/v5/"
	}
	return "https://rpc.ote.gandi.net/xmlrpc/"
}

// Client holds the configuration of a gandi client
type Client struct {
	// Key is the API key to provide to gandi
	Key string
	// Url is the base URL of the gandi API
	Url string
}

// New creates a new gandi client for the given system
func New(apiKey string, system SystemType) *Client {
	return &Client{
		Key: apiKey,
		Url: system.Url(),
	}
}

// Call performs an acual XML RPC call to the gandi API
func (c *Client) Call(serviceMethod string, args []interface{}, reply interface{}) error {
	rpc, err := xmlrpc.NewClient(c.Url, nil)
	if err != nil {
		return err
	}
	return rpc.Call(serviceMethod, args, reply)
}

// DoRest performs a request to gandi LiveDNS api and optionnally decodes the reply
func (c *Client) DoRest(req *http.Request, decoded interface{}) (*http.Response, error) {
	if decoded != nil {
		req.Header.Set("Accept", "application/json")
	}
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	//
	defer func() { err = resp.Body.Close() }()
	if decoded != nil {
		b, e := ioutil.ReadAll(resp.Body)
		if e != nil {
			return nil, e
		}
		if len(b) > 0 {
			e = json.Unmarshal(b, decoded)
			if e != nil {
				return nil, e
			}
		}
		resp.Body = ioutil.NopCloser(bytes.NewReader(b))
	}
	return resp, err
}

// NewJSONRequest creates a new authenticated to gandi live DNS REST API.
// If data is not null, it will be encoded as json and prodived in the request body
func (c *Client) NewJSONRequest(method string, url string, data interface{}) (*http.Request, error) {
	var reader io.Reader
	if data != nil {
		b, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, fmt.Sprintf("%s/%s", strings.TrimRight(c.Url, "/"), strings.TrimLeft(url, "/")), reader)
	if err != nil {
		return nil, err
	}
	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("X-Api-Key", c.Key)
	return req, nil
}

// Get performs a Get request to gandi Live DNS api and decodes the returned data if a not null decoded pointer is provided
func (c *Client) Get(URI string, decoded interface{}) (*http.Response, error) {
	req, err := c.NewJSONRequest("GET", URI, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.DoRest(req, decoded)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected http code %d on URL %v. expecting %d", resp.StatusCode, resp.Request.URL, http.StatusOK)
	}
	return resp, err
}

// Delete performs a Delete request to gandi Live DNS api and decodes the returned data if a not null decoded pointer is provided
func (c *Client) Delete(URI string, decoded interface{}) (*http.Response, error) {
	req, err := c.NewJSONRequest("DELETE", URI, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.DoRest(req, decoded)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("Unexpected http code %d on URL %v. expecting %d", resp.StatusCode, resp.Request.URL, http.StatusNoContent)
	}
	return resp, err
}

// Post performs a Post request request to gandi Live DNS api
// - with data encoded as JSON if a not null data pointer is provided
// - decodes the returned data if a not null decoded pointer is provided
// - ensures the status code is an HTTP accepted
func (c *Client) Post(URI string, data interface{}, decoded interface{}) (*http.Response, error) {
	req, err := c.NewJSONRequest("POST", URI, data)
	if err != nil {
		return nil, err
	}
	resp, err := c.DoRest(req, decoded)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("Unexpected http code %d on URL %v. expecting %d", resp.StatusCode, resp.Request.URL, http.StatusCreated)
	}
	return resp, err
}

// Put performs a Put request to gandi Live DNS api
// - with data encoded as JSON if a not null data pointer is provided
// - decodes the returned data if a not null decoded pointer is provided
func (c *Client) Put(URI string, data interface{}, decoded interface{}) (*http.Response, error) {
	req, err := c.NewJSONRequest("PUT", URI, data)
	if err != nil {
		return nil, err
	}
	return c.DoRest(req, decoded)
}

// Patch performs a Patch request to gandi Live DNS api
// - with data encoded as JSON if a not null data pointer is provided
// - decodes the returned data if a not null decoded pointer is provided
// - ensures the status code is an HTTP accepted
func (c *Client) Patch(URI string, data interface{}, decoded interface{}) (*http.Response, error) {
	req, err := c.NewJSONRequest("PATCH", URI, data)
	if err != nil {
		return nil, err
	}
	resp, err := c.DoRest(req, decoded)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("Unexpected http code %d on URL %v. expecting %d", resp.StatusCode, resp.Request.URL, http.StatusAccepted)
	}
	return resp, err
}
