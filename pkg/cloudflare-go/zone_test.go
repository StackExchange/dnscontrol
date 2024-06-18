package cloudflare

import (
	"context"
	"crypto/md5"   //nolint:gosec
	"encoding/hex" // for generating IDs
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
)

// mockID returns a hex string of length 32, suitable for all kinds of IDs
// used in the Cloudflare API.
func mockID(seed string) string {
	arr := md5.Sum([]byte(seed)) //nolint:gosec
	return hex.EncodeToString(arr[:])
}

func mustParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		panic(err)
	}
	return t
}

func mockZone(i int) *Zone {
	zoneName := fmt.Sprintf("%d.example.com", i)
	ownerName := "Test Account"

	return &Zone{
		ID:      mockID(zoneName),
		Name:    zoneName,
		DevMode: 0,
		OriginalNS: []string{
			"linda.ns.cloudflare.com",
			"merlin.ns.cloudflare.com",
		},
		OriginalRegistrar: "cloudflare, inc. (id: 1910)",
		OriginalDNSHost:   "",
		CreatedOn:         mustParseTime("2021-07-28T05:06:20.736244Z"),
		ModifiedOn:        mustParseTime("2021-07-28T05:06:20.736244Z"),
		NameServers: []string{
			"abby.ns.cloudflare.com",
			"noel.ns.cloudflare.com",
		},
		Owner: Owner{
			ID:        mockID(ownerName),
			Email:     "",
			Name:      ownerName,
			OwnerType: "organization",
		},
		Permissions: []string{
			"#access:read",
			"#analytics:read",
			"#auditlogs:read",
			"#billing:read",
			"#dns_records:read",
			"#lb:read",
			"#legal:read",
			"#logs:read",
			"#member:read",
			"#organization:read",
			"#ssl:read",
			"#stream:read",
			"#subscription:read",
			"#waf:read",
			"#webhooks:read",
			"#worker:read",
			"#zone:read",
			"#zone_settings:read",
		},
		Plan: ZonePlan{
			ZonePlanCommon: ZonePlanCommon{
				ID:       "0feeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
				Name:     "Free Website",
				Currency: "USD",
			},
			IsSubscribed:      false,
			CanSubscribe:      false,
			LegacyID:          "free",
			LegacyDiscount:    false,
			ExternallyManaged: false,
		},
		PlanPending: ZonePlan{
			ZonePlanCommon: ZonePlanCommon{
				ID: "",
			},
			IsSubscribed:      false,
			CanSubscribe:      false,
			LegacyID:          "",
			LegacyDiscount:    false,
			ExternallyManaged: false,
		},
		Status: "active",
		Paused: false,
		Type:   "full",
		Host: struct {
			Name    string
			Website string
		}{
			Name:    "",
			Website: "",
		},
		VanityNS:    nil,
		Betas:       nil,
		DeactReason: "",
		Meta: ZoneMeta{
			PageRuleQuota:     3,
			WildcardProxiable: false,
			PhishingDetected:  false,
		},
		Account: Account{
			ID:   mockID(ownerName),
			Name: ownerName,
		},
		VerificationKey: "",
	}
}

func mockZonesResponse(total, page, start, count int) *ZonesResponse {
	zones := make([]Zone, count)
	for i := range zones {
		zones[i] = *mockZone(start + i)
	}

	return &ZonesResponse{
		Result: zones,
		ResultInfo: ResultInfo{
			Page:       page,
			PerPage:    50,
			TotalPages: (total + 49) / 50,
			Count:      count,
			Total:      total,
		},
		Response: Response{
			Success:  true,
			Errors:   []ResponseInfo{},
			Messages: []ResponseInfo{},
		},
	}
}

