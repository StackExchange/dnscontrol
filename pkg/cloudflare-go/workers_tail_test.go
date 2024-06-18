package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	testScriptName = "this-is_my_script-01"
	testTailID     = "03dc9f77817b488fb26c5861ec18f791"
)

func TestWorkersTail_StartWorkersTail(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/workers/scripts/%s/tails", testAccountID, testScriptName), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
  "success": true,
  "errors": [],
  "messages": [],
  "result": {
    "id": "03dc9f77817b488fb26c5861ec18f791",
    "url": "wss://tail.developers.workers.dev/03dc9f77817b488fb26c5861ec18f791",
    "expires_at": "2021-08-20T19:15:51Z"
  }
}`)
	})

	_, err := client.StartWorkersTail(context.Background(), AccountIdentifier(testAccountID), "")
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingScriptName, err)
	}

	res, err := client.StartWorkersTail(context.Background(), AccountIdentifier(testAccountID), testScriptName)
	expiresAt, _ := time.Parse(time.RFC3339, "2021-08-20T19:15:51Z")
	want := WorkersTail{
		ID:        "03dc9f77817b488fb26c5861ec18f791",
		URL:       "wss://tail.developers.workers.dev/03dc9f77817b488fb26c5861ec18f791",
		ExpiresAt: &expiresAt,
	}

	if assert.NoError(t, err) {
		assert.Equal(t, want, res)
	}
}

func TestWorkersTail_ListWorkersTail(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/workers/scripts/%s/tails", testAccountID, testScriptName), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
  "success": true,
  "errors": [],
  "messages": [],
  "result": {
    "id": "03dc9f77817b488fb26c5861ec18f791",
    "url": "wss://tail.developers.workers.dev/03dc9f77817b488fb26c5861ec18f791",
    "expires_at": "2021-08-20T19:15:51Z"
  }
}`)
	})

	_, err := client.ListWorkersTail(context.Background(), AccountIdentifier(testAccountID), ListWorkersTailParameters{ScriptName: ""})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingScriptName, err)
	}

	res, err := client.ListWorkersTail(context.Background(), AccountIdentifier(testAccountID), ListWorkersTailParameters{ScriptName: testScriptName})
	expiresAt, _ := time.Parse(time.RFC3339, "2021-08-20T19:15:51Z")
	want := WorkersTail{
		ID:        "03dc9f77817b488fb26c5861ec18f791",
		URL:       "wss://tail.developers.workers.dev/03dc9f77817b488fb26c5861ec18f791",
		ExpiresAt: &expiresAt,
	}

	if assert.NoError(t, err) {
		assert.Equal(t, want, res)
	}
}

func TestWorkersTail_DeleteWorkersTail(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/workers/scripts/%s/tails/%s", testAccountID, testScriptName, testTailID), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
  "success": true,
  "errors": [],
  "messages": [],
}`)
	})

	err := client.DeleteWorkersTail(context.Background(), AccountIdentifier(testAccountID), "", "")
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingScriptName, err)
	}

	err = client.DeleteWorkersTail(context.Background(), AccountIdentifier(testAccountID), testScriptName, "")
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingTailID, err)
	}

	err = client.DeleteWorkersTail(context.Background(), AccountIdentifier(testAccountID), testScriptName, testTailID)
	assert.NoError(t, err)
}
