package cloudflare

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	testAccessApplicationID = "480f4f69-1a28-4fdd-9240-1ed29f0ac1db"
)

func TestAccessApplications(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": [
				{
					"id": "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
					"created_at": "2014-01-01T05:20:00.12345Z",
					"updated_at": "2014-01-01T05:20:00.12345Z",
					"aud": "737646a56ab1df6ec9bddc7e5ca84eaf3b0768850f3ffb5d74f1534911fe3893",
					"name": "Admin Site",
					"domain": "test.example.com/admin",
					"type": "self_hosted",
					"session_duration": "24h",
					"allowed_idps": ["f174e90a-fafe-4643-bbbc-4a0ed4fc8415"],
					"auto_redirect_to_identity": false,
					"enable_binding_cookie": false,
					"custom_deny_url": "https://www.example.com",
					"custom_non_identity_deny_url": "https://blocked.com",
					"custom_deny_message": "denied!",
					"http_only_cookie_attribute": true,
					"same_site_cookie_attribute": "strict",
					"logo_url": "https://www.example.com/example.png",
					"skip_interstitial": true,
					"app_launcher_visible": true,
					"service_auth_401_redirect": true,
					"path_cookie_attribute": true,
					"custom_pages": ["480f4f69-1a28-4fdd-9240-1ed29f0ac1dc"],
					"tags": ["engineers"],
					"allow_authenticate_via_warp": true,
					"options_preflight_bypass": false,
					"policies": [
						{
							"id": "699d98642c564d2e855e9661899b7252",
							"precedence": 1,
							"reusable": true,
							"decision": "allow",
							"created_at": "2014-01-01T05:20:00.12345Z",
							"updated_at": "2014-01-01T05:20:00.12345Z",
							"name": "Allow devs",
							"include": [
								{
									"email": {
										"email": "test@example.com"
									}
								}
							]
						}
					]
				}
			],
			"result_info": {
				"page": 1,
				"per_page": 20,
				"count": 1,
				"total_count": 1
			}
		}
		`)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")

	want := []AccessApplication{{
		ID:                       "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
		CreatedAt:                &createdAt,
		UpdatedAt:                &updatedAt,
		AUD:                      "737646a56ab1df6ec9bddc7e5ca84eaf3b0768850f3ffb5d74f1534911fe3893",
		Name:                     "Admin Site",
		Domain:                   "test.example.com/admin",
		Type:                     "self_hosted",
		SessionDuration:          "24h",
		AllowedIdps:              []string{"f174e90a-fafe-4643-bbbc-4a0ed4fc8415"},
		AutoRedirectToIdentity:   BoolPtr(false),
		EnableBindingCookie:      BoolPtr(false),
		AppLauncherVisible:       BoolPtr(true),
		ServiceAuth401Redirect:   BoolPtr(true),
		CustomDenyMessage:        "denied!",
		CustomDenyURL:            "https://www.example.com",
		SameSiteCookieAttribute:  "strict",
		HttpOnlyCookieAttribute:  BoolPtr(true),
		LogoURL:                  "https://www.example.com/example.png",
		SkipInterstitial:         BoolPtr(true),
		PathCookieAttribute:      BoolPtr(true),
		CustomPages:              []string{"480f4f69-1a28-4fdd-9240-1ed29f0ac1dc"},
		Tags:                     []string{"engineers"},
		CustomNonIdentityDenyURL: "https://blocked.com",
		AllowAuthenticateViaWarp: BoolPtr(true),
		OptionsPreflightBypass:   BoolPtr(false),
		Policies: []AccessPolicy{
			{
				ID:         "699d98642c564d2e855e9661899b7252",
				Precedence: 1,
				Reusable:   BoolPtr(true),
				Decision:   "allow",
				CreatedAt:  &createdAt,
				UpdatedAt:  &updatedAt,
				Name:       "Allow devs",
				Include: []any{
					map[string]interface{}{"email": map[string]interface{}{"email": "test@example.com"}},
				},
			},
		},
	}}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/apps", handler)

	actual, _, err := client.ListAccessApplications(context.Background(), AccountIdentifier(testAccountID), ListAccessApplicationsParams{})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/apps", handler)

	actual, _, err = client.ListAccessApplications(context.Background(), ZoneIdentifier(testZoneID), ListAccessApplicationsParams{})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestAccessApplication(t *testing.T) {
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
				"id": "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
				"created_at": "2014-01-01T05:20:00.12345Z",
				"updated_at": "2014-01-01T05:20:00.12345Z",
				"aud": "737646a56ab1df6ec9bddc7e5ca84eaf3b0768850f3ffb5d74f1534911fe3893",
				"name": "Admin Site",
				"domain": "test.example.com/admin",
				"self_hosted_domains": ["test.example.com/admin", "test.example.com/admin2"],
				"type": "self_hosted",
				"session_duration": "24h",
				"allowed_idps": ["f174e90a-fafe-4643-bbbc-4a0ed4fc8415"],
				"auto_redirect_to_identity": false,
				"enable_binding_cookie": false,
				"custom_deny_url": "https://www.example.com",
				"custom_non_identity_deny_url": "https://blocked.com",
				"custom_deny_message": "denied!",
				"logo_url": "https://www.example.com/example.png",
				"skip_interstitial": true,
				"app_launcher_visible": true,
				"service_auth_401_redirect": true,
				"http_only_cookie_attribute": false,
				"path_cookie_attribute": false,
				"allow_authenticate_via_warp": false,
				"options_preflight_bypass": false,
				"policies": [
					{
						"id": "699d98642c564d2e855e9661899b7252",
						"precedence": 1,
						"reusable": true,
						"decision": "allow",
						"created_at": "2014-01-01T05:20:00.12345Z",
						"updated_at": "2014-01-01T05:20:00.12345Z",
						"name": "Allow devs",
						"include": [
							{
								"email": {
									"email": "test@example.com"
								}
							}
						]
					}
				]
			}
		}
		`)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")

	want := AccessApplication{
		ID:                       "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
		CreatedAt:                &createdAt,
		UpdatedAt:                &updatedAt,
		AUD:                      "737646a56ab1df6ec9bddc7e5ca84eaf3b0768850f3ffb5d74f1534911fe3893",
		Name:                     "Admin Site",
		Domain:                   "test.example.com/admin",
		SelfHostedDomains:        []string{"test.example.com/admin", "test.example.com/admin2"},
		Type:                     "self_hosted",
		SessionDuration:          "24h",
		AllowedIdps:              []string{"f174e90a-fafe-4643-bbbc-4a0ed4fc8415"},
		AutoRedirectToIdentity:   BoolPtr(false),
		EnableBindingCookie:      BoolPtr(false),
		AppLauncherVisible:       BoolPtr(true),
		ServiceAuth401Redirect:   BoolPtr(true),
		CustomDenyMessage:        "denied!",
		CustomDenyURL:            "https://www.example.com",
		LogoURL:                  "https://www.example.com/example.png",
		SkipInterstitial:         BoolPtr(true),
		HttpOnlyCookieAttribute:  BoolPtr(false),
		PathCookieAttribute:      BoolPtr(false),
		CustomNonIdentityDenyURL: "https://blocked.com",
		AllowAuthenticateViaWarp: BoolPtr(false),
		OptionsPreflightBypass:   BoolPtr(false),
		Policies: []AccessPolicy{
			{
				ID:         "699d98642c564d2e855e9661899b7252",
				Precedence: 1,
				Reusable:   BoolPtr(true),
				Decision:   "allow",
				CreatedAt:  &createdAt,
				UpdatedAt:  &updatedAt,
				Name:       "Allow devs",
				Include: []any{
					map[string]interface{}{"email": map[string]interface{}{"email": "test@example.com"}},
				},
			},
		},
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/apps/480f4f69-1a28-4fdd-9240-1ed29f0ac1db", handler)

	actual, err := client.GetAccessApplication(context.Background(), AccountIdentifier(testAccountID), "480f4f69-1a28-4fdd-9240-1ed29f0ac1db")

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/apps/480f4f69-1a28-4fdd-9240-1ed29f0ac1db", handler)

	actual, err = client.GetAccessApplication(context.Background(), ZoneIdentifier(testZoneID), "480f4f69-1a28-4fdd-9240-1ed29f0ac1db")

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateAccessApplications(t *testing.T) {
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
				"id": "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
				"created_at": "2014-01-01T05:20:00.12345Z",
				"updated_at": "2014-01-01T05:20:00.12345Z",
				"aud": "737646a56ab1df6ec9bddc7e5ca84eaf3b0768850f3ffb5d74f1534911fe3893",
				"name": "Admin Site",
				"domain": "test.example.com/admin",
				"self_hosted_domains": ["test.example.com/admin", "test.example.com/admin2"],
				"type": "self_hosted",
				"session_duration": "24h",
				"allowed_idps": ["f174e90a-fafe-4643-bbbc-4a0ed4fc8415"],
				"auto_redirect_to_identity": false,
				"enable_binding_cookie": false,
				"custom_deny_url": "https://www.example.com",
				"custom_deny_message": "denied!",
				"custom_non_identity_deny_url": "https://blocked.com",
				"logo_url": "https://www.example.com/example.png",
				"skip_interstitial": true,
				"app_launcher_visible": true,
				"service_auth_401_redirect": true,
				"tags": ["engineers"],
				"allow_authenticate_via_warp": false,
				"options_preflight_bypass": true,
				"policies": [
					{
						"id": "699d98642c564d2e855e9661899b7252",
						"precedence": 1,
						"reusable": true,
						"decision": "allow",
						"created_at": "2014-01-01T05:20:00.12345Z",
						"updated_at": "2014-01-01T05:20:00.12345Z",
						"name": "Allow devs",
						"include": [
							{
								"email": {
									"email": "test@example.com"
								}
							}
						]
					}
				]
			}
		}
		`)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	fullAccessApplication := AccessApplication{
		ID:                       "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
		Name:                     "Admin Site",
		Domain:                   "test.example.com/admin",
		SelfHostedDomains:        []string{"test.example.com/admin", "test.example.com/admin2"},
		Type:                     "self_hosted",
		SessionDuration:          "24h",
		AUD:                      "737646a56ab1df6ec9bddc7e5ca84eaf3b0768850f3ffb5d74f1534911fe3893",
		AllowedIdps:              []string{"f174e90a-fafe-4643-bbbc-4a0ed4fc8415"},
		AutoRedirectToIdentity:   BoolPtr(false),
		EnableBindingCookie:      BoolPtr(false),
		AppLauncherVisible:       BoolPtr(true),
		ServiceAuth401Redirect:   BoolPtr(true),
		CustomDenyMessage:        "denied!",
		CustomDenyURL:            "https://www.example.com",
		LogoURL:                  "https://www.example.com/example.png",
		SkipInterstitial:         BoolPtr(true),
		CreatedAt:                &createdAt,
		UpdatedAt:                &updatedAt,
		CustomNonIdentityDenyURL: "https://blocked.com",
		Tags:                     []string{"engineers"},
		AllowAuthenticateViaWarp: BoolPtr(false),
		OptionsPreflightBypass:   BoolPtr(true),
		Policies: []AccessPolicy{
			{
				ID:         "699d98642c564d2e855e9661899b7252",
				Precedence: 1,
				Reusable:   BoolPtr(true),
				Decision:   "allow",
				CreatedAt:  &createdAt,
				UpdatedAt:  &updatedAt,
				Name:       "Allow devs",
				Include: []any{
					map[string]interface{}{"email": map[string]interface{}{"email": "test@example.com"}},
				},
			},
		},
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/apps", handler)

	actual, err := client.CreateAccessApplication(context.Background(), AccountIdentifier(testAccountID), CreateAccessApplicationParams{
		Name:            "Admin Site",
		Domain:          "test.example.com/admin",
		SessionDuration: "24h",
		Policies:        []string{"699d98642c564d2e855e9661899b7252"},
	})

	if assert.NoError(t, err) {
		assert.Equal(t, fullAccessApplication, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/apps", handler)

	actual, err = client.CreateAccessApplication(context.Background(), ZoneIdentifier(testZoneID), CreateAccessApplicationParams{
		Name:            "Admin Site",
		Domain:          "test.example.com/admin",
		SessionDuration: "24h",
		Policies:        []string{"699d98642c564d2e855e9661899b7252"},
	})

	if assert.NoError(t, err) {
		assert.Equal(t, fullAccessApplication, actual)
	}
}

func TestUpdateAccessApplication(t *testing.T) {
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
				"id": "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
				"created_at": "2014-01-01T05:20:00.12345Z",
				"updated_at": "2014-01-01T05:20:00.12345Z",
				"aud": "737646a56ab1df6ec9bddc7e5ca84eaf3b0768850f3ffb5d74f1534911fe3893",
				"name": "Admin Site",
				"domain": "test.example.com/admin",
				"self_hosted_domains": ["test.example.com/admin", "test.example.com/admin2"],
				"type": "self_hosted",
				"session_duration": "24h",
				"allowed_idps": ["f174e90a-fafe-4643-bbbc-4a0ed4fc8415"],
				"auto_redirect_to_identity": false,
				"enable_binding_cookie": false,
				"custom_deny_url": "https://www.example.com",
				"custom_deny_message": "denied!",
				"custom_non_identity_deny_url": "https://blocked.com",
				"logo_url": "https://www.example.com/example.png",
				"skip_interstitial": true,
				"app_launcher_visible": true,
				"service_auth_401_redirect": true,
				"tags": ["engineers"],
				"allow_authenticate_via_warp": true,
				"options_preflight_bypass": true,
				"policies": [
					{
						"id": "699d98642c564d2e855e9661899b7252",
						"precedence": 1,
						"reusable": true,
						"decision": "allow",
						"created_at": "2014-01-01T05:20:00.12345Z",
						"updated_at": "2014-01-01T05:20:00.12345Z",
						"name": "Allow devs",
						"include": [
							{
								"email": {
									"email": "test@example.com"
								}
							}
						]
					}
				]
			}
		}
		`)
	}

	fullAccessApplication := AccessApplication{
		ID:                       "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
		Name:                     "Admin Site",
		Domain:                   "test.example.com/admin",
		SelfHostedDomains:        []string{"test.example.com/admin", "test.example.com/admin2"},
		Type:                     "self_hosted",
		SessionDuration:          "24h",
		AUD:                      "737646a56ab1df6ec9bddc7e5ca84eaf3b0768850f3ffb5d74f1534911fe3893",
		AllowedIdps:              []string{"f174e90a-fafe-4643-bbbc-4a0ed4fc8415"},
		AutoRedirectToIdentity:   BoolPtr(false),
		EnableBindingCookie:      BoolPtr(false),
		AppLauncherVisible:       BoolPtr(true),
		ServiceAuth401Redirect:   BoolPtr(true),
		CustomDenyMessage:        "denied!",
		CustomDenyURL:            "https://www.example.com",
		LogoURL:                  "https://www.example.com/example.png",
		CustomNonIdentityDenyURL: "https://blocked.com",
		Tags:                     []string{"engineers"},
		SkipInterstitial:         BoolPtr(true),
		AllowAuthenticateViaWarp: BoolPtr(true),
		OptionsPreflightBypass:   BoolPtr(true),
		CreatedAt:                &createdAt,
		UpdatedAt:                &updatedAt,
		Policies: []AccessPolicy{
			{
				ID:         "699d98642c564d2e855e9661899b7252",
				Precedence: 1,
				Reusable:   BoolPtr(true),
				Decision:   "allow",
				CreatedAt:  &createdAt,
				UpdatedAt:  &updatedAt,
				Name:       "Allow devs",
				Include: []any{
					map[string]interface{}{"email": map[string]interface{}{"email": "test@example.com"}},
				},
			},
		},
	}

	params := UpdateAccessApplicationParams{
		ID:                       "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
		Name:                     "Admin Site",
		Domain:                   "test.example.com/admin",
		SelfHostedDomains:        []string{"test.example.com/admin", "test.example.com/admin2"},
		Type:                     "self_hosted",
		SessionDuration:          "24h",
		AUD:                      "737646a56ab1df6ec9bddc7e5ca84eaf3b0768850f3ffb5d74f1534911fe3893",
		AllowedIdps:              []string{"f174e90a-fafe-4643-bbbc-4a0ed4fc8415"},
		AutoRedirectToIdentity:   BoolPtr(false),
		EnableBindingCookie:      BoolPtr(false),
		AppLauncherVisible:       BoolPtr(true),
		ServiceAuth401Redirect:   BoolPtr(true),
		CustomDenyMessage:        "denied!",
		CustomDenyURL:            "https://www.example.com",
		LogoURL:                  "https://www.example.com/example.png",
		SkipInterstitial:         BoolPtr(true),
		CustomNonIdentityDenyURL: "https://blocked.com",
		Tags:                     []string{"engineers"},
		AllowAuthenticateViaWarp: BoolPtr(true),
		OptionsPreflightBypass:   BoolPtr(true),
		Policies:                 &[]string{"699d98642c564d2e855e9661899b7252"},
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/apps/480f4f69-1a28-4fdd-9240-1ed29f0ac1db", handler)

	actual, err := client.UpdateAccessApplication(context.Background(), AccountIdentifier(testAccountID), params)

	if assert.NoError(t, err) {
		assert.Equal(t, fullAccessApplication, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/apps/480f4f69-1a28-4fdd-9240-1ed29f0ac1db", handler)

	actual, err = client.UpdateAccessApplication(context.Background(), ZoneIdentifier(testZoneID), UpdateAccessApplicationParams{
		ID:                       "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
		Name:                     "Admin Site",
		Domain:                   "test.example.com/admin",
		SelfHostedDomains:        []string{"test.example.com/admin", "test.example.com/admin2"},
		Type:                     "self_hosted",
		SessionDuration:          "24h",
		AUD:                      "737646a56ab1df6ec9bddc7e5ca84eaf3b0768850f3ffb5d74f1534911fe3893",
		AllowedIdps:              []string{"f174e90a-fafe-4643-bbbc-4a0ed4fc8415"},
		AutoRedirectToIdentity:   BoolPtr(false),
		EnableBindingCookie:      BoolPtr(false),
		AppLauncherVisible:       BoolPtr(true),
		ServiceAuth401Redirect:   BoolPtr(true),
		CustomDenyMessage:        "denied!",
		CustomDenyURL:            "https://www.example.com",
		LogoURL:                  "https://www.example.com/example.png",
		SkipInterstitial:         BoolPtr(true),
		CustomNonIdentityDenyURL: "https://blocked.com",
		OptionsPreflightBypass:   BoolPtr(true),
		Policies:                 &[]string{"699d98642c564d2e855e9661899b7252"},
	})

	if assert.NoError(t, err) {
		assert.Equal(t, fullAccessApplication, actual)
	}
}

func TestUpdateAccessApplicationOmitPolicies(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		reqBody, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.NotContains(t, string(reqBody), "policies")
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
				"created_at": "2014-01-01T05:20:00.12345Z",
				"updated_at": "2014-01-01T05:20:00.12345Z",
				"aud": "737646a56ab1df6ec9bddc7e5ca84eaf3b0768850f3ffb5d74f1534911fe3893",
				"name": "Admin Site",
				"domain": "test.example.com/admin",
				"self_hosted_domains": ["test.example.com/admin", "test.example.com/admin2"],
				"type": "self_hosted",
				"session_duration": "24h",
				"allowed_idps": ["f174e90a-fafe-4643-bbbc-4a0ed4fc8415"],
				"auto_redirect_to_identity": false,
				"enable_binding_cookie": false,
				"custom_deny_url": "https://www.example.com",
				"custom_deny_message": "denied!",
				"custom_non_identity_deny_url": "https://blocked.com",
				"logo_url": "https://www.example.com/example.png",
				"skip_interstitial": true,
				"app_launcher_visible": true,
				"service_auth_401_redirect": true,
				"tags": ["engineers"],
				"allow_authenticate_via_warp": true,
				"options_preflight_bypass": true,
				"policies": [
					{
						"id": "699d98642c564d2e855e9661899b7252",
						"precedence": 1,
						"reusable": true,
						"decision": "allow",
						"created_at": "2014-01-01T05:20:00.12345Z",
						"updated_at": "2014-01-01T05:20:00.12345Z",
						"name": "Allow devs",
						"include": [
							{
								"email": {
									"email": "test@example.com"
								}
							}
						]
					}
				]
			}
		}
		`)
	}

	fullAccessApplication := AccessApplication{
		ID:                       "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
		Name:                     "Admin Site",
		Domain:                   "test.example.com/admin",
		SelfHostedDomains:        []string{"test.example.com/admin", "test.example.com/admin2"},
		Type:                     "self_hosted",
		SessionDuration:          "24h",
		AUD:                      "737646a56ab1df6ec9bddc7e5ca84eaf3b0768850f3ffb5d74f1534911fe3893",
		AllowedIdps:              []string{"f174e90a-fafe-4643-bbbc-4a0ed4fc8415"},
		AutoRedirectToIdentity:   BoolPtr(false),
		EnableBindingCookie:      BoolPtr(false),
		AppLauncherVisible:       BoolPtr(true),
		ServiceAuth401Redirect:   BoolPtr(true),
		CustomDenyMessage:        "denied!",
		CustomDenyURL:            "https://www.example.com",
		LogoURL:                  "https://www.example.com/example.png",
		CustomNonIdentityDenyURL: "https://blocked.com",
		Tags:                     []string{"engineers"},
		SkipInterstitial:         BoolPtr(true),
		AllowAuthenticateViaWarp: BoolPtr(true),
		OptionsPreflightBypass:   BoolPtr(true),
		CreatedAt:                &createdAt,
		UpdatedAt:                &updatedAt,
		Policies: []AccessPolicy{
			{
				ID:         "699d98642c564d2e855e9661899b7252",
				Precedence: 1,
				Reusable:   BoolPtr(true),
				Decision:   "allow",
				CreatedAt:  &createdAt,
				UpdatedAt:  &updatedAt,
				Name:       "Allow devs",
				Include: []any{
					map[string]interface{}{"email": map[string]interface{}{"email": "test@example.com"}},
				},
			},
		},
	}

	params := UpdateAccessApplicationParams{
		ID:                       "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
		Name:                     "Admin Site",
		Domain:                   "test.example.com/admin",
		SelfHostedDomains:        []string{"test.example.com/admin", "test.example.com/admin2"},
		Type:                     "self_hosted",
		SessionDuration:          "24h",
		AUD:                      "737646a56ab1df6ec9bddc7e5ca84eaf3b0768850f3ffb5d74f1534911fe3893",
		AllowedIdps:              []string{"f174e90a-fafe-4643-bbbc-4a0ed4fc8415"},
		AutoRedirectToIdentity:   BoolPtr(false),
		EnableBindingCookie:      BoolPtr(false),
		AppLauncherVisible:       BoolPtr(true),
		ServiceAuth401Redirect:   BoolPtr(true),
		CustomDenyMessage:        "denied!",
		CustomDenyURL:            "https://www.example.com",
		LogoURL:                  "https://www.example.com/example.png",
		SkipInterstitial:         BoolPtr(true),
		CustomNonIdentityDenyURL: "https://blocked.com",
		Tags:                     []string{"engineers"},
		AllowAuthenticateViaWarp: BoolPtr(true),
		OptionsPreflightBypass:   BoolPtr(true),
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/apps/480f4f69-1a28-4fdd-9240-1ed29f0ac1db", handler)

	actual, err := client.UpdateAccessApplication(context.Background(), AccountIdentifier(testAccountID), params)

	if assert.NoError(t, err) {
		assert.Equal(t, fullAccessApplication, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/apps/480f4f69-1a28-4fdd-9240-1ed29f0ac1db", handler)

	actual, err = client.UpdateAccessApplication(context.Background(), ZoneIdentifier(testZoneID), UpdateAccessApplicationParams{
		ID:                       "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
		Name:                     "Admin Site",
		Domain:                   "test.example.com/admin",
		SelfHostedDomains:        []string{"test.example.com/admin", "test.example.com/admin2"},
		Type:                     "self_hosted",
		SessionDuration:          "24h",
		AUD:                      "737646a56ab1df6ec9bddc7e5ca84eaf3b0768850f3ffb5d74f1534911fe3893",
		AllowedIdps:              []string{"f174e90a-fafe-4643-bbbc-4a0ed4fc8415"},
		AutoRedirectToIdentity:   BoolPtr(false),
		EnableBindingCookie:      BoolPtr(false),
		AppLauncherVisible:       BoolPtr(true),
		ServiceAuth401Redirect:   BoolPtr(true),
		CustomDenyMessage:        "denied!",
		CustomDenyURL:            "https://www.example.com",
		LogoURL:                  "https://www.example.com/example.png",
		SkipInterstitial:         BoolPtr(true),
		CustomNonIdentityDenyURL: "https://blocked.com",
		OptionsPreflightBypass:   BoolPtr(true),
	})

	if assert.NoError(t, err) {
		assert.Equal(t, fullAccessApplication, actual)
	}
}

func TestUpdateAccessApplicationWithMissingID(t *testing.T) {
	setup()
	defer teardown()

	_, err := client.UpdateAccessApplication(context.Background(), AccountIdentifier(testAccountID), UpdateAccessApplicationParams{})
	assert.EqualError(t, err, "access application ID cannot be empty")

	_, err = client.UpdateAccessApplication(context.Background(), AccountIdentifier(testAccountID), UpdateAccessApplicationParams{})
	assert.EqualError(t, err, "access application ID cannot be empty")
}

func TestDeleteAccessApplication(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
      "success": true,
      "errors": [],
      "messages": [],
      "result": {
        "id": "699d98642c564d2e855e9661899b7252"
      }
    }
    `)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/apps/480f4f69-1a28-4fdd-9240-1ed29f0ac1db", handler)
	err := client.DeleteAccessApplication(context.Background(), AccountIdentifier(testAccountID), "480f4f69-1a28-4fdd-9240-1ed29f0ac1db")

	assert.NoError(t, err)

	mux.HandleFunc("/zones/"+testZoneID+"/access/apps/480f4f69-1a28-4fdd-9240-1ed29f0ac1db", handler)
	err = client.DeleteAccessApplication(context.Background(), ZoneIdentifier(testZoneID), "480f4f69-1a28-4fdd-9240-1ed29f0ac1db")

	assert.NoError(t, err)
}