func TestZoneAnalyticsDashboard(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		assert.Equal(t, "2015-01-01T12:23:00Z", r.URL.Query().Get("since"))
		assert.Equal(t, "2015-01-02T12:23:00Z", r.URL.Query().Get("until"))
		assert.Equal(t, "true", r.URL.Query().Get("continuous"))

		w.Header().Set("content-type", "application/json")
		// JSON data from: https://api.cloudflare.com/#zone-analytics-properties
		fmt.Fprintf(w, `{
          "success": true,
          "errors": [],
          "messages": [],
          "result": {
            "totals": {
              "since": "2015-01-01T12:23:00Z",
              "until": "2015-01-02T12:23:00Z",
              "requests": {
                "all": 1234085328,
                "cached": 1234085328,
                "uncached": 13876154,
                "content_type": {
                  "css": 15343,
                  "html": 1234213,
                  "javascript": 318236,
                  "gif": 23178,
                  "jpeg": 1982048
                },
                "country": {
                  "US": 4181364,
                  "AG": 37298,
                  "GI": 293846
                },
                "ssl": {
                  "encrypted": 12978361,
                  "unencrypted": 781263
                },
                "http_status": {
                  "200": 13496983,
                  "301": 283,
                  "400": 187936,
                  "402": 1828,
                  "404": 1293
                }
              },
              "bandwidth": {
                "all": 213867451,
                "cached": 113205063,
                "uncached": 113205063,
                "content_type": {
                  "css": 237421,
                  "html": 1231290,
                  "javascript": 123245,
                  "gif": 1234242,
                  "jpeg": 784278
                },
                "country": {
                  "US": 123145433,
                  "AG": 2342483,
                  "GI": 984753
                },
                "ssl": {
                  "encrypted": 37592942,
                  "unencrypted": 237654192
                }
              },
              "threats": {
                "all": 23423873,
                "country": {
                  "US": 123,
                  "CN": 523423,
                  "AU": 91
                },
                "type": {
                  "user.ban.ip": 123,
                  "hot.ban.unknown": 5324,
                  "macro.chl.captchaErr": 1341,
                  "macro.chl.jschlErr": 5323
                }
              },
              "pageviews": {
                "all": 5724723,
                "search_engines": {
                  "googlebot": 35272,
                  "pingdom": 13435,
                  "bingbot": 5372,
                  "baidubot": 1345
                }
              },
              "uniques": {
                "all": 12343
              }
            },
            "timeseries": [
              {
                "since": "2015-01-01T12:23:00Z",
                "until": "2015-01-02T12:23:00Z",
                "requests": {
                  "all": 1234085328,
                  "cached": 1234085328,
                  "uncached": 13876154,
                  "content_type": {
                    "css": 15343,
                    "html": 1234213,
                    "javascript": 318236,
                    "gif": 23178,
                    "jpeg": 1982048
                  },
                  "country": {
                    "US": 4181364,
                    "AG": 37298,
                    "GI": 293846
                  },
                  "ssl": {
                    "encrypted": 12978361,
                    "unencrypted": 781263
                  },
                  "http_status": {
                    "200": 13496983,
                    "301": 283,
                    "400": 187936,
                    "402": 1828,
                    "404": 1293
                  }
                },
                "bandwidth": {
                  "all": 213867451,
                  "cached": 113205063,
                  "uncached": 113205063,
                  "content_type": {
                    "css": 237421,
                    "html": 1231290,
                    "javascript": 123245,
                    "gif": 1234242,
                    "jpeg": 784278
                  },
                  "country": {
                    "US": 123145433,
                    "AG": 2342483,
                    "GI": 984753
                  },
                  "ssl": {
                    "encrypted": 37592942,
                    "unencrypted": 237654192
                  }
                },
                "threats": {
                  "all": 23423873,
                  "country": {
                    "US": 123,
                    "CN": 523423,
                    "AU": 91
                  },
                  "type": {
                    "user.ban.ip": 123,
                    "hot.ban.unknown": 5324,
                    "macro.chl.captchaErr": 1341,
                    "macro.chl.jschlErr": 5323
                  }
                },
                "pageviews": {
                  "all": 5724723,
                  "search_engines": {
                    "googlebot": 35272,
                    "pingdom": 13435,
                    "bingbot": 5372,
                    "baidubot": 1345
                  }
                },
                "uniques": {
                  "all": 12343
                }
              }
            ]
          },
          "query": {
            "since": "2015-01-01T12:23:00Z",
            "until": "2015-01-02T12:23:00Z",
            "time_delta": 60
          }
        }`)
	}

	mux.HandleFunc("/zones/foo/analytics/dashboard", handler)

	since, _ := time.Parse(time.RFC3339, "2015-01-01T12:23:00Z")
	until, _ := time.Parse(time.RFC3339, "2015-01-02T12:23:00Z")
	data := ZoneAnalytics{
		Since: since,
		Until: until,
		Requests: struct {
			All         int            `json:"all"`
			Cached      int            `json:"cached"`
			Uncached    int            `json:"uncached"`
			ContentType map[string]int `json:"content_type"`
			Country     map[string]int `json:"country"`
			SSL         struct {
				Encrypted   int `json:"encrypted"`
				Unencrypted int `json:"unencrypted"`
			} `json:"ssl"`
			HTTPStatus map[string]int `json:"http_status"`
		}{
			All:      1234085328,
			Cached:   1234085328,
			Uncached: 13876154,
			ContentType: map[string]int{
				"css":        15343,
				"html":       1234213,
				"javascript": 318236,
				"gif":        23178,
				"jpeg":       1982048,
			},
			Country: map[string]int{
				"US": 4181364,
				"AG": 37298,
				"GI": 293846,
			},
			SSL: struct {
				Encrypted   int `json:"encrypted"`
				Unencrypted int `json:"unencrypted"`
			}{
				Encrypted:   12978361,
				Unencrypted: 781263,
			},
			HTTPStatus: map[string]int{
				"200": 13496983,
				"301": 283,
				"400": 187936,
				"402": 1828,
				"404": 1293,
			},
		},
		Bandwidth: struct {
			All         int            `json:"all"`
			Cached      int            `json:"cached"`
			Uncached    int            `json:"uncached"`
			ContentType map[string]int `json:"content_type"`
			Country     map[string]int `json:"country"`
			SSL         struct {
				Encrypted   int `json:"encrypted"`
				Unencrypted int `json:"unencrypted"`
			} `json:"ssl"`
		}{
			All:      213867451,
			Cached:   113205063,
			Uncached: 113205063,
			ContentType: map[string]int{
				"css":        237421,
				"html":       1231290,
				"javascript": 123245,
				"gif":        1234242,
				"jpeg":       784278,
			},
			Country: map[string]int{
				"US": 123145433,
				"AG": 2342483,
				"GI": 984753,
			},
			SSL: struct {
				Encrypted   int `json:"encrypted"`
				Unencrypted int `json:"unencrypted"`
			}{
				Encrypted:   37592942,
				Unencrypted: 237654192,
			},
		},
		Threats: struct {
			All     int            `json:"all"`
			Country map[string]int `json:"country"`
			Type    map[string]int `json:"type"`
		}{
			All: 23423873,
			Country: map[string]int{
				"US": 123,
				"CN": 523423,
				"AU": 91,
			},
			Type: map[string]int{
				"user.ban.ip":          123,
				"hot.ban.unknown":      5324,
				"macro.chl.captchaErr": 1341,
				"macro.chl.jschlErr":   5323,
			},
		},
		Pageviews: struct {
			All           int            `json:"all"`
			SearchEngines map[string]int `json:"search_engines"`
		}{
			All: 5724723,
			SearchEngines: map[string]int{
				"googlebot": 35272,
				"pingdom":   13435,
				"bingbot":   5372,
				"baidubot":  1345,
			},
		},
		Uniques: struct {
			All int `json:"all"`
		}{
			All: 12343,
		},
	}
	want := ZoneAnalyticsData{
		Totals:     data,
		Timeseries: []ZoneAnalytics{data},
	}

	continuous := true
	d, err := client.ZoneAnalyticsDashboard(context.Background(), "foo", ZoneAnalyticsOptions{
		Since:      &since,
		Until:      &until,
		Continuous: &continuous,
	})
	if assert.NoError(t, err) {
		assert.Equal(t, want, d)
	}

	_, err = client.ZoneAnalyticsDashboard(context.Background(), "bar", ZoneAnalyticsOptions{})
	assert.Error(t, err)
}

