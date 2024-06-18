package cloudflare

import (
	"context"
	"fmt"
	"net/http"

	"github.com/goccy/go-json"
)

// AccessCACertificate is the structure of the CA certificate used for
// short-lived certificates.
type AccessCACertificate struct {
	ID        string `json:"id"`
	Aud       string `json:"aud"`
	PublicKey string `json:"public_key"`
}

// AccessCACertificateListResponse represents the response of all CA
// certificates within Access.
type AccessCACertificateListResponse struct {
	Response
	Result []AccessCACertificate `json:"result"`
	ResultInfo
}

// AccessCACertificateResponse represents the response of a single CA
// certificate.
type AccessCACertificateResponse struct {
	Response
	Result AccessCACertificate `json:"result"`
}

type ListAccessCACertificatesParams struct {
	ResultInfo
}

type CreateAccessCACertificateParams struct {
	ApplicationID string
}

// ListAccessCACertificates returns all AccessCACertificate within Access.
//
// Account API reference: https://developers.cloudflare.com/api/operations/access-short-lived-certificate-c-as-list-short-lived-certificate-c-as
// Zone API reference: https://developers.cloudflare.com/api/operations/zone-level-access-short-lived-certificate-c-as-list-short-lived-certificate-c-as
func (api *API) ListAccessCACertificates(ctx context.Context, rc *ResourceContainer, params ListAccessCACertificatesParams) ([]AccessCACertificate, *ResultInfo, error) {
	baseURL := fmt.Sprintf("/%s/%s/access/apps/ca", rc.Level, rc.Identifier)

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

	var accessCACertificates []AccessCACertificate
	var r AccessCACertificateListResponse

	for {
		uri := buildURI(baseURL, params)
		res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
		if err != nil {
			return []AccessCACertificate{}, &ResultInfo{}, fmt.Errorf("%s: %w", errMakeRequestError, err)
		}

		err = json.Unmarshal(res, &r)
		if err != nil {
			return []AccessCACertificate{}, &ResultInfo{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
		}
		accessCACertificates = append(accessCACertificates, r.Result...)
		params.ResultInfo = r.ResultInfo.Next()
		if params.ResultInfo.Done() || !autoPaginate {
			break
		}
	}

	return accessCACertificates, &r.ResultInfo, nil
}

// GetAccessCACertificate returns a single CA certificate associated within
// Access.
//
// Account API reference: https://developers.cloudflare.com/api/operations/access-short-lived-certificate-c-as-get-a-short-lived-certificate-ca
// Zone API reference: https://developers.cloudflare.com/api/operations/zone-level-access-short-lived-certificate-c-as-get-a-short-lived-certificate-ca
func (api *API) GetAccessCACertificate(ctx context.Context, rc *ResourceContainer, applicationID string) (AccessCACertificate, error) {
	uri := fmt.Sprintf("/%s/%s/access/apps/%s/ca", rc.Level, rc.Identifier, applicationID)

	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return AccessCACertificate{}, err
	}

	var accessCAResponse AccessCACertificateResponse
	err = json.Unmarshal(res, &accessCAResponse)
	if err != nil {
		return AccessCACertificate{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return accessCAResponse.Result, nil
}

// CreateAccessCACertificate creates a new CA certificate for an AccessApplication.
//
// Account API reference: https://developers.cloudflare.com/api/operations/access-short-lived-certificate-c-as-create-a-short-lived-certificate-ca
// Zone API reference: https://developers.cloudflare.com/api/operations/zone-level-access-short-lived-certificate-c-as-create-a-short-lived-certificate-ca
func (api *API) CreateAccessCACertificate(ctx context.Context, rc *ResourceContainer, params CreateAccessCACertificateParams) (AccessCACertificate, error) {
	uri := fmt.Sprintf(
		"/%s/%s/access/apps/%s/ca",
		rc.Level,
		rc.Identifier,
		params.ApplicationID,
	)

	res, err := api.makeRequestContext(ctx, http.MethodPost, uri, nil)
	if err != nil {
		return AccessCACertificate{}, err
	}

	var accessCACertificate AccessCACertificateResponse
	err = json.Unmarshal(res, &accessCACertificate)
	if err != nil {
		return AccessCACertificate{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return accessCACertificate.Result, nil
}

// DeleteAccessCACertificate deletes an Access CA certificate on a defined
// AccessApplication.
//
// Account API reference: https://developers.cloudflare.com/api/operations/access-short-lived-certificate-c-as-delete-a-short-lived-certificate-ca
// Zone API reference: https://developers.cloudflare.com/api/operations/zone-level-access-short-lived-certificate-c-as-delete-a-short-lived-certificate-ca
func (api *API) DeleteAccessCACertificate(ctx context.Context, rc *ResourceContainer, applicationID string) error {
	uri := fmt.Sprintf(
		"/%s/%s/access/apps/%s/ca",
		rc.Level,
		rc.Identifier,
		applicationID,
	)

	_, err := api.makeRequestContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return err
	}

	return nil
}