func TestRevokeAccessApplicationTokens(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
      "success": true,
      "errors": [],
      "messages": []
    }
    `)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/apps/480f4f69-1a28-4fdd-9240-1ed29f0ac1db/revoke-tokens", handler)
	err := client.RevokeAccessApplicationTokens(context.Background(), AccountIdentifier(testAccountID), "480f4f69-1a28-4fdd-9240-1ed29f0ac1db")

	assert.NoError(t, err)

	mux.HandleFunc("/zones/"+testZoneID+"/access/apps/480f4f69-1a28-4fdd-9240-1ed29f0ac1db/revoke-tokens", handler)
	err = client.RevokeAccessApplicationTokens(context.Background(), ZoneIdentifier(testZoneID), "480f4f69-1a28-4fdd-9240-1ed29f0ac1db")

	assert.NoError(t, err)
}

func TestAccessApplicationWithCORS(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
      "success": true,
      "errors": [],
      "messages": [

      ],
      "result":{
        "id": "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
        "created_at": "2014-01-01T05:20:00.12345Z",
        "updated_at": "2014-01-01T05:20:00.12345Z",
        "aud": "737646a56ab1df6ec9bddc7e5ca84eaf3b0768850f3ffb5d74f1534911fe3893",
        "name": "Admin Site",
        "domain": "test.example.com/admin",
        "type": "self_hosted",
        "session_duration": "24h",
        "cors_headers": {
          "allowed_methods": [
            "GET"
          ],
          "allowed_origins": [
            "https://example.com"
          ],
          "allow_all_headers": true,
          "max_age": -1
        }
      }
    }
    `)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")

	want := AccessApplication{
		ID:              "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
		CreatedAt:       &createdAt,
		UpdatedAt:       &updatedAt,
		AUD:             "737646a56ab1df6ec9bddc7e5ca84eaf3b0768850f3ffb5d74f1534911fe3893",
		Name:            "Admin Site",
		Domain:          "test.example.com/admin",
		Type:            "self_hosted",
		SessionDuration: "24h",
		CorsHeaders: &AccessApplicationCorsHeaders{
			AllowedMethods:  []string{http.MethodGet},
			AllowedOrigins:  []string{"https://example.com"},
			AllowAllHeaders: true,
			MaxAge:          -1,
		},
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/apps/480f4f69-1a28-4fdd-9240-1ed29f0ac1db", handler)

	actual, err := client.GetAccessApplication(context.Background(), AccountIdentifier(testAccountID), "480f4f69-1a28-4fdd-9240-1ed29f0ac1db")

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/apps/480f4f69-1a28-4fdd-9240-1ed29f0ac1db", handler)

	actual, err = client.GetAccessApplication(context.Background(), ZoneIdentifier(testZoneID), "480f4f69-1a28-4fdd-9240-1ed29f0ac1db")

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreatePrivateAccessApplication(t *testing.T) {
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
				"id": "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
				"created_at": "2014-01-01T05:20:00.12345Z",
				"updated_at": "2014-01-01T05:20:00.12345Z",
				"aud": "737646a56ab1df6ec9bddc7e5ca84eaf3b0768850f3ffb5d74f1534911fe3893",
				"name": "Private Admin Site",
				"private_address": "198.51.100.0",
				"type": "private_ip",
		    "gateway_rules": [
					{"id": "d9c61460-6f4d-4c40-89ea-2f552c9a8466"},
					{"id": "bc5ee7e7-9773-47a3-835e-b9b9799ebb92"}
				],
				"session_duration": "24h",
				"allowed_idps": ["f174e90a-fafe-4643-bbbc-4a0ed4fc8415"],
				"auto_redirect_to_identity": false,
				"enable_binding_cookie": false,
				"custom_deny_url": "https://www.example.com",
				"custom_deny_message": "denied!",
				"logo_url": "https://www.example.com/example.png",
				"skip_interstitial": true,
				"app_launcher_visible": false,
				"service_auth_401_redirect": false
			}
		}
		`)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	fullAccessApplication := AccessApplication{
		ID:             "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
		Name:           "Private Admin Site",
		PrivateAddress: "198.51.100.0",
		Type:           "private_ip",
		GatewayRules: []AccessApplicationGatewayRule{
			{ID: "d9c61460-6f4d-4c40-89ea-2f552c9a8466"},
			{ID: "bc5ee7e7-9773-47a3-835e-b9b9799ebb92"},
		},
		SessionDuration:        "24h",
		AUD:                    "737646a56ab1df6ec9bddc7e5ca84eaf3b0768850f3ffb5d74f1534911fe3893",
		AllowedIdps:            []string{"f174e90a-fafe-4643-bbbc-4a0ed4fc8415"},
		AutoRedirectToIdentity: BoolPtr(false),
		EnableBindingCookie:    BoolPtr(false),
		AppLauncherVisible:     BoolPtr(false),
		ServiceAuth401Redirect: BoolPtr(false),
		CustomDenyMessage:      "denied!",
		CustomDenyURL:          "https://www.example.com",
		LogoURL:                "https://www.example.com/example.png",
		SkipInterstitial:       BoolPtr(true),
		CreatedAt:              &createdAt,
		UpdatedAt:              &updatedAt,
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/apps", handler)

	actual, err := client.CreateAccessApplication(context.Background(), AccountIdentifier(testAccountID), CreateAccessApplicationParams{
		Name:            "Admin Site",
		PrivateAddress:  "198.51.100.0",
		SessionDuration: "24h",
		Type:            "private_ip",
	})

	if assert.NoError(t, err) {
		assert.Equal(t, fullAccessApplication, actual)
	}
}

func TestCreateSAMLSaasAccessApplications(t *testing.T) {
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
				"id": "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
				"created_at": "2014-01-01T05:20:00.12345Z",
				"updated_at": "2014-01-01T05:20:00.12345Z",
				"aud": "737646a56ab1df6ec9bddc7e5ca84eaf3b0768850f3ffb5d74f1534911fe3893",
				"name": "Admin Saas App",
				"domain": "example.cloudflareaccess.com/cdn-cgi/access/sso/saml/737646a56ab1df6ec9bddc7e5ca84eaf3b0768850f3ffb5d74f1534911fe3893",
				"type": "saas",
				"session_duration": "24h",
				"allowed_idps": [],
				"auto_redirect_to_identity": false,
				"enable_binding_cookie": false,
				"custom_deny_url": "https://www.example.com",
				"custom_deny_message": "denied!",
				"logo_url": "https://www.example.com/example.png",
				"skip_interstitial": true,
				"app_launcher_visible": true,
				"service_auth_401_redirect": true,
				"custom_non_identity_deny_url": "https://blocked.com",
				"tags": ["engineers"],
				"saas_app": {
					"consumer_service_url": "https://saas.example.com",
					"sp_entity_id": "dash.example.com",
					"name_id_format": "id",
					"default_relay_state": "https://saas.example.com",
					"custom_attributes": [
						{
							"name": "test1",
							"name_format": "urn:oasis:names:tc:SAML:2.0:attrname-format:unspecified",
							"source": {
								"name": "test1"
							}
						},
						{
							"name": "test2",
							"name_format": "urn:oasis:names:tc:SAML:2.0:attrname-format:basic",
							"source": {
								"name": "test2"
							}
						},
						{
							"name": "test3",
							"name_format": "urn:oasis:names:tc:SAML:2.0:attrname-format:uri",
							"source": {
								"name": "test3"
							}
						}
					],
					"name_id_transform_jsonata": "$substringBefore(email, '@') & '+sandbox@' & $substringAfter(email, '@')",
					"saml_attribute_transform_jsonata": "$ ~>| groups | {'group_name': name} |"
				}
			}
		}
		`)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	fullAccessApplication := AccessApplication{
		ID:                     "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
		Name:                   "Admin Saas App",
		Domain:                 "example.cloudflareaccess.com/cdn-cgi/access/sso/saml/737646a56ab1df6ec9bddc7e5ca84eaf3b0768850f3ffb5d74f1534911fe3893",
		Type:                   "saas",
		SessionDuration:        "24h",
		AUD:                    "737646a56ab1df6ec9bddc7e5ca84eaf3b0768850f3ffb5d74f1534911fe3893",
		AllowedIdps:            []string{},
		AutoRedirectToIdentity: BoolPtr(false),
		EnableBindingCookie:    BoolPtr(false),
		AppLauncherVisible:     BoolPtr(true),
		ServiceAuth401Redirect: BoolPtr(true),
		CustomDenyMessage:      "denied!",
		CustomDenyURL:          "https://www.example.com",
		LogoURL:                "https://www.example.com/example.png",
		SkipInterstitial:       BoolPtr(true),
		SaasApplication: &SaasApplication{
			ConsumerServiceUrl: "https://saas.example.com",
			SPEntityID:         "dash.example.com",
			NameIDFormat:       "id",
			DefaultRelayState:  "https://saas.example.com",
			CustomAttributes: []SAMLAttributeConfig{
				{
					Name:       "test1",
					NameFormat: "urn:oasis:names:tc:SAML:2.0:attrname-format:unspecified",
					Source: SourceConfig{
						Name: "test1",
					},
				},
				{
					Name:       "test2",
					NameFormat: "urn:oasis:names:tc:SAML:2.0:attrname-format:basic",
					Source: SourceConfig{
						Name: "test2",
					},
				},
				{
					Name:       "test3",
					NameFormat: "urn:oasis:names:tc:SAML:2.0:attrname-format:uri",
					Source: SourceConfig{
						Name: "test3",
					},
				},
			},
			NameIDTransformJsonata:        "$substringBefore(email, '@') & '+sandbox@' & $substringAfter(email, '@')",
			SamlAttributeTransformJsonata: "$ ~>| groups | {'group_name': name} |",
		},
		CreatedAt:                &createdAt,
		UpdatedAt:                &updatedAt,
		CustomNonIdentityDenyURL: "https://blocked.com",
		Tags:                     []string{"engineers"},
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/apps", handler)

	actual, err := client.CreateAccessApplication(context.Background(), AccountIdentifier(testAccountID), CreateAccessApplicationParams{
		Name: "Admin Saas Site",
		SaasApplication: &SaasApplication{
			ConsumerServiceUrl: "https://examplesaas.com",
			SPEntityID:         "TEST12345678",
			NameIDFormat:       "id",
		},
		SessionDuration: "24h",
	})

	if assert.NoError(t, err) {
		assert.Equal(t, fullAccessApplication, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/apps", handler)

	actual, err = client.CreateAccessApplication(context.Background(), ZoneIdentifier(testZoneID), CreateAccessApplicationParams{
		Name: "Admin Saas Site",
		SaasApplication: &SaasApplication{
			ConsumerServiceUrl: "https://saas.example.com",
			SPEntityID:         "TEST12345678",
			NameIDFormat:       "id",
		},
		SessionDuration: "24h",
	})

	if assert.NoError(t, err) {
		assert.Equal(t, fullAccessApplication, actual)
	}
}

func TestCreateOIDCSaasAccessApplications(t *testing.T) {
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
				"id": "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
				"created_at": "2014-01-01T05:20:00.12345Z",
				"updated_at": "2014-01-01T05:20:00.12345Z",
				"aud": "737646a56ab1df6ec9bddc7e5ca84eaf3b0768850f3ffb5d74f1534911fe3893",
				"name": "Admin OIDC Saas App",
				"domain": "example.cloudflareaccess.com/cdn-cgi/access/sso/oidc/737646a56ab1df6ec9bddc7e5ca84eaf3b0768850f3ffb5d74f1534911fe3893",
				"type": "saas",
				"session_duration": "24h",
				"allowed_idps": [],
				"auto_redirect_to_identity": false,
				"enable_binding_cookie": false,
				"custom_deny_url": "https://www.example.com",
				"custom_deny_message": "denied!",
				"logo_url": "https://www.example.com/example.png",
				"skip_interstitial": true,
				"app_launcher_visible": true,
				"service_auth_401_redirect": true,
				"custom_non_identity_deny_url": "https://blocked.com",
				"tags": ["engineers"],
				"saas_app": {
					"auth_type": "oidc",
					"client_id": "737646a56ab1df6ec9bddc7e5ca84eaf3b0768850f3ffb5d74f1534911fe3893",
					"client_secret": "secret",
					"redirect_uris": ["https://saas.example.com"],
					"grant_types": ["authorization_code", "hybrid", "implicit"],
					"scopes": ["openid", "email", "profile", "groups"],
					"app_launcher_url": "https://saas.example.com",
					"group_filter_regex": ".*",
					"allow_pkce_without_client_secret": false,
					"custom_claims": [
						{
							"name": "test1",
							"source": {
								"name": "test1"
							},
							"required": true,
							"scope": "profile"
						}
					],
					"hybrid_and_implicit_options": {
						"return_id_token_from_authorization_endpoint": true,
						"return_access_token_from_authorization_endpoint": true
					},
					"access_token_lifetime": "1m"
				}
			}
		}
		`)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	fullAccessApplication := AccessApplication{
		ID:                     "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
		Name:                   "Admin OIDC Saas App",
		Domain:                 "example.cloudflareaccess.com/cdn-cgi/access/sso/oidc/737646a56ab1df6ec9bddc7e5ca84eaf3b0768850f3ffb5d74f1534911fe3893",
		Type:                   "saas",
		SessionDuration:        "24h",
		AUD:                    "737646a56ab1df6ec9bddc7e5ca84eaf3b0768850f3ffb5d74f1534911fe3893",
		AllowedIdps:            []string{},
		AutoRedirectToIdentity: BoolPtr(false),
		EnableBindingCookie:    BoolPtr(false),
		AppLauncherVisible:     BoolPtr(true),
		ServiceAuth401Redirect: BoolPtr(true),
		CustomDenyMessage:      "denied!",
		CustomDenyURL:          "https://www.example.com",
		LogoURL:                "https://www.example.com/example.png",
		SkipInterstitial:       BoolPtr(true),
		SaasApplication: &SaasApplication{
			AuthType:                     "oidc",
			ClientID:                     "737646a56ab1df6ec9bddc7e5ca84eaf3b0768850f3ffb5d74f1534911fe3893",
			ClientSecret:                 "secret",
			RedirectURIs:                 []string{"https://saas.example.com"},
			GrantTypes:                   []string{"authorization_code", "hybrid", "implicit"},
			Scopes:                       []string{"openid", "email", "profile", "groups"},
			AppLauncherURL:               "https://saas.example.com",
			GroupFilterRegex:             ".*",
			AllowPKCEWithoutClientSecret: BoolPtr(false),
			CustomClaims: []OIDCClaimConfig{
				{
					Name:     "test1",
					Source:   SourceConfig{Name: "test1"},
					Required: BoolPtr(true),
					Scope:    "profile",
				},
			},
			HybridAndImplicitOptions: &AccessApplicationHybridAndImplicitOptions{
				ReturnIDTokenFromAuthorizationEndpoint:     BoolPtr(true),
				ReturnAccessTokenFromAuthorizationEndpoint: BoolPtr(true),
			},
			AccessTokenLifetime: "1m",
		},
		CreatedAt:                &createdAt,
		UpdatedAt:                &updatedAt,
		CustomNonIdentityDenyURL: "https://blocked.com",
		Tags:                     []string{"engineers"},
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/apps", handler)

	actual, err := client.CreateAccessApplication(context.Background(), AccountIdentifier(testAccountID), CreateAccessApplicationParams{
		Name: "Admin Saas Site",
		SaasApplication: &SaasApplication{
			AuthType:                     "oidc",
			GrantTypes:                   []string{"authorization_code", "hybrid", "implicit"},
			RedirectURIs:                 []string{"https://saas.example.com"},
			AppLauncherURL:               "https://saas.example.com",
			GroupFilterRegex:             ".*",
			AllowPKCEWithoutClientSecret: BoolPtr(false),
			CustomClaims: []OIDCClaimConfig{
				{
					Name:     "test1",
					Source:   SourceConfig{Name: "test1"},
					Required: BoolPtr(true),
					Scope:    "profile",
				},
			},
			HybridAndImplicitOptions: &AccessApplicationHybridAndImplicitOptions{
				ReturnIDTokenFromAuthorizationEndpoint:     BoolPtr(true),
				ReturnAccessTokenFromAuthorizationEndpoint: BoolPtr(true),
			},
			AccessTokenLifetime: "1m",
		},
		SessionDuration: "24h",
	})

	if assert.NoError(t, err) {
		assert.Equal(t, fullAccessApplication, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/apps", handler)

	actual, err = client.CreateAccessApplication(context.Background(), ZoneIdentifier(testZoneID), CreateAccessApplicationParams{
		Name: "Admin Saas Site",
		SaasApplication: &SaasApplication{
			AuthType:                     "oidc",
			GrantTypes:                   []string{"authorization_code", "hybrid", "implicit"},
			RedirectURIs:                 []string{"https://saas.example.com"},
			AppLauncherURL:               "https://saas.example.com",
			GroupFilterRegex:             ".*",
			AllowPKCEWithoutClientSecret: BoolPtr(false),
			CustomClaims: []OIDCClaimConfig{
				{
					Name:     "test1",
					Source:   SourceConfig{Name: "test1"},
					Required: BoolPtr(true),
					Scope:    "profile",
				},
			},
			HybridAndImplicitOptions: &AccessApplicationHybridAndImplicitOptions{
				ReturnIDTokenFromAuthorizationEndpoint:     BoolPtr(true),
				ReturnAccessTokenFromAuthorizationEndpoint: BoolPtr(true),
			},
			AccessTokenLifetime: "1m",
		},
		SessionDuration: "24h",
	})

	if assert.NoError(t, err) {
		assert.Equal(t, fullAccessApplication, actual)
	}
}

func TestCreateApplicationWithAccessAppLauncherCustomization(t *testing.T) {
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
				"id": "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
				"created_at": "2014-01-01T05:20:00.12345Z",
				"updated_at": "2014-01-01T05:20:00.12345Z",
				"aud": "737646a56ab1df6ec9bddc7e5ca84eaf3b0768850f3ffb5d74f1534911fe3893",
				"name": "App Launcher",
				"type": "app_launcher",
				"session_duration": "24h",
				"auto_redirect_to_identity": false,
				"enable_binding_cookie": false,
				"custom_deny_url": "https://www.example.com",
				"custom_deny_message": "denied!",
				"logo_url": "https://www.example.com/example.png",
				"skip_interstitial": true,
				"app_launcher_visible": false,
				"service_auth_401_redirect": false,
				"landing_page_design": {
					"title": "A title",
					"message": "a message",
					"image_url": "https://www.example.com/example.png",
					"button_color": "green",
					"button_text_color": "red"
				},
				"header_bg_color": "red",
				"bg_color": "blue",
				"footer_links": [
					{
						"url": "https://somesite.com",
						"name": "bug"
					}
				]
			}
		}
		`)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	fullAccessApplication := AccessApplication{
		ID:                     "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
		Name:                   "App Launcher",
		Type:                   "app_launcher",
		SessionDuration:        "24h",
		AUD:                    "737646a56ab1df6ec9bddc7e5ca84eaf3b0768850f3ffb5d74f1534911fe3893",
		AutoRedirectToIdentity: BoolPtr(false),
		EnableBindingCookie:    BoolPtr(false),
		AppLauncherVisible:     BoolPtr(false),
		ServiceAuth401Redirect: BoolPtr(false),
		CustomDenyMessage:      "denied!",
		CustomDenyURL:          "https://www.example.com",
		LogoURL:                "https://www.example.com/example.png",
		SkipInterstitial:       BoolPtr(true),
		CreatedAt:              &createdAt,
		UpdatedAt:              &updatedAt,
		AccessAppLauncherCustomization: AccessAppLauncherCustomization{
			LandingPageDesign: AccessLandingPageDesign{
				Title:           "A title",
				Message:         "a message",
				ImageURL:        "https://www.example.com/example.png",
				ButtonColor:     "green",
				ButtonTextColor: "red",
			},
			HeaderBackgroundColor: "red",
			BackgroundColor:       "blue",
			FooterLinks: []AccessFooterLink{
				{
					URL:  "https://somesite.com",
					Name: "bug",
				},
			},
		},
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/apps", handler)

	actual, err := client.CreateAccessApplication(context.Background(), AccountIdentifier(testAccountID), CreateAccessApplicationParams{
		Name:            "Admin Site",
		SessionDuration: "24h",
		Type:            "app_launcher",
		AccessAppLauncherCustomization: AccessAppLauncherCustomization{
			LandingPageDesign: AccessLandingPageDesign{
				Title:           "A title",
				Message:         "a message",
				ImageURL:        "https://www.example.com/example.png",
				ButtonColor:     "green",
				ButtonTextColor: "red",
			},
			HeaderBackgroundColor: "red",
			BackgroundColor:       "blue",
			FooterLinks: []AccessFooterLink{
				{
					URL:  "https://somesite.com",
					Name: "bug",
				},
			},
		},
	})

	if assert.NoError(t, err) {
		assert.Equal(t, fullAccessApplication, actual)
	}
}

