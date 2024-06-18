package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const testTurnstileWidgetSiteKey = "0x4AAF00AAAABn0R22HWm-YUc"

var (
	turnstileWidgetCreatedOn, _  = time.Parse(time.RFC3339, "2014-01-01T05:20:00.123123Z")
	turnstileWidgetModifiedOn, _ = time.Parse(time.RFC3339, "2014-01-01T05:20:00.123123Z")
	expectedTurnstileWidget      = TurnstileWidget{
		SiteKey:    "0x4AAF00AAAABn0R22HWm-YUc",
		Secret:     "0x4AAF00AAAABn0R22HWm098HVBjhdsYUc",
		CreatedOn:  &turnstileWidgetCreatedOn,
		ModifiedOn: &turnstileWidgetModifiedOn,
		Name:       "blog.cloudflare.com login form",
		Domains: []string{
			"203.0.113.1",
			"cloudflare.com",
			"blog.example.com",
		},
		Mode:         "invisible",
		BotFightMode: true,
		Region:       "world",
		OffLabel:     false,
	}
)

func TestTurnstileWidget_Create(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/challenges/widgets", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `
			{
			  "success": true,
			  "errors": [],
			  "messages": [],
			  "result": {
				"sitekey": "0x4AAF00AAAABn0R22HWm-YUc",
				"secret": "0x4AAF00AAAABn0R22HWm098HVBjhdsYUc",
				"created_on": "2014-01-01T05:20:00.123123Z",
				"modified_on": "2014-01-01T05:20:00.123123Z",
				"name": "blog.cloudflare.com login form",
				"domains": [
				  "203.0.113.1",
				  "cloudflare.com",
				  "blog.example.com"
				],
				"mode": "invisible",
				"bot_fight_mode": true,
				"region": "world",
				"offlabel": false
			  }
			}`)
	})

	// Make sure missing account ID is thrown
	_, err := client.CreateTurnstileWidget(context.Background(), AccountIdentifier(""), CreateTurnstileWidgetParams{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	out, err := client.CreateTurnstileWidget(context.Background(), AccountIdentifier(testAccountID), CreateTurnstileWidgetParams{
		Name:         "blog.cloudflare.com login form",
		Mode:         "invisible",
		BotFightMode: true,
		Domains: []string{
			"203.0.113.1",
			"cloudflare.com",
			"blog.example.com",
		},
		Region:   "world",
		OffLabel: false,
	})
	if assert.NoError(t, err) {
		assert.Equal(t, expectedTurnstileWidget, out, "create challenge_widgets structs not equal")
	}
}

func TestTurnstileWidget_List(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/challenges/widgets", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		assert.Equal(t, "asc", r.URL.Query().Get("order"))
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `
{
  "success": true,
  "errors": [],
  "messages": [],
  "result": [
    {
      "sitekey": "0x4AAF00AAAABn0R22HWm-YUc",
      "secret": "0x4AAF00AAAABn0R22HWm098HVBjhdsYUc",
      "created_on": "2014-01-01T05:20:00.123123Z",
      "modified_on": "2014-01-01T05:20:00.123123Z",
      "name": "blog.cloudflare.com login form",
      "domains": [
        "203.0.113.1",
        "cloudflare.com",
        "blog.example.com"
      ],
      "mode": "invisible",
      "bot_fight_mode": true,
	  "region": "world",
	  "offlabel": false
    }
  ],
  "result_info": {
    "page": 1,
    "per_page": 20,
    "count": 1,
    "total_count": 1
  }
}`)
	})

	_, _, err := client.ListTurnstileWidgets(context.Background(), AccountIdentifier(""), ListTurnstileWidgetParams{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	out, results, err := client.ListTurnstileWidgets(context.Background(), AccountIdentifier(testAccountID), ListTurnstileWidgetParams{
		Order: OrderDirectionAsc,
	})
	if assert.NoError(t, err) {
		assert.Equal(t, 1, len(out), "expected 1 challenge_widgets")
		assert.Equal(t, 20, results.PerPage, "expected 20 per page")
		assert.Equal(t, expectedTurnstileWidget, out[0], "list challenge_widgets structs not equal")
	}
}

func TestTurnstileWidget_Get(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/challenges/widgets/"+testTurnstileWidgetSiteKey, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `
{
  "success": true,
  "errors": [],
  "messages": [],
  "result": {
    "sitekey": "0x4AAF00AAAABn0R22HWm-YUc",
    "secret": "0x4AAF00AAAABn0R22HWm098HVBjhdsYUc",
    "created_on": "2014-01-01T05:20:00.123123Z",
    "modified_on": "2014-01-01T05:20:00.123123Z",
    "name": "blog.cloudflare.com login form",
    "domains": [
      "203.0.113.1",
      "cloudflare.com",
      "blog.example.com"
    ],
    "mode": "invisible",
	"bot_fight_mode": true,
	"region": "world",
	"offlabel": false
  }
}`)
	})

	_, err := client.GetTurnstileWidget(context.Background(), AccountIdentifier(""), "")
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	_, err = client.GetTurnstileWidget(context.Background(), AccountIdentifier(testAccountID), "")
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingSiteKey, err)
	}

	out, err := client.GetTurnstileWidget(context.Background(), AccountIdentifier(testAccountID), testTurnstileWidgetSiteKey)

	if assert.NoError(t, err) {
		assert.Equal(t, expectedTurnstileWidget, out, "get challenge_widgets structs not equal")
	}
}

func TestTurnstileWidgets_Update(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/challenges/widgets/"+testTurnstileWidgetSiteKey, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `
{
  "success": true,
  "errors": [],
  "messages": [],
  "result": {
    "sitekey": "0x4AAF00AAAABn0R22HWm-YUc",
    "secret": "0x4AAF00AAAABn0R22HWm098HVBjhdsYUc",
    "created_on": "2014-01-01T05:20:00.123123Z",
    "modified_on": "2014-01-01T05:20:00.123123Z",
    "name": "blog.cloudflare.com login form",
    "domains": [
      "203.0.113.1",
      "cloudflare.com",
      "blog.example.com"
    ],
    "mode": "invisible",
	"bot_fight_mode": true,
	"region": "world",
	"offlabel": false
  }
}`)
	})

	_, err := client.UpdateTurnstileWidget(context.Background(), AccountIdentifier(""), UpdateTurnstileWidgetParams{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	_, err = client.UpdateTurnstileWidget(context.Background(), AccountIdentifier(testAccountID), UpdateTurnstileWidgetParams{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingSiteKey, err)
	}

	out, err := client.UpdateTurnstileWidget(context.Background(), AccountIdentifier(testAccountID), UpdateTurnstileWidgetParams{
		SiteKey: testTurnstileWidgetSiteKey,
	})
	if assert.NoError(t, err) {
		assert.Equal(t, expectedTurnstileWidget, out, "update challenge_widgets structs not equal")
	}
}

func TestTurnstileWidgets_RotateSecret(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/challenges/widgets/"+testTurnstileWidgetSiteKey+"/rotate_secret", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `
{
  "success": true,
  "errors": [],
  "messages": [],
  "result": {
    "sitekey": "0x4AAF00AAAABn0R22HWm-YUc",
    "secret": "0x4AAF00AAAABn0R22HWm098HVBjhdsYUc",
    "created_on": "2014-01-01T05:20:00.123123Z",
    "modified_on": "2014-01-01T05:20:00.123123Z",
    "name": "blog.cloudflare.com login form",
    "domains": [
      "203.0.113.1",
      "cloudflare.com",
      "blog.example.com"
    ],
    "mode": "invisible",
	"bot_fight_mode": true,
	"region": "world",
	"offlabel": false
  }
}`)
	})

	// Make sure missing account ID is thrown
	_, err := client.RotateTurnstileWidget(context.Background(), AccountIdentifier(""), RotateTurnstileWidgetParams{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	_, err = client.RotateTurnstileWidget(context.Background(), AccountIdentifier(testAccountID), RotateTurnstileWidgetParams{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingSiteKey, err)
	}

	out, err := client.RotateTurnstileWidget(context.Background(), AccountIdentifier(testAccountID), RotateTurnstileWidgetParams{SiteKey: testTurnstileWidgetSiteKey})
	if assert.NoError(t, err) {
		assert.Equal(t, expectedTurnstileWidget, out, "rotate challenge_widgets structs not equal")
	}
}

func TestTurnstileWidgets_Delete(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/challenges/widgets/"+testTurnstileWidgetSiteKey, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `
{
  "success": true,
  "errors": [],
  "messages": [],
  "result": {
    "sitekey": "0x4AAF00AAAABn0R22HWm-YUc",
    "secret": "0x4AAF00AAAABn0R22HWm098HVBjhdsYUc",
    "created_on": "2014-01-01T05:20:00.123123Z",
    "modified_on": "2014-01-01T05:20:00.123123Z",
    "name": "blog.cloudflare.com login form",
    "domains": [
      "203.0.113.1",
      "cloudflare.com",
      "blog.example.com"
    ],
    "mode": "invisible",
	"bot_fight_mode": true,
	"region": "world",
	"offlabel": false
  }
}`)
	})

	// Make sure missing account ID is thrown
	err := client.DeleteTurnstileWidget(context.Background(), AccountIdentifier(""), "")
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	err = client.DeleteTurnstileWidget(context.Background(), AccountIdentifier(testAccountID), "")
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingSiteKey, err)
	}

	err = client.DeleteTurnstileWidget(context.Background(), AccountIdentifier(testAccountID), testTurnstileWidgetSiteKey)
	assert.NoError(t, err)
}
