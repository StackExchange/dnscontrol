package cloudflare

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

const (
	errEmptyCredentials                       = "invalid credentials: key & email must not be empty" //nolint:gosec,unused
	errEmptyAPIToken                          = "invalid credentials: API Token must not be empty"   //nolint:gosec,unused
	errInternalServiceError                   = "internal service error"
	errMakeRequestError                       = "error from makeRequest"
	errUnmarshalError                         = "error unmarshalling the JSON response"
	errUnmarshalErrorBody                     = "error unmarshalling the JSON response error body"
	errRequestNotSuccessful                   = "error reported by API"
	errMissingAccountID                       = "required missing account ID"
	errMissingZoneID                          = "required missing zone ID"
	errMissingAccountOrZoneID                 = "either account ID or zone ID must be provided"
	errAccountIDAndZoneIDAreMutuallyExclusive = "account ID and zone ID are mutually exclusive"
	errMissingResourceIdentifier              = "required missing resource identifier"
	errOperationStillRunning                  = "bulk operation did not finish before timeout"
	errOperationUnexpectedStatus              = "bulk operation returned an unexpected status"
	errResultInfo                             = "incorrect pagination info (result_info) in responses"
	errManualPagination                       = "unexpected pagination options passed to functions that handle pagination automatically"
	errInvalidResourceIdentifer               = "invalid resource identifier: %s"
	errInvalidZoneIdentifer                   = "invalid zone identifier: %s"
	errAPIKeysAndTokensAreMutuallyExclusive   = "API keys and tokens are mutually exclusive" //nolint:gosec
	errMissingCredentials                     = "no credentials provided"

	errInvalidResourceContainerAccess        = "requested resource container (%q) is not supported for this endpoint"
	errRequiredAccountLevelResourceContainer = "this endpoint requires using an account level resource container and identifiers"
	errRequiredZoneLevelResourceContainer    = "this endpoint requires using a zone level resource container and identifiers"
)

var (
	ErrAPIKeysAndTokensAreMutuallyExclusive   = errors.New(errAPIKeysAndTokensAreMutuallyExclusive)
	ErrMissingCredentials                     = errors.New(errMissingCredentials)
	ErrMissingAccountID                       = errors.New(errMissingAccountID)
	ErrMissingZoneID                          = errors.New(errMissingZoneID)
	ErrAccountIDOrZoneIDAreRequired           = errors.New(errMissingAccountOrZoneID)
	ErrAccountIDAndZoneIDAreMutuallyExclusive = errors.New(errAccountIDAndZoneIDAreMutuallyExclusive)
	ErrMissingResourceIdentifier              = errors.New(errMissingResourceIdentifier)

	ErrRequiredAccountLevelResourceContainer = errors.New(errRequiredAccountLevelResourceContainer)
	ErrRequiredZoneLevelResourceContainer    = errors.New(errRequiredZoneLevelResourceContainer)
)

type ErrorType string

const (
	ErrorTypeRequest        ErrorType = "request"
	ErrorTypeAuthentication ErrorType = "authentication"
	ErrorTypeAuthorization  ErrorType = "authorization"
	ErrorTypeNotFound       ErrorType = "not_found"
	ErrorTypeRateLimit      ErrorType = "rate_limit"
	ErrorTypeService        ErrorType = "service"
)

type Error struct {
	// The classification of error encountered.
	Type ErrorType

	// StatusCode is the HTTP status code from the response.
	StatusCode int

	// Errors is all of the error messages and codes, combined.
	Errors []ResponseInfo

	// ErrorCodes is a list of all the error codes.
	ErrorCodes []int

	// ErrorMessages is a list of all the error codes.
	ErrorMessages []string

	// Messages is a list of informational messages provided by the endpoint.
	Messages []ResponseInfo

	// RayID is the internal identifier for the request that was made.
	RayID string
}

func (e Error) Error() string {
	var errString string
	errMessages := []string{}
	for _, err := range e.Errors {
		m := ""
		if err.Message != "" {
			m += err.Message
		}

		if err.Code != 0 {
			m += fmt.Sprintf(" (%d)", err.Code)
		}

		errMessages = append(errMessages, m)
	}

	msgs := []string{}
	for _, m := range e.Messages {
		msgs = append(msgs, m.Message)
	}

	errString += strings.Join(errMessages, ", ")

	// `Messages` is primarily used for additional validation failure notes in
	// page rules. This shouldn't be used going forward but instead, use the
	// error fields appropriately.
	if len(msgs) > 0 {
		errString += "\n" + strings.Join(msgs, "  \n")
	}

	return errString
}

// RequestError is for 4xx errors that we encounter not covered elsewhere
// (generally bad payloads).
type RequestError struct {
	cloudflareError *Error
}

func (e RequestError) Error() string {
	return e.cloudflareError.Error()
}

func (e RequestError) Errors() []ResponseInfo {
	return e.cloudflareError.Errors
}

func (e RequestError) ErrorCodes() []int {
	return e.cloudflareError.ErrorCodes
}

func (e RequestError) ErrorMessages() []string {
	return e.cloudflareError.ErrorMessages
}

func (e RequestError) InternalErrorCodeIs(code int) bool {
	return e.cloudflareError.InternalErrorCodeIs(code)
}

func (e RequestError) Messages() []ResponseInfo {
	return e.cloudflareError.Messages
}

func (e RequestError) RayID() string {
	return e.cloudflareError.RayID
}

func (e RequestError) Type() ErrorType {
	return e.cloudflareError.Type
}

func NewRequestError(e *Error) RequestError {
	return RequestError{
		cloudflareError: e,
	}
}

// RatelimitError is for HTTP 429s where the service is telling the client to
// slow down.
type RatelimitError struct {
	cloudflareError *Error
}

