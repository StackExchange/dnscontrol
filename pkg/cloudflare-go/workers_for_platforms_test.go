package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	listDispatchNamespaces = `{
    "result": [
		{
            "namespace_id": "6446f71d-13b3-4bbc-a8a4-9e18760499c8",
            "namespace_name": "test",
            "created_on": "2024-02-20T17:26:15.4134Z",
            "created_by": "4e599df4216133509abaac54b109a647",
            "modified_on": "2024-02-20T17:26:15.4134Z",
            "modified_by": "4e599df4216133509abaac54b109a647"
        },
        {
            "namespace_id": "d6851dad-d412-4509-ae13-a364bc5f125a",
            "namespace_name": "test-2",
            "created_on": "2024-02-20T20:28:36.560575Z",
            "created_by": "4e599df4216133509abaac54b109a647",
            "modified_on": "2024-02-20T20:28:36.560575Z",
            "modified_by": "4e599df4216133509abaac54b109a647"
        }
	],
    "success": true,
    "errors": [],
    "messages": []
}`

	getDispatchNamespace = `{
	"result": {
		"namespace_id": "6446f71d-13b3-4bbc-a8a4-9e18760499c8",
		"namespace_name": "test",
		"created_on": "2024-02-20T17:26:15.4134Z",
		"created_by": "4e599df4216133509abaac54b109a647",
		"modified_on": "2024-02-20T17:26:15.4134Z",
		"modified_by": "4e599df4216133509abaac54b109a647"
	},
    "success": true,
    "errors": [],
    "messages": []
}`

	deleteDispatchNamespace = `{
    "result": null,
    "success": true,
    "errors": [],
    "messages": []
}`
)

func TestListWorkersForPlatformsDispatchNamespaces(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/workers/dispatch/namespaces", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, listDispatchNamespaces)
	})

	res, err := client.ListWorkersForPlatformsDispatchNamespaces(context.Background(), AccountIdentifier(testAccountID))

	assert.NoError(t, err)
	assert.Len(t, res.Result, 2)

	assert.Equal(t, "6446f71d-13b3-4bbc-a8a4-9e18760499c8", res.Result[0].NamespaceId)
	assert.Equal(t, "test", res.Result[0].NamespaceName)
	assert.Equal(t, "4e599df4216133509abaac54b109a647", res.Result[0].CreatedBy)
	assert.Equal(t, "4e599df4216133509abaac54b109a647", res.Result[0].ModifiedBy)

	assert.Equal(t, "d6851dad-d412-4509-ae13-a364bc5f125a", res.Result[1].NamespaceId)
	assert.Equal(t, "test-2", res.Result[1].NamespaceName)
	assert.Equal(t, "4e599df4216133509abaac54b109a647", res.Result[1].CreatedBy)
	assert.Equal(t, "4e599df4216133509abaac54b109a647", res.Result[1].ModifiedBy)
}

func TestGetWorkersForPlatformsDispatchNamespace(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/workers/dispatch/namespaces/test", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, getDispatchNamespace)
	})

	res, err := client.GetWorkersForPlatformsDispatchNamespace(context.Background(), AccountIdentifier(testAccountID), "test")

	assert.NoError(t, err)

	assert.Equal(t, "6446f71d-13b3-4bbc-a8a4-9e18760499c8", res.Result.NamespaceId)
	assert.Equal(t, "test", res.Result.NamespaceName)
	assert.Equal(t, "4e599df4216133509abaac54b109a647", res.Result.CreatedBy)
	assert.Equal(t, "4e599df4216133509abaac54b109a647", res.Result.ModifiedBy)
}

func TestCreateWorkersForPlatformsDispatchNamespace(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/workers/dispatch/namespaces", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, getDispatchNamespace)
	})

	res, err := client.CreateWorkersForPlatformsDispatchNamespace(context.Background(), AccountIdentifier(testAccountID), CreateWorkersForPlatformsDispatchNamespaceParams{
		Name: "test",
	})

	assert.NoError(t, err)

	assert.Equal(t, "6446f71d-13b3-4bbc-a8a4-9e18760499c8", res.Result.NamespaceId)
	assert.Equal(t, "test", res.Result.NamespaceName)
	assert.Equal(t, "4e599df4216133509abaac54b109a647", res.Result.CreatedBy)
	assert.Equal(t, "4e599df4216133509abaac54b109a647", res.Result.ModifiedBy)
}

func TestDeleteWorkersForPlatformsDispatchNamespace(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/workers/dispatch/namespaces/test", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, deleteDispatchNamespace)
	})

	err := client.DeleteWorkersForPlatformsDispatchNamespace(context.Background(), AccountIdentifier(testAccountID), "test")

	assert.NoError(t, err)
}
