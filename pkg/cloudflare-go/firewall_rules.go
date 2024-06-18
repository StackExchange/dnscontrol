package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/goccy/go-json"
)

// FirewallRule is the struct of the firewall rule.
type FirewallRule struct {
	ID          string      `json:"id,omitempty"`
	Paused      bool        `json:"paused"`
	Description string      `json:"description"`
	Action      string      `json:"action"`
	Priority    interface{} `json:"priority"`
	Filter      Filter      `json:"filter"`
	Products    []string    `json:"products,omitempty"`
	Ref         string      `json:"ref,omitempty"`
	CreatedOn   time.Time   `json:"created_on,omitempty"`
	ModifiedOn  time.Time   `json:"modified_on,omitempty"`
}

// FirewallRulesDetailResponse is the API response for the firewall
// rules.
type FirewallRulesDetailResponse struct {
	Result     []FirewallRule `json:"result"`
	ResultInfo `json:"result_info"`
	Response
}

// FirewallRuleResponse is the API response that is returned
// for requesting a single firewall rule on a zone.
type FirewallRuleResponse struct {
	Result     FirewallRule `json:"result"`
	ResultInfo `json:"result_info"`
	Response
}

// FirewallRuleCreateParams contains required and optional params
// for creating a firewall rule.
type FirewallRuleCreateParams struct {
	ID          string      `json:"id,omitempty"`
	Paused      bool        `json:"paused"`
	Description string      `json:"description"`
	Action      string      `json:"action"`
	Priority    interface{} `json:"priority"`
	Filter      Filter      `json:"filter"`
	Products    []string    `json:"products,omitempty"`
	Ref         string      `json:"ref,omitempty"`
}

// FirewallRuleUpdateParams contains required and optional params
// for updating a firewall rule.
type FirewallRuleUpdateParams struct {
	ID          string      `json:"id"`
	Paused      bool        `json:"paused"`
	Description string      `json:"description"`
	Action      string      `json:"action"`
	Priority    interface{} `json:"priority"`
	Filter      Filter      `json:"filter"`
	Products    []string    `json:"products,omitempty"`
	Ref         string      `json:"ref,omitempty"`
}

type FirewallRuleListParams struct {
	ResultInfo
}

// FirewallRules returns all firewall rules.
//
// Automatically paginates all results unless `params.PerPage` and `params.Page`
// is set.
//
// API reference: https://developers.cloudflare.com/firewall/api/cf-firewall-rules/get/#get-all-rules
func (api *API) FirewallRules(ctx context.Context, rc *ResourceContainer, params FirewallRuleListParams) ([]FirewallRule, *ResultInfo, error) {
	autoPaginate := true
	if params.PerPage >= 1 || params.Page >= 1 {
		autoPaginate = false
	}
	if params.PerPage < 1 {
		params.PerPage = 50
	}
	if params.Page < 1 {
		params.Page = 1
	}

	var firewallRules []FirewallRule
	var fResponse FirewallRulesDetailResponse
	for {
		fResponse = FirewallRulesDetailResponse{}
		uri := buildURI(fmt.Sprintf("/zones/%s/firewall/rules", rc.Identifier), params)

		res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
		if err != nil {
			return []FirewallRule{}, &ResultInfo{}, err
		}

		err = json.Unmarshal(res, &fResponse)
		if err != nil {
			return []FirewallRule{}, &ResultInfo{}, fmt.Errorf("failed to unmarshal filters JSON data: %w", err)
		}

		firewallRules = append(firewallRules, fResponse.Result...)
		params.ResultInfo = fResponse.ResultInfo.Next()

		if params.ResultInfo.Done() || !autoPaginate {
			break
		}
	}

	return firewallRules, &fResponse.ResultInfo, nil
}

