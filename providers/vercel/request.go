package vercel

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	vercelClient "github.com/vercel/terraform-provider-vercel/client"
)

type clientRequest struct {
	ctx              context.Context
	method           string
	url              string
	body             string
	errorOnNoContent bool
}

func (cr *clientRequest) toHTTPRequest() (*http.Request, error) {
	r, err := http.NewRequestWithContext(
		cr.ctx,
		cr.method,
		cr.url,
		strings.NewReader(cr.body),
	)
	if err != nil {
		return nil, err
	}
	// Use a custom user agent for dnscontrol
	r.Header.Set("User-Agent", "dnscontrol https://github.com/StackExchange/dnscontrol/pull/3542")
	if cr.body != "" {
		r.Header.Set("Content-Type", "application/json")
	}
	return r, nil
}

// doRequest is a helper function for consistently requesting data from vercel.
// Adapted from github.com/vercel/terraform-provider-vercel/client/request.go
//
// This manages:
// - Setting the default Content-Type for requests with a body
// - Setting the User-Agent
// - Authorization via the Bearer token
// - Converting error responses into an inspectable type
// - Unmarshaling responses
// - Parsing a Retry-After header in the case of rate limits being hit
// - In the case of a rate-limit being hit, trying again aftera period of time
func (c *vercelProvider) doRequest(req clientRequest, v interface{}) error {
	r, err := req.toHTTPRequest()
	if err != nil {
		return err
	}
	retryAfter, err := c._doRequest(r, v, req.errorOnNoContent)
	for retries := 0; retries < 3; retries++ {
		if retryAfter > 0 && retryAfter < 5*60 { // and the retry time is less than 5 minutes
			fmt.Printf("Rate limit was hit. Retrying after %d seconds. Error: %v\n", retryAfter, err)

			time.Sleep(time.Duration(retryAfter) * time.Second)
			r, err = req.toHTTPRequest()
			if err != nil {
				return err
			}
			retryAfter, err = c._doRequest(r, v, req.errorOnNoContent)
			if err != nil {
				continue
			}
			return nil
		} else {
			break
		}
	}

	return err
}

func (c *vercelProvider) _doRequest(req *http.Request, v interface{}, errorOnNoContent bool) (int, error) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.apiToken))

	// Use a default http client if one isn't available on the provider (though we don't store one yet)
	// For now, we'll just create one or use http.DefaultClient.
	// Better to use a client with timeout.
	httpClient := &http.Client{
		Timeout: 5 * 60 * time.Second,
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error doing http request: %w", err)
	}

	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode >= 300 {
		var errorResponse vercelClient.APIError
		if string(responseBody) == "" {
			errorResponse.StatusCode = resp.StatusCode
			return 0, errorResponse
		}
		err = json.Unmarshal(responseBody, &struct {
			Error *vercelClient.APIError `json:"error"`
		}{
			Error: &errorResponse,
		})
		if err != nil {
			// Try to unmarshal directly if it's not wrapped in "error"
			if err2 := json.Unmarshal(responseBody, &errorResponse); err2 != nil {
				return 0, fmt.Errorf("error unmarshaling response for status code %d: %w", resp.StatusCode, err)
			}
		}
		errorResponse.StatusCode = resp.StatusCode
		errorResponse.RawMessage = responseBody

		var retryAfter int
		// We can't set retryAfter on errorResponse because it's unexported in vercelClient.APIError
		// So we return it separately.
		if resp.StatusCode == 429 {
			retryAfter = 1000 // default
			retryAfterRaw := resp.Header.Get("Retry-After")
			if retryAfterRaw != "" {
				ra, err := strconv.Atoi(retryAfterRaw)
				if err == nil && ra > 0 {
					retryAfter = ra
				}
			}
		}
		return retryAfter, errorResponse
	}

	if v == nil {
		return 0, nil
	}

	if errorOnNoContent && resp.StatusCode == 204 {
		return 0, vercelClient.APIError{
			StatusCode: 204,
			Code:       "no_content",
			Message:    "No content",
		}
	}

	err = json.Unmarshal(responseBody, v)
	if err != nil {
		return 0, fmt.Errorf("error unmarshaling response %s: %w", responseBody, err)
	}

	return 0, nil
}
