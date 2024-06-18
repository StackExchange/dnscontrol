package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/goccy/go-json"
)

type AccessApprovalGroup struct {
	EmailListUuid   string   `json:"email_list_uuid,omitempty"`
	EmailAddresses  []string `json:"email_addresses,omitempty"`
	ApprovalsNeeded int      `json:"approvals_needed,omitempty"`
}

// AccessPolicy defines a policy for allowing or disallowing access to
// one or more Access applications.
type AccessPolicy struct {
	ID string `json:"id,omitempty"`
	// Precedence is the order in which the policy is executed in an Access application.
	// As a general rule, lower numbers take precedence over higher numbers.
	// This field can only be zero when a reusable policy is requested outside the context
	// of an Access application.
	Precedence int        `json:"precedence"`
	Decision   string     `json:"decision"`
	CreatedAt  *time.Time `json:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at"`
	Reusable   *bool      `json:"reusable,omitempty"`
	Name       string     `json:"name"`

	IsolationRequired            *bool                 `json:"isolation_required,omitempty"`
	SessionDuration              *string               `json:"session_duration,omitempty"`
	PurposeJustificationRequired *bool                 `json:"purpose_justification_required,omitempty"`
	PurposeJustificationPrompt   *string               `json:"purpose_justification_prompt,omitempty"`
	ApprovalRequired             *bool                 `json:"approval_required,omitempty"`
	ApprovalGroups               []AccessApprovalGroup `json:"approval_groups"`

	// The include policy works like an OR logical operator. The user must
	// satisfy one of the rules.
	Include []interface{} `json:"include"`

	// The exclude policy works like a NOT logical operator. The user must
	// not satisfy all the rules in exclude.
	Exclude []interface{} `json:"exclude"`

	// The require policy works like a AND logical operator. The user must
	// satisfy all the rules in require.
	Require []interface{} `json:"require"`
}

// AccessPolicyListResponse represents the response from the list
// access policies endpoint.
type AccessPolicyListResponse struct {
	Result []AccessPolicy `json:"result"`
	Response
	ResultInfo `json:"result_info"`
}

// AccessPolicyDetailResponse is the API response, containing a single
// access policy.
type AccessPolicyDetailResponse struct {
	Success  bool         `json:"success"`
	Errors   []string     `json:"errors"`
	Messages []string     `json:"messages"`
	Result   AccessPolicy `json:"result"`
}

type ListAccessPoliciesParams struct {
	// ApplicationID is the application ID to list attached access policies for.
	// If omitted, only reusable policies for the account are returned.
	ApplicationID string `json:"-"`
	ResultInfo
}

type GetAccessPolicyParams struct {
	PolicyID string `json:"-"`
	// ApplicationID is the application ID for which to scope the policy for.
	// Optional, but if included, the policy returned will include its execution precedence within the application.
	ApplicationID string `json:"-"`
}

type CreateAccessPolicyParams struct {
	// ApplicationID is the application ID for which to create the policy for.
	// Pass an empty value to create a reusable policy.
	ApplicationID string `json:"-"`

	// Precedence is the order in which the policy is executed in an Access application.
	// As a general rule, lower numbers take precedence over higher numbers.
	// This field is ignored when creating a reusable policy.
	// Read more here https://developers.cloudflare.com/cloudflare-one/policies/access/#order-of-execution
	Precedence int    `json:"precedence"`
	Decision   string `json:"decision"`
	Name       string `json:"name"`

	IsolationRequired            *bool                 `json:"isolation_required,omitempty"`
	SessionDuration              *string               `json:"session_duration,omitempty"`
	PurposeJustificationRequired *bool                 `json:"purpose_justification_required,omitempty"`
	PurposeJustificationPrompt   *string               `json:"purpose_justification_prompt,omitempty"`
	ApprovalRequired             *bool                 `json:"approval_required,omitempty"`
	ApprovalGroups               []AccessApprovalGroup `json:"approval_groups"`

	// The include policy works like an OR logical operator. The user must
	// satisfy one of the rules.
	Include []interface{} `json:"include"`

	// The exclude policy works like a NOT logical operator. The user must
	// not satisfy all the rules in exclude.
	Exclude []interface{} `json:"exclude"`

	// The require policy works like a AND logical operator. The user must
	// satisfy all the rules in require.
	Require []interface{} `json:"require"`
}

