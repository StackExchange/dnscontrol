// Package cloudflare implements the Cloudflare v4 API.
package cloudflare

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-json"

	"golang.org/x/time/rate"
)

var (
	Version string = "v4"

	// Deprecated: Use `client.New` configuration instead.
	apiURL = fmt.Sprintf("%s://%s%s", defaultScheme, defaultHostname, defaultBasePath)
)

const (
	// AuthKeyEmail specifies that we should authenticate with API key and email address.
	AuthKeyEmail = 1 << iota
	// AuthUserService specifies that we should authenticate with a User-Service key.
	AuthUserService
	// AuthToken specifies that we should authenticate with an API Token.
	AuthToken
)

// API holds the configuration for the current API client. A client should not
// be modified concurrently.
type API struct {
	APIKey            string
	APIEmail          string
	APIUserServiceKey string
	APIToken          string
	BaseURL           string
	UserAgent         string
	headers           http.Header
	httpClient        *http.Client
	authType          int
	rateLimiter       *rate.Limiter
	retryPolicy       RetryPolicy
	logger            Logger
	Debug             bool
}

// newClient provides shared logic for New and NewWithUserServiceKey.
func newClient(opts ...Option) (*API, error) {
	silentLogger := log.New(io.Discard, "", log.LstdFlags)

	api := &API{
		BaseURL:     fmt.Sprintf("%s://%s%s", defaultScheme, defaultHostname, defaultBasePath),
		UserAgent:   userAgent + "/" + Version,
		headers:     make(http.Header),
		rateLimiter: rate.NewLimiter(rate.Limit(4), 1), // 4rps equates to default api limit (1200 req/5 min)
		retryPolicy: RetryPolicy{
			MaxRetries:    3,
			MinRetryDelay: 1 * time.Second,
			MaxRetryDelay: 30 * time.Second,
		},
		logger: silentLogger,
	}

	err := api.parseOptions(opts...)
	if err != nil {
		return nil, fmt.Errorf("options parsing failed: %w", err)
	}

	// Fall back to http.DefaultClient if the package user does not provide
	// their own.
	if api.httpClient == nil {
		api.httpClient = http.DefaultClient
	}

	return api, nil
}

// New creates a new Cloudflare v4 API client.
func New(key, email string, opts ...Option) (*API, error) {
	if key == "" || email == "" {
		return nil, errors.New(errEmptyCredentials)
	}

	api, err := newClient(opts...)
	if err != nil {
		return nil, err
	}

	api.APIKey = key
	api.APIEmail = email
	api.authType = AuthKeyEmail

	return api, nil
}

// NewWithAPIToken creates a new Cloudflare v4 API client using API Tokens.
func NewWithAPIToken(token string, opts ...Option) (*API, error) {
	if token == "" {
		return nil, errors.New(errEmptyAPIToken)
	}

	api, err := newClient(opts...)
	if err != nil {
		return nil, err
	}

	api.APIToken = token
	api.authType = AuthToken

	return api, nil
}

// NewWithUserServiceKey creates a new Cloudflare v4 API client using service key authentication.
func NewWithUserServiceKey(key string, opts ...Option) (*API, error) {
	if key == "" {
		return nil, errors.New(errEmptyCredentials)
	}

	api, err := newClient(opts...)
	if err != nil {
		return nil, err
	}

	api.APIUserServiceKey = key
	api.authType = AuthUserService

	return api, nil
}

// SetAuthType sets the authentication method (AuthKeyEmail, AuthToken, or AuthUserService).
func (api *API) SetAuthType(authType int) {
	api.authType = authType
}

// ZoneIDByName retrieves a zone's ID from the name.
func (api *API) ZoneIDByName(zoneName string) (string, error) {
	zoneName = normalizeZoneName(zoneName)
	res, err := api.ListZonesContext(context.Background(), WithZoneFilters(zoneName, "", ""))
	if err != nil {
		return "", fmt.Errorf("ListZonesContext command failed: %w", err)
	}

	switch len(res.Result) {
	case 0:
		return "", errors.New("zone could not be found")
	case 1:
		return res.Result[0].ID, nil
	default:
		return "", errors.New("ambiguous zone name; an account ID might help")
	}
}

// makeRequest makes a HTTP request and returns the body as a byte slice,
// closing it before returning. params will be serialized to JSON.
func (api *API) makeRequest(method, uri string, params interface{}) ([]byte, error) {
	return api.makeRequestWithAuthType(context.Background(), method, uri, params, api.authType)
}

func (api *API) makeRequestContext(ctx context.Context, method, uri string, params interface{}) ([]byte, error) {
	return api.makeRequestWithAuthType(ctx, method, uri, params, api.authType)
}