func TestZoneAnalyticsByColocation(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		assert.Equal(t, "2015-01-01T12:23:00Z", r.URL.Query().Get("since"))
		assert.Equal(t, "2015-01-02T12:23:00Z", r.URL.Query().Get("until"))
		assert.Equal(t, "true", r.URL.Query().Get("continuous"))

		w.Header().Set("content-type", "application/json")
		// JSON data from: https://api.cloudflare.com/#zone-analytics-analytics-by-co-locations
		fmt.Fprintf(w, `{
          "success": true,
          "errors": [],
          "messages": [],
          "result": [
            {
              "colo_id": "SFO",
              "timeseries": [
                {
                  "since": "2015-01-01T12:23:00Z",
                  "until": "2015-01-02T12:23:00Z",
                  "requests": {
                    "all": 1234085328,
                    "cached": 1234085328,
                    "uncached": 13876154,
                    "content_type": {
                      "css": 15343,
                      "html": 1234213,
                      "javascript": 318236,
                      "gif": 23178,
                      "jpeg": 1982048
                    },
                    "country": {
                      "US": 4181364,
                      "AG": 37298,
                      "GI": 293846
                    },
                    "ssl": {
                      "encrypted": 12978361,
                      "unencrypted": 781263
                    },
                    "http_status": {
                      "200": 13496983,
                      "301": 283,
                      "400": 187936,
                      "402": 1828,
                      "404": 1293
                    }
                  },
                  "bandwidth": {
                    "all": 213867451,
                    "cached": 113205063,
                    "uncached": 113205063,
                    "content_type": {
                      "css": 237421,
                      "html": 1231290,
                      "javascript": 123245,
                      "gif": 1234242,
                      "jpeg": 784278
                    },
                    "country": {
                      "US": 123145433,
                      "AG": 2342483,
                      "GI": 984753
                    },
                    "ssl": {
                      "encrypted": 37592942,
                      "unencrypted": 237654192
                    }
                  },
                  "threats": {
                    "all": 23423873,
                    "country": {
                      "US": 123,
                      "CN": 523423,
                      "AU": 91
                    },
                    "type": {
                      "user.ban.ip": 123,
                      "hot.ban.unknown": 5324,
                      "macro.chl.captchaErr": 1341,
                      "macro.chl.jschlErr": 5323
                    }
                  },
                  "pageviews": {
                    "all": 5724723,
                    "search_engines": {
                      "googlebot": 35272,
                      "pingdom": 13435,
                      "bingbot": 5372,
                      "baidubot": 1345
                    }
                  },
                  "uniques": {
                    "all": 12343
                  }
                }
              ]
            }
          ],
          "query": {
            "since": "2015-01-01T12:23:00Z",
            "until": "2015-01-02T12:23:00Z",
            "time_delta": 60
          }
        }`)
	}

	mux.HandleFunc("/zones/foo/analytics/colos", handler)

	since, _ := time.Parse(time.RFC3339, "2015-01-01T12:23:00Z")
	until, _ := time.Parse(time.RFC3339, "2015-01-02T12:23:00Z")
	data := ZoneAnalytics{
		Since: since,
		Until: until,
		Requests: struct {
			All         int            `json:"all"`
			Cached      int            `json:"cached"`
			Uncached    int            `json:"uncached"`
			ContentType map[string]int `json:"content_type"`
			Country     map[string]int `json:"country"`
			SSL         struct {
				Encrypted   int `json:"encrypted"`
				Unencrypted int `json:"unencrypted"`
			} `json:"ssl"`
			HTTPStatus map[string]int `json:"http_status"`
		}{
			All:      1234085328,
			Cached:   1234085328,
			Uncached: 13876154,
			ContentType: map[string]int{
				"css":        15343,
				"html":       1234213,
				"javascript": 318236,
				"gif":        23178,
				"jpeg":       1982048,
			},
			Country: map[string]int{
				"US": 4181364,
				"AG": 37298,
				"GI": 293846,
			},
			SSL: struct {
				Encrypted   int `json:"encrypted"`
				Unencrypted int `json:"unencrypted"`
			}{
				Encrypted:   12978361,
				Unencrypted: 781263,
			},
			HTTPStatus: map[string]int{
				"200": 13496983,
				"301": 283,
				"400": 187936,
				"402": 1828,
				"404": 1293,
			},
		},
		Bandwidth: struct {
			All         int            `json:"all"`
			Cached      int            `json:"cached"`
			Uncached    int            `json:"uncached"`
			ContentType map[string]int `json:"content_type"`
			Country     map[string]int `json:"country"`
			SSL         struct {
				Encrypted   int `json:"encrypted"`
				Unencrypted int `json:"unencrypted"`
			} `json:"ssl"`
		}{
			All:      213867451,
			Cached:   113205063,
			Uncached: 113205063,
			ContentType: map[string]int{
				"css":        237421,
				"html":       1231290,
				"javascript": 123245,
				"gif":        1234242,
				"jpeg":       784278,
			},
			Country: map[string]int{
				"US": 123145433,
				"AG": 2342483,
				"GI": 984753,
			},
			SSL: struct {
				Encrypted   int `json:"encrypted"`
				Unencrypted int `json:"unencrypted"`
			}{
				Encrypted:   37592942,
				Unencrypted: 237654192,
			},
		},
		Threats: struct {
			All     int            `json:"all"`
			Country map[string]int `json:"country"`
			Type    map[string]int `json:"type"`
		}{
			All: 23423873,
			Country: map[string]int{
				"US": 123,
				"CN": 523423,
				"AU": 91,
			},
			Type: map[string]int{
				"user.ban.ip":          123,
				"hot.ban.unknown":      5324,
				"macro.chl.captchaErr": 1341,
				"macro.chl.jschlErr":   5323,
			},
		},
		Pageviews: struct {
			All           int            `json:"all"`
			SearchEngines map[string]int `json:"search_engines"`
		}{
			All: 5724723,
			SearchEngines: map[string]int{
				"googlebot": 35272,
				"pingdom":   13435,
				"bingbot":   5372,
				"baidubot":  1345,
			},
		},
		Uniques: struct {
			All int `json:"all"`
		}{
			All: 12343,
		},
	}
	want := []ZoneAnalyticsColocation{
		{
			ColocationID: "SFO",
			Timeseries:   []ZoneAnalytics{data},
		},
	}

	continuous := true
	d, err := client.ZoneAnalyticsByColocation(context.Background(), "foo", ZoneAnalyticsOptions{
		Since:      &since,
		Until:      &until,
		Continuous: &continuous,
	})
	if assert.NoError(t, err) {
		assert.Equal(t, want, d)
	}

	_, err = client.ZoneAnalyticsDashboard(context.Background(), "bar", ZoneAnalyticsOptions{})
	assert.Error(t, err)
}

