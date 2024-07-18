package cloudflare

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/goccy/go-json"
)

// CustomHostnameStatus is the enumeration of valid state values in the CustomHostnameSSL.
type CustomHostnameStatus string

const (
	// PENDING status represents state of CustomHostname is pending.
	PENDING CustomHostnameStatus = "pending"
	// ACTIVE status represents state of CustomHostname is active.
	ACTIVE CustomHostnameStatus = "active"
	// MOVED status represents state of CustomHostname is moved.
	MOVED CustomHostnameStatus = "moved"
	// DELETED status represents state of CustomHostname is deleted.
	DELETED CustomHostnameStatus = "deleted"
	// BLOCKED status represents state of CustomHostname is blocked from going active.
	BLOCKED CustomHostnameStatus = "blocked"
)

// CustomHostnameSSLSettings represents the SSL settings for a custom hostname.
type CustomHostnameSSLSettings struct {
	HTTP2         string   `json:"http2,omitempty"`
	HTTP3         string   `json:"http3,omitempty"`
	TLS13         string   `json:"tls_1_3,omitempty"`
	MinTLSVersion string   `json:"min_tls_version,omitempty"`
	Ciphers       []string `json:"ciphers,omitempty"`
	EarlyHints    string   `json:"early_hints,omitempty"`
}

// CustomHostnameOwnershipVerification represents ownership verification status of a given custom hostname.
type CustomHostnameOwnershipVerification struct {
	Type  string `json:"type,omitempty"`
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

// SSLValidationError represents errors that occurred during SSL validation.
type SSLValidationError struct {
	Message string `json:"message,omitempty"`
}

// CustomHostnameSSLCertificates represent certificate properties like issuer, expires date and etc.
type CustomHostnameSSLCertificates struct {
	Issuer            string     `json:"issuer"`
	SerialNumber      string     `json:"serial_number"`
	Signature         string     `json:"signature"`
	ExpiresOn         *time.Time `json:"expires_on"`
	IssuedOn          *time.Time `json:"issued_on"`
	FingerprintSha256 string     `json:"fingerprint_sha256"`
	ID                string     `json:"id"`
}

// CustomHostnameSSL represents the SSL section in a given custom hostname.
type CustomHostnameSSL struct {
	ID                   string                          `json:"id,omitempty"`
	Status               string                          `json:"status,omitempty"`
	Method               string                          `json:"method,omitempty"`
	Type                 string                          `json:"type,omitempty"`
	Wildcard             *bool                           `json:"wildcard,omitempty"`
	CustomCertificate    string                          `json:"custom_certificate,omitempty"`
	CustomKey            string                          `json:"custom_key,omitempty"`
	CertificateAuthority string                          `json:"certificate_authority,omitempty"`
	Issuer               string                          `json:"issuer,omitempty"`
	SerialNumber         string                          `json:"serial_number,omitempty"`
	Settings             CustomHostnameSSLSettings       `json:"settings,omitempty"`
	Certificates         []CustomHostnameSSLCertificates `json:"certificates,omitempty"`
	// Deprecated: use ValidationRecords.
	// If there a single validation record, this will equal ValidationRecords[0] for backwards compatibility.
	SSLValidationRecord
	ValidationRecords []SSLValidationRecord `json:"validation_records,omitempty"`
	ValidationErrors  []SSLValidationError  `json:"validation_errors,omitempty"`
	BundleMethod      string                `json:"bundle_method,omitempty"`
}

// CustomMetadata defines custom metadata for the hostname. This requires logic to be implemented by Cloudflare to act on the data provided.
type CustomMetadata map[string]interface{}

// CustomHostname represents a custom hostname in a zone.
type CustomHostname struct {
	ID                        string                                  `json:"id,omitempty"`
	Hostname                  string                                  `json:"hostname,omitempty"`
	CustomOriginServer        string                                  `json:"custom_origin_server,omitempty"`
	CustomOriginSNI           string                                  `json:"custom_origin_sni,omitempty"`
	SSL                       *CustomHostnameSSL                      `json:"ssl,omitempty"`
	CustomMetadata            *CustomMetadata                         `json:"custom_metadata,omitempty"`
	Status                    CustomHostnameStatus                    `json:"status,omitempty"`
	VerificationErrors        []string                                `json:"verification_errors,omitempty"`
	OwnershipVerification     CustomHostnameOwnershipVerification     `json:"ownership_verification,omitempty"`
	OwnershipVerificationHTTP CustomHostnameOwnershipVerificationHTTP `json:"ownership_verification_http,omitempty"`
	CreatedAt                 *time.Time                              `json:"created_at,omitempty"`
}

// CustomHostnameOwnershipVerificationHTTP represents a response from the Custom Hostnames endpoints.
type CustomHostnameOwnershipVerificationHTTP struct {
	HTTPUrl  string `json:"http_url,omitempty"`
	HTTPBody string `json:"http_body,omitempty"`
}

// CustomHostnameResponse represents a response from the Custom Hostnames endpoints.
type CustomHostnameResponse struct {
	Result CustomHostname `json:"result"`
	Response
}

// CustomHostnameListResponse represents a response from the Custom Hostnames endpoints.
type CustomHostnameListResponse struct {
	Result []CustomHostname `json:"result"`
	Response
	ResultInfo `json:"result_info"`
}

// CustomHostnameFallbackOrigin represents a Custom Hostnames Fallback Origin.
type CustomHostnameFallbackOrigin struct {
	Origin string   `json:"origin,omitempty"`
	Status string   `json:"status,omitempty"`
	Errors []string `json:"errors,omitempty"`
}

// CustomHostnameFallbackOriginResponse represents a response from the Custom Hostnames Fallback Origin endpoint.
type CustomHostnameFallbackOriginResponse struct {
	Result CustomHostnameFallbackOrigin `json:"result"`
	Response
}

// UpdateCustomHostnameSSL modifies SSL configuration for the given custom
// hostname in the given zone.
//
// API reference: https://api.cloudflare.com/#custom-hostname-for-a-zone-update-custom-hostname-configuration
func (api *API) UpdateCustomHostnameSSL(ctx context.Context, zoneID string, customHostnameID string, ssl *CustomHostnameSSL) (*CustomHostnameResponse, error) {
	uri := fmt.Sprintf("/zones/%s/custom_hostnames/%s", zoneID, customHostnameID)
	ch := CustomHostname{
		SSL: ssl,
	}
	res, err := api.makeRequestContext(ctx, http.MethodPatch, uri, ch)
	if err != nil {
		return nil, err
	}

	var response *CustomHostnameResponse
	err = json.Unmarshal(res, &response)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}
	return response, nil
}

