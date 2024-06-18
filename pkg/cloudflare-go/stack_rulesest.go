package cloudflare

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

// DeleteRulesetRule removes a ruleset rule based on the ruleset ID +
// ruleset rule ID.
//
// API reference: https://developers.cloudflare.com/api/operations/deleteZoneRulesetRule
func (api *API) DeleteRulesetRule(ctx context.Context, rc *ResourceContainer, rulesetID, rulesetRuleID string) error {
	uri := fmt.Sprintf("/%s/%s/rulesets/%s/rules/%s", rc.Level, rc.Identifier, rulesetID, rulesetRuleID)
	res, err := api.makeRequestContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return err
	}

	// The API is not implementing the standard response blob but returns an
	// empty response (204) in case of a success. So we are checking for the
	// response body size here.
	if len(res) > 0 {
		return fmt.Errorf(errMakeRequestError+": %w", errors.New(string(res)))
	}

	return nil
}
