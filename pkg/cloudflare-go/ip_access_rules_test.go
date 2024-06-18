package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListIPAccessRules(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"result":[
						{
							"allowed_modes": [
								"whitelist",
								"block",
								"challenge",
								"js_challenge",
								"managed_challenge"
							],
							"configuration": {
								"target": "ip",
								"value": "198.51.100.1"
							},
							"created_on": "2014-01-01T05:20:00.12345Z",
							"id": "f2d427378e7542acb295380d352e2ebd",
							"mode": "whitelist",
							"modified_on": "2014-01-01T05:20:00.12345Z",
							"notes": "Whitelisting this IP."
						},
						{
							"allowed_modes": [
								"whitelist",
								"block",
								"challenge",
								"js_challenge",
								"managed_challenge"
							],
							"configuration": {
								"target": "ip",
								"value": "198.51.100.2"
							},
							"created_on": "2014-01-01T05:20:00.12345Z",
							"id": "92f17202ed8bd63d69a66b86a49a8f6b",
							"mode": "block",
							"modified_on": "2014-01-01T05:20:00.12345Z",
							"notes": "This rule is enabled because of an event that occurred on date X."
						},
						{
							"allowed_modes": [
								"whitelist",
								"block",
								"challenge",
								"js_challenge",
								"managed_challenge"
							],
							"configuration": {
								"target": "ip",
								"value": "198.51.100.3"
							},
							"created_on": "2014-01-01T05:20:00.12345Z",
							"id": "4ae338944d6143378c3cf05a7c77d983",
							"mode": "challenge",
							"modified_on": "2014-01-01T05:20:00.12345Z",
							"notes": "This rule is enabled because of an event that occurred on date Y."
						},
						{
							"allowed_modes": [
								"whitelist",
								"block",
								"challenge",
								"js_challenge",
								"managed_challenge"
							],
							"configuration": {
								"target": "ip",
								"value": "198.51.100.4"
							},
							"created_on": "2014-01-01T05:20:00.12345Z",
							"id": "52161eb6af4241bb9d4b32394be72fdf",
							"mode": "js_challenge",
							"modified_on": "2014-01-01T05:20:00.12345Z",
							"notes": "This rule is enabled because of an event that occurred on date Z."
						},
						{
							"allowed_modes": [
								"whitelist",
								"block",
								"challenge",
								"js_challenge",
								"managed_challenge"
							],
							"configuration": {
								"target": "ip",
								"value": "198.51.100.5"
							},
							"created_on": "2014-01-01T05:20:00.12345Z",
							"id": "cbf4b7a5a2a24e59a03044d6d44ceb09",
							"mode": "managed_challenge",
							"modified_on": "2014-01-01T05:20:00.12345Z",
							"notes": "This rule is enabled because we like the challenge page."
						}
			],
			"success":true,
			"errors":null,
			"messages":null,
			"result_info":{
				"page":1,
				"per_page":25,
				"count":5,
				"total_count":5
			}
		}
		`)
	}

	mux.HandleFunc("/zones/d56084adb405e0b7e32c52321bf07be6/firewall/access_rules/rules", handler)
	want := []IPAccessRule{
		{
			AllowedModes: []IPAccessRulesModeOption{
				IPAccessRulesModeWhitelist,
				IPAccessRulesModeBlock,
				IPAccessRulesModeChallenge,
				IPAccessRulesModeJsChallenge,
				IPAccessRulesModeManagedChallenge,
			},
			ID: "f2d427378e7542acb295380d352e2ebd",
			Configuration: IPAccessRuleConfiguration{
				Target: "ip",
				Value:  "198.51.100.1",
			},
			CreatedOn:  "2014-01-01T05:20:00.12345Z",
			Mode:       "whitelist",
			ModifiedOn: "2014-01-01T05:20:00.12345Z",
			Notes:      "Whitelisting this IP.",
		},
		{
			AllowedModes: []IPAccessRulesModeOption{
				IPAccessRulesModeWhitelist,
				IPAccessRulesModeBlock,
				IPAccessRulesModeChallenge,
				IPAccessRulesModeJsChallenge,
				IPAccessRulesModeManagedChallenge,
			},
			ID: "92f17202ed8bd63d69a66b86a49a8f6b",
			Configuration: IPAccessRuleConfiguration{
				Target: "ip",
				Value:  "198.51.100.2",
			},
			CreatedOn:  "2014-01-01T05:20:00.12345Z",
			Mode:       "block",
			ModifiedOn: "2014-01-01T05:20:00.12345Z",
			Notes:      "This rule is enabled because of an event that occurred on date X.",
		},
		{
			AllowedModes: []IPAccessRulesModeOption{
				IPAccessRulesModeWhitelist,
				IPAccessRulesModeBlock,
				IPAccessRulesModeChallenge,
				IPAccessRulesModeJsChallenge,
				IPAccessRulesModeManagedChallenge,
			},
			ID: "4ae338944d6143378c3cf05a7c77d983",
			Configuration: IPAccessRuleConfiguration{
				Target: "ip",
				Value:  "198.51.100.3",
			},
			CreatedOn:  "2014-01-01T05:20:00.12345Z",
			Mode:       "challenge",
			ModifiedOn: "2014-01-01T05:20:00.12345Z",
			Notes:      "This rule is enabled because of an event that occurred on date Y.",
		},
		{
			AllowedModes: []IPAccessRulesModeOption{
				IPAccessRulesModeWhitelist,
				IPAccessRulesModeBlock,
				IPAccessRulesModeChallenge,
				IPAccessRulesModeJsChallenge,
				IPAccessRulesModeManagedChallenge,
			},
			ID: "52161eb6af4241bb9d4b32394be72fdf",
			Configuration: IPAccessRuleConfiguration{
				Target: "ip",
				Value:  "198.51.100.4",
			},
			CreatedOn:  "2014-01-01T05:20:00.12345Z",
			Mode:       "js_challenge",
			ModifiedOn: "2014-01-01T05:20:00.12345Z",
			Notes:      "This rule is enabled because of an event that occurred on date Z.",
		},
		{
			AllowedModes: []IPAccessRulesModeOption{
				IPAccessRulesModeWhitelist,
				IPAccessRulesModeBlock,
				IPAccessRulesModeChallenge,
				IPAccessRulesModeJsChallenge,
				IPAccessRulesModeManagedChallenge,
			},
			ID: "cbf4b7a5a2a24e59a03044d6d44ceb09",
			Configuration: IPAccessRuleConfiguration{
				Target: "ip",
				Value:  "198.51.100.5",
			},
			CreatedOn:  "2014-01-01T05:20:00.12345Z",
			Mode:       "managed_challenge",
			ModifiedOn: "2014-01-01T05:20:00.12345Z",
			Notes:      "This rule is enabled because we like the challenge page.",
		},
	}

	actual, _, err := client.ListIPAccessRules(context.Background(),
		ZoneIdentifier("d56084adb405e0b7e32c52321bf07be6"), ListIPAccessRulesParams{})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestListIPAccessRulesWithParams(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"result":[
						{
							"allowed_modes": [
								"whitelist",
								"block",
								"challenge",
								"js_challenge",
								"managed_challenge"
							],
							"configuration": {
								"target": "ip",
								"value": "198.51.100.5"
							},
							"created_on": "2014-01-01T05:20:00.12345Z",
							"id": "cbf4b7a5a2a24e59a03044d6d44ceb09",
							"mode": "managed_challenge",
							"modified_on": "2014-01-01T05:20:00.12345Z",
							"notes": "This rule is enabled because we like the challenge page."
						}
			],
			"success":true,
			"errors":null,
			"messages":null,
			"result_info":{
				"page":1,
				"per_page":100,
				"count":1,
				"total_count":1
			}
		}
		`)
	}

	mux.HandleFunc("/zones/d56084adb405e0b7e32c52321bf07be6/firewall/access_rules/rules", handler)
	want := []IPAccessRule{
		{
			AllowedModes: []IPAccessRulesModeOption{
				IPAccessRulesModeWhitelist,
				IPAccessRulesModeBlock,
				IPAccessRulesModeChallenge,
				IPAccessRulesModeJsChallenge,
				IPAccessRulesModeManagedChallenge,
			},
			ID: "cbf4b7a5a2a24e59a03044d6d44ceb09",
			Configuration: IPAccessRuleConfiguration{
				Target: "ip",
				Value:  "198.51.100.5",
			},
			CreatedOn:  "2014-01-01T05:20:00.12345Z",
			Mode:       "managed_challenge",
			ModifiedOn: "2014-01-01T05:20:00.12345Z",
			Notes:      "This rule is enabled because we like the challenge page.",
		},
	}

	actual, _, err := client.ListIPAccessRules(context.Background(),
		ZoneIdentifier("d56084adb405e0b7e32c52321bf07be6"), ListIPAccessRulesParams{
			Direction: "asc",
			Filters: ListIPAccessRulesFilters{
				Configuration: IPAccessRuleConfiguration{
					Target: "ip",
				},
				Match: "any",
				Mode:  "manage_challenge",
				Notes: "This rule is enabled because we like the challenge page.",
			},
			Order: IPAccessRulesConfigurationTarget,
			PaginationOptions: PaginationOptions{
				Page:    1,
				PerPage: 100,
			},
		})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}
