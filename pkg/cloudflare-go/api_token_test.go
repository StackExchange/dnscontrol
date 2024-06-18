package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAPITokens(t *testing.T) {
	setup()
	defer teardown()

	issuedOn, _ := time.Parse(time.RFC3339, "2018-07-01T05:20:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2018-07-02T05:20:00Z")
	notBefore, _ := time.Parse(time.RFC3339, "2018-07-01T05:20:00Z")
	expiresOn, _ := time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
      "success": true,
      "errors": [],
      "messages": [],
      "result": [
        {
          "id": "ed17574386854bf78a67040be0a770b0",
          "name": "readonly token",
          "status": "active",
          "issued_on": "2018-07-01T05:20:00Z",
          "modified_on": "2018-07-02T05:20:00Z",
          "not_before": "2018-07-01T05:20:00Z",
          "expires_on": "2020-01-01T00:00:00Z",
          "policies": [
            {
              "id": "f267e341f3dd4697bd3b9f71dd96247f",
              "effect": "allow",
              "resources": {
                "com.cloudflare.api.account.zone.eb78d65290b24279ba6f44721b3ea3c4": "*",
                "com.cloudflare.api.account.zone.22b1de5f1c0e4b3ea97bb1e963b06a43": "*"
              },
              "permission_groups": [
                {
                  "id": "c8fed203ed3043cba015a93ad1616f1f",
                  "name": "Zone Read"
                },
                {
                  "id": "82e64a83756745bbbb1c9c2701bf816b",
                  "name": "DNS Read"
                }
              ]
            }
          ],
          "condition": {
            "request.ip": {
              "in": [
                "192.0.2.0/24",
                "2001:db8::/48"
              ],
              "not_in": [
                "198.51.100.0/24",
                "2001:db8:1::/48"
              ]
            }
          }
        }
      ]
    }`)
	}

	mux.HandleFunc("/user/tokens", handler)

	resources := make(map[string]interface{})
	resources["com.cloudflare.api.account.zone.eb78d65290b24279ba6f44721b3ea3c4"] = "*"
	resources["com.cloudflare.api.account.zone.22b1de5f1c0e4b3ea97bb1e963b06a43"] = "*"

	expectedAPIToken := APIToken{
		ID:         "ed17574386854bf78a67040be0a770b0",
		Name:       "readonly token",
		Status:     "active",
		IssuedOn:   &issuedOn,
		ModifiedOn: &modifiedOn,
		NotBefore:  &notBefore,
		ExpiresOn:  &expiresOn,
		Policies: []APITokenPolicies{{
			ID:        "f267e341f3dd4697bd3b9f71dd96247f",
			Effect:    "allow",
			Resources: resources,
			PermissionGroups: []APITokenPermissionGroups{
				{
					ID:   "c8fed203ed3043cba015a93ad1616f1f",
					Name: "Zone Read",
				},
				{
					ID:   "82e64a83756745bbbb1c9c2701bf816b",
					Name: "DNS Read",
				},
			},
		}},
		Condition: &APITokenCondition{
			RequestIP: &APITokenRequestIPCondition{
				In:    []string{"192.0.2.0/24", "2001:db8::/48"},
				NotIn: []string{"198.51.100.0/24", "2001:db8:1::/48"},
			},
		},
	}

	actual, err := client.APITokens(context.Background())

	if assert.NoError(t, err) {
		assert.Equal(t, []APIToken{expectedAPIToken}, actual)
	}
}

func TestGetAPIToken(t *testing.T) {
	setup()
	defer teardown()

	issuedOn, _ := time.Parse(time.RFC3339, "2018-07-01T05:20:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2018-07-02T05:20:00Z")
	notBefore, _ := time.Parse(time.RFC3339, "2018-07-01T05:20:00Z")
	expiresOn, _ := time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
      "success": true,
      "errors": [],
      "messages": [],
      "result": {
          "id": "ed17574386854bf78a67040be0a770b0",
          "name": "readonly token",
          "status": "active",
          "issued_on": "2018-07-01T05:20:00Z",
          "modified_on": "2018-07-02T05:20:00Z",
          "not_before": "2018-07-01T05:20:00Z",
          "expires_on": "2020-01-01T00:00:00Z",
          "policies": [
            {
              "id": "f267e341f3dd4697bd3b9f71dd96247f",
              "effect": "allow",
              "resources": {
                "com.cloudflare.api.account.zone.eb78d65290b24279ba6f44721b3ea3c4": "*",
                "com.cloudflare.api.account.zone.22b1de5f1c0e4b3ea97bb1e963b06a43": "*"
              },
              "permission_groups": [
                {
                  "id": "c8fed203ed3043cba015a93ad1616f1f",
                  "name": "Zone Read"
                },
                {
                  "id": "82e64a83756745bbbb1c9c2701bf816b",
                  "name": "DNS Read"
                }
              ]
            }
          ],
          "condition": {
            "request.ip": {
              "in": [
                "192.0.2.0/24",
                "2001:db8::/48"
              ],
              "not_in": [
                "198.51.100.0/24",
                "2001:db8:1::/48"
              ]
            }
          }
        }
      }`)
	}

	mux.HandleFunc("/user/tokens/ed17574386854bf78a67040be0a770b0", handler)

	resources := make(map[string]interface{})
	resources["com.cloudflare.api.account.zone.eb78d65290b24279ba6f44721b3ea3c4"] = "*"
	resources["com.cloudflare.api.account.zone.22b1de5f1c0e4b3ea97bb1e963b06a43"] = "*"

	expectedAPIToken := APIToken{
		ID:         "ed17574386854bf78a67040be0a770b0",
		Name:       "readonly token",
		Status:     "active",
		IssuedOn:   &issuedOn,
		ModifiedOn: &modifiedOn,
		NotBefore:  &notBefore,
		ExpiresOn:  &expiresOn,
		Policies: []APITokenPolicies{{
			ID:        "f267e341f3dd4697bd3b9f71dd96247f",
			Effect:    "allow",
			Resources: resources,
			PermissionGroups: []APITokenPermissionGroups{
				{
					ID:   "c8fed203ed3043cba015a93ad1616f1f",
					Name: "Zone Read",
				},
				{
					ID:   "82e64a83756745bbbb1c9c2701bf816b",
					Name: "DNS Read",
				},
			},
		}},
		Condition: &APITokenCondition{
			RequestIP: &APITokenRequestIPCondition{
				In:    []string{"192.0.2.0/24", "2001:db8::/48"},
				NotIn: []string{"198.51.100.0/24", "2001:db8:1::/48"},
			},
		},
	}

	actual, err := client.GetAPIToken(context.Background(), "ed17574386854bf78a67040be0a770b0")

	if assert.NoError(t, err) {
		assert.Equal(t, expectedAPIToken, actual)
	}
}