func (api *API) makeRequestContextWithHeaders(ctx context.Context, method, uri string, params interface{}, headers http.Header) ([]byte, error) {
	return api.makeRequestWithAuthTypeAndHeaders(ctx, method, uri, params, api.authType, headers)
}

func (api *API) makeRequestWithAuthType(ctx context.Context, method, uri string, params interface{}, authType int) ([]byte, error) {
	return api.makeRequestWithAuthTypeAndHeaders(ctx, method, uri, params, authType, nil)
}

// APIResponse holds the structure for a response from the API. It looks alot
// like `http.Response` however, uses a `[]byte` for the `Body` instead of a
// `io.ReadCloser`.
//
// This may go away in the experimental client in favour of `http.Response`.
type APIResponse struct {
	Body       []byte
	Status     string
	StatusCode int
	Headers    http.Header
}

func (api *API) makeRequestWithAuthTypeAndHeaders(ctx context.Context, method, uri string, params interface{}, authType int, headers http.Header) ([]byte, error) {
	res, err := api.makeRequestWithAuthTypeAndHeadersComplete(ctx, method, uri, params, authType, headers)
	if err != nil {
		return nil, err
	}
	return res.Body, err
}

// Use this method if an API response can have different Content-Type headers and different body formats.
func (api *API) makeRequestContextWithHeadersComplete(ctx context.Context, method, uri string, params interface{}, headers http.Header) (*APIResponse, error) {
	return api.makeRequestWithAuthTypeAndHeadersComplete(ctx, method, uri, params, api.authType, headers)
}

func (api *API) makeRequestWithAuthTypeAndHeadersComplete(ctx context.Context, method, uri string, params interface{}, authType int, headers http.Header) (*APIResponse, error) {
	var err error
	var resp *http.Response
	var respErr error
	var respBody []byte

	for i := 0; i <= api.retryPolicy.MaxRetries; i++ {
		var reqBody io.Reader
		if params != nil {
			if r, ok := params.(io.Reader); ok {
				reqBody = r
			} else if paramBytes, ok := params.([]byte); ok {
				reqBody = bytes.NewReader(paramBytes)
			} else {
				var jsonBody []byte
				jsonBody, err = json.Marshal(params)
				if err != nil {
					return nil, fmt.Errorf("error marshalling params to JSON: %w", err)
				}
				reqBody = bytes.NewReader(jsonBody)
			}
		}

		if i > 0 {
			// expect the backoff introduced here on errored requests to dominate the effect of rate limiting
			// don't need a random component here as the rate limiter should do something similar
			// nb time duration could truncate an arbitrary float. Since our inputs are all ints, we should be ok
			sleepDuration := time.Duration(math.Pow(2, float64(i-1)) * float64(api.retryPolicy.MinRetryDelay))

			if sleepDuration > api.retryPolicy.MaxRetryDelay {
				sleepDuration = api.retryPolicy.MaxRetryDelay
			}
			// useful to do some simple logging here, maybe introduce levels later
			api.logger.Printf("Sleeping %s before retry attempt number %d for request %s %s", sleepDuration.String(), i, method, uri)

			select {
			case <-time.After(sleepDuration):
			case <-ctx.Done():
				return nil, fmt.Errorf("operation aborted during backoff: %w", ctx.Err())
			}
		}

		err = api.rateLimiter.Wait(ctx)
		if err != nil {
			return nil, fmt.Errorf("error caused by request rate limiting: %w", err)
		}

		resp, respErr = api.request(ctx, method, uri, reqBody, authType, headers)

		// short circuit processing on context timeouts
		if respErr != nil && errors.Is(respErr, context.DeadlineExceeded) {
			return nil, respErr
		}

		// retry if the server is rate limiting us or if it failed
		// assumes server operations are rolled back on failure
		if respErr != nil || resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
				respErr = errors.New("exceeded available rate limit retries")
			}

			if respErr == nil {
				respErr = fmt.Errorf("received %s response (HTTP %d), please try again later", strings.ToLower(http.StatusText(resp.StatusCode)), resp.StatusCode)
			}
			continue
		} else {
			respBody, err = io.ReadAll(resp.Body)
			defer resp.Body.Close()
			if err != nil {
				return nil, fmt.Errorf("could not read response body: %w", err)
			}

			break
		}
	}

	// still had an error after all retries
	if respErr != nil {
		return nil, respErr
	}

	if resp.StatusCode >= http.StatusBadRequest {
		if strings.HasSuffix(resp.Request.URL.Path, "/filters/validate-expr") {
			return nil, fmt.Errorf("%s", respBody)
		}

		if resp.StatusCode >= http.StatusInternalServerError {
			return nil, &ServiceError{cloudflareError: &Error{
				StatusCode: resp.StatusCode,
				RayID:      resp.Header.Get("cf-ray"),
				Errors: []ResponseInfo{{
					Message: errInternalServiceError,
				}},
			}}
		}

		errBody := &Response{}
		err = json.Unmarshal(respBody, &errBody)
		if err != nil {
			return nil, fmt.Errorf(errUnmarshalErrorBody+": %w", err)
		}

		errCodes := make([]int, 0, len(errBody.Errors))
		errMsgs := make([]string, 0, len(errBody.Errors))
		for _, e := range errBody.Errors {
			errCodes = append(errCodes, e.Code)
			errMsgs = append(errMsgs, e.Message)
		}

		err := &Error{
			StatusCode:    resp.StatusCode,
			RayID:         resp.Header.Get("cf-ray"),
			Errors:        errBody.Errors,
			ErrorCodes:    errCodes,
			ErrorMessages: errMsgs,
			Messages:      errBody.Messages,
		}

		switch resp.StatusCode {
		case http.StatusUnauthorized:
			err.Type = ErrorTypeAuthorization
			return nil, &AuthorizationError{cloudflareError: err}
		case http.StatusForbidden:
			err.Type = ErrorTypeAuthentication
			return nil, &AuthenticationError{cloudflareError: err}
		case http.StatusNotFound:
			err.Type = ErrorTypeNotFound
			return nil, &NotFoundError{cloudflareError: err}
		case http.StatusTooManyRequests:
			err.Type = ErrorTypeRateLimit
			return nil, &RatelimitError{cloudflareError: err}
		default:
			err.Type = ErrorTypeRequest
			return nil, &RequestError{cloudflareError: err}
		}
	}

	return &APIResponse{
		Body:       respBody,
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Headers:    resp.Header,
	}, nil
}

