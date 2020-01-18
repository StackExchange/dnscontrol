// Package opensrs provides a client for the OpenSRS API.
// In order to use this package you will need a OpenSRS account.
package opensrs

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	// Version identifies the current library version.
	// This is a pro-forma convention given that Go dependencies
	// tends to be fetched directly from the repo.
	// It is also used in the user-agent identify the client.
	Version = "0.0.1"

	// defaultBaseURL to the OpenSRS production API.
	defaultBaseURL = "https://rr-n1-tor.opensrs.net:55443"
	//defaultBaseURL = "https://horizon.opensrs.net:55443"

	// userAgent represents the default user agent used
	// when no other user agent is set.
	defaultUserAgent = "opensrs-go/" + Version
)

// Client represents a client to the OpenSRS API.
type Client struct {
	// HttpClient is the underlying HTTP client
	// used to communicate with the API.
	HttpClient *http.Client

	// Credentials used for accessing the OpenSRS API
	Credentials Credentials

	// BaseURL for API requests.
	// Defaults to the public OpenSRS API, but can be set to a different endpoint (e.g. the sandbox).
	BaseURL string

	// UserAgent used when communicating with the OpenSRS API.
	UserAgent string

	// Services used for talking to different parts of the OpenSRS API.
	Domains *DomainsService

	// Set to true to output debugging logs during API calls
	Debug bool
}

// NewClient returns a new OpenSRS API client using the given credentials.
func NewClient(credentials Credentials) *Client {
	c := &Client{Credentials: credentials, HttpClient: &http.Client{}, BaseURL: defaultBaseURL}
	c.Domains = &DomainsService{client: c}
	return c
}

// NewRequest creates an API request.
// The path is expected to be a relative path and will be resolved
// according to the BaseURL of the Client. Paths should always be specified without a preceding slash.
func (c *Client) NewRequest(method, path string, payload interface{}) (*http.Request, error) {
	url := c.BaseURL + path

	body := new(bytes.Buffer)
	if payload != nil {
		xml, err := ToXml(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(xml)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "text/xml")
	req.Header.Add("Accept", "text/xml")
	req.Header.Add("User-Agent", formatUserAgent(c.UserAgent))
	for key, value := range c.Credentials.Headers(body.Bytes()) {
		req.Header.Add(key, value)
	}

	return req, nil
}

// formatUserAgent builds the final user agent to use for HTTP requests.
//
// If no custom user agent is provided, the default user agent is used.
//
//     opensrs-go/1.0
//
// If a custom user agent is provided, the final user agent is the combination of the custom user agent
// prepended by the default user agent.
//
//     opensrs-go/1.0 customAgentFlag
//
func formatUserAgent(customUserAgent string) string {
	if customUserAgent == "" {
		return defaultUserAgent
	}

	return fmt.Sprintf("%s %s", defaultUserAgent, customUserAgent)
}

func (c *Client) post(action string, object string, attributes OpsRequestAttributes, obj *OpsResponse) (*http.Response, error) {
	payload := OpsRequest{Action: action, Object: object, Protocol: "XCP", Attributes: attributes}
	req, err := c.NewRequest("POST", "", payload)
	if err != nil {
		return nil, err
	}

	return c.Do(req, obj)
}

// Do sends an API request and returns the API response.
//
// The API response is JSON decoded and stored in the value pointed by obj,
// or returned as an error if an API error has occurred.
// If obj implements the io.Writer interface, the raw response body will be written to obj,
// without attempting to decode it.
func (c *Client) Do(req *http.Request, obj *OpsResponse) (*http.Response, error) {
	if c.Debug {
		log.Printf("Executing request (%v): %#v", req.URL, req)
	}

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if c.Debug {
		log.Printf("Response received: %#v", resp)
	}

	err = CheckResponse(resp)
	if err != nil {
		return resp, err
	}

	// If obj implements the io.Writer,
	// the response body is decoded into v.
	if obj != nil {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		err = FromXml(b, obj)
		if err != nil {
			return resp, err
		}

		err = CheckOpsResponse(resp, obj)
		if err != nil {
			return resp, err
		}
	}

	return resp, err
}

// An ErrorResponse represents an API response that generated an error.
type ErrorResponse struct {
	HttpResponse *http.Response
	OpsResponse  *OpsResponse

	// human-readable message
	Message string `json:"message"`
}

// Error implements the error interface.
func (r *ErrorResponse) Error() string {
	s := fmt.Sprintf("%v %v: ",
		r.HttpResponse.Request.Method, r.HttpResponse.Request.URL)
	if r.OpsResponse != nil {
		s = s + fmt.Sprintf("%v %v",
			r.OpsResponse.ResponseCode,
			r.OpsResponse.ResponseText)
	} else {
		s = s + fmt.Sprintf("%v %v",
			r.HttpResponse.StatusCode, r.Message)
	}
	return s
}

// CheckResponse checks the API response for errors, and returns them if present.
// A response is considered an error if the status code is different than 2xx. Specific requests
// may have additional requirements, but this is sufficient in most of the cases.
func CheckResponse(resp *http.Response) error {
	if code := resp.StatusCode; 200 <= code && code <= 299 {
		return nil
	}

	errorResponse := &ErrorResponse{}
	errorResponse.HttpResponse = resp

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = FromXml(b, errorResponse)
	if err != nil {
		return err
	}

	return errorResponse
}

func CheckOpsResponse(resp *http.Response, or *OpsResponse) error {
	if or.IsSuccess == "1" {
		return nil
	}

	errorResponse := &ErrorResponse{}
	errorResponse.HttpResponse = resp
	errorResponse.OpsResponse = or

	return errorResponse
}
