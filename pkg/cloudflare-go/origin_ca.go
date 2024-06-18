package cloudflare

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/goccy/go-json"
)

// OriginCACertificate represents a Cloudflare-issued certificate.
//
// API reference: https://api.cloudflare.com/#cloudflare-ca
type OriginCACertificate struct {
	ID              string    `json:"id"`
	Certificate     string    `json:"certificate"`
	Hostnames       []string  `json:"hostnames"`
	ExpiresOn       time.Time `json:"expires_on"`
	RequestType     string    `json:"request_type"`
	RequestValidity int       `json:"requested_validity"`
	RevokedAt       time.Time `json:"revoked_at,omitempty"`
	CSR             string    `json:"csr"`
}

type CreateOriginCertificateParams struct {
	ID              string    `json:"id"`
	Certificate     string    `json:"certificate"`
	Hostnames       []string  `json:"hostnames"`
	ExpiresOn       time.Time `json:"expires_on"`
	RequestType     string    `json:"request_type"`
	RequestValidity int       `json:"requested_validity"`
	RevokedAt       time.Time `json:"revoked_at,omitempty"`
	CSR             string    `json:"csr"`
}

// UnmarshalJSON handles custom parsing from an API response to an OriginCACertificate
// http://choly.ca/post/go-json-marshalling/
func (c *OriginCACertificate) UnmarshalJSON(data []byte) error {
	type Alias OriginCACertificate

	aux := &struct {
		ExpiresOn string `json:"expires_on"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	var err error

	if err = json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// This format comes from time.Time.String() source
	c.ExpiresOn, err = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", aux.ExpiresOn)

	if err != nil {
		c.ExpiresOn, err = time.Parse(time.RFC3339, aux.ExpiresOn)
	}

	if err != nil {
		return err
	}

	return nil
}

// ListOriginCertificatesParams represents the parameters used to list
// Cloudflare-issued certificates.
type ListOriginCertificatesParams struct {
	ZoneID string `url:"zone_id,omitempty"`
}

// OriginCACertificateID represents the ID of the revoked certificate from the Revoke Certificate endpoint.
type OriginCACertificateID struct {
	ID string `json:"id"`
}

// originCACertificateResponse represents the response from the Create Certificate and the Certificate Details endpoints.
type originCACertificateResponse struct {
	Response
	Result OriginCACertificate `json:"result"`
}

// originCACertificateResponseList represents the response from the List Certificates endpoint.
type originCACertificateResponseList struct {
	Response
	Result     []OriginCACertificate `json:"result"`
	ResultInfo ResultInfo            `json:"result_info"`
}

// originCACertificateResponseRevoke represents the response from the Revoke Certificate endpoint.
type originCACertificateResponseRevoke struct {
	Response
	Result OriginCACertificateID `json:"result"`
}

// CreateOriginCACertificate creates a Cloudflare-signed certificate.
//
// API reference: https://api.cloudflare.com/#cloudflare-ca-create-certificate
func (api *API) CreateOriginCACertificate(ctx context.Context, params CreateOriginCertificateParams) (*OriginCACertificate, error) {
	res, err := api.makeRequestContext(ctx, http.MethodPost, "/certificates", params)
	if err != nil {
		return &OriginCACertificate{}, err
	}

	var originResponse *originCACertificateResponse

	err = json.Unmarshal(res, &originResponse)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	if !originResponse.Success {
		return nil, errors.New(errRequestNotSuccessful)
	}

	return &originResponse.Result, nil
}

// ListOriginCACertificates lists all Cloudflare-issued certificates.
//
// API reference: https://api.cloudflare.com/#cloudflare-ca-list-certificates
func (api *API) ListOriginCACertificates(ctx context.Context, params ListOriginCertificatesParams) ([]OriginCACertificate, error) {
	uri := buildURI("/certificates", params)
	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	var originResponse *originCACertificateResponseList

	err = json.Unmarshal(res, &originResponse)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	if !originResponse.Success {
		return nil, errors.New(errRequestNotSuccessful)
	}

	return originResponse.Result, nil
}

// GetOriginCACertificate returns the details for a Cloudflare-issued
// certificate.
//
// API reference: https://api.cloudflare.com/#cloudflare-ca-certificate-details
func (api *API) GetOriginCACertificate(ctx context.Context, certificateID string) (*OriginCACertificate, error) {
	uri := fmt.Sprintf("/certificates/%s", certificateID)
	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	var originResponse *originCACertificateResponse

	err = json.Unmarshal(res, &originResponse)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	if !originResponse.Success {
		return nil, errors.New(errRequestNotSuccessful)
	}

	return &originResponse.Result, nil
}

// RevokeOriginCACertificate revokes a created certificate for a zone.
//
// API reference: https://api.cloudflare.com/#cloudflare-ca-revoke-certificate
func (api *API) RevokeOriginCACertificate(ctx context.Context, certificateID string) (*OriginCACertificateID, error) {
	uri := fmt.Sprintf("/certificates/%s", certificateID)
	res, err := api.makeRequestContext(ctx, http.MethodDelete, uri, nil)

	if err != nil {
		return nil, err
	}

	var originResponse *originCACertificateResponseRevoke

	err = json.Unmarshal(res, &originResponse)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	if !originResponse.Success {
		return nil, errors.New(errRequestNotSuccessful)
	}

	return &originResponse.Result, nil
}

// Gets the Cloudflare Origin CA Root Certificate for a given algorithm in PEM format.
// Algorithm must be one of ['ecc', 'rsa'].
func GetOriginCARootCertificate(algorithm string) ([]byte, error) {
	var url string
	switch algorithm {
	case "ecc":
		url = originCARootCertEccURL
	case "rsa":
		url = originCARootCertRsaURL
	default:
		return nil, fmt.Errorf("invalid algorithm: must be one of ['ecc', 'rsa']")
	}

	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(errRequestNotSuccessful)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Response body could not be read: %w", err)
	}

	return body, nil
}
