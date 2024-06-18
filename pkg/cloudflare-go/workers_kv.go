package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/goccy/go-json"
)

// CreateWorkersKVNamespaceParams provides parameters for creating and updating storage namespaces.
type CreateWorkersKVNamespaceParams struct {
	Title string `json:"title"`
}

type UpdateWorkersKVNamespaceParams struct {
	NamespaceID string `json:"-"`
	Title       string `json:"title"`
}

// WorkersKVPair is used in an array in the request to the bulk KV api.
type WorkersKVPair struct {
	Key           string      `json:"key"`
	Value         string      `json:"value"`
	Expiration    int         `json:"expiration,omitempty"`
	ExpirationTTL int         `json:"expiration_ttl,omitempty"`
	Metadata      interface{} `json:"metadata,omitempty"`
	Base64        bool        `json:"base64,omitempty"`
}

// WorkersKVNamespaceResponse is the response received when creating storage namespaces.
type WorkersKVNamespaceResponse struct {
	Response
	Result WorkersKVNamespace `json:"result"`
}

// WorkersKVNamespace contains the unique identifier and title of a storage namespace.
type WorkersKVNamespace struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// ListWorkersKVNamespacesResponse contains a slice of storage namespaces associated with an
// account, pagination information, and an embedded response struct.
type ListWorkersKVNamespacesResponse struct {
	Response
	Result     []WorkersKVNamespace `json:"result"`
	ResultInfo `json:"result_info"`
}

// StorageKey is a key name used to identify a storage value.
type StorageKey struct {
	Name       string      `json:"name"`
	Expiration int         `json:"expiration"`
	Metadata   interface{} `json:"metadata"`
}

// ListStorageKeysResponse contains a slice of keys belonging to a storage namespace,
// pagination information, and an embedded response struct.
type ListStorageKeysResponse struct {
	Response
	Result     []StorageKey `json:"result"`
	ResultInfo `json:"result_info"`
}

type ListWorkersKVNamespacesParams struct {
	ResultInfo
}

type WriteWorkersKVEntryParams struct {
	NamespaceID string
	Key         string
	Value       []byte
}

type WriteWorkersKVEntriesParams struct {
	NamespaceID string
	KVs         []*WorkersKVPair
}

type GetWorkersKVParams struct {
	NamespaceID string
	Key         string
}

type DeleteWorkersKVEntryParams struct {
	NamespaceID string
	Key         string
}

type DeleteWorkersKVEntriesParams struct {
	NamespaceID string
	Keys        []string
}

type ListWorkersKVsParams struct {
	NamespaceID string `url:"-"`
	Limit       int    `url:"limit,omitempty"`
	Cursor      string `url:"cursor,omitempty"`
	Prefix      string `url:"prefix,omitempty"`
}

