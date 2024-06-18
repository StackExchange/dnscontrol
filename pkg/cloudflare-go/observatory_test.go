package cloudflare

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var testURL = "example.com/a/b"
var escapedTestURL = url.PathEscape(testURL)
var region = "us-central1"
var regionLabel = "Iowa, USA"
var frequency = "DAILY"
var observatoryTestID = "52cc96c9-b709-4ffe-9048-338853d3db46"
var date = time.Now().UTC()

var pageJSON = fmt.Sprintf(`
{
  "url": "%[1]s",
  "region": {
    "value": "%[2]s",
    "label": "%[3]s"
  },
  "scheduleFrequency": "%[4]s",
  "tests": [
    {
      "id": "%[5]s",
      "date": "%[6]s",
      "url": "%[1]s",
      "scheduleFrequency": "%[4]s",
      "region": {
        "value": "%[2]s",
        "label": "%[3]s"
      },
      "mobileReport": {
        "performanceScore": 100,
        "ttfb": 10,
        "fcp": 10,
        "lcp": 10,
        "tti": 10,
        "tbt": 10,
        "si": 10,
        "cls": 0.10,
        "state": "COMPLETED",
        "deviceType": "DESKTOP"
      },
      "desktopReport": {
        "performanceScore": 100,
        "ttfb": 10,
        "fcp": 10,
        "lcp": 10,
        "tti": 10,
        "tbt": 10,
        "si": 10,
        "cls": 0.10,
        "state": "COMPLETED",
        "deviceType": "DESKTOP"
      }
    }
  ]
}`, testURL, region, regionLabel, frequency, observatoryTestID, date.Format(time.RFC3339Nano))

var scheduledPageTestJSON = fmt.Sprintf(`
{
  "schedule": {
    "url": "%[1]s",
    "region": "%[2]s",
    "frequency": "%[3]s"
  },
  "test": %[4]s
}
`, testURL, region, frequency, pageTestJSON)

var scheduleJSON = fmt.Sprintf(`
{
  "url": "%[1]s",
  "region": "%[2]s",
  "frequency": "%[3]s"
}
`, testURL, region, frequency)

var pageTestJSON = fmt.Sprintf(`
{
  "id": "%[1]s",
  "date": "%[2]s",
  "url": "%[3]s",
  "region": {
    "value": "%[4]s",
    "label": "%[5]s"
  },
  "scheduleFrequency": "%[6]s",
  "mobileReport": %[7]s,
  "desktopReport": %[7]s
}`, observatoryTestID, date.Format(time.RFC3339Nano), testURL, region, regionLabel, frequency, reportJSON, reportJSON)

var reportJSON = `
{
  "state": "COMPLETED",
  "deviceType": "DESKTOP",
  "performanceScore": 100,
  "ttfb": 10,
  "fcp": 10,
  "lcp": 10,
  "tti": 10,
  "tbt": 10,
  "si": 10,
  "cls": 0.10
}
`

var report = ObservatoryLighthouseReport{
	PerformanceScore: 100,
	State:            "COMPLETED",
	DeviceType:       "DESKTOP",
	TTFB:             10,
	FCP:              10,
	LCP:              10,
	TTI:              10,
	TBT:              10,
	SI:               10,
	CLS:              0.10,
	Error:            nil,
}

var page = ObservatoryPage{
	URL: testURL,
	Region: labeledRegion{
		Value: region,
		Label: regionLabel,
	},
	ScheduleFrequency: frequency,
	Tests: []ObservatoryPageTest{
		pageTest,
	},
}
var pageTest = ObservatoryPageTest{
	ID:   observatoryTestID,
	Date: &date,
	URL:  testURL,
	Region: labeledRegion{
		Value: region,
		Label: regionLabel,
	},
	ScheduleFrequency: &frequency,
	MobileReport:      report,
	DesktopReport:     report,
}

var scheduledPageTest = ObservatoryScheduledPageTest{
	Schedule: ObservatorySchedule{
		URL:       testURL,
		Region:    region,
		Frequency: frequency,
	},
	Test: pageTest,
}

var schedule = ObservatorySchedule{
	URL:       testURL,
	Region:    region,
	Frequency: frequency,
}

