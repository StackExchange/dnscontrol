package cloudflare

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListDLPDatasets(t *testing.T) {
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
					"created_at": "2019-08-24T14:15:22Z",
					"description": "string",
					"id": "497f6eca-6276-4993-bfeb-53cbbbba6f08",
					"name": "string",
					"num_cells": 0,
					"secret": true,
					"status": "empty",
					"updated_at": "2019-08-24T14:15:22Z",
					"uploads": [
					  {
						"num_cells": 0,
						"status": "empty",
						"version": 0
					  }
					]
				  }
			]
		}`)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2019-08-24T14:15:22Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2019-08-24T14:15:22Z")

	secret := true
	want := []DLPDataset{
		{
			CreatedAt:   &createdAt,
			Description: "string",
			ID:          "497f6eca-6276-4993-bfeb-53cbbbba6f08",
			Name:        "string",
			NumCells:    0,
			Secret:      &secret,
			Status:      "empty",
			UpdatedAt:   &updatedAt,
			Uploads: []DLPDatasetUpload{
				{
					NumCells: 0,
					Status:   "empty",
					Version:  0,
				},
			},
		},
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/dlp/datasets", handler)

	actual, err := client.ListDLPDatasets(context.Background(), AccountIdentifier(testAccountID), ListDLPDatasetsParams{})
	require.NoError(t, err)
	require.Equal(t, want, actual)
}

func TestGetDLPDataset(t *testing.T) {
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
				"created_at": "2019-08-24T14:15:22Z",
				"description": "string",
				"id": "497f6eca-6276-4993-bfeb-53cbbbba6f08",
				"name": "string",
				"num_cells": 0,
				"secret": true,
				"status": "empty",
				"updated_at": "2019-08-24T14:15:22Z",
				"uploads": [
					{
						"num_cells": 0,
						"status": "empty",
						"version": 0
					}
				]
			}
		}`)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2019-08-24T14:15:22Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2019-08-24T14:15:22Z")

	secret := true
	want := DLPDataset{
		CreatedAt:   &createdAt,
		Description: "string",
		ID:          "497f6eca-6276-4993-bfeb-53cbbbba6f08",
		Name:        "string",
		NumCells:    0,
		Secret:      &secret,
		Status:      "empty",
		UpdatedAt:   &updatedAt,
		Uploads: []DLPDatasetUpload{
			{
				NumCells: 0,
				Status:   "empty",
				Version:  0,
			},
		},
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/dlp/datasets/497f6eca-6276-4993-bfeb-53cbbbba6f08", handler)

	actual, err := client.GetDLPDataset(context.Background(), AccountIdentifier(testAccountID), "497f6eca-6276-4993-bfeb-53cbbbba6f08")
	require.NoError(t, err)
	require.Equal(t, want, actual)
}

func TestCreateDLPDataset(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")

		var reqBody CreateDLPDatasetParams
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		require.Nil(t, err)

		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"max_cells": 0,
				"secret": "1234",
				"version": 0,
				"dataset": {
					"created_at": "2019-08-24T14:15:22Z",
					"description": "`+reqBody.Description+`",
					"id": "497f6eca-6276-4993-bfeb-53cbbbba6f08",
					"name": "`+reqBody.Name+`",
					"num_cells": 0,
					"secret": `+fmt.Sprintf("%t", *reqBody.Secret)+`,
					"status": "empty",
					"updated_at": "2019-08-24T14:15:22Z",
					"uploads": [
						{
							"num_cells": 0,
							"status": "empty",
							"version": 0
						}
					]
				}
			}
		}`)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2019-08-24T14:15:22Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2019-08-24T14:15:22Z")

	secret := true
	want := CreateDLPDatasetResult{
		MaxCells: 0,
		Secret:   "1234",
		Version:  0,
		Dataset: DLPDataset{
			CreatedAt:   &createdAt,
			Description: "string",
			ID:          "497f6eca-6276-4993-bfeb-53cbbbba6f08",
			Name:        "string",
			NumCells:    0,
			Secret:      &secret,
			Status:      "empty",
			UpdatedAt:   &updatedAt,
			Uploads: []DLPDatasetUpload{
				{
					NumCells: 0,
					Status:   "empty",
					Version:  0,
				},
			},
		},
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/dlp/datasets", handler)

	actual, err := client.CreateDLPDataset(context.Background(), AccountIdentifier(testAccountID), CreateDLPDatasetParams{Description: "string", Name: "string", Secret: &secret})
	require.NoError(t, err)
	require.Equal(t, want, actual)
}

func TestDeleteDLPDataset(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": null
		}`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/dlp/datasets/497f6eca-6276-4993-bfeb-53cbbbba6f08", handler)

	err := client.DeleteDLPDataset(context.Background(), AccountIdentifier(testAccountID), "497f6eca-6276-4993-bfeb-53cbbbba6f08")
	require.NoError(t, err)
}

