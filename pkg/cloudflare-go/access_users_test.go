package cloudflare

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testAccessUserID        = "access-user-id"
	testAccessUserSessionID = "access-user-session-id"

	expectedListAccessUserResult = []AccessUser{
		{
			AccessSeat:          BoolPtr(false),
			ActiveDeviceCount:   2,
			CreatedAt:           "2014-01-01T05:20:00.12345Z",
			Email:               "jdoe@example.com",
			GatewaySeat:         BoolPtr(false),
			ID:                  "f3b12456-80dd-4e89-9f5f-ba3dfff12365",
			LastSuccessfulLogin: "2020-07-01T05:20:00Z",
			Name:                "Jane Doe",
			SeatUID:             "",
			UID:                 "",
			UpdatedAt:           "2014-01-01T05:20:00.12345Z",
		},
		{
			AccessSeat:          BoolPtr(true),
			ActiveDeviceCount:   2,
			CreatedAt:           "2024-01-01T05:20:00.12345Z",
			Email:               "jhondoe@example.com",
			GatewaySeat:         BoolPtr(true),
			ID:                  "c3b12456-80dd-4e89-9f5f-ba3dfff12367",
			LastSuccessfulLogin: "2020-07-01T05:20:00Z",
			Name:                "Jhon Doe",
			SeatUID:             "",
			UID:                 "",
			UpdatedAt:           "2014-01-01T05:20:00.12345Z",
		},
	}

	expectedGetAccessUserActiveSessionsResult = AccessUserActiveSessionResult{
		Expiration: 1694813506,
		Metadata: AccessUserActiveSessionMetadata{
			Apps: map[string]AccessUserActiveSessionMetadataApp{
				"property1": {
					Hostname: "test.example.com",
					Name:     "app name",
					Type:     "self_hosted",
					UID:      "cc2a8145-0128-4429-87f3-872c4d380c4e",
				},
				"property2": {
					Hostname: "test.example.com",
					Name:     "app name",
					Type:     "self_hosted",
					UID:      "cc2a8145-0128-4429-87f3-872c4d380c4e",
				},
			},
			Expires: 1694813506,
			IAT:     1694791905,
			Nonce:   "X1aXj1lFVcqqyoXF",
			TTL:     21600,
		},
		Name: "string",
	}

	expectedGetAccessUserFailedLoginsResult = AccessUserFailedLoginResult{
		Expiration: 0,
		Metadata: AccessUserFailedLoginMetadata{
			AppName:   "Test App",
			Aud:       "39691c1480a2352a18ece567debc2b32552686cbd38eec0887aa18d5d3f00c04",
			Datetime:  "2022-02-02T21:54:34.914Z",
			RayID:     "6d76a8a42ead4133",
			UserEmail: "test@cloudflare.com",
			UserUUID:  "57171132-e453-4ee8-b2a5-8cbaad333207",
		},
	}

	expectedGetAccessUserLastSeenIdentityResult = GetAccessUserLastSeenIdentityResult{
		AccountID:  "1234567890",
		AuthStatus: "NONE",
		CommonName: "",
		DevicePosture: map[string]AccessUserDevicePosture{
			"property1": {
				Check: AccessUserDevicePostureCheck{
					Exists: BoolPtr(true),
					Path:   "string",
				},
				Data:        map[string]interface{}{},
				Description: "string",
				Error:       "string",
				ID:          "string",
				RuleName:    "string",
				Success:     BoolPtr(true),
				Timestamp:   "string",
				Type:        "string",
			},
			"property2": {
				Check: AccessUserDevicePostureCheck{
					Exists: BoolPtr(true),
					Path:   "string",
				},
				Data:        map[string]interface{}{},
				Description: "string",
				Error:       "string",
				ID:          "string",
				RuleName:    "string",
				Success:     BoolPtr(true),
				Timestamp:   "string",
				Type:        "string",
			},
		},
		DeviceID: "",
		DeviceSessions: map[string]AccessUserDeviceSession{
			"property1": {
				LastAuthenticated: 1638832687,
			},
			"property2": {
				LastAuthenticated: 1638832687,
			},
		},
		Email: "test@cloudflare.com",
		Geo: AccessUserIdentityGeo{
			Country: "US",
		},
		IAT: 1694791905,
		IDP: AccessUserIDP{
			ID:   "string",
			Type: "string",
		},
		IP:        "127.0.0.0",
		IsGateway: BoolPtr(false),
		IsWarp:    BoolPtr(false),
		MtlsAuth: AccessUserMTLSAuth{
			AuthStatus:    "string",
			CertIssuerDN:  "string",
			CertIssuerSKI: "string",
			CertPresented: BoolPtr(true),
			CertSerial:    "string",
		},
		ServiceTokenID:     "",
		ServiceTokenStatus: BoolPtr(false),
		UserUUID:           "57cf8cf2-f55a-4588-9ac9-f5e41e9f09b4",
		Version:            2,
	}

	expectedGetAccessUserSingleActiveSessionResult = GetAccessUserSingleActiveSessionResult{
		AccountID:  "1234567890",
		AuthStatus: "NONE",
		CommonName: "",
		DevicePosture: map[string]AccessUserDevicePosture{
			"property1": {
				Check: AccessUserDevicePostureCheck{
					Exists: BoolPtr(true),
					Path:   "string",
				},
				Data:        map[string]interface{}{},
				Description: "string",
				Error:       "string",
				ID:          "string",
				RuleName:    "string",
				Success:     BoolPtr(true),
				Timestamp:   "string",
				Type:        "string",
			},
			"property2": {
				Check: AccessUserDevicePostureCheck{
					Exists: BoolPtr(true),
					Path:   "string",
				},
				Data:        map[string]interface{}{},
				Description: "string",
				Error:       "string",
				ID:          "string",
				RuleName:    "string",
				Success:     BoolPtr(true),
				Timestamp:   "string",
				Type:        "string",
			},
		},
		DeviceID: "",
		DeviceSessions: map[string]AccessUserDeviceSession{
			"property1": {
				LastAuthenticated: 1638832687,
			},
			"property2": {
				LastAuthenticated: 1638832687,
			},
		},
		Email: "test@cloudflare.com",
		Geo: AccessUserIdentityGeo{
			Country: "US",
		},
		IAT: 1694791905,
		IDP: AccessUserIDP{
			ID:   "string",
			Type: "string",
		},
		IP:        "127.0.0.0",
		IsGateway: BoolPtr(false),
		IsWarp:    BoolPtr(false),
		MtlsAuth: AccessUserMTLSAuth{
			AuthStatus:    "string",
			CertIssuerDN:  "string",
			CertIssuerSKI: "string",
			CertPresented: BoolPtr(true),
			CertSerial:    "string",
		},
		ServiceTokenID:     "",
		ServiceTokenStatus: BoolPtr(false),
		UserUUID:           "57cf8cf2-f55a-4588-9ac9-f5e41e9f09b4",
		Version:            2,
		IsActive:           BoolPtr(true),
	}
)