func TestWithPagination(t *testing.T) {
	opt := reqOption{
		params: url.Values{},
	}
	popts := PaginationOptions{
		Page:    45,
		PerPage: 500,
	}
	of := WithPagination(popts)
	of(&opt)

	tests := []struct {
		name     string
		expected string
	}{
		{"page", "45"},
		{"per_page", "500"},
	}

	for _, tt := range tests {
		if got := opt.params.Get(tt.name); got != tt.expected {
			t.Errorf("expected param %s to be %s, got %s", tt.name, tt.expected, got)
		}
	}
}

func TestZoneFilters(t *testing.T) {
	opt := reqOption{
		params: url.Values{},
	}
	of := WithZoneFilters("example.org", "", "")
	of(&opt)

	if got := opt.params.Get("name"); got != "example.org" {
		t.Errorf("expected param %s to be %s, got %s", "name", "example.org", got)
	}
}

var createdAndModifiedOn, _ = time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
var expectedFullZoneSetup = Zone{
	ID:      "023e105f4ecef8ad9ca31a8372d0c353",
	Name:    "example.com",
	DevMode: 7200,
	OriginalNS: []string{
		"ns1.originaldnshost.com",
		"ns2.originaldnshost.com",
	},
	OriginalRegistrar: "GoDaddy",
	OriginalDNSHost:   "NameCheap",
	CreatedOn:         createdAndModifiedOn,
	ModifiedOn:        createdAndModifiedOn,
	Owner: Owner{
		ID:        "7c5dae5552338874e5053f2534d2767a",
		Email:     "user@example.com",
		OwnerType: "user",
	},
	Account: Account{
		ID:   "01a7362d577a6c3019a474fd6f485823",
		Name: "Demo Account",
	},
	Permissions: []string{"#zone:read", "#zone:edit"},
	Plan: ZonePlan{
		ZonePlanCommon: ZonePlanCommon{
			ID:        "e592fd9519420ba7405e1307bff33214",
			Name:      "Pro Plan",
			Price:     20,
			Currency:  "USD",
			Frequency: "monthly",
		},
		LegacyID:     "pro",
		IsSubscribed: true,
		CanSubscribe: true,
	},
	PlanPending: ZonePlan{
		ZonePlanCommon: ZonePlanCommon{
			ID:        "e592fd9519420ba7405e1307bff33214",
			Name:      "Pro Plan",
			Price:     20,
			Currency:  "USD",
			Frequency: "monthly",
		},
		LegacyID:     "pro",
		IsSubscribed: true,
		CanSubscribe: true,
	},
	Status:      "active",
	Paused:      false,
	Type:        "full",
	NameServers: []string{"tony.ns.cloudflare.com", "woz.ns.cloudflare.com"},
}
var expectedPartialZoneSetup = Zone{
	ID:      "023e105f4ecef8ad9ca31a8372d0c353",
	Name:    "example.com",
	DevMode: 7200,
	OriginalNS: []string{
		"ns1.originaldnshost.com",
		"ns2.originaldnshost.com",
	},
	OriginalRegistrar: "GoDaddy",
	OriginalDNSHost:   "NameCheap",
	CreatedOn:         createdAndModifiedOn,
	ModifiedOn:        createdAndModifiedOn,
	Owner: Owner{
		ID:        "7c5dae5552338874e5053f2534d2767a",
		Email:     "user@example.com",
		OwnerType: "user",
	},
	Account: Account{
		ID:   "01a7362d577a6c3019a474fd6f485823",
		Name: "Demo Account",
	},
	Permissions: []string{"#zone:read", "#zone:edit"},
	Plan: ZonePlan{
		ZonePlanCommon: ZonePlanCommon{
			ID:        "e592fd9519420ba7405e1307bff33214",
			Name:      "Pro Plan",
			Price:     20,
			Currency:  "USD",
			Frequency: "monthly",
		},
		LegacyID:     "pro",
		IsSubscribed: true,
		CanSubscribe: true,
	},
	PlanPending: ZonePlan{
		ZonePlanCommon: ZonePlanCommon{
			ID:        "e592fd9519420ba7405e1307bff33214",
			Name:      "Pro Plan",
			Price:     20,
			Currency:  "USD",
			Frequency: "monthly",
		},
		LegacyID:     "pro",
		IsSubscribed: true,
		CanSubscribe: true,
	},
	Status:      "active",
	Paused:      false,
	Type:        "partial",
	NameServers: []string{"tony.ns.cloudflare.com", "woz.ns.cloudflare.com"},
}

