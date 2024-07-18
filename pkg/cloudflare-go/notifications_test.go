package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/goccy/go-json"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testWebhookID = "fe49ee055d23404e9d58f9110b210c8d"
	testPolicyID  = "6ec8a5145d0d2263a36fad55c03cb43d"
)

var (
	notificationTimestamp = time.Date(2021, 05, 01, 10, 47, 01, 01, time.UTC)
)

func TestGetEligibleNotificationDestinations(t *testing.T) {
	setup()
	defer teardown()

	expected := NotificationMechanisms{
		Email:     NotificationMechanismMetaData{true, true, "email"},
		PagerDuty: NotificationMechanismMetaData{true, true, "pagerduty"},
		Webhooks:  NotificationMechanismMetaData{true, true, "webhooks"},
	}
	b, err := json.Marshal(expected)
	require.NoError(t, err)
	require.NotEmpty(t, b)

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
  "success": true,
  "errors": [],
  "messages": [],
  "result":%s
}`, string(b))
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/alerting/v3/destinations/eligible", handler)

	actual, err := client.GetEligibleNotificationDestinations(context.Background(), testAccountID)
	require.Nil(t, err)
	require.NotNil(t, actual)
	assert.Equal(t, expected, actual.Result)
}
func TestGetAvailableNotificationTypes(t *testing.T) {
	setup()
	defer teardown()

	expected := make(NotificationsGroupedByProduct, 1)
	alert1 := NotificationAlertWithDescription{Type: "secondary_dns_zone_successfully_updated", DisplayName: "Secondary DNS Successfully Updated", Description: "Secondary zone transfers are succeeding, the zone has been updated."}
	alert2 := NotificationAlertWithDescription{Type: "secondary_dns_zone_validation_warning", DisplayName: "Secondary DNSSEC Validation Warning", Description: "The transferred DNSSEC zone is incorrectly configured."}
	expected["DNS"] = []NotificationAlertWithDescription{alert1, alert2}

	b, err := json.Marshal(expected)
	require.NoError(t, err)
	require.NotEmpty(t, b)

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
  									"success": true,
  									"errors": [],
  									"messages": [],
  									"result":%s
								}`,
			string(b))
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/alerting/v3/available_alerts", handler)

	actual, err := client.GetAvailableNotificationTypes(context.Background(), testAccountID)
	require.Nil(t, err)
	require.NotNil(t, actual)
	assert.Equal(t, expected, actual.Result)
}
func TestListPagerDutyDestinations(t *testing.T) {
	setup()
	defer teardown()

	expected := NotificationPagerDutyResource{ID: "valid-uuid", Name: "my pagerduty connection"}
	b, err := json.Marshal(expected)
	require.NoError(t, err)
	require.NotEmpty(t, b)

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		_, err = fmt.Fprintf(w, `{
  									"success": true,
  									"errors": [],
  									"messages": [],
  									"result":%s
								}`,
			string(b))
		require.NoError(t, err)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/alerting/v3/destinations/pagerduty", handler)

	actual, err := client.ListPagerDutyNotificationDestinations(context.Background(), testAccountID)
	require.Nil(t, err)
	require.NotNil(t, actual)
	assert.Equal(t, expected, actual.Result)
}

func TestCreateNotificationPolicy(t *testing.T) {
	setup()
	defer teardown()

	mechanisms := make(map[string]NotificationMechanismIntegrations)
	mechanisms["email"] = []NotificationMechanismData{{Name: "email to send notification", ID: "test@gmail.com"}}
	policy := NotificationPolicy{
		Description: "Notifies when my zones are under attack",
		Name:        "CF DOS attack alert - L4",
		Enabled:     true,
		AlertType:   "dos_attack_l4",
		Mechanisms:  mechanisms,
		Conditions:  nil,
		Filters:     nil,
	}
	b, err := json.Marshal(policy)
	require.NoError(t, err)
	require.NotEmpty(t, b)

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		_, err = fmt.Fprintf(w, `{
  									"success": true,
  									"errors": [],
  									"messages": [],
  									"result":%s
								}`,
			string(b))
		require.NoError(t, err)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/alerting/v3/policies", handler)
	res, err := client.CreateNotificationPolicy(context.Background(), testAccountID, policy)
	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestGetNotificationPolicy(t *testing.T) {
	setup()
	defer teardown()

	mechanisms := make(map[string]NotificationMechanismIntegrations)
	mechanisms["email"] = []NotificationMechanismData{{Name: "email to send notification", ID: "test@gmail.com"}}
	policy := NotificationPolicy{
		ID:          testPolicyID,
		Description: "Notifies when my zones are under attack",
		Name:        "CF DOS attack alert - L4",
		Enabled:     true,
		AlertType:   "dos_attack_l4",
		Mechanisms:  mechanisms,
		Conditions:  nil,
		Filters:     nil,
		Created:     notificationTimestamp,
		Modified:    notificationTimestamp,
	}
	b, err := json.Marshal(policy)
	require.NoError(t, err)
	require.NotEmpty(t, b)

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		_, err = fmt.Fprintf(w, `{
  									"success": true,
  									"errors": [],
  									"messages": [],
  									"result":%s
								}`,
			string(b))
		require.NoError(t, err)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/alerting/v3/policies/"+testPolicyID, handler)

	res, err := client.GetNotificationPolicy(context.Background(), testAccountID, testPolicyID)
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, policy, res.Result)
}

