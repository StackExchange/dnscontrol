package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTeamsAccount(t *testing.T) {
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
				"id": "%s",
				"provider_name": "cf",
				"gateway_tag": "1234"
			}
		}
		`, testAccountID)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/gateway", handler)

	actual, err := client.TeamsAccount(context.Background(), testAccountID)

	if assert.NoError(t, err) {
		assert.Equal(t, actual.ProviderName, "cf")
		assert.Equal(t, actual.GatewayTag, "1234")
		assert.Equal(t, actual.ID, testAccountID)
	}
}

func TestTeamsAccountConfiguration(t *testing.T) {
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
				"settings": {
					"antivirus": {
					"enabled_download_phase": true,
						"notification_settings": {
					  		"enabled":true,
					   		"msg":"msg",
					   		"support_url":"https://hi.com"
						}
				 	},
					"tls_decrypt": {
						"enabled": true
					},
					"protocol_detection": {
						"enabled": true
					},
					"fips": {
						"tls": true
					},
					"activity_log": {
						"enabled": true
					},
					"block_page": {
						"enabled": true,
						"name": "Cloudflare",
						"footer_text": "--footer--",
						"header_text": "--header--",
						"mailto_address": "admin@example.com",
						"mailto_subject": "Blocked User Inquiry",
						"logo_path": "https://logos.com/a.png",
						"background_color": "#ff0000",
						"suppress_footer": true
					},
					"browser_isolation": {
						"url_browser_isolation_enabled": true,
						"non_identity_enabled": true
					},
					"body_scanning": {
						"inspection_mode": "deep"
					},
					"extended_email_matching": {
						"enabled": true
					}
				}
			}
		}
		`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/gateway/configuration", handler)

	actual, err := client.TeamsAccountConfiguration(context.Background(), testAccountID)

	if assert.NoError(t, err) {
		assert.Equal(t, actual.Settings, TeamsAccountSettings{
			Antivirus: &TeamsAntivirus{
				EnabledDownloadPhase: true,
				NotificationSettings: &TeamsNotificationSettings{
					Enabled:    &trueValue,
					Message:    "msg",
					SupportURL: "https://hi.com",
				},
			},
			ActivityLog:       &TeamsActivityLog{Enabled: true},
			TLSDecrypt:        &TeamsTLSDecrypt{Enabled: true},
			ProtocolDetection: &TeamsProtocolDetection{Enabled: true},
			FIPS:              &TeamsFIPS{TLS: true},
			BodyScanning:      &TeamsBodyScanning{InspectionMode: "deep"},

			BlockPage: &TeamsBlockPage{
				Enabled:         BoolPtr(true),
				FooterText:      "--footer--",
				HeaderText:      "--header--",
				LogoPath:        "https://logos.com/a.png",
				BackgroundColor: "#ff0000",
				Name:            "Cloudflare",
				MailtoAddress:   "admin@example.com",
				MailtoSubject:   "Blocked User Inquiry",
				SuppressFooter:  BoolPtr(true),
			},
			BrowserIsolation: &BrowserIsolation{
				UrlBrowserIsolationEnabled: BoolPtr(true),
				NonIdentityEnabled:         BoolPtr(true),
			},
			ExtendedEmailMatching: &TeamsExtendedEmailMatching{
				Enabled: BoolPtr(true),
			},
		})
	}
}

func TestTeamsAccountUpdateConfiguration(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'put', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"settings": {
					"antivirus": {
						"enabled_download_phase": false
					},
					"tls_decrypt": {
						"enabled": true
					},
					"activity_log": {
						"enabled": true
					},
					"protocol_detection": {
						"enabled": true
					},
					"extended_email_matching": {
						"enabled": true
					},
					"custom_certificate": {
						"enabled": true
					}
				}
			}
		}
		`)
	}

	settings := TeamsAccountSettings{
		Antivirus:         &TeamsAntivirus{EnabledDownloadPhase: false},
		ActivityLog:       &TeamsActivityLog{Enabled: true},
		TLSDecrypt:        &TeamsTLSDecrypt{Enabled: true},
		ProtocolDetection: &TeamsProtocolDetection{Enabled: true},
		ExtendedEmailMatching: &TeamsExtendedEmailMatching{
			Enabled: BoolPtr(true),
		},
		CustomCertificate: &TeamsCustomCertificate{
			Enabled: BoolPtr(true),
		},
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/gateway/configuration", handler)

	configuration := TeamsConfiguration{
		Settings: settings,
	}
	actual, err := client.TeamsAccountUpdateConfiguration(context.Background(), testAccountID, configuration)

	if assert.NoError(t, err) {
		assert.Equal(t, actual, configuration)
	}
}

func TestTeamsAccountGetLoggingConfiguration(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {"settings_by_rule_type":{"dns":{"log_all":false,"log_blocks":true}},"redact_pii":true}
		}`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/gateway/logging", handler)

	actual, err := client.TeamsAccountLoggingConfiguration(context.Background(), testAccountID)

	if assert.NoError(t, err) {
		assert.Equal(t, actual, TeamsLoggingSettings{
			RedactPii: true,
			LoggingSettingsByRuleType: map[TeamsRuleType]TeamsAccountLoggingConfiguration{
				TeamsDnsRuleType: {LogAll: false, LogBlocks: true},
			},
		})
	}
}

