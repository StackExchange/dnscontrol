package cloudflare

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAPIShieldSchemaId = "0f05e4fc-1c36-4bd2-a5e7-b29d32d3dc8e"

func TestGetAPIShieldSchema(t *testing.T) {
	endpoint := fmt.Sprintf("/zones/%s/api_gateway/user_schemas/%s", testZoneID, testAPIShieldSchemaId)
	response := `{
		"success" : true,
		"errors": [],
		"messages": [],
		"result": {
			"schema_id": "0f05e4fc-1c36-4bd2-a5e7-b29d32d3dc8e",
			"name": "petstore.json",
			"kind": "openapi_v3",
			"created_at": "2023-03-02T15:46:06.000000Z",
			"validation_enabled": true,
			"source": "{}"
		}
	}`

	tests := []struct {
		name           string
		getParams      GetAPIShieldSchemaParams
		expectedParams url.Values
	}{
		{
			name: "without omit source",
			getParams: GetAPIShieldSchemaParams{
				SchemaID: testAPIShieldSchemaId,
			},
			expectedParams: url.Values{},
		},
		{
			name: "with omit source=true",
			getParams: GetAPIShieldSchemaParams{
				SchemaID:   testAPIShieldSchemaId,
				OmitSource: BoolPtr(true),
			},
			expectedParams: url.Values{
				"omit_source": []string{"true"},
			},
		},
		{
			name: "with omit source=false",
			getParams: GetAPIShieldSchemaParams{
				SchemaID:   testAPIShieldSchemaId,
				OmitSource: BoolPtr(false),
			},
			expectedParams: url.Values{
				"omit_source": []string{"false"},
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

			actual, err := client.GetAPIShieldSchema(
				context.Background(),
				ZoneIdentifier(testZoneID),
				test.getParams,
			)

			createdAt := time.Date(2023, time.March, 2, 15, 46, 6, 0, time.UTC)
			expected := &APIShieldSchema{
				ID:                testAPIShieldSchemaId,
				Name:              "petstore.json",
				Kind:              "openapi_v3",
				Source:            "{}",
				CreatedAt:         &createdAt,
				ValidationEnabled: true,
			}

			if assert.NoError(t, err) {
				assert.Equal(t, expected, actual)
			}
		})
	}
}