type UpdateAccessPolicyParams struct {
	// ApplicationID is the application ID that owns the existing policy.
	// Pass an empty value if the existing policy is reusable.
	ApplicationID string `json:"-"`
	PolicyID      string `json:"-"`

	// Precedence is the order in which the policy is executed in an Access application.
	// As a general rule, lower numbers take precedence over higher numbers.
	// This field is ignored when updating a reusable policy.
	Precedence int    `json:"precedence"`
	Decision   string `json:"decision"`
	Name       string `json:"name"`

	IsolationRequired            *bool                 `json:"isolation_required,omitempty"`
	SessionDuration              *string               `json:"session_duration,omitempty"`
	PurposeJustificationRequired *bool                 `json:"purpose_justification_required,omitempty"`
	PurposeJustificationPrompt   *string               `json:"purpose_justification_prompt,omitempty"`
	ApprovalRequired             *bool                 `json:"approval_required,omitempty"`
	ApprovalGroups               []AccessApprovalGroup `json:"approval_groups"`

	// The include policy works like an OR logical operator. The user must
	// satisfy one of the rules.
	Include []interface{} `json:"include"`

	// The exclude policy works like a NOT logical operator. The user must
	// not satisfy all the rules in exclude.
	Exclude []interface{} `json:"exclude"`

	// The require policy works like a AND logical operator. The user must
	// satisfy all the rules in require.
	Require []interface{} `json:"require"`
}

type DeleteAccessPolicyParams struct {
	// ApplicationID is the application ID the policy belongs to.
	// If the existing policy is reusable, this field must be omitted. Otherwise, it is required.
	ApplicationID string `json:"-"`
	PolicyID      string `json:"-"`
}

