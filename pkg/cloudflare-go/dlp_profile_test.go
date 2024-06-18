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

func TestDLPProfiles(t *testing.T) {
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
				{
					"id": "d658f520-6ecb-4a34-a725-ba37243c2d28",
					"name": "U.S. Social Security Numbers",
					"entries": [
						{
							"id": "111b9d4b-a5c6-40f0-957d-9d53b25dd84a",
							"name": "SSN Numeric Detection",
							"profile_id": "d658f520-6ecb-4a34-a725-ba37243c2d28",
							"enabled": false,
							"type": "predefined"
						},
						{
							"id": "aec08712-ee49-4109-9d9f-3b229c5b3dcd",
							"name": "SSN Text",
							"profile_id": "d658f520-6ecb-4a34-a725-ba37243c2d28",
							"enabled": false,
							"type": "predefined"
						}
					],
					"type": "predefined",
					"allowed_match_count": 0,
					"context_awareness": {
						"enabled": true,
						"skip": {
							"files": true
						}
					},
					"ocr_enabled": false
				},
				{
					"id": "29678c26-a191-428d-9f63-6e20a4a636a4",
					"name": "Example Custom Profile",
					"entries": [
						{
							"id": "ef79b054-12d4-4067-bb30-b85f6267b91c",
							"name": "matches credit card regex",
							"profile_id": "29678c26-a191-428d-9f63-6e20a4a636a4",
							"created_at": "2022-10-18T08:00:56Z",
							"updated_at": "2022-10-18T08:00:57Z",
							"pattern": {
								"regex": "^4[0-9]$",
								"validation": "luhn"
							},
							"enabled": true,
							"type": "custom"
						}
					],
					"created_at": "2022-10-18T08:00:56Z",
					"updated_at": "2022-10-18T08:00:57Z",
					"type": "custom",
					"description": "just a custom profile example",
					"allowed_match_count": 1,
					"ocr_enabled": true 
				}
			]
		}
		`)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2022-10-18T08:00:56Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2022-10-18T08:00:57Z")

	want := []DLPProfile{
		{
			ID:                "d658f520-6ecb-4a34-a725-ba37243c2d28",
			Name:              "U.S. Social Security Numbers",
			Type:              "predefined",
			Description:       "",
			AllowedMatchCount: 0,
			ContextAwareness: &DLPContextAwareness{
				Enabled: BoolPtr(true),
				Skip: DLPContextAwarenessSkip{
					Files: BoolPtr(true),
				},
			},
			OCREnabled: BoolPtr(false),
			Entries: []DLPEntry{
				{
					ID:        "111b9d4b-a5c6-40f0-957d-9d53b25dd84a",
					Name:      "SSN Numeric Detection",
					ProfileID: "d658f520-6ecb-4a34-a725-ba37243c2d28",
					Type:      "predefined",
					Enabled:   BoolPtr(false),
				},
				{
					ID:   "aec08712-ee49-4109-9d9f-3b229c5b3dcd",
					Name: "SSN Text", ProfileID: "d658f520-6ecb-4a34-a725-ba37243c2d28",
					Type:    "predefined",
					Enabled: BoolPtr(false),
				},
			},
		},
		{
			ID:                "29678c26-a191-428d-9f63-6e20a4a636a4",
			Name:              "Example Custom Profile",
			Type:              "custom",
			Description:       "just a custom profile example",
			AllowedMatchCount: 1,
			// Omit ContextAwareness to test ContextAwareness optionality
			OCREnabled: BoolPtr(true),
			Entries: []DLPEntry{
				{
					ID:        "ef79b054-12d4-4067-bb30-b85f6267b91c",
					Name:      "matches credit card regex",
					ProfileID: "29678c26-a191-428d-9f63-6e20a4a636a4",
					Enabled:   BoolPtr(true),
					Type:      "custom",
					Pattern: &DLPPattern{
						Regex:      "^4[0-9]$",
						Validation: "luhn",
					},
					CreatedAt: &createdAt,
					UpdatedAt: &updatedAt,
				},
			},
			CreatedAt: &createdAt,
			UpdatedAt: &updatedAt,
		}}

	mux.HandleFunc("/accounts/"+testAccountID+"/dlp/profiles", handler)

	actual, err := client.ListDLPProfiles(context.Background(), AccountIdentifier(testAccountID), ListDLPProfilesParams{})
	require.NoError(t, err)
	require.Equal(t, want, actual)
}

func TestGetDLPProfile(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "29678c26-a191-428d-9f63-6e20a4a636a4",
				"name": "Example Custom Profile",
				"entries": [
					{
						"id": "ef79b054-12d4-4067-bb30-b85f6267b91c",
						"name": "matches credit card regex",
						"profile_id": "29678c26-a191-428d-9f63-6e20a4a636a4",
						"created_at": "2022-10-18T08:00:56Z",
						"updated_at": "2022-10-18T08:00:57Z",
						"pattern": {
							"regex": "^4[0-9]$",
							"validation": "luhn"
						},
						"enabled": true,
						"type": "custom"
					}
				],
				"created_at": "2022-10-18T08:00:56Z",
				"updated_at": "2022-10-18T08:00:57Z",
				"type": "custom",
				"description": "just a custom profile example",
				"allowed_match_count": 42,
				"context_awareness": {
					"enabled": false,
					"skip": {
						"files": false
					}
				}
			}
		}`)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2022-10-18T08:00:56Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2022-10-18T08:00:57Z")

	want := DLPProfile{
		ID:                "29678c26-a191-428d-9f63-6e20a4a636a4",
		Name:              "Example Custom Profile",
		Type:              "custom",
		Description:       "just a custom profile example",
		AllowedMatchCount: 42,
		ContextAwareness: &DLPContextAwareness{
			Enabled: BoolPtr(false),
			Skip: DLPContextAwarenessSkip{
				Files: BoolPtr(false),
			},
		},
		Entries: []DLPEntry{
			{
				ID:        "ef79b054-12d4-4067-bb30-b85f6267b91c",
				Name:      "matches credit card regex",
				ProfileID: "29678c26-a191-428d-9f63-6e20a4a636a4",
				Enabled:   BoolPtr(true),
				Pattern: &DLPPattern{
					Regex:      "^4[0-9]$",
					Validation: "luhn",
				},
				Type:      "custom",
				CreatedAt: &createdAt,
				UpdatedAt: &updatedAt,
			},
		},
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/dlp/profiles/29678c26-a191-428d-9f63-6e20a4a636a4", handler)

	actual, err := client.GetDLPProfile(context.Background(), AccountIdentifier(testAccountID), "29678c26-a191-428d-9f63-6e20a4a636a4")
	require.NoError(t, err)
	require.Equal(t, want, actual)
}