func TestCreateZoneFullSetup(t *testing.T) {
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
				"id": "023e105f4ecef8ad9ca31a8372d0c353",
				"name": "example.com",
				"development_mode": 7200,
				"original_name_servers": [
					"ns1.originaldnshost.com",
					"ns2.originaldnshost.com"
				],
				"original_registrar": "GoDaddy",
				"original_dnshost": "NameCheap",
				"created_on": "2014-01-01T05:20:00.12345Z",
				"modified_on": "2014-01-01T05:20:00.12345Z",
				"activated_on": "2014-01-02T00:01:00.12345Z",
				"owner": {
					"id": "7c5dae5552338874e5053f2534d2767a",
					"email": "user@example.com",
					"type": "user"
				},
				"account": {
					"id": "01a7362d577a6c3019a474fd6f485823",
					"name": "Demo Account"
				},
				"permissions": [
					"#zone:read",
					"#zone:edit"
				],
				"plan": {
					"id": "e592fd9519420ba7405e1307bff33214",
					"name": "Pro Plan",
					"price": 20,
					"currency": "USD",
					"frequency": "monthly",
					"legacy_id": "pro",
					"is_subscribed": true,
					"can_subscribe": true
				},
				"plan_pending": {
					"id": "e592fd9519420ba7405e1307bff33214",
					"name": "Pro Plan",
					"price": 20,
					"currency": "USD",
					"frequency": "monthly",
					"legacy_id": "pro",
					"is_subscribed": true,
					"can_subscribe": true
				},
				"status": "active",
				"paused": false,
				"type": "full",
				"name_servers": [
					"tony.ns.cloudflare.com",
					"woz.ns.cloudflare.com"
				]
			}
		}
		`)
	}

	mux.HandleFunc("/zones", handler)

	actual, err := client.CreateZone(context.Background(), "example.com", false, Account{ID: "01a7362d577a6c3019a474fd6f485823"}, "full")

	if assert.NoError(t, err) {
		assert.Equal(t, expectedFullZoneSetup, actual)
	}
}

func TestCreateZonePartialSetup(t *testing.T) {
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
				"id": "023e105f4ecef8ad9ca31a8372d0c353",
				"name": "example.com",
				"development_mode": 7200,
				"original_name_servers": [
					"ns1.originaldnshost.com",
					"ns2.originaldnshost.com"
				],
				"original_registrar": "GoDaddy",
				"original_dnshost": "NameCheap",
				"created_on": "2014-01-01T05:20:00.12345Z",
				"modified_on": "2014-01-01T05:20:00.12345Z",
				"activated_on": "2014-01-02T00:01:00.12345Z",
				"owner": {
					"id": "7c5dae5552338874e5053f2534d2767a",
					"email": "user@example.com",
					"type": "user"
				},
				"account": {
					"id": "01a7362d577a6c3019a474fd6f485823",
					"name": "Demo Account"
				},
				"permissions": [
					"#zone:read",
					"#zone:edit"
				],
				"plan": {
					"id": "e592fd9519420ba7405e1307bff33214",
					"name": "Pro Plan",
					"price": 20,
					"currency": "USD",
					"frequency": "monthly",
					"legacy_id": "pro",
					"is_subscribed": true,
					"can_subscribe": true
				},
				"plan_pending": {
					"id": "e592fd9519420ba7405e1307bff33214",
					"name": "Pro Plan",
					"price": 20,
					"currency": "USD",
					"frequency": "monthly",
					"legacy_id": "pro",
					"is_subscribed": true,
					"can_subscribe": true
				},
				"status": "active",
				"paused": false,
				"type": "partial",
				"name_servers": [
					"tony.ns.cloudflare.com",
					"woz.ns.cloudflare.com"
				]
			}
		}
		`)
	}

	mux.HandleFunc("/zones", handler)

	actual, err := client.CreateZone(context.Background(), "example.com", false, Account{ID: "01a7362d577a6c3019a474fd6f485823"}, "partial")

	if assert.NoError(t, err) {
		assert.Equal(t, expectedPartialZoneSetup, actual)
	}
}

