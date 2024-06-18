package cloudflare

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/goccy/go-json"
)

var (
	ErrMissingServiceTokenUUID = errors.New("missing required service token UUID")
)

// AccessServiceToken represents an Access Service Token.
type AccessServiceToken struct {
	ClientID  string     `json:"client_id"`
	CreatedAt *time.Time `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at"`
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	UpdatedAt *time.Time `json:"updated_at"`
	Duration  string     `json:"duration,omitempty"`
}

// AccessServiceTokenUpdateResponse represents the response from the API
// when a new Service Token is updated. This base struct is also used in the
// Create as they are very similar responses.
type AccessServiceTokenUpdateResponse struct {
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	ExpiresAt *time.Time `json:"expires_at"`
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	ClientID  string     `json:"client_id"`
	Duration  string     `json:"duration,omitempty"`
}

// AccessServiceTokenRefreshResponse represents the response from the API
// when an existing service token is refreshed to last longer.
type AccessServiceTokenRefreshResponse struct {
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	ExpiresAt *time.Time `json:"expires_at"`
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	ClientID  string     `json:"client_id"`
	Duration  string     `json:"duration,omitempty"`
}

// AccessServiceTokenCreateResponse is the same API response as the Update
// operation with the exception that the `ClientSecret` is present in a
// Create operation.
type AccessServiceTokenCreateResponse struct {
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
	ExpiresAt    *time.Time `json:"expires_at"`
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	ClientID     string     `json:"client_id"`
	ClientSecret string     `json:"client_secret"`
	Duration     string     `json:"duration,omitempty"`
}

// AccessServiceTokenRotateResponse is the same API response as the Create
// operation.
type AccessServiceTokenRotateResponse struct {
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
	ExpiresAt    *time.Time `json:"expires_at"`
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	ClientID     string     `json:"client_id"`
	ClientSecret string     `json:"client_secret"`
	Duration     string     `json:"duration,omitempty"`
}

// AccessServiceTokensListResponse represents the response from the list
// Access Service Tokens endpoint.
type AccessServiceTokensListResponse struct {
	Result []AccessServiceToken `json:"result"`
	Response
	ResultInfo `json:"result_info"`
}

// AccessServiceTokensDetailResponse is the API response, containing a single
// Access Service Token.
type AccessServiceTokensDetailResponse struct {
	Success  bool               `json:"success"`
	Errors   []string           `json:"errors"`
	Messages []string           `json:"messages"`
	Result   AccessServiceToken `json:"result"`
}

// AccessServiceTokensCreationDetailResponse is the API response, containing a
// single Access Service Token.
type AccessServiceTokensCreationDetailResponse struct {
	Success  bool                             `json:"success"`
	Errors   []string                         `json:"errors"`
	Messages []string                         `json:"messages"`
	Result   AccessServiceTokenCreateResponse `json:"result"`
}

// AccessServiceTokensUpdateDetailResponse is the API response, containing a
// single Access Service Token.
type AccessServiceTokensUpdateDetailResponse struct {
	Success  bool                             `json:"success"`
	Errors   []string                         `json:"errors"`
	Messages []string                         `json:"messages"`
	Result   AccessServiceTokenUpdateResponse `json:"result"`
}

// AccessServiceTokensRefreshDetailResponse is the API response, containing a
// single Access Service Token.
type AccessServiceTokensRefreshDetailResponse struct {
	Success  bool                              `json:"success"`
	Errors   []string                          `json:"errors"`
	Messages []string                          `json:"messages"`
	Result   AccessServiceTokenRefreshResponse `json:"result"`
}

// AccessServiceTokensRotateSecretDetailResponse is the API response, containing a
// single Access Service Token.
type AccessServiceTokensRotateSecretDetailResponse struct {
	Success  bool                             `json:"success"`
	Errors   []string                         `json:"errors"`
	Messages []string                         `json:"messages"`
	Result   AccessServiceTokenRotateResponse `json:"result"`
}

type ListAccessServiceTokensParams struct{}

type CreateAccessServiceTokenParams struct {
	Name     string `json:"name"`
	Duration string `json:"duration,omitempty"`
}

type UpdateAccessServiceTokenParams struct {
	Name     string `json:"name"`
	UUID     string `json:"-"`
	Duration string `json:"duration,omitempty"`
}

