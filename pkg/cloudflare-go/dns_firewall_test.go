package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDNSFirewallUserAnalytics_UserLevel(t *testing.T) {
	setup()
	defer teardown()

	now := time.Now().UTC()
	since := now.Add(-1 * time.Hour)
	until := now

	handler := func(w http.ResponseWriter, r *http.Request) {
		expectedMetrics := "queryCount,uncachedCount,staleCount,responseTimeAvg,responseTimeMedia,responseTime90th,responseTime99th"

		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET'")
		assert.Equal(t, expectedMetrics, r.URL.Query().Get("metrics"), "Expected many metrics in URL parameter")
		assert.Equal(t, since.Format(time.RFC3339), r.URL.Query().Get("since"), "Expected since parameter in URL")
		assert.Equal(t, until.Format(time.RFC3339), r.URL.Query().Get("until"), "Expected until parameter in URL")

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
		  "result": {
			"totals":{
				"queryCount": 5,
				"uncachedCount":6,
				"staleCount":7,
				"responseTimeAvg":1.0,
				"responseTimeMedian":2.0,
				"responseTime90th":3.0,
				"responseTime99th":4.0
			  }
		  },
		  "success": true,
		  "errors": null,
		  "messages": null
		}`)
	}

	mux.HandleFunc("/user/dns_firewall/12345/dns_analytics/report", handler)
	want := DNSFirewallAnalytics{
		Totals: DNSFirewallAnalyticsMetrics{
			QueryCount:         Int64Ptr(5),
			UncachedCount:      Int64Ptr(6),
			StaleCount:         Int64Ptr(7),
			ResponseTimeAvg:    Float64Ptr(1.0),
			ResponseTimeMedian: Float64Ptr(2.0),
			ResponseTime90th:   Float64Ptr(3.0),
			ResponseTime99th:   Float64Ptr(4.0),
		},
	}

	actual, err := client.GetDNSFirewallUserAnalytics(context.Background(), UserIdentifier("foo"), GetDNSFirewallUserAnalyticsParams{ClusterID: "12345", DNSFirewallUserAnalyticsOptions: DNSFirewallUserAnalyticsOptions{
		Metrics: []string{
			"queryCount",
			"uncachedCount",
			"staleCount",
			"responseTimeAvg",
			"responseTimeMedia",
			"responseTime90th",
			"responseTime99th",
		},
		Since: &since,
		Until: &until,
	}})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestDNSFirewallUserAnalytics_AccountLevel(t *testing.T) {
	setup()
	defer teardown()

	now := time.Now().UTC()
	since := now.Add(-1 * time.Hour)
	until := now

	handler := func(w http.ResponseWriter, r *http.Request) {
		expectedMetrics := "queryCount,uncachedCount,staleCount,responseTimeAvg,responseTimeMedia,responseTime90th,responseTime99th"

		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET'")
		assert.Equal(t, expectedMetrics, r.URL.Query().Get("metrics"), "Expected many metrics in URL parameter")
		assert.Equal(t, since.Format(time.RFC3339), r.URL.Query().Get("since"), "Expected since parameter in URL")
		assert.Equal(t, until.Format(time.RFC3339), r.URL.Query().Get("until"), "Expected until parameter in URL")

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
		  "result": {
			"totals":{
				"queryCount": 5,
				"uncachedCount":6,
				"staleCount":7,
				"responseTimeAvg":1.0,
				"responseTimeMedian":2.0,
				"responseTime90th":3.0,
				"responseTime99th":4.0
			  }
		  },
		  "success": true,
		  "errors": null,
		  "messages": null
		}`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/dns_firewall/12345/dns_analytics/report", handler)
	want := DNSFirewallAnalytics{
		Totals: DNSFirewallAnalyticsMetrics{
			QueryCount:         Int64Ptr(5),
			UncachedCount:      Int64Ptr(6),
			StaleCount:         Int64Ptr(7),
			ResponseTimeAvg:    Float64Ptr(1.0),
			ResponseTimeMedian: Float64Ptr(2.0),
			ResponseTime90th:   Float64Ptr(3.0),
			ResponseTime99th:   Float64Ptr(4.0),
		},
	}

	actual, err := client.GetDNSFirewallUserAnalytics(context.Background(), AccountIdentifier(testAccountID), GetDNSFirewallUserAnalyticsParams{ClusterID: "12345", DNSFirewallUserAnalyticsOptions: DNSFirewallUserAnalyticsOptions{
		Metrics: []string{
			"queryCount",
			"uncachedCount",
			"staleCount",
			"responseTimeAvg",
			"responseTimeMedia",
			"responseTime90th",
			"responseTime99th",
		},
		Since: &since,
		Until: &until,
	}})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}