func TestFallbackOrigin_FallbackOrigin(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/zones/foo/fallback_origin", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
"success": true,
"errors": [],
"messages": [],
"result": {
    "id": "fallback_origin",
    "value": "app.example.com",
    "editable": true
  }
}`)
	})

	fallbackOrigin, err := client.FallbackOrigin(context.Background(), "foo")

	want := FallbackOrigin{
		ID:    "fallback_origin",
		Value: "app.example.com",
	}

	if assert.NoError(t, err) {
		assert.Equal(t, want, fallbackOrigin)
	}
}

func TestFallbackOrigin_UpdateFallbackOrigin(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/zones/foo/fallback_origin", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `
{
  "success": true,
  "errors": [],
  "messages": [],
  "result": {
    "id": "fallback_origin",
    "value": "app.example.com",
		"editable": true
  }
}`)
	})

	response, err := client.UpdateFallbackOrigin(context.Background(), "foo", FallbackOrigin{Value: "app.example.com"})

	want := &FallbackOriginResponse{
		Result: FallbackOrigin{
			ID:    "fallback_origin",
			Value: "app.example.com",
		},
		Response: Response{Success: true, Errors: []ResponseInfo{}, Messages: []ResponseInfo{}},
	}

	if assert.NoError(t, err) {
		assert.Equal(t, want, response)
	}
}

func Test_normalizeZoneName(t *testing.T) {
	tests := []struct {
		name     string
		zone     string
		expected string
	}{
		{
			name:     "unicode stays unicode",
			zone:     "ünì¢øðe.tld",
			expected: "ünì¢øðe.tld",
		}, {
			name:     "valid punycode is normalized to unicode",
			zone:     "xn--ne-7ca90ava1cya.tld",
			expected: "ünì¢øðe.tld",
		}, {
			name:     "valid punycode in second label",
			zone:     "example.xn--j6w193g",
			expected: "example.香港",
		}, {
			name:     "invalid punycode is returned without change",
			zone:     "xn-invalid.xn-invalid-tld",
			expected: "xn-invalid.xn-invalid-tld",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := normalizeZoneName(tt.zone)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestZonePartialHasVerificationKey(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		// JSON data from: https://api.cloudflare.com/#zone-zone-details (plus an undocumented field verification_key from curl to API)
		fmt.Fprintf(w, `{
  "result": {
    "id": "foo",
    "name": "bar",
    "status": "active",
    "paused": false,
    "type": "partial",
    "development_mode": 0,
    "verification_key": "foo-bar",
    "original_name_servers": ["a","b","c","d"],
    "original_registrar": null,
    "original_dnshost": null,
    "modified_on": "2019-09-04T15:11:43.409805Z",
    "created_on": "2018-12-06T14:33:38.410126Z",
    "activated_on": "2018-12-06T14:34:39.274528Z",
    "meta": {
      "step": 4,
      "wildcard_proxiable": true,
      "custom_certificate_quota": 1,
      "page_rule_quota": 100,
      "phishing_detected": false,
      "multiple_railguns_allowed": false
    },
    "owner": {
      "id": "bbbbbbbbbbbbbbbbbbbbbbbb",
      "type": "organization",
      "name": "OrgName"
    },
    "account": {
      "id": "aaaaaaaaaaaaaaaaaaaaaaaa",
      "name": "AccountName"
    },
    "permissions": [
      "#access:edit",
      "#access:read",
      "#analytics:read",
      "#app:edit",
      "#auditlogs:read",
      "#billing:read",
      "#cache_purge:edit",
      "#dns_records:edit",
      "#dns_records:read",
      "#lb:edit",
      "#lb:read",
      "#legal:read",
      "#logs:edit",
      "#logs:read",
      "#member:read",
      "#organization:edit",
      "#organization:read",
      "#ssl:edit",
      "#ssl:read",
      "#stream:edit",
      "#stream:read",
      "#subscription:edit",
      "#subscription:read",
      "#waf:edit",
      "#waf:read",
      "#webhooks:edit",
      "#webhooks:read",
      "#worker:edit",
      "#worker:read",
      "#zone:edit",
      "#zone:read",
      "#zone_settings:edit",
      "#zone_settings:read"
    ],
    "plan": {
      "id": "94f3b7b768b0458b56d2cac4fe5ec0f9",
      "name": "Enterprise Website",
      "price": 0,
      "currency": "USD",
      "frequency": "monthly",
      "is_subscribed": true,
      "can_subscribe": true,
      "legacy_id": "enterprise",
      "legacy_discount": false,
      "externally_managed": true
    }
  },
  "success": true,
  "errors": [],
  "messages": []
}`)
	}

	mux.HandleFunc("/zones/foo", handler)

	z, err := client.ZoneDetails(context.Background(), "foo")
	if assert.NoError(t, err) {
		assert.NotEmpty(t, z.VerificationKey)
		assert.Equal(t, z.VerificationKey, "foo-bar")
	}
}

func TestZoneDNSSECSetting(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		// JSON data from: https://api.cloudflare.com/#dnssec-properties
		fmt.Fprintf(w, `{
			"result": {
				"status": "active",
				"flags": 257,
				"algorithm": "13",
				"key_type": "ECDSAP256SHA256",
				"digest_type": "2",
				"digest_algorithm": "SHA256",
				"digest": "48E939042E82C22542CB377B580DFDC52A361CEFDC72E7F9107E2B6BD9306A45",
				"ds": "example.com. 3600 IN DS 16953 13 2 48E939042E82C22542CB377B580DFDC52A361CEFDC72E7F9107E2B6BD9306A45",
				"key_tag": 42,
				"public_key": "oXiGYrSTO+LSCJ3mohc8EP+CzF9KxBj8/ydXJ22pKuZP3VAC3/Md/k7xZfz470CoRyZJ6gV6vml07IC3d8xqhA==",
				"modified_on": "2014-01-01T05:20:00Z"
  			}
		}`)
	}

	mux.HandleFunc("/zones/foo/dnssec", handler)

	z, err := client.ZoneDNSSECSetting(context.Background(), "foo")
	if assert.NoError(t, err) {
		assert.Equal(t, z.Status, "active")
		assert.Equal(t, z.Flags, 257)
		assert.Equal(t, z.Algorithm, "13")
		assert.Equal(t, z.KeyType, "ECDSAP256SHA256")
		assert.Equal(t, z.DigestType, "2")
		assert.Equal(t, z.DigestAlgorithm, "SHA256")
		assert.Equal(t, z.Digest, "48E939042E82C22542CB377B580DFDC52A361CEFDC72E7F9107E2B6BD9306A45")
		assert.Equal(t, z.DS, "example.com. 3600 IN DS 16953 13 2 48E939042E82C22542CB377B580DFDC52A361CEFDC72E7F9107E2B6BD9306A45")
		assert.Equal(t, z.KeyTag, 42)
		assert.Equal(t, z.PublicKey, "oXiGYrSTO+LSCJ3mohc8EP+CzF9KxBj8/ydXJ22pKuZP3VAC3/Md/k7xZfz470CoRyZJ6gV6vml07IC3d8xqhA==")
		time, _ := time.Parse("2006-01-02T15:04:05Z", "2014-01-01T05:20:00Z")
		assert.Equal(t, z.ModifiedOn, time)
	}
}

func TestDeleteZoneDNSSEC(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		// JSON data from: https://api.cloudflare.com/#dnssec-properties
		fmt.Fprintf(w, `{
			"result": "foo"
		}`)
	}

	mux.HandleFunc("/zones/foo/dnssec", handler)

	z, err := client.DeleteZoneDNSSEC(context.Background(), "foo")
	if assert.NoError(t, err) {
		assert.Equal(t, z, "foo")
	}
}

func TestUpdateZoneDNSSEC(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		// JSON data from: https://api.cloudflare.com/#dnssec-properties
		fmt.Fprintf(w, `{
			"result": {
				"status": "active",
				"flags": 257,
				"algorithm": "13",
				"key_type": "ECDSAP256SHA256",
				"digest_type": "2",
				"digest_algorithm": "SHA256",
				"digest": "48E939042E82C22542CB377B580DFDC52A361CEFDC72E7F9107E2B6BD9306A45",
				"ds": "example.com. 3600 IN DS 16953 13 2 48E939042E82C22542CB377B580DFDC52A361CEFDC72E7F9107E2B6BD9306A45",
				"key_tag": 42,
				"public_key": "oXiGYrSTO+LSCJ3mohc8EP+CzF9KxBj8/ydXJ22pKuZP3VAC3/Md/k7xZfz470CoRyZJ6gV6vml07IC3d8xqhA==",
				"modified_on": "2014-01-01T05:20:00Z"
  			}
		}`)
	}

	mux.HandleFunc("/zones/foo/dnssec", handler)

	z, err := client.UpdateZoneDNSSEC(context.Background(), "foo", ZoneDNSSECUpdateOptions{
		Status: "active",
	})
	if assert.NoError(t, err) {
		assert.Equal(t, z.Status, "active")
		assert.Equal(t, z.Flags, 257)
		assert.Equal(t, z.Algorithm, "13")
		assert.Equal(t, z.KeyType, "ECDSAP256SHA256")
		assert.Equal(t, z.DigestType, "2")
		assert.Equal(t, z.DigestAlgorithm, "SHA256")
		assert.Equal(t, z.Digest, "48E939042E82C22542CB377B580DFDC52A361CEFDC72E7F9107E2B6BD9306A45")
		assert.Equal(t, z.DS, "example.com. 3600 IN DS 16953 13 2 48E939042E82C22542CB377B580DFDC52A361CEFDC72E7F9107E2B6BD9306A45")
		assert.Equal(t, z.KeyTag, 42)
		assert.Equal(t, z.PublicKey, "oXiGYrSTO+LSCJ3mohc8EP+CzF9KxBj8/ydXJ22pKuZP3VAC3/Md/k7xZfz470CoRyZJ6gV6vml07IC3d8xqhA==")
		time, _ := time.Parse("2006-01-02T15:04:05Z", "2014-01-01T05:20:00Z")
		assert.Equal(t, z.ModifiedOn, time)
	}
}

func TestZoneSetType(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"result": {
				"type": "partial",
				"verification_key": "000000000-000000000",
				"modified_on": "2014-01-01T05:20:00Z"
				}
		}`)
	}

	mux.HandleFunc("/zones/foo", handler)

	z, err := client.ZoneSetType(context.Background(), "foo", "partial")
	if assert.NoError(t, err) {
		assert.Equal(t, z.Type, "partial")
		assert.Equal(t, z.VerificationKey, "000000000-000000000")
		time, _ := time.Parse("2006-01-02T15:04:05Z", "2014-01-01T05:20:00Z")
		assert.Equal(t, z.ModifiedOn, time)
	}
}

