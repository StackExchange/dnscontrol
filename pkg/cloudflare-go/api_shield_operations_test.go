package cloudflare

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAPIShieldOperationId = "9def2cb0-3ed0-4737-92ca-f09efa4718fd"

func TestGetAPIShieldOperation(t *testing.T) {
	setup()
	t.Cleanup(teardown)

	endpoint := fmt.Sprintf("/zones/%s/api_gateway/operations/%s", testZoneID, testAPIShieldOperationId)
	response := `{
		"success" : true,
		"errors": [],
		"messages": [],
		"result": {
			"operation_id": "9def2cb0-3ed0-4737-92ca-f09efa4718fd",
			"method": "POST",
			"host": "api.cloudflare.com",
			"endpoint": "/client/v4/zones",
			"last_updated": "2023-03-02T15:46:06.000000Z"
		}
	}`

	handler := func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		require.Empty(t, r.URL.Query())
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, response)
	}

	mux.HandleFunc(endpoint, handler)

	actual, err := client.GetAPIShieldOperation(
		context.Background(),
		ZoneIdentifier(testZoneID),
		GetAPIShieldOperationParams{
			OperationID: testAPIShieldOperationId,
		},
	)

	time := time.Date(2023, time.March, 2, 15, 46, 6, 0, time.UTC)
	expected := &APIShieldOperation{
		APIShieldBasicOperation: APIShieldBasicOperation{
			Method:   "POST",
			Host:     "api.cloudflare.com",
			Endpoint: "/client/v4/zones",
		},
		ID:          testAPIShieldOperationId,
		LastUpdated: &time,
		Features:    nil,
	}

	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func TestGetAPIShieldOperationWithParams(t *testing.T) {
	endpoint := fmt.Sprintf("/zones/%s/api_gateway/operations/%s", testZoneID, testAPIShieldOperationId)
	response := `{
		"success" : true,
		"errors": [],
		"messages": [],
		"result": {
			"operation_id": "9def2cb0-3ed0-4737-92ca-f09efa4718fd",
			"method": "POST",
			"host": "api.cloudflare.com",
			"endpoint": "/client/v4/zones",
			"last_updated": "2023-03-02T15:46:06.000000Z",
			"features":{
				"thresholds":{},
				"parameter_schemas":{}
			}
		}
	}`

	tests := []struct {
		name           string
		getParams      GetAPIShieldOperationParams
		expectedParams url.Values
	}{
		{
			name: "one feature",
			getParams: GetAPIShieldOperationParams{
				OperationID: testAPIShieldOperationId,
				Features:    []string{"thresholds"},
			},
			expectedParams: url.Values{
				"feature": []string{"thresholds"},
			},
		},
		{
			name: "more than one feature",
			getParams: GetAPIShieldOperationParams{
				OperationID: testAPIShieldOperationId,
				Features:    []string{"thresholds", "parameter_schemas"},
			},
			expectedParams: url.Values{
				"feature": []string{"thresholds", "parameter_schemas"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setup()
			t.Cleanup(teardown)
			handler := func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
				require.Equal(t, test.expectedParams, r.URL.Query())
				w.Header().Set("content-type", "application/json")
				fmt.Fprint(w, response)
			}

			mux.HandleFunc(endpoint, handler)

			actual, err := client.GetAPIShieldOperation(
				context.Background(),
				ZoneIdentifier(testZoneID),
				test.getParams,
			)

			time := time.Date(2023, time.March, 2, 15, 46, 6, 0, time.UTC)
			expected := &APIShieldOperation{
				APIShieldBasicOperation: APIShieldBasicOperation{
					Method:   "POST",
					Host:     "api.cloudflare.com",
					Endpoint: "/client/v4/zones",
				},
				ID:          "9def2cb0-3ed0-4737-92ca-f09efa4718fd",
				LastUpdated: &time,
				Features: map[string]any{
					"thresholds":        map[string]any{},
					"parameter_schemas": map[string]any{},
				},
			}

			if assert.NoError(t, err) {
				assert.Equal(t, expected, actual)
			}
		})
	}
}