func TestListAccessUsers_ZoneIsNotSupported(t *testing.T) {
	setup()
	defer teardown()

	_, _, err := client.ListAccessUsers(context.Background(), testZoneRC, AccessUserParams{})
	assert.EqualError(t, err, fmt.Sprintf(errInvalidResourceContainerAccess, ZoneRouteLevel))
}

func TestListAccessUsers(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		userList, err := json.Marshal(expectedListAccessUserResult)
		assert.NoError(t, err, "Error marshaling expectedListAccessUserResult")

		fmt.Fprintf(w, `{
			"errors": [],
			"messages": [],
			"result": %s,
			"success": true,
			"result_info": {
			  "count": 2,
			  "page": 1,
			  "per_page": 100,
			  "total_count": 2
			}
		  }`, string(userList))
	}
	mux.HandleFunc("/accounts/"+testAccountID+"/access/users", handler)

	actual, _, err := client.ListAccessUsers(context.Background(), testAccountRC, AccessUserParams{})

	if assert.NoError(t, err) {
		assert.Equal(t, expectedListAccessUserResult, actual)
	}
}

func TestListAccessUsersWithPagination(t *testing.T) {
	setup()
	defer teardown()
	// page 1 of 2
	page := 1

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")

		userList, err := json.Marshal(expectedListAccessUserResult)
		assert.NoError(t, err, "Error marshaling expectedListAccessUserResult")

		fmt.Fprintf(w, `{
			"errors": [],
			"messages": [],
			"result": %s,
			"success": true,
			"result_info": {
			  "count": 2,
			  "page": %d,
			  "per_page": 2,
			  "total_count": 4
			}
		  }`, string(userList), page)
		// increment page for the next call
		page++
	}
	mux.HandleFunc("/accounts/"+testAccountID+"/access/users", handler)

	actual, _, err := client.ListAccessUsers(context.Background(), testAccountRC, AccessUserParams{})
	expected := []AccessUser{}
	// two pages of the same expectedResult
	expected = append(expected, expectedListAccessUserResult...)
	expected = append(expected, expectedListAccessUserResult...)
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func TestGetGetAccessUserActiveSessions_ZoneIsNotSupported(t *testing.T) {
	setup()
	defer teardown()

	_, err := client.GetAccessUserActiveSessions(context.Background(), testZoneRC, testAccessUserID)
	assert.EqualError(t, err, fmt.Sprintf(errInvalidResourceContainerAccess, ZoneRouteLevel))
}

