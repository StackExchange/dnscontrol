package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestListRulesets(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
      "result": [
        {
          "id": "2c0fc9fa937b11eaa1b71c4d701ab86e",
          "name": "my example ruleset",
          "description": "Test magic transit ruleset",
          "kind": "root",
          "version": "1",
          "last_updated": "2020-12-02T20:24:07.776073Z",
          "phase": "magic_transit"
        }
      ],
      "success": true,
      "errors": [],
      "messages": []
    }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/rulesets", handler)
	mux.HandleFunc("/zones/"+testZoneID+"/rulesets", handler)

	lastUpdated, _ := time.Parse(time.RFC3339, "2020-12-02T20:24:07.776073Z")

	want := []Ruleset{
		{
			ID:          "2c0fc9fa937b11eaa1b71c4d701ab86e",
			Name:        "my example ruleset",
			Description: "Test magic transit ruleset",
			Kind:        "root",
			Version:     StringPtr("1"),
			LastUpdated: &lastUpdated,
			Phase:       string(RulesetPhaseMagicTransit),
		},
	}

	zoneActual, err := client.ListRulesets(context.Background(), ZoneIdentifier(testZoneID), ListRulesetsParams{})
	if assert.NoError(t, err) {
		assert.Equal(t, want, zoneActual)
	}

	accountActual, err := client.ListRulesets(context.Background(), AccountIdentifier(testAccountID), ListRulesetsParams{})
	if assert.NoError(t, err) {
		assert.Equal(t, want, accountActual)
	}
}

func TestGetRuleset_MagicTransit(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
      "result": {
        "id": "2c0fc9fa937b11eaa1b71c4d701ab86e",
        "name": "my example ruleset",
        "description": "Test magic transit ruleset",
        "kind": "root",
        "version": "1",
        "last_updated": "2020-12-02T20:24:07.776073Z",
        "phase": "magic_transit"
      },
      "success": true,
      "errors": [],
      "messages": []
    }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/rulesets/2c0fc9fa937b11eaa1b71c4d701ab86e", handler)
	mux.HandleFunc("/zones/"+testZoneID+"/rulesets/2c0fc9fa937b11eaa1b71c4d701ab86e", handler)

	lastUpdated, _ := time.Parse(time.RFC3339, "2020-12-02T20:24:07.776073Z")

	want := Ruleset{
		ID:          "2c0fc9fa937b11eaa1b71c4d701ab86e",
		Name:        "my example ruleset",
		Description: "Test magic transit ruleset",
		Kind:        "root",
		Version:     StringPtr("1"),
		LastUpdated: &lastUpdated,
		Phase:       string(RulesetPhaseMagicTransit),
	}

	zoneActual, err := client.GetRuleset(context.Background(), ZoneIdentifier(testZoneID), "2c0fc9fa937b11eaa1b71c4d701ab86e")
	if assert.NoError(t, err) {
		assert.Equal(t, want, zoneActual)
	}

	accountActual, err := client.GetRuleset(context.Background(), AccountIdentifier(testAccountID), "2c0fc9fa937b11eaa1b71c4d701ab86e")
	if assert.NoError(t, err) {
		assert.Equal(t, want, accountActual)
	}
}