func TestListAPIShieldSchemas(t *testing.T) {
	endpoint := fmt.Sprintf("/zones/%s/api_gateway/user_schemas", testZoneID)
	response := `{
		"success" : true,
		"errors": [],
		"messages": [],
		"result": [
			{
				"schema_id": "0f05e4fc-1c36-4bd2-a5e7-b29d32d3dc8e",
				"name": "petstore.json",
				"kind": "openapi_v3",
				"created_at": "2023-03-02T15:46:06.000000Z",
				"validation_enabled": true,
				"source": "{}"
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
		getParams      ListAPIShieldSchemasParams
		expectedParams url.Values
	}{
		{
			name:           "with empty params",
			getParams:      ListAPIShieldSchemasParams{},
			expectedParams: url.Values{},
		},
		{
			name: "all params",
			getParams: ListAPIShieldSchemasParams{
				OmitSource:        BoolPtr(true),
				ValidationEnabled: BoolPtr(false),
				PaginationOptions: PaginationOptions{
					Page:    1,
					PerPage: 25,
				},
			},
			expectedParams: url.Values{
				"omit_source":        []string{"true"},
				"validation_enabled": []string{"false"},
				"page":               []string{"1"},
				"per_page":           []string{"25"},
			},
		},
		{
			name: "with omit source",
			getParams: ListAPIShieldSchemasParams{
				OmitSource: BoolPtr(true),
			},
			expectedParams: url.Values{
				"omit_source": []string{"true"},
			},
		},
		{
			name: "with validation enabled",
			getParams: ListAPIShieldSchemasParams{
				ValidationEnabled: BoolPtr(true),
			},
			expectedParams: url.Values{
				"validation_enabled": []string{"true"},
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

			actual, resultInfo, err := client.ListAPIShieldSchemas(
				context.Background(),
				ZoneIdentifier(testZoneID),
				test.getParams,
			)

			createdAt := time.Date(2023, time.March, 2, 15, 46, 6, 0, time.UTC)
			expected := []APIShieldSchema{
				{
					ID:                testAPIShieldSchemaId,
					Name:              "petstore.json",
					Kind:              "openapi_v3",
					Source:            "{}",
					CreatedAt:         &createdAt,
					ValidationEnabled: true,
				},
			}

			if assert.NoError(t, err) {
				assert.Equal(t, expected, actual)

				expectedResultInfo := ResultInfo{
					Page:    3,
					PerPage: 20,
					Count:   1,
					Total:   2000,
				}
				assert.Equal(t, expectedResultInfo, resultInfo)
			}
		})
	}
}

func TestDeleteAPIShieldSchema(t *testing.T) {
	setup()
	t.Cleanup(teardown)

	endpoint := fmt.Sprintf("/zones/%s/api_gateway/user_schemas/%s", testZoneID, testAPIShieldSchemaId)
	response := `{"result":{},"success":true,"errors":[],"messages":[]}`

	handler := func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		require.Empty(t, r.URL.Query())
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, response)
	}

	mux.HandleFunc(endpoint, handler)

	err := client.DeleteAPIShieldSchema(
		context.Background(),
		ZoneIdentifier(testZoneID),
		DeleteAPIShieldSchemaParams{
			SchemaID: testAPIShieldSchemaId,
		},
	)

	assert.NoError(t, err)
}

func TestUpdateAPIShieldSchema(t *testing.T) {
	setup()
	t.Cleanup(teardown)

	endpoint := fmt.Sprintf("/zones/%s/api_gateway/user_schemas/%s", testZoneID, testAPIShieldSchemaId)
	response := `{
		"success" : true,
		"errors": [],
		"messages": [],
		"result": {
			"schema_id": "0f05e4fc-1c36-4bd2-a5e7-b29d32d3dc8e",
			"name": "petstore.json",
			"kind": "openapi_v3",
			"validation_enabled": true
		}
	}`

	handler := func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)
		require.Empty(t, r.URL.Query())

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.Equal(t, `{"validation_enabled":true}`, string(body))

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, response)
	}

	mux.HandleFunc(endpoint, handler)

	actual, err := client.UpdateAPIShieldSchema(
		context.Background(),
		ZoneIdentifier(testZoneID),
		UpdateAPIShieldSchemaParams{
			SchemaID:          testAPIShieldSchemaId,
			ValidationEnabled: BoolPtr(true),
		},
	)

	// patch result is a cut down representation of the schema
	// so metadata like created date is not populated
	expected := &APIShieldSchema{
		ID:                testAPIShieldSchemaId,
		Name:              "petstore.json",
		Kind:              "openapi_v3",
		ValidationEnabled: true,
	}

	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}

	assert.NoError(t, err)
}

func TestCreateAPIShieldSchema(t *testing.T) {
	setup()
	t.Cleanup(teardown)

	endpoint := fmt.Sprintf("/zones/%s/api_gateway/user_schemas", testZoneID)
	response := `{
		"success" : true,
		"errors": [],
		"messages": [],
		"result": {
			"schema": {
				"schema_id": "0f05e4fc-1c36-4bd2-a5e7-b29d32d3dc8e",
				"name": "petstore.json",
				"kind": "openapi_v3",
				"created_at": "2023-03-02T15:46:06.000000Z",
				"validation_enabled": true,
				"source": "{}"
			},
			"upload_details": {
				"warnings": [
					{
						"code": 28,
						"message": "unsupported media type: application/octet-stream",
						"locations": [
							".paths[\"/user/{username}\"].put"
						]
					}
				]
			}
		}
	}`

	handler := func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)

		// Check that form data has been sent correctly.
		mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		require.NoError(t, err)
		require.True(t, strings.HasPrefix(mediaType, "multipart/form-data"))

		expectedParts := []struct {
			partKeys map[string]string
			data     []byte
		}{
			{
				partKeys: map[string]string{"name": "name"},
				data:     []byte("petstore.json"),
			},
			{
				partKeys: map[string]string{"name": "kind"},
				data:     []byte("openapi_v3"),
			},
			{
				partKeys: map[string]string{"name": "validation_enabled"},
				data:     []byte("true"),
			},
			{
				partKeys: map[string]string{"filename": "petstore.json", "name": "file"},
				data:     []byte("{}"),
			},
		}

		multiPartReader := multipart.NewReader(r.Body, params["boundary"])
		partNum := 0
		for {
			part, err := multiPartReader.NextPart()
			if errors.Is(err, io.EOF) {
				break
			}
			require.NoError(t, err)

			// more parts found - this is unexpected
			require.Less(t, partNum, len(expectedParts))

			value, err := io.ReadAll(part)
			require.NoError(t, err)

			_, dParams, err := mime.ParseMediaType(part.Header.Get("Content-Disposition"))
			require.NoError(t, err)

			require.Equal(t, expectedParts[partNum].partKeys, dParams)
			require.Equal(t, expectedParts[partNum].data, value)
			partNum++
		}

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, response)
	}

	mux.HandleFunc(endpoint, handler)

	actual, err := client.CreateAPIShieldSchema(
		context.Background(),
		ZoneIdentifier(testZoneID),
		CreateAPIShieldSchemaParams{
			Name:              "petstore.json",
			Kind:              "openapi_v3",
			Source:            strings.NewReader("{}"),
			ValidationEnabled: BoolPtr(true),
		},
	)

	createdAt := time.Date(2023, time.March, 2, 15, 46, 6, 0, time.UTC)
	expected := &APIShieldCreateSchemaResult{
		Schema: APIShieldSchema{
			ID:                testAPIShieldSchemaId,
			Name:              "petstore.json",
			Kind:              "openapi_v3",
			Source:            "{}",
			CreatedAt:         &createdAt,
			ValidationEnabled: true,
		},
		Events: APIShieldCreateSchemaEvents{
			Warnings: []APIShieldCreateSchemaEventWithLocations{
				{
					APIShieldCreateSchemaEvent: APIShieldCreateSchemaEvent{
						Code:    28,
						Message: "unsupported media type: application/octet-stream",
					},
					Locations: []string{
						".paths[\"/user/{username}\"].put",
					},
				},
			},
		},
	}

	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func TestCreateAPIShieldSchemaError(t *testing.T) {
	setup()
	t.Cleanup(teardown)

	endpoint := fmt.Sprintf("/zones/%s/api_gateway/user_schemas", testZoneID)
	response := `{
		"success" : false,
		"errors": [
			{
			  "code": 21400,
			  "message": "only default parameter styles are supported: path, form (.paths[\"/pet/{petId}\"].get.parameters[0])"
			}
		],
		"messages": [],
		"result": {
			"upload_details": {
				"errors": [
					{
						"code": 22,
						"message": "only default parameter styles are supported: path, form",
						"locations": [
							".paths[\"/pet/{petId}\"].get.parameters[0]"
						]
					}
				]
			}
		}
	}`

	handler := func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(400)
		fmt.Fprint(w, response)
	}

	mux.HandleFunc(endpoint, handler)

	_, err := client.CreateAPIShieldSchema(
		context.Background(),
		ZoneIdentifier(testZoneID),
		CreateAPIShieldSchemaParams{
			Name:              "petstore.json",
			Kind:              "openapi_v3",
			Source:            strings.NewReader("{}"),
			ValidationEnabled: BoolPtr(true),
		},
	)

	require.ErrorContains(t, err, "only default parameter styles are supported: path, form (.paths[\"/pet/{petId}\"].get.parameters[0]) (21400)")
}

func TestMustProvideSchemaID(t *testing.T) {
	setup()
	t.Cleanup(teardown)

	_, err := client.GetAPIShieldSchema(context.Background(), ZoneIdentifier(testZoneID), GetAPIShieldSchemaParams{})
	require.ErrorContains(t, err, "schema ID must be provided")

	_, err = client.UpdateAPIShieldSchema(context.Background(), ZoneIdentifier(testZoneID), UpdateAPIShieldSchemaParams{})
	require.ErrorContains(t, err, "schema ID must be provided")

	err = client.DeleteAPIShieldSchema(context.Background(), ZoneIdentifier(testZoneID), DeleteAPIShieldSchemaParams{})
	require.ErrorContains(t, err, "schema ID must be provided")
}

func TestGetAPIShieldSchemaValidationSettings(t *testing.T) {
	endpoint := fmt.Sprintf("/zones/%s/api_gateway/settings/schema_validation", testZoneID)
	response := `{
		"success" : true,
		"errors": [],
		"messages": [],
		"result": {
			"validation_default_mitigation_action": "log",
			"validation_override_mitigation_action": "none"
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

	actual, err := client.GetAPIShieldSchemaValidationSettings(
		context.Background(),
		ZoneIdentifier(testZoneID),
	)

	none := "none"
	expected := &APIShieldSchemaValidationSettings{
		DefaultMitigationAction:  "log",
		OverrideMitigationAction: &none,
	}

	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func TestUpdateAPIShieldSchemaValidationSettings(t *testing.T) {
	endpoint := fmt.Sprintf("/zones/%s/api_gateway/settings/schema_validation", testZoneID)
	response := `{
		"success" : true,
		"errors": [],
		"messages": [],
		"result": {
			"validation_default_mitigation_action": "log",
			"validation_override_mitigation_action": null
		}
	}`

	setup()
	t.Cleanup(teardown)
	handler := func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)
		require.Empty(t, r.URL.Query())
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, response)
	}

	mux.HandleFunc(endpoint, handler)

	actual, err := client.UpdateAPIShieldSchemaValidationSettings(
		context.Background(),
		ZoneIdentifier(testZoneID),
		UpdateAPIShieldSchemaValidationSettingsParams{}, // specifying nil is ok for fields
	)

	expected := &APIShieldSchemaValidationSettings{
		DefaultMitigationAction:  "log",
		OverrideMitigationAction: nil,
	}

	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func TestGetAPIShieldOperationSchemaValidationSettings(t *testing.T) {
	endpoint := fmt.Sprintf("/zones/%s/api_gateway/operations/%s/schema_validation", testZoneID, testAPIShieldOperationId)
	response := `{
		"success" : true,
		"errors": [],
		"messages": [],
		"result": {
			"mitigation_action": "log"
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

	actual, err := client.GetAPIShieldOperationSchemaValidationSettings(
		context.Background(),
		ZoneIdentifier(testZoneID),
		GetAPIShieldOperationSchemaValidationSettingsParams{OperationID: testAPIShieldOperationId},
	)

	log := "log"
	expected := &APIShieldOperationSchemaValidationSettings{
		MitigationAction: &log,
	}

	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func TestUpdateAPIShieldOperationSchemaValidationSettings(t *testing.T) {
	endpoint := fmt.Sprintf("/zones/%s/api_gateway/operations/schema_validation", testZoneID)
	response := fmt.Sprintf(`{
		"success" : true,
		"errors": [],
		"messages": [],
		"result": {
			"%s": null
		}
	}`, testAPIShieldOperationId)

	setup()
	t.Cleanup(teardown)
	handler := func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)
		require.Empty(t, r.URL.Query())

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		expected := fmt.Sprintf(`{"%s":{"mitigation_action":null}}`, testAPIShieldOperationId)
		require.Equal(t, expected, string(body))

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, response)
	}

	mux.HandleFunc(endpoint, handler)

	actual, err := client.UpdateAPIShieldOperationSchemaValidationSettings(
		context.Background(),
		ZoneIdentifier(testZoneID),
		UpdateAPIShieldOperationSchemaValidationSettings{
			testAPIShieldOperationId: APIShieldOperationSchemaValidationSettings{
				MitigationAction: nil,
			},
		},
	)

	expected := &UpdateAPIShieldOperationSchemaValidationSettings{
		testAPIShieldOperationId: APIShieldOperationSchemaValidationSettings{
			MitigationAction: nil,
		},
	}

	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}
