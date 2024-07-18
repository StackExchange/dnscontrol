package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTeamsRules(t *testing.T) {
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
				  "id": "7559a944-3dd7-41bf-b183-360a814a8c36",
				  "name": "rule1",
				  "description": "rule description",
				  "precedence": 1000,
				  "enabled": false,
				  "action": "isolate",
				  "filters": [
					"http"
				  ],
				  "created_at": "2014-01-01T05:20:00.12345Z",
				  "updated_at": "2014-01-01T05:20:00.12345Z",
				  "deleted_at": null,
				  "traffic": "http.host == \"example.com\"",
				  "identity": "",
				  "version": 1,
				  "rule_settings": {
					"block_page_enabled": false,
					"block_reason": "",
					"override_ips": null,
					"override_host": "",
					"l4override": null,
					"biso_admin_controls": null,
					"add_headers": null,
					"check_session": {
						"enforce": true,
						"duration": "15m0s"
					},
                    "insecure_disable_dnssec_validation": false,
					"untrusted_cert": {
						"action": "error"
					},
					"dns_resolvers": {
						"ipv4": [
							{"ip": "10.0.0.2", "port": 5053},
							{
								"ip": "192.168.0.2",
								"vnet_id": "16fd7a32-11f0-4687-a0bb-7031d241e184",
								"route_through_private_network": true
							}
						],
						"ipv6": [
							{"ip": "2460::1"}
						]
					},
					"notification_settings": {
						"enabled": true,
						"msg": "message",
						"support_url": "https://hello.com"
					}
				  }
				},
				{
				  "id": "9ae57318-f32e-46b3-b889-48dd6dcc49af",
				  "name": "rule2",
				  "description": "rule description 2",
				  "precedence": 2000,
				  "enabled": true,
				  "action": "block",
				  "filters": [
					"http"
				  ],
				  "created_at": "2014-01-01T05:20:00.12345Z",
				  "updated_at": "2014-01-01T05:20:00.12345Z",
				  "deleted_at": null,
				  "traffic": "http.host == \"abcd.com\"",
				  "identity": "",
				  "version": 1,
				  "rule_settings": {
					"block_page_enabled": true,
					"block_reason": "",
					"override_ips": null,
					"override_host": "",
					"l4override": null,
					"biso_admin_controls": null,
					"add_headers": null,
					"check_session": null,
                    "insecure_disable_dnssec_validation": true,
					"untrusted_cert": {
						"action": "pass_through"
					},
					"resolve_dns_through_cloudflare": true
				  }
				}
			]
		}
		`)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")

	want := []TeamsRule{{
		ID:            "7559a944-3dd7-41bf-b183-360a814a8c36",
		Name:          "rule1",
		Description:   "rule description",
		Precedence:    1000,
		Enabled:       false,
		Action:        Isolate,
		Filters:       []TeamsFilterType{HttpFilter},
		Traffic:       `http.host == "example.com"`,
		DevicePosture: "",
		Identity:      "",
		Version:       1,
		RuleSettings: TeamsRuleSettings{
			BlockPageEnabled:  false,
			BlockReason:       "",
			OverrideIPs:       nil,
			OverrideHost:      "",
			L4Override:        nil,
			AddHeaders:        nil,
			BISOAdminControls: nil,
			CheckSession: &TeamsCheckSessionSettings{
				Enforce:  true,
				Duration: Duration{900 * time.Second},
			},
			InsecureDisableDNSSECValidation: false,
			UntrustedCertSettings: &UntrustedCertSettings{
				Action: UntrustedCertError,
			},
			DnsResolverSettings: &TeamsDnsResolverSettings{
				V4Resolvers: []TeamsDnsResolverAddressV4{
					{
						TeamsDnsResolverAddress{
							IP:   "10.0.0.2",
							Port: IntPtr(5053),
						},
					},
					{
						TeamsDnsResolverAddress{
							IP:                         "192.168.0.2",
							VnetID:                     "16fd7a32-11f0-4687-a0bb-7031d241e184",
							RouteThroughPrivateNetwork: BoolPtr(true),
						},
					},
				},
				V6Resolvers: []TeamsDnsResolverAddressV6{
					{
						TeamsDnsResolverAddress{
							IP: "2460::1",
						},
					},
				},
			},
			NotificationSettings: &TeamsNotificationSettings{
				Enabled:    BoolPtr(true),
				Message:    "message",
				SupportURL: "https://hello.com",
			},
		},
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
		DeletedAt: nil,
	},
		{
			ID:            "9ae57318-f32e-46b3-b889-48dd6dcc49af",
			Name:          "rule2",
			Description:   "rule description 2",
			Precedence:    2000,
			Enabled:       true,
			Action:        Block,
			Filters:       []TeamsFilterType{HttpFilter},
			Traffic:       `http.host == "abcd.com"`,
			Identity:      "",
			DevicePosture: "",
			Version:       1,
			RuleSettings: TeamsRuleSettings{
				BlockPageEnabled:  true,
				BlockReason:       "",
				OverrideIPs:       nil,
				OverrideHost:      "",
				L4Override:        nil,
				AddHeaders:        nil,
				BISOAdminControls: nil,
				CheckSession:      nil,
				// setting is invalid for block rules, just testing serialization here
				InsecureDisableDNSSECValidation: true,
				UntrustedCertSettings: &UntrustedCertSettings{
					Action: UntrustedCertPassthrough,
				},
				ResolveDnsThroughCloudflare: BoolPtr(true),
			},
			CreatedAt: &createdAt,
			UpdatedAt: &updatedAt,
			DeletedAt: nil,
		}}

	mux.HandleFunc("/accounts/"+testAccountID+"/gateway/rules", handler)

	actual, err := client.TeamsRules(context.Background(), testAccountID)

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestTeamsRule(t *testing.T) {
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
				"id": "7559a944-3dd7-41bf-b183-360a814a8c36",
				"name": "rule1",
				"description": "rule description",
				"precedence": 1000,
				"enabled": false,
				"action": "isolate",
				"filters": [
					"http"
				],
				"created_at": "2014-01-01T05:20:00.12345Z",
				"updated_at": "2014-01-01T05:20:00.12345Z",
				"deleted_at": null,
				"traffic": "http.host == \"abcd.com\"",
				"identity": "",
				"version": 1,
				"rule_settings": {
					"block_page_enabled": false,
					"block_reason": "",
					"override_ips": null,
					"override_host": "",
					"l4override": null,
					"biso_admin_controls": null,
					"add_headers": null,
					"check_session": {
						"enforce": true,
						"duration": "15m0s"
					},
                    "insecure_disable_dnssec_validation": false,
					"untrusted_cert": {
						"action": "block"
					}
				}
			}
		}
		`)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")

	want := TeamsRule{
		ID:            "7559a944-3dd7-41bf-b183-360a814a8c36",
		Name:          "rule1",
		Description:   "rule description",
		Precedence:    1000,
		Enabled:       false,
		Action:        Isolate,
		Filters:       []TeamsFilterType{HttpFilter},
		Traffic:       `http.host == "abcd.com"`,
		Identity:      "",
		DevicePosture: "",
		Version:       1,
		RuleSettings: TeamsRuleSettings{
			BlockPageEnabled:  false,
			BlockReason:       "",
			OverrideIPs:       nil,
			OverrideHost:      "",
			L4Override:        nil,
			AddHeaders:        nil,
			BISOAdminControls: nil,
			CheckSession: &TeamsCheckSessionSettings{
				Enforce:  true,
				Duration: Duration{900 * time.Second},
			},
			InsecureDisableDNSSECValidation: false,
			UntrustedCertSettings: &UntrustedCertSettings{
				Action: UntrustedCertBlock,
			},
		},
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
		DeletedAt: nil,
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/gateway/rules/7559a9443dd741bfb183360a814a8c36", handler)

	actual, err := client.TeamsRule(context.Background(), testAccountID, "7559a9443dd741bfb183360a814a8c36")

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestTeamsCreateHTTPRule(t *testing.T) {
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
				"name": "rule1",
				"description": "rule description",
				"precedence": 1000,
				"enabled": false,
				"action": "isolate",
				"filters": [
					"http"
				],
				"traffic": "http.host == \"abcd.com\"",
				"identity": "",
				"rule_settings": {
					"block_page_enabled": false,
					"biso_admin_controls": {"dp": true, "du": true, "dk": true},
					"add_headers": {
						"X-Test": ["abcd"]
					},
					"check_session": {
						"enforce": true,
						"duration": "5m0s"
					},
                    "insecure_disable_dnssec_validation": false
				}
			}
		}
		`)
	}

	want := TeamsRule{
		Name:          "rule1",
		Description:   "rule description",
		Precedence:    1000,
		Enabled:       false,
		Action:        Isolate,
		Filters:       []TeamsFilterType{HttpFilter},
		Traffic:       `http.host == "abcd.com"`,
		Identity:      "",
		DevicePosture: "",
		RuleSettings: TeamsRuleSettings{
			BlockPageEnabled: false,
			BlockReason:      "",
			OverrideIPs:      nil,
			OverrideHost:     "",
			L4Override:       nil,
			AddHeaders:       http.Header{"X-Test": []string{"abcd"}},
			BISOAdminControls: &TeamsBISOAdminControlSettings{
				DisablePrinting: true,
				DisableKeyboard: true,
				DisableUpload:   true,
			},
			CheckSession: &TeamsCheckSessionSettings{
				Enforce:  true,
				Duration: Duration{300 * time.Second},
			},
			InsecureDisableDNSSECValidation: false,
			EgressSettings:                  nil,
		},
		DeletedAt: nil,
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/gateway/rules", handler)

	actual, err := client.TeamsCreateRule(context.Background(), testAccountID, want)

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestTeamsCreateEgressRule(t *testing.T) {
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
				"name": "egress via chicago",
				"description": "rule description",
				"precedence": 1000,
				"enabled": false,
				"action": "egress",
				"filters": [
					"egress"
				],
				"traffic": "net.src.geo.country == \"US\"",
				"identity": "",
				"rule_settings": {
					"egress": {
						"ipv6": "2a06:98c1:54::c61/64",
						"ipv4": "2.2.2.2",
						"ipv4_fallback": "1.1.1.1"
					}
				}
			}
		}
		`)
	}

	want := TeamsRule{
		Name:          "egress via chicago",
		Description:   "rule description",
		Precedence:    1000,
		Enabled:       false,
		Action:        Egress,
		Filters:       []TeamsFilterType{EgressFilter},
		Traffic:       `net.src.geo.country == "US"`,
		Identity:      "",
		DevicePosture: "",
		RuleSettings: TeamsRuleSettings{
			BlockPageEnabled:                false,
			BlockReason:                     "",
			OverrideIPs:                     nil,
			OverrideHost:                    "",
			L4Override:                      nil,
			AddHeaders:                      nil,
			BISOAdminControls:               nil,
			CheckSession:                    nil,
			InsecureDisableDNSSECValidation: false,
			EgressSettings: &EgressSettings{
				Ipv6Range:    "2a06:98c1:54::c61/64",
				Ipv4:         "2.2.2.2",
				Ipv4Fallback: "1.1.1.1",
			},
		},
		DeletedAt: nil,
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/gateway/rules", handler)

	actual, err := client.TeamsCreateRule(context.Background(), testAccountID, want)

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestTeamsCreateL4Rule(t *testing.T) {
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
				"name": "block 4.4.4.4",
				"description": "rule description",
				"precedence": 1000,
				"enabled": true,
				"action": "audit_ssh",
				"filters": [
					"l4"
				],
				"traffic": "net.src.geo.country == \"US\"",
				"identity": "",
				"rule_settings": {
					"audit_ssh": { "command_logging": true }
				}
			}
		}
		`)
	}

	want := TeamsRule{
		Name:          "block 4.4.4.4",
		Description:   "rule description",
		Precedence:    1000,
		Enabled:       true,
		Action:        AuditSSH,
		Filters:       []TeamsFilterType{L4Filter},
		Traffic:       `net.src.geo.country == "US"`,
		Identity:      "",
		DevicePosture: "",
		RuleSettings: TeamsRuleSettings{
			BlockPageEnabled:                false,
			BlockReason:                     "",
			OverrideIPs:                     nil,
			OverrideHost:                    "",
			L4Override:                      nil,
			AddHeaders:                      nil,
			BISOAdminControls:               nil,
			CheckSession:                    nil,
			InsecureDisableDNSSECValidation: false,
			EgressSettings:                  nil,
			AuditSSH: &AuditSSHRuleSettings{
				CommandLogging: true,
			},
		},
		DeletedAt: nil,
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/gateway/rules", handler)

	actual, err := client.TeamsCreateRule(context.Background(), testAccountID, want)

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestTeamsCreateResolverPolicy(t *testing.T) {
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
				"name": "resolve 4.4.4.4",
				"description": "rule description",
				"precedence": 1000,
				"enabled": true,
				"action": "resolve",
				"filters": [
					"dns_resolver"
				],
				"traffic": "any(dns.domains[*] == \"scottstots.com\")",
				"identity": "",
				"rule_settings": {
					"audit_ssh": { "command_logging": true },
					"resolve_dns_through_cloudflare": true
				}
			}
		}
		`)
	}

	want := TeamsRule{
		Name:          "resolve 4.4.4.4",
		Description:   "rule description",
		Precedence:    1000,
		Enabled:       true,
		Action:        Resolve,
		Filters:       []TeamsFilterType{DnsResolverFilter},
		Traffic:       `any(dns.domains[*] == "scottstots.com")`,
		Identity:      "",
		DevicePosture: "",
		RuleSettings: TeamsRuleSettings{
			BlockPageEnabled:                false,
			BlockReason:                     "",
			OverrideIPs:                     nil,
			OverrideHost:                    "",
			L4Override:                      nil,
			AddHeaders:                      nil,
			BISOAdminControls:               nil,
			CheckSession:                    nil,
			InsecureDisableDNSSECValidation: false,
			EgressSettings:                  nil,
			AuditSSH: &AuditSSHRuleSettings{
				CommandLogging: true,
			},
			ResolveDnsThroughCloudflare: BoolPtr(true),
		},
		DeletedAt: nil,
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/gateway/rules", handler)

	actual, err := client.TeamsCreateRule(context.Background(), testAccountID, want)

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestTeamsUpdateRule(t *testing.T) {
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
				"id": "7559a944-3dd7-41bf-b183-360a814a8c36",
				"name": "rule_name_change",
				"description": "rule new description",
				"precedence": 3000,
				"enabled": true,
				"action": "block",
				"filters": [
					"http"
				],
				"traffic": "",
				"identity": "",
				"created_at": "2014-01-01T05:20:00.12345Z",
				"updated_at": "2014-02-01T05:20:00.12345Z",
				"rule_settings": {
					"block_page_enabled": false,
					"block_reason": "",
					"override_ips": null,
					"override_host": "",
					"l4override": null,
					"biso_admin_controls": null,
					"add_headers": null,
					"check_session": null,
                    "insecure_disable_dnssec_validation": false
				}
			}
		}
		`)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2014-02-01T05:20:00.12345Z")

	want := TeamsRule{
		ID:            "7559a944-3dd7-41bf-b183-360a814a8c36",
		Name:          "rule_name_change",
		Description:   "rule new description",
		Precedence:    3000,
		Enabled:       true,
		Action:        Block,
		Filters:       []TeamsFilterType{HttpFilter},
		Traffic:       "",
		Identity:      "",
		DevicePosture: "",
		RuleSettings: TeamsRuleSettings{
			BlockPageEnabled:                false,
			BlockReason:                     "",
			OverrideIPs:                     nil,
			OverrideHost:                    "",
			L4Override:                      nil,
			AddHeaders:                      nil,
			BISOAdminControls:               nil,
			CheckSession:                    nil,
			InsecureDisableDNSSECValidation: false,
		},
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
		DeletedAt: nil,
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/gateway/rules/7559a9443dd741bfb183360a814a8c36", handler)

	actual, err := client.TeamsUpdateRule(context.Background(), testAccountID, "7559a9443dd741bfb183360a814a8c36", want)

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestTeamsPatchRule(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method 'Patch', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"name": "rule_name_change",
				"description": "rule new description",
				"precedence": 3000,
				"enabled": true,
				"action": "block",
				"rule_settings": {
					"block_page_enabled": false,
					"block_reason": "",
					"override_ips": null,
					"override_host": "",
					"l4override": null,
					"biso_admin_controls": null,
					"add_headers": null,
					"check_session": null,
                    "insecure_disable_dnssec_validation": false
				}
			}
		}
		`)
	}

	reqBody := TeamsRulePatchRequest{
		Name:        "rule_name_change",
		Description: "rule new description",
		Precedence:  3000,
		Enabled:     true,
		Action:      Block,
		RuleSettings: TeamsRuleSettings{
			BlockPageEnabled:                false,
			BlockReason:                     "",
			OverrideIPs:                     nil,
			OverrideHost:                    "",
			L4Override:                      nil,
			AddHeaders:                      nil,
			BISOAdminControls:               nil,
			CheckSession:                    nil,
			InsecureDisableDNSSECValidation: false,
		},
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/gateway/rules/7559a9443dd741bfb183360a814a8c36", handler)

	actual, err := client.TeamsPatchRule(context.Background(), testAccountID, "7559a9443dd741bfb183360a814a8c36", reqBody)

	if assert.NoError(t, err) {
		assert.Equal(t, actual.Name, "rule_name_change")
		assert.Equal(t, actual.Description, "rule new description")
		assert.Equal(t, actual.Precedence, uint64(3000))
		assert.Equal(t, actual.Action, Block)
		assert.Equal(t, actual.Enabled, true)
	}
}

func TestTeamsDeleteRule(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'Delete', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": nil
		}
		`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/gateway/rules/7559a9443dd741bfb183360a814a8c36", handler)

	err := client.TeamsDeleteRule(context.Background(), testAccountID, "7559a9443dd741bfb183360a814a8c36")

	assert.NoError(t, err)
}
