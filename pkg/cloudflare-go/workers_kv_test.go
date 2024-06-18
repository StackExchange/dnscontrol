package cloudflare

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkersKV_CreateWorkersKVNamespace(t *testing.T) {
	setup()
	defer teardown()

	response := `{
		"result": {
			"id" : "3aeaxxxxee014exxxx4cf66xxxxc0448",
			"title": "test_namespace"
		},
		"success": true,
		"errors": [],
		"messages": []
	}`

	mux.HandleFunc("/accounts/"+testAccountID+"/storage/kv/namespaces", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/javascript")
		fmt.Fprint(w, response)
	})

	res, err := client.CreateWorkersKVNamespace(context.Background(), AccountIdentifier(testAccountID), CreateWorkersKVNamespaceParams{Title: "Namespace"})
	want := WorkersKVNamespaceResponse{
		successResponse,
		WorkersKVNamespace{
			ID:    "3aeaxxxxee014exxxx4cf66xxxxc0448",
			Title: "test_namespace",
		},
	}

	if assert.NoError(t, err) {
		assert.Equal(t, want.Response, res.Response)
	}
}

func TestWorkersKV_DeleteWorkersKVNamespace(t *testing.T) {
	setup()
	defer teardown()

	namespace := "3aeaxxxxee014exxxx4cf66xxxxc0448"
	response := `{
		"success": true,
		"errors": [],
		"messages": []
	}`

	mux.HandleFunc(fmt.Sprintf("/accounts/"+testAccountID+"/storage/kv/namespaces/%s", namespace), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/javascript")
		fmt.Fprint(w, response)
	})

	res, err := client.DeleteWorkersKVNamespace(context.Background(), AccountIdentifier(testAccountID), namespace)
	want := successResponse

	if assert.NoError(t, err) {
		assert.Equal(t, want, res)
	}
}

func TestWorkersKV_ListWorkersKVNamespaces(t *testing.T) {
	setup()
	defer teardown()

	response := `{
		"result": [
			{"id": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
			"title": "test_namespace_1"
			},
			{"id": "yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy",
			"title": "test_namespace_2"
			}
		],
		"success": true,
		"errors": [],
		"messages": [],
		"result_info": {
			"page": 1,
			"per_page": 100,
			"count": 2,
			"total_count": 2,
			"total_pages": 1
		}
	}`

	mux.HandleFunc("/accounts/"+testAccountID+"/storage/kv/namespaces", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/javascript")
		fmt.Fprint(w, response)
	})

	res, _, err := client.ListWorkersKVNamespaces(context.Background(), AccountIdentifier(testAccountID), ListWorkersKVNamespacesParams{})
	want := []WorkersKVNamespace{
		{
			ID:    "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
			Title: "test_namespace_1",
		},
		{
			ID:    "yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy",
			Title: "test_namespace_2",
		},
	}

	if assert.NoError(t, err) {
		sort.Slice(res, func(i, j int) bool {
			return res[i].ID < res[j].ID
		})
		sort.Slice(want, func(i, j int) bool {
			return want[i].ID < want[j].ID
		})
		assert.Equal(t, res, want)
	}
}

func TestWorkersKV_ListWorkersKVNamespaceMultiplePages(t *testing.T) {
	setup()
	defer teardown()

	response1 := `{
		"result": [
			{"id": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
			"title": "test_namespace_1"
			}
		],
		"success": true,
		"errors": [],
		"messages": [],
		"result_info": {
			"page": 1,
			"per_page": 100,
			"count": 1,
			"total_count": 2,
			"total_pages": 2
		}
	}`

	response2 := `{
		"result": [
			{"id": "yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy",
			"title": "test_namespace_2"
			}
		],
		"success": true,
		"errors": [],
		"messages": [],
		"result_info": {
			"page": 2,
			"per_page": 100,
			"count": 1,
			"total_count": 2,
			"total_pages": 2
		}
	}`

	mux.HandleFunc("/accounts/"+testAccountID+"/storage/kv/namespaces", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/javascript")

		if r.URL.Query().Get("page") == "1" {
			fmt.Fprint(w, response1)
			return
		} else if r.URL.Query().Get("page") == "2" {
			fmt.Fprint(w, response2)
			return
		} else {
			panic(errors.New("Got a request for an unexpected page"))
		}
	})

	res, _, err := client.ListWorkersKVNamespaces(context.Background(), AccountIdentifier(testAccountID), ListWorkersKVNamespacesParams{})
	want := []WorkersKVNamespace{
		{
			ID:    "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
			Title: "test_namespace_1",
		},
		{
			ID:    "yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy",
			Title: "test_namespace_2",
		},
	}

	if assert.NoError(t, err) {
		sort.Slice(res, func(i, j int) bool {
			return res[i].ID < res[j].ID
		})
		sort.Slice(want, func(i, j int) bool {
			return want[i].ID < want[j].ID
		})
		assert.Equal(t, res, want)
	}
}