func TestListAPIShieldOperations(t *testing.T) {
	endpoint := fmt.Sprintf("/zones/%s/api_gateway/operations", testZoneID)
	response := `{
		"success" : true,
		"errors": [],
		"messages": [],
		"result": [
			{
				"operation_id": "9def2cb0-3ed0-4737-92ca-f09efa4718fd",
				"method": "POST",
				"host": "api.cloudflare.com",
				"endpoint": "/client/v4/zones",
				"last_updated": "2023-03-02T15:46:06.000000Z"
			}

		],
		"result_info": {
			"page": 3,
			"per_page": 20,
			"count": 1,
			"total_count": 2000
		}
	}`

	setup()
	t.Cleanup(teardown)
	handler := func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		require.Empty(t, r.URL.Query())
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, response)
	}

	mux.HandleFunc(endpoint, handler)

	actual, actualResultInfo, err := client.ListAPIShieldOperations(
		context.Background(),
		ZoneIdentifier(testZoneID),
		ListAPIShieldOperationsParams{},
	)

	time := time.Date(2023, time.March, 2, 15, 46, 6, 0, time.UTC)
	expectedOps := []APIShieldOperation{
		{
			APIShieldBasicOperation: APIShieldBasicOperation{
				Method:   "POST",
				Host:     "api.cloudflare.com",
				Endpoint: "/client/v4/zones",
			},
			ID:          "9def2cb0-3ed0-4737-92ca-f09efa4718fd",
			LastUpdated: &time,
			Features:    nil,
		},
	}

	expectedResultInfo := ResultInfo{
		Page:    3,
		PerPage: 20,
		Count:   1,
		Total:   2000,
	}

	if assert.NoError(t, err) {
		assert.Equal(t, expectedOps, actual)
		assert.Equal(t, expectedResultInfo, actualResultInfo)
	}
}

func TestListAPIShieldOperationsWithParams(t *testing.T) {
	endpoint := fmt.Sprintf("/zones/%s/api_gateway/operations", testZoneID)
	response := `{
		"success" : true,
		"errors": [],
		"messages": [],
		"result": [
			{
				"operation_id": "9def2cb0-3ed0-4737-92ca-f09efa4718fd",
				"method": "POST",
				"host": "api.cloudflare.com",
				"endpoint": "/client/v4/zones",
				"last_updated": "2023-03-02T15:46:06.000000Z",
				"features": {
					"thresholds": {}
                }
			}
		],
		"result_info": {
			"page": 3,
			"per_page": 20,
			"count": 1,
			"total_count": 2000
		}
	}`

	tests := []struct {
		name           string
		params         ListAPIShieldOperationsParams
		expectedParams url.Values
	}{
		{
			name: "all params",
			params: ListAPIShieldOperationsParams{
				Features:  []string{"thresholds", "parameter_schemas"},
				Direction: "desc",
				OrderBy:   "host",
				APIShieldListOperationsFilters: APIShieldListOperationsFilters{
					Hosts:    []string{"api.cloudflare.com", "developers.cloudflare.com"},
					Methods:  []string{"GET", "PUT"},
					Endpoint: "/client",
				},
				PaginationOptions: PaginationOptions{
					Page:    1,
					PerPage: 25,
				},
			},
			expectedParams: url.Values{
				"feature":   []string{"thresholds", "parameter_schemas"},
				"direction": []string{"desc"},
				"order":     []string{"host"},
				"host":      []string{"api.cloudflare.com", "developers.cloudflare.com"},
				"method":    []string{"GET", "PUT"},
				"endpoint":  []string{"/client"},
				"page":      []string{"1"},
				"per_page":  []string{"25"},
			},
		},
		{
			name: "features only",
			params: ListAPIShieldOperationsParams{
				Features: []string{"thresholds", "parameter_schemas"},
			},
			expectedParams: url.Values{
				"feature": []string{"thresholds", "parameter_schemas"},
			},
		},
		{
			name: "direction only",
			params: ListAPIShieldOperationsParams{
				Direction: "desc",
			},
			expectedParams: url.Values{
				"direction": []string{"desc"},
			},
		},
		{
			name: "order only",
			params: ListAPIShieldOperationsParams{
				OrderBy: "host",
			},
			expectedParams: url.Values{
				"order": []string{"host"},
			},
		},
		{
			name: "hosts only",
			params: ListAPIShieldOperationsParams{
				APIShieldListOperationsFilters: APIShieldListOperationsFilters{
					Hosts: []string{"api.cloudflare.com", "developers.cloudflare.com"},
				},
			},
			expectedParams: url.Values{
				"host": []string{"api.cloudflare.com", "developers.cloudflare.com"},
			},
		},
		{
			name: "methods only",
			params: ListAPIShieldOperationsParams{
				APIShieldListOperationsFilters: APIShieldListOperationsFilters{
					Methods: []string{"GET", "PUT"},
				},
			},
			expectedParams: url.Values{
				"method": []string{"GET", "PUT"},
			},
		},
		{
			name: "endpoint only",
			params: ListAPIShieldOperationsParams{
				APIShieldListOperationsFilters: APIShieldListOperationsFilters{
					Endpoint: "/client",
				},
			},
			expectedParams: url.Values{
				"endpoint": []string{"/client"},
			},
		},
		{
			name: "pagination only",
			params: ListAPIShieldOperationsParams{
				PaginationOptions: PaginationOptions{
					Page:    1,
					PerPage: 25,
				},
			},
			expectedParams: url.Values{
				"page":     []string{"1"},
				"per_page": []string{"25"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setup()
			t.Cleanup(teardown)
			handler := func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
				require.Equal(t, test.expectedParams, r.URL.Query())
				w.Header().Set("content-type", "application/json")
				fmt.Fprint(w, response)
			}

			mux.HandleFunc(endpoint, handler)

			actual, _, err := client.ListAPIShieldOperations(
				context.Background(),
				ZoneIdentifier(testZoneID),
				test.params,
			)

			time := time.Date(2023, time.March, 2, 15, 46, 6, 0, time.UTC)
			expected := []APIShieldOperation{
				{
					APIShieldBasicOperation: APIShieldBasicOperation{
						Method:   "POST",
						Host:     "api.cloudflare.com",
						Endpoint: "/client/v4/zones",
					},
					ID:          "9def2cb0-3ed0-4737-92ca-f09efa4718fd",
					LastUpdated: &time,
					Features: map[string]any{
						"thresholds": map[string]any{},
					},
				},
			}

			if assert.NoError(t, err) {
				assert.Equal(t, expected, actual)
			}
		})
	}
}