func TestGetGetAccessUserActiveSessions(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"errors": [],
			"messages": [],
			"result": [
			  {
				"expiration": 1694813506,
				"metadata": {
				  "apps": {
					"property1": {
					  "hostname": "test.example.com",
					  "name": "app name",
					  "type": "self_hosted",
					  "uid": "cc2a8145-0128-4429-87f3-872c4d380c4e"
					},
					"property2": {
					  "hostname": "test.example.com",
					  "name": "app name",
					  "type": "self_hosted",
					  "uid": "cc2a8145-0128-4429-87f3-872c4d380c4e"
					}
				  },
				  "expires": 1694813506,
				  "iat": 1694791905,
				  "nonce": "X1aXj1lFVcqqyoXF",
				  "ttl": 21600
				},
				"name": "string"
			  }
			],
			"success": true,
			"result_info": {
			  "count": 1,
			  "page": 1,
			  "per_page": 20,
			  "total_count": 2000
			}
		  }
		`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/users/"+testAccessUserID+"/active_sessions", handler)

	actual, err := client.GetAccessUserActiveSessions(context.Background(), testAccountRC, testAccessUserID)
	if err != nil {
		t.Fatal(err)
	}

	if assert.NoError(t, err) {
		assert.Equal(t, []AccessUserActiveSessionResult{expectedGetAccessUserActiveSessionsResult}, actual)
	}
}

func TestGetAccessUserSingleActiveSession_ZoneIsNotSupported(t *testing.T) {
	setup()
	defer teardown()

	_, err := client.GetAccessUserSingleActiveSession(context.Background(), testZoneRC, testAccessUserID, testAccessUserSessionID)
	assert.EqualError(t, err, fmt.Sprintf(errInvalidResourceContainerAccess, ZoneRouteLevel))
}

func TestGetAccessUserSingleActiveSession(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"errors": [],
			"messages": [],
			"result": {
			  "account_id": "1234567890",
			  "auth_status": "NONE",
			  "common_name": "",
			  "devicePosture": {
				"property1": {
				  "check": {
					"exists": true,
					"path": "string"
				  },
				  "data": {},
				  "description": "string",
				  "error": "string",
				  "id": "string",
				  "rule_name": "string",
				  "success": true,
				  "timestamp": "string",
				  "type": "string"
				},
				"property2": {
				  "check": {
					"exists": true,
					"path": "string"
				  },
				  "data": {},
				  "description": "string",
				  "error": "string",
				  "id": "string",
				  "rule_name": "string",
				  "success": true,
				  "timestamp": "string",
				  "type": "string"
				}
			  },
			  "device_id": "",
			  "device_sessions": {
				"property1": {
				  "last_authenticated": 1638832687
				},
				"property2": {
				  "last_authenticated": 1638832687
				}
			  },
			  "email": "test@cloudflare.com",
			  "geo": {
				"country": "US"
			  },
			  "iat": 1694791905,
			  "idp": {
				"id": "string",
				"type": "string"
			  },
			  "ip": "127.0.0.0",
			  "is_gateway": false,
			  "is_warp": false,
			  "mtls_auth": {
				"auth_status": "string",
				"cert_issuer_dn": "string",
				"cert_issuer_ski": "string",
				"cert_presented": true,
				"cert_serial": "string"
			  },
			  "service_token_id": "",
			  "service_token_status": false,
			  "user_uuid": "57cf8cf2-f55a-4588-9ac9-f5e41e9f09b4",
			  "version": 2,
			  "isActive": true
			},
			"success": true
		  }
		`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/users/"+testAccessUserID+"/active_sessions/"+testAccessUserSessionID, handler)

	actual, err := client.GetAccessUserSingleActiveSession(context.Background(), testAccountRC, testAccessUserID, testAccessUserSessionID)

	if assert.NoError(t, err) {
		assert.Equal(t, expectedGetAccessUserSingleActiveSessionResult, actual)
	}
}