func TestGetRuleset_WAF(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
      "result": {
        "id": "70339d97bdb34195bbf054b1ebe81f76",
        "name": "Cloudflare Normalization Ruleset",
        "description": "Created by the Cloudflare security team, this ruleset provides normalization on the URL path",
        "kind": "managed",
        "version": "1",
        "rules": [
          {
            "id": "78723a9e0c7c4c6dbec5684cb766231d",
            "version": "1",
            "action": "rewrite",
            "action_parameters": {
              "uri": {
                "path": {
                  "expression": "normalize_url_path(raw.http.request.uri.path)"
                },
                "origin": false
              }
            },
            "description": "Normalization on the URL path, without propagating it to the origin",
            "last_updated": "2020-12-18T09:28:09.655749Z",
            "ref": "272936dc447b41fe976255ff6b768ec0",
            "enabled": true
          }
        ],
        "last_updated": "2020-12-18T09:28:09.655749Z",
        "phase": "http_request_sanitize"
      },
      "success": true,
      "errors": [],
      "messages": []
    }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/rulesets/b232b534beea4e00a21dcbb7a8a545e9", handler)
	mux.HandleFunc("/zones/"+testZoneID+"/rulesets/b232b534beea4e00a21dcbb7a8a545e9", handler)

	lastUpdated, _ := time.Parse(time.RFC3339, "2020-12-18T09:28:09.655749Z")

	rules := []RulesetRule{{
		ID:      "78723a9e0c7c4c6dbec5684cb766231d",
		Version: StringPtr("1"),
		Action:  string(RulesetRuleActionRewrite),
		ActionParameters: &RulesetRuleActionParameters{
			URI: &RulesetRuleActionParametersURI{
				Path: &RulesetRuleActionParametersURIPath{
					Expression: "normalize_url_path(raw.http.request.uri.path)",
				},
				Origin: BoolPtr(false),
			},
		},
		Description: "Normalization on the URL path, without propagating it to the origin",
		LastUpdated: &lastUpdated,
		Ref:         "272936dc447b41fe976255ff6b768ec0",
		Enabled:     BoolPtr(true),
	}}

	want := Ruleset{
		ID:          "70339d97bdb34195bbf054b1ebe81f76",
		Name:        "Cloudflare Normalization Ruleset",
		Description: "Created by the Cloudflare security team, this ruleset provides normalization on the URL path",
		Kind:        string(RulesetKindManaged),
		Version:     StringPtr("1"),
		LastUpdated: &lastUpdated,
		Phase:       string(RulesetPhaseHTTPRequestSanitize),
		Rules:       rules,
	}

	zoneActual, err := client.GetRuleset(context.Background(), ZoneIdentifier(testZoneID), "b232b534beea4e00a21dcbb7a8a545e9")
	if assert.NoError(t, err) {
		assert.Equal(t, want, zoneActual)
	}

	accountActual, err := client.GetRuleset(context.Background(), AccountIdentifier(testAccountID), "b232b534beea4e00a21dcbb7a8a545e9")
	if assert.NoError(t, err) {
		assert.Equal(t, want, accountActual)
	}
}

