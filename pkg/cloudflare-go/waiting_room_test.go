package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"time"

	"github.com/stretchr/testify/assert"
)

var waitingRoomID = "699d98642c564d2e855e9661899b7252"
var waitingRoomEventID = "25756b2dfe6e378a06b033b670413757"
var waitingRoomRuleID = "25756b2dfe6e378a06b033b670413757"
var testTimestampWaitingRoom = time.Now().UTC()
var testTimestampWaitingRoomEvent = time.Now().UTC()
var testTimestampWaitingRoomEventPrequeue = time.Now().UTC()
var testTimestampWaitingRoomEventStart = testTimestampWaitingRoomEventPrequeue.Add(5 * time.Minute)
var testTimestampWaitingRoomEventEnd = testTimestampWaitingRoomEventStart.Add(1 * time.Minute)
var waitingRoomJSON = fmt.Sprintf(`
    {
      "id": "%s",
      "created_on": "%s",
      "modified_on": "%s",
      "name": "production_webinar",
      "description": "Production - DO NOT MODIFY",
      "queueing_method": "random",
      "suspended": false,
      "host": "shop.example.com",
      "path": "/shop/checkout",
      "queue_all": true,
      "new_users_per_minute": 600,
      "total_active_users": 1000,
      "session_duration": 10,
      "disable_session_renewal": false,
      "json_response_enabled": true,
      "custom_page_html": "{{#waitTimeKnown}} {{waitTime}} mins {{/waitTimeKnown}} {{^waitTimeKnown}} Queue all enabled {{/waitTimeKnown}}",
      "default_template_language": "en-US",
      "next_event_prequeue_start_time": null,
      "next_event_start_time": "%s",
	  "cookie_suffix": "example_shop",
	  "additional_routes": [{"host": "shop2.example.com", "path": "/shop/checkout"}],
	  "queueing_status_code": 200
    }
   `, waitingRoomID, testTimestampWaitingRoom.Format(time.RFC3339Nano), testTimestampWaitingRoom.Format(time.RFC3339Nano),
	testTimestampWaitingRoomEventStart.Format(time.RFC3339Nano))

var waitingRoomEventJSON = fmt.Sprintf(`
    {
      "id": "%s",
      "created_on": "%s",
      "modified_on": "%s",
      "name": "production_webinar_event",
      "description": "Production event - DO NOT MODIFY",
      "suspended": false,
      "prequeue_start_time": "%s",
      "event_start_time": "%s",
      "event_end_time": "%s",
      "shuffle_at_event_start": false,
      "new_users_per_minute": 2000,
      "total_active_users": 2500,
      "session_duration": null,
      "disable_session_renewal": null,
      "queueing_method": "random",
      "custom_page_html": "{{#waitTimeKnown}} {{waitTime}} mins {{/waitTimeKnown}} {{^waitTimeKnown}} Event is prequeueing / Queue all enabled {{/waitTimeKnown}}"
    }
   `, waitingRoomEventID, testTimestampWaitingRoomEvent.Format(time.RFC3339Nano),
	testTimestampWaitingRoomEvent.Format(time.RFC3339Nano),
	testTimestampWaitingRoomEventPrequeue.Format(time.RFC3339Nano),
	testTimestampWaitingRoomEventStart.Format(time.RFC3339Nano),
	testTimestampWaitingRoomEventEnd.Format(time.RFC3339Nano))

var waitingRoomStatusJSON = fmt.Sprintf(`
    {
      "status": "queueing",
      "event_id": "%s",
      "estimated_queued_users": 10,
      "estimated_total_active_users": 9,
      "max_estimated_time_minutes": 5
    }
   `, waitingRoomEventID)

var waitingRoomRuleJSON = fmt.Sprintf(`
{
  "id": "%s",
  "version": "1",
  "description": "bypass ip",
  "action": "bypass_waiting_room",
  "expression": "ip.src in {1.2.3.4 5.6.7.8}",
  "enabled": true,
  "last_updated": "%s"
}
`, waitingRoomRuleID, testTimestampWaitingRoom.Format(time.RFC3339Nano))

