package joker

import (
	"fmt"
	"io"
	"net/url"
	"strings"
)

// authenticate logs in to Joker DMAPI and stores the session ID.
func (api *jokerProvider) authenticate() error {
	data := url.Values{}

	if api.apiKey != "" {
		data.Set("api-key", api.apiKey)
	} else {
		data.Set("username", api.username)
		data.Set("password", api.password)
	}

	resp, err := api.httpClient.PostForm(api.apiURL+"login", data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Parse the response headers and body
	respStr := string(body)
	headers, _ := api.parseResponse(respStr)

	if headers["Status-Code"] != "" && headers["Status-Code"] != "0" {
		return fmt.Errorf("login failed: %s", headers["Status-Text"])
	}

	authSID := headers["Auth-Sid"]
	if authSID == "" {
		return fmt.Errorf("no Auth-Sid received from login. Response: %s", respStr)
	}

	api.authSID = authSID
	return nil
}

// parseResponse parses the Joker DMAPI response format.
func (api *jokerProvider) parseResponse(response string) (map[string]string, string) {
	headers := make(map[string]string)
	lines := strings.Split(response, "\n")

	var bodyStart int
	for i, line := range lines {
		if line == "" {
			bodyStart = i + 1
			break
		}
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	body := ""
	if bodyStart < len(lines) {
		body = strings.Join(lines[bodyStart:], "\n")
	}

	return headers, body
}

// makeRequest makes an authenticated request to Joker DMAPI.
func (api *jokerProvider) makeRequest(endpoint string, params url.Values) (map[string]string, string, error) {
	if params == nil {
		params = url.Values{}
	}
	params.Set("auth-sid", api.authSID)

	resp, err := api.httpClient.PostForm(api.apiURL+endpoint, params)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	headers, responseBody := api.parseResponse(string(body))

	if headers["Status-Code"] != "" && headers["Status-Code"] != "0" {
		statusText := headers["Status-Text"]
		// Check for session expiration and attempt to renew
		if strings.Contains(statusText, "Auth-Sid") || strings.Contains(statusText, "session") ||
			strings.Contains(statusText, "authorization") || strings.Contains(statusText, "authentication") {
			// Try to re-authenticate
			if authErr := api.authenticate(); authErr == nil {
				// Retry the request with new session
				params.Set("auth-sid", api.authSID)
				resp2, err2 := api.httpClient.PostForm(api.apiURL+endpoint, params)
				if err2 != nil {
					return nil, "", err2
				}
				defer resp2.Body.Close()

				body2, err2 := io.ReadAll(resp2.Body)
				if err2 != nil {
					return nil, "", err2
				}

				headers2, responseBody2 := api.parseResponse(string(body2))
				if headers2["Status-Code"] != "" && headers2["Status-Code"] != "0" {
					return nil, "", fmt.Errorf("API error after re-auth: %s (Status-Code: %s)", headers2["Status-Text"], headers2["Status-Code"])
				}
				return headers2, responseBody2, nil
			}
		}
		return nil, "", fmt.Errorf("API error: %s (Status-Code: %s)", statusText, headers["Status-Code"])
	}

	return headers, responseBody, nil
}