func TestGetRuleset_SetCacheSettings(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
      "result": {
        "id": "70339d97bdb34195bbf054b1ebe81f76",
        "name": "Cloudflare Cache Rules Ruleset",
        "description": "This ruleset provides cache settings modifications",
        "kind": "zone",
        "version": "1",
        "rules": [
          {
            "id": "78723a9e0c7c4c6dbec5684cb766231d",
            "version": "1",
            "action": "set_cache_settings",
            "action_parameters": {
				"cache": true,
				"edge_ttl":{"mode":"respect_origin","default":60,"status_code_ttl":[{"status_code":200,"value":30},{"status_code_range":{"from":201,"to":300},"value":20}]},
				"browser_ttl":{"mode":"override_origin","default":10},
				"serve_stale":{"disable_stale_while_updating":true},
				"respect_strong_etags":true,
				"cache_key":{
					"cache_deception_armor":true,
					"ignore_query_strings_order":true,
					"custom_key": {
						"query_string":{"include":"*"},
						"header":{"include":["habc","hdef"],"check_presence":["hfizz","hbuzz"],"exclude_origin":true},
						"cookie":{"include":["cabc","cdef"],"check_presence":["cfizz","cbuzz"]},
						"user":{
							"device_type":true,
							"geo":true,
							"lang":true
						},
						"host":{
							"resolved":true
						}
					}
				},
				"additional_cacheable_ports": [1,2,3,4],
				"origin_cache_control": true,
				"read_timeout": 1000,
				"origin_error_page_passthru":true,
				"cache_reserve": {
					"eligible": true,
					"minimum_file_size": 1000
				}
			},
			"description": "Set all available cache settings in one rule",
			"last_updated": "2020-12-18T09:28:09.655749Z",
			"ref": "272936dc447b41fe976255ff6b768ec0",
			"enabled": true
          }
        ],
        "last_updated": "2020-12-18T09:28:09.655749Z",
        "phase": "http_request_cache_settings"
      },
      "success": true,
      "errors": [],
      "messages": []
    }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/rulesets/b232b534beea4e00a21dcbb7a8a545e9", handler)
	mux.HandleFunc("/zones/"+testZoneID+"/rulesets/b232b534beea4e00a21dcbb7a8a545e9", handler)

	lastUpdated, _ := time.Parse(time.RFC3339, "2020-12-18T09:28:09.655749Z")

	rules := []RulesetRule{{
		ID:      "78723a9e0c7c4c6dbec5684cb766231d",
		Version: StringPtr("1"),
		Action:  string(RulesetRuleActionSetCacheSettings),
		ActionParameters: &RulesetRuleActionParameters{
			Cache: BoolPtr(true),
			EdgeTTL: &RulesetRuleActionParametersEdgeTTL{
				Mode:    "respect_origin",
				Default: UintPtr(60),
				StatusCodeTTL: []RulesetRuleActionParametersStatusCodeTTL{
					{
						StatusCodeValue: UintPtr(200),
						Value:           IntPtr(30),
					},
					{
						StatusCodeRange: &RulesetRuleActionParametersStatusCodeRange{
							From: UintPtr(201),
							To:   UintPtr(300),
						},
						Value: IntPtr(20),
					},
				},
			},
			BrowserTTL: &RulesetRuleActionParametersBrowserTTL{
				Mode:    "override_origin",
				Default: UintPtr(10),
			},
			ServeStale: &RulesetRuleActionParametersServeStale{
				DisableStaleWhileUpdating: BoolPtr(true),
			},
			RespectStrongETags: BoolPtr(true),
			CacheKey: &RulesetRuleActionParametersCacheKey{
				IgnoreQueryStringsOrder: BoolPtr(true),
				CacheDeceptionArmor:     BoolPtr(true),
				CustomKey: &RulesetRuleActionParametersCustomKey{
					Query: &RulesetRuleActionParametersCustomKeyQuery{
						Include: &RulesetRuleActionParametersCustomKeyList{
							All: true,
						},
					},
					Header: &RulesetRuleActionParametersCustomKeyHeader{
						RulesetRuleActionParametersCustomKeyFields: RulesetRuleActionParametersCustomKeyFields{
							Include:       []string{"habc", "hdef"},
							CheckPresence: []string{"hfizz", "hbuzz"},
						},
						ExcludeOrigin: BoolPtr(true),
					},
					Cookie: &RulesetRuleActionParametersCustomKeyCookie{
						Include:       []string{"cabc", "cdef"},
						CheckPresence: []string{"cfizz", "cbuzz"},
					},
					User: &RulesetRuleActionParametersCustomKeyUser{
						DeviceType: BoolPtr(true),
						Geo:        BoolPtr(true),
						Lang:       BoolPtr(true),
					},
					Host: &RulesetRuleActionParametersCustomKeyHost{
						Resolved: BoolPtr(true),
					},
				},
			},
			AdditionalCacheablePorts: []int{1, 2, 3, 4},
			OriginCacheControl:       BoolPtr(true),
			ReadTimeout:              UintPtr(1000),
			OriginErrorPagePassthru:  BoolPtr(true),
			CacheReserve: &RulesetRuleActionParametersCacheReserve{
				Eligible:        BoolPtr(true),
				MinimumFileSize: UintPtr(1000),
			},
		},
		Description: "Set all available cache settings in one rule",
		LastUpdated: &lastUpdated,
		Ref:         "272936dc447b41fe976255ff6b768ec0",
		Enabled:     BoolPtr(true),
	}}

	want := Ruleset{
		ID:          "70339d97bdb34195bbf054b1ebe81f76",
		Name:        "Cloudflare Cache Rules Ruleset",
		Description: "This ruleset provides cache settings modifications",
		Kind:        string(RulesetKindZone),
		Version:     StringPtr("1"),
		LastUpdated: &lastUpdated,
		Phase:       string(RulesetPhaseHTTPRequestCacheSettings),
		Rules:       rules,
	}

	zoneActual, err := client.GetRuleset(context.Background(), ZoneIdentifier(testZoneID), "b232b534beea4e00a21dcbb7a8a545e9")
	if assert.NoError(t, err) {
		assert.Equal(t, want, zoneActual)
	}

	accountActual, err := client.GetRuleset(context.Background(), AccountIdentifier(testAccountID), "b232b534beea4e00a21dcbb7a8a545e9")
	if assert.NoError(t, err) {
		assert.Equal(t, want, accountActual)
	}
}