// request makes a HTTP request to the given API endpoint, returning the raw
// *http.Response, or an error if one occurred. The caller is responsible for
// closing the response body.
func (api *API) request(ctx context.Context, method, uri string, reqBody io.Reader, authType int, headers http.Header) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, api.BaseURL+uri, reqBody)
	if err != nil {
		return nil, fmt.Errorf("HTTP request creation failed: %w", err)
	}

	combinedHeaders := make(http.Header)
	copyHeader(combinedHeaders, api.headers)
	copyHeader(combinedHeaders, headers)
	req.Header = combinedHeaders

	if authType&AuthKeyEmail != 0 {
		req.Header.Set("X-Auth-Key", api.APIKey)
		req.Header.Set("X-Auth-Email", api.APIEmail)
	}
	if authType&AuthUserService != 0 {
		req.Header.Set("X-Auth-User-Service-Key", api.APIUserServiceKey)
	}
	if authType&AuthToken != 0 {
		req.Header.Set("Authorization", "Bearer "+api.APIToken)
	}

	if api.UserAgent != "" {
		req.Header.Set("User-Agent", api.UserAgent)
	}

	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	if api.Debug {
		dump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			return nil, err
		}

		// Strip out any sensitive information from the request payload.
		sensitiveKeys := []string{api.APIKey, api.APIEmail, api.APIToken, api.APIUserServiceKey}
		for _, key := range sensitiveKeys {
			if key != "" {
				valueRegex := regexp.MustCompile(fmt.Sprintf("(?m)%s", key))
				dump = valueRegex.ReplaceAll(dump, []byte("[redacted]"))
			}
		}
		log.Printf("\n%s", string(dump))
	}

	resp, err := api.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	if api.Debug {
		dump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return resp, err
		}
		log.Printf("\n%s", string(dump))
	}

	return resp, nil
}

// copyHeader copies all headers for `source` and sets them on `target`.
// based on https://godoc.org/github.com/golang/gddo/httputil/header#Copy
func copyHeader(target, source http.Header) {
	for k, vs := range source {
		target[k] = vs
	}
}

// ResponseInfo contains a code and message returned by the API as errors or
// informational messages inside the response.
type ResponseInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Response is a template.  There will also be a result struct.  There will be a
// unique response type for each response, which will include this type.
type Response struct {
	Success  bool           `json:"success"`
	Errors   []ResponseInfo `json:"errors"`
	Messages []ResponseInfo `json:"messages"`
}

// ResultInfoCursors contains information about cursors.
type ResultInfoCursors struct {
	Before string `json:"before" url:"before,omitempty"`
	After  string `json:"after" url:"after,omitempty"`
}