func TestGetAccessUserFailedLogins_ZoneIsNotSupported(t *testing.T) {
	setup()
	defer teardown()

	_, err := client.GetAccessUserFailedLogins(context.Background(), testZoneRC, testAccessUserID)
	assert.EqualError(t, err, fmt.Sprintf(errInvalidResourceContainerAccess, ZoneRouteLevel))
}

func TestGetAccessUserFailedLogins(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"errors": [],
			"messages": [],
			"result": [
			  {
				"expiration": 0,
				"metadata": {
				  "app_name": "Test App",
				  "aud": "39691c1480a2352a18ece567debc2b32552686cbd38eec0887aa18d5d3f00c04",
				  "datetime": "2022-02-02T21:54:34.914Z",
				  "ray_id": "6d76a8a42ead4133",
				  "user_email": "test@cloudflare.com",
				  "user_uuid": "57171132-e453-4ee8-b2a5-8cbaad333207"
				}
			  }
			],
			"success": true,
			"result_info": {
			  "count": 1,
			  "page": 1,
			  "per_page": 20,
			  "total_count": 2000
			}
		  }
		`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/users/"+testAccessUserID+"/failed_logins", handler)

	actual, err := client.GetAccessUserFailedLogins(context.Background(), testAccountRC, testAccessUserID)

	if assert.NoError(t, err) {
		assert.Equal(t, []AccessUserFailedLoginResult{expectedGetAccessUserFailedLoginsResult}, actual)
	}
}

func TestGetAccessUserLastSeenIdentity_ZoneIsNotSupported(t *testing.T) {
	setup()
	defer teardown()

	_, err := client.GetAccessUserLastSeenIdentity(context.Background(), testZoneRC, testAccessUserID)
	assert.EqualError(t, err, fmt.Sprintf(errInvalidResourceContainerAccess, ZoneRouteLevel))
}

func TestGetAccessUserLastSeenIdentity(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
  "errors": [],
  "messages": [],
  "result": {
    "account_id": "1234567890",
    "auth_status": "NONE",
    "common_name": "",
    "devicePosture": {
      "property1": {
        "check": {
          "exists": true,
          "path": "string"
        },
        "data": {},
        "description": "string",
        "error": "string",
        "id": "string",
        "rule_name": "string",
        "success": true,
        "timestamp": "string",
        "type": "string"
      },
      "property2": {
        "check": {
          "exists": true,
          "path": "string"
        },
        "data": {},
        "description": "string",
        "error": "string",
        "id": "string",
        "rule_name": "string",
        "success": true,
        "timestamp": "string",
        "type": "string"
      }
    },
    "device_id": "",
    "device_sessions": {
      "property1": {
        "last_authenticated": 1638832687
      },
      "property2": {
        "last_authenticated": 1638832687
      }
    },
    "email": "test@cloudflare.com",
    "geo": {
      "country": "US"
    },
    "iat": 1694791905,
    "idp": {
      "id": "string",
      "type": "string"
    },
    "ip": "127.0.0.0",
    "is_gateway": false,
    "is_warp": false,
    "mtls_auth": {
      "auth_status": "string",
      "cert_issuer_dn": "string",
      "cert_issuer_ski": "string",
      "cert_presented": true,
      "cert_serial": "string"
    },
    "service_token_id": "",
    "service_token_status": false,
    "user_uuid": "57cf8cf2-f55a-4588-9ac9-f5e41e9f09b4",
    "version": 2
  },
  "success": true
}
		`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/access/users/"+testAccessUserID+"/last_seen_identity", handler)

	actual, err := client.GetAccessUserLastSeenIdentity(context.Background(), testAccountRC, testAccessUserID)

	if assert.NoError(t, err) {
		assert.Equal(t, expectedGetAccessUserLastSeenIdentityResult, actual)
	}
}