func TestGetRuleset_SetConfig(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
      "result": {
        "id": "70339d97bdb34195bbf054b1ebe81f76",
        "name": "Cloudflare Config Rules Ruleset",
        "description": "This ruleset provides config rules modifications",
        "kind": "zone",
        "version": "1",
        "rules": [
          {
            "id": "78723a9e0c7c4c6dbec5684cb766231d",
            "version": "1",
            "action": "set_config",
            "action_parameters": {
				"automatic_https_rewrites":true,
				"autominify":{"html":true, "css":true, "js":true},
				"bic":true,
				"disable_apps":true,
				"polish":"off",
				"disable_zaraz":true,
				"disable_railgun":true,
				"email_obfuscation":true,
				"mirage":true,
				"opportunistic_encryption":true,
				"rocket_loader":true,
				"security_level":"off",
				"server_side_excludes":true,
				"ssl":"off",
				"sxg":true,
				"hotlink_protection":true,
				"fonts":true,
				"disable_rum":true
            },
			"description": "Set all available config rules in one rule",
			"last_updated": "2020-12-18T09:28:09.655749Z",
			"ref": "272936dc447b41fe976255ff6b768ec0",
			"enabled": true
          }
        ],
        "last_updated": "2020-12-18T09:28:09.655749Z",
        "phase": "http_config_settings"
      },
      "success": true,
      "errors": [],
      "messages": []
    }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/rulesets/b232b534beea4e00a21dcbb7a8a545e9", handler)
	mux.HandleFunc("/zones/"+testZoneID+"/rulesets/b232b534beea4e00a21dcbb7a8a545e9", handler)

	lastUpdated, _ := time.Parse(time.RFC3339, "2020-12-18T09:28:09.655749Z")

	rules := []RulesetRule{{
		ID:      "78723a9e0c7c4c6dbec5684cb766231d",
		Version: StringPtr("1"),
		Action:  string(RulesetRuleActionSetConfig),
		ActionParameters: &RulesetRuleActionParameters{
			AutomaticHTTPSRewrites: BoolPtr(true),
			AutoMinify: &RulesetRuleActionParametersAutoMinify{
				HTML: true,
				CSS:  true,
				JS:   true,
			},
			BrowserIntegrityCheck:   BoolPtr(true),
			DisableApps:             BoolPtr(true),
			DisableZaraz:            BoolPtr(true),
			DisableRailgun:          BoolPtr(true),
			DisableRUM:              BoolPtr(true),
			EmailObfuscation:        BoolPtr(true),
			Fonts:                   BoolPtr(true),
			Mirage:                  BoolPtr(true),
			OpportunisticEncryption: BoolPtr(true),
			Polish:                  PolishOff.IntoRef(),
			RocketLoader:            BoolPtr(true),
			SecurityLevel:           SecurityLevelOff.IntoRef(),
			ServerSideExcludes:      BoolPtr(true),
			SSL:                     SSLOff.IntoRef(),
			SXG:                     BoolPtr(true),
			HotLinkProtection:       BoolPtr(true),
		},
		Description: "Set all available config rules in one rule",
		LastUpdated: &lastUpdated,
		Ref:         "272936dc447b41fe976255ff6b768ec0",
		Enabled:     BoolPtr(true),
	}}

	want := Ruleset{
		ID:          "70339d97bdb34195bbf054b1ebe81f76",
		Name:        "Cloudflare Config Rules Ruleset",
		Description: "This ruleset provides config rules modifications",
		Kind:        string(RulesetKindZone),
		Version:     StringPtr("1"),
		LastUpdated: &lastUpdated,
		Phase:       string(RulesetPhaseHTTPConfigSettings),
		Rules:       rules,
	}

	zoneActual, err := client.GetRuleset(context.Background(), ZoneIdentifier(testZoneID), "b232b534beea4e00a21dcbb7a8a545e9")
	if assert.NoError(t, err) {
		assert.Equal(t, want, zoneActual)
	}

	accountActual, err := client.GetRuleset(context.Background(), AccountIdentifier(testAccountID), "b232b534beea4e00a21dcbb7a8a545e9")
	if assert.NoError(t, err) {
		assert.Equal(t, want, accountActual)
	}
}