var waitingRoomPagePreviewJSON = `
    {
      "preview_url": "http://waitingrooms.dev/preview/35af8c12-6d68-4608-babb-b53435a5ddfb"
    }
    `

var waitingRoomSettingsJSON = `
    {
      "search_engine_crawler_bypass": true
    }
   `

var waitingRoom = WaitingRoom{
	ID:                         waitingRoomID,
	CreatedOn:                  testTimestampWaitingRoom,
	ModifiedOn:                 testTimestampWaitingRoom,
	Name:                       "production_webinar",
	Description:                "Production - DO NOT MODIFY",
	QueueingMethod:             "random",
	Suspended:                  false,
	Host:                       "shop.example.com",
	Path:                       "/shop/checkout",
	QueueAll:                   true,
	NewUsersPerMinute:          600,
	TotalActiveUsers:           1000,
	SessionDuration:            10,
	DisableSessionRenewal:      false,
	JsonResponseEnabled:        true,
	CustomPageHTML:             "{{#waitTimeKnown}} {{waitTime}} mins {{/waitTimeKnown}} {{^waitTimeKnown}} Queue all enabled {{/waitTimeKnown}}",
	DefaultTemplateLanguage:    "en-US",
	NextEventStartTime:         &testTimestampWaitingRoomEventStart,
	NextEventPrequeueStartTime: nil,
	CookieSuffix:               "example_shop",
	AdditionalRoutes:           []*WaitingRoomRoute{{Host: "shop2.example.com", Path: "/shop/checkout"}},
	QueueingStatusCode:         200,
}

var waitingRoomEvent = WaitingRoomEvent{
	ID:                    waitingRoomEventID,
	CreatedOn:             testTimestampWaitingRoomEvent,
	ModifiedOn:            testTimestampWaitingRoomEvent,
	Name:                  "production_webinar_event",
	Description:           "Production event - DO NOT MODIFY",
	Suspended:             false,
	PrequeueStartTime:     &testTimestampWaitingRoomEventPrequeue,
	EventStartTime:        testTimestampWaitingRoomEventStart,
	EventEndTime:          testTimestampWaitingRoomEventEnd,
	ShuffleAtEventStart:   false,
	NewUsersPerMinute:     2000,
	TotalActiveUsers:      2500,
	SessionDuration:       0,
	DisableSessionRenewal: nil,
	QueueingMethod:        "random",
	CustomPageHTML:        "{{#waitTimeKnown}} {{waitTime}} mins {{/waitTimeKnown}} {{^waitTimeKnown}} Event is prequeueing / Queue all enabled {{/waitTimeKnown}}",
}

var waitingRoomStatus = WaitingRoomStatus{
	Status:                    "queueing",
	EventID:                   waitingRoomEventID,
	EstimatedQueuedUsers:      10,
	EstimatedTotalActiveUsers: 9,
	MaxEstimatedTimeMinutes:   5,
}

var waitingRoomPagePreview = WaitingRoomPagePreviewURL{
	PreviewURL: "http://waitingrooms.dev/preview/35af8c12-6d68-4608-babb-b53435a5ddfb",
}

var waitingRoomRule = WaitingRoomRule{
	ID:          waitingRoomRuleID,
	Version:     "1",
	Action:      "bypass_waiting_room",
	Expression:  "ip.src in {1.2.3.4 5.6.7.8}",
	Description: "bypass ip",
	Enabled:     BoolPtr(true),
	LastUpdated: &testTimestampWaitingRoom,
}

var waitingRoomSettings = WaitingRoomSettings{
	SearchEngineCrawlerBypass: true,
}

var waitingRoomSettingsUpdate = UpdateWaitingRoomSettingsParams{
	SearchEngineCrawlerBypass: BoolPtr(true),
}

