package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/goccy/go-json"
)

// ZoneLockdown represents a Zone Lockdown rule. A rule only permits access to
// the provided URL pattern(s) from the given IP address(es) or subnet(s).
type ZoneLockdown struct {
	ID             string               `json:"id"`
	Description    string               `json:"description"`
	URLs           []string             `json:"urls"`
	Configurations []ZoneLockdownConfig `json:"configurations"`
	Paused         bool                 `json:"paused"`
	Priority       int                  `json:"priority,omitempty"`
	CreatedOn      *time.Time           `json:"created_on,omitempty"`
	ModifiedOn     *time.Time           `json:"modified_on,omitempty"`
}

// ZoneLockdownConfig represents a Zone Lockdown config, which comprises
// a Target ("ip" or "ip_range") and a Value (an IP address or IP+mask,
// respectively.)
type ZoneLockdownConfig struct {
	Target string `json:"target"`
	Value  string `json:"value"`
}

// ZoneLockdownResponse represents a response from the Zone Lockdown endpoint.
type ZoneLockdownResponse struct {
	Result ZoneLockdown `json:"result"`
	Response
	ResultInfo `json:"result_info"`
}

// ZoneLockdownListResponse represents a response from the List Zone Lockdown
// endpoint.
type ZoneLockdownListResponse struct {
	Result []ZoneLockdown `json:"result"`
	Response
	ResultInfo `json:"result_info"`
}

// ZoneLockdownCreateParams contains required and optional params
// for creating a zone lockdown.
type ZoneLockdownCreateParams struct {
	Description    string               `json:"description"`
	URLs           []string             `json:"urls"`
	Configurations []ZoneLockdownConfig `json:"configurations"`
	Paused         bool                 `json:"paused"`
	Priority       int                  `json:"priority,omitempty"`
}

// ZoneLockdownUpdateParams contains required and optional params
// for updating a zone lockdown.
type ZoneLockdownUpdateParams struct {
	ID             string               `json:"id"`
	Description    string               `json:"description"`
	URLs           []string             `json:"urls"`
	Configurations []ZoneLockdownConfig `json:"configurations"`
	Paused         bool                 `json:"paused"`
	Priority       int                  `json:"priority,omitempty"`
}

type LockdownListParams struct {
	ResultInfo
}

// CreateZoneLockdown creates a Zone ZoneLockdown rule for the given zone ID.
//
// API reference: https://api.cloudflare.com/#zone-ZoneLockdown-create-a-ZoneLockdown-rule
func (api *API) CreateZoneLockdown(ctx context.Context, rc *ResourceContainer, params ZoneLockdownCreateParams) (ZoneLockdown, error) {
	uri := fmt.Sprintf("/zones/%s/firewall/lockdowns", rc.Identifier)
	res, err := api.makeRequestContext(ctx, http.MethodPost, uri, params)
	if err != nil {
		return ZoneLockdown{}, err
	}

	response := &ZoneLockdownResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return ZoneLockdown{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return response.Result, nil
}

// UpdateZoneLockdown updates a Zone ZoneLockdown rule (based on the ID) for the given zone ID.
//
// API reference: https://api.cloudflare.com/#zone-ZoneLockdown-update-ZoneLockdown-rule
func (api *API) UpdateZoneLockdown(ctx context.Context, rc *ResourceContainer, params ZoneLockdownUpdateParams) (ZoneLockdown, error) {
	uri := fmt.Sprintf("/zones/%s/firewall/lockdowns/%s", rc.Identifier, params.ID)
	res, err := api.makeRequestContext(ctx, http.MethodPut, uri, params)
	if err != nil {
		return ZoneLockdown{}, err
	}

	response := &ZoneLockdownResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return ZoneLockdown{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return response.Result, nil
}

// DeleteZoneLockdown deletes a Zone ZoneLockdown rule (based on the ID) for the given zone ID.
//
// API reference: https://api.cloudflare.com/#zone-ZoneLockdown-delete-ZoneLockdown-rule
func (api *API) DeleteZoneLockdown(ctx context.Context, rc *ResourceContainer, id string) (ZoneLockdown, error) {
	uri := fmt.Sprintf("/zones/%s/firewall/lockdowns/%s", rc.Identifier, id)
	res, err := api.makeRequestContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return ZoneLockdown{}, err
	}

	response := &ZoneLockdownResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return ZoneLockdown{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return response.Result, nil
}

// ZoneLockdown retrieves a Zone ZoneLockdown rule (based on the ID) for the given zone ID.
//
// API reference: https://api.cloudflare.com/#zone-ZoneLockdown-ZoneLockdown-rule-details
func (api *API) ZoneLockdown(ctx context.Context, rc *ResourceContainer, id string) (ZoneLockdown, error) {
	uri := fmt.Sprintf("/zones/%s/firewall/lockdowns/%s", rc.Identifier, id)
	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return ZoneLockdown{}, err
	}

	response := &ZoneLockdownResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return ZoneLockdown{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return response.Result, nil
}

// ListZoneLockdowns retrieves every Zone ZoneLockdown rules for a given zone ID.
//
// Automatically paginates all results unless `params.PerPage` and `params.Page`
// is set.
//
// API reference: https://api.cloudflare.com/#zone-ZoneLockdown-list-ZoneLockdown-rules
func (api *API) ListZoneLockdowns(ctx context.Context, rc *ResourceContainer, params LockdownListParams) ([]ZoneLockdown, *ResultInfo, error) {
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

	var zoneLockdowns []ZoneLockdown
	var zResponse ZoneLockdownListResponse
	for {
		zResponse = ZoneLockdownListResponse{}
		uri := buildURI(fmt.Sprintf("/zones/%s/firewall/lockdowns", rc.Identifier), params)

		res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
		if err != nil {
			return []ZoneLockdown{}, &ResultInfo{}, err
		}

		err = json.Unmarshal(res, &zResponse)
		if err != nil {
			return []ZoneLockdown{}, &ResultInfo{}, fmt.Errorf("failed to unmarshal filters JSON data: %w", err)
		}

		zoneLockdowns = append(zoneLockdowns, zResponse.Result...)
		params.ResultInfo = zResponse.ResultInfo.Next()

		if params.ResultInfo.Done() || !autoPaginate {
			break
		}
	}

	return zoneLockdowns, &zResponse.ResultInfo, nil
}