func TestGetRuleset_RedirectFromValue(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
      "result": {
        "id": "70339d97bdb34195bbf054b1ebe81f76",
        "name": "Cloudflare Redirect Rules Ruleset",
        "description": "This ruleset provides redirect from value",
        "kind": "zone",
        "version": "1",
        "rules": [
          {
            "id": "78723a9e0c7c4c6dbec5684cb766231d",
            "version": "1",
            "action": "redirect",
            "action_parameters": {
				"from_value": {
					"status_code": 301,
					"target_url": {
						"value": "some_host.com"
					},
					"preserve_query_string": true
				}
			},
			"description": "Set dynamic redirect from value",
			"last_updated": "2020-12-18T09:28:09.655749Z",
			"ref": "272936dc447b41fe976255ff6b768ec0",
			"enabled": true
          }
        ],
        "last_updated": "2020-12-18T09:28:09.655749Z",
        "phase": "http_request_dynamic_redirect"
      },
      "success": true,
      "errors": [],
      "messages": []
    }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/rulesets/b232b534beea4e00a21dcbb7a8a545e9", handler)
	mux.HandleFunc("/zones/"+testZoneID+"/rulesets/b232b534beea4e00a21dcbb7a8a545e9", handler)

	lastUpdated, _ := time.Parse(time.RFC3339, "2020-12-18T09:28:09.655749Z")

	rules := []RulesetRule{{
		ID:      "78723a9e0c7c4c6dbec5684cb766231d",
		Version: StringPtr("1"),
		Action:  string(RulesetRuleActionRedirect),
		ActionParameters: &RulesetRuleActionParameters{
			FromValue: &RulesetRuleActionParametersFromValue{
				StatusCode: 301,
				TargetURL: RulesetRuleActionParametersTargetURL{
					Value: "some_host.com",
				},
				PreserveQueryString: BoolPtr(true),
			},
		},
		Description: "Set dynamic redirect from value",
		LastUpdated: &lastUpdated,
		Ref:         "272936dc447b41fe976255ff6b768ec0",
		Enabled:     BoolPtr(true),
	}}

	want := Ruleset{
		ID:          "70339d97bdb34195bbf054b1ebe81f76",
		Name:        "Cloudflare Redirect Rules Ruleset",
		Description: "This ruleset provides redirect from value",
		Kind:        string(RulesetKindZone),
		Version:     StringPtr("1"),
		LastUpdated: &lastUpdated,
		Phase:       string(RulesetPhaseHTTPRequestDynamicRedirect),
		Rules:       rules,
	}

	zoneActual, err := client.GetRuleset(context.Background(), ZoneIdentifier(testZoneID), "b232b534beea4e00a21dcbb7a8a545e9")
	if assert.NoError(t, err) {
		assert.Equal(t, want, zoneActual)
	}

	accountActual, err := client.GetRuleset(context.Background(), AccountIdentifier(testAccountID), "b232b534beea4e00a21dcbb7a8a545e9")
	if assert.NoError(t, err) {
		assert.Equal(t, want, accountActual)
	}
}

func TestGetRuleset_CompressResponse(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
      "result": {
        "id": "70339d97bdb34195bbf054b1ebe81f76",
        "name": "Cloudflare compress response ruleset",
        "description": "This ruleset provides response compression rules",
        "kind": "zone",
        "version": "1",
        "rules": [
          {
            "id": "78723a9e0c7c4c6dbec5684cb766231d",
            "version": "1",
            "action": "compress_response",
            "action_parameters": {
				"algorithms": [ { "name": "brotli" }, { "name": "default" } ]
            },
			"description": "Compress response rule",
			"last_updated": "2020-12-18T09:28:09.655749Z",
			"ref": "272936dc447b41fe976255ff6b768ec0",
			"enabled": true
          }
        ],
        "last_updated": "2020-12-18T09:28:09.655749Z",
        "phase": "http_response_compression"
      },
      "success": true,
      "errors": [],
      "messages": []
    }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/rulesets/b232b534beea4e00a21dcbb7a8a545e9", handler)
	mux.HandleFunc("/zones/"+testZoneID+"/rulesets/b232b534beea4e00a21dcbb7a8a545e9", handler)

	lastUpdated, _ := time.Parse(time.RFC3339, "2020-12-18T09:28:09.655749Z")

	rules := []RulesetRule{{
		ID:      "78723a9e0c7c4c6dbec5684cb766231d",
		Version: StringPtr("1"),
		Action:  string(RulesetRuleActionCompressResponse),
		ActionParameters: &RulesetRuleActionParameters{
			Algorithms: []RulesetRuleActionParametersCompressionAlgorithm{
				{Name: "brotli"},
				{Name: "default"},
			},
		},
		Description: "Compress response rule",
		LastUpdated: &lastUpdated,
		Ref:         "272936dc447b41fe976255ff6b768ec0",
		Enabled:     BoolPtr(true),
	}}

	want := Ruleset{
		ID:          "70339d97bdb34195bbf054b1ebe81f76",
		Name:        "Cloudflare compress response ruleset",
		Description: "This ruleset provides response compression rules",
		Kind:        string(RulesetKindZone),
		Version:     StringPtr("1"),
		LastUpdated: &lastUpdated,
		Phase:       string(RulesetPhaseHTTPResponseCompression),
		Rules:       rules,
	}

	zoneActual, err := client.GetRuleset(context.Background(), ZoneIdentifier(testZoneID), "b232b534beea4e00a21dcbb7a8a545e9")
	if assert.NoError(t, err) {
		assert.Equal(t, want, zoneActual)
	}

	accountActual, err := client.GetRuleset(context.Background(), AccountIdentifier(testAccountID), "b232b534beea4e00a21dcbb7a8a545e9")
	if assert.NoError(t, err) {
		assert.Equal(t, want, accountActual)
	}
}