func TestCreateDLPCustomProfiles(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")

		var reqBody DLPProfilesCreateRequest
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		require.Nil(t, err)
		requestProfile := reqBody.Profiles[0]

		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": [{
				"id": "29678c26-a191-428d-9f63-6e20a4a636a4",
				"name": "`+requestProfile.Name+`",
				"entries": [
					{
						"id": "ef79b054-12d4-4067-bb30-b85f6267b91c",
						"name": "`+requestProfile.Entries[0].Name+`",
						"profile_id": "29678c26-a191-428d-9f63-6e20a4a636a4",
						"created_at": "2022-10-18T08:00:56Z",
						"updated_at": "2022-10-18T08:00:57Z",
						"pattern": {
							"regex": "`+requestProfile.Entries[0].Pattern.Regex+`",
							"validation": "`+requestProfile.Entries[0].Pattern.Validation+`"
						},
						"enabled": `+fmt.Sprintf("%t", Bool(requestProfile.Entries[0].Enabled))+`,
						"type": "custom"
					}
				],
				"created_at": "2022-10-18T08:00:56Z",
				"updated_at": "2022-10-18T08:00:57Z",
				"type": "custom",
				"description": "`+requestProfile.Description+`",
				"allowed_match_count": 0,
				"ocr_enabled": true
			}]
		}`)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2022-10-18T08:00:56Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2022-10-18T08:00:57Z")

	want := []DLPProfile{
		{
			ID:          "29678c26-a191-428d-9f63-6e20a4a636a4",
			Name:        "Example Custom Profile",
			Type:        "custom",
			Description: "just a custom profile example",
			Entries: []DLPEntry{
				{
					ID:        "ef79b054-12d4-4067-bb30-b85f6267b91c",
					Name:      "matches credit card regex",
					ProfileID: "29678c26-a191-428d-9f63-6e20a4a636a4",
					Enabled:   BoolPtr(true),
					Type:      "custom",
					Pattern: &DLPPattern{
						Regex:      "^4[0-9]$",
						Validation: "luhn",
					},
					CreatedAt: &createdAt,
					UpdatedAt: &updatedAt,
				},
			},
			CreatedAt:         &createdAt,
			UpdatedAt:         &updatedAt,
			AllowedMatchCount: 0,
			OCREnabled:        BoolPtr(true),
		},
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/dlp/profiles/custom", handler)

	profiles := []DLPProfile{
		{
			Name:        "Example Custom Profile",
			Description: "just a custom profile example",
			Type:        "custom",
			Entries: []DLPEntry{
				{
					Name:    "matches credit card regex",
					Enabled: BoolPtr(true),
					Pattern: &DLPPattern{
						Regex:      "^4[0-9]$",
						Validation: "luhn",
					},
				},
			},
			AllowedMatchCount: 0,
		},
	}
	actual, err := client.CreateDLPProfiles(context.Background(), AccountIdentifier(testAccountID), CreateDLPProfilesParams{Profiles: profiles, Type: "custom"})
	require.NoError(t, err)
	require.Equal(t, want, actual)
}

func TestCreateDLPCustomProfile(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")

		var reqBody DLPProfilesCreateRequest
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		require.Nil(t, err)
		requestProfile := reqBody.Profiles[0]

		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": [{
				"id": "29678c26-a191-428d-9f63-6e20a4a636a4",
				"name": "`+requestProfile.Name+`",
				"entries": [
					{
						"id": "ef79b054-12d4-4067-bb30-b85f6267b91c",
						"name": "`+requestProfile.Entries[0].Name+`",
						"profile_id": "29678c26-a191-428d-9f63-6e20a4a636a4",
						"created_at": "2022-10-18T08:00:56Z",
						"updated_at": "2022-10-18T08:00:57Z",
						"pattern": {
							"regex": "`+requestProfile.Entries[0].Pattern.Regex+`",
							"validation": "`+requestProfile.Entries[0].Pattern.Validation+`"
						},
						"enabled": `+fmt.Sprintf("%t", Bool(requestProfile.Entries[0].Enabled))+`,
						"type": "custom"
					}
				],
				"created_at": "2022-10-18T08:00:56Z",
				"updated_at": "2022-10-18T08:00:57Z",
				"type": "custom",
				"description": "`+requestProfile.Description+`",
				"allowed_match_count": 0
			}]
		}`)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2022-10-18T08:00:56Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2022-10-18T08:00:57Z")

	want := []DLPProfile{{

		ID:          "29678c26-a191-428d-9f63-6e20a4a636a4",
		Name:        "Example Custom Profile",
		Type:        "custom",
		Description: "just a custom profile example",
		Entries: []DLPEntry{
			{
				ID:        "ef79b054-12d4-4067-bb30-b85f6267b91c",
				Name:      "matches credit card regex",
				ProfileID: "29678c26-a191-428d-9f63-6e20a4a636a4",
				Type:      "custom",
				Enabled:   BoolPtr(true),
				Pattern: &DLPPattern{
					Regex:      "^4[0-9]$",
					Validation: "luhn",
				},
				CreatedAt: &createdAt,
				UpdatedAt: &updatedAt,
			},
		},
		CreatedAt:         &createdAt,
		UpdatedAt:         &updatedAt,
		AllowedMatchCount: 0,
	}}

	mux.HandleFunc("/accounts/"+testAccountID+"/dlp/profiles/custom", handler)

	profiles := []DLPProfile{{
		Name:        "Example Custom Profile",
		Description: "just a custom profile example",
		Entries: []DLPEntry{
			{
				Name:    "matches credit card regex",
				Enabled: BoolPtr(true),
				Pattern: &DLPPattern{
					Regex:      "^4[0-9]$",
					Validation: "luhn",
				},
			},
		},
		AllowedMatchCount: 0,
	}}

	actual, err := client.CreateDLPProfiles(context.Background(), AccountIdentifier(testAccountID), CreateDLPProfilesParams{
		Profiles: profiles,
		Type:     "custom",
	})
	require.NoError(t, err)
	require.Equal(t, want, actual)
}