func TestListNotificationPolicies(t *testing.T) {
	setup()
	defer teardown()

	mechanisms := make(map[string]NotificationMechanismIntegrations)
	mechanisms["email"] = []NotificationMechanismData{{Name: "email to send notification", ID: "test@gmail.com"}}
	policy := NotificationPolicy{
		ID:          testPolicyID,
		Description: "Notifies when my zones are under attack",
		Name:        "CF DOS attack alert - L4",
		Enabled:     true,
		AlertType:   "dos_attack_l4",
		Mechanisms:  mechanisms,
		Conditions:  nil,
		Filters:     nil,
		Created:     time.Date(2021, 05, 01, 10, 47, 01, 01, time.UTC),
		Modified:    time.Date(2021, 05, 01, 10, 47, 01, 01, time.UTC),
	}
	policies := []NotificationPolicy{
		policy,
	}
	b, err := json.Marshal(policies)
	require.NoError(t, err)
	require.NotEmpty(t, b)

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		_, err = fmt.Fprintf(w, `{
  									"success": true,
  									"errors": [],
  									"messages": [],
  									"result":%s
								}`,
			string(b))
		require.NoError(t, err)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/alerting/v3/policies", handler)

	res, err := client.ListNotificationPolicies(context.Background(), testAccountID)
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, policies, res.Result)
}

func TestUpdateNotificationPolicy(t *testing.T) {
	setup()
	defer teardown()

	mechanisms := make(map[string]NotificationMechanismIntegrations)
	mechanisms["email"] = []NotificationMechanismData{{Name: "email to send notification", ID: "test@gmail.com"}}
	policy := NotificationPolicy{
		ID:          testPolicyID,
		Description: "Notifies when my zones are under attack",
		Name:        "CF DOS attack alert - L4",
		Enabled:     true,
		AlertType:   "dos_attack_l4",
		Mechanisms:  mechanisms,
		Conditions:  nil,
		Filters:     nil,
		Created:     time.Date(2021, 05, 01, 10, 47, 01, 01, time.UTC),
		Modified:    time.Date(2021, 05, 01, 10, 47, 01, 01, time.UTC),
	}
	b, err := json.Marshal(policy)
	require.NoError(t, err)
	require.NotEmpty(t, b)

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		_, err = fmt.Fprintf(w, `{
  									"success": true,
  									"errors": [],
  									"messages": [],
  									"result":%s
								}`,
			string(b))
		require.NoError(t, err)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/alerting/v3/policies/"+testPolicyID, handler)

	res, err := client.UpdateNotificationPolicy(context.Background(), testAccountID, &policy)
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, testPolicyID, res.Result.ID)
}

func TestDeleteNotificationPolicy(t *testing.T) {
	setup()
	defer teardown()

	result := NotificationResource{ID: testPolicyID}
	b, err := json.Marshal(result)
	require.NoError(t, err)
	require.NotNil(t, b)

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		_, err = fmt.Fprintf(w, `{
  									"success": true,
  									"errors": [],
  									"messages": [],
  									"result":%s
								}`,
			string(b))
		require.NoError(t, err)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/alerting/v3/policies/"+testPolicyID, handler)

	res, err := client.DeleteNotificationPolicy(context.Background(), testAccountID, testPolicyID)
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, testPolicyID, res.Result.ID)
}

func TestCreateNotificationWebhooks(t *testing.T) {
	setup()
	defer teardown()

	webhook := NotificationUpsertWebhooks{
		Name:   "my test webhook",
		URL:    "https://example.com",
		Secret: "mischief-managed", // optional
	}

	result := NotificationResource{ID: testWebhookID}

	b, err := json.Marshal(result)
	require.NoError(t, err)
	require.NotEmpty(t, b)

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		_, err = fmt.Fprintf(w, `{
  									"success": true,
  									"errors": [],
  									"messages": [],
  									"result":%s
								}`,
			string(b))
		require.NoError(t, err)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/alerting/v3/destinations/webhooks", handler)

	res, err := client.CreateNotificationWebhooks(context.Background(), testAccountID, &webhook)
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, testWebhookID, res.Result.ID)
}

