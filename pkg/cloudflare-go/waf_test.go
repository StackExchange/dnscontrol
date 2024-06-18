package cloudflare

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListWAFPackages(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		// JSON data from: https://api.cloudflare.com/#waf-rule-packages-properties
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": [
				{
				"id": "a25a9a7e9c00afc1fb2e0245519d725b",
				"name": "WordPress rules",
				"description": "Common WordPress exploit protections",
				"detection_mode": "traditional",
				"zone_id": "023e105f4ecef8ad9ca31a8372d0c353",
				"status": "active"
				}
			],
			"result_info": {
				"page": 1,
				"per_page": 20,
				"count": 1,
				"total_count": 1
			}
		}`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/firewall/waf/packages", handler)

	want := []WAFPackage{
		{
			ID:            "a25a9a7e9c00afc1fb2e0245519d725b",
			Name:          "WordPress rules",
			Description:   "Common WordPress exploit protections",
			ZoneID:        "023e105f4ecef8ad9ca31a8372d0c353",
			DetectionMode: "traditional",
			Sensitivity:   "",
			ActionMode:    "",
		},
	}

	d, err := client.ListWAFPackages(context.Background(), testZoneID)

	if assert.NoError(t, err) {
		assert.Equal(t, want, d)
	}

	_, err = client.ListWAFRules(context.Background(), testZoneID, "123")
	assert.Error(t, err)
}

func TestListWAFPackagesMultiplePages(t *testing.T) {
	setup()
	defer teardown()

	page := 1
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)

		reqURI, err := url.ParseRequestURI(r.RequestURI)
		assert.NoError(t, err)

		query, err := url.ParseQuery(reqURI.RawQuery)
		assert.NoError(t, err)

		assert.Equal(t, query, url.Values{"page": []string{strconv.Itoa(page)}, "per_page": []string{"100"}})

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": [
				{
				"id": "fake_id_number_%[1]d",
				"name": "Fake rule name %[1]d",
				"description": "Fake rule description %[1]d",
				"detection_mode": "traditional",
				"zone_id": "%[2]s",
				"status": "active"
				}
			],
			"result_info": {
				"page": %[1]d,
				"per_page": 1,
				"total_pages": 2,
				"count": 1,
				"total_count": 2
			}
		}`, page, testZoneID)

		page++
	}

	mux.HandleFunc("/zones/"+testZoneID+"/firewall/waf/packages", handler)

	want := []WAFPackage{
		{
			ID:            "fake_id_number_1",
			Name:          "Fake rule name 1",
			Description:   "Fake rule description 1",
			ZoneID:        testZoneID,
			DetectionMode: "traditional",
			Sensitivity:   "",
			ActionMode:    "",
		},
		{
			ID:            "fake_id_number_2",
			Name:          "Fake rule name 2",
			Description:   "Fake rule description 2",
			ZoneID:        testZoneID,
			DetectionMode: "traditional",
			Sensitivity:   "",
			ActionMode:    "",
		},
	}

	d, err := client.ListWAFPackages(context.Background(), testZoneID)

	if assert.NoError(t, err) {
		assert.Equal(t, want, d)
	}

	_, err = client.ListWAFRules(context.Background(), testZoneID, "123")
	assert.Error(t, err)
}

