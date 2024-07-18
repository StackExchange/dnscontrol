package cloudflare

import (
	"context"
	"fmt"
	"net/http"

	"github.com/goccy/go-json"
)

// AccountRole defines the roles that a member can have attached.
type AccountRole struct {
	ID          string                           `json:"id"`
	Name        string                           `json:"name"`
	Description string                           `json:"description"`
	Permissions map[string]AccountRolePermission `json:"permissions"`
}

// AccountRolePermission is the shared structure for all permissions
// that can be assigned to a member.
type AccountRolePermission struct {
	Read bool `json:"read"`
	Edit bool `json:"edit"`
}

// AccountRolesListResponse represents the list response from the
// account roles.
type AccountRolesListResponse struct {
	Result []AccountRole `json:"result"`
	Response
	ResultInfo `json:"result_info"`
}

// AccountRoleDetailResponse is the API response, containing a single
// account role.
type AccountRoleDetailResponse struct {
	Success  bool        `json:"success"`
	Errors   []string    `json:"errors"`
	Messages []string    `json:"messages"`
	Result   AccountRole `json:"result"`
}

type ListAccountRolesParams struct {
	ResultInfo
}

// ListAccountRoles returns all roles of an account.
//
// API reference: https://developers.cloudflare.com/api/operations/account-roles-list-roles
func (api *API) ListAccountRoles(ctx context.Context, rc *ResourceContainer, params ListAccountRolesParams) ([]AccountRole, error) {
	if rc.Identifier == "" {
		return []AccountRole{}, ErrMissingAccountID
	}
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
	var roles []AccountRole
	var r AccountRolesListResponse
	for {
		uri := buildURI(fmt.Sprintf("/accounts/%s/roles", rc.Identifier), params)

		res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
		if err != nil {
			return []AccountRole{}, err
		}

		err = json.Unmarshal(res, &r)
		if err != nil {
			return []AccountRole{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
		}
		roles = append(roles, r.Result...)
		params.ResultInfo = r.ResultInfo.Next()
		if params.ResultInfo.Done() || !autoPaginate {
			break
		}
	}

	return roles, nil
}

// GetAccountRole returns the details of a single account role.
//
// API reference: https://developers.cloudflare.com/api/operations/account-roles-role-details
func (api *API) GetAccountRole(ctx context.Context, rc *ResourceContainer, roleID string) (AccountRole, error) {
	if rc.Identifier == "" {
		return AccountRole{}, ErrMissingAccountID
	}
	uri := fmt.Sprintf("/accounts/%s/roles/%s", rc.Identifier, roleID)

	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return AccountRole{}, fmt.Errorf("%s: %w", errMakeRequestError, err)
	}

	var accountRole AccountRoleDetailResponse
	err = json.Unmarshal(res, &accountRole)
	if err != nil {
		return AccountRole{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return accountRole.Result, nil
}