func TestCreateAPIToken(t *testing.T) {
	setup()
	defer teardown()

	// issuedOn, _ := time.Parse(time.RFC3339, "2018-07-01T05:20:00Z")
	// modifiedOn, _ := time.Parse(time.RFC3339, "2018-07-02T05:20:00Z")
	notBefore, _ := time.Parse(time.RFC3339, "2018-07-01T05:20:00Z")
	expiresOn, _ := time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
    "success":true,
    "errors":[],
    "messages":[],
    "result":{
      "id":"ed17574386854bf78a67040be0a770b0",
      "name":"readonly token",
      "status":"active",
      "issued_on":"2018-07-01T05:20:00Z",
      "modified_on":"2018-07-02T05:20:00Z",
      "not_before":"2018-07-01T05:20:00Z",
      "expires_on":"2020-01-01T00:00:00Z",
      "policies":[
        {
          "id":"f267e341f3dd4697bd3b9f71dd96247f",
          "effect":"allow",
          "resources":{
            "com.cloudflare.api.account.zone.eb78d65290b24279ba6f44721b3ea3c4":"*",
            "com.cloudflare.api.account.zone.22b1de5f1c0e4b3ea97bb1e963b06a43":"*"
          },
          "permission_groups":[
            {
              "id":"c8fed203ed3043cba015a93ad1616f1f",
              "name":"Zone Read"
            },
            {
              "id":"82e64a83756745bbbb1c9c2701bf816b",
              "name":"DNS Read"
            }
          ]
        }
      ]
    }
  }`)
	}

	mux.HandleFunc("/user/tokens", handler)

	resources := make(map[string]interface{})
	resources["com.cloudflare.api.account.zone.eb78d65290b24279ba6f44721b3ea3c4"] = "*"
	resources["com.cloudflare.api.account.zone.22b1de5f1c0e4b3ea97bb1e963b06a43"] = "*"

	tokenToCreate := APIToken{
		Name:      "readonly token",
		NotBefore: &notBefore,
		ExpiresOn: &expiresOn,
		Policies: []APITokenPolicies{{
			ID:        "f267e341f3dd4697bd3b9f71dd96247f",
			Effect:    "allow",
			Resources: resources,
			PermissionGroups: []APITokenPermissionGroups{
				{
					ID:   "c8fed203ed3043cba015a93ad1616f1f",
					Name: "Zone Read",
				},
				{
					ID:   "82e64a83756745bbbb1c9c2701bf816b",
					Name: "DNS Read",
				},
			},
		}},
	}

	actual, err := client.CreateAPIToken(context.Background(), tokenToCreate)

	if assert.NoError(t, err) {
		assert.Equal(t, tokenToCreate.Name, actual.Name)
		assert.Equal(t, tokenToCreate.Policies, actual.Policies)
	}
}

func TestUpdateAPIToken(t *testing.T) {
	setup()
	defer teardown()

	issuedOn, _ := time.Parse(time.RFC3339, "2018-07-01T05:20:00Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2018-07-02T05:20:00Z")
	notBefore, _ := time.Parse(time.RFC3339, "2018-07-01T05:20:00Z")
	expiresOn, _ := time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
  "success": true,
  "errors": [],
  "messages": [],
  "result": {
    "id": "ed17574386854bf78a67040be0a770b0",
    "name": "readonly token",
    "status": "active",
    "issued_on": "2018-07-01T05:20:00Z",
    "modified_on": "2018-07-02T05:20:00Z",
    "not_before": "2018-07-01T05:20:00Z",
    "expires_on": "2020-01-01T00:00:00Z",
    "policies": [
      {
        "id": "f267e341f3dd4697bd3b9f71dd96247f",
        "effect": "allow",
        "resources": {
          "com.cloudflare.api.account.zone.eb78d65290b24279ba6f44721b3ea3c4": "*",
          "com.cloudflare.api.account.zone.22b1de5f1c0e4b3ea97bb1e963b06a43": "*"
        },
        "permission_groups": [
          {
            "id": "c8fed203ed3043cba015a93ad1616f1f",
            "name": "Zone Read"
          },
          {
            "id": "82e64a83756745bbbb1c9c2701bf816b",
            "name": "DNS Read"
          }
        ]
      }
    ],
    "condition": {
      "request.ip": {
        "in": [
          "198.51.100.0/24",
          "2400:cb00::/32"
        ],
        "not_in": [
          "198.51.100.0/24",
          "2400:cb00::/32"
        ]
      }
    }
  }
}`)
	}

	mux.HandleFunc("/user/tokens/ed17574386854bf78a67040be0a770b0", handler)

	resources := make(map[string]interface{})
	resources["com.cloudflare.api.account.zone.eb78d65290b24279ba6f44721b3ea3c4"] = "*"
	resources["com.cloudflare.api.account.zone.22b1de5f1c0e4b3ea97bb1e963b06a43"] = "*"

	expectedAPIToken := APIToken{
		ID:         "ed17574386854bf78a67040be0a770b0",
		Name:       "readonly token",
		Status:     "active",
		IssuedOn:   &issuedOn,
		ModifiedOn: &modifiedOn,
		NotBefore:  &notBefore,
		ExpiresOn:  &expiresOn,
		Policies: []APITokenPolicies{{
			ID:        "f267e341f3dd4697bd3b9f71dd96247f",
			Effect:    "allow",
			Resources: resources,
			PermissionGroups: []APITokenPermissionGroups{
				{
					ID:   "c8fed203ed3043cba015a93ad1616f1f",
					Name: "Zone Read",
				},
				{
					ID:   "82e64a83756745bbbb1c9c2701bf816b",
					Name: "DNS Read",
				},
			},
		},
		},
		Condition: &APITokenCondition{
			RequestIP: &APITokenRequestIPCondition{
				In:    []string{"198.51.100.0/24", "2400:cb00::/32"},
				NotIn: []string{"198.51.100.0/24", "2400:cb00::/32"},
			},
		},
	}

	actual, err := client.UpdateAPIToken(context.Background(), "ed17574386854bf78a67040be0a770b0", APIToken{})

	if assert.NoError(t, err) {
		assert.Equal(t, expectedAPIToken, actual)
	}
}

