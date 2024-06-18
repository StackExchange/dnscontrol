package cloudflare

import (
	"context"
	"fmt"
	"net/http"

	"github.com/goccy/go-json"
)

type WorkersSubdomain struct {
	Name string `json:"name,omitempty"`
}

type WorkersSubdomainResponse struct {
	Response
	Result WorkersSubdomain
}

// WorkersCreateSubdomain Creates a Workers subdomain for an account.
//
// API reference: https://developers.cloudflare.com/api/operations/worker-subdomain-create-subdomain
func (api *API) WorkersCreateSubdomain(ctx context.Context, rc *ResourceContainer, params WorkersSubdomain) (WorkersSubdomain, error) {
	if rc.Identifier == "" {
		return WorkersSubdomain{}, ErrMissingAccountID
	}

	uri := fmt.Sprintf("/accounts/%s/workers/subdomain", rc.Identifier)
	res, err := api.makeRequestContext(ctx, http.MethodPut, uri, params)
	if err != nil {
		return WorkersSubdomain{}, err
	}
	var r WorkersSubdomainResponse
	if err := json.Unmarshal(res, &r); err != nil {
		return WorkersSubdomain{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}
	return r.Result, nil
}

// WorkersGetSubdomain Creates a Workers subdomain for an account.
//
// API reference: https://developers.cloudflare.com/api/operations/worker-subdomain-get-subdomain
func (api *API) WorkersGetSubdomain(ctx context.Context, rc *ResourceContainer) (WorkersSubdomain, error) {
	if rc.Identifier == "" {
		return WorkersSubdomain{}, ErrMissingAccountID
	}

	uri := fmt.Sprintf("/accounts/%s/workers/subdomain", rc.Identifier)
	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return WorkersSubdomain{}, err
	}
	var r WorkersSubdomainResponse
	if err := json.Unmarshal(res, &r); err != nil {
		return WorkersSubdomain{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}
	return r.Result, nil
}