// ListAccessPolicies returns all access policies that match the parameters.
//
// Account API reference: https://developers.cloudflare.com/api/operations/access-policies-list-access-policies
// Zone API reference: https://developers.cloudflare.com/api/operations/zone-level-access-policies-list-access-policies
func (api *API) ListAccessPolicies(ctx context.Context, rc *ResourceContainer, params ListAccessPoliciesParams) ([]AccessPolicy, *ResultInfo, error) {
	var baseURL string
	if params.ApplicationID != "" {
		baseURL = fmt.Sprintf(
			"/%s/%s/access/apps/%s/policies",
			rc.Level,
			rc.Identifier,
			params.ApplicationID,
		)
	} else {
		baseURL = fmt.Sprintf(
			"/%s/%s/access/policies",
			rc.Level,
			rc.Identifier,
		)
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

	var accessPolicies []AccessPolicy
	var r AccessPolicyListResponse
	for {
		uri := buildURI(baseURL, params)
		res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
		if err != nil {
			return []AccessPolicy{}, &ResultInfo{}, fmt.Errorf("%s: %w", errMakeRequestError, err)
		}

		err = json.Unmarshal(res, &r)
		if err != nil {
			return []AccessPolicy{}, &ResultInfo{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
		}
		accessPolicies = append(accessPolicies, r.Result...)
		params.ResultInfo = r.ResultInfo.Next()
		if params.ResultInfo.Done() || !autoPaginate {
			break
		}
	}

	return accessPolicies, &r.ResultInfo, nil
}

// GetAccessPolicy returns a single policy based on the policy ID.
//
// Account API reference: https://developers.cloudflare.com/api/operations/access-policies-get-an-access-policy
// Zone API reference: https://developers.cloudflare.com/api/operations/zone-level-access-policies-get-an-access-policy
func (api *API) GetAccessPolicy(ctx context.Context, rc *ResourceContainer, params GetAccessPolicyParams) (AccessPolicy, error) {
	var uri string
	if params.ApplicationID != "" {
		uri = fmt.Sprintf(
			"/%s/%s/access/apps/%s/policies/%s",
			rc.Level,
			rc.Identifier,
			params.ApplicationID,
			params.PolicyID,
		)
	} else {
		uri = fmt.Sprintf(
			"/%s/%s/access/policies/%s",
			rc.Level,
			rc.Identifier,
			params.PolicyID,
		)
	}

	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return AccessPolicy{}, fmt.Errorf("%s: %w", errMakeRequestError, err)
	}

	var accessPolicyDetailResponse AccessPolicyDetailResponse
	err = json.Unmarshal(res, &accessPolicyDetailResponse)
	if err != nil {
		return AccessPolicy{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return accessPolicyDetailResponse.Result, nil
}

// CreateAccessPolicy creates a new access policy.
//
// Account API reference: https://developers.cloudflare.com/api/operations/access-policies-create-an-access-policy
// Zone API reference: https://developers.cloudflare.com/api/operations/zone-level-access-policies-create-an-access-policy
func (api *API) CreateAccessPolicy(ctx context.Context, rc *ResourceContainer, params CreateAccessPolicyParams) (AccessPolicy, error) {
	var uri string
	if params.ApplicationID != "" {
		uri = fmt.Sprintf(
			"/%s/%s/access/apps/%s/policies",
			rc.Level,
			rc.Identifier,
			params.ApplicationID,
		)
	} else {
		uri = fmt.Sprintf(
			"/%s/%s/access/policies",
			rc.Level,
			rc.Identifier,
		)
	}

	res, err := api.makeRequestContext(ctx, http.MethodPost, uri, params)
	if err != nil {
		return AccessPolicy{}, fmt.Errorf("%s: %w", errMakeRequestError, err)
	}

	var accessPolicyDetailResponse AccessPolicyDetailResponse
	err = json.Unmarshal(res, &accessPolicyDetailResponse)
	if err != nil {
		return AccessPolicy{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return accessPolicyDetailResponse.Result, nil
}

// UpdateAccessPolicy updates an existing access policy.
//
// Account API reference: https://developers.cloudflare.com/api/operations/access-policies-update-an-access-policy
// Zone API reference: https://developers.cloudflare.com/api/operations/zone-level-access-policies-update-an-access-policy
func (api *API) UpdateAccessPolicy(ctx context.Context, rc *ResourceContainer, params UpdateAccessPolicyParams) (AccessPolicy, error) {
	if params.PolicyID == "" {
		return AccessPolicy{}, fmt.Errorf("access policy ID cannot be empty")
	}

	var uri string
	if params.ApplicationID != "" {
		uri = fmt.Sprintf(
			"/%s/%s/access/apps/%s/policies/%s",
			rc.Level,
			rc.Identifier,
			params.ApplicationID,
			params.PolicyID,
		)
	} else {
		uri = fmt.Sprintf(
			"/%s/%s/access/policies/%s",
			rc.Level,
			rc.Identifier,
			params.PolicyID,
		)
	}

	res, err := api.makeRequestContext(ctx, http.MethodPut, uri, params)
	if err != nil {
		return AccessPolicy{}, fmt.Errorf("%s: %w", errMakeRequestError, err)
	}

	var accessPolicyDetailResponse AccessPolicyDetailResponse
	err = json.Unmarshal(res, &accessPolicyDetailResponse)
	if err != nil {
		return AccessPolicy{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return accessPolicyDetailResponse.Result, nil
}

// DeleteAccessPolicy deletes an access policy.
//
// Account API reference: https://developers.cloudflare.com/api/operations/access-policies-delete-an-access-policy
// Zone API reference: https://developers.cloudflare.com/api/operations/zone-level-access-policies-delete-an-access-policy
func (api *API) DeleteAccessPolicy(ctx context.Context, rc *ResourceContainer, params DeleteAccessPolicyParams) error {
	var uri string
	if params.ApplicationID != "" {
		uri = fmt.Sprintf(
			"/%s/%s/access/apps/%s/policies/%s",
			rc.Level,
			rc.Identifier,
			params.ApplicationID,
			params.PolicyID,
		)
	} else {
		uri = fmt.Sprintf(
			"/%s/%s/access/policies/%s",
			rc.Level,
			rc.Identifier,
			params.PolicyID,
		)
	}

	_, err := api.makeRequestContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", errMakeRequestError, err)
	}

	return nil
}