func TestUpdateDLPCustomProfile(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")

		var requestProfile DLPProfile
		err := json.NewDecoder(r.Body).Decode(&requestProfile)
		require.Nil(t, err)

		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "29678c26-a191-428d-9f63-6e20a4a636a4",
				"name": "`+requestProfile.Name+`",
				"entries": [
					{
						"id": "ef79b054-12d4-4067-bb30-b85f6267b91c",
						"name": "`+requestProfile.Entries[0].Name+`",
						"profile_id": "29678c26-a191-428d-9f63-6e20a4a636a4",
						"created_at": "2022-10-18T08:00:56Z",
						"updated_at": "2022-10-18T08:00:57Z",
						"pattern": {
							"regex": "`+requestProfile.Entries[0].Pattern.Regex+`",
							"validation": "`+requestProfile.Entries[0].Pattern.Validation+`"
						},
						"enabled": `+fmt.Sprintf("%t", Bool(requestProfile.Entries[0].Enabled))+`,
						"type": "custom"
					}
				],
				"created_at": "2022-10-18T08:00:56Z",
				"updated_at": "2022-10-18T08:00:57Z",
				"type": "custom",
				"description": "`+requestProfile.Description+`",
				"allowed_match_count": 0
			}
		}`)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2022-10-18T08:00:56Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2022-10-18T08:00:57Z")

	want := DLPProfile{

		ID:          "29678c26-a191-428d-9f63-6e20a4a636a4",
		Name:        "Example Custom Profile",
		Type:        "custom",
		Description: "just a custom profile example",
		Entries: []DLPEntry{
			{
				ID:        "ef79b054-12d4-4067-bb30-b85f6267b91c",
				Name:      "matches credit card regex",
				ProfileID: "29678c26-a191-428d-9f63-6e20a4a636a4",
				Enabled:   BoolPtr(true),
				Type:      "custom",
				Pattern: &DLPPattern{
					Regex:      "^4[0-9]$",
					Validation: "luhn",
				},
				CreatedAt: &createdAt,
				UpdatedAt: &updatedAt,
			},
		},
		CreatedAt:         &createdAt,
		UpdatedAt:         &updatedAt,
		AllowedMatchCount: 0,
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/dlp/profiles/custom/29678c26-a191-428d-9f63-6e20a4a636a4", handler)

	customProfile := DLPProfile{
		Name:        "Example Custom Profile",
		Description: "just a custom profile example",
		Entries: []DLPEntry{
			{
				Name:    "matches credit card regex",
				Enabled: BoolPtr(true),
				Pattern: &DLPPattern{
					Regex:      "^4[0-9]$",
					Validation: "luhn",
				},
			},
		},
		AllowedMatchCount: 0,
	}
	actual, err := client.UpdateDLPProfile(context.Background(), AccountIdentifier(testAccountID), UpdateDLPProfileParams{
		ProfileID: "29678c26-a191-428d-9f63-6e20a4a636a4",
		Type:      "custom",
		Profile:   customProfile,
	})
	require.NoError(t, err)
	require.Equal(t, want, actual)
}

func TestUpdateDLPPredefinedProfile(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")

		var requestProfile DLPProfile
		err := json.NewDecoder(r.Body).Decode(&requestProfile)
		require.Nil(t, err)

		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "29678c26-a191-428d-9f63-6e20a4a636a4",
				"name": "Example predefined profile",
				"entries": [
					{
						"id": "ef79b054-12d4-4067-bb30-b85f6267b91c",
						"name": "Example predefined entry",
						"profile_id": "29678c26-a191-428d-9f63-6e20a4a636a4",
						"enabled": `+fmt.Sprintf("%t", Bool(requestProfile.Entries[0].Enabled))+`,
						"type": "predefined"
					}
				],
				"type": "predefined",
				"description": "example predefined profile",
				"allowed_match_count": 0,
				"context_awareness": {
					"enabled": true,
					"skip": {
						"files": true
					}
				}
			}
		}`)
	}

	want := DLPProfile{
		ID:                "29678c26-a191-428d-9f63-6e20a4a636a4",
		Name:              "Example predefined profile",
		Type:              "predefined",
		Description:       "example predefined profile",
		AllowedMatchCount: 0,
		ContextAwareness: &DLPContextAwareness{
			Enabled: BoolPtr(true),
			Skip: DLPContextAwarenessSkip{
				Files: BoolPtr(true),
			},
		},
		Entries: []DLPEntry{
			{
				ID:        "ef79b054-12d4-4067-bb30-b85f6267b91c",
				Name:      "Example predefined entry",
				ProfileID: "29678c26-a191-428d-9f63-6e20a4a636a4",
				Type:      "predefined",
				Enabled:   BoolPtr(true),
			},
		},
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/dlp/profiles/predefined/29678c26-a191-428d-9f63-6e20a4a636a4", handler)

	actual, err := client.UpdateDLPProfile(context.Background(), AccountIdentifier(testAccountID), UpdateDLPProfileParams{
		ProfileID: "29678c26-a191-428d-9f63-6e20a4a636a4",
		Type:      "predefined",
		Profile: DLPProfile{
			Entries: []DLPEntry{
				{
					ID:      "29678c26-a191-428d-9f63-6e20a4a636a4",
					Enabled: BoolPtr(true),
				},
			},
		}})
	require.NoError(t, err)
	require.Equal(t, want, actual)
}

func TestDeleteDLPCustomProfile(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": []
		}`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/dlp/profiles/custom/29678c26-a191-428d-9f63-6e20a4a636a4", handler)

	err := client.DeleteDLPProfile(context.Background(), AccountIdentifier(testAccountID), "29678c26-a191-428d-9f63-6e20a4a636a4")
	require.NoError(t, err)
}