// UpdateCustomHostname modifies configuration for the given custom
// hostname in the given zone.
//
// API reference: https://api.cloudflare.com/#custom-hostname-for-a-zone-update-custom-hostname-configuration
func (api *API) UpdateCustomHostname(ctx context.Context, zoneID string, customHostnameID string, ch CustomHostname) (*CustomHostnameResponse, error) {
	uri := fmt.Sprintf("/zones/%s/custom_hostnames/%s", zoneID, customHostnameID)
	res, err := api.makeRequestContext(ctx, http.MethodPatch, uri, ch)
	if err != nil {
		return nil, err
	}

	var response *CustomHostnameResponse
	err = json.Unmarshal(res, &response)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}
	return response, nil
}

// DeleteCustomHostname deletes a custom hostname (and any issued SSL
// certificates).
//
// API reference: https://api.cloudflare.com/#custom-hostname-for-a-zone-delete-a-custom-hostname-and-any-issued-ssl-certificates-
func (api *API) DeleteCustomHostname(ctx context.Context, zoneID string, customHostnameID string) error {
	uri := fmt.Sprintf("/zones/%s/custom_hostnames/%s", zoneID, customHostnameID)
	res, err := api.makeRequestContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return err
	}

	var response *CustomHostnameResponse
	err = json.Unmarshal(res, &response)
	if err != nil {
		return fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return nil
}

// CreateCustomHostname creates a new custom hostname and requests that an SSL certificate be issued for it.
//
// API reference: https://api.cloudflare.com/#custom-hostname-for-a-zone-create-custom-hostname
func (api *API) CreateCustomHostname(ctx context.Context, zoneID string, ch CustomHostname) (*CustomHostnameResponse, error) {
	uri := fmt.Sprintf("/zones/%s/custom_hostnames", zoneID)
	res, err := api.makeRequestContext(ctx, http.MethodPost, uri, ch)
	if err != nil {
		return nil, err
	}

	var response *CustomHostnameResponse
	err = json.Unmarshal(res, &response)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return response, nil
}

