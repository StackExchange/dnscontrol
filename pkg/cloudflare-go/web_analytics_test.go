package cloudflare

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var siteTag = "46c32e0ea0e85e90aa1a6df4596b831e"
var siteToken = "75300e6c2c5648d983fcef2a6c03d14e"
var rulesetID = "2e8804e9-674f-4652-94a4-1c664d0d6764"
var ruleID = "3caf59c9-eda3-4f99-a4a3-ee5fc2358a78"

// var snippetFormat = `\u003c!-- Cloudflare Web Analytics --\u003e\u003cscript defer src='https://static.cloudflareinsights.com/beacon.min.js' data-cf-beacon='{\"token\": \"%s\"}'\u003e\u003c/script\u003e\u003c!-- End Cloudflare Web Analytics --\u003e`.
var snippetFormat = `%s`
var createdTimestamp = time.Now().UTC()
var siteJSON = fmt.Sprintf(`
{
  "site_tag": "%s",
  "site_token": "%s",
  "created": "%s",
  "snippet": "%s",
  "auto_install": true,
  "ruleset": {
    "zone_tag": "%s",
    "zone_name": "example.com",
    "enabled": true,
    "id": "%s"
  },
  "rules": [
    {
      "host": "example.com",
      "paths": [
        "*"
      ],
      "inclusive": true,
      "created": "%s",
      "is_paused": false,
      "priority": 1000,
      "id": "%s"
    }
  ]
}
`, siteTag, siteToken, createdTimestamp.Format(time.RFC3339Nano), fmt.Sprintf(snippetFormat, siteToken), testZoneID, rulesetID, createdTimestamp.Format(time.RFC3339Nano), ruleID)

var rulesetJSON = fmt.Sprintf(`
{
  "id": "%s",
  "zone_tag": "%s",
  "zone_name": "%s",
  "enabled": true
}
`, rulesetID, testZoneID, "example.com")

var ruleJSON = fmt.Sprintf(`
{
  "id": "%s",
  "host": "example.com",
  "paths": [
    "*"
  ],
  "inclusive": true,
  "created": "%s",
  "is_paused": false,
  "priority": 1000
}
`, ruleID, createdTimestamp.Format(time.RFC3339Nano))

var site = WebAnalyticsSite{
	SiteTag:     siteTag,
	SiteToken:   siteToken,
	AutoInstall: true,
	Snippet:     fmt.Sprintf(snippetFormat, siteToken),
	Ruleset:     ruleset,
	Rules: []WebAnalyticsRule{
		rule,
	},
	Created: TimePtr(createdTimestamp.UTC()),
}

var ruleset = WebAnalyticsRuleset{
	ID:       rulesetID,
	ZoneTag:  testZoneID,
	ZoneName: "example.com",
	Enabled:  true,
}

var rule = WebAnalyticsRule{
	ID:   ruleID,
	Host: "example.com",
	Paths: []string{
		"*",
	},
	Inclusive: true,
	Created:   TimePtr(createdTimestamp.UTC()),
	IsPaused:  false,
	Priority:  1000,
}

func TestListWebAnalyticsSites(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "10", r.URL.Query().Get("per_page"))
		assert.Equal(t, "host", r.URL.Query().Get("order_by"))
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			  "success": true,
			  "errors": [],
			  "messages": [],
			  "result": [
			    %s
			  ],
              "result_info": {
                "page": 1,
                "per_page": 10,
                "count": 1,
                "total_count": 1,
                "total_pages": 1
              }
			}
		`, siteJSON)
	}
	mux.HandleFunc("/accounts/"+testAccountID+"/rum/site_info/list", handler)
	want := []WebAnalyticsSite{site}
	actual, resultInfo, err := client.ListWebAnalyticsSites(context.Background(), AccountIdentifier(testAccountID), ListWebAnalyticsSitesParams{
		ResultInfo: ResultInfo{
			Page:    0,
			PerPage: 0,
		},
		OrderBy: "host",
	})
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
		assert.Equal(t, &ResultInfo{
			Page:       1,
			PerPage:    10,
			TotalPages: 1,
			Count:      1,
			Total:      1,
		}, resultInfo)
		assert.Len(t, actual, 1)
	}
}

func TestGetWebAnalyticsSite(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			  "success": true,
			  "errors": [],
			  "messages": [],
			  "result": %s
			}
		`, siteJSON)
	}
	mux.HandleFunc("/accounts/"+testAccountID+"/rum/site_info/"+siteTag, handler)
	want := site
	actual, err := client.GetWebAnalyticsSite(context.Background(), AccountIdentifier(testAccountID), GetWebAnalyticsSiteParams{
		SiteTag: siteTag,
	})
	if assert.NoError(t, err) {
		assert.Equal(t, &want, actual)
	}
}

