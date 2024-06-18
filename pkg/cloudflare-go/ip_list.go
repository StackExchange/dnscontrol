package cloudflare

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/goccy/go-json"
)

// The definitions in this file are deprecated and should be removed after
// enough time is given for users to migrate. Use the more general `List`
// methods instead.

const (
	// IPListTypeIP specifies a list containing IP addresses.
	IPListTypeIP = "ip"
)

// IPListBulkOperation contains information about a Bulk Operation.
type IPListBulkOperation struct {
	ID        string     `json:"id"`
	Status    string     `json:"status"`
	Error     string     `json:"error"`
	Completed *time.Time `json:"completed"`
}

// IPList contains information about an IP List.
type IPList struct {
	ID                    string     `json:"id"`
	Name                  string     `json:"name"`
	Description           string     `json:"description"`
	Kind                  string     `json:"kind"`
	NumItems              int        `json:"num_items"`
	NumReferencingFilters int        `json:"num_referencing_filters"`
	CreatedOn             *time.Time `json:"created_on"`
	ModifiedOn            *time.Time `json:"modified_on"`
}

// IPListItem contains information about a single IP List Item.
type IPListItem struct {
	ID         string     `json:"id"`
	IP         string     `json:"ip"`
	Comment    string     `json:"comment"`
	CreatedOn  *time.Time `json:"created_on"`
	ModifiedOn *time.Time `json:"modified_on"`
}

// IPListCreateRequest contains data for a new IP List.
type IPListCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Kind        string `json:"kind"`
}

// IPListItemCreateRequest contains data for a new IP List Item.
type IPListItemCreateRequest struct {
	IP      string `json:"ip"`
	Comment string `json:"comment"`
}

// IPListItemDeleteRequest wraps IP List Items that shall be deleted.
type IPListItemDeleteRequest struct {
	Items []IPListItemDeleteItemRequest `json:"items"`
}

// IPListItemDeleteItemRequest contains single IP List Items that shall be deleted.
type IPListItemDeleteItemRequest struct {
	ID string `json:"id"`
}

// IPListUpdateRequest contains data for an IP List update.
type IPListUpdateRequest struct {
	Description string `json:"description"`
}

// IPListResponse contains a single IP List.
type IPListResponse struct {
	Response
	Result IPList `json:"result"`
}

// IPListItemCreateResponse contains information about the creation of an IP List Item.
type IPListItemCreateResponse struct {
	Response
	Result struct {
		OperationID string `json:"operation_id"`
	} `json:"result"`
}

// IPListListResponse contains a slice of IP Lists.
type IPListListResponse struct {
	Response
	Result []IPList `json:"result"`
}

// IPListBulkOperationResponse contains information about a Bulk Operation.
type IPListBulkOperationResponse struct {
	Response
	Result IPListBulkOperation `json:"result"`
}

// IPListDeleteResponse contains information about the deletion of an IP List.
type IPListDeleteResponse struct {
	Response
	Result struct {
		ID string `json:"id"`
	} `json:"result"`
}

// IPListItemsListResponse contains information about IP List Items.
type IPListItemsListResponse struct {
	Response
	ResultInfo `json:"result_info"`
	Result     []IPListItem `json:"result"`
}

// IPListItemDeleteResponse contains information about the deletion of an IP List Item.
type IPListItemDeleteResponse struct {
	Response
	Result struct {
		OperationID string `json:"operation_id"`
	} `json:"result"`
}

// IPListItemsGetResponse contains information about a single IP List Item.
type IPListItemsGetResponse struct {
	Response
	Result IPListItem `json:"result"`
}