func TestListObservatoryPages(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		assert.Equal(t, "/zones/"+testZoneID+"/speed_api/pages", r.URL.EscapedPath())
		fmt.Fprintf(w, `{
			  "success": true,
			  "errors": [],
			  "messages": [],
			  "result": [
			    %s
			  ]
			}
		`, pageJSON)
	}
	mux.HandleFunc("/zones/"+testZoneID+"/speed_api/pages", handler)
	want := []ObservatoryPage{
		page,
	}
	pages, err := client.ListObservatoryPages(context.Background(), ZoneIdentifier(testZoneID), ListObservatoryPagesParams{})
	if assert.NoError(t, err) {
		assert.Equal(t, want, pages)
	}
}

func TestObservatoryPageTrend(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		assert.Equal(t, "DESKTOP", r.URL.Query().Get("deviceType"))
		assert.Equal(t, "America/Chicago", r.URL.Query().Get("tz"))
		assert.Equal(t, "fcp,lcp", r.URL.Query().Get("metrics"))
		assert.Equal(t, "/zones/"+testZoneID+"/speed_api/pages/"+escapedTestURL+"/trend", r.URL.EscapedPath())
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			  "success": true,
			  "errors": [],
			  "messages": [],
			  "result": {
			    "performanceScore": [null, 100],
			    "ttfb": [null, 10],
			    "fcp": [null, 10],
			    "lcp": [null, 10],
			    "tti": [null, 10],
			    "tbt": [null, 10],
			    "si": [null, 10],
			    "cls": [null, 0.10]
			  }
			}
		`)
	}
	mux.HandleFunc("/zones/"+testZoneID+"/speed_api/pages/"+testURL+"/trend", handler)
	want := ObservatoryPageTrend{
		PerformanceScore: []*int{nil, IntPtr(100)},
		TTFB:             []*int{nil, IntPtr(10)},
		FCP:              []*int{nil, IntPtr(10)},
		LCP:              []*int{nil, IntPtr(10)},
		TTI:              []*int{nil, IntPtr(10)},
		TBT:              []*int{nil, IntPtr(10)},
		SI:               []*int{nil, IntPtr(10)},
		CLS:              []*float64{nil, Float64Ptr(0.10)},
	}
	trend, err := client.GetObservatoryPageTrend(context.Background(), ZoneIdentifier(testZoneID), GetObservatoryPageTrendParams{
		URL:        testURL,
		Region:     region,
		DeviceType: "DESKTOP",
		Start:      &date,
		End:        &date,
		Timezone:   "America/Chicago",
		Metrics:    []string{"fcp,lcp"},
	})
	if assert.NoError(t, err) {
		assert.Equal(t, &want, trend)
	}
}

func TestListObservatoryPageTests(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		assert.Equal(t, region, r.URL.Query().Get("region"))
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		assert.Equal(t, "10", r.URL.Query().Get("per_page"))
		assert.Equal(t, "/zones/"+testZoneID+"/speed_api/pages/"+escapedTestURL+"/tests", r.URL.EscapedPath())
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			  "success": true,
			  "errors": [],
			  "messages": [],
			  "result": [%s]
			}
		`, pageTestJSON)
	}
	mux.HandleFunc("/zones/"+testZoneID+"/speed_api/pages/"+testURL+"/tests", handler)
	want := []ObservatoryPageTest{
		pageTest,
	}
	tests, _, err := client.ListObservatoryPageTests(context.Background(), ZoneIdentifier(testZoneID), ListObservatoryPageTestParams{
		URL: testURL,
		ResultInfo: ResultInfo{
			Page:    1,
			PerPage: 10,
		},
		Region: region,
	})
	if assert.NoError(t, err) {
		assert.Equal(t, want, tests)
	}
}