// CreateWorkersKVNamespace creates a namespace under the given title.
// A 400 is returned if the account already owns a namespace with this title.
// A namespace must be explicitly deleted to be replaced.
//
// API reference: https://developers.cloudflare.com/api/operations/workers-kv-namespace-create-a-namespace
func (api *API) CreateWorkersKVNamespace(ctx context.Context, rc *ResourceContainer, params CreateWorkersKVNamespaceParams) (WorkersKVNamespaceResponse, error) {
	if rc.Level != AccountRouteLevel {
		return WorkersKVNamespaceResponse{}, ErrRequiredAccountLevelResourceContainer
	}

	if rc.Identifier == "" {
		return WorkersKVNamespaceResponse{}, ErrMissingIdentifier
	}
	uri := fmt.Sprintf("/accounts/%s/storage/kv/namespaces", rc.Identifier)
	res, err := api.makeRequestContext(ctx, http.MethodPost, uri, params)
	if err != nil {
		return WorkersKVNamespaceResponse{}, err
	}

	result := WorkersKVNamespaceResponse{}
	if err := json.Unmarshal(res, &result); err != nil {
		return result, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return result, err
}

// ListWorkersKVNamespaces lists storage namespaces.
//
// API reference: https://developers.cloudflare.com/api/operations/workers-kv-namespace-list-namespaces
func (api *API) ListWorkersKVNamespaces(ctx context.Context, rc *ResourceContainer, params ListWorkersKVNamespacesParams) ([]WorkersKVNamespace, *ResultInfo, error) {
	if rc.Level != AccountRouteLevel {
		return []WorkersKVNamespace{}, &ResultInfo{}, ErrRequiredAccountLevelResourceContainer
	}

	if rc.Identifier == "" {
		return []WorkersKVNamespace{}, &ResultInfo{}, ErrMissingIdentifier
	}

	autoPaginate := true
	if params.PerPage >= 1 || params.Page >= 1 {
		autoPaginate = false
	}
	if params.PerPage < 1 {
		params.PerPage = 50
	}
	if params.Page < 1 {
		params.Page = 1
	}

	var namespaces []WorkersKVNamespace
	var nsResponse ListWorkersKVNamespacesResponse
	for {
		nsResponse = ListWorkersKVNamespacesResponse{}
		uri := buildURI(fmt.Sprintf("/accounts/%s/storage/kv/namespaces", rc.Identifier), params)

		res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
		if err != nil {
			return []WorkersKVNamespace{}, &ResultInfo{}, err
		}

		err = json.Unmarshal(res, &nsResponse)
		if err != nil {
			return []WorkersKVNamespace{}, &ResultInfo{}, fmt.Errorf("failed to unmarshal workers KV namespaces JSON data: %w", err)
		}

		namespaces = append(namespaces, nsResponse.Result...)
		params.ResultInfo = nsResponse.ResultInfo.Next()

		if params.ResultInfo.Done() || !autoPaginate {
			break
		}
	}

	return namespaces, &nsResponse.ResultInfo, nil
}

// DeleteWorkersKVNamespace deletes the namespace corresponding to the given ID.
//
// API reference: https://developers.cloudflare.com/api/operations/workers-kv-namespace-remove-a-namespace
func (api *API) DeleteWorkersKVNamespace(ctx context.Context, rc *ResourceContainer, namespaceID string) (Response, error) {
	uri := fmt.Sprintf("/accounts/%s/storage/kv/namespaces/%s", rc.Identifier, namespaceID)
	res, err := api.makeRequestContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return Response{}, err
	}

	result := Response{}
	if err := json.Unmarshal(res, &result); err != nil {
		return result, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return result, err
}

// UpdateWorkersKVNamespace modifies a KV namespace based on the ID.
//
// API reference: https://developers.cloudflare.com/api/operations/workers-kv-namespace-rename-a-namespace
func (api *API) UpdateWorkersKVNamespace(ctx context.Context, rc *ResourceContainer, params UpdateWorkersKVNamespaceParams) (Response, error) {
	if rc.Level != AccountRouteLevel {
		return Response{}, ErrRequiredAccountLevelResourceContainer
	}

	if rc.Identifier == "" {
		return Response{}, ErrMissingIdentifier
	}

	uri := fmt.Sprintf("/accounts/%s/storage/kv/namespaces/%s", rc.Identifier, params.NamespaceID)
	res, err := api.makeRequestContext(ctx, http.MethodPut, uri, params)
	if err != nil {
		return Response{}, err
	}

	result := Response{}
	if err := json.Unmarshal(res, &result); err != nil {
		return result, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return result, err
}

// WriteWorkersKVEntry writes a single KV value based on the key.
//
// API reference: https://developers.cloudflare.com/api/operations/workers-kv-namespace-write-key-value-pair-with-metadata
func (api *API) WriteWorkersKVEntry(ctx context.Context, rc *ResourceContainer, params WriteWorkersKVEntryParams) (Response, error) {
	if rc.Level != AccountRouteLevel {
		return Response{}, ErrRequiredAccountLevelResourceContainer
	}

	if rc.Identifier == "" {
		return Response{}, ErrMissingIdentifier
	}

	uri := fmt.Sprintf("/accounts/%s/storage/kv/namespaces/%s/values/%s", rc.Identifier, params.NamespaceID, url.PathEscape(params.Key))
	res, err := api.makeRequestContextWithHeaders(
		ctx, http.MethodPut, uri, params.Value, http.Header{"Content-Type": []string{"application/octet-stream"}},
	)
	if err != nil {
		return Response{}, err
	}

	result := Response{}
	if err := json.Unmarshal(res, &result); err != nil {
		return result, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return result, err
}

// WriteWorkersKVEntries writes multiple KVs at once.
//
// API reference: https://developers.cloudflare.com/api/operations/workers-kv-namespace-write-multiple-key-value-pairs
func (api *API) WriteWorkersKVEntries(ctx context.Context, rc *ResourceContainer, params WriteWorkersKVEntriesParams) (Response, error) {
	if rc.Level != AccountRouteLevel {
		return Response{}, ErrRequiredAccountLevelResourceContainer
	}

	if rc.Identifier == "" {
		return Response{}, ErrMissingIdentifier
	}

	uri := fmt.Sprintf("/accounts/%s/storage/kv/namespaces/%s/bulk", rc.Identifier, params.NamespaceID)
	res, err := api.makeRequestContextWithHeaders(
		ctx, http.MethodPut, uri, params.KVs, http.Header{"Content-Type": []string{"application/json"}},
	)
	if err != nil {
		return Response{}, err
	}

	result := Response{}
	if err := json.Unmarshal(res, &result); err != nil {
		return result, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return result, err
}

// GetWorkersKV returns the value associated with the given key in the
// given namespace.
//
// API reference: https://developers.cloudflare.com/api/operations/workers-kv-namespace-read-key-value-pair
func (api API) GetWorkersKV(ctx context.Context, rc *ResourceContainer, params GetWorkersKVParams) ([]byte, error) {
	if rc.Level != AccountRouteLevel {
		return []byte(``), ErrRequiredAccountLevelResourceContainer
	}

	if rc.Identifier == "" {
		return []byte(``), ErrMissingIdentifier
	}
	uri := fmt.Sprintf("/accounts/%s/storage/kv/namespaces/%s/values/%s", rc.Identifier, params.NamespaceID, url.PathEscape(params.Key))
	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// DeleteWorkersKVEntry deletes a key and value for a provided storage namespace.
//
// API reference: https://developers.cloudflare.com/api/operations/workers-kv-namespace-delete-key-value-pair
func (api API) DeleteWorkersKVEntry(ctx context.Context, rc *ResourceContainer, params DeleteWorkersKVEntryParams) (Response, error) {
	if rc.Level != AccountRouteLevel {
		return Response{}, ErrRequiredAccountLevelResourceContainer
	}

	if rc.Identifier == "" {
		return Response{}, ErrMissingIdentifier
	}
	uri := fmt.Sprintf("/accounts/%s/storage/kv/namespaces/%s/values/%s", rc.Identifier, params.NamespaceID, url.PathEscape(params.Key))
	res, err := api.makeRequestContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return Response{}, err
	}

	result := Response{}
	if err := json.Unmarshal(res, &result); err != nil {
		return result, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}
	return result, err
}

// DeleteWorkersKVEntries deletes multiple KVs at once.
//
// API reference: https://developers.cloudflare.com/api/operations/workers-kv-namespace-delete-multiple-key-value-pairs
func (api *API) DeleteWorkersKVEntries(ctx context.Context, rc *ResourceContainer, params DeleteWorkersKVEntriesParams) (Response, error) {
	if rc.Level != AccountRouteLevel {
		return Response{}, ErrRequiredAccountLevelResourceContainer
	}

	if rc.Identifier == "" {
		return Response{}, ErrMissingIdentifier
	}
	uri := fmt.Sprintf("/accounts/%s/storage/kv/namespaces/%s/bulk", rc.Identifier, params.NamespaceID)
	res, err := api.makeRequestContextWithHeaders(
		ctx, http.MethodDelete, uri, params.Keys, http.Header{"Content-Type": []string{"application/json"}},
	)
	if err != nil {
		return Response{}, err
	}

	result := Response{}
	if err := json.Unmarshal(res, &result); err != nil {
		return result, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return result, err
}

// ListWorkersKVKeys lists a namespace's keys.
//
// API Reference: https://developers.cloudflare.com/api/operations/workers-kv-namespace-list-a-namespace'-s-keys
func (api API) ListWorkersKVKeys(ctx context.Context, rc *ResourceContainer, params ListWorkersKVsParams) (ListStorageKeysResponse, error) {
	if rc.Level != AccountRouteLevel {
		return ListStorageKeysResponse{}, ErrRequiredAccountLevelResourceContainer
	}

	if rc.Identifier == "" {
		return ListStorageKeysResponse{}, ErrMissingIdentifier
	}

	uri := buildURI(
		fmt.Sprintf("/accounts/%s/storage/kv/namespaces/%s/keys", rc.Identifier, params.NamespaceID),
		params,
	)
	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return ListStorageKeysResponse{}, err
	}

	result := ListStorageKeysResponse{}
	if err := json.Unmarshal(res, &result); err != nil {
		return result, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}
	return result, err
}