func TestWAFPackage(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		// JSON data from: https://api.cloudflare.com/#waf-rule-packages-properties
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result":
			{
				"id": "a25a9a7e9c00afc1fb2e0245519d725b",
				"name": "WordPress rules",
				"description": "Common WordPress exploit protections",
				"detection_mode": "traditional",
				"zone_id": "023e105f4ecef8ad9ca31a8372d0c353",
				"status": "active"
			}
		}`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/firewall/waf/packages/a25a9a7e9c00afc1fb2e0245519d725b", handler)

	want := WAFPackage{
		ID:            "a25a9a7e9c00afc1fb2e0245519d725b",
		Name:          "WordPress rules",
		Description:   "Common WordPress exploit protections",
		ZoneID:        "023e105f4ecef8ad9ca31a8372d0c353",
		DetectionMode: "traditional",
		Sensitivity:   "",
		ActionMode:    "",
	}

	d, err := client.WAFPackage(context.Background(), testZoneID, "a25a9a7e9c00afc1fb2e0245519d725b")

	if assert.NoError(t, err) {
		assert.Equal(t, want, d)
	}

	_, err = client.WAFPackage(context.Background(), testZoneID, "123")
	assert.Error(t, err)
}

func TestUpdateWAFPackage(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		defer r.Body.Close()

		assert.Equal(t, `{"sensitivity":"high","action_mode":"challenge"}`, string(body), "Expected body '{\"sensitivity\":\"high\",\"action_mode\":\"challenge\"}', got %s", string(body))

		w.Header().Set("content-type", "application/json")
		// JSON data from: https://api.cloudflare.com/#waf-rules-properties
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "a25a9a7e9c00afc1fb2e0245519d725b",
				"name": "OWASP ModSecurity Core Rule Set",
				"description": "Covers OWASP Top 10 vulnerabilities, and more.",
				"detection_mode": "anomaly",
				"zone_id": "023e105f4ecef8ad9ca31a8372d0c353",
				"status": "active",
				"sensitivity": "high",
				"action_mode": "challenge"
			}
		}`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/firewall/waf/packages/a25a9a7e9c00afc1fb2e0245519d725b", handler)

	want := WAFPackage{
		ID:            "a25a9a7e9c00afc1fb2e0245519d725b",
		Name:          "OWASP ModSecurity Core Rule Set",
		Description:   "Covers OWASP Top 10 vulnerabilities, and more.",
		ZoneID:        "023e105f4ecef8ad9ca31a8372d0c353",
		DetectionMode: "anomaly",
		Sensitivity:   "high",
		ActionMode:    "challenge",
	}

	d, err := client.UpdateWAFPackage(context.Background(), testZoneID, "a25a9a7e9c00afc1fb2e0245519d725b", WAFPackageOptions{Sensitivity: "high", ActionMode: "challenge"})

	if assert.NoError(t, err) {
		assert.Equal(t, want, d)
	}
}

func TestListWAFGroups(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)

		w.Header().Set("content-type", "application/json")

		// JSON data from: https://api.cloudflare.com/#waf-rule-groups-properties
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": [
				{
					"id": "de677e5818985db1285d0e80225f06e5",
					"name": "Project Honey Pot",
					"description": "Group designed to protect against IP addresses that are a threat and typically used to launch DDoS attacks",
					"rules_count": 10,
					"modified_rules_count": 2,
					"package_id": "a25a9a7e9c00afc1fb2e0245519d725b",
					"mode": "on",
					"allowed_modes": [
						"on",
						"off"
					]
				}
			],
			"result_info": {
				"page": 1,
				"per_page": 20,
				"count": 1,
				"total_count": 1
			}
			}`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/firewall/waf/packages/a25a9a7e9c00afc1fb2e0245519d725b/groups", handler)

	want := []WAFGroup{
		{
			ID:                 "de677e5818985db1285d0e80225f06e5",
			Name:               "Project Honey Pot",
			Description:        "Group designed to protect against IP addresses that are a threat and typically used to launch DDoS attacks",
			RulesCount:         10,
			ModifiedRulesCount: 2,
			PackageID:          "a25a9a7e9c00afc1fb2e0245519d725b",
			Mode:               "on",
			AllowedModes:       []string{"on", "off"},
		},
	}

	d, err := client.ListWAFGroups(context.Background(), testZoneID, "a25a9a7e9c00afc1fb2e0245519d725b")

	if assert.NoError(t, err) {
		assert.Equal(t, want, d)
	}

	_, err = client.ListWAFGroups(context.Background(), testZoneID, "123")
	assert.Error(t, err)
}

func TestListWAFGroupsMultiplePages(t *testing.T) {
	setup()
	defer teardown()
	packageID := "efgh456"

	page := 1
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)

		reqURI, err := url.ParseRequestURI(r.RequestURI)
		assert.NoError(t, err)

		query, err := url.ParseQuery(reqURI.RawQuery)
		assert.NoError(t, err)

		assert.Equal(t, query, url.Values{"page": []string{strconv.Itoa(page)}, "per_page": []string{"100"}})

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": [
				{
					"id": "fake_group_id_%[1]d",
					"name": "Fake Group Name %[1]d",
					"description": "Fake Group Description %[1]d",
					"rules_count": 1%[1]d,
					"modified_rules_count": %[1]d,
					"package_id": "%[2]s",
					"mode": "on",
					"allowed_modes": [
						"on",
						"off"
					]
				}
			],
			"result_info": {
				"page": %[1]d,
				"per_page": 1,
				"total_pages": 2,
				"count": 1,
				"total_count": 2
			}
		}`, page, packageID)

		page++
	}

	mux.HandleFunc("/zones/"+testZoneID+"/firewall/waf/packages/"+packageID+"/groups", handler)

	want := []WAFGroup{
		{
			ID:                 "fake_group_id_1",
			Name:               "Fake Group Name 1",
			Description:        "Fake Group Description 1",
			RulesCount:         11,
			ModifiedRulesCount: 1,
			PackageID:          packageID,
			Mode:               "on",
			AllowedModes:       []string{"on", "off"},
		},
		{
			ID:                 "fake_group_id_2",
			Name:               "Fake Group Name 2",
			Description:        "Fake Group Description 2",
			RulesCount:         12,
			ModifiedRulesCount: 2,
			PackageID:          packageID,
			Mode:               "on",
			AllowedModes:       []string{"on", "off"},
		},
	}

	d, err := client.ListWAFGroups(context.Background(), testZoneID, packageID)

	if assert.NoError(t, err) {
		assert.Equal(t, want, d)
	}

	_, err = client.ListWAFGroups(context.Background(), testZoneID, "123")
	assert.Error(t, err)
}