func TestWorkersKV_UpdateWorkersKVNamespace(t *testing.T) {
	setup()
	defer teardown()

	namespace := "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	response := `{
		"result": null,
		"success": true,
		"errors": [],
		"messages": []
	}`

	mux.HandleFunc(fmt.Sprintf("/accounts/"+testAccountID+"/storage/kv/namespaces/%s", namespace), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/javascript")
		fmt.Fprint(w, response)
	})

	res, err := client.UpdateWorkersKVNamespace(context.Background(), AccountIdentifier(testAccountID), UpdateWorkersKVNamespaceParams{Title: "Namespace", NamespaceID: namespace})
	want := successResponse

	if assert.NoError(t, err) {
		assert.Equal(t, want, res)
	}
}

func TestWorkersKV_WriteWorkersKVEntry(t *testing.T) {
	setup()
	defer teardown()

	key := "test_key"
	value := []byte("test_value")
	namespace := "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	response := `{
		"result": null,
		"success": true,
		"errors": [],
		"messages": []
	}`

	mux.HandleFunc(fmt.Sprintf("/accounts/"+testAccountID+"/storage/kv/namespaces/%s/values/%s", namespace, key), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/octet-stream")
		fmt.Fprint(w, response)
	})

	want := successResponse
	res, err := client.WriteWorkersKVEntry(context.Background(), AccountIdentifier(testAccountID), WriteWorkersKVEntryParams{NamespaceID: namespace, Key: key, Value: value})

	if assert.NoError(t, err) {
		assert.Equal(t, want, res)
	}
}

func TestWorkersKV_WriteWorkersKVEntries(t *testing.T) {
	setup()
	defer teardown()

	kvs := []*WorkersKVPair{
		{Key: "key1", Value: "value1"},
		{Key: "key2", Value: "value2"},
		{Key: "key3", Value: "value3", Metadata: "meta3", Base64: true},
	}

	namespace := "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	response := `{
		"result": null,
		"success": true,
		"errors": [],
		"messages": []
	}`

	mux.HandleFunc(fmt.Sprintf("/accounts/"+testAccountID+"/storage/kv/namespaces/%s/bulk", namespace), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, response)
	})

	want := successResponse
	res, err := client.WriteWorkersKVEntries(context.Background(), AccountIdentifier(testAccountID), WriteWorkersKVEntriesParams{NamespaceID: namespace, KVs: kvs})
	require.NoError(t, err)
	assert.Equal(t, want, res)
}

func TestWorkersKV_ReadWorkersKV(t *testing.T) {
	setup()
	defer teardown()

	key := "test_key"
	namespace := "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

	mux.HandleFunc(fmt.Sprintf("/accounts/"+testAccountID+"/storage/kv/namespaces/%s/values/%s", namespace, key), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "text/plain")
		fmt.Fprint(w, "test_value")
	})

	res, err := client.GetWorkersKV(context.Background(), AccountIdentifier(testAccountID), GetWorkersKVParams{NamespaceID: namespace, Key: key})
	want := []byte("test_value")

	if assert.NoError(t, err) {
		assert.Equal(t, want, res)
	}
}

func TestWorkersKV_DeleteWorkersKVEntry(t *testing.T) {
	setup()
	defer teardown()

	key := "test_key"
	namespace := "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	response := `{
		"result": null,
		"success": true,
		"errors": [],
		"messages": []
	}`

	mux.HandleFunc(fmt.Sprintf("/accounts/"+testAccountID+"/storage/kv/namespaces/%s/values/%s", namespace, key), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/javascript")
		fmt.Fprint(w, response)
	})

	res, err := client.DeleteWorkersKVEntry(context.Background(), AccountIdentifier(testAccountID), DeleteWorkersKVEntryParams{NamespaceID: namespace, Key: key})
	want := successResponse

	if assert.NoError(t, err) {
		assert.Equal(t, want, res)
	}
}

func TestWorkersKV_DeleteWorkersKVBulk(t *testing.T) {
	setup()
	defer teardown()

	keys := []string{"key1", "key2", "key3"}

	namespace := "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	response := `{
		"result": null,
		"success": true,
		"errors": [],
		"messages": []
	}`

	mux.HandleFunc(fmt.Sprintf("/accounts/"+testAccountID+"/storage/kv/namespaces/%s/bulk", namespace), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, response)
	})

	want := successResponse
	res, err := client.DeleteWorkersKVEntries(context.Background(), AccountIdentifier(testAccountID), DeleteWorkersKVEntriesParams{NamespaceID: namespace, Keys: keys})
	require.NoError(t, err)
	assert.Equal(t, want, res)
}

func TestWorkersKV_ListKeys(t *testing.T) {
	setup()
	defer teardown()

	namespace := "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	response := `{
		"result": [
			{"name": "test_key_1"},
			{"name": "test_key_2"}
		],
		"success": true,
		"errors": [],
		"messages": [],
		"result_info": {
			"page": 1,
			"per_page": 20,
			"count": 2,
			"total_count": 2,
			"total_pages": 1
		}
	}`

	mux.HandleFunc(fmt.Sprintf("/accounts/"+testAccountID+"/storage/kv/namespaces/%s/keys", namespace), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/javascript")
		fmt.Fprint(w, response)
	})

	res, err := client.ListWorkersKVKeys(context.Background(), AccountIdentifier(testAccountID), ListWorkersKVsParams{NamespaceID: namespace})

	want := ListStorageKeysResponse{
		successResponse,
		[]StorageKey{
			{Name: "test_key_1"},
			{Name: "test_key_2"},
		},
		ResultInfo{
			Page:       1,
			PerPage:    20,
			Count:      2,
			TotalPages: 1,
			Total:      2,
		},
	}

	if assert.NoError(t, err) {
		assert.Equal(t, want.Response, res.Response)
		assert.Equal(t, want.ResultInfo, res.ResultInfo)

		sort.Slice(res.Result, func(i, j int) bool {
			return res.Result[i].Name < res.Result[j].Name
		})

		sort.Slice(want.Result, func(i, j int) bool {
			return want.Result[i].Name < want.Result[j].Name
		})
		assert.Equal(t, want.Result, res.Result)
	}
}

func TestWorkersKV_ListKeysWithParameters(t *testing.T) {
	setup()
	defer teardown()

	cursor := "AArAbNSOuYcr4HmzGH02-cfDN8Ck9ejOwkn_Ai5rsn7S9NEqVJBenU9-gYRlrsziyjKLx48hNDLvtYzBAmkPsLGdye8ECr5PqFYcIOfUITdhkyTc1x6bV8nmyjz5DO-XaZH4kYY1KfqT8NRBIe5sic6yYt3FUDttGjafy0ivi-Up-TkVdRB0OxCf3O3OB-svG6DXheV5XTdDNrNx1o_CVqy2l2j0F4iKV1qFe_KhdkjC7Y6QjhUZ1MOb3J_uznNYVCoxZ-bVAAsJmXA"
	namespace := "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	response := `{
		"result": [
			{
				"name": "test_key_1",
				"metadata": "test_key_1_meta"
			},
			{
				"name": "test_key_2",
				"metadata": {
					"test2_meta_key": "test2_meta_value"
				}
			}
		],
		"success": true,
		"errors": [],
		"messages": [],
		"result_info": {
			"page": 1,
			"per_page": 20,
			"count": 2,
			"total_count": 2,
			"total_pages": 1,
			"cursor": "` + cursor + `"
		}
	}`

	mux.HandleFunc(fmt.Sprintf("/accounts/"+testAccountID+"/storage/kv/namespaces/%s/keys", namespace), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/javascript")
		fmt.Fprint(w, response)
	})

	limit, prefix := 25, "test-prefix"
	res, err := client.ListWorkersKVKeys(context.Background(), AccountIdentifier(testAccountID), ListWorkersKVsParams{
		NamespaceID: namespace,
		Limit:       limit,
		Prefix:      prefix,
	})

	want := ListStorageKeysResponse{
		successResponse,
		[]StorageKey{
			{
				Name:     "test_key_1",
				Metadata: "test_key_1_meta",
			},
			{
				Name: "test_key_2",
				Metadata: map[string]interface{}{
					"test2_meta_key": "test2_meta_value",
				},
			},
		},
		ResultInfo{
			Page:       1,
			PerPage:    20,
			Count:      2,
			TotalPages: 1,
			Total:      2,
			Cursor:     cursor,
		},
	}

	if assert.NoError(t, err) {
		assert.Equal(t, want.Response, res.Response)
		assert.Equal(t, want.ResultInfo, res.ResultInfo)

		sort.Slice(res.Result, func(i, j int) bool {
			return res.Result[i].Name < res.Result[j].Name
		})

		sort.Slice(want.Result, func(i, j int) bool {
			return want.Result[i].Name < want.Result[j].Name
		})
		assert.Equal(t, want.Result, res.Result)
	}
}