// ListIPLists lists all IP Lists.
//
// API reference: https://api.cloudflare.com/#rules-lists-list-lists
//
// Deprecated: Use `ListLists` instead.
func (api *API) ListIPLists(ctx context.Context, accountID string) ([]IPList, error) {
	uri := fmt.Sprintf("/accounts/%s/rules/lists", accountID)
	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return []IPList{}, err
	}

	result := IPListListResponse{}
	if err := json.Unmarshal(res, &result); err != nil {
		return []IPList{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return result.Result, nil
}

// CreateIPList creates a new IP List.
//
// API reference: https://api.cloudflare.com/#rules-lists-create-list
//
// Deprecated: Use `CreateList` instead.
func (api *API) CreateIPList(ctx context.Context, accountID, name, description, kind string) (IPList,
	error) {
	uri := fmt.Sprintf("/accounts/%s/rules/lists", accountID)
	res, err := api.makeRequestContext(ctx, http.MethodPost, uri,
		IPListCreateRequest{Name: name, Description: description, Kind: kind})
	if err != nil {
		return IPList{}, err
	}

	result := IPListResponse{}
	if err := json.Unmarshal(res, &result); err != nil {
		return IPList{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return result.Result, nil
}

// GetIPList returns a single IP List
//
// API reference: https://api.cloudflare.com/#rules-lists-get-list
//
// Deprecated: Use `GetList` instead.
func (api *API) GetIPList(ctx context.Context, accountID, ID string) (IPList, error) {
	uri := fmt.Sprintf("/accounts/%s/rules/lists/%s", accountID, ID)
	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return IPList{}, err
	}

	result := IPListResponse{}
	if err := json.Unmarshal(res, &result); err != nil {
		return IPList{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return result.Result, nil
}

// UpdateIPList updates the description of an existing IP List.
//
// API reference: https://api.cloudflare.com/#rules-lists-update-list
//
// Deprecated: Use `UpdateList` instead.
func (api *API) UpdateIPList(ctx context.Context, accountID, ID, description string) (IPList, error) {
	uri := fmt.Sprintf("/accounts/%s/rules/lists/%s", accountID, ID)
	res, err := api.makeRequestContext(ctx, http.MethodPut, uri, IPListUpdateRequest{Description: description})
	if err != nil {
		return IPList{}, err
	}

	result := IPListResponse{}
	if err := json.Unmarshal(res, &result); err != nil {
		return IPList{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return result.Result, nil
}

// DeleteIPList deletes an IP List.
//
// API reference: https://api.cloudflare.com/#rules-lists-delete-list
//
// Deprecated: Use `DeleteList` instead.
func (api *API) DeleteIPList(ctx context.Context, accountID, ID string) (IPListDeleteResponse, error) {
	uri := fmt.Sprintf("/accounts/%s/rules/lists/%s", accountID, ID)
	res, err := api.makeRequestContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return IPListDeleteResponse{}, err
	}

	result := IPListDeleteResponse{}
	if err := json.Unmarshal(res, &result); err != nil {
		return IPListDeleteResponse{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return result, nil
}

// ListIPListItems returns a list with all items in an IP List.
//
// API reference: https://api.cloudflare.com/#rules-lists-list-list-items
//
// Deprecated: Use `ListListItems` instead.
func (api *API) ListIPListItems(ctx context.Context, accountID, ID string) ([]IPListItem, error) {
	var list []IPListItem
	var cursor string
	var cursorQuery string

	for {
		if len(cursor) > 0 {
			cursorQuery = fmt.Sprintf("?cursor=%s", cursor)
		}
		uri := fmt.Sprintf("/accounts/%s/rules/lists/%s/items%s", accountID, ID, cursorQuery)
		res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
		if err != nil {
			return []IPListItem{}, err
		}

		result := IPListItemsListResponse{}
		if err := json.Unmarshal(res, &result); err != nil {
			return []IPListItem{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
		}

		list = append(list, result.Result...)
		if cursor = result.ResultInfo.Cursors.After; cursor == "" {
			break
		}
	}

	return list, nil
}

// CreateIPListItemAsync creates a new IP List Item asynchronously. Users have
// to poll the operation status by using the operation_id returned by this
// function.
//
// API reference: https://api.cloudflare.com/#rules-lists-create-list-items
//
// Deprecated: Use `CreateListItemAsync` instead.
func (api *API) CreateIPListItemAsync(ctx context.Context, accountID, ID, ip, comment string) (IPListItemCreateResponse, error) {
	uri := fmt.Sprintf("/accounts/%s/rules/lists/%s/items", accountID, ID)
	res, err := api.makeRequestContext(ctx, http.MethodPost, uri, []IPListItemCreateRequest{{IP: ip, Comment: comment}})
	if err != nil {
		return IPListItemCreateResponse{}, err
	}

	result := IPListItemCreateResponse{}
	if err := json.Unmarshal(res, &result); err != nil {
		return IPListItemCreateResponse{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return result, nil
}

// CreateIPListItem creates a new IP List Item synchronously and returns the
// current set of IP List Items.
//
// Deprecated: Use `CreateListItem` instead.
func (api *API) CreateIPListItem(ctx context.Context, accountID, ID, ip, comment string) ([]IPListItem, error) {
	result, err := api.CreateIPListItemAsync(ctx, accountID, ID, ip, comment)

	if err != nil {
		return []IPListItem{}, err
	}

	err = api.pollIPListBulkOperation(ctx, accountID, result.Result.OperationID)
	if err != nil {
		return []IPListItem{}, err
	}

	return api.ListIPListItems(ctx, accountID, ID)
}

// CreateIPListItemsAsync bulk creates many IP List Items asynchronously. Users
// have to poll the operation status by using the operation_id returned by this
// function.
//
// API reference: https://api.cloudflare.com/#rules-lists-create-list-items
//
// Deprecated: Use `CreateListItemsAsync` instead.
func (api *API) CreateIPListItemsAsync(ctx context.Context, accountID, ID string, items []IPListItemCreateRequest) (
	IPListItemCreateResponse, error) {
	uri := fmt.Sprintf("/accounts/%s/rules/lists/%s/items", accountID, ID)
	res, err := api.makeRequestContext(ctx, http.MethodPost, uri, items)
	if err != nil {
		return IPListItemCreateResponse{}, err
	}

	result := IPListItemCreateResponse{}
	if err := json.Unmarshal(res, &result); err != nil {
		return IPListItemCreateResponse{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return result, nil
}

// CreateIPListItems bulk creates many IP List Items synchronously and returns
// the current set of IP List Items.
//
// Deprecated: Use `CreateListItems` instead.
func (api *API) CreateIPListItems(ctx context.Context, accountID, ID string, items []IPListItemCreateRequest) (
	[]IPListItem, error) {
	result, err := api.CreateIPListItemsAsync(ctx, accountID, ID, items)
	if err != nil {
		return []IPListItem{}, err
	}

	err = api.pollIPListBulkOperation(ctx, accountID, result.Result.OperationID)
	if err != nil {
		return []IPListItem{}, err
	}

	return api.ListIPListItems(ctx, accountID, ID)
}

// ReplaceIPListItemsAsync replaces all IP List Items asynchronously. Users have
// to poll the operation status by using the operation_id returned by this
// function.
//
// API reference: https://api.cloudflare.com/#rules-lists-replace-list-items
//
// Deprecated: Use `ReplaceListItemsAsync` instead.
func (api *API) ReplaceIPListItemsAsync(ctx context.Context, accountID, ID string, items []IPListItemCreateRequest) (
	IPListItemCreateResponse, error) {
	uri := fmt.Sprintf("/accounts/%s/rules/lists/%s/items", accountID, ID)
	res, err := api.makeRequestContext(ctx, http.MethodPut, uri, items)
	if err != nil {
		return IPListItemCreateResponse{}, err
	}

	result := IPListItemCreateResponse{}
	if err := json.Unmarshal(res, &result); err != nil {
		return IPListItemCreateResponse{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return result, nil
}

// ReplaceIPListItems replaces all IP List Items synchronously and returns the
// current set of IP List Items.
//
// Deprecated: Use `ReplaceListItems` instead.
func (api *API) ReplaceIPListItems(ctx context.Context, accountID, ID string, items []IPListItemCreateRequest) (
	[]IPListItem, error) {
	result, err := api.ReplaceIPListItemsAsync(ctx, accountID, ID, items)
	if err != nil {
		return []IPListItem{}, err
	}

	err = api.pollIPListBulkOperation(ctx, accountID, result.Result.OperationID)
	if err != nil {
		return []IPListItem{}, err
	}

	return api.ListIPListItems(ctx, accountID, ID)
}

// DeleteIPListItemsAsync removes specific Items of an IP List by their ID
// asynchronously. Users have to poll the operation status by using the
// operation_id returned by this function.
//
// API reference: https://api.cloudflare.com/#rules-lists-delete-list-items
//
// Deprecated: Use `DeleteListItemsAsync` instead.
func (api *API) DeleteIPListItemsAsync(ctx context.Context, accountID, ID string, items IPListItemDeleteRequest) (
	IPListItemDeleteResponse, error) {
	uri := fmt.Sprintf("/accounts/%s/rules/lists/%s/items", accountID, ID)
	res, err := api.makeRequestContext(ctx, http.MethodDelete, uri, items)
	if err != nil {
		return IPListItemDeleteResponse{}, err
	}

	result := IPListItemDeleteResponse{}
	if err := json.Unmarshal(res, &result); err != nil {
		return IPListItemDeleteResponse{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return result, nil
}

// DeleteIPListItems removes specific Items of an IP List by their ID
// synchronously and returns the current set of IP List Items.
//
// Deprecated: Use `DeleteListItems` instead.
func (api *API) DeleteIPListItems(ctx context.Context, accountID, ID string, items IPListItemDeleteRequest) (
	[]IPListItem, error) {
	result, err := api.DeleteIPListItemsAsync(ctx, accountID, ID, items)
	if err != nil {
		return []IPListItem{}, err
	}

	err = api.pollIPListBulkOperation(ctx, accountID, result.Result.OperationID)
	if err != nil {
		return []IPListItem{}, err
	}

	return api.ListIPListItems(ctx, accountID, ID)
}

// GetIPListItem returns a single IP List Item.
//
// API reference: https://api.cloudflare.com/#rules-lists-get-list-item
//
// Deprecated: Use `GetListItem` instead.
func (api *API) GetIPListItem(ctx context.Context, accountID, listID, id string) (IPListItem, error) {
	uri := fmt.Sprintf("/accounts/%s/rules/lists/%s/items/%s", accountID, listID, id)
	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return IPListItem{}, err
	}

	result := IPListItemsGetResponse{}
	if err := json.Unmarshal(res, &result); err != nil {
		return IPListItem{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return result.Result, nil
}

// GetIPListBulkOperation returns the status of a bulk operation.
//
// API reference: https://api.cloudflare.com/#rules-lists-get-bulk-operation
//
// Deprecated: Use `GetListBulkOperation` instead.
func (api *API) GetIPListBulkOperation(ctx context.Context, accountID, ID string) (IPListBulkOperation, error) {
	uri := fmt.Sprintf("/accounts/%s/rules/lists/bulk_operations/%s", accountID, ID)
	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return IPListBulkOperation{}, err
	}

	result := IPListBulkOperationResponse{}
	if err := json.Unmarshal(res, &result); err != nil {
		return IPListBulkOperation{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return result.Result, nil
}

// pollIPListBulkOperation implements synchronous behaviour for some
// asynchronous endpoints. bulk-operation status can be either pending, running,
// failed or completed.
func (api *API) pollIPListBulkOperation(ctx context.Context, accountID, ID string) error {
	for i := uint8(0); i < 16; i++ {
		sleepDuration := 1 << (i / 2) * time.Second
		select {
		case <-time.After(sleepDuration):
		case <-ctx.Done():
			return fmt.Errorf("operation aborted during backoff: %w", ctx.Err())
		}

		bulkResult, err := api.GetIPListBulkOperation(ctx, accountID, ID)
		if err != nil {
			return err
		}

		switch bulkResult.Status {
		case "failed":
			return errors.New(bulkResult.Error)
		case "pending", "running":
			continue
		case "completed":
			return nil
		default:
			return fmt.Errorf("%s: %s", errOperationUnexpectedStatus, bulkResult.Status)
		}
	}

	return errors.New(errOperationStillRunning)
}
