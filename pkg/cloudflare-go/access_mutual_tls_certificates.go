package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/goccy/go-json"
)

// AccessMutualTLSCertificate is the structure of a single Access Mutual TLS
// certificate.
type AccessMutualTLSCertificate struct {
	ID                  string    `json:"id,omitempty"`
	CreatedAt           time.Time `json:"created_at,omitempty"`
	UpdatedAt           time.Time `json:"updated_at,omitempty"`
	ExpiresOn           time.Time `json:"expires_on,omitempty"`
	Name                string    `json:"name,omitempty"`
	Fingerprint         string    `json:"fingerprint,omitempty"`
	Certificate         string    `json:"certificate,omitempty"`
	AssociatedHostnames []string  `json:"associated_hostnames,omitempty"`
}

// AccessMutualTLSCertificateListResponse is the API response for all Access
// Mutual TLS certificates.
type AccessMutualTLSCertificateListResponse struct {
	Response
	Result     []AccessMutualTLSCertificate `json:"result"`
	ResultInfo `json:"result_info"`
}

// AccessMutualTLSCertificateDetailResponse is the API response for a single
// Access Mutual TLS certificate.
type AccessMutualTLSCertificateDetailResponse struct {
	Response
	Result AccessMutualTLSCertificate `json:"result"`
}

type ListAccessMutualTLSCertificatesParams struct {
	ResultInfo
}

type CreateAccessMutualTLSCertificateParams struct {
	ExpiresOn           time.Time `json:"expires_on,omitempty"`
	Name                string    `json:"name,omitempty"`
	Fingerprint         string    `json:"fingerprint,omitempty"`
	Certificate         string    `json:"certificate,omitempty"`
	AssociatedHostnames []string  `json:"associated_hostnames,omitempty"`
}

type UpdateAccessMutualTLSCertificateParams struct {
	ID                  string    `json:"-"`
	ExpiresOn           time.Time `json:"expires_on,omitempty"`
	Name                string    `json:"name,omitempty"`
	Fingerprint         string    `json:"fingerprint,omitempty"`
	Certificate         string    `json:"certificate,omitempty"`
	AssociatedHostnames []string  `json:"associated_hostnames,omitempty"`
}

type AccessMutualTLSHostnameSettings struct {
	ChinaNetwork                *bool  `json:"china_network,omitempty"`
	ClientCertificateForwarding *bool  `json:"client_certificate_forwarding,omitempty"`
	Hostname                    string `json:"hostname,omitempty"`
}

type GetAccessMutualTLSHostnameSettingsResponse struct {
	Response
	Result []AccessMutualTLSHostnameSettings `json:"result"`
}

type UpdateAccessMutualTLSHostnameSettingsParams struct {
	Settings []AccessMutualTLSHostnameSettings `json:"settings,omitempty"`
}