func TestUpdateDLPDataset(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")

		var reqBody UpdateDLPDatasetParams
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		require.Nil(t, err)

		var description string
		if reqBody.Description == nil {
			description = "string"
		} else {
			description = *reqBody.Description
		}

		var name string
		if reqBody.Name == nil {
			name = "string"
		} else {
			name = *reqBody.Name
		}

		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"created_at": "2019-08-24T14:15:22Z",
				"description": "`+description+`",
				"id": "497f6eca-6276-4993-bfeb-53cbbbba6f08",
				"name": "`+name+`",
				"num_cells": 0,
				"secret": true,
				"status": "empty",
				"updated_at": "2019-08-24T14:15:22Z",
				"uploads": [
					{
						"num_cells": 0,
						"status": "empty",
						"version": 0
					}
				]
			}
		}`)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2019-08-24T14:15:22Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2019-08-24T14:15:22Z")

	secret := true
	want := DLPDataset{
		CreatedAt:   &createdAt,
		Description: "new_desc",
		ID:          "497f6eca-6276-4993-bfeb-53cbbbba6f08",
		Name:        "string",
		NumCells:    0,
		Secret:      &secret,
		Status:      "empty",
		UpdatedAt:   &updatedAt,
		Uploads: []DLPDatasetUpload{
			{
				NumCells: 0,
				Status:   "empty",
				Version:  0,
			},
		},
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/dlp/datasets/497f6eca-6276-4993-bfeb-53cbbbba6f08", handler)

	description := "new_desc"
	actual, err := client.UpdateDLPDataset(context.Background(), AccountIdentifier(testAccountID), UpdateDLPDatasetParams{Description: &description, Name: nil, DatasetID: "497f6eca-6276-4993-bfeb-53cbbbba6f08"})
	require.NoError(t, err)
	require.Equal(t, want, actual)
}

func TestCreateDLPDatasetUpload(t *testing.T) {
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
				"max_cells": 0,
				"secret": "1234",
				"version": 1
			}
		}`)
	}

	want := CreateDLPDatasetUploadResult{
		MaxCells: 0,
		Secret:   "1234",
		Version:  1,
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/dlp/datasets/497f6eca-6276-4993-bfeb-53cbbbba6f08/upload", handler)

	actual, err := client.CreateDLPDatasetUpload(context.Background(), AccountIdentifier(testAccountID), CreateDLPDatasetUploadParams{DatasetID: "497f6eca-6276-4993-bfeb-53cbbbba6f08"})
	require.NoError(t, err)
	require.Equal(t, want, actual)
}

func TestUploadDLPDatasetVersion(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.Equal(t, []byte{1, 2, 3, 4}, body)

		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"created_at": "2019-08-24T14:15:22Z",
				"description": "string",
				"id": "497f6eca-6276-4993-bfeb-53cbbbba6f08",
				"name": "string",
				"num_cells": 5,
				"secret": true,
				"status": "complete",
				"updated_at": "2019-08-24T14:15:22Z",
				"uploads": [
					{
						"num_cells": 0,
						"status": "empty",
						"version": 0
					},
					{
						"num_cells": 5,
						"status": "complete",
						"version": 1
					}
				]
			}
		}`)
	}

	createdAt, _ := time.Parse(time.RFC3339, "2019-08-24T14:15:22Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2019-08-24T14:15:22Z")

	secret := true
	want := DLPDataset{
		CreatedAt:   &createdAt,
		Description: "string",
		ID:          "497f6eca-6276-4993-bfeb-53cbbbba6f08",
		Name:        "string",
		NumCells:    5,
		Secret:      &secret,
		Status:      "complete",
		UpdatedAt:   &updatedAt,
		Uploads: []DLPDatasetUpload{
			{
				NumCells: 0,
				Status:   "empty",
				Version:  0,
			},
			{
				NumCells: 5,
				Status:   "complete",
				Version:  1,
			},
		},
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/dlp/datasets/497f6eca-6276-4993-bfeb-53cbbbba6f08/upload/1", handler)

	actual, err := client.UploadDLPDatasetVersion(context.Background(), AccountIdentifier(testAccountID), UploadDLPDatasetVersionParams{DatasetID: "497f6eca-6276-4993-bfeb-53cbbbba6f08", Version: 1, Body: []byte{1, 2, 3, 4}})
	require.NoError(t, err)
	require.Equal(t, want, actual)
}