func TestListNotificationWebhooks(t *testing.T) {
	setup()
	defer teardown()

	webhook := NotificationWebhookIntegration{
		ID:          testWebhookID,
		Name:        "my test webhook",
		URL:         "https://example.com",
		Type:        "generic",
		CreatedAt:   notificationTimestamp,
		LastSuccess: &notificationTimestamp,
		LastFailure: &notificationTimestamp,
	}
	webhooks := []NotificationWebhookIntegration{webhook}
	b, err := json.Marshal(webhooks)
	require.NoError(t, err)
	require.NotEmpty(t, b)

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		_, err = fmt.Fprintf(w, `{
  									"success": true,
  									"errors": [],
  									"messages": [],
  									"result":%s
								}`,
			string(b))
		require.NoError(t, err)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/alerting/v3/destinations/webhooks", handler)

	res, err := client.ListNotificationWebhooks(context.Background(), testAccountID)
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, webhooks, res.Result)
}

func TestGetNotificationWebhooks(t *testing.T) {
	setup()
	defer teardown()

	webhook := NotificationWebhookIntegration{
		ID:          testWebhookID,
		Name:        "my test webhook",
		URL:         "https://example.com",
		Type:        "generic",
		CreatedAt:   notificationTimestamp,
		LastSuccess: &notificationTimestamp,
		LastFailure: &notificationTimestamp,
	}
	b, err := json.Marshal(webhook)
	require.NoError(t, err)
	require.NotEmpty(t, b)

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		_, err = fmt.Fprintf(w, `{
  									"success": true,
  									"errors": [],
  									"messages": [],
  									"result":%s
								}`,
			string(b))
		require.NoError(t, err)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/alerting/v3/destinations/webhooks/"+testWebhookID, handler)

	res, err := client.GetNotificationWebhooks(context.Background(), testAccountID, testWebhookID)
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, webhook, res.Result)
}

func TestUpdateNotificationWebhooks(t *testing.T) {
	setup()
	defer teardown()

	result := NotificationResource{ID: testWebhookID}
	b, err := json.Marshal(result)
	require.NoError(t, err)
	require.NotEmpty(t, b)

	webhook := NotificationUpsertWebhooks{
		Name:   "my test webhook with a new name",
		URL:    "https://example.com",
		Secret: "mischief-managed",
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		_, err = fmt.Fprintf(w, `{
  									"success": true,
  									"errors": [],
  									"messages": [],
  									"result":%s
								}`,
			string(b))
		require.NoError(t, err)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/alerting/v3/destinations/webhooks/"+testWebhookID, handler)

	res, err := client.UpdateNotificationWebhooks(context.Background(), testAccountID, testWebhookID, &webhook)
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, testWebhookID, res.Result.ID)
}

func TestDeleteNotificationWebhooks(t *testing.T) {
	setup()
	defer teardown()

	result := NotificationResource{ID: testWebhookID}
	b, err := json.Marshal(result)
	require.NoError(t, err)
	require.NotEmpty(t, b)

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		_, err = fmt.Fprintf(w, `{
  									"success": true,
  									"errors": [],
  									"messages": [],
  									"result":%s
								}`,
			string(b))
		require.NoError(t, err)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/alerting/v3/destinations/webhooks/"+testWebhookID, handler)

	res, err := client.DeleteNotificationWebhooks(context.Background(), testAccountID, testWebhookID)
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, testWebhookID, res.Result.ID)
}

func TestListNotificationHistory(t *testing.T) {
	setup()
	defer teardown()

	expected := []NotificationHistory{
		{
			ID:            "some-id",
			Name:          "some-name",
			Description:   "some-description",
			AlertBody:     "some-alert-body",
			AlertType:     "some-alert-type",
			Mechanism:     "some-mechanism",
			MechanismType: "some-mechanism-type",
			Sent:          notificationTimestamp,
		},
	}

	expectedResultInfo := ResultInfo{
		Page:    0,
		PerPage: 25,
		Count:   1,
	}

	pageOptions := PaginationOptions{
		PerPage: 25,
		Page:    1,
	}

	timeRange := TimeRange{
		Since:  time.Now().Add(-15 * time.Minute).Format(time.RFC3339),
		Before: time.Now().Format(time.RFC3339),
	}

	historyFilters := AlertHistoryFilter{TimeRange: timeRange, PaginationOptions: pageOptions}

	alertHistory, err := json.Marshal(expected)
	require.NoError(t, err)
	require.NotNil(t, alertHistory)

	resultInfo, err := json.Marshal(expectedResultInfo)
	require.NoError(t, err)
	require.NotNil(t, resultInfo)

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		_, err = fmt.Fprintf(w, `{
  									"success": true,
  									"errors": [],
  									"messages": [],
									"result_info": %s,
  									"result": %s
								}`,
			string(resultInfo),
			string(alertHistory))
		if err != nil {
			return
		}
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/alerting/v3/history", handler)

	actualResult, actualResultInfo, err := client.ListNotificationHistory(context.Background(), testAccountID, historyFilters)
	require.Nil(t, err)
	require.NotNil(t, actualResult)
	require.Equal(t, expected, actualResult)
	require.Equal(t, expectedResultInfo, actualResultInfo)
}
