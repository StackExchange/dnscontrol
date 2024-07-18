package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/goccy/go-json"
)

// User describes a user account.
type User struct {
	ID         string     `json:"id,omitempty"`
	Email      string     `json:"email,omitempty"`
	FirstName  string     `json:"first_name,omitempty"`
	LastName   string     `json:"last_name,omitempty"`
	Username   string     `json:"username,omitempty"`
	Telephone  string     `json:"telephone,omitempty"`
	Country    string     `json:"country,omitempty"`
	Zipcode    string     `json:"zipcode,omitempty"`
	CreatedOn  *time.Time `json:"created_on,omitempty"`
	ModifiedOn *time.Time `json:"modified_on,omitempty"`
	APIKey     string     `json:"api_key,omitempty"`
	TwoFA      bool       `json:"two_factor_authentication_enabled,omitempty"`
	Betas      []string   `json:"betas,omitempty"`
	Accounts   []Account  `json:"organizations,omitempty"`
}

// UserResponse wraps a response containing User accounts.
type UserResponse struct {
	Response
	Result User `json:"result"`
}

// userBillingProfileResponse wraps a response containing Billing Profile information.
type userBillingProfileResponse struct {
	Response
	Result UserBillingProfile
}

// UserBillingProfile contains Billing Profile information.
type UserBillingProfile struct {
	ID              string     `json:"id,omitempty"`
	FirstName       string     `json:"first_name,omitempty"`
	LastName        string     `json:"last_name,omitempty"`
	Address         string     `json:"address,omitempty"`
	Address2        string     `json:"address2,omitempty"`
	Company         string     `json:"company,omitempty"`
	City            string     `json:"city,omitempty"`
	State           string     `json:"state,omitempty"`
	ZipCode         string     `json:"zipcode,omitempty"`
	Country         string     `json:"country,omitempty"`
	Telephone       string     `json:"telephone,omitempty"`
	CardNumber      string     `json:"card_number,omitempty"`
	CardExpiryYear  int        `json:"card_expiry_year,omitempty"`
	CardExpiryMonth int        `json:"card_expiry_month,omitempty"`
	VAT             string     `json:"vat,omitempty"`
	CreatedOn       *time.Time `json:"created_on,omitempty"`
	EditedOn        *time.Time `json:"edited_on,omitempty"`
}

type UserBillingHistoryResponse struct {
	Response
	Result     []UserBillingHistory `json:"result"`
	ResultInfo ResultInfo           `json:"result_info"`
}

type UserBillingHistory struct {
	ID          string                 `json:"id,omitempty"`
	Type        string                 `json:"type,omitempty"`
	Action      string                 `json:"action,omitempty"`
	Description string                 `json:"description,omitempty"`
	OccurredAt  *time.Time             `json:"occurred_at,omitempty"`
	Amount      float32                `json:"amount,omitempty"`
	Currency    string                 `json:"currency,omitempty"`
	Zone        userBillingHistoryZone `json:"zone"`
}

type userBillingHistoryZone struct {
	Name string `json:"name,omitempty"`
}

type UserBillingOptions struct {
	PaginationOptions
	Order      string     `url:"order,omitempty"`
	Type       string     `url:"type,omitempty"`
	OccurredAt *time.Time `url:"occurred_at,omitempty"`
	Action     string     `url:"action,omitempty"`
}

// UserDetails provides information about the logged-in user.
//
// API reference: https://api.cloudflare.com/#user-user-details
func (api *API) UserDetails(ctx context.Context) (User, error) {
	var r UserResponse
	res, err := api.makeRequestContext(ctx, http.MethodGet, "/user", nil)
	if err != nil {
		return User{}, err
	}

	err = json.Unmarshal(res, &r)
	if err != nil {
		return User{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return r.Result, nil
}

// UpdateUser updates the properties of the given user.
//
// API reference: https://api.cloudflare.com/#user-update-user
func (api *API) UpdateUser(ctx context.Context, user *User) (User, error) {
	var r UserResponse
	res, err := api.makeRequestContext(ctx, http.MethodPatch, "/user", user)
	if err != nil {
		return User{}, err
	}

	err = json.Unmarshal(res, &r)
	if err != nil {
		return User{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return r.Result, nil
}

// UserBillingProfile returns the billing profile of the user.
//
// API reference: https://api.cloudflare.com/#user-billing-profile
func (api *API) UserBillingProfile(ctx context.Context) (UserBillingProfile, error) {
	var r userBillingProfileResponse
	res, err := api.makeRequestContext(ctx, http.MethodGet, "/user/billing/profile", nil)
	if err != nil {
		return UserBillingProfile{}, err
	}

	err = json.Unmarshal(res, &r)
	if err != nil {
		return UserBillingProfile{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return r.Result, nil
}

// UserBillingHistory return the billing history of the user
//
// API reference: https://api.cloudflare.com/#user-billing-history-billing-history-details
func (api *API) UserBillingHistory(ctx context.Context, pageOpts UserBillingOptions) ([]UserBillingHistory, error) {
	uri := buildURI("/user/billing/history", pageOpts)
	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return []UserBillingHistory{}, err
	}
	var r UserBillingHistoryResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return []UserBillingHistory{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}
	return r.Result, nil
}
