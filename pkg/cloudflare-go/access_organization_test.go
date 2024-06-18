package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAccessOrganization(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"created_at": "2014-01-01T05:20:00.12345Z",
				"updated_at": "2014-01-01T05:20:00.12345Z",
				"name": "Widget Corps Internal Applications",
				"auth_domain": "test.cloudflareaccess.com",
				"is_ui_read_only": false,
				"user_seat_expiration_inactive_time": "720h",
				"auto_redirect_to_identity": true,
				"allow_authenticate_via_warp": true,
				"warp_auth_session_duration": "24h",
				"session_duration": "12h",
				"login_design": {
					"background_color": "#c5ed1b",
					"logo_path": "https://example.com/logo.png",
					"text_color": "#c5ed1b",
					"header_text": "Widget Corp",
					"footer_text": "© Widget Corp"
				}
			}
		}
		`)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")

	want := AccessOrganization{
		Name:                     "Widget Corps Internal Applications",
		CreatedAt:                &createdAt,
		UpdatedAt:                &updatedAt,
		AuthDomain:               "test.cloudflareaccess.com",
		AllowAuthenticateViaWarp: BoolPtr(true),
		WarpAuthSessionDuration:  StringPtr("24h"),
		LoginDesign: AccessOrganizationLoginDesign{
			BackgroundColor: "#c5ed1b",
			LogoPath:        "https://example.com/logo.png",
			TextColor:       "#c5ed1b",
			HeaderText:      "Widget Corp",
			FooterText:      "© Widget Corp",
		},
		IsUIReadOnly:                   BoolPtr(false),
		SessionDuration:                StringPtr("12h"),
		UserSeatExpirationInactiveTime: "720h",
		AutoRedirectToIdentity:         BoolPtr(true),
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/organizations", handler)

	actual, _, err := client.GetAccessOrganization(context.Background(), testAccountRC, GetAccessOrganizationParams{})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/organizations", handler)

	actual, _, err = client.GetAccessOrganization(context.Background(), testZoneRC, GetAccessOrganizationParams{})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateAccessOrganization(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"created_at": "2014-01-01T05:20:00.12345Z",
				"updated_at": "2014-01-01T05:20:00.12345Z",
				"name": "Widget Corps Internal Applications",
				"auth_domain": "test.cloudflareaccess.com",
				"allow_authenticate_via_warp": true,
				"warp_auth_session_duration": "24h",
				"is_ui_read_only": true,
				"session_duration": "12h",
				"login_design": {
					"background_color": "#c5ed1b",
					"logo_path": "https://example.com/logo.png",
					"text_color": "#c5ed1b",
					"header_text": "Widget Corp",
					"footer_text": "© Widget Corp"
				}
			}
		}
		`)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")

	want := AccessOrganization{
		CreatedAt:                &createdAt,
		UpdatedAt:                &updatedAt,
		Name:                     "Widget Corps Internal Applications",
		AuthDomain:               "test.cloudflareaccess.com",
		AllowAuthenticateViaWarp: BoolPtr(true),
		WarpAuthSessionDuration:  StringPtr("24h"),
		LoginDesign: AccessOrganizationLoginDesign{
			BackgroundColor: "#c5ed1b",
			LogoPath:        "https://example.com/logo.png",
			TextColor:       "#c5ed1b",
			HeaderText:      "Widget Corp",
			FooterText:      "© Widget Corp",
		},
		IsUIReadOnly:    BoolPtr(true),
		SessionDuration: StringPtr("12h"),
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/organizations", handler)

	actual, err := client.CreateAccessOrganization(context.Background(), testAccountRC, CreateAccessOrganizationParams{
		Name:       "Widget Corps Internal Applications",
		AuthDomain: "test.cloudflareaccess.com",
		LoginDesign: AccessOrganizationLoginDesign{
			BackgroundColor: "#c5ed1b",
			LogoPath:        "https://example.com/logo.png",
			TextColor:       "#c5ed1b",
			HeaderText:      "Widget Corp",
			FooterText:      "© Widget Corp",
		},
		IsUIReadOnly:    BoolPtr(true),
		SessionDuration: StringPtr("12h"),
	})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/organizations", handler)

	actual, err = client.CreateAccessOrganization(context.Background(), testZoneRC, CreateAccessOrganizationParams{
		Name:       "Widget Corps Internal Applications",
		AuthDomain: "test.cloudflareaccess.com",
		LoginDesign: AccessOrganizationLoginDesign{
			BackgroundColor: "#c5ed1b",
			LogoPath:        "https://example.com/logo.png",
			TextColor:       "#c5ed1b",
			HeaderText:      "Widget Corp",
			FooterText:      "© Widget Corp",
		},
		IsUIReadOnly:    BoolPtr(true),
		SessionDuration: StringPtr("12h"),
	})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestUpdateAccessOrganization(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"created_at": "2014-01-01T05:20:00.12345Z",
				"updated_at": "2014-01-01T05:20:00.12345Z",
				"name": "Widget Corps Internal Applications",
				"auth_domain": "test.cloudflareaccess.com",
				"allow_authenticate_via_warp": false,
				"warp_auth_session_duration": "18h",
				"login_design": {
					"background_color": "#c5ed1b",
					"logo_path": "https://example.com/logo.png",
					"text_color": "#c5ed1b",
					"header_text": "Widget Corp",
					"footer_text": "© Widget Corp"
				},
				"is_ui_read_only": false,
				"ui_read_only_toggle_reason": "this is my reason",
				"session_duration": "12h"
			}
		}
		`)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")

	want := AccessOrganization{
		CreatedAt:                &createdAt,
		UpdatedAt:                &updatedAt,
		Name:                     "Widget Corps Internal Applications",
		AuthDomain:               "test.cloudflareaccess.com",
		WarpAuthSessionDuration:  StringPtr("18h"),
		AllowAuthenticateViaWarp: BoolPtr(false),
		LoginDesign: AccessOrganizationLoginDesign{
			BackgroundColor: "#c5ed1b",
			LogoPath:        "https://example.com/logo.png",
			TextColor:       "#c5ed1b",
			HeaderText:      "Widget Corp",
			FooterText:      "© Widget Corp",
		},
		IsUIReadOnly:           BoolPtr(false),
		UIReadOnlyToggleReason: "this is my reason",
		SessionDuration:        StringPtr("12h"),
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/organizations", handler)

	actual, err := client.UpdateAccessOrganization(context.Background(), testAccountRC, UpdateAccessOrganizationParams{
		Name:       "Widget Corps Internal Applications",
		AuthDomain: "test.cloudflareaccess.com",
		LoginDesign: AccessOrganizationLoginDesign{
			BackgroundColor: "#c5ed1b",
			LogoPath:        "https://example.com/logo.png",
			TextColor:       "#c5ed1b",
			HeaderText:      "Widget Corp",
			FooterText:      "© Widget Corp",
		},
		WarpAuthSessionDuration:  StringPtr("18h"),
		AllowAuthenticateViaWarp: BoolPtr(false),
		IsUIReadOnly:             BoolPtr(false),
		SessionDuration:          StringPtr("12h"),
		UIReadOnlyToggleReason:   "this is my reason",
	})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/organizations", handler)

	actual, err = client.UpdateAccessOrganization(context.Background(), testZoneRC, UpdateAccessOrganizationParams{
		Name:       "Widget Corps Internal Applications",
		AuthDomain: "test.cloudflareaccess.com",
		LoginDesign: AccessOrganizationLoginDesign{
			BackgroundColor: "#c5ed1b",
			LogoPath:        "https://example.com/logo.png",
			TextColor:       "#c5ed1b",
			HeaderText:      "Widget Corp",
			FooterText:      "© Widget Corp",
		},
		WarpAuthSessionDuration:  StringPtr("18h"),
		AllowAuthenticateViaWarp: BoolPtr(false),
		IsUIReadOnly:             BoolPtr(false),
		UIReadOnlyToggleReason:   "this is my reason",
		SessionDuration:          StringPtr("12h"),
	})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}