// CustomHostnames fetches custom hostnames for the given zone,
// by applying filter.Hostname if not empty and scoping the result to page'th 50 items.
//
// The returned ResultInfo can be used to implement pagination.
//
// API reference: https://api.cloudflare.com/#custom-hostname-for-a-zone-list-custom-hostnames
func (api *API) CustomHostnames(ctx context.Context, zoneID string, page int, filter CustomHostname) ([]CustomHostname, ResultInfo, error) {
	v := url.Values{}
	v.Set("per_page", "50")
	v.Set("page", strconv.Itoa(page))
	if filter.Hostname != "" {
		v.Set("hostname", filter.Hostname)
	}

	uri := fmt.Sprintf("/zones/%s/custom_hostnames?%s", zoneID, v.Encode())
	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return []CustomHostname{}, ResultInfo{}, err
	}
	var customHostnameListResponse CustomHostnameListResponse
	err = json.Unmarshal(res, &customHostnameListResponse)
	if err != nil {
		return []CustomHostname{}, ResultInfo{}, err
	}

	return customHostnameListResponse.Result, customHostnameListResponse.ResultInfo, nil
}

// CustomHostname inspects the given custom hostname in the given zone.
//
// API reference: https://api.cloudflare.com/#custom-hostname-for-a-zone-custom-hostname-configuration-details
func (api *API) CustomHostname(ctx context.Context, zoneID string, customHostnameID string) (CustomHostname, error) {
	uri := fmt.Sprintf("/zones/%s/custom_hostnames/%s", zoneID, customHostnameID)
	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return CustomHostname{}, err
	}

	var response CustomHostnameResponse
	err = json.Unmarshal(res, &response)
	if err != nil {
		return CustomHostname{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return response.Result, nil
}

// CustomHostnameIDByName retrieves the ID for the given hostname in the given zone.
func (api *API) CustomHostnameIDByName(ctx context.Context, zoneID string, hostname string) (string, error) {
	customHostnames, _, err := api.CustomHostnames(ctx, zoneID, 1, CustomHostname{Hostname: hostname})
	if err != nil {
		return "", fmt.Errorf("CustomHostnames command failed: %w", err)
	}
	for _, ch := range customHostnames {
		if ch.Hostname == hostname {
			return ch.ID, nil
		}
	}
	return "", errors.New("CustomHostname could not be found")
}

// UpdateCustomHostnameFallbackOrigin modifies the Custom Hostname Fallback origin in the given zone.
//
// API reference: https://api.cloudflare.com/#custom-hostname-fallback-origin-for-a-zone-update-fallback-origin-for-custom-hostnames
func (api *API) UpdateCustomHostnameFallbackOrigin(ctx context.Context, zoneID string, chfo CustomHostnameFallbackOrigin) (*CustomHostnameFallbackOriginResponse, error) {
	uri := fmt.Sprintf("/zones/%s/custom_hostnames/fallback_origin", zoneID)
	res, err := api.makeRequestContext(ctx, http.MethodPut, uri, chfo)
	if err != nil {
		return nil, err
	}

	var response *CustomHostnameFallbackOriginResponse
	err = json.Unmarshal(res, &response)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}
	return response, nil
}

// DeleteCustomHostnameFallbackOrigin deletes the Custom Hostname Fallback origin in the given zone.
//
// API reference: https://api.cloudflare.com/#custom-hostname-fallback-origin-for-a-zone-delete-fallback-origin-for-custom-hostnames
func (api *API) DeleteCustomHostnameFallbackOrigin(ctx context.Context, zoneID string) error {
	uri := fmt.Sprintf("/zones/%s/custom_hostnames/fallback_origin", zoneID)
	res, err := api.makeRequestContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return err
	}

	var response *CustomHostnameFallbackOriginResponse
	err = json.Unmarshal(res, &response)
	if err != nil {
		return fmt.Errorf("%s: %w", errUnmarshalError, err)
	}
	return nil
}

// CustomHostnameFallbackOrigin inspects the Custom Hostname Fallback origin in the given zone.
//
// API reference: https://api.cloudflare.com/#custom-hostname-fallback-origin-for-a-zone-properties
func (api *API) CustomHostnameFallbackOrigin(ctx context.Context, zoneID string) (CustomHostnameFallbackOrigin, error) {
	uri := fmt.Sprintf("/zones/%s/custom_hostnames/fallback_origin", zoneID)
	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return CustomHostnameFallbackOrigin{}, err
	}

	var response CustomHostnameFallbackOriginResponse
	err = json.Unmarshal(res, &response)
	if err != nil {
		return CustomHostnameFallbackOrigin{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return response.Result, nil
}