func TestCreateObservatoryPageTest(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		b, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.True(t, strings.Contains(string(b), region))
		assert.Equal(t, "/zones/"+testZoneID+"/speed_api/pages/"+escapedTestURL+"/tests", r.URL.EscapedPath())
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			  "success": true,
			  "errors": [],
			  "messages": [],
			  "result": %s
			}
		`, pageTestJSON)
	}
	mux.HandleFunc("/zones/"+testZoneID+"/speed_api/pages/"+testURL+"/tests", handler)
	want := pageTest
	test, err := client.CreateObservatoryPageTest(context.Background(), ZoneIdentifier(testZoneID), CreateObservatoryPageTestParams{
		URL: testURL,
		Settings: CreateObservatoryPageTestSettings{
			Region: region,
		},
	})
	if assert.NoError(t, err) {
		assert.Equal(t, &want, test)
	}
}

func TestDeleteObservatoryPageTests(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		assert.Equal(t, region, r.URL.Query().Get("region"))
		assert.Equal(t, "/zones/"+testZoneID+"/speed_api/pages/"+escapedTestURL+"/tests", r.URL.EscapedPath())
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			  "success": true,
			  "errors": [],
			  "messages": [],
			  "result": {
                "count": 2
              }
			}
		`)
	}
	mux.HandleFunc("/zones/"+testZoneID+"/speed_api/pages/"+testURL+"/tests", handler)
	want := 2
	count, err := client.DeleteObservatoryPageTests(context.Background(), ZoneIdentifier(testZoneID), DeleteObservatoryPageTestsParams{
		URL:    testURL,
		Region: region,
	})
	if assert.NoError(t, err) {
		assert.Equal(t, &want, count)
	}
}

func TestGetObservatoryPageTest(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)

		assert.Equal(t, "/zones/"+testZoneID+"/speed_api/pages/"+escapedTestURL+"/tests/"+observatoryTestID, r.URL.EscapedPath())
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			  "success": true,
			  "errors": [],
			  "messages": [],
			  "result": %s
			}
		`, pageTestJSON)
	}
	mux.HandleFunc("/zones/"+testZoneID+"/speed_api/pages/"+testURL+"/tests/"+observatoryTestID, handler)
	want := pageTest
	test, err := client.GetObservatoryPageTest(context.Background(), ZoneIdentifier(testZoneID), GetObservatoryPageTestParams{
		TestID: observatoryTestID,
		URL:    testURL,
	})
	if assert.NoError(t, err) {
		assert.Equal(t, &want, test)
	}
}

func TestCreateObservatoryScheduledPageTest(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		assert.Equal(t, frequency, r.URL.Query().Get("frequency"))
		assert.Equal(t, region, r.URL.Query().Get("region"))
		assert.Equal(t, "/zones/"+testZoneID+"/speed_api/schedule/"+escapedTestURL, r.URL.EscapedPath())
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			  "success": true,
			  "errors": [],
			  "messages": [],
			  "result": %s
			}
		`, scheduledPageTestJSON)
	}
	mux.HandleFunc("/zones/"+testZoneID+"/speed_api/schedule/"+testURL, handler)
	want := scheduledPageTest
	pages, err := client.CreateObservatoryScheduledPageTest(context.Background(), ZoneIdentifier(testZoneID), CreateObservatoryScheduledPageTestParams{
		Frequency: frequency,
		URL:       testURL,
		Region:    region,
	})
	if assert.NoError(t, err) {
		assert.Equal(t, &want, pages)
	}
}

func TestObservatoryScheduledPageTest(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		assert.Equal(t, region, r.URL.Query().Get("region"))
		assert.Equal(t, "/zones/"+testZoneID+"/speed_api/schedule/"+escapedTestURL, r.URL.EscapedPath())
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			  "success": true,
			  "errors": [],
			  "messages": [],
			  "result": %s
			}
		`, scheduleJSON)
	}
	mux.HandleFunc("/zones/"+testZoneID+"/speed_api/schedule/"+testURL, handler)
	want := schedule
	schedule, err := client.GetObservatoryScheduledPageTest(context.Background(), ZoneIdentifier(testZoneID), GetObservatoryScheduledPageTestParams{
		URL:    testURL,
		Region: region,
	})
	if assert.NoError(t, err) {
		assert.Equal(t, &want, schedule)
	}
}

func TestDeleteObservatoryScheduledPageTest(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		assert.Equal(t, region, r.URL.Query().Get("region"))
		assert.Equal(t, "/zones/"+testZoneID+"/speed_api/schedule/"+escapedTestURL, r.URL.EscapedPath())
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			  "success": true,
			  "errors": [],
			  "messages": [],
			  "result": { 
                "count": 2
              }
			}
		`)
	}
	mux.HandleFunc("/zones/"+testZoneID+"/speed_api/schedule/"+testURL, handler)
	want := 2
	count, err := client.DeleteObservatoryScheduledPageTest(context.Background(), ZoneIdentifier(testZoneID), DeleteObservatoryScheduledPageTestParams{
		URL:    testURL,
		Region: region,
	})
	if assert.NoError(t, err) {
		assert.Equal(t, &want, count)
	}
}
