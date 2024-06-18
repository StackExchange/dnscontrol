package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestListMagicFirewallRulesets(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"result": [
				{
					"id": "2c0fc9fa937b11eaa1b71c4d701ab86e",
					"name": "ruleset1",
					"description": "Test Firewall Ruleset",
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

	lastUpdated, _ := time.Parse(time.RFC3339, "2020-12-02T20:24:07.776073Z")

	want := []MagicFirewallRuleset{
		{
			ID:          "2c0fc9fa937b11eaa1b71c4d701ab86e",
			Name:        "ruleset1",
			Description: "Test Firewall Ruleset",
			Kind:        "root",
			Version:     "1",
			LastUpdated: &lastUpdated,
			Phase:       MagicFirewallRulesetPhaseMagicTransit,
		},
	}

	actual, err := client.ListMagicFirewallRulesets(context.Background(), testAccountID)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestGetMagicFirewallRuleset(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"result": {
				"id": "2c0fc9fa937b11eaa1b71c4d701ab86e",
				"name": "ruleset1",
				"description": "Test Firewall Ruleset",
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

	mux.HandleFunc("/accounts/"+testAccountID+"/rulesets/2c0fc9fa937b11eaa1b71c4d701ab86e", handler)

	lastUpdated, _ := time.Parse(time.RFC3339, "2020-12-02T20:24:07.776073Z")
	rules := []MagicFirewallRulesetRule{{
		ID:      "62449e2e0de149619edb35e59c10d801",
		Version: "1",
		Action:  MagicFirewallRulesetRuleActionSkip,
		ActionParameters: &MagicFirewallRulesetRuleActionParameters{
			Ruleset: "current",
		},
		Expression:  "tcp.dstport in { 32768..65535 }",
		Description: "Allow TCP Ephemeral Ports",
		LastUpdated: &lastUpdated,
		Ref:         "72449e2e0de149619edb35e59c10d801",
		Enabled:     true,
	}}

	want := MagicFirewallRuleset{
		ID:          "2c0fc9fa937b11eaa1b71c4d701ab86e",
		Name:        "ruleset1",
		Description: "Test Firewall Ruleset",
		Kind:        "root",
		Version:     "1",
		LastUpdated: &lastUpdated,
		Phase:       MagicFirewallRulesetPhaseMagicTransit,
		Rules:       rules,
	}

	actual, err := client.GetMagicFirewallRuleset(context.Background(), testAccountID, "2c0fc9fa937b11eaa1b71c4d701ab86e")
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateMagicFirewallRuleset(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"result": {
				"id": "2c0fc9fa937b11eaa1b71c4d701ab86e",
				"name": "ruleset1",
				"description": "Test Firewall Ruleset",
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

	lastUpdated, _ := time.Parse(time.RFC3339, "2020-12-02T20:24:07.776073Z")
	rules := []MagicFirewallRulesetRule{{
		ID:      "62449e2e0de149619edb35e59c10d801",
		Version: "1",
		Action:  MagicFirewallRulesetRuleActionSkip,
		ActionParameters: &MagicFirewallRulesetRuleActionParameters{
			Ruleset: "current",
		},
		Expression:  "tcp.dstport in { 32768..65535 }",
		Description: "Allow TCP Ephemeral Ports",
		LastUpdated: &lastUpdated,
		Ref:         "72449e2e0de149619edb35e59c10d801",
		Enabled:     true,
	}}

	want := MagicFirewallRuleset{
		ID:          "2c0fc9fa937b11eaa1b71c4d701ab86e",
		Name:        "ruleset1",
		Description: "Test Firewall Ruleset",
		Kind:        "root",
		Version:     "1",
		LastUpdated: &lastUpdated,
		Phase:       MagicFirewallRulesetPhaseMagicTransit,
		Rules:       rules,
	}

	actual, err := client.CreateMagicFirewallRuleset(context.Background(), testAccountID, "ruleset1", "Test Firewall Ruleset", rules)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestUpdateMagicFirewallRuleset(t *testing.T) {
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

	lastUpdated, _ := time.Parse(time.RFC3339, "2020-12-02T20:24:07.776073Z")
	rules := []MagicFirewallRulesetRule{{
		ID:      "62449e2e0de149619edb35e59c10d801",
		Version: "1",
		Action:  MagicFirewallRulesetRuleActionSkip,
		ActionParameters: &MagicFirewallRulesetRuleActionParameters{
			Ruleset: "current",
		},
		Expression:  "tcp.dstport in { 32768..65535 }",
		Description: "Allow TCP Ephemeral Ports",
		LastUpdated: &lastUpdated,
		Ref:         "72449e2e0de149619edb35e59c10d801",
		Enabled:     true,
	}, {
		ID:      "62449e2e0de149619edb35e59c10d802",
		Version: "1",
		Action:  MagicFirewallRulesetRuleActionSkip,
		ActionParameters: &MagicFirewallRulesetRuleActionParameters{
			Ruleset: "current",
		},
		Expression:  "udp.dstport in { 32768..65535 }",
		Description: "Allow UDP Ephemeral Ports",
		LastUpdated: &lastUpdated,
		Ref:         "72449e2e0de149619edb35e59c10d801",
		Enabled:     true,
	}}

	want := MagicFirewallRuleset{
		ID:          "2c0fc9fa937b11eaa1b71c4d701ab86e",
		Name:        "ruleset1",
		Description: "Test Firewall Ruleset Update",
		Kind:        "root",
		Version:     "1",
		LastUpdated: &lastUpdated,
		Phase:       MagicFirewallRulesetPhaseMagicTransit,
		Rules:       rules,
	}

	actual, err := client.UpdateMagicFirewallRuleset(context.Background(), testAccountID, "2c0fc9fa937b11eaa1b71c4d701ab86e", "Test Firewall Ruleset Update", rules)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

// Firewall API is not implementing the standard response blob but returns an empty response (204) in case
// of a success. So we are checking for the response body size here
// TODO, This is going to be changed by MFW-63.
func TestDeleteMagicFirewallRuleset(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, ``)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/rulesets/2c0fc9fa937b11eaa1b71c4d701ab86e", handler)

	err := client.DeleteMagicFirewallRuleset(context.Background(), testAccountID, "2c0fc9fa937b11eaa1b71c4d701ab86e")
	assert.NoError(t, err)
}