func TestWAFGroup(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		// JSON data from: https://api.cloudflare.com/#waf-rule-groups-rule-group-details
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "de677e5818985db1285d0e80225f06e5",
				"name": "Project Honey Pot",
				"description": "Group designed to protect against IP addresses that are a threat and typically used to launch DDoS attacks",
				"rules_count": 10,
				"modified_rules_count": 2,
				"package_id": "a25a9a7e9c00afc1fb2e0245519d725b",
				"mode": "on",
				"allowed_modes": [
					"on",
					"off"
				]
			}
		}`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/firewall/waf/packages/a25a9a7e9c00afc1fb2e0245519d725b/groups/de677e5818985db1285d0e80225f06e5", handler)

	want := WAFGroup{
		ID:                 "de677e5818985db1285d0e80225f06e5",
		Name:               "Project Honey Pot",
		Description:        "Group designed to protect against IP addresses that are a threat and typically used to launch DDoS attacks",
		RulesCount:         10,
		ModifiedRulesCount: 2,
		PackageID:          "a25a9a7e9c00afc1fb2e0245519d725b",
		Mode:               "on",
		AllowedModes:       []string{"on", "off"},
	}

	d, err := client.WAFGroup(context.Background(), testZoneID, "a25a9a7e9c00afc1fb2e0245519d725b", "de677e5818985db1285d0e80225f06e5")

	if assert.NoError(t, err) {
		assert.Equal(t, want, d)
	}

	_, err = client.WAFGroup(context.Background(), testZoneID, "a25a9a7e9c00afc1fb2e0245519d725b", "123")
	assert.Error(t, err)
}

func TestUpdateWAFGroup(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		defer r.Body.Close()

		assert.Equal(t, `{"mode":"on"}`, string(body), "Expected body '{\"mode\":\"on\"}', got %s", string(body))

		w.Header().Set("content-type", "application/json")
		// JSON data from: https://api.cloudflare.com/#waf-rule-groups-edit-rule-group
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "de677e5818985db1285d0e80225f06e5",
				"name": "Project Honey Pot",
				"description": "Group designed to protect against IP addresses that are a threat and typically used to launch DDoS attacks",
				"rules_count": 10,
				"modified_rules_count": 2,
				"package_id": "a25a9a7e9c00afc1fb2e0245519d725b",
				"mode": "on",
				"allowed_modes": [
					"on",
					"off"
				]
			}
		}`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/firewall/waf/packages/a25a9a7e9c00afc1fb2e0245519d725b/groups/de677e5818985db1285d0e80225f06e5", handler)

	want := WAFGroup{
		ID:                 "de677e5818985db1285d0e80225f06e5",
		Name:               "Project Honey Pot",
		Description:        "Group designed to protect against IP addresses that are a threat and typically used to launch DDoS attacks",
		RulesCount:         10,
		ModifiedRulesCount: 2,
		PackageID:          "a25a9a7e9c00afc1fb2e0245519d725b",
		Mode:               "on",
		AllowedModes:       []string{"on", "off"},
	}

	d, err := client.UpdateWAFGroup(context.Background(), testZoneID, "a25a9a7e9c00afc1fb2e0245519d725b", "de677e5818985db1285d0e80225f06e5", "on")

	if assert.NoError(t, err) {
		assert.Equal(t, want, d)
	}
}