func TestCreateWebAnalyticsSite(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.True(t, strings.Contains(string(body), `"auto_install":true`))
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			  "success": true,
			  "errors": [],
			  "messages": [],
			  "result": %s
			}
		`, siteJSON)
	}
	mux.HandleFunc("/accounts/"+testAccountID+"/rum/site_info", handler)
	want := site
	actual, err := client.CreateWebAnalyticsSite(context.Background(), AccountIdentifier(testAccountID), CreateWebAnalyticsSiteParams{
		ZoneTag: testZoneID,
		//AutoInstall: BoolPtr(true),  // should default to true
	})
	if assert.NoError(t, err) {
		assert.Equal(t, &want, actual)
	}
}

func TestUpdateWebAnalyticsSite(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			  "success": true,
			  "errors": [],
			  "messages": [],
			  "result": %s
			}
		`, siteJSON)
	}
	mux.HandleFunc("/accounts/"+testAccountID+"/rum/site_info/"+siteTag, handler)
	want := site
	actual, err := client.UpdateWebAnalyticsSite(context.Background(), AccountIdentifier(testAccountID), UpdateWebAnalyticsSiteParams{
		SiteTag:     site.SiteTag,
		Host:        "example.com",
		AutoInstall: BoolPtr(true),
	})
	if assert.NoError(t, err) {
		assert.Equal(t, &want, actual)
	}
}

func TestDeleteWebAnalyticsSite(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			  "success": true,
			  "errors": [],
			  "messages": [],
			  "result": {
                "site_tag": "%s"
              }
			}
		`, siteTag)
	}
	mux.HandleFunc("/accounts/"+testAccountID+"/rum/site_info/"+siteTag, handler)
	want := siteTag
	actual, err := client.DeleteWebAnalyticsSite(context.Background(), AccountIdentifier(testAccountID), DeleteWebAnalyticsSiteParams{
		SiteTag: siteTag,
	})
	if assert.NoError(t, err) {
		assert.Equal(t, &want, actual)
	}
}

func TestListWebAnalyticsRules(t *testing.T) {
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
                "ruleset": %s,
                "rules": [
                  %s
                ]
              }
			}
		`, rulesetJSON, ruleJSON)
	}
	mux.HandleFunc("/accounts/"+testAccountID+"/rum/v2/"+rulesetID+"/rules", handler)
	want := WebAnalyticsRulesetRules{
		Ruleset: ruleset,
		Rules:   []WebAnalyticsRule{rule},
	}
	actual, err := client.ListWebAnalyticsRules(context.Background(), AccountIdentifier(testAccountID), ListWebAnalyticsRulesParams{
		RulesetID: rulesetID,
	})
	if assert.NoError(t, err) {
		assert.Equal(t, &want, actual)
	}
}

func TestCreateWebAnalyticsRule(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			  "success": true,
			  "errors": [],
			  "messages": [],
			  "result": %s
			}
		`, ruleJSON)
	}
	mux.HandleFunc("/accounts/"+testAccountID+"/rum/v2/"+rulesetID+"/rule", handler)
	want := rule
	actual, err := client.CreateWebAnalyticsRule(context.Background(), AccountIdentifier(testAccountID), CreateWebAnalyticsRuleParams{
		RulesetID: rulesetID,
		Rule: CreateWebAnalyticsRule{
			Host:      "example.com",
			Paths:     []string{"*"},
			Inclusive: true,
			IsPaused:  false,
		},
	})
	if assert.NoError(t, err) {
		assert.Equal(t, &want, actual)
	}
}

func TestUpdateWebAnalyticsRule(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			  "success": true,
			  "errors": [],
			  "messages": [],
			  "result": %s
			}
		`, ruleJSON)
	}
	mux.HandleFunc("/accounts/"+testAccountID+"/rum/v2/"+rulesetID+"/rule/"+ruleID, handler)
	want := rule
	actual, err := client.UpdateWebAnalyticsRule(context.Background(), AccountIdentifier(testAccountID), UpdateWebAnalyticsRuleParams{
		RulesetID: rulesetID,
		RuleID:    ruleID,
		Rule: CreateWebAnalyticsRule{
			Host:      "example.com",
			Paths:     []string{"*"},
			Inclusive: true,
			IsPaused:  false,
		},
	})
	if assert.NoError(t, err) {
		assert.Equal(t, &want, actual)
	}
}

func TestDeleteWebAnalyticsRule(t *testing.T) {
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
                "id": "%s"
              }
			}
		`, ruleID)
	}
	mux.HandleFunc("/accounts/"+testAccountID+"/rum/v2/"+rulesetID+"/rule/"+ruleID, handler)
	want := ruleID
	actual, err := client.DeleteWebAnalyticsRule(context.Background(), AccountIdentifier(testAccountID), DeleteWebAnalyticsRuleParams{
		RulesetID: rulesetID,
		RuleID:    ruleID,
	})
	if assert.NoError(t, err) {
		assert.Equal(t, &want, actual)
	}
}