func TestCreateAccessApplicationWithSCIMProvisioning(t *testing.T) {
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
				"id": "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
				"created_at": "2014-01-01T05:20:00.12345Z",
				"updated_at": "2014-01-01T05:20:00.12345Z",
				"aud": "737646a56ab1df6ec9bddc7e5ca84eaf3b0768850f3ffb5d74f1534911fe3893",
				"name": "Admin SCIM App",
				"domain": "example.cloudflareaccess.com/cdn-cgi/access/sso/oidc/737646a56ab1df6ec9bddc7e5ca84eaf3b0768850f3ffb5d74f1534911fe3893",
				"type": "saas",
				"session_duration": "24h",
				"allowed_idps": [],
				"auto_redirect_to_identity": false,
				"enable_binding_cookie": false,
				"custom_deny_url": "https://www.example.com",
				"custom_deny_message": "denied!",
				"logo_url": "https://www.example.com/example.png",
				"skip_interstitial": true,
				"app_launcher_visible": true,
				"service_auth_401_redirect": true,
				"custom_non_identity_deny_url": "https://blocked.com",
				"tags": ["engineers"],
				"scim_config": {
					"enabled": true,
					"remote_uri": "https://scim.com",
					"authentication": {
						 "scheme": "oauthbearertoken",
						 "token": "1234567890"
					},
					"idp_uid": "1234567",
					"deactivate_on_delete": true,
					"mappings": [
						{
							"schema": "urn:ietf:params:scim:schemas:core:2.0:User",
							"enabled": true,
							"filter": "title pr or userType eq \"Intern\"",
							"transform_jsonata": "$merge([$, {'userName': $substringBefore($.userName, '@') & '+test@' & $substringAfter($.userName, '@')}])",
							"operations": {
								"create": true,
								"update": true,
								"delete": true
							}
						}
					]
				}
			}
		}
		`)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	fullAccessApplication := AccessApplication{
		ID:                       "480f4f69-1a28-4fdd-9240-1ed29f0ac1db",
		Name:                     "Admin SCIM App",
		Domain:                   "example.cloudflareaccess.com/cdn-cgi/access/sso/oidc/737646a56ab1df6ec9bddc7e5ca84eaf3b0768850f3ffb5d74f1534911fe3893",
		Type:                     "saas",
		SessionDuration:          "24h",
		AUD:                      "737646a56ab1df6ec9bddc7e5ca84eaf3b0768850f3ffb5d74f1534911fe3893",
		AllowedIdps:              []string{},
		AutoRedirectToIdentity:   BoolPtr(false),
		EnableBindingCookie:      BoolPtr(false),
		AppLauncherVisible:       BoolPtr(true),
		ServiceAuth401Redirect:   BoolPtr(true),
		CustomDenyMessage:        "denied!",
		CustomDenyURL:            "https://www.example.com",
		LogoURL:                  "https://www.example.com/example.png",
		SkipInterstitial:         BoolPtr(true),
		CreatedAt:                &createdAt,
		UpdatedAt:                &updatedAt,
		CustomNonIdentityDenyURL: "https://blocked.com",
		Tags:                     []string{"engineers"},
		SCIMConfig: &AccessApplicationSCIMConfig{
			Enabled:   BoolPtr(true),
			RemoteURI: "https://scim.com",
			Authentication: &AccessApplicationScimAuthenticationJson{
				Value: &AccessApplicationScimAuthenticationOauthBearerToken{
					Token:                  "1234567890",
					baseScimAuthentication: baseScimAuthentication{Scheme: AccessApplicationScimAuthenticationSchemeOauthBearerToken},
				},
			},
			IdPUID:             "1234567",
			DeactivateOnDelete: BoolPtr(true),
			Mappings: []*AccessApplicationScimMapping{
				{
					Schema:           "urn:ietf:params:scim:schemas:core:2.0:User",
					Enabled:          BoolPtr(true),
					Filter:           "title pr or userType eq \"Intern\"",
					TransformJsonata: "$merge([$, {'userName': $substringBefore($.userName, '@') & '+test@' & $substringAfter($.userName, '@')}])",
					Operations: &AccessApplicationScimMappingOperations{
						Create: BoolPtr(true),
						Update: BoolPtr(true),
						Delete: BoolPtr(true),
					},
				},
			},
		},
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/apps", handler)

	actual, err := client.CreateAccessApplication(context.Background(), AccountIdentifier(testAccountID), CreateAccessApplicationParams{
		Name: "Admin Saas Site",
		SCIMConfig: &AccessApplicationSCIMConfig{
			Enabled:   BoolPtr(true),
			RemoteURI: "https://scim.com",
			Authentication: &AccessApplicationScimAuthenticationJson{
				Value: &AccessApplicationScimAuthenticationOauthBearerToken{
					Token:                  "1234567890",
					baseScimAuthentication: baseScimAuthentication{Scheme: AccessApplicationScimAuthenticationSchemeOauthBearerToken},
				},
			},
			IdPUID:             "1234567",
			DeactivateOnDelete: BoolPtr(true),
			Mappings: []*AccessApplicationScimMapping{
				{
					Schema:           "urn:ietf:params:scim:schemas:core:2.0:User",
					Enabled:          BoolPtr(true),
					Filter:           "title pr or userType eq \"Intern\"",
					TransformJsonata: "$merge([$, {'userName': $substringBefore($.userName, '@') & '+test@' & $substringAfter($.userName, '@')}])",
					Operations: &AccessApplicationScimMappingOperations{
						Create: BoolPtr(true),
						Update: BoolPtr(true),
						Delete: BoolPtr(true),
					},
				},
			},
		},
	})

	if assert.NoError(t, err) {
		assert.Equal(t, fullAccessApplication, actual)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/access/apps", handler)

	actual, err = client.CreateAccessApplication(context.Background(), ZoneIdentifier(testZoneID), CreateAccessApplicationParams{
		Name: "Admin SCIM Site",
		SCIMConfig: &AccessApplicationSCIMConfig{
			Enabled:   BoolPtr(true),
			RemoteURI: "https://scim.com",
			Authentication: &AccessApplicationScimAuthenticationJson{
				Value: &AccessApplicationScimAuthenticationOauthBearerToken{
					Token:                  "1234567890",
					baseScimAuthentication: baseScimAuthentication{Scheme: AccessApplicationScimAuthenticationSchemeOauthBearerToken},
				},
			},
			IdPUID:             "1234567",
			DeactivateOnDelete: BoolPtr(true),
			Mappings: []*AccessApplicationScimMapping{
				{
					Schema:           "urn:ietf:params:scim:schemas:core:2.0:User",
					Enabled:          BoolPtr(true),
					Filter:           "title pr or userType eq \"Intern\"",
					TransformJsonata: "$merge([$, {'userName': $substringBefore($.userName, '@') & '+test@' & $substringAfter($.userName, '@')}])",
					Operations: &AccessApplicationScimMappingOperations{
						Create: BoolPtr(true),
						Update: BoolPtr(true),
						Delete: BoolPtr(true),
					},
				},
			},
		},
	})

	if assert.NoError(t, err) {
		assert.Equal(t, fullAccessApplication, actual)
	}
}