func TestTeamsAccountUpdateLoggingConfiguration(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {"settings_by_rule_type":{"dns":{"log_all":false,"log_blocks":true}, "http":{"log_all":true,"log_blocks":false}, "l4": {"log_all": false, "log_blocks": true}},"redact_pii":true}
		}`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/gateway/logging", handler)

	actual, err := client.TeamsAccountUpdateLoggingConfiguration(context.Background(), testAccountID, TeamsLoggingSettings{
		RedactPii: true,
		LoggingSettingsByRuleType: map[TeamsRuleType]TeamsAccountLoggingConfiguration{
			TeamsDnsRuleType: {
				LogAll:    false,
				LogBlocks: true,
			},
			TeamsHttpRuleType: {
				LogAll: true,
			},
			TeamsL4RuleType: {
				LogBlocks: true,
			},
		},
	})

	if assert.NoError(t, err) {
		assert.Equal(t, actual, TeamsLoggingSettings{
			RedactPii: true,
			LoggingSettingsByRuleType: map[TeamsRuleType]TeamsAccountLoggingConfiguration{
				TeamsDnsRuleType:  {LogAll: false, LogBlocks: true},
				TeamsHttpRuleType: {LogAll: true, LogBlocks: false},
				TeamsL4RuleType:   {LogAll: false, LogBlocks: true},
			},
		})
	}
}

func TestTeamsAccountGetDeviceConfiguration(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {"gateway_proxy_enabled": true,"gateway_udp_proxy_enabled":false, "root_certificate_installation_enabled":true, "use_zt_virtual_ip":false}
		}`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/devices/settings", handler)

	actual, err := client.TeamsAccountDeviceConfiguration(context.Background(), testAccountID)

	if assert.NoError(t, err) {
		assert.Equal(t, actual, TeamsDeviceSettings{
			GatewayProxyEnabled:                true,
			GatewayProxyUDPEnabled:             false,
			RootCertificateInstallationEnabled: true,
			UseZTVirtualIP:                     BoolPtr(false),
		})
	}
}

func TestTeamsAccountUpdateDeviceConfiguration(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {"gateway_proxy_enabled": true,"gateway_udp_proxy_enabled":true, "root_certificate_installation_enabled":true, "use_zt_virtual_ip":true}
		}`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/devices/settings", handler)

	actual, err := client.TeamsAccountDeviceUpdateConfiguration(context.Background(), testAccountID, TeamsDeviceSettings{
		GatewayProxyUDPEnabled:             true,
		GatewayProxyEnabled:                true,
		RootCertificateInstallationEnabled: true,
		UseZTVirtualIP:                     BoolPtr(true),
	})

	if assert.NoError(t, err) {
		assert.Equal(t, actual, TeamsDeviceSettings{
			GatewayProxyEnabled:                true,
			GatewayProxyUDPEnabled:             true,
			RootCertificateInstallationEnabled: true,
			UseZTVirtualIP:                     BoolPtr(true),
		})
	}
}

func TestTeamsAccountGetConnectivityConfiguration(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {"icmp_proxy_enabled": false,"offramp_warp_enabled":false}
		}`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/zerotrust/connectivity_settings", handler)

	actual, err := client.TeamsAccountConnectivityConfiguration(context.Background(), testAccountID)

	if assert.NoError(t, err) {
		assert.Equal(t, actual, TeamsConnectivitySettings{
			ICMPProxyEnabled:   BoolPtr(false),
			OfframpWARPEnabled: BoolPtr(false),
		})
	}
}

func TestTeamsAccountUpdateConnectivityConfiguration(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {"icmp_proxy_enabled": true,"offramp_warp_enabled":true}
		}`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/zerotrust/connectivity_settings", handler)

	actual, err := client.TeamsAccountConnectivityUpdateConfiguration(context.Background(), testAccountID, TeamsConnectivitySettings{
		ICMPProxyEnabled:   BoolPtr(true),
		OfframpWARPEnabled: BoolPtr(true),
	})

	if assert.NoError(t, err) {
		assert.Equal(t, actual, TeamsConnectivitySettings{
			ICMPProxyEnabled:   BoolPtr(true),
			OfframpWARPEnabled: BoolPtr(true),
		})
	}
}