// ListAccessMutualTLSCertificates returns all Access TLS certificates
//
// Account API Reference: https://developers.cloudflare.com/api/operations/access-mtls-authentication-list-mtls-certificates
// Zone API Reference: https://developers.cloudflare.com/api/operations/zone-level-access-mtls-authentication-list-mtls-certificates
func (api *API) ListAccessMutualTLSCertificates(ctx context.Context, rc *ResourceContainer, params ListAccessMutualTLSCertificatesParams) ([]AccessMutualTLSCertificate, *ResultInfo, error) {
	baseURL := fmt.Sprintf(
		"/%s/%s/access/certificates",
		rc.Level,
		rc.Identifier,
	)

	autoPaginate := true
	if params.PerPage >= 1 || params.Page >= 1 {
		autoPaginate = false
	}

	if params.PerPage < 1 {
		params.PerPage = 25
	}

	if params.Page < 1 {
		params.Page = 1
	}

	var accessCertificates []AccessMutualTLSCertificate
	var r AccessMutualTLSCertificateListResponse

	for {
		uri := buildURI(baseURL, params)
		res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
		if err != nil {
			return []AccessMutualTLSCertificate{}, &ResultInfo{}, fmt.Errorf("%s: %w", errMakeRequestError, err)
		}

		err = json.Unmarshal(res, &r)
		if err != nil {
			return []AccessMutualTLSCertificate{}, &ResultInfo{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
		}
		accessCertificates = append(accessCertificates, r.Result...)
		params.ResultInfo = r.ResultInfo.Next()
		if params.ResultInfo.Done() || !autoPaginate {
			break
		}
	}

	return accessCertificates, &r.ResultInfo, nil
}

// GetAccessMutualTLSCertificate returns a single Access Mutual TLS
// certificate.
//
// Account API Reference: https://developers.cloudflare.com/api/operations/access-mtls-authentication-get-an-mtls-certificate
// Zone API Reference: https://developers.cloudflare.com/api/operations/zone-level-access-mtls-authentication-get-an-mtls-certificate
func (api *API) GetAccessMutualTLSCertificate(ctx context.Context, rc *ResourceContainer, certificateID string) (AccessMutualTLSCertificate, error) {
	uri := fmt.Sprintf(
		"/%s/%s/access/certificates/%s",
		rc.Level,
		rc.Identifier,
		certificateID,
	)

	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return AccessMutualTLSCertificate{}, fmt.Errorf("%s: %w", errMakeRequestError, err)
	}

	var accessMutualTLSCertificateDetailResponse AccessMutualTLSCertificateDetailResponse
	err = json.Unmarshal(res, &accessMutualTLSCertificateDetailResponse)
	if err != nil {
		return AccessMutualTLSCertificate{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return accessMutualTLSCertificateDetailResponse.Result, nil
}

// CreateAccessMutualTLSCertificate creates an Access TLS Mutual
// certificate.
//
// Account API Reference: https://developers.cloudflare.com/api/operations/access-mtls-authentication-add-an-mtls-certificate
// Zone API Reference: https://developers.cloudflare.com/api/operations/zone-level-access-mtls-authentication-add-an-mtls-certificate
func (api *API) CreateAccessMutualTLSCertificate(ctx context.Context, rc *ResourceContainer, params CreateAccessMutualTLSCertificateParams) (AccessMutualTLSCertificate, error) {
	uri := fmt.Sprintf(
		"/%s/%s/access/certificates",
		rc.Level,
		rc.Identifier,
	)

	res, err := api.makeRequestContext(ctx, http.MethodPost, uri, params)
	if err != nil {
		return AccessMutualTLSCertificate{}, fmt.Errorf("%s: %w", errMakeRequestError, err)
	}

	var accessMutualTLSCertificateDetailResponse AccessMutualTLSCertificateDetailResponse
	err = json.Unmarshal(res, &accessMutualTLSCertificateDetailResponse)
	if err != nil {
		return AccessMutualTLSCertificate{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return accessMutualTLSCertificateDetailResponse.Result, nil
}

// UpdateAccessMutualTLSCertificate updates an account level Access TLS Mutual
// certificate.
//
// Account API Reference: https://developers.cloudflare.com/api/operations/access-mtls-authentication-update-an-mtls-certificate
// Zone API Reference: https://developers.cloudflare.com/api/operations/zone-level-access-mtls-authentication-update-an-mtls-certificate
func (api *API) UpdateAccessMutualTLSCertificate(ctx context.Context, rc *ResourceContainer, params UpdateAccessMutualTLSCertificateParams) (AccessMutualTLSCertificate, error) {
	uri := fmt.Sprintf(
		"/%s/%s/access/certificates/%s",
		rc.Level,
		rc.Identifier,
		params.ID,
	)

	res, err := api.makeRequestContext(ctx, http.MethodPut, uri, params)
	if err != nil {
		return AccessMutualTLSCertificate{}, fmt.Errorf("%s: %w", errMakeRequestError, err)
	}

	var accessMutualTLSCertificateDetailResponse AccessMutualTLSCertificateDetailResponse
	err = json.Unmarshal(res, &accessMutualTLSCertificateDetailResponse)
	if err != nil {
		return AccessMutualTLSCertificate{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return accessMutualTLSCertificateDetailResponse.Result, nil
}

// DeleteAccessMutualTLSCertificate destroys an Access Mutual
// TLS certificate.
//
// Account API Reference: https://developers.cloudflare.com/api/operations/access-mtls-authentication-delete-an-mtls-certificate
// Zone API Reference: https://developers.cloudflare.com/api/operations/zone-level-access-mtls-authentication-delete-an-mtls-certificate
func (api *API) DeleteAccessMutualTLSCertificate(ctx context.Context, rc *ResourceContainer, certificateID string) error {
	uri := fmt.Sprintf(
		"/%s/%s/access/certificates/%s",
		rc.Level,
		rc.Identifier,
		certificateID,
	)

	res, err := api.makeRequestContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", errMakeRequestError, err)
	}

	var accessMutualTLSCertificateDetailResponse AccessMutualTLSCertificateDetailResponse
	err = json.Unmarshal(res, &accessMutualTLSCertificateDetailResponse)
	if err != nil {
		return fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return nil
}

// GetAccessMutualTLSHostnameSettings returns all Access mTLS hostname settings.
//
// Account API Reference: https://developers.cloudflare.com/api/operations/access-mtls-authentication-update-an-mtls-certificate-settings
// Zone API Reference: https://developers.cloudflare.com/api/operations/zone-level-access-mtls-authentication-list-mtls-certificates-hostname-settings
func (api *API) GetAccessMutualTLSHostnameSettings(ctx context.Context, rc *ResourceContainer) ([]AccessMutualTLSHostnameSettings, error) {
	uri := fmt.Sprintf(
		"/%s/%s/access/certificates/settings",
		rc.Level,
		rc.Identifier,
	)

	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return []AccessMutualTLSHostnameSettings{}, fmt.Errorf("%s: %w", errMakeRequestError, err)
	}

	var accessMutualTLSHostnameSettingsResponse GetAccessMutualTLSHostnameSettingsResponse
	err = json.Unmarshal(res, &accessMutualTLSHostnameSettingsResponse)
	if err != nil {
		return []AccessMutualTLSHostnameSettings{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return accessMutualTLSHostnameSettingsResponse.Result, nil
}

// UpdateAccessMutualTLSHostnameSettings updates Access mTLS certificate hostname settings.
//
// Account API Reference: https://developers.cloudflare.com/api/operations/access-mtls-authentication-update-an-mtls-certificate-settings
// Zone API Reference: https://developers.cloudflare.com/api/operations/zone-level-access-mtls-authentication-update-an-mtls-certificate-settings
func (api *API) UpdateAccessMutualTLSHostnameSettings(ctx context.Context, rc *ResourceContainer, params UpdateAccessMutualTLSHostnameSettingsParams) ([]AccessMutualTLSHostnameSettings, error) {
	uri := fmt.Sprintf(
		"/%s/%s/access/certificates/settings",
		rc.Level,
		rc.Identifier,
	)

	res, err := api.makeRequestContext(ctx, http.MethodPut, uri, params)
	if err != nil {
		return []AccessMutualTLSHostnameSettings{}, fmt.Errorf("%s: %w", errMakeRequestError, err)
	}

	var accessMutualTLSHostnameSettingsResponse GetAccessMutualTLSHostnameSettingsResponse
	err = json.Unmarshal(res, &accessMutualTLSHostnameSettingsResponse)
	if err != nil {
		return []AccessMutualTLSHostnameSettings{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return accessMutualTLSHostnameSettingsResponse.Result, nil
}