func TestCreateRuleset(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
      "result": {
        "id": "2c0fc9fa937b11eaa1b71c4d701ab86e",
        "name": "my example ruleset",
        "description": "Test magic transit ruleset",
        "kind": "root",
        "version": "1",
        "last_updated": "2020-12-02T20:24:07.776073Z",
        "phase": "magic_transit",
        "rules": [
          {
            "id": "62449e2e0de149619edb35e59c10d801",
            "version": "1",
            "action": "skip",
            "action_parameters":{
              "ruleset":"current"
            },
            "expression": "tcp.dstport in { 32768..65535 }",
            "description": "Allow TCP Ephemeral Ports",
            "last_updated": "2020-12-02T20:24:07.776073Z",
            "ref": "72449e2e0de149619edb35e59c10d801",
            "enabled": true
          }
        ]
      },
      "success": true,
      "errors": [],
      "messages": []
    }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/rulesets", handler)
	mux.HandleFunc("/zones/"+testZoneID+"/rulesets", handler)

	lastUpdated, _ := time.Parse(time.RFC3339, "2020-12-02T20:24:07.776073Z")

	rules := []RulesetRule{{
		ID:      "62449e2e0de149619edb35e59c10d801",
		Version: StringPtr("1"),
		Action:  string(RulesetRuleActionSkip),
		ActionParameters: &RulesetRuleActionParameters{
			Ruleset: "current",
		},
		Expression:  "tcp.dstport in { 32768..65535 }",
		Description: "Allow TCP Ephemeral Ports",
		LastUpdated: &lastUpdated,
		Ref:         "72449e2e0de149619edb35e59c10d801",
		Enabled:     BoolPtr(true),
	}}

	newRuleset := CreateRulesetParams{
		Name:        "my example ruleset",
		Description: "Test magic transit ruleset",
		Kind:        "root",
		Phase:       string(RulesetPhaseMagicTransit),
		Rules:       rules,
	}

	want := Ruleset{
		ID:          "2c0fc9fa937b11eaa1b71c4d701ab86e",
		Name:        "my example ruleset",
		Description: "Test magic transit ruleset",
		Kind:        "root",
		Version:     StringPtr("1"),
		LastUpdated: &lastUpdated,
		Phase:       string(RulesetPhaseMagicTransit),
		Rules:       rules,
	}

	zoneActual, err := client.CreateRuleset(context.Background(), ZoneIdentifier(testZoneID), newRuleset)
	if assert.NoError(t, err) {
		assert.Equal(t, want, zoneActual)
	}

	accountActual, err := client.CreateRuleset(context.Background(), AccountIdentifier(testAccountID), newRuleset)
	if assert.NoError(t, err) {
		assert.Equal(t, want, accountActual)
	}
}

