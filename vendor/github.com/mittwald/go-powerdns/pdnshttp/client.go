package pdnshttp

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"
)

type Client struct {
	baseURL       string
	httpClient    *http.Client
	authenticator ClientAuthenticator
	debugOutput   io.Writer
}

// NewClient returns a new PowerDNS HTTP client
func NewClient(baseURL string, hc *http.Client, auth ClientAuthenticator, debugOutput io.Writer) *Client {
	c := Client{
		baseURL:       baseURL,
		httpClient:    hc,
		authenticator: auth,
		debugOutput:   debugOutput,
	}

	return &c
}

// NewRequest builds a new request. Usually, this method should not be used;
// prefer using the "Get", "Post", ... methods if possible.
func (c *Client) NewRequest(method string, path string, body io.Reader) (*http.Request, error) {
	path = strings.TrimPrefix(path, "/")
	req, err := http.NewRequest(method, c.baseURL+"/"+path, body)
	if err != nil {
		return nil, err
	}

	if c.authenticator != nil {
		if err := c.authenticator.OnRequest(req); err != nil {
			return nil, err
		}
	}

	return req, err
}

// Get executes a GET request
func (c *Client) Get(ctx context.Context, path string, out interface{}, opts ...RequestOption) error {
	return c.doRequest(ctx, http.MethodGet, path, out, opts...)
}

// Post executes a POST request
func (c *Client) Post(ctx context.Context, path string, out interface{}, opts ...RequestOption) error {
	return c.doRequest(ctx, http.MethodPost, path, out, opts...)
}

// Put executes a PUT request
func (c *Client) Put(ctx context.Context, path string, out interface{}, opts ...RequestOption) error {
	return c.doRequest(ctx, http.MethodPut, path, out, opts...)
}

// Patch executes a PATCH request
func (c *Client) Patch(ctx context.Context, path string, out interface{}, opts ...RequestOption) error {
	return c.doRequest(ctx, http.MethodPatch, path, out, opts...)
}

// Delete executes a DELETE request
func (c *Client) Delete(ctx context.Context, path string, out interface{}, opts ...RequestOption) error {
	return c.doRequest(ctx, http.MethodDelete, path, out, opts...)
}

func (c *Client) doRequest(ctx context.Context, method string, path string, out interface{}, opts ...RequestOption) error {
	req, err := c.NewRequest(method, path, nil)
	if err != nil {
		return err
	}

	for i := range opts {
		if err := opts[i](req); err != nil {
			return err
		}
	}

	req = req.WithContext(ctx)

	reqDump, _ := httputil.DumpRequestOut(req, true)
	c.debugOutput.Write(reqDump)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	resDump, _ := httputil.DumpResponse(res, true)
	c.debugOutput.Write(resDump)

	if res.StatusCode == http.StatusNotFound {
		return ErrNotFound{URL: req.URL.String()}
	} else if res.StatusCode >= 400 {
		return ErrUnexpectedStatus{URL: req.URL.String(), StatusCode: res.StatusCode}
	}

	if out != nil {
		if w, ok := out.(io.Writer); ok {
			_, err := io.Copy(w, res.Body)
			return err
		}

		dec := json.NewDecoder(res.Body)
		err = dec.Decode(out)
		if err != nil {
			return err
		}
	}

	return nil
}