func TestCreateAPIShieldOperations(t *testing.T) {
	setup()
	t.Cleanup(teardown)

	endpoint := fmt.Sprintf("/zones/%s/api_gateway/operations", testZoneID)
	response := `{
		"success" : true,
		"errors": [],
		"messages": [],
		"result": [
			{
				"operation_id": "9def2cb0-3ed0-4737-92ca-f09efa4718fd",
				"method": "POST",
				"host": "api.cloudflare.com",
				"endpoint": "/client/v4/zones",
				"last_updated": "2023-03-02T15:46:06.000000Z"
			}
		]
	}`

	handler := func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.Equal(t, []byte(`[{"method":"POST","host":"api.cloudflare.com","endpoint":"/client/v4/zones"}]`), body)

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, response)
	}

	mux.HandleFunc(endpoint, handler)

	actual, err := client.CreateAPIShieldOperations(
		context.Background(),
		ZoneIdentifier(testZoneID),
		CreateAPIShieldOperationsParams{
			Operations: []APIShieldBasicOperation{
				{
					Method:   "POST",
					Host:     "api.cloudflare.com",
					Endpoint: "/client/v4/zones",
				},
			},
		},
	)

	time := time.Date(2023, time.March, 2, 15, 46, 6, 0, time.UTC)
	expected := []APIShieldOperation{
		{
			APIShieldBasicOperation: APIShieldBasicOperation{
				Method:   "POST",
				Host:     "api.cloudflare.com",
				Endpoint: "/client/v4/zones",
			},
			ID:          "9def2cb0-3ed0-4737-92ca-f09efa4718fd",
			LastUpdated: &time,
			Features:    nil,
		},
	}

	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func TestDeleteAPIShieldOperation(t *testing.T) {
	setup()
	t.Cleanup(teardown)

	endpoint := fmt.Sprintf("/zones/%s/api_gateway/operations/%s", testZoneID, testAPIShieldOperationId)
	response := `{"result":{},"success":true,"errors":[],"messages":[]}`

	handler := func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		require.Empty(t, r.URL.Query())
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, response)
	}

	mux.HandleFunc(endpoint, handler)

	err := client.DeleteAPIShieldOperation(
		context.Background(),
		ZoneIdentifier(testZoneID),
		DeleteAPIShieldOperationParams{
			OperationID: testAPIShieldOperationId,
		},
	)

	assert.NoError(t, err)
}