func TestListWAFRules(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		// JSON data from: https://api.cloudflare.com/#waf-rules-properties
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": [
				{
				"id": "f939de3be84e66e757adcdcb87908023",
				"description": "SQL injection prevention for SELECT statements",
				"priority": "5",
				"group": {
					"id": "de677e5818985db1285d0e80225f06e5",
					"name": "Project Honey Pot"
				},
				"package_id": "a25a9a7e9c00afc1fb2e0245519d725b",
				"allowed_modes": [
					"on",
					"off"
				],
				"mode": "on"
				}
			],
			"result_info": {
				"page": 1,
				"per_page": 20,
				"count": 1,
				"total_count": 1
			}
		}`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/firewall/waf/packages/a25a9a7e9c00afc1fb2e0245519d725b/rules", handler)

	want := []WAFRule{
		{
			ID:          "f939de3be84e66e757adcdcb87908023",
			Description: "SQL injection prevention for SELECT statements",
			Priority:    "5",
			PackageID:   "a25a9a7e9c00afc1fb2e0245519d725b",
			Group: struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}{
				ID:   "de677e5818985db1285d0e80225f06e5",
				Name: "Project Honey Pot",
			},
			Mode:         "on",
			DefaultMode:  "",
			AllowedModes: []string{"on", "off"},
		},
	}

	d, err := client.ListWAFRules(context.Background(), testZoneID, "a25a9a7e9c00afc1fb2e0245519d725b")

	if assert.NoError(t, err) {
		assert.Equal(t, want, d)
	}

	_, err = client.ListWAFRules(context.Background(), testZoneID, "123")
	assert.Error(t, err)
}

func TestListWAFRulesMultiplePages(t *testing.T) {
	setup()
	defer teardown()
	packageID := "efgh456"

	page := 1
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)

		reqURI, err := url.ParseRequestURI(r.RequestURI)
		assert.NoError(t, err)

		query, err := url.ParseQuery(reqURI.RawQuery)
		assert.NoError(t, err)

		assert.Equal(t, query, url.Values{"page": []string{strconv.Itoa(page)}, "per_page": []string{"100"}})

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": [
				{
				"id": "fake_rule_id_%[1]d",
				"description": "Fake Rule Description %[1]d",
				"priority": "%[1]d",
				"group": {
					"id": "fake_group_id_%[1]d",
					"name": "Fake Group Name %[1]d"
				},
				"package_id": "%[2]s",
				"allowed_modes": [
					"on",
					"off"
				],
				"mode": "on"
				}
			],
			"result_info": {
				"page": %[1]d,
				"per_page": 1,
				"total_pages": 2,
				"count": 1,
				"total_count": 2
			}
		}`, page, packageID)

		page++
	}

	mux.HandleFunc("/zones/"+testZoneID+"/firewall/waf/packages/"+packageID+"/rules", handler)

	want := []WAFRule{
		{
			ID:          "fake_rule_id_1",
			Description: "Fake Rule Description 1",
			Priority:    "1",
			PackageID:   packageID,
			Group: struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}{
				ID:   "fake_group_id_1",
				Name: "Fake Group Name 1",
			},
			Mode:         "on",
			DefaultMode:  "",
			AllowedModes: []string{"on", "off"},
		},
		{
			ID:          "fake_rule_id_2",
			Description: "Fake Rule Description 2",
			Priority:    "2",
			PackageID:   packageID,
			Group: struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}{
				ID:   "fake_group_id_2",
				Name: "Fake Group Name 2",
			},
			Mode:         "on",
			DefaultMode:  "",
			AllowedModes: []string{"on", "off"},
		},
	}

	d, err := client.ListWAFRules(context.Background(), testZoneID, packageID)

	if assert.NoError(t, err) {
		assert.Equal(t, want, d)
	}

	_, err = client.ListWAFRules(context.Background(), testZoneID, "123")
	assert.Error(t, err)
}