var waitingRoomSettingsPatch = PatchWaitingRoomSettingsParams{
	SearchEngineCrawlerBypass: BoolPtr(true),
}

func TestListWaitingRooms(t *testing.T) {
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
			    %s
			  ]
			}
		`, waitingRoomJSON)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/waiting_rooms", handler)
	want := []WaitingRoom{waitingRoom}

	actual, err := client.ListWaitingRooms(context.Background(), testZoneID)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestListWaitingRoomsNoResult(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			  "success": true,
			  "errors": [],
			  "messages": [],
			  "result": []
			}
		`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/waiting_rooms", handler)
	want := []WaitingRoom{}

	actual, err := client.ListWaitingRooms(context.Background(), testZoneID)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestWaitingRoom(t *testing.T) {
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
		`, waitingRoomJSON)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/waiting_rooms/699d98642c564d2e855e9661899b7252", handler)
	want := waitingRoom

	actual, err := client.WaitingRoom(context.Background(), testZoneID, "699d98642c564d2e855e9661899b7252")
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestWaitingRoomNotFound(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			  "success": null,
			  "errors": [
			  	{
      			"code": 1001,
      			"message": "Object not found."
    			}
    		],
			  "messages": []
			}
		`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/waiting_rooms/699d98642c564d2e855e9661899b7252", handler)

	_, err := client.WaitingRoom(context.Background(), testZoneID, "699d98642c564d2e855e9661899b7252")
	assert.NotNil(t, err)
}

func TestCreateWaitingRoom(t *testing.T) {
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
		`, waitingRoomJSON)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/waiting_rooms", handler)
	want := &waitingRoom

	actual, err := client.CreateWaitingRoom(context.Background(), testZoneID, waitingRoom)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateWaitingRoomError(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			 "success": false,
			 "errors": [
			  	{
      			"code": 1002,
      			"message": "new_users_per_minute must be in range [200, total_active_users]: invalid data"
    			}
    		],
			  "messages": []
			}
		`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/waiting_rooms", handler)

	_, err := client.CreateWaitingRoom(context.Background(), testZoneID, waitingRoom)
	assert.NotNil(t, err)
}