func parsePage(t *testing.T, total int, s string) (int, bool) {
	if s == "" {
		return 1, true
	}

	page, err := strconv.Atoi(s)
	if !assert.NoError(t, err) {
		return 0, false
	}

	if !assert.LessOrEqual(t, page, total) || !assert.GreaterOrEqual(t, page, 1) {
		return 0, false
	}

	return page, true
}

func TestListZones(t *testing.T) {
	setup()
	defer teardown()

	const (
		total     = 392
		totalPage = (total + 49) / 50
	)

	handler := func(w http.ResponseWriter, r *http.Request) {
		switch {
		case !assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method):
			return
		case !assert.Equal(t, "50", r.URL.Query().Get("per_page")):
			return
		}

		page, ok := parsePage(t, totalPage, r.URL.Query().Get("page"))
		if !ok {
			return
		}

		start := (page - 1) * 50

		count := 50
		if page == totalPage {
			count = total - start
		}

		w.Header().Set("content-type", "application/json")
		err := json.NewEncoder(w).Encode(mockZonesResponse(total, page, start, count))
		assert.NoError(t, err)
	}

	mux.HandleFunc("/zones", handler)

	zones, err := client.ListZones(context.Background())
	if !assert.NoError(t, err) || !assert.Equal(t, total, len(zones)) {
		return
	}

	for i, zone := range zones {
		assert.Equal(t, *mockZone(i), zone)
	}
}