func (api *API) ListAccessServiceTokens(ctx context.Context, rc *ResourceContainer, params ListAccessServiceTokensParams) ([]AccessServiceToken, ResultInfo, error) {
	uri := fmt.Sprintf("/%s/%s/access/service_tokens", rc.Level, rc.Identifier)

	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return []AccessServiceToken{}, ResultInfo{}, err
	}

	var accessServiceTokensListResponse AccessServiceTokensListResponse
	err = json.Unmarshal(res, &accessServiceTokensListResponse)
	if err != nil {
		return []AccessServiceToken{}, ResultInfo{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return accessServiceTokensListResponse.Result, accessServiceTokensListResponse.ResultInfo, nil
}

func (api *API) CreateAccessServiceToken(ctx context.Context, rc *ResourceContainer, params CreateAccessServiceTokenParams) (AccessServiceTokenCreateResponse, error) {
	uri := fmt.Sprintf("/%s/%s/access/service_tokens", rc.Level, rc.Identifier)
	res, err := api.makeRequestContext(ctx, http.MethodPost, uri, params)

	if err != nil {
		return AccessServiceTokenCreateResponse{}, err
	}

	var accessServiceTokenCreation AccessServiceTokensCreationDetailResponse
	err = json.Unmarshal(res, &accessServiceTokenCreation)
	if err != nil {
		return AccessServiceTokenCreateResponse{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return accessServiceTokenCreation.Result, nil
}

func (api *API) UpdateAccessServiceToken(ctx context.Context, rc *ResourceContainer, params UpdateAccessServiceTokenParams) (AccessServiceTokenUpdateResponse, error) {
	if params.UUID == "" {
		return AccessServiceTokenUpdateResponse{}, ErrMissingServiceTokenUUID
	}

	uri := fmt.Sprintf("/%s/%s/access/service_tokens/%s", rc.Level, rc.Identifier, params.UUID)

	res, err := api.makeRequestContext(ctx, http.MethodPut, uri, params)
	if err != nil {
		return AccessServiceTokenUpdateResponse{}, err
	}

	var accessServiceTokenUpdate AccessServiceTokensUpdateDetailResponse
	err = json.Unmarshal(res, &accessServiceTokenUpdate)
	if err != nil {
		return AccessServiceTokenUpdateResponse{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return accessServiceTokenUpdate.Result, nil
}

func (api *API) DeleteAccessServiceToken(ctx context.Context, rc *ResourceContainer, uuid string) (AccessServiceTokenUpdateResponse, error) {
	uri := fmt.Sprintf("/%s/%s/access/service_tokens/%s", rc.Level, rc.Identifier, uuid)

	res, err := api.makeRequestContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return AccessServiceTokenUpdateResponse{}, err
	}

	var accessServiceTokenUpdate AccessServiceTokensUpdateDetailResponse
	err = json.Unmarshal(res, &accessServiceTokenUpdate)
	if err != nil {
		return AccessServiceTokenUpdateResponse{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return accessServiceTokenUpdate.Result, nil
}

// RefreshAccessServiceToken updates the expiry of an Access Service Token
// in place.
//
// API reference: https://api.cloudflare.com/#access-service-tokens-refresh-a-service-token
func (api *API) RefreshAccessServiceToken(ctx context.Context, rc *ResourceContainer, id string) (AccessServiceTokenRefreshResponse, error) {
	uri := fmt.Sprintf("/%s/%s/access/service_tokens/%s/refresh", rc.Level, rc.Identifier, id)

	res, err := api.makeRequestContext(ctx, http.MethodPost, uri, nil)
	if err != nil {
		return AccessServiceTokenRefreshResponse{}, err
	}

	var accessServiceTokenRefresh AccessServiceTokensRefreshDetailResponse
	err = json.Unmarshal(res, &accessServiceTokenRefresh)
	if err != nil {
		return AccessServiceTokenRefreshResponse{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return accessServiceTokenRefresh.Result, nil
}

// RotateAccessServiceToken rotates the client secret of an Access Service
// Token in place.
// API reference: https://api.cloudflare.com/#access-service-tokens-rotate-a-service-token
func (api *API) RotateAccessServiceToken(ctx context.Context, rc *ResourceContainer, id string) (AccessServiceTokenRotateResponse, error) {
	uri := fmt.Sprintf("/%s/%s/access/service_tokens/%s/rotate", rc.Level, rc.Identifier, id)

	res, err := api.makeRequestContext(ctx, http.MethodPost, uri, nil)
	if err != nil {
		return AccessServiceTokenRotateResponse{}, err
	}

	var accessServiceTokenRotate AccessServiceTokensRotateSecretDetailResponse
	err = json.Unmarshal(res, &accessServiceTokenRotate)
	if err != nil {
		return AccessServiceTokenRotateResponse{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return accessServiceTokenRotate.Result, nil
}