func (e RatelimitError) Error() string {
	return e.cloudflareError.Error()
}

func (e RatelimitError) Errors() []ResponseInfo {
	return e.cloudflareError.Errors
}

func (e RatelimitError) ErrorCodes() []int {
	return e.cloudflareError.ErrorCodes
}

func (e RatelimitError) ErrorMessages() []string {
	return e.cloudflareError.ErrorMessages
}

func (e RatelimitError) InternalErrorCodeIs(code int) bool {
	return e.cloudflareError.InternalErrorCodeIs(code)
}

func (e RatelimitError) RayID() string {
	return e.cloudflareError.RayID
}

func (e RatelimitError) Type() ErrorType {
	return e.cloudflareError.Type
}

func NewRatelimitError(e *Error) RatelimitError {
	return RatelimitError{
		cloudflareError: e,
	}
}

// ServiceError is a handler for 5xx errors returned to the client.
type ServiceError struct {
	cloudflareError *Error
}

func (e ServiceError) Error() string {
	return e.cloudflareError.Error()
}

func (e ServiceError) Errors() []ResponseInfo {
	return e.cloudflareError.Errors
}

func (e ServiceError) ErrorCodes() []int {
	return e.cloudflareError.ErrorCodes
}

func (e ServiceError) ErrorMessages() []string {
	return e.cloudflareError.ErrorMessages
}

func (e ServiceError) InternalErrorCodeIs(code int) bool {
	return e.cloudflareError.InternalErrorCodeIs(code)
}

func (e ServiceError) RayID() string {
	return e.cloudflareError.RayID
}

func (e ServiceError) Type() ErrorType {
	return e.cloudflareError.Type
}

func NewServiceError(e *Error) ServiceError {
	return ServiceError{
		cloudflareError: e,
	}
}

// AuthenticationError is for HTTP 401 responses.
type AuthenticationError struct {
	cloudflareError *Error
}

func (e AuthenticationError) Error() string {
	return e.cloudflareError.Error()
}

func (e AuthenticationError) Errors() []ResponseInfo {
	return e.cloudflareError.Errors
}

func (e AuthenticationError) ErrorCodes() []int {
	return e.cloudflareError.ErrorCodes
}

func (e AuthenticationError) ErrorMessages() []string {
	return e.cloudflareError.ErrorMessages
}

func (e AuthenticationError) InternalErrorCodeIs(code int) bool {
	return e.cloudflareError.InternalErrorCodeIs(code)
}

func (e AuthenticationError) RayID() string {
	return e.cloudflareError.RayID
}

func (e AuthenticationError) Type() ErrorType {
	return e.cloudflareError.Type
}

func NewAuthenticationError(e *Error) AuthenticationError {
	return AuthenticationError{
		cloudflareError: e,
	}
}

// AuthorizationError is for HTTP 403 responses.
type AuthorizationError struct {
	cloudflareError *Error
}

func (e AuthorizationError) Error() string {
	return e.cloudflareError.Error()
}

func (e AuthorizationError) Errors() []ResponseInfo {
	return e.cloudflareError.Errors
}

func (e AuthorizationError) ErrorCodes() []int {
	return e.cloudflareError.ErrorCodes
}

func (e AuthorizationError) ErrorMessages() []string {
	return e.cloudflareError.ErrorMessages
}

func (e AuthorizationError) InternalErrorCodeIs(code int) bool {
	return e.cloudflareError.InternalErrorCodeIs(code)
}

func (e AuthorizationError) RayID() string {
	return e.cloudflareError.RayID
}

func (e AuthorizationError) Type() ErrorType {
	return e.cloudflareError.Type
}

func NewAuthorizationError(e *Error) AuthorizationError {
	return AuthorizationError{
		cloudflareError: e,
	}
}

// NotFoundError is for HTTP 404 responses.
type NotFoundError struct {
	cloudflareError *Error
}

func (e NotFoundError) Error() string {
	return e.cloudflareError.Error()
}

func (e NotFoundError) Errors() []ResponseInfo {
	return e.cloudflareError.Errors
}

func (e NotFoundError) ErrorCodes() []int {
	return e.cloudflareError.ErrorCodes
}

func (e NotFoundError) ErrorMessages() []string {
	return e.cloudflareError.ErrorMessages
}

func (e NotFoundError) InternalErrorCodeIs(code int) bool {
	return e.cloudflareError.InternalErrorCodeIs(code)
}

func (e NotFoundError) RayID() string {
	return e.cloudflareError.RayID
}

func (e NotFoundError) Type() ErrorType {
	return e.cloudflareError.Type
}

func NewNotFoundError(e *Error) NotFoundError {
	return NotFoundError{
		cloudflareError: e,
	}
}

// ClientError returns a boolean whether or not the raised error was caused by
// something client side.
func (e *Error) ClientError() bool {
	return e.StatusCode >= http.StatusBadRequest &&
		e.StatusCode < http.StatusInternalServerError
}

// ClientRateLimited returns a boolean whether or not the raised error was
// caused by too many requests from the client.
func (e *Error) ClientRateLimited() bool {
	return e.Type == ErrorTypeRateLimit
}

// InternalErrorCodeIs returns a boolean whether or not the desired internal
// error code is present in `e.InternalErrorCodes`.
func (e *Error) InternalErrorCodeIs(code int) bool {
	for _, errCode := range e.ErrorCodes {
		if errCode == code {
			return true
		}
	}

	return false
}

// ErrorMessageContains returns a boolean whether or not a substring exists in
// any of the `e.ErrorMessages` slice entries.
func (e *Error) ErrorMessageContains(s string) bool {
	for _, errMsg := range e.ErrorMessages {
		if strings.Contains(errMsg, s) {
			return true
		}
	}
	return false
}