func TestDeleteRuleset(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, ``)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/rulesets/2c0fc9fa937b11eaa1b71c4d701ab86e", handler)
	mux.HandleFunc("/zones/"+testZoneID+"/rulesets/2c0fc9fa937b11eaa1b71c4d701ab86e", handler)

	zErr := client.DeleteRuleset(context.Background(), ZoneIdentifier(testZoneID), "2c0fc9fa937b11eaa1b71c4d701ab86e")
	assert.NoError(t, zErr)

	aErr := client.DeleteRuleset(context.Background(), AccountIdentifier(testAccountID), "2c0fc9fa937b11eaa1b71c4d701ab86e")
	assert.NoError(t, aErr)
}

func TestUpdateRuleset(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
      "result": {
        "id": "2c0fc9fa937b11eaa1b71c4d701ab86e",
        "name": "ruleset1",
        "description": "Test Firewall Ruleset Update",
        "kind": "root",
        "version": "1",
        "last_updated": "2020-12-02T20:24:07.776073Z",
        "phase": "magic_transit",
        "rules": [
          {
            "id": "62449e2e0de149619edb35e59c10d801",
            "version": "1",
            "action": "skip",
            "action_parameters":{
              "ruleset":"current"
            },
            "expression": "tcp.dstport in { 32768..65535 }",
            "description": "Allow TCP Ephemeral Ports",
            "last_updated": "2020-12-02T20:24:07.776073Z",
            "ref": "72449e2e0de149619edb35e59c10d801",
            "enabled": true
          },
          {
            "id": "62449e2e0de149619edb35e59c10d802",
            "version": "1",
            "action": "skip",
            "action_parameters":{
              "ruleset":"current"
            },
            "expression": "udp.dstport in { 32768..65535 }",
            "description": "Allow UDP Ephemeral Ports",
            "last_updated": "2020-12-02T20:24:07.776073Z",
            "ref": "72449e2e0de149619edb35e59c10d801",
            "enabled": true
          }
        ]
      },
      "success": true,
      "errors": [],
      "messages": []
    }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/rulesets/2c0fc9fa937b11eaa1b71c4d701ab86e", handler)
	mux.HandleFunc("/zones/"+testZoneID+"/rulesets/2c0fc9fa937b11eaa1b71c4d701ab86e", handler)

	lastUpdated, _ := time.Parse(time.RFC3339, "2020-12-02T20:24:07.776073Z")

	rules := []RulesetRule{{
		ID:      "62449e2e0de149619edb35e59c10d801",
		Version: StringPtr("1"),
		Action:  string(RulesetRuleActionSkip),
		ActionParameters: &RulesetRuleActionParameters{
			Ruleset: "current",
		},
		Expression:  "tcp.dstport in { 32768..65535 }",
		Description: "Allow TCP Ephemeral Ports",
		LastUpdated: &lastUpdated,
		Ref:         "72449e2e0de149619edb35e59c10d801",
		Enabled:     BoolPtr(true),
	}, {
		ID:      "62449e2e0de149619edb35e59c10d802",
		Version: StringPtr("1"),
		Action:  string(RulesetRuleActionSkip),
		ActionParameters: &RulesetRuleActionParameters{
			Ruleset: "current",
		},
		Expression:  "udp.dstport in { 32768..65535 }",
		Description: "Allow UDP Ephemeral Ports",
		LastUpdated: &lastUpdated,
		Ref:         "72449e2e0de149619edb35e59c10d801",
		Enabled:     BoolPtr(true),
	}}

	updated := UpdateRulesetParams{
		ID:          "2c0fc9fa937b11eaa1b71c4d701ab86e",
		Description: "Test Firewall Ruleset Update",
		Rules:       rules,
	}

	want := Ruleset{
		ID:          "2c0fc9fa937b11eaa1b71c4d701ab86e",
		Name:        "ruleset1",
		Description: "Test Firewall Ruleset Update",
		Kind:        "root",
		Version:     StringPtr("1"),
		LastUpdated: &lastUpdated,
		Phase:       string(RulesetPhaseMagicTransit),
		Rules:       rules,
	}

	zoneActual, err := client.UpdateRuleset(context.Background(), ZoneIdentifier(testZoneID), updated)
	if assert.NoError(t, err) {
		assert.Equal(t, want, zoneActual)
	}

	accountActual, err := client.UpdateRuleset(context.Background(), AccountIdentifier(testAccountID), updated)
	if assert.NoError(t, err) {
		assert.Equal(t, want, accountActual)
	}
}