func TestRollAPIToken(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
      "success": true,
      "errors": [],
      "messages": [],
      "result": "8M7wS6hCpXVc-DoRnPPY_UCWPgy8aea4Wy6kCe5T"
    }
    `)
	}

	mux.HandleFunc("/user/tokens/ed17574386854bf78a67040be0a770b0/value", handler)

	actual, err := client.RollAPIToken(context.Background(), "ed17574386854bf78a67040be0a770b0")

	if assert.NoError(t, err) {
		assert.Equal(t, "8M7wS6hCpXVc-DoRnPPY_UCWPgy8aea4Wy6kCe5T", actual)
	}
}

func TestVerifyAPIToken(t *testing.T) {
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
        "id": "ed17574386854bf78a67040be0a770b0",
        "status": "active",
        "not_before": "2018-07-01T05:20:00Z",
        "expires_on": "2020-01-01T00:00:00Z"
      }
    }`)
	}

	mux.HandleFunc("/user/tokens/verify", handler)

	actual, err := client.VerifyAPIToken(context.Background())

	if assert.NoError(t, err) {
		assert.Equal(t, "active", actual.Status)
	}
}

func TestDeleteAPIToken(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
      "success": true,
      "errors": [],
      "messages": [],
      "result": {
        "id": "9a7806061c88ada191ed06f989cc3dac"
      }
    }`)
	}

	mux.HandleFunc("/user/tokens/ed17574386854bf78a67040be0a770b0", handler)

	err := client.DeleteAPIToken(context.Background(), "ed17574386854bf78a67040be0a770b0")
	assert.NoError(t, err)
}

func TestListAPITokensPermissionGroups(t *testing.T) {
	setup()
	defer teardown()

	var pgID = "47aa30b6eb97ecae0518b750d6b142b6"

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
		  "success": true,
		  "errors": [],
		  "messages": [],
		  "result": [
			{
				"id": %q,
				"name": "DNS Read",
				"scopes": ["com.cloudflare.api.account.zone"]
			}
		  ]
		}
		`, pgID)
	}

	mux.HandleFunc("/user/tokens/permission_groups", handler)
	want := []APITokenPermissionGroups{{
		ID:     pgID,
		Name:   "DNS Read",
		Scopes: []string{"com.cloudflare.api.account.zone"},
	}}
	actual, err := client.ListAPITokensPermissionGroups(context.Background())
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}