// FirewallRule returns a single firewall rule based on the ID.
//
// API reference: https://developers.cloudflare.com/firewall/api/cf-firewall-rules/get/#get-by-rule-id
func (api *API) FirewallRule(ctx context.Context, rc *ResourceContainer, firewallRuleID string) (FirewallRule, error) {
	uri := fmt.Sprintf("/zones/%s/firewall/rules/%s", rc.Identifier, firewallRuleID)

	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return FirewallRule{}, err
	}

	var firewallRuleResponse FirewallRuleResponse
	err = json.Unmarshal(res, &firewallRuleResponse)
	if err != nil {
		return FirewallRule{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return firewallRuleResponse.Result, nil
}

// CreateFirewallRules creates new firewall rules.
//
// API reference: https://developers.cloudflare.com/firewall/api/cf-firewall-rules/post/
func (api *API) CreateFirewallRules(ctx context.Context, rc *ResourceContainer, params []FirewallRuleCreateParams) ([]FirewallRule, error) {
	uri := fmt.Sprintf("/zones/%s/firewall/rules", rc.Identifier)

	res, err := api.makeRequestContext(ctx, http.MethodPost, uri, params)
	if err != nil {
		return []FirewallRule{}, err
	}

	var firewallRulesDetailResponse FirewallRulesDetailResponse
	err = json.Unmarshal(res, &firewallRulesDetailResponse)
	if err != nil {
		return []FirewallRule{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return firewallRulesDetailResponse.Result, nil
}

// UpdateFirewallRule updates a single firewall rule.
//
// API reference: https://developers.cloudflare.com/firewall/api/cf-firewall-rules/put/#update-a-single-rule
func (api *API) UpdateFirewallRule(ctx context.Context, rc *ResourceContainer, params FirewallRuleUpdateParams) (FirewallRule, error) {
	if params.ID == "" {
		return FirewallRule{}, fmt.Errorf("firewall rule ID cannot be empty")
	}

	uri := fmt.Sprintf("/zones/%s/firewall/rules/%s", rc.Identifier, params.ID)

	res, err := api.makeRequestContext(ctx, http.MethodPut, uri, params)
	if err != nil {
		return FirewallRule{}, err
	}

	var firewallRuleResponse FirewallRuleResponse
	err = json.Unmarshal(res, &firewallRuleResponse)
	if err != nil {
		return FirewallRule{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return firewallRuleResponse.Result, nil
}

// UpdateFirewallRules updates a single firewall rule.
//
// API reference: https://developers.cloudflare.com/firewall/api/cf-firewall-rules/put/#update-multiple-rules
func (api *API) UpdateFirewallRules(ctx context.Context, rc *ResourceContainer, params []FirewallRuleUpdateParams) ([]FirewallRule, error) {
	for _, firewallRule := range params {
		if firewallRule.ID == "" {
			return []FirewallRule{}, fmt.Errorf("firewall ID cannot be empty")
		}
	}

	uri := fmt.Sprintf("/zones/%s/firewall/rules", rc.Identifier)

	res, err := api.makeRequestContext(ctx, http.MethodPut, uri, params)
	if err != nil {
		return []FirewallRule{}, err
	}

	var firewallRulesDetailResponse FirewallRulesDetailResponse
	err = json.Unmarshal(res, &firewallRulesDetailResponse)
	if err != nil {
		return []FirewallRule{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return firewallRulesDetailResponse.Result, nil
}

// DeleteFirewallRule deletes a single firewall rule.
//
// API reference: https://developers.cloudflare.com/firewall/api/cf-firewall-rules/delete/#delete-a-single-rule
func (api *API) DeleteFirewallRule(ctx context.Context, rc *ResourceContainer, firewallRuleID string) error {
	if firewallRuleID == "" {
		return fmt.Errorf("firewall rule ID cannot be empty")
	}

	uri := fmt.Sprintf("/zones/%s/firewall/rules/%s", rc.Identifier, firewallRuleID)

	_, err := api.makeRequestContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return err
	}

	return nil
}

// DeleteFirewallRules deletes multiple firewall rules at once.
//
// API reference: https://developers.cloudflare.com/firewall/api/cf-firewall-rules/delete/#delete-multiple-rules
func (api *API) DeleteFirewallRules(ctx context.Context, rc *ResourceContainer, firewallRuleIDs []string) error {
	v := url.Values{}

	for _, ruleID := range firewallRuleIDs {
		v.Add("id", ruleID)
	}

	uri := fmt.Sprintf("/zones/%s/firewall/rules?%s", rc.Identifier, v.Encode())

	_, err := api.makeRequestContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return err
	}

	return nil
}
