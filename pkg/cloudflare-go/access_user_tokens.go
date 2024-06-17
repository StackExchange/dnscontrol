package cloudflare

import (
	"context"
	"fmt"
	"net/http"
)

type RevokeAccessUserTokensParams struct {
	Email string `json:"email"`
}

// RevokeAccessUserTokens revokes any outstanding tokens issued for a specific user
// Access User.
func (api *API) RevokeAccessUserTokens(ctx context.Context, rc *ResourceContainer, params RevokeAccessUserTokensParams) error {
	uri := fmt.Sprintf("/%s/%s/access/organizations/revoke_user", rc.Level, rc.Identifier)

	_, err := api.makeRequestContext(ctx, http.MethodPost, uri, params)
	if err != nil {
		return err
	}

	return nil
}