func TestWAFRule(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		// JSON data from: https://api.cloudflare.com/#waf-rules-properties
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "f939de3be84e66e757adcdcb87908023",
				"description": "SQL injection prevention for SELECT statements",
				"priority": "5",
				"group": {
					"id": "de677e5818985db1285d0e80225f06e5",
					"name": "Project Honey Pot"
				},
				"package_id": "a25a9a7e9c00afc1fb2e0245519d725b",
				"allowed_modes": [
					"on",
					"off"
				],
				"mode": "on"
			}
		}`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/firewall/waf/packages/a25a9a7e9c00afc1fb2e0245519d725b/rules/f939de3be84e66e757adcdcb87908023", handler)

	want := WAFRule{
		ID:          "f939de3be84e66e757adcdcb87908023",
		Description: "SQL injection prevention for SELECT statements",
		Priority:    "5",
		PackageID:   "a25a9a7e9c00afc1fb2e0245519d725b",
		Group: struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}{
			ID:   "de677e5818985db1285d0e80225f06e5",
			Name: "Project Honey Pot",
		},
		Mode:         "on",
		DefaultMode:  "",
		AllowedModes: []string{"on", "off"},
	}

	d, err := client.WAFRule(context.Background(), testZoneID, "a25a9a7e9c00afc1fb2e0245519d725b", "f939de3be84e66e757adcdcb87908023")

	if assert.NoError(t, err) {
		assert.Equal(t, want, d)
	}

	_, err = client.ListWAFRules(context.Background(), testZoneID, "123")
	assert.Error(t, err)
}

func TestUpdateWAFRule(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		defer r.Body.Close()

		assert.Equal(t, `{"mode":"on"}`, string(body), "Expected method '{\"mode\":\"on\"}', got %s", string(body))

		w.Header().Set("content-type", "application/json")
		// JSON data from: https://api.cloudflare.com/#waf-rules-properties
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "f939de3be84e66e757adcdcb87908023",
				"description": "SQL injection prevention for SELECT statements",
				"priority": "5",
				"group": {
					"id": "de677e5818985db1285d0e80225f06e5",
					"name": "Project Honey Pot"
				},
				"package_id": "a25a9a7e9c00afc1fb2e0245519d725b",
				"allowed_modes": [
					"on",
					"off"
				],
				"mode": "on"
			}
		}`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/firewall/waf/packages/a25a9a7e9c00afc1fb2e0245519d725b/rules/f939de3be84e66e757adcdcb87908023", handler)

	want := WAFRule{
		ID:          "f939de3be84e66e757adcdcb87908023",
		Description: "SQL injection prevention for SELECT statements",
		Priority:    "5",
		PackageID:   "a25a9a7e9c00afc1fb2e0245519d725b",
		Group: struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}{
			ID:   "de677e5818985db1285d0e80225f06e5",
			Name: "Project Honey Pot",
		},
		Mode:         "on",
		DefaultMode:  "",
		AllowedModes: []string{"on", "off"},
	}

	d, err := client.UpdateWAFRule(context.Background(), testZoneID, "a25a9a7e9c00afc1fb2e0245519d725b", "f939de3be84e66e757adcdcb87908023", "on")

	if assert.NoError(t, err) {
		assert.Equal(t, want, d)
	}

	_, err = client.ListWAFRules(context.Background(), testZoneID, "123")
	assert.Error(t, err)
}
