package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	deviceSettingsPolicyID         = "a842fa8a-a583-482e-9cd9-eb43362949fd"
	deviceSettingsPolicyMatch      = "identity.email == \"test@example.com\""
	deviceSettingsPolicyPrecedence = 10

	defaultDeviceSettingsPolicy = DeviceSettingsPolicy{
		ServiceModeV2: &ServiceModeV2{
			Mode: "warp",
		},
		DisableAutoFallback: BoolPtr(false),
		FallbackDomains: &[]FallbackDomain{
			{Suffix: "invalid"},
			{Suffix: "test"},
		},
		Exclude: &[]SplitTunnel{
			{Address: "10.0.0.0/8"},
			{Address: "100.64.0.0/10"},
		},
		GatewayUniqueID:    StringPtr("t1235"),
		SupportURL:         StringPtr(""),
		CaptivePortal:      IntPtr(180),
		AllowModeSwitch:    BoolPtr(false),
		SwitchLocked:       BoolPtr(false),
		AllowUpdates:       BoolPtr(false),
		AutoConnect:        IntPtr(0),
		AllowedToLeave:     BoolPtr(true),
		Enabled:            BoolPtr(true),
		PolicyID:           nil,
		Name:               nil,
		Match:              nil,
		Precedence:         nil,
		Default:            true,
		ExcludeOfficeIps:   BoolPtr(false),
		Description:        nil,
		LANAllowMinutes:    nil,
		LANAllowSubnetSize: nil,
	}

	nonDefaultDeviceSettingsPolicy = DeviceSettingsPolicy{
		ServiceModeV2: &ServiceModeV2{
			Mode: "warp",
		},
		DisableAutoFallback: BoolPtr(false),
		FallbackDomains: &[]FallbackDomain{
			{Suffix: "invalid"},
			{Suffix: "test"},
		},
		Exclude: &[]SplitTunnel{
			{Address: "10.0.0.0/8"},
			{Address: "100.64.0.0/10"},
		},
		GatewayUniqueID:    StringPtr("t1235"),
		SupportURL:         StringPtr(""),
		CaptivePortal:      IntPtr(180),
		AllowModeSwitch:    BoolPtr(false),
		SwitchLocked:       BoolPtr(false),
		AllowUpdates:       BoolPtr(false),
		AutoConnect:        IntPtr(0),
		AllowedToLeave:     BoolPtr(true),
		PolicyID:           &deviceSettingsPolicyID,
		Enabled:            BoolPtr(true),
		Name:               StringPtr("test"),
		Match:              &deviceSettingsPolicyMatch,
		Precedence:         &deviceSettingsPolicyPrecedence,
		Default:            false,
		ExcludeOfficeIps:   BoolPtr(true),
		Description:        StringPtr("Test Description"),
		LANAllowMinutes:    UintPtr(120),
		LANAllowSubnetSize: UintPtr(31),
	}

	defaultDeviceSettingsPolicyJson = `{
		"service_mode_v2": {
			"mode": "warp"
		},
		"disable_auto_fallback": false,
		"fallback_domains": [
			{
				"suffix": "invalid"
			},
			{
				"suffix": "test"
			}
		],
		"exclude": [
			{
				"address": "10.0.0.0/8"
			},
			{
				"address": "100.64.0.0/10"
			}
		],
		"gateway_unique_id": "t1235",
		"support_url": "",
		"captive_portal": 180,
		"allow_mode_switch": false,
		"switch_locked": false,
		"allow_updates": false,
		"auto_connect": 0,
		"allowed_to_leave": true,
		"enabled": true,
		"default": true,
		"exclude_office_ips":false
	}`

	nonDefaultDeviceSettingsPolicyJson = fmt.Sprintf(`{
		"service_mode_v2": {
			"mode": "warp"
		},
		"disable_auto_fallback": false,
		"fallback_domains": [
			{
				"suffix": "invalid"
			},
			{
				"suffix": "test"
			}
		],
		"exclude": [
			{
				"address": "10.0.0.0/8"
			},
			{
				"address": "100.64.0.0/10"
			}
		],
		"gateway_unique_id": "t1235",
		"support_url": "",
		"captive_portal": 180,
		"allow_mode_switch": false,
		"switch_locked": false,
		"allow_updates": false,
		"auto_connect": 0,
		"allowed_to_leave": true,
		"policy_id": "%s",
		"enabled": true,
		"name": "test",
		"match": %#v,
		"precedence": 10,
		"default": false,
		"exclude_office_ips":true,
		"description":"Test Description",
		"lan_allow_minutes": 120,
		"lan_allow_subnet_size": 31
	}`, deviceSettingsPolicyID, deviceSettingsPolicyMatch)
)

