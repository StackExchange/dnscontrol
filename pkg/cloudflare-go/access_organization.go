package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/goccy/go-json"
)

// AccessOrganization represents an Access organization.
type AccessOrganization struct {
	CreatedAt                      *time.Time                    `json:"created_at"`
	UpdatedAt                      *time.Time                    `json:"updated_at"`
	Name                           string                        `json:"name"`
	AuthDomain                     string                        `json:"auth_domain"`
	LoginDesign                    AccessOrganizationLoginDesign `json:"login_design"`
	IsUIReadOnly                   *bool                         `json:"is_ui_read_only,omitempty"`
	UIReadOnlyToggleReason         string                        `json:"ui_read_only_toggle_reason,omitempty"`
	UserSeatExpirationInactiveTime string                        `json:"user_seat_expiration_inactive_time,omitempty"`
	AutoRedirectToIdentity         *bool                         `json:"auto_redirect_to_identity,omitempty"`
	SessionDuration                *string                       `json:"session_duration,omitempty"`
	CustomPages                    AccessOrganizationCustomPages `json:"custom_pages,omitempty"`
	WarpAuthSessionDuration        *string                       `json:"warp_auth_session_duration,omitempty"`
	AllowAuthenticateViaWarp       *bool                         `json:"allow_authenticate_via_warp,omitempty"`
}

// AccessOrganizationLoginDesign represents the login design options.
type AccessOrganizationLoginDesign struct {
	BackgroundColor string `json:"background_color"`
	LogoPath        string `json:"logo_path"`
	TextColor       string `json:"text_color"`
	HeaderText      string `json:"header_text"`
	FooterText      string `json:"footer_text"`
}

type AccessOrganizationCustomPages struct {
	Forbidden      AccessCustomPageType `json:"forbidden,omitempty"`
	IdentityDenied AccessCustomPageType `json:"identity_denied,omitempty"`
}

// AccessOrganizationListResponse represents the response from the list
// access organization endpoint.
type AccessOrganizationListResponse struct {
	Result AccessOrganization `json:"result"`
	Response
	ResultInfo `json:"result_info"`
}

// AccessOrganizationDetailResponse is the API response, containing a
// single access organization.
type AccessOrganizationDetailResponse struct {
	Success  bool               `json:"success"`
	Errors   []string           `json:"errors"`
	Messages []string           `json:"messages"`
	Result   AccessOrganization `json:"result"`
}

type GetAccessOrganizationParams struct{}

type CreateAccessOrganizationParams struct {
	Name                           string                        `json:"name"`
	AuthDomain                     string                        `json:"auth_domain"`
	LoginDesign                    AccessOrganizationLoginDesign `json:"login_design"`
	IsUIReadOnly                   *bool                         `json:"is_ui_read_only,omitempty"`
	UIReadOnlyToggleReason         string                        `json:"ui_read_only_toggle_reason,omitempty"`
	UserSeatExpirationInactiveTime string                        `json:"user_seat_expiration_inactive_time,omitempty"`
	AutoRedirectToIdentity         *bool                         `json:"auto_redirect_to_identity,omitempty"`
	SessionDuration                *string                       `json:"session_duration,omitempty"`
	CustomPages                    AccessOrganizationCustomPages `json:"custom_pages,omitempty"`
	WarpAuthSessionDuration        *string                       `json:"warp_auth_session_duration,omitempty"`
	AllowAuthenticateViaWarp       *bool                         `json:"allow_authenticate_via_warp,omitempty"`
}

type UpdateAccessOrganizationParams struct {
	Name                           string                        `json:"name"`
	AuthDomain                     string                        `json:"auth_domain"`
	LoginDesign                    AccessOrganizationLoginDesign `json:"login_design"`
	IsUIReadOnly                   *bool                         `json:"is_ui_read_only,omitempty"`
	UIReadOnlyToggleReason         string                        `json:"ui_read_only_toggle_reason,omitempty"`
	UserSeatExpirationInactiveTime string                        `json:"user_seat_expiration_inactive_time,omitempty"`
	AutoRedirectToIdentity         *bool                         `json:"auto_redirect_to_identity,omitempty"`
	SessionDuration                *string                       `json:"session_duration,omitempty"`
	CustomPages                    AccessOrganizationCustomPages `json:"custom_pages,omitempty"`
	WarpAuthSessionDuration        *string                       `json:"warp_auth_session_duration,omitempty"`
	AllowAuthenticateViaWarp       *bool                         `json:"allow_authenticate_via_warp,omitempty"`
}

func (api *API) GetAccessOrganization(ctx context.Context, rc *ResourceContainer, params GetAccessOrganizationParams) (AccessOrganization, ResultInfo, error) {
	uri := fmt.Sprintf("/%s/%s/access/organizations", rc.Level, rc.Identifier)

	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return AccessOrganization{}, ResultInfo{}, err
	}

	var accessOrganizationListResponse AccessOrganizationListResponse
	err = json.Unmarshal(res, &accessOrganizationListResponse)
	if err != nil {
		return AccessOrganization{}, ResultInfo{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return accessOrganizationListResponse.Result, accessOrganizationListResponse.ResultInfo, nil
}

func (api *API) CreateAccessOrganization(ctx context.Context, rc *ResourceContainer, params CreateAccessOrganizationParams) (AccessOrganization, error) {
	uri := fmt.Sprintf("/%s/%s/access/organizations", rc.Level, rc.Identifier)

	res, err := api.makeRequestContext(ctx, http.MethodPost, uri, params)
	if err != nil {
		return AccessOrganization{}, err
	}

	var accessOrganizationDetailResponse AccessOrganizationDetailResponse
	err = json.Unmarshal(res, &accessOrganizationDetailResponse)
	if err != nil {
		return AccessOrganization{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return accessOrganizationDetailResponse.Result, nil
}

// UpdateAccessOrganization updates the Access organisation details.
//
// Account API reference: https://api.cloudflare.com/#access-organizations-update-access-organization
// Zone API reference: https://api.cloudflare.com/#zone-level-access-organizations-update-access-organization
func (api *API) UpdateAccessOrganization(ctx context.Context, rc *ResourceContainer, params UpdateAccessOrganizationParams) (AccessOrganization, error) {
	uri := fmt.Sprintf("/%s/%s/access/organizations", rc.Level, rc.Identifier)

	res, err := api.makeRequestContext(ctx, http.MethodPut, uri, params)
	if err != nil {
		return AccessOrganization{}, err
	}

	var accessOrganizationDetailResponse AccessOrganizationDetailResponse
	err = json.Unmarshal(res, &accessOrganizationDetailResponse)
	if err != nil {
		return AccessOrganization{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return accessOrganizationDetailResponse.Result, nil
}
