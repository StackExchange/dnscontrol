package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestListLists(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"result": [
				{
					"id": "2c0fc9fa937b11eaa1b71c4d701ab86e",
					"name": "list1",
					"description": "This is a note.",
					"kind": "ip",
					"num_items": 10,
					"num_referencing_filters": 2,
					"created_on": "2020-01-01T08:00:00Z",
					"modified_on": "2020-01-10T14:00:00Z"
				}
			],
			"success": true,
			"errors": [],
			"messages": []
		}`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/rules/lists", handler)

	createdOn, _ := time.Parse(time.RFC3339, "2020-01-01T08:00:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2020-01-10T14:00:00Z")

	want := []List{
		{
			ID:                    "2c0fc9fa937b11eaa1b71c4d701ab86e",
			Name:                  "list1",
			Description:           "This is a note.",
			Kind:                  "ip",
			NumItems:              10,
			NumReferencingFilters: 2,
			CreatedOn:             &createdOn,
			ModifiedOn:            &modifiedOn,
		},
	}

	actual, err := client.ListLists(context.Background(), AccountIdentifier(testAccountID), ListListsParams{})
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateList(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"result": {
				"id": "2c0fc9fa937b11eaa1b71c4d701ab86e",
				"name": "list1",
				"description": "This is a note.",
				"kind": "ip",
				"num_items": 10,
				"num_referencing_filters": 2,
				"created_on": "2020-01-01T08:00:00Z",
				"modified_on": "2020-01-10T14:00:00Z"
			},
			"success": true,
			"errors": [],
			"messages": []
		}`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/rules/lists", handler)

	createdOn, _ := time.Parse(time.RFC3339, "2020-01-01T08:00:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2020-01-10T14:00:00Z")

	want := List{
		ID:                    "2c0fc9fa937b11eaa1b71c4d701ab86e",
		Name:                  "list1",
		Description:           "This is a note.",
		Kind:                  "ip",
		NumItems:              10,
		NumReferencingFilters: 2,
		CreatedOn:             &createdOn,
		ModifiedOn:            &modifiedOn,
	}

	actual, err := client.CreateList(context.Background(), AccountIdentifier(testAccountID), ListCreateParams{
		Name: "list1", Description: "This is a note.", Kind: "ip",
	})
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestGetList(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"result": {
				"id": "2c0fc9fa937b11eaa1b71c4d701ab86e",
				"name": "list1",
				"description": "This is a note.",
				"kind": "ip",
				"num_items": 10,
				"num_referencing_filters": 2,
				"created_on": "2020-01-01T08:00:00Z",
				"modified_on": "2020-01-10T14:00:00Z"
			},
			"success": true,
			"errors": [],
			"messages": []
		}`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/rules/lists/2c0fc9fa937b11eaa1b71c4d701ab86e", handler)

	createdOn, _ := time.Parse(time.RFC3339, "2020-01-01T08:00:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2020-01-10T14:00:00Z")

	want := List{
		ID:                    "2c0fc9fa937b11eaa1b71c4d701ab86e",
		Name:                  "list1",
		Description:           "This is a note.",
		Kind:                  "ip",
		NumItems:              10,
		NumReferencingFilters: 2,
		CreatedOn:             &createdOn,
		ModifiedOn:            &modifiedOn,
	}

	actual, err := client.GetList(context.Background(), AccountIdentifier(testAccountID), "2c0fc9fa937b11eaa1b71c4d701ab86e")
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestUpdateList(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"result": {
				"id": "2c0fc9fa937b11eaa1b71c4d701ab86e",
				"name": "list1",
				"description": "This note was updated.",
				"kind": "ip",
				"num_items": 10,
				"num_referencing_filters": 2,
				"created_on": "2020-01-01T08:00:00Z",
				"modified_on": "2020-01-10T14:00:00Z"
			},
			"success": true,
			"errors": [],
			"messages": []
		}`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/rules/lists/2c0fc9fa937b11eaa1b71c4d701ab86e", handler)

	createdOn, _ := time.Parse(time.RFC3339, "2020-01-01T08:00:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2020-01-10T14:00:00Z")

	want := List{
		ID:                    "2c0fc9fa937b11eaa1b71c4d701ab86e",
		Name:                  "list1",
		Description:           "This note was updated.",
		Kind:                  "ip",
		NumItems:              10,
		NumReferencingFilters: 2,
		CreatedOn:             &createdOn,
		ModifiedOn:            &modifiedOn,
	}

	actual, err := client.UpdateList(context.Background(), AccountIdentifier(testAccountID),
		ListUpdateParams{
			ID: "2c0fc9fa937b11eaa1b71c4d701ab86e", Description: "This note was updated.",
		},
	)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestDeleteList(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"result": {
				"id": "34b12448945f11eaa1b71c4d701ab86e"
			},
			"success": true,
			"errors": [],
			"messages": []
		}`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/rules/lists/2c0fc9fa937b11eaa1b71c4d701ab86e", handler)

	want := ListDeleteResponse{}
	want.Success = true
	want.Errors = []ResponseInfo{}
	want.Messages = []ResponseInfo{}
	want.Result.ID = "34b12448945f11eaa1b71c4d701ab86e"

	actual, err := client.DeleteList(context.Background(), AccountIdentifier(testAccountID), "2c0fc9fa937b11eaa1b71c4d701ab86e")
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestListsItemsIP(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")

		if len(r.URL.Query().Get("cursor")) > 0 && r.URL.Query().Get("cursor") == "yyy" {
			fmt.Fprint(w, `{
			"result": [
    			{
      				"id": "2c0fc9fa937b11eaa1b71c4d701ab86e",
      				"ip": "192.0.2.2",
      				"comment": "Another Private IP address",
      				"created_on": "2020-01-01T08:00:00Z",
      				"modified_on": "2020-01-10T14:00:00Z"
    			}
  			],
  			"result_info": {
    			"cursors": {
					"before": "xxx"
				}
			},
			"success": true,
			"errors": [],
			"messages": []
		}`)
		} else {
			fmt.Fprint(w, `{
			"result": [
    			{
      				"id": "2c0fc9fa937b11eaa1b71c4d701ab86e",
      				"ip": "192.0.2.1",
      				"comment": "Private IP address",
      				"created_on": "2020-01-01T08:00:00Z",
      				"modified_on": "2020-01-10T14:00:00Z"
    			}
  			],
  			"result_info": {
    			"cursors": {
					"before": "xxx",
					"after": "yyy"
				}
			},
			"success": true,
			"errors": [],
			"messages": []
		}`)
		}
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/rules/lists/2c0fc9fa937b11eaa1b71c4d701ab86e/items", handler)

	createdOn, _ := time.Parse(time.RFC3339, "2020-01-01T08:00:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2020-01-10T14:00:00Z")

	want := []ListItem{
		{
			ID:         "2c0fc9fa937b11eaa1b71c4d701ab86e",
			IP:         StringPtr("192.0.2.1"),
			Comment:    "Private IP address",
			CreatedOn:  &createdOn,
			ModifiedOn: &modifiedOn,
		},
		{
			ID:         "2c0fc9fa937b11eaa1b71c4d701ab86e",
			IP:         StringPtr("192.0.2.2"),
			Comment:    "Another Private IP address",
			CreatedOn:  &createdOn,
			ModifiedOn: &modifiedOn,
		},
	}

	actual, err := client.ListListItems(context.Background(), AccountIdentifier(testAccountID), ListListItemsParams{
		ID: "2c0fc9fa937b11eaa1b71c4d701ab86e",
	})
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestListsItemsRedirect(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")

		if len(r.URL.Query().Get("cursor")) > 0 && r.URL.Query().Get("cursor") == "yyy" {
			fmt.Fprint(w, `{
			"result": [
				{
      				"id": "1c0fc9fa937b11eaa1b71c4d701ab86e",
      				"redirect": {
						"source_url": "www.3fonteinen.be",
						"target_url": "https://shop.3fonteinen.be"
					},
      				"comment": "3F redirect",
      				"created_on": "2020-01-01T08:00:00Z",
      				"modified_on": "2020-01-10T14:00:00Z"
    			}
  			],
  			"result_info": {
    			"cursors": {
					"before": "xxx"
				}
			},
			"success": true,
			"errors": [],
			"messages": []
		}`)
		} else {
			fmt.Fprint(w, `{
			"result": [
				{
      				"id": "0c0fc9fa937b11eaa1b71c4d701ab86e",
      				"redirect": {
						"source_url": "http://cloudflare.com",
						"include_subdomains": true,
						"target_url": "https://cloudflare.com",
						"status_code": 302,
						"preserve_query_string": true,
						"subpath_matching": true,
						"preserve_path_suffix": false
					},
      				"comment": "Cloudflare http redirect",
      				"created_on": "2020-01-01T08:00:00Z",
      				"modified_on": "2020-01-10T14:00:00Z"
    			}
  			],
  			"result_info": {
    			"cursors": {
					"before": "xxx",
					"after": "yyy"
				}
			},
			"success": true,
			"errors": [],
			"messages": []
		}`)
		}
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/rules/lists/0c0fc9fa937b11eaa1b71c4d701ab86e/items", handler)

	createdOn, _ := time.Parse(time.RFC3339, "2020-01-01T08:00:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2020-01-10T14:00:00Z")

	want := []ListItem{
		{
			ID: "0c0fc9fa937b11eaa1b71c4d701ab86e",
			Redirect: &Redirect{
				SourceUrl:           "http://cloudflare.com",
				IncludeSubdomains:   BoolPtr(true),
				TargetUrl:           "https://cloudflare.com",
				StatusCode:          IntPtr(302),
				PreserveQueryString: BoolPtr(true),
				SubpathMatching:     BoolPtr(true),
				PreservePathSuffix:  BoolPtr(false),
			},
			Comment:    "Cloudflare http redirect",
			CreatedOn:  &createdOn,
			ModifiedOn: &modifiedOn,
		},
		{
			ID: "1c0fc9fa937b11eaa1b71c4d701ab86e",
			Redirect: &Redirect{
				SourceUrl: "www.3fonteinen.be",
				TargetUrl: "https://shop.3fonteinen.be",
			},
			Comment:    "3F redirect",
			CreatedOn:  &createdOn,
			ModifiedOn: &modifiedOn,
		},
	}

	actual, err := client.ListListItems(
		context.Background(),
		AccountIdentifier(testAccountID),
		ListListItemsParams{ID: "0c0fc9fa937b11eaa1b71c4d701ab86e"},
	)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestListsItemsHostname(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"result": [
				{
      				"id": "0c0fc9fa937b11eaa1b71c4d701ab86e",
      				"hostname": {
						"url_hostname": "cloudflare.com"
					},
      				"comment": "CF hostname",
      				"created_on": "2023-01-01T08:00:00Z",
      				"modified_on": "2023-01-10T14:00:00Z"
    			}
  			],
  			"result_info": {
    			"cursors": {
					"before": "xxx"
				}
			},
			"success": true,
			"errors": [],
			"messages": []
		}`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/rules/lists/0c0fc9fa937b11eaa1b71c4d701ab86e/items", handler)

	createdOn, _ := time.Parse(time.RFC3339, "2023-01-01T08:00:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2023-01-10T14:00:00Z")

	want := []ListItem{
		{
			ID: "0c0fc9fa937b11eaa1b71c4d701ab86e",
			Hostname: &Hostname{
				UrlHostname: "cloudflare.com",
			},
			Comment:    "CF hostname",
			CreatedOn:  &createdOn,
			ModifiedOn: &modifiedOn,
		},
	}

	actual, err := client.ListListItems(
		context.Background(),
		AccountIdentifier(testAccountID),
		ListListItemsParams{ID: "0c0fc9fa937b11eaa1b71c4d701ab86e"},
	)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestListsItemsASN(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")

		fmt.Fprint(w, `{
			"result": [
				{
      				"id": "0c0fc9fa937b11eaa1b71c4d701ab86e",
      				"asn": 3456,
      				"comment": "ASN",
      				"created_on": "2023-01-01T08:00:00Z",
      				"modified_on": "2023-01-10T14:00:00Z"
    			}
  			],
  			"result_info": {
    			"cursors": {
					"before": "xxx"
				}
			},
			"success": true,
			"errors": [],
			"messages": []
		}`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/rules/lists/0c0fc9fa937b11eaa1b71c4d701ab86e/items", handler)

	createdOn, _ := time.Parse(time.RFC3339, "2023-01-01T08:00:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2023-01-10T14:00:00Z")

	want := []ListItem{
		{
			ID:         "0c0fc9fa937b11eaa1b71c4d701ab86e",
			ASN:        Uint32Ptr(3456),
			Comment:    "ASN",
			CreatedOn:  &createdOn,
			ModifiedOn: &modifiedOn,
		},
	}

	actual, err := client.ListListItems(
		context.Background(),
		AccountIdentifier(testAccountID),
		ListListItemsParams{ID: "0c0fc9fa937b11eaa1b71c4d701ab86e"},
	)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateListItemsIP(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"result": {
				"operation_id": "4da8780eeb215e6cb7f48dd981c4ea02"
			},
			"success": true,
			"errors": [],
			"messages": []
		}`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/rules/lists/2c0fc9fa937b11eaa1b71c4d701ab86e/items", handler)

	want := ListItemCreateResponse{}
	want.Success = true
	want.Errors = []ResponseInfo{}
	want.Messages = []ResponseInfo{}
	want.Result.OperationID = "4da8780eeb215e6cb7f48dd981c4ea02"

	actual, err := client.CreateListItemsAsync(context.Background(), AccountIdentifier(testAccountID), ListCreateItemsParams{
		ID: "2c0fc9fa937b11eaa1b71c4d701ab86e",
		Items: []ListItemCreateRequest{{
			IP:      StringPtr("192.0.2.1"),
			Comment: "Private IP",
		}, {
			IP:      StringPtr("192.0.2.2"),
			Comment: "Another Private IP",
		}}})
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateListItemsRedirect(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"result": {
				"operation_id": "4da8780eeb215e6cb7f48dd981c4ea02"
			},
			"success": true,
			"errors": [],
			"messages": []
		}`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/rules/lists/0c0fc9fa937b11eaa1b71c4d701ab86e/items", handler)

	want := ListItemCreateResponse{}
	want.Success = true
	want.Errors = []ResponseInfo{}
	want.Messages = []ResponseInfo{}
	want.Result.OperationID = "4da8780eeb215e6cb7f48dd981c4ea02"

	actual, err := client.CreateListItemsAsync(context.Background(), AccountIdentifier(testAccountID), ListCreateItemsParams{
		ID: "0c0fc9fa937b11eaa1b71c4d701ab86e",
		Items: []ListItemCreateRequest{{
			Redirect: &Redirect{
				SourceUrl: "www.3fonteinen.be",
				TargetUrl: "https://shop.3fonteinen.be",
			},
			Comment: "redirect 3F",
		}, {
			Redirect: &Redirect{
				SourceUrl: "www.cf.com",
				TargetUrl: "https://cloudflare.com",
			},
			Comment: "Redirect cf",
		}}})
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateListItemsHostname(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"result": {
				"operation_id": "4da8780eeb215e6cb7f48dd981c4ea02"
			},
			"success": true,
			"errors": [],
			"messages": []
		}`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/rules/lists/0c0fc9fa937b11eaa1b71c4d701ab86e/items", handler)

	want := ListItemCreateResponse{}
	want.Success = true
	want.Errors = []ResponseInfo{}
	want.Messages = []ResponseInfo{}
	want.Result.OperationID = "4da8780eeb215e6cb7f48dd981c4ea02"

	actual, err := client.CreateListItemsAsync(context.Background(), AccountIdentifier(testAccountID), ListCreateItemsParams{
		ID: "0c0fc9fa937b11eaa1b71c4d701ab86e",
		Items: []ListItemCreateRequest{{
			Hostname: &Hostname{
				UrlHostname: "3fonteinen.be", // ie. only match 3fonteinen.be
			},
			Comment: "hostname 3F",
		}, {
			Hostname: &Hostname{
				UrlHostname: "*.cf.com", // ie. match all subdomains of cf.com but not cf.com
			},
			Comment: "Hostname cf",
		}, {
			Hostname: &Hostname{
				UrlHostname: "*.abc.com", // ie. equivalent to match all subdomains of abc.com excluding abc.com
			},
			Comment: "Hostname abc",
		}}})
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateListItemsASN(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"result": {
				"operation_id": "4da8780eeb215e6cb7f48dd981c4ea02"
			},
			"success": true,
			"errors": [],
			"messages": []
		}`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/rules/lists/0c0fc9fa937b11eaa1b71c4d701ab86e/items", handler)

	want := ListItemCreateResponse{}
	want.Success = true
	want.Errors = []ResponseInfo{}
	want.Messages = []ResponseInfo{}
	want.Result.OperationID = "4da8780eeb215e6cb7f48dd981c4ea02"

	actual, err := client.CreateListItemsAsync(context.Background(), AccountIdentifier(testAccountID), ListCreateItemsParams{
		ID: "0c0fc9fa937b11eaa1b71c4d701ab86e",
		Items: []ListItemCreateRequest{{
			ASN:     Uint32Ptr(458),
			Comment: "ASN 458",
		}, {
			ASN:     Uint32Ptr(789),
			Comment: "ASN 789",
		}}})
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestReplaceListItemsIP(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"result": {
				"operation_id": "4da8780eeb215e6cb7f48dd981c4ea02"
			},
			"success": true,
			"errors": [],
			"messages": []
		}`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/rules/lists/2c0fc9fa937b11eaa1b71c4d701ab86e/items", handler)

	want := ListItemCreateResponse{}
	want.Success = true
	want.Errors = []ResponseInfo{}
	want.Messages = []ResponseInfo{}
	want.Result.OperationID = "4da8780eeb215e6cb7f48dd981c4ea02"

	actual, err := client.ReplaceListItemsAsync(context.Background(), AccountIdentifier(testAccountID), ListReplaceItemsParams{
		ID: "2c0fc9fa937b11eaa1b71c4d701ab86e",
		Items: []ListItemCreateRequest{{
			IP:      StringPtr("192.0.2.1"),
			Comment: "Private IP",
		}, {
			IP:      StringPtr("192.0.2.2"),
			Comment: "Another Private IP",
		}}})
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestReplaceListItemsRedirect(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"result": {
				"operation_id": "4da8780eeb215e6cb7f48dd981c4ea02"
			},
			"success": true,
			"errors": [],
			"messages": []
		}`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/rules/lists/2c0fc9fa937b11eaa1b71c4d701ab86e/items", handler)

	want := ListItemCreateResponse{}
	want.Success = true
	want.Errors = []ResponseInfo{}
	want.Messages = []ResponseInfo{}
	want.Result.OperationID = "4da8780eeb215e6cb7f48dd981c4ea02"

	actual, err := client.ReplaceListItemsAsync(context.Background(), AccountIdentifier(testAccountID), ListReplaceItemsParams{
		ID: "2c0fc9fa937b11eaa1b71c4d701ab86e",
		Items: []ListItemCreateRequest{{
			Redirect: &Redirect{
				SourceUrl: "www.3fonteinen.be",
				TargetUrl: "https://shop.3fonteinen.be",
			},
			Comment: "redirect 3F",
		}, {
			Redirect: &Redirect{
				SourceUrl: "www.cf.com",
				TargetUrl: "https://cloudflare.com",
			},
			Comment: "Redirect cf",
		}}})
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestReplaceListItemsHostname(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"result": {
				"operation_id": "4da8780eeb215e6cb7f48dd981c4ea02"
			},
			"success": true,
			"errors": [],
			"messages": []
		}`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/rules/lists/2c0fc9fa937b11eaa1b71c4d701ab86e/items", handler)

	want := ListItemCreateResponse{}
	want.Success = true
	want.Errors = []ResponseInfo{}
	want.Messages = []ResponseInfo{}
	want.Result.OperationID = "4da8780eeb215e6cb7f48dd981c4ea02"

	actual, err := client.ReplaceListItemsAsync(context.Background(), AccountIdentifier(testAccountID), ListReplaceItemsParams{
		ID: "2c0fc9fa937b11eaa1b71c4d701ab86e",
		Items: []ListItemCreateRequest{{
			Hostname: &Hostname{
				UrlHostname: "3fonteinen.be",
			},
			Comment: "hostname 3F",
		}, {
			Hostname: &Hostname{
				UrlHostname: "cf.com",
			},
			Comment: "Hostname cf",
		}}})
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestReplaceListItemsASN(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"result": {
				"operation_id": "4da8780eeb215e6cb7f48dd981c4ea02"
			},
			"success": true,
			"errors": [],
			"messages": []
		}`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/rules/lists/2c0fc9fa937b11eaa1b71c4d701ab86e/items", handler)

	want := ListItemCreateResponse{}
	want.Success = true
	want.Errors = []ResponseInfo{}
	want.Messages = []ResponseInfo{}
	want.Result.OperationID = "4da8780eeb215e6cb7f48dd981c4ea02"

	actual, err := client.ReplaceListItemsAsync(context.Background(), AccountIdentifier(testAccountID), ListReplaceItemsParams{
		ID: "2c0fc9fa937b11eaa1b71c4d701ab86e",
		Items: []ListItemCreateRequest{{
			ASN:     Uint32Ptr(4567),
			Comment: "ASN 4567",
		}, {
			ASN:     Uint32Ptr(8901),
			Comment: "ASN 8901",
		}}})
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestDeleteListItems(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"result": {
				"operation_id": "4da8780eeb215e6cb7f48dd981c4ea02"
			},
			"success": true,
			"errors": [],
			"messages": []
		}`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/rules/lists/2c0fc9fa937b11eaa1b71c4d701ab86e/items", handler)

	want := ListItemDeleteResponse{}
	want.Success = true
	want.Errors = []ResponseInfo{}
	want.Messages = []ResponseInfo{}
	want.Result.OperationID = "4da8780eeb215e6cb7f48dd981c4ea02"

	actual, err := client.DeleteListItemsAsync(context.Background(), AccountIdentifier(testAccountID), ListDeleteItemsParams{
		ID: "2c0fc9fa937b11eaa1b71c4d701ab86e",
		Items: ListItemDeleteRequest{[]ListItemDeleteItemRequest{{
			ID: "34b12448945f11eaa1b71c4d701ab86e",
		}}}})
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestGetListItemIP(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"result": {
				"id": "2c0fc9fa937b11eaa1b71c4d701ab86e",
				"ip": "192.0.2.1",
				"comment": "Private IP address",
				"created_on": "2020-01-01T08:00:00Z",
				"modified_on": "2020-01-10T14:00:00Z"
			},
			"success": true,
			"errors": [],
			"messages": []
		}`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/rules/lists/2c0fc9fa937b11eaa1b71c4d701ab86e/items/"+
		"34b12448945f11eaa1b71c4d701ab86e", handler)

	createdOn, _ := time.Parse(time.RFC3339, "2020-01-01T08:00:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2020-01-10T14:00:00Z")

	want := ListItem{
		ID:         "2c0fc9fa937b11eaa1b71c4d701ab86e",
		IP:         StringPtr("192.0.2.1"),
		Comment:    "Private IP address",
		CreatedOn:  &createdOn,
		ModifiedOn: &modifiedOn,
	}

	actual, err := client.GetListItem(context.Background(), AccountIdentifier(testAccountID), "2c0fc9fa937b11eaa1b71c4d701ab86e", "34b12448945f11eaa1b71c4d701ab86e")
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestGetListItemHostname(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"result": {
				"id": "2c0fc9fa937b11eaa1b71c4d701ab86e",
				"hostname": {
					"url_hostname": "cloudflare.com"
				},
				"comment": "CF Hostname",
				"created_on": "2023-01-01T08:00:00Z",
				"modified_on": "2023-01-10T14:00:00Z"
			},
			"success": true,
			"errors": [],
			"messages": []
		}`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/rules/lists/2c0fc9fa937b11eaa1b71c4d701ab86e/items/"+
		"34b12448945f11eaa1b71c4d701ab86e", handler)

	createdOn, _ := time.Parse(time.RFC3339, "2023-01-01T08:00:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2023-01-10T14:00:00Z")

	want := ListItem{
		ID: "2c0fc9fa937b11eaa1b71c4d701ab86e",
		Hostname: &Hostname{
			UrlHostname: "cloudflare.com",
		},
		Comment:    "CF Hostname",
		CreatedOn:  &createdOn,
		ModifiedOn: &modifiedOn,
	}

	actual, err := client.GetListItem(context.Background(), AccountIdentifier(testAccountID), "2c0fc9fa937b11eaa1b71c4d701ab86e", "34b12448945f11eaa1b71c4d701ab86e")
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestGetListItemASN(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"result": {
				"id": "2c0fc9fa937b11eaa1b71c4d701ab86e",
				"asn": 5555,
				"comment": "asn 5555",
				"created_on": "2023-01-01T08:00:00Z",
				"modified_on": "2023-01-10T14:00:00Z"
			},
			"success": true,
			"errors": [],
			"messages": []
		}`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/rules/lists/2c0fc9fa937b11eaa1b71c4d701ab86e/items/"+
		"34b12448945f11eaa1b71c4d701ab86e", handler)

	createdOn, _ := time.Parse(time.RFC3339, "2023-01-01T08:00:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2023-01-10T14:00:00Z")

	want := ListItem{
		ID:         "2c0fc9fa937b11eaa1b71c4d701ab86e",
		ASN:        Uint32Ptr(5555),
		Comment:    "asn 5555",
		CreatedOn:  &createdOn,
		ModifiedOn: &modifiedOn,
	}

	actual, err := client.GetListItem(context.Background(), AccountIdentifier(testAccountID), "2c0fc9fa937b11eaa1b71c4d701ab86e", "34b12448945f11eaa1b71c4d701ab86e")
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestPollListTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()

	start := time.Now()
	err := client.pollListBulkOperation(ctx, AccountIdentifier(testAccountID), "list1")
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.WithinDuration(t, start, time.Now(), time.Second,
		"pollListBulkOperation took too much time with an expiring context")
}