func TestUpdateDeviceClientCertificates(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": null,
			"messages": null,
			"result": {"enabled": true}
		}`)
	}

	want := DeviceClientCertificates{
		Response: Response{
			Success:  true,
			Errors:   nil,
			Messages: nil,
		},
		Result: Enabled{true},
	}

	mux.HandleFunc("/zones/"+testZoneID+"/devices/policy/certificates", handler)

	actual, err := client.UpdateDeviceClientCertificates(context.Background(), ZoneIdentifier(testZoneID), UpdateDeviceClientCertificatesParams{Enabled: BoolPtr(true)})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestGetDeviceClientCertificates(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": null,
			"messages": null,
			"result": {"enabled": false}
		}`)
	}

	want := DeviceClientCertificates{
		Response: Response{
			Success:  true,
			Errors:   nil,
			Messages: nil,
		},
		Result: Enabled{false},
	}

	mux.HandleFunc("/zones/"+testZoneID+"/devices/policy/certificates", handler)

	actual, err := client.GetDeviceClientCertificates(context.Background(), ZoneIdentifier(testZoneID), GetDeviceClientCertificatesParams{})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateDeviceSettingsPolicy(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": null,
			"messages": null,
			"result": %s
		}`, nonDefaultDeviceSettingsPolicyJson)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/devices/policy", handler)

	actual, err := client.CreateDeviceSettingsPolicy(context.Background(), AccountIdentifier(testAccountID), CreateDeviceSettingsPolicyParams{
		Precedence:         IntPtr(10),
		Match:              &deviceSettingsPolicyMatch,
		Name:               StringPtr("test"),
		Description:        StringPtr("Test Description"),
		LANAllowMinutes:    UintPtr(120),
		LANAllowSubnetSize: UintPtr(31),
	})

	if assert.NoError(t, err) {
		assert.Equal(t, nonDefaultDeviceSettingsPolicy, actual)
	}
}

func TestUpdateDefaultDeviceSettingsPolicy(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": null,
			"messages": null,
			"result": %s
		}`, defaultDeviceSettingsPolicyJson)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/devices/policy", handler)

	actual, err := client.UpdateDefaultDeviceSettingsPolicy(context.Background(), AccountIdentifier(testAccountID), UpdateDefaultDeviceSettingsPolicyParams{})

	if assert.NoError(t, err) {
		assert.Equal(t, defaultDeviceSettingsPolicy, actual)
	}
}

func TestUpdateDeviceSettingsPolicy(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": null,
			"messages": null,
			"result": %s
		}`, nonDefaultDeviceSettingsPolicyJson)
	}

	precedence := 10

	mux.HandleFunc("/accounts/"+testAccountID+"/devices/policy/"+deviceSettingsPolicyID, handler)

	actual, err := client.UpdateDeviceSettingsPolicy(context.Background(), AccountIdentifier(testAccountID), UpdateDeviceSettingsPolicyParams{
		PolicyID:   &deviceSettingsPolicyID,
		Precedence: &precedence,
	})

	if assert.NoError(t, err) {
		assert.Equal(t, nonDefaultDeviceSettingsPolicy, actual)
	}
}

func TestDeleteDeviceSettingsPolicy(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": null,
			"messages": null,
			"result": [ %s ]
		}`, defaultDeviceSettingsPolicyJson)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/devices/policy/"+deviceSettingsPolicyID, handler)

	actual, err := client.DeleteDeviceSettingsPolicy(context.Background(), AccountIdentifier(testAccountID), deviceSettingsPolicyID)

	if assert.NoError(t, err) {
		assert.Equal(t, []DeviceSettingsPolicy{defaultDeviceSettingsPolicy}, actual)
	}
}

func TestGetDefaultDeviceSettings(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": null,
			"messages": null,
			"result": %s
		}`, defaultDeviceSettingsPolicyJson)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/devices/policy", handler)

	actual, err := client.GetDefaultDeviceSettingsPolicy(context.Background(), AccountIdentifier(testAccountID), GetDefaultDeviceSettingsPolicyParams{})

	if assert.NoError(t, err) {
		assert.Equal(t, defaultDeviceSettingsPolicy, actual)
	}
}

func TestGetDeviceSettings(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": null,
			"messages": null,
			"result": %s
		}`, nonDefaultDeviceSettingsPolicyJson)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/devices/policy/"+deviceSettingsPolicyID, handler)

	actual, err := client.GetDeviceSettingsPolicy(context.Background(), AccountIdentifier(testAccountID), GetDeviceSettingsPolicyParams{PolicyID: &deviceSettingsPolicyID})

	if assert.NoError(t, err) {
		assert.Equal(t, nonDefaultDeviceSettingsPolicy, actual)
	}
}

func TestListDeviceSettingsPolicies(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": null,
			"messages": null,
			"result": [%s],
			"result_info": {
  				"count": 1,
  				"page": 1,
  				"per_page": 20,
  				"total_count": 1
			}
		}`, nonDefaultDeviceSettingsPolicyJson)
	}

	want := []DeviceSettingsPolicy{nonDefaultDeviceSettingsPolicy}

	mux.HandleFunc("/accounts/"+testAccountID+"/devices/policies", handler)

	actual, resultInfo, err := client.ListDeviceSettingsPolicies(context.Background(), AccountIdentifier(testAccountID), ListDeviceSettingsPoliciesParams{
		ResultInfo: ResultInfo{
			Page:    1,
			PerPage: 20,
		},
	})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
		assert.Equal(t, &ResultInfo{
			Count:   1,
			Page:    1,
			PerPage: 20,
			Total:   1,
		}, resultInfo)
		assert.Len(t, actual, 1)
	}
}
