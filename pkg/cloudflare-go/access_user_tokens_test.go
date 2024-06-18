package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRevokeAccessUserTokens(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"result": true
		}
		`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/organizations/revoke_user", handler)

	err := client.RevokeAccessUserTokens(context.Background(), testAccountRC, RevokeAccessUserTokensParams{Email: "test@example.com"})

	assert.NoError(t, err)
}

func TestRevokeZoneLevelAccessUserTokens(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"result": true
		}
		`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/organizations/revoke_user", handler)

	err := client.RevokeAccessUserTokens(context.Background(), testZoneRC, RevokeAccessUserTokensParams{Email: "test@example.com"})

	assert.NoError(t, err)
}