func TestUpdateWaitingRoom(t *testing.T) {
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
		`, waitingRoomJSON)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/waiting_rooms/699d98642c564d2e855e9661899b7252", handler)
	want := waitingRoom

	actual, err := client.UpdateWaitingRoom(context.Background(), testZoneID, waitingRoom)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestChangeWaitingRoom(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			  "success": true,
			  "errors": [],
			  "messages": [],
			  "result": %s
			}
		`, waitingRoomJSON)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/waiting_rooms/699d98642c564d2e855e9661899b7252", handler)
	want := waitingRoom

	actual, err := client.ChangeWaitingRoom(context.Background(), testZoneID, "699d98642c564d2e855e9661899b7252", WaitingRoom{TotalActiveUsers: 400})
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestDeleteWaitingRoom(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			  "success": true,
			  "errors": [],
			  "messages": [],
			  "result": {
			    "id": "699d98642c564d2e855e9661899b7252"
			  }
			}
		`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/waiting_rooms/699d98642c564d2e855e9661899b7252", handler)

	err := client.DeleteWaitingRoom(context.Background(), testZoneID, "699d98642c564d2e855e9661899b7252")
	assert.NoError(t, err)
}

func TestWaitingRoomStatus(t *testing.T) {
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
		`, waitingRoomStatusJSON)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/waiting_rooms/699d98642c564d2e855e9661899b7252/status", handler)
	want := waitingRoomStatus

	actual, err := client.WaitingRoomStatus(context.Background(), testZoneID, "699d98642c564d2e855e9661899b7252")
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestWaitingRoomPagePreview(t *testing.T) {
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
		`, waitingRoomPagePreviewJSON)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/waiting_rooms/preview", handler)
	want := waitingRoomPagePreview

	actual, err := client.WaitingRoomPagePreview(context.Background(), testZoneID, "{{#waitTimeKnown}} {{waitTime}} mins {{/waitTimeKnown}} {{^waitTimeKnown}} Queue all enabled {{/waitTimeKnown}}")
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateWaitingRoomEvent(t *testing.T) {
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
		`, waitingRoomEventJSON)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/waiting_rooms/699d98642c564d2e855e9661899b7252/events", handler)
	want := &waitingRoomEvent

	actual, err := client.CreateWaitingRoomEvent(context.Background(), testZoneID, "699d98642c564d2e855e9661899b7252", waitingRoomEvent)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestListWaitingRoomEvents(t *testing.T) {
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
			    %s
			  ]
			}
		`, waitingRoomEventJSON)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/waiting_rooms/699d98642c564d2e855e9661899b7252/events", handler)
	want := []WaitingRoomEvent{waitingRoomEvent}

	actual, err := client.ListWaitingRoomEvents(context.Background(), testZoneID, "699d98642c564d2e855e9661899b7252")
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestListWaitingRoomEventsNoResult(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			  "success": true,
			  "errors": [],
			  "messages": [],
			  "result": []
			}
		`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/waiting_rooms/699d98642c564d2e855e9661899b7252/events", handler)
	want := []WaitingRoomEvent{}

	actual, err := client.ListWaitingRoomEvents(context.Background(), testZoneID, "699d98642c564d2e855e9661899b7252")
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestWaitingRoomEvent(t *testing.T) {
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
		`, waitingRoomEventJSON)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/waiting_rooms/699d98642c564d2e855e9661899b7252/events/25756b2dfe6e378a06b033b670413757", handler)
	want := waitingRoomEvent

	actual, err := client.WaitingRoomEvent(context.Background(), testZoneID, "699d98642c564d2e855e9661899b7252", "25756b2dfe6e378a06b033b670413757")
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestWaitingRoomEventPreview(t *testing.T) {
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
		`, waitingRoomEventJSON)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/waiting_rooms/699d98642c564d2e855e9661899b7252/events/25756b2dfe6e378a06b033b670413757/details", handler)
	want := waitingRoomEvent

	actual, err := client.WaitingRoomEventPreview(context.Background(), testZoneID, "699d98642c564d2e855e9661899b7252", "25756b2dfe6e378a06b033b670413757")
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestUpdateWaitingRoomEvent(t *testing.T) {
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
		`, waitingRoomEventJSON)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/waiting_rooms/699d98642c564d2e855e9661899b7252/events/25756b2dfe6e378a06b033b670413757", handler)
	want := waitingRoomEvent

	actual, err := client.UpdateWaitingRoomEvent(context.Background(), testZoneID, "699d98642c564d2e855e9661899b7252", waitingRoomEvent)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestChangeWaitingRoomEvent(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PATCH', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			  "success": true,
			  "errors": [],
			  "messages": [],
			  "result": %s
			}
		`, waitingRoomEventJSON)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/waiting_rooms/699d98642c564d2e855e9661899b7252/events/25756b2dfe6e378a06b033b670413757", handler)
	want := waitingRoomEvent

	actual, err := client.UpdateWaitingRoomEvent(context.Background(), testZoneID, "699d98642c564d2e855e9661899b7252", waitingRoomEvent)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestDeleteWaitingRoomEvent(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			  "success": true,
			  "errors": [],
			  "messages": [],
			  "result": {
			    "id": "25756b2dfe6e378a06b033b670413757"
			  }
			}
		`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/waiting_rooms/699d98642c564d2e855e9661899b7252/events/25756b2dfe6e378a06b033b670413757", handler)

	err := client.DeleteWaitingRoomEvent(context.Background(), testZoneID, "699d98642c564d2e855e9661899b7252", "25756b2dfe6e378a06b033b670413757")
	assert.NoError(t, err)
}

func TestListWaitingRoomRules(t *testing.T) {
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
			    %s
			  ]
			}
		`, waitingRoomRuleJSON)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/waiting_rooms/699d98642c564d2e855e9661899b7252/rules", handler)
	want := []WaitingRoomRule{waitingRoomRule}

	actual, err := client.ListWaitingRoomRules(context.Background(), ZoneIdentifier(testZoneID), ListWaitingRoomRuleParams{WaitingRoomID: "699d98642c564d2e855e9661899b7252"})
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateWaitingRoomRule(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			  "success": true,
			  "errors": [],
			  "messages": [],
			  "result": [
			    %s
			  ]
			}
		`, waitingRoomRuleJSON)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/waiting_rooms/699d98642c564d2e855e9661899b7252/rules", handler)
	want := []WaitingRoomRule{waitingRoomRule}

	actual, err := client.CreateWaitingRoomRule(context.Background(), ZoneIdentifier(testZoneID), CreateWaitingRoomRuleParams{
		WaitingRoomID: "699d98642c564d2e855e9661899b7252",
		Rule:          waitingRoomRule,
	})
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestUpdateWaitingRoomRule(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			  "success": true,
			  "errors": [],
			  "messages": [],
			  "result": [
			    %s
			  ]
			}
		`, waitingRoomRuleJSON)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/waiting_rooms/699d98642c564d2e855e9661899b7252/rules/"+waitingRoomRuleID, handler)
	want := []WaitingRoomRule{waitingRoomRule}

	actual, err := client.UpdateWaitingRoomRule(context.Background(), ZoneIdentifier(testZoneID), UpdateWaitingRoomRuleParams{
		WaitingRoomID: "699d98642c564d2e855e9661899b7252",
		Rule:          waitingRoomRule,
	})
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestDeleteWaitingRoomRule(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			  "success": true,
			  "errors": [],
			  "messages": [],
			  "result": []
			}
		`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/waiting_rooms/699d98642c564d2e855e9661899b7252/rules/"+waitingRoomRuleID, handler)
	want := []WaitingRoomRule{}

	actual, err := client.DeleteWaitingRoomRule(context.Background(), ZoneIdentifier(testZoneID), DeleteWaitingRoomRuleParams{
		WaitingRoomID: "699d98642c564d2e855e9661899b7252",
		RuleID:        waitingRoomRuleID,
	})
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestReplaceWaitingRoomRules(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			  "success": true,
			  "errors": [],
			  "messages": [],
			  "result": [
			    %s
			  ]
			}
		`, waitingRoomRuleJSON)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/waiting_rooms/699d98642c564d2e855e9661899b7252/rules", handler)
	want := []WaitingRoomRule{waitingRoomRule}

	actual, err := client.ReplaceWaitingRoomRules(context.Background(), ZoneIdentifier(testZoneID), ReplaceWaitingRoomRuleParams{
		WaitingRoomID: "699d98642c564d2e855e9661899b7252",
		Rules:         want,
	})
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestWaitingRoomSettings(t *testing.T) {
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
		`, waitingRoomSettingsJSON)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/waiting_rooms/settings", handler)
	want := waitingRoomSettings

	actual, err := client.GetWaitingRoomSettings(context.Background(), ZoneIdentifier(testZoneID))
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestUpdateWaitingRoomSettings(t *testing.T) {
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
		`, waitingRoomSettingsJSON)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/waiting_rooms/settings", handler)
	want := waitingRoomSettings

	actual, err := client.UpdateWaitingRoomSettings(context.Background(), ZoneIdentifier(testZoneID), waitingRoomSettingsUpdate)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestChangeWaitingRoomSettings(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			  "success": true,
			  "errors": [],
			  "messages": [],
			  "result": %s
			}
		`, waitingRoomSettingsJSON)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/waiting_rooms/settings", handler)
	want := waitingRoomSettings

	actual, err := client.PatchWaitingRoomSettings(context.Background(), ZoneIdentifier(testZoneID), waitingRoomSettingsPatch)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}