// ResultInfo contains metadata about the Response.
type ResultInfo struct {
	Page       int               `json:"page" url:"page,omitempty"`
	PerPage    int               `json:"per_page" url:"per_page,omitempty"`
	TotalPages int               `json:"total_pages" url:"-"`
	Count      int               `json:"count" url:"-"`
	Total      int               `json:"total_count" url:"-"`
	Cursor     string            `json:"cursor" url:"cursor,omitempty"`
	Cursors    ResultInfoCursors `json:"cursors" url:"cursors,omitempty"`
}

// RawResponse keeps the result as JSON form.
type RawResponse struct {
	Response
	Result     json.RawMessage `json:"result"`
	ResultInfo *ResultInfo     `json:"result_info,omitempty"`
}

// Raw makes a HTTP request with user provided params and returns the
// result as a RawResponse, which contains the untouched JSON result.
func (api *API) Raw(ctx context.Context, method, endpoint string, data interface{}, headers http.Header) (RawResponse, error) {
	var r RawResponse
	res, err := api.makeRequestContextWithHeaders(ctx, method, endpoint, data, headers)
	if err != nil {
		return r, err
	}

	if err := json.Unmarshal(res, &r); err != nil {
		return r, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}
	return r, nil
}

// PaginationOptions can be passed to a list request to configure paging
// These values will be defaulted if omitted, and PerPage has min/max limits set by resource.
type PaginationOptions struct {
	Page    int `json:"page,omitempty" url:"page,omitempty"`
	PerPage int `json:"per_page,omitempty" url:"per_page,omitempty"`
}

// RetryPolicy specifies number of retries and min/max retry delays
// This config is used when the client exponentially backs off after errored requests.
type RetryPolicy struct {
	MaxRetries    int
	MinRetryDelay time.Duration
	MaxRetryDelay time.Duration
}

// Logger defines the interface this library needs to use logging
// This is a subset of the methods implemented in the log package.
type Logger interface {
	Printf(format string, v ...interface{})
}

// ReqOption is a functional option for configuring API requests.
type ReqOption func(opt *reqOption)

type reqOption struct {
	params url.Values
}

// WithZoneFilters applies a filter based on zone properties.
func WithZoneFilters(zoneName, accountID, status string) ReqOption {
	return func(opt *reqOption) {
		if zoneName != "" {
			opt.params.Set("name", normalizeZoneName(zoneName))
		}

		if accountID != "" {
			opt.params.Set("account.id", accountID)
		}

		if status != "" {
			opt.params.Set("status", status)
		}
	}
}

// WithPagination configures the pagination for a response.
func WithPagination(opts PaginationOptions) ReqOption {
	return func(opt *reqOption) {
		if opts.Page > 0 {
			opt.params.Set("page", strconv.Itoa(opts.Page))
		}

		if opts.PerPage > 0 {
			opt.params.Set("per_page", strconv.Itoa(opts.PerPage))
		}
	}
}

// checkResultInfo checks whether ResultInfo is reasonable except that it currently
// ignores the cursor information. perPage, page, and count are the requested #items
// per page, the requested page number, and the actual length of the Result array.
//
// Responses from the actual Cloudflare servers should pass all these checks (or we
// discover a serious bug in the Cloudflare servers). However, the unit tests can
// easily violate these constraints and this utility function can help debugging.
// Correct pagination information is crucial for more advanced List* functions that
// handle pagination automatically and fetch different pages in parallel.
//
// TODO: check cursors as well.
func checkResultInfo(perPage, page, count int, info *ResultInfo) bool {
	if info.Cursor != "" || info.Cursors.Before != "" || info.Cursors.After != "" {
		panic("checkResultInfo could not handle cursors yet.")
	}

	switch {
	case info.PerPage != perPage || info.Page != page || info.Count != count:
		return false

	case info.PerPage <= 0:
		return false

	case info.Total == 0 && info.TotalPages == 0 && info.Page == 1 && info.Count == 0:
		return true

	case info.Total <= 0 || info.TotalPages <= 0:
		return false

	case info.Total > info.PerPage*info.TotalPages || info.Total <= info.PerPage*(info.TotalPages-1):
		return false
	}

	switch {
	case info.Page > info.TotalPages || info.Page <= 0:
		return false

	case info.Page < info.TotalPages:
		return info.Count == info.PerPage

	case info.Page == info.TotalPages:
		return info.Count == info.Total-info.PerPage*(info.TotalPages-1)

	default:
		// This is actually impossible, but Go compiler does not know trichotomy
		panic("checkResultInfo: impossible")
	}
}

type OrderDirection string

const (
	OrderDirectionAsc  OrderDirection = "asc"
	OrderDirectionDesc OrderDirection = "desc"
)