func TestListZonesFailingPages(t *testing.T) {
	setup()
	defer teardown()

	const (
		total     = 1489
		totalPage = (total + 49) / 50
	)

	// the pages to reject
	isReject := func(i int) bool { return i == 4 || i == 7 }

	handler := func(w http.ResponseWriter, r *http.Request) {
		switch {
		case !assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method):
			return
		case !assert.Equal(t, "50", r.URL.Query().Get("per_page")):
			return
		}

		page, ok := parsePage(t, totalPage, r.URL.Query().Get("page"))
		switch {
		case !ok:
			return
		case isReject(page):
			return
		}

		start := (page - 1) * 50

		count := 50
		if page == totalPage {
			count = total - start
		}

		w.Header().Set("content-type", "application/json")
		err := json.NewEncoder(w).Encode(mockZonesResponse(total, page, start, count))
		assert.NoError(t, err)
	}

	mux.HandleFunc("/zones", handler)

	_, err := client.ListZones(context.Background())
	assert.Error(t, err)
}

func TestListZonesContextManualPagination1(t *testing.T) {
	_, err := client.ListZonesContext(context.Background(), WithPagination(PaginationOptions{Page: 2}))
	assert.EqualError(t, err, errManualPagination)
}

func TestListZonesContextManualPagination2(t *testing.T) {
	_, err := client.ListZonesContext(context.Background(), WithPagination(PaginationOptions{PerPage: 30}))
	assert.EqualError(t, err, errManualPagination)
}

func TestUpdateZoneSSLSettings(t *testing.T) {
	setup()
	defer teardown()
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		// JSON data from: https://api.cloudflare.com/#zone-settings-properties
		_, _ = fmt.Fprintf(w, `{
			"result": {
				"id": "ssl",
				"value": "off",
				"editable": true,
				"modified_on": "2014-01-01T05:20:00.12345Z"
			}
		}`)
	}
	mux.HandleFunc("/zones/foo/settings/ssl", handler)
	s, err := client.UpdateZoneSSLSettings(context.Background(), "foo", "off")
	if assert.NoError(t, err) {
		assert.Equal(t, s.ID, "ssl")
		assert.Equal(t, s.Value, "off")
		assert.Equal(t, s.Editable, true)
		assert.Equal(t, s.ModifiedOn, "2014-01-01T05:20:00.12345Z")
	}
}

func TestGetZoneSetting(t *testing.T) {
	setup()
	defer teardown()
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		_, _ = fmt.Fprintf(w, `{
			"result": {
				"id": "ssl",
				"value": "off",
				"editable": true,
				"modified_on": "2014-01-01T05:20:00.12345Z"
			}
		}`)
	}
	mux.HandleFunc("/zones/foo/settings/ssl", handler)
	s, err := client.GetZoneSetting(context.Background(), ZoneIdentifier("foo"), GetZoneSettingParams{Name: "ssl"})
	if assert.NoError(t, err) {
		assert.Equal(t, s.ID, "ssl")
		assert.Equal(t, s.Value, "off")
		assert.Equal(t, s.Editable, true)
		assert.Equal(t, s.ModifiedOn, "2014-01-01T05:20:00.12345Z")
	}
}

func TestGetZoneSettingWithCustomPathPrefix(t *testing.T) {
	setup()
	defer teardown()
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		_, _ = fmt.Fprintf(w, `{
			"result": {
				"id": "ssl",
				"value": "off",
				"editable": true,
				"modified_on": "2014-01-01T05:20:00.12345Z"
			}
		}`)
	}
	mux.HandleFunc("/zones/foo/my_custom_path/ssl", handler)
	s, err := client.GetZoneSetting(context.Background(), ZoneIdentifier("foo"), GetZoneSettingParams{Name: "ssl", PathPrefix: "my_custom_path"})
	if assert.NoError(t, err) {
		assert.Equal(t, s.ID, "ssl")
		assert.Equal(t, s.Value, "off")
		assert.Equal(t, s.Editable, true)
		assert.Equal(t, s.ModifiedOn, "2014-01-01T05:20:00.12345Z")
	}
}

func TestUpdateZoneSetting(t *testing.T) {
	setup()
	defer teardown()
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		_, _ = fmt.Fprintf(w, `{
			"result": {
				"id": "ssl",
				"value": "off",
				"editable": true,
				"modified_on": "2014-01-01T05:20:00.12345Z"
			}
		}`)
	}
	mux.HandleFunc("/zones/foo/settings/ssl", handler)
	s, err := client.UpdateZoneSetting(context.Background(), ZoneIdentifier("foo"), UpdateZoneSettingParams{Name: "ssl", Value: "off"})
	if assert.NoError(t, err) {
		assert.Equal(t, s.ID, "ssl")
		assert.Equal(t, s.Value, "off")
		assert.Equal(t, s.Editable, true)
		assert.Equal(t, s.ModifiedOn, "2014-01-01T05:20:00.12345Z")
	}
}

func TestUpdateZoneSettingWithCustomPathPrefix(t *testing.T) {
	setup()
	defer teardown()
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		_, _ = fmt.Fprintf(w, `{
			"result": {
				"id": "ssl",
				"value": "off",
				"editable": true,
				"modified_on": "2014-01-01T05:20:00.12345Z"
			}
		}`)
	}
	mux.HandleFunc("/zones/foo/my_custom_path/ssl", handler)
	s, err := client.UpdateZoneSetting(context.Background(), ZoneIdentifier("foo"), UpdateZoneSettingParams{Name: "ssl", PathPrefix: "my_custom_path", Value: "off"})
	if assert.NoError(t, err) {
		assert.Equal(t, s.ID, "ssl")
		assert.Equal(t, s.Value, "off")
		assert.Equal(t, s.Editable, true)
		assert.Equal(t, s.ModifiedOn, "2014-01-01T05:20:00.12345Z")
	}
}
