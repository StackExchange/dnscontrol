package cloudflare

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateLoadBalancerPool(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		b, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if assert.NoError(t, err) {
			assert.JSONEq(t, `{
              "description": "Primary data center - Provider XYZ",
              "name": "primary-dc-1",
              "enabled": true,
              "monitor": "f1aba936b94213e5b8dca0c0dbf1f9cc",
              "latitude": 55,
              "longitude": -12.5,
              "load_shedding": {
                "default_percent": 50,
                "default_policy": "random",
                "session_percent": 10,
                "session_policy": "hash"
              },
              "origin_steering": {
                "policy": "random"
              },
              "origins": [
                {
                  "name": "app-server-1",
                  "address": "198.51.100.1",
                  "enabled": true,
                  "weight": 1,
                  "header": {
                      "Host": [
                          "example.com"
                      ]
                  },
			      "virtual_network_id":"a5624d4e-044a-4ff0-b3e1-e2465353d4b4"
                }
              ],
              "notification_email": "someone@example.com",
              "check_regions": [
                "WEU"
              ]
            }`, string(b))
		}
		fmt.Fprint(w, `{
            "success": true,
            "errors": [],
            "messages": [],
            "result": {
              "id": "17b5962d775c646f3f9725cbc7a53df4",
              "created_on": "2014-01-01T05:20:00.12345Z",
              "modified_on": "2014-02-01T05:20:00.12345Z",
              "description": "Primary data center - Provider XYZ",
              "name": "primary-dc-1",
              "enabled": true,
              "minimum_origins": 1,
              "monitor": "f1aba936b94213e5b8dca0c0dbf1f9cc",
              "latitude": 55,
              "longitude": -12.5,
              "load_shedding": {
                "default_percent": 50,
                "default_policy": "random",
                "session_percent": 10,
                "session_policy": "hash"
              },
              "origin_steering": {
                "policy": "random"
              },
              "origins": [
                {
                  "name": "app-server-1",
                  "address": "198.51.100.1",
                  "enabled": true,
                  "weight": 1,
                  "header": {
                      "Host": [
                          "example.com"
                      ]
                  },
			      "virtual_network_id":"a5624d4e-044a-4ff0-b3e1-e2465353d4b4"
                }
              ],
              "notification_email": "someone@example.com",
              "check_regions": [
                "WEU"
              ],
			  "healthy": true
            }
        }`)
	}

	fptr := func(f float32) *float32 {
		return &f
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/load_balancers/pools", handler)
	createdOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2014-02-01T05:20:00.12345Z")
	want := LoadBalancerPool{
		ID:             "17b5962d775c646f3f9725cbc7a53df4",
		CreatedOn:      &createdOn,
		ModifiedOn:     &modifiedOn,
		Description:    "Primary data center - Provider XYZ",
		Name:           "primary-dc-1",
		Enabled:        true,
		MinimumOrigins: IntPtr(1),
		Monitor:        "f1aba936b94213e5b8dca0c0dbf1f9cc",
		Latitude:       fptr(55),
		Longitude:      fptr(-12.5),
		LoadShedding: &LoadBalancerLoadShedding{
			DefaultPercent: 50,
			DefaultPolicy:  "random",
			SessionPercent: 10,
			SessionPolicy:  "hash",
		},
		OriginSteering: &LoadBalancerOriginSteering{
			Policy: "random",
		},
		Origins: []LoadBalancerOrigin{
			{
				Name:    "app-server-1",
				Address: "198.51.100.1",
				Enabled: true,
				Weight:  1,
				Header: map[string][]string{
					"Host": {"example.com"},
				},
				VirtualNetworkID: "a5624d4e-044a-4ff0-b3e1-e2465353d4b4",
			},
		},
		NotificationEmail: "someone@example.com",
		CheckRegions: []string{
			"WEU",
		},
		Healthy: BoolPtr(true),
	}
	request := LoadBalancerPool{
		Description: "Primary data center - Provider XYZ",
		Name:        "primary-dc-1",
		Enabled:     true,
		Monitor:     "f1aba936b94213e5b8dca0c0dbf1f9cc",
		Latitude:    fptr(55),
		Longitude:   fptr(-12.5),
		LoadShedding: &LoadBalancerLoadShedding{
			DefaultPercent: 50,
			DefaultPolicy:  "random",
			SessionPercent: 10,
			SessionPolicy:  "hash",
		},
		OriginSteering: &LoadBalancerOriginSteering{
			Policy: "random",
		},
		Origins: []LoadBalancerOrigin{
			{
				Name:    "app-server-1",
				Address: "198.51.100.1",
				Enabled: true,
				Weight:  1,
				Header: map[string][]string{
					"Host": {"example.com"},
				},
				VirtualNetworkID: "a5624d4e-044a-4ff0-b3e1-e2465353d4b4",
			},
		},
		NotificationEmail: "someone@example.com",
		CheckRegions: []string{
			"WEU",
		},
	}

	actual, err := client.CreateLoadBalancerPool(context.Background(), AccountIdentifier(testAccountID), CreateLoadBalancerPoolParams{LoadBalancerPool: request})
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateLoadBalancerPool_MinimumOriginsZero(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		b, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if assert.NoError(t, err) {
			assert.JSONEq(t, `{
              "description": "Primary data center - Provider XYZ",
              "name": "primary-dc-2",
              "minimum_origins": 0,
              "enabled": true,
              "check_regions": null,
              "origins": null
            }`, string(b))
		}
		fmt.Fprint(w, `{
            "success": true,
            "errors": [],
            "messages": [],
            "result": {
              "description": "Primary data center - Provider XYZ",
              "created_on": "2014-01-01T05:20:00.12345Z",
              "modified_on": "2014-02-01T05:20:00.12345Z",
              "id": "f6fea70e5154b4c563bd549ef405b7d7",
              "enabled": true,
              "minimum_origins": 0,
              "name": "primary-dc-2",
              "notification_email": "",
              "check_regions": null,
              "origins": []
            }
        }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/load_balancers/pools", handler)
	createdOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2014-02-01T05:20:00.12345Z")
	want := LoadBalancerPool{
		ID:                "f6fea70e5154b4c563bd549ef405b7d7",
		CreatedOn:         &createdOn,
		ModifiedOn:        &modifiedOn,
		Description:       "Primary data center - Provider XYZ",
		Name:              "primary-dc-2",
		Enabled:           true,
		MinimumOrigins:    IntPtr(0),
		Origins:           []LoadBalancerOrigin{},
		NotificationEmail: "",
	}
	request := LoadBalancerPool{
		Description:    "Primary data center - Provider XYZ",
		Name:           "primary-dc-2",
		Enabled:        true,
		MinimumOrigins: IntPtr(0),
	}

	actual, err := client.CreateLoadBalancerPool(context.Background(), AccountIdentifier(testAccountID), CreateLoadBalancerPoolParams{LoadBalancerPool: request})
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateLoadBalancerPool_ZoneIsNotSupported(t *testing.T) {
	setup()
	defer teardown()

	_, err := client.CreateLoadBalancerPool(context.Background(), ZoneIdentifier(testZoneID), CreateLoadBalancerPoolParams{})
	if assert.Error(t, err) {
		assert.Equal(t, fmt.Sprintf(errInvalidResourceContainerAccess, ZoneRouteLevel), err.Error())
	}
}

func TestListLoadBalancerPools(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
            "success": true,
            "errors": [],
            "messages": [],
            "result": [
                {
                    "id": "17b5962d775c646f3f9725cbc7a53df4",
                    "created_on": "2014-01-01T05:20:00.12345Z",
                    "modified_on": "2014-02-01T05:20:00.12345Z",
                    "description": "Primary data center - Provider XYZ",
                    "name": "primary-dc-1",
                    "enabled": true,
                    "monitor": "f1aba936b94213e5b8dca0c0dbf1f9cc",
                    "origin_steering": {
                      "policy": "random"
                    },
                    "origins": [
                      {
                        "name": "app-server-1",
                        "address": "198.51.100.1",
                        "enabled": true,
                        "weight": 1,
					    "virtual_network_id":"a5624d4e-044a-4ff0-b3e1-e2465353d4b4"
                      }
                    ],
                    "notification_email": "someone@example.com"
                }
            ],
            "result_info": {
                "page": 1,
                "per_page": 20,
                "count": 1,
                "total_count": 1
            }
        }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/load_balancers/pools", handler)
	createdOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2014-02-01T05:20:00.12345Z")
	want := []LoadBalancerPool{
		{
			ID:          "17b5962d775c646f3f9725cbc7a53df4",
			CreatedOn:   &createdOn,
			ModifiedOn:  &modifiedOn,
			Description: "Primary data center - Provider XYZ",
			Name:        "primary-dc-1",
			Enabled:     true,
			Monitor:     "f1aba936b94213e5b8dca0c0dbf1f9cc",
			OriginSteering: &LoadBalancerOriginSteering{
				Policy: "random",
			},
			Origins: []LoadBalancerOrigin{
				{
					Name:             "app-server-1",
					Address:          "198.51.100.1",
					Enabled:          true,
					Weight:           1,
					VirtualNetworkID: "a5624d4e-044a-4ff0-b3e1-e2465353d4b4",
				},
			},
			NotificationEmail: "someone@example.com",
		},
	}

	actual, err := client.ListLoadBalancerPools(context.Background(), AccountIdentifier(testAccountID), ListLoadBalancerPoolParams{})
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestListLoadBalancerPool_ZoneIsNotSupported(t *testing.T) {
	setup()
	defer teardown()

	_, err := client.ListLoadBalancerPools(context.Background(), ZoneIdentifier(testZoneID), ListLoadBalancerPoolParams{})
	if assert.Error(t, err) {
		assert.Equal(t, fmt.Sprintf(errInvalidResourceContainerAccess, ZoneRouteLevel), err.Error())
	}
}

func TestGetLoadBalancerPool(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
            "success": true,
            "errors": [],
            "messages": [],
            "result": {
              "id": "17b5962d775c646f3f9725cbc7a53df4",
              "created_on": "2014-01-01T05:20:00.12345Z",
              "modified_on": "2014-02-01T05:20:00.12345Z",
              "description": "Primary data center - Provider XYZ",
              "name": "primary-dc-1",
              "enabled": true,
              "monitor": "f1aba936b94213e5b8dca0c0dbf1f9cc",
              "origin_steering": {
                "policy": "random"
              },
              "origins": [
                {
                  "name": "app-server-1",
                  "address": "198.51.100.1",
                  "enabled": true,
                  "weight": 1,
				  "virtual_network_id":"a5624d4e-044a-4ff0-b3e1-e2465353d4b4"
                }
              ],
              "notification_email": "someone@example.com"
            }
        }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/load_balancers/pools/17b5962d775c646f3f9725cbc7a53df4", handler)
	createdOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2014-02-01T05:20:00.12345Z")
	want := LoadBalancerPool{
		ID:          "17b5962d775c646f3f9725cbc7a53df4",
		CreatedOn:   &createdOn,
		ModifiedOn:  &modifiedOn,
		Description: "Primary data center - Provider XYZ",
		Name:        "primary-dc-1",
		Enabled:     true,
		Monitor:     "f1aba936b94213e5b8dca0c0dbf1f9cc",
		OriginSteering: &LoadBalancerOriginSteering{
			Policy: "random",
		},
		Origins: []LoadBalancerOrigin{
			{
				Name:             "app-server-1",
				Address:          "198.51.100.1",
				Enabled:          true,
				Weight:           1,
				VirtualNetworkID: "a5624d4e-044a-4ff0-b3e1-e2465353d4b4",
			},
		},
		NotificationEmail: "someone@example.com",
	}

	actual, err := client.GetLoadBalancerPool(context.Background(), AccountIdentifier(testAccountID), "17b5962d775c646f3f9725cbc7a53df4")
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}

	_, err = client.GetLoadBalancerPool(context.Background(), AccountIdentifier(testAccountID), "bar")
	assert.Error(t, err)
}

func TestGetLoadBalancerPool_ZoneIsNotSupported(t *testing.T) {
	setup()
	defer teardown()

	_, err := client.GetLoadBalancerPool(context.Background(), ZoneIdentifier(testZoneID), "foo")
	if assert.Error(t, err) {
		assert.Equal(t, fmt.Sprintf(errInvalidResourceContainerAccess, ZoneRouteLevel), err.Error())
	}
}

func TestDeleteLoadBalancerPool(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
            "success": true,
            "errors": [],
            "messages": [],
            "result": {
              "id": "17b5962d775c646f3f9725cbc7a53df4"
            }
        }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/load_balancers/pools/17b5962d775c646f3f9725cbc7a53df4", handler)
	assert.NoError(t, client.DeleteLoadBalancerPool(context.Background(), AccountIdentifier(testAccountID), "17b5962d775c646f3f9725cbc7a53df4"))
	assert.Error(t, client.DeleteLoadBalancerPool(context.Background(), AccountIdentifier(testAccountID), "bar"))
}

func TestDeleteLoadBalancerPool_ZoneIsNotSupported(t *testing.T) {
	setup()
	defer teardown()

	err := client.DeleteLoadBalancerPool(context.Background(), ZoneIdentifier(testZoneID), "foo")
	if assert.Error(t, err) {
		assert.Equal(t, fmt.Sprintf(errInvalidResourceContainerAccess, ZoneRouteLevel), err.Error())
	}
}

func TestUpdateLoadBalancerPool(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		b, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if assert.NoError(t, err) {
			assert.JSONEq(t, `{
              "id": "17b5962d775c646f3f9725cbc7a53df4",
              "description": "Primary data center - Provider XYZZY",
              "name": "primary-dc-2",
              "enabled": false,
              "origin_steering": {
                "policy": "random"
              },
              "origins": [
                {
                  "name": "app-server-2",
                  "address": "198.51.100.2",
                  "enabled": false,
                  "weight": 1,
                  "header": {
                      "Host": [
                          "example.com"
                      ]
                  },
			      "virtual_network_id":"a5624d4e-044a-4ff0-b3e1-e2465353d4b4"
                }
              ],
              "notification_email": "nobody@example.com",
              "check_regions": [
                "WEU"
              ]
						}`, string(b))
		}
		fmt.Fprint(w, `{
            "success": true,
            "errors": [],
            "messages": [],
            "result": {
              "id": "17b5962d775c646f3f9725cbc7a53df4",
              "created_on": "2014-01-01T05:20:00.12345Z",
              "modified_on": "2017-02-01T05:20:00.12345Z",
              "description": "Primary data center - Provider XYZZY",
              "name": "primary-dc-2",
              "enabled": false,
              "origin_steering": {
                "policy": "random"
              },
              "origins": [
                {
                  "name": "app-server-2",
                  "address": "198.51.100.2",
                  "enabled": false,
                  "weight": 1,
                  "header": {
                      "Host": [
                          "example.com"
                      ]
                  },
			      "virtual_network_id":"a5624d4e-044a-4ff0-b3e1-e2465353d4b4"
                }
              ],
              "notification_email": "nobody@example.com",
              "check_regions": [
                "WEU"
              ]
            }
        }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/load_balancers/pools/17b5962d775c646f3f9725cbc7a53df4", handler)
	createdOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2017-02-01T05:20:00.12345Z")
	want := LoadBalancerPool{
		ID:          "17b5962d775c646f3f9725cbc7a53df4",
		CreatedOn:   &createdOn,
		ModifiedOn:  &modifiedOn,
		Description: "Primary data center - Provider XYZZY",
		Name:        "primary-dc-2",
		Enabled:     false,
		OriginSteering: &LoadBalancerOriginSteering{
			Policy: "random",
		},
		Origins: []LoadBalancerOrigin{
			{
				Name:    "app-server-2",
				Address: "198.51.100.2",
				Enabled: false,
				Weight:  1,
				Header: map[string][]string{
					"Host": {"example.com"},
				},
				VirtualNetworkID: "a5624d4e-044a-4ff0-b3e1-e2465353d4b4",
			},
		},
		NotificationEmail: "nobody@example.com",
		CheckRegions: []string{
			"WEU",
		},
	}
	request := LoadBalancerPool{
		ID:          "17b5962d775c646f3f9725cbc7a53df4",
		Description: "Primary data center - Provider XYZZY",
		Name:        "primary-dc-2",
		Enabled:     false,
		OriginSteering: &LoadBalancerOriginSteering{
			Policy: "random",
		},
		Origins: []LoadBalancerOrigin{
			{
				Name:    "app-server-2",
				Address: "198.51.100.2",
				Enabled: false,
				Weight:  1,
				Header: map[string][]string{
					"Host": {"example.com"},
				},
				VirtualNetworkID: "a5624d4e-044a-4ff0-b3e1-e2465353d4b4",
			},
		},
		NotificationEmail: "nobody@example.com",
		CheckRegions: []string{
			"WEU",
		},
	}

	actual, err := client.UpdateLoadBalancerPool(context.Background(), AccountIdentifier(testAccountID), UpdateLoadBalancerPoolParams{LoadBalancer: request})
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestUpdateLoadBalancerPool_ZoneIsNotSupported(t *testing.T) {
	setup()
	defer teardown()

	_, err := client.UpdateLoadBalancerPool(context.Background(), ZoneIdentifier(testZoneID), UpdateLoadBalancerPoolParams{})
	if assert.Error(t, err) {
		assert.Equal(t, fmt.Sprintf(errInvalidResourceContainerAccess, ZoneRouteLevel), err.Error())
	}
}

func TestCreateLoadBalancerMonitor(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		b, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if assert.NoError(t, err) {
			assert.JSONEq(t, `{
              "type": "https",
              "description": "Login page monitor",
              "method": "GET",
              "path": "/health",
              "header": {
                "Host": [
                  "example.com"
                ],
                "X-App-ID": [
                  "abc123"
                ]
              },
              "timeout": 3,
              "retries": 0,
              "interval": 90,
              "consecutive_up": 2,
              "consecutive_down": 2,
              "expected_body": "alive",
              "expected_codes": "2xx",
              "follow_redirects": true,
              "allow_insecure": true,
              "probe_zone": ""
						}`, string(b))
		}
		fmt.Fprint(w, `{
            "success": true,
            "errors": [],
            "messages": [],
            "result": {
                "id": "f1aba936b94213e5b8dca0c0dbf1f9cc",
                "created_on": "2014-01-01T05:20:00.12345Z",
                "modified_on": "2014-02-01T05:20:00.12345Z",
                "type": "https",
                "description": "Login page monitor",
                "method": "GET",
                "path": "/health",
                "header": {
                  "Host": [
                    "example.com"
                  ],
                  "X-App-ID": [
                    "abc123"
                  ]
                },
                "timeout": 3,
                "retries": 0,
                "interval": 90,
                "consecutive_up": 2,
                "consecutive_down": 2,
                "expected_body": "alive",
                "expected_codes": "2xx",
                "follow_redirects": true,
                "allow_insecure": true,
                "probe_zone": ""
            }
        }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/load_balancers/monitors", handler)
	createdOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2014-02-01T05:20:00.12345Z")
	want := LoadBalancerMonitor{
		ID:          "f1aba936b94213e5b8dca0c0dbf1f9cc",
		CreatedOn:   &createdOn,
		ModifiedOn:  &modifiedOn,
		Type:        "https",
		Description: "Login page monitor",
		Method:      http.MethodGet,
		Path:        "/health",
		Header: map[string][]string{
			"Host":     {"example.com"},
			"X-App-ID": {"abc123"},
		},
		Timeout:         3,
		Retries:         0,
		Interval:        90,
		ConsecutiveUp:   2,
		ConsecutiveDown: 2,
		ExpectedBody:    "alive",
		ExpectedCodes:   "2xx",

		FollowRedirects: true,
		AllowInsecure:   true,
	}
	request := LoadBalancerMonitor{
		Type:        "https",
		Description: "Login page monitor",
		Method:      http.MethodGet,
		Path:        "/health",
		Header: map[string][]string{
			"Host":     {"example.com"},
			"X-App-ID": {"abc123"},
		},
		Timeout:         3,
		Retries:         0,
		Interval:        90,
		ConsecutiveUp:   2,
		ConsecutiveDown: 2,
		ExpectedBody:    "alive",
		ExpectedCodes:   "2xx",

		FollowRedirects: true,
		AllowInsecure:   true,
	}

	actual, err := client.CreateLoadBalancerMonitor(context.Background(), AccountIdentifier(testAccountID), CreateLoadBalancerMonitorParams{LoadBalancerMonitor: request})
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateLoadBalancerMonitor_ZoneIsNotSupported(t *testing.T) {
	setup()
	defer teardown()

	_, err := client.CreateLoadBalancerMonitor(context.Background(), ZoneIdentifier(testZoneID), CreateLoadBalancerMonitorParams{})
	if assert.Error(t, err) {
		assert.Equal(t, fmt.Sprintf(errInvalidResourceContainerAccess, ZoneRouteLevel), err.Error())
	}
}

func TestListLoadBalancerMonitors(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
            "success": true,
            "errors": [],
            "messages": [],
            "result": [
                {
                    "id": "f1aba936b94213e5b8dca0c0dbf1f9cc",
                    "created_on": "2014-01-01T05:20:00.12345Z",
                    "modified_on": "2014-02-01T05:20:00.12345Z",
                    "type": "https",
                    "description": "Login page monitor",
                    "method": "GET",
                    "path": "/health",
                    "header": {
                      "Host": [
                        "example.com"
                      ],
                      "X-App-ID": [
                        "abc123"
                      ]
                    },
                    "timeout": 3,
                    "retries": 0,
                    "interval": 90,
                    "consecutive_up": 2,
                    "consecutive_down": 2,
                    "expected_body": "alive",
                    "expected_codes": "2xx"
                }
            ],
            "result_info": {
                "page": 1,
                "per_page": 20,
                "count": 1,
                "total_count": 1
            }
        }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/load_balancers/monitors", handler)
	createdOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2014-02-01T05:20:00.12345Z")
	want := []LoadBalancerMonitor{
		{
			ID:          "f1aba936b94213e5b8dca0c0dbf1f9cc",
			CreatedOn:   &createdOn,
			ModifiedOn:  &modifiedOn,
			Type:        "https",
			Description: "Login page monitor",
			Method:      http.MethodGet,
			Path:        "/health",
			Header: map[string][]string{
				"Host":     {"example.com"},
				"X-App-ID": {"abc123"},
			},
			Timeout:         3,
			Retries:         0,
			Interval:        90,
			ConsecutiveUp:   2,
			ConsecutiveDown: 2,
			ExpectedBody:    "alive",
			ExpectedCodes:   "2xx",
		},
	}

	actual, err := client.ListLoadBalancerMonitors(context.Background(), AccountIdentifier(testAccountID), ListLoadBalancerMonitorParams{})
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestListLoadBalancerMonitors_ZoneIsNotSupported(t *testing.T) {
	setup()
	defer teardown()

	_, err := client.ListLoadBalancerMonitors(context.Background(), ZoneIdentifier(testZoneID), ListLoadBalancerMonitorParams{})
	if assert.Error(t, err) {
		assert.Equal(t, fmt.Sprintf(errInvalidResourceContainerAccess, ZoneRouteLevel), err.Error())
	}
}

func TestGetLoadBalancerMonitor(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
            "success": true,
            "errors": [],
            "messages": [],
            "result": {
                "id": "f1aba936b94213e5b8dca0c0dbf1f9cc",
                "created_on": "2014-01-01T05:20:00.12345Z",
                "modified_on": "2014-02-01T05:20:00.12345Z",
                "type": "https",
                "description": "Login page monitor",
                "method": "GET",
                "path": "/health",
                "header": {
                  "Host": [
                    "example.com"
                  ],
                  "X-App-ID": [
                    "abc123"
                  ]
                },
                "timeout": 3,
                "retries": 0,
                "interval": 90,
                "consecutive_up": 2,
                "consecutive_down": 2,
                "expected_body": "alive",
                "expected_codes": "2xx",
                "follow_redirects": true,
                "allow_insecure": true,
                "probe_zone": ""
            }
        }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/load_balancers/monitors/f1aba936b94213e5b8dca0c0dbf1f9cc", handler)
	createdOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2014-02-01T05:20:00.12345Z")
	want := LoadBalancerMonitor{
		ID:          "f1aba936b94213e5b8dca0c0dbf1f9cc",
		CreatedOn:   &createdOn,
		ModifiedOn:  &modifiedOn,
		Type:        "https",
		Description: "Login page monitor",
		Method:      http.MethodGet,
		Path:        "/health",
		Header: map[string][]string{
			"Host":     {"example.com"},
			"X-App-ID": {"abc123"},
		},
		Timeout:         3,
		Retries:         0,
		Interval:        90,
		ConsecutiveUp:   2,
		ConsecutiveDown: 2,
		ExpectedBody:    "alive",
		ExpectedCodes:   "2xx",

		FollowRedirects: true,
		AllowInsecure:   true,
	}

	actual, err := client.GetLoadBalancerMonitor(context.Background(), AccountIdentifier(testAccountID), "f1aba936b94213e5b8dca0c0dbf1f9cc")
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}

	_, err = client.GetLoadBalancerMonitor(context.Background(), AccountIdentifier(testAccountID), "bar")
	assert.Error(t, err)
}

func TestGetLoadBalancerMonitor_ZoneIsNotSupported(t *testing.T) {
	setup()
	defer teardown()

	_, err := client.GetLoadBalancerMonitor(context.Background(), ZoneIdentifier(testZoneID), "foo")
	if assert.Error(t, err) {
		assert.Equal(t, fmt.Sprintf(errInvalidResourceContainerAccess, ZoneRouteLevel), err.Error())
	}
}

func TestDeleteLoadBalancerMonitor(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
            "success": true,
            "errors": [],
            "messages": [],
            "result": {
              "id": "f1aba936b94213e5b8dca0c0dbf1f9cc"
            }
        }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/load_balancers/monitors/f1aba936b94213e5b8dca0c0dbf1f9cc", handler)
	assert.NoError(t, client.DeleteLoadBalancerMonitor(context.Background(), AccountIdentifier(testAccountID), "f1aba936b94213e5b8dca0c0dbf1f9cc"))
	assert.Error(t, client.DeleteLoadBalancerMonitor(context.Background(), AccountIdentifier(testAccountID), "bar"))
}

func TestDeleteLoadBalancerMonitor_ZoneIsNotSupported(t *testing.T) {
	setup()
	defer teardown()

	err := client.DeleteLoadBalancerMonitor(context.Background(), ZoneIdentifier(testZoneID), "foo")
	if assert.Error(t, err) {
		assert.Equal(t, fmt.Sprintf(errInvalidResourceContainerAccess, ZoneRouteLevel), err.Error())
	}
}

func TestUpdateLoadBalancerMonitor(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		b, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if assert.NoError(t, err) {
			assert.JSONEq(t, `{
                "id": "f1aba936b94213e5b8dca0c0dbf1f9cc",
                "type": "http",
                "description": "Login page monitor",
                "method": "GET",
                "path": "/status",
                "header": {
                  "Host": [
                    "example.com"
                  ],
                  "X-App-ID": [
                    "easy"
                  ]
                },
                "timeout": 3,
                "retries": 0,
                "interval": 90,
                "consecutive_up": 2,
                "consecutive_down": 2,
                "expected_body": "kicking",
                "expected_codes": "200",
                "follow_redirects": true,
                "allow_insecure": true,
                "probe_zone": ""
						}`, string(b))
		}
		fmt.Fprint(w, `{
            "success": true,
            "errors": [],
            "messages": [],
            "result": {
                "id": "f1aba936b94213e5b8dca0c0dbf1f9cc",
                "created_on": "2014-01-01T05:20:00.12345Z",
                "modified_on": "2017-02-01T05:20:00.12345Z",
                "type": "http",
                "description": "Login page monitor",
                "method": "GET",
                "path": "/status",
                "header": {
                  "Host": [
                    "example.com"
                  ],
                  "X-App-ID": [
                    "easy"
                  ]
                },
                "timeout": 3,
                "retries": 0,
                "interval": 90,
                "consecutive_up": 2,
                "consecutive_down": 2,
                "expected_body": "kicking",
                "expected_codes": "200",
                "follow_redirects": true,
                "allow_insecure": true,
                "probe_zone": ""
            }
        }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/load_balancers/monitors/f1aba936b94213e5b8dca0c0dbf1f9cc", handler)
	createdOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2017-02-01T05:20:00.12345Z")
	want := LoadBalancerMonitor{
		ID:          "f1aba936b94213e5b8dca0c0dbf1f9cc",
		CreatedOn:   &createdOn,
		ModifiedOn:  &modifiedOn,
		Type:        "http",
		Description: "Login page monitor",
		Method:      http.MethodGet,
		Path:        "/status",
		Header: map[string][]string{
			"Host":     {"example.com"},
			"X-App-ID": {"easy"},
		},
		Timeout:         3,
		Retries:         0,
		Interval:        90,
		ConsecutiveUp:   2,
		ConsecutiveDown: 2,
		ExpectedBody:    "kicking",
		ExpectedCodes:   "200",

		FollowRedirects: true,
		AllowInsecure:   true,
	}
	request := LoadBalancerMonitor{
		ID:          "f1aba936b94213e5b8dca0c0dbf1f9cc",
		Type:        "http",
		Description: "Login page monitor",
		Method:      http.MethodGet,
		Path:        "/status",
		Header: map[string][]string{
			"Host":     {"example.com"},
			"X-App-ID": {"easy"},
		},
		Timeout:         3,
		Retries:         0,
		Interval:        90,
		ConsecutiveUp:   2,
		ConsecutiveDown: 2,
		ExpectedBody:    "kicking",
		ExpectedCodes:   "200",

		FollowRedirects: true,
		AllowInsecure:   true,
	}

	actual, err := client.UpdateLoadBalancerMonitor(context.Background(), AccountIdentifier(testAccountID), UpdateLoadBalancerMonitorParams{LoadBalancerMonitor: request})
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestUpdateLoadBalancerMonitor_ZoneIsNotSupported(t *testing.T) {
	setup()
	defer teardown()

	_, err := client.UpdateLoadBalancerMonitor(context.Background(), ZoneIdentifier(testZoneID), UpdateLoadBalancerMonitorParams{})
	if assert.Error(t, err) {
		assert.Equal(t, fmt.Sprintf(errInvalidResourceContainerAccess, ZoneRouteLevel), err.Error())
	}
}

func TestCreateLoadBalancer(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		b, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if assert.NoError(t, err) {
			assert.JSONEq(t, `{
              "description": "Load Balancer for www.example.com",
              "name": "www.example.com",
              "ttl": 30,
              "fallback_pool": "17b5962d775c646f3f9725cbc7a53df4",
              "default_pools": [
                "de90f38ced07c2e2f4df50b1f61d4194",
                "9290f38c5d07c2e2f4df57b1f61d4196",
                "00920f38ce07c2e2f4df50b1f61d4194"
              ],
              "region_pools": {
                "WNAM": [
                  "de90f38ced07c2e2f4df50b1f61d4194",
                  "9290f38c5d07c2e2f4df57b1f61d4196"
                ],
                "ENAM": [
                  "00920f38ce07c2e2f4df50b1f61d4194"
                ]
              },
              "country_pools": {
                "US": [
                  "de90f38ced07c2e2f4df50b1f61d4194"
                ],
                "GB": [
                  "abd90f38ced07c2e2f4df50b1f61d4194"
                ]
              },
              "pop_pools": {
                "LAX": [
                  "de90f38ced07c2e2f4df50b1f61d4194",
                  "9290f38c5d07c2e2f4df57b1f61d4196"
                ],
                "LHR": [
                  "abd90f38ced07c2e2f4df50b1f61d4194",
                  "f9138c5d07c2e2f4df57b1f61d4196"
                ],
                "SJC": [
                  "00920f38ce07c2e2f4df50b1f61d4194"
                ]
			  },
			  "random_steering": {
			    "default_weight": 0.2,
			    "pool_weights": {
			        "9290f38c5d07c2e2f4df57b1f61d4196": 0.6,
			        "de90f38ced07c2e2f4df50b1f61d4194": 0.4
			    }
			  },
			  "adaptive_routing": {
				"failover_across_pools": true
			  },
			  "location_strategy": {
				"prefer_ecs": "always",
				"mode": "resolver_ip"
			  },
			  "rules": [
				  {
					  "name": "example rule",
					  "condition": "cf.load_balancer.region == \"SAF\"",
					  "disabled": false,
					  "priority": 0,
					  "overrides": {
						  "region_pools": {
							  "SAF": ["de90f38ced07c2e2f4df50b1f61d4194"]
						  },
						  "adaptive_routing": {
							"failover_across_pools": false
						  },
						  "location_strategy": {
							"prefer_ecs": "never",
							"mode": "pop"
						  }
					  }
				  }
			  ],
              "proxied": true,
              "session_affinity": "cookie",
              "session_affinity_ttl": 5000,
              "session_affinity_attributes": {
                "samesite": "Strict",
                "secure": "Always",
                "drain_duration": 60,
                "zero_downtime_failover": "sticky"
              }
            }`, string(b))
		}

		fmt.Fprint(w, `{
            "success": true,
            "errors": [],
            "messages": [],
            "result": {
                "id": "699d98642c564d2e855e9661899b7252",
                "created_on": "2014-01-01T05:20:00.12345Z",
                "modified_on": "2014-02-01T05:20:00.12345Z",
                "description": "Load Balancer for www.example.com",
                "name": "www.example.com",
                "ttl": 30,
                "fallback_pool": "17b5962d775c646f3f9725cbc7a53df4",
                "default_pools": [
                  "de90f38ced07c2e2f4df50b1f61d4194",
                  "9290f38c5d07c2e2f4df57b1f61d4196",
                  "00920f38ce07c2e2f4df50b1f61d4194"
                ],
                "region_pools": {
                  "WNAM": [
                    "de90f38ced07c2e2f4df50b1f61d4194",
                    "9290f38c5d07c2e2f4df57b1f61d4196"
                  ],
                  "ENAM": [
                    "00920f38ce07c2e2f4df50b1f61d4194"
                  ]
                },
                "country_pools": {
                  "US": [
                    "de90f38ced07c2e2f4df50b1f61d4194"
                  ],
                  "GB": [
                    "abd90f38ced07c2e2f4df50b1f61d4194"
                  ]
                },
                "pop_pools": {
                  "LAX": [
                    "de90f38ced07c2e2f4df50b1f61d4194",
                    "9290f38c5d07c2e2f4df57b1f61d4196"
                  ],
                  "LHR": [
                    "abd90f38ced07c2e2f4df50b1f61d4194",
                    "f9138c5d07c2e2f4df57b1f61d4196"
                  ],
                  "SJC": [
                    "00920f38ce07c2e2f4df50b1f61d4194"
                  ]
				},
				"random_steering": {
				   "default_weight": 0.2,
				    "pool_weights": {
				        "9290f38c5d07c2e2f4df57b1f61d4196": 0.6,
				        "de90f38ced07c2e2f4df50b1f61d4194": 0.4
				    }
				},
				"adaptive_routing": {
					"failover_across_pools": true
				},
				"location_strategy": {
					"prefer_ecs": "always",
					"mode": "resolver_ip"
				},
				"rules": [
				  {
					  "name": "example rule",
					  "condition": "cf.load_balancer.region == \"SAF\"",
					  "overrides": {
						  "region_pools": {
							  "SAF": ["de90f38ced07c2e2f4df50b1f61d4194"]
						  },
						  "adaptive_routing": {
							"failover_across_pools": false
						  },
						  "location_strategy": {
							"prefer_ecs": "never",
							"mode": "pop"
						  }
					  }
				  }
			  ],
                "proxied": true,
                "session_affinity": "cookie",
                "session_affinity_ttl": 5000,
                "session_affinity_attributes": {
                    "samesite": "Strict",
                    "secure": "Always",
                    "drain_duration": 60,
	                "zero_downtime_failover": "sticky"
                }
            }
        }`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/load_balancers", handler)
	createdOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2014-02-01T05:20:00.12345Z")
	want := LoadBalancer{
		ID:           "699d98642c564d2e855e9661899b7252",
		CreatedOn:    &createdOn,
		ModifiedOn:   &modifiedOn,
		Description:  "Load Balancer for www.example.com",
		Name:         "www.example.com",
		TTL:          30,
		FallbackPool: "17b5962d775c646f3f9725cbc7a53df4",
		DefaultPools: []string{
			"de90f38ced07c2e2f4df50b1f61d4194",
			"9290f38c5d07c2e2f4df57b1f61d4196",
			"00920f38ce07c2e2f4df50b1f61d4194",
		},
		RegionPools: map[string][]string{
			"WNAM": {
				"de90f38ced07c2e2f4df50b1f61d4194",
				"9290f38c5d07c2e2f4df57b1f61d4196",
			},
			"ENAM": {
				"00920f38ce07c2e2f4df50b1f61d4194",
			},
		},
		CountryPools: map[string][]string{
			"US": {
				"de90f38ced07c2e2f4df50b1f61d4194",
			},
			"GB": {
				"abd90f38ced07c2e2f4df50b1f61d4194",
			},
		},
		PopPools: map[string][]string{
			"LAX": {
				"de90f38ced07c2e2f4df50b1f61d4194",
				"9290f38c5d07c2e2f4df57b1f61d4196",
			},
			"LHR": {
				"abd90f38ced07c2e2f4df50b1f61d4194",
				"f9138c5d07c2e2f4df57b1f61d4196",
			},
			"SJC": {
				"00920f38ce07c2e2f4df50b1f61d4194",
			},
		},
		RandomSteering: &RandomSteering{
			DefaultWeight: 0.2,
			PoolWeights: map[string]float64{
				"9290f38c5d07c2e2f4df57b1f61d4196": 0.6,
				"de90f38ced07c2e2f4df50b1f61d4194": 0.4,
			},
		},
		AdaptiveRouting: &AdaptiveRouting{
			FailoverAcrossPools: BoolPtr(true),
		},
		LocationStrategy: &LocationStrategy{
			PreferECS: "always",
			Mode:      "resolver_ip",
		},
		Rules: []*LoadBalancerRule{
			{
				Name:      "example rule",
				Condition: "cf.load_balancer.region == \"SAF\"",
				Overrides: LoadBalancerRuleOverrides{
					RegionPools: map[string][]string{
						"SAF": {"de90f38ced07c2e2f4df50b1f61d4194"},
					},
					AdaptiveRouting: &AdaptiveRouting{
						FailoverAcrossPools: BoolPtr(false),
					},
					LocationStrategy: &LocationStrategy{
						PreferECS: "never",
						Mode:      "pop",
					},
				},
			},
		},
		Proxied:        true,
		Persistence:    "cookie",
		PersistenceTTL: 5000,
		SessionAffinityAttributes: &SessionAffinityAttributes{
			SameSite:             "Strict",
			Secure:               "Always",
			DrainDuration:        60,
			ZeroDowntimeFailover: "sticky",
		},
	}
	request := LoadBalancer{
		Description:  "Load Balancer for www.example.com",
		Name:         "www.example.com",
		TTL:          30,
		FallbackPool: "17b5962d775c646f3f9725cbc7a53df4",
		DefaultPools: []string{
			"de90f38ced07c2e2f4df50b1f61d4194",
			"9290f38c5d07c2e2f4df57b1f61d4196",
			"00920f38ce07c2e2f4df50b1f61d4194",
		},
		RegionPools: map[string][]string{
			"WNAM": {
				"de90f38ced07c2e2f4df50b1f61d4194",
				"9290f38c5d07c2e2f4df57b1f61d4196",
			},
			"ENAM": {
				"00920f38ce07c2e2f4df50b1f61d4194",
			},
		},
		CountryPools: map[string][]string{
			"US": {
				"de90f38ced07c2e2f4df50b1f61d4194",
			},
			"GB": {
				"abd90f38ced07c2e2f4df50b1f61d4194",
			},
		},
		PopPools: map[string][]string{
			"LAX": {
				"de90f38ced07c2e2f4df50b1f61d4194",
				"9290f38c5d07c2e2f4df57b1f61d4196",
			},
			"LHR": {
				"abd90f38ced07c2e2f4df50b1f61d4194",
				"f9138c5d07c2e2f4df57b1f61d4196",
			},
			"SJC": {
				"00920f38ce07c2e2f4df50b1f61d4194",
			},
		},
		RandomSteering: &RandomSteering{
			DefaultWeight: 0.2,
			PoolWeights: map[string]float64{
				"9290f38c5d07c2e2f4df57b1f61d4196": 0.6,
				"de90f38ced07c2e2f4df50b1f61d4194": 0.4,
			},
		},
		AdaptiveRouting: &AdaptiveRouting{
			FailoverAcrossPools: BoolPtr(true),
		},
		LocationStrategy: &LocationStrategy{
			PreferECS: "always",
			Mode:      "resolver_ip",
		},
		Rules: []*LoadBalancerRule{
			{
				Name:      "example rule",
				Condition: "cf.load_balancer.region == \"SAF\"",
				Overrides: LoadBalancerRuleOverrides{
					RegionPools: map[string][]string{
						"SAF": {"de90f38ced07c2e2f4df50b1f61d4194"},
					},
					AdaptiveRouting: &AdaptiveRouting{
						FailoverAcrossPools: BoolPtr(false),
					},
					LocationStrategy: &LocationStrategy{
						PreferECS: "never",
						Mode:      "pop",
					},
				},
			},
		},
		Proxied:        true,
		Persistence:    "cookie",
		PersistenceTTL: 5000,
		SessionAffinityAttributes: &SessionAffinityAttributes{
			SameSite:             "Strict",
			Secure:               "Always",
			DrainDuration:        60,
			ZeroDowntimeFailover: "sticky",
		},
	}

	actual, err := client.CreateLoadBalancer(context.Background(), ZoneIdentifier(testZoneID), CreateLoadBalancerParams{LoadBalancer: request})
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateLoadBalancer_AccountIsNotSupported(t *testing.T) {
	setup()
	defer teardown()

	_, err := client.CreateLoadBalancer(context.Background(), AccountIdentifier(testAccountID), CreateLoadBalancerParams{})
	if assert.Error(t, err) {
		assert.Equal(t, fmt.Sprintf(errInvalidResourceContainerAccess, AccountRouteLevel), err.Error())
	}
}

func TestListLoadBalancers(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
            "success": true,
            "errors": [],
            "messages": [],
            "result": [
                {
                    "id": "699d98642c564d2e855e9661899b7252",
                    "created_on": "2014-01-01T05:20:00.12345Z",
                    "modified_on": "2014-02-01T05:20:00.12345Z",
                    "description": "Load Balancer for www.example.com",
                    "name": "www.example.com",
                    "ttl": 30,
                    "fallback_pool": "17b5962d775c646f3f9725cbc7a53df4",
                    "default_pools": [
                      "de90f38ced07c2e2f4df50b1f61d4194",
                      "9290f38c5d07c2e2f4df57b1f61d4196",
                      "00920f38ce07c2e2f4df50b1f61d4194"
                    ],
                    "region_pools": {
                      "WNAM": [
                        "de90f38ced07c2e2f4df50b1f61d4194",
                        "9290f38c5d07c2e2f4df57b1f61d4196"
                      ],
                      "ENAM": [
                        "00920f38ce07c2e2f4df50b1f61d4194"
                      ]
                    },
                    "country_pools": {
                      "US": [
                        "de90f38ced07c2e2f4df50b1f61d4194"
                      ],
                      "GB": [
                        "abd90f38ced07c2e2f4df50b1f61d4194"
                      ]
                    },
                    "pop_pools": {
                      "LAX": [
                        "de90f38ced07c2e2f4df50b1f61d4194",
                        "9290f38c5d07c2e2f4df57b1f61d4196"
                      ],
                      "LHR": [
                        "abd90f38ced07c2e2f4df50b1f61d4194",
                        "f9138c5d07c2e2f4df57b1f61d4196"
                      ],
                      "SJC": [
                        "00920f38ce07c2e2f4df50b1f61d4194"
                      ]
                    },
                    "random_steering": {
                        "default_weight": 0.2,
                        "pool_weights": {
                            "9290f38c5d07c2e2f4df57b1f61d4196": 0.6,
                            "de90f38ced07c2e2f4df50b1f61d4194": 0.4
                        }
                    },
					"adaptive_routing": {
						"failover_across_pools": true
					},
					"location_strategy": {
						"prefer_ecs": "always",
						"mode": "resolver_ip"
					},
                    "proxied": true
                }
            ],
            "result_info": {
                "page": 1,
                "per_page": 20,
                "count": 1,
                "total_count": 1
            }
        }`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/load_balancers", handler)
	createdOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2014-02-01T05:20:00.12345Z")
	want := []LoadBalancer{
		{
			ID:           "699d98642c564d2e855e9661899b7252",
			CreatedOn:    &createdOn,
			ModifiedOn:   &modifiedOn,
			Description:  "Load Balancer for www.example.com",
			Name:         "www.example.com",
			TTL:          30,
			FallbackPool: "17b5962d775c646f3f9725cbc7a53df4",
			DefaultPools: []string{
				"de90f38ced07c2e2f4df50b1f61d4194",
				"9290f38c5d07c2e2f4df57b1f61d4196",
				"00920f38ce07c2e2f4df50b1f61d4194",
			},
			RegionPools: map[string][]string{
				"WNAM": {
					"de90f38ced07c2e2f4df50b1f61d4194",
					"9290f38c5d07c2e2f4df57b1f61d4196",
				},
				"ENAM": {
					"00920f38ce07c2e2f4df50b1f61d4194",
				},
			},
			CountryPools: map[string][]string{
				"US": {
					"de90f38ced07c2e2f4df50b1f61d4194",
				},
				"GB": {
					"abd90f38ced07c2e2f4df50b1f61d4194",
				},
			},
			PopPools: map[string][]string{
				"LAX": {
					"de90f38ced07c2e2f4df50b1f61d4194",
					"9290f38c5d07c2e2f4df57b1f61d4196",
				},
				"LHR": {
					"abd90f38ced07c2e2f4df50b1f61d4194",
					"f9138c5d07c2e2f4df57b1f61d4196",
				},
				"SJC": {
					"00920f38ce07c2e2f4df50b1f61d4194",
				},
			},
			RandomSteering: &RandomSteering{
				DefaultWeight: 0.2,
				PoolWeights: map[string]float64{
					"9290f38c5d07c2e2f4df57b1f61d4196": 0.6,
					"de90f38ced07c2e2f4df50b1f61d4194": 0.4,
				},
			},
			AdaptiveRouting: &AdaptiveRouting{
				FailoverAcrossPools: BoolPtr(true),
			},
			LocationStrategy: &LocationStrategy{
				PreferECS: "always",
				Mode:      "resolver_ip",
			},
			Proxied: true,
		},
	}

	actual, err := client.ListLoadBalancers(context.Background(), ZoneIdentifier(testZoneID), ListLoadBalancerParams{})
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestListLoadBalancer_AccountIsNotSupported(t *testing.T) {
	setup()
	defer teardown()

	_, err := client.ListLoadBalancers(context.Background(), AccountIdentifier(testAccountID), ListLoadBalancerParams{})
	if assert.Error(t, err) {
		assert.Equal(t, fmt.Sprintf(errInvalidResourceContainerAccess, AccountRouteLevel), err.Error())
	}
}

func TestGetLoadBalancer(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
            "success": true,
            "errors": [],
            "messages": [],
            "result": {
                "id": "699d98642c564d2e855e9661899b7252",
                "created_on": "2014-01-01T05:20:00.12345Z",
                "modified_on": "2014-02-01T05:20:00.12345Z",
                "description": "Load Balancer for www.example.com",
                "name": "www.example.com",
                "ttl": 30,
                "fallback_pool": "17b5962d775c646f3f9725cbc7a53df4",
                "default_pools": [
                  "de90f38ced07c2e2f4df50b1f61d4194",
                  "9290f38c5d07c2e2f4df57b1f61d4196",
                  "00920f38ce07c2e2f4df50b1f61d4194"
                ],
                "region_pools": {
                  "WNAM": [
                    "de90f38ced07c2e2f4df50b1f61d4194",
                    "9290f38c5d07c2e2f4df57b1f61d4196"
                  ],
                  "ENAM": [
                    "00920f38ce07c2e2f4df50b1f61d4194"
                  ]
                },
                "country_pools": {
                  "US": [
                    "de90f38ced07c2e2f4df50b1f61d4194"
                  ],
                  "GB": [
                    "abd90f38ced07c2e2f4df50b1f61d4194"
                  ]
                },
                "pop_pools": {
                  "LAX": [
                    "de90f38ced07c2e2f4df50b1f61d4194",
                    "9290f38c5d07c2e2f4df57b1f61d4196"
                  ],
                  "LHR": [
                    "abd90f38ced07c2e2f4df50b1f61d4194",
                    "f9138c5d07c2e2f4df57b1f61d4196"
                  ],
                  "SJC": [
                    "00920f38ce07c2e2f4df50b1f61d4194"
                  ]
                },
                "random_steering": {
                    "default_weight": 0.2,
                    "pool_weights": {
                        "9290f38c5d07c2e2f4df57b1f61d4196": 0.6,
                        "de90f38ced07c2e2f4df50b1f61d4194": 0.4
                    }
                },
				"adaptive_routing": {
					"failover_across_pools": true
				},
				"location_strategy": {
					"prefer_ecs": "always",
					"mode": "resolver_ip"
				},
                "proxied": true
            }
        }`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/load_balancers/699d98642c564d2e855e9661899b7252", handler)
	createdOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2014-02-01T05:20:00.12345Z")
	want := LoadBalancer{
		ID:           "699d98642c564d2e855e9661899b7252",
		CreatedOn:    &createdOn,
		ModifiedOn:   &modifiedOn,
		Description:  "Load Balancer for www.example.com",
		Name:         "www.example.com",
		TTL:          30,
		FallbackPool: "17b5962d775c646f3f9725cbc7a53df4",
		DefaultPools: []string{
			"de90f38ced07c2e2f4df50b1f61d4194",
			"9290f38c5d07c2e2f4df57b1f61d4196",
			"00920f38ce07c2e2f4df50b1f61d4194",
		},
		RegionPools: map[string][]string{
			"WNAM": {
				"de90f38ced07c2e2f4df50b1f61d4194",
				"9290f38c5d07c2e2f4df57b1f61d4196",
			},
			"ENAM": {
				"00920f38ce07c2e2f4df50b1f61d4194",
			},
		},
		CountryPools: map[string][]string{
			"US": {
				"de90f38ced07c2e2f4df50b1f61d4194",
			},
			"GB": {
				"abd90f38ced07c2e2f4df50b1f61d4194",
			},
		},
		PopPools: map[string][]string{
			"LAX": {
				"de90f38ced07c2e2f4df50b1f61d4194",
				"9290f38c5d07c2e2f4df57b1f61d4196",
			},
			"LHR": {
				"abd90f38ced07c2e2f4df50b1f61d4194",
				"f9138c5d07c2e2f4df57b1f61d4196",
			},
			"SJC": {
				"00920f38ce07c2e2f4df50b1f61d4194",
			},
		},
		RandomSteering: &RandomSteering{
			DefaultWeight: 0.2,
			PoolWeights: map[string]float64{
				"9290f38c5d07c2e2f4df57b1f61d4196": 0.6,
				"de90f38ced07c2e2f4df50b1f61d4194": 0.4,
			},
		},
		AdaptiveRouting: &AdaptiveRouting{
			FailoverAcrossPools: BoolPtr(true),
		},
		LocationStrategy: &LocationStrategy{
			PreferECS: "always",
			Mode:      "resolver_ip",
		},
		Proxied: true,
	}

	actual, err := client.GetLoadBalancer(context.Background(), ZoneIdentifier(testZoneID), "699d98642c564d2e855e9661899b7252")
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}

	_, err = client.GetLoadBalancer(context.Background(), ZoneIdentifier(testZoneID), "bar")
	assert.Error(t, err)
}

func TestGetLoadBalancer_AccountIsNotSupported(t *testing.T) {
	setup()
	defer teardown()

	_, err := client.GetLoadBalancer(context.Background(), AccountIdentifier(testAccountID), "foo")
	if assert.Error(t, err) {
		assert.Equal(t, fmt.Sprintf(errInvalidResourceContainerAccess, AccountRouteLevel), err.Error())
	}
}

func TestDeleteLoadBalancer(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
            "success": true,
            "errors": [],
            "messages": [],
            "result": {
              "id": "699d98642c564d2e855e9661899b7252"
            }
        }`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/load_balancers/699d98642c564d2e855e9661899b7252", handler)
	assert.NoError(t, client.DeleteLoadBalancer(context.Background(), ZoneIdentifier(testZoneID), "699d98642c564d2e855e9661899b7252"))
	assert.Error(t, client.DeleteLoadBalancer(context.Background(), ZoneIdentifier(testZoneID), "bar"))
}

func TestDeleteLoadBalancer_AccountIsNotSupported(t *testing.T) {
	setup()
	defer teardown()

	err := client.DeleteLoadBalancer(context.Background(), AccountIdentifier(testAccountID), "foo")
	if assert.Error(t, err) {
		assert.Equal(t, fmt.Sprintf(errInvalidResourceContainerAccess, AccountRouteLevel), err.Error())
	}
}

func TestUpdateLoadBalancer(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		b, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if assert.NoError(t, err) {
			assert.JSONEq(t, `{
                "id": "699d98642c564d2e855e9661899b7252",
                "description": "Load Balancer for www.example.com",
                "name": "www.example.com",
                "ttl": 30,
                "fallback_pool": "17b5962d775c646f3f9725cbc7a53df4",
                "default_pools": [
                  "00920f38ce07c2e2f4df50b1f61d4194"
                ],
                "region_pools": {
                  "WNAM": [
                    "9290f38c5d07c2e2f4df57b1f61d4196"
                  ],
                  "ENAM": [
                    "00920f38ce07c2e2f4df50b1f61d4194"
                  ]
                },
                "country_pools": {
                  "US": [
                    "de90f38ced07c2e2f4df50b1f61d4194"
                  ],
                  "GB": [
                    "f9138c5d07c2e2f4df57b1f61d4196"
                  ]
                },
                "pop_pools": {
                  "LAX": [
                    "9290f38c5d07c2e2f4df57b1f61d4196"
                  ],
                  "LHR": [
                    "f9138c5d07c2e2f4df57b1f61d4196"
                  ],
                  "SJC": [
                    "00920f38ce07c2e2f4df50b1f61d4194"
                  ]
                },
                "random_steering": {
                    "default_weight": 0.5,
                    "pool_weights": {
                        "9290f38c5d07c2e2f4df57b1f61d4196": 0.2
                    }
                },
				"adaptive_routing": {
					"failover_across_pools": false
				},
				"location_strategy": {
					"prefer_ecs": "never",
					"mode": "pop"
				},
                "proxied": true,
                "session_affinity": "none",
                "session_affinity_attributes": {
                  "samesite": "Strict",
                  "secure": "Always",
				  "zero_downtime_failover": "sticky"
                }
			}`, string(b))
		}
		fmt.Fprint(w, `{
            "success": true,
            "errors": [],
            "messages": [],
            "result": {
                "id": "699d98642c564d2e855e9661899b7252",
                "created_on": "2014-01-01T05:20:00.12345Z",
                "modified_on": "2017-02-01T05:20:00.12345Z",
                "description": "Load Balancer for www.example.com",
                "name": "www.example.com",
                "ttl": 30,
                "fallback_pool": "17b5962d775c646f3f9725cbc7a53df4",
                "default_pools": [
                  "00920f38ce07c2e2f4df50b1f61d4194"
                ],
                "region_pools": {
                  "WNAM": [
                    "9290f38c5d07c2e2f4df57b1f61d4196"
                  ],
                  "ENAM": [
                    "00920f38ce07c2e2f4df50b1f61d4194"
                  ]
                },
                "country_pools": {
                  "US": [
                    "de90f38ced07c2e2f4df50b1f61d4194"
                  ],
                  "GB": [
                    "f9138c5d07c2e2f4df57b1f61d4196"
                  ]
                },
                "pop_pools": {
                  "LAX": [
                    "9290f38c5d07c2e2f4df57b1f61d4196"
                  ],
                  "LHR": [
                    "f9138c5d07c2e2f4df57b1f61d4196"
                  ],
                  "SJC": [
                    "00920f38ce07c2e2f4df50b1f61d4194"
                  ]
                },
                "random_steering": {
                    "default_weight": 0.5,
                    "pool_weights": {
                        "9290f38c5d07c2e2f4df57b1f61d4196": 0.2
                    }
                },
				"adaptive_routing": {
					"failover_across_pools": false
				},
				"location_strategy": {
					"prefer_ecs": "never",
					"mode": "pop"
				},
                "proxied": true,
                "session_affinity": "none",
                "session_affinity_attributes": {
                  "samesite": "Strict",
                  "secure": "Always",
	              "zero_downtime_failover": "sticky"
                }
            }
        }`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/load_balancers/699d98642c564d2e855e9661899b7252", handler)
	createdOn, _ := time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	modifiedOn, _ := time.Parse(time.RFC3339, "2017-02-01T05:20:00.12345Z")
	want := LoadBalancer{
		ID:           "699d98642c564d2e855e9661899b7252",
		CreatedOn:    &createdOn,
		ModifiedOn:   &modifiedOn,
		Description:  "Load Balancer for www.example.com",
		Name:         "www.example.com",
		TTL:          30,
		FallbackPool: "17b5962d775c646f3f9725cbc7a53df4",
		DefaultPools: []string{
			"00920f38ce07c2e2f4df50b1f61d4194",
		},
		RegionPools: map[string][]string{
			"WNAM": {
				"9290f38c5d07c2e2f4df57b1f61d4196",
			},
			"ENAM": {
				"00920f38ce07c2e2f4df50b1f61d4194",
			},
		},
		CountryPools: map[string][]string{
			"US": {
				"de90f38ced07c2e2f4df50b1f61d4194",
			},
			"GB": {
				"f9138c5d07c2e2f4df57b1f61d4196",
			},
		},
		PopPools: map[string][]string{
			"LAX": {
				"9290f38c5d07c2e2f4df57b1f61d4196",
			},
			"LHR": {
				"f9138c5d07c2e2f4df57b1f61d4196",
			},
			"SJC": {
				"00920f38ce07c2e2f4df50b1f61d4194",
			},
		},
		RandomSteering: &RandomSteering{
			DefaultWeight: 0.5,
			PoolWeights: map[string]float64{
				"9290f38c5d07c2e2f4df57b1f61d4196": 0.2,
			},
		},
		AdaptiveRouting: &AdaptiveRouting{
			FailoverAcrossPools: BoolPtr(false),
		},
		LocationStrategy: &LocationStrategy{
			PreferECS: "never",
			Mode:      "pop",
		},
		Proxied:     true,
		Persistence: "none",
		SessionAffinityAttributes: &SessionAffinityAttributes{
			SameSite:             "Strict",
			Secure:               "Always",
			ZeroDowntimeFailover: "sticky",
		},
	}
	request := LoadBalancer{
		ID:           "699d98642c564d2e855e9661899b7252",
		Description:  "Load Balancer for www.example.com",
		Name:         "www.example.com",
		TTL:          30,
		FallbackPool: "17b5962d775c646f3f9725cbc7a53df4",
		DefaultPools: []string{
			"00920f38ce07c2e2f4df50b1f61d4194",
		},
		RegionPools: map[string][]string{
			"WNAM": {
				"9290f38c5d07c2e2f4df57b1f61d4196",
			},
			"ENAM": {
				"00920f38ce07c2e2f4df50b1f61d4194",
			},
		},
		CountryPools: map[string][]string{
			"US": {
				"de90f38ced07c2e2f4df50b1f61d4194",
			},
			"GB": {
				"f9138c5d07c2e2f4df57b1f61d4196",
			},
		},
		PopPools: map[string][]string{
			"LAX": {
				"9290f38c5d07c2e2f4df57b1f61d4196",
			},
			"LHR": {
				"f9138c5d07c2e2f4df57b1f61d4196",
			},
			"SJC": {
				"00920f38ce07c2e2f4df50b1f61d4194",
			},
		},
		RandomSteering: &RandomSteering{
			DefaultWeight: 0.5,
			PoolWeights: map[string]float64{
				"9290f38c5d07c2e2f4df57b1f61d4196": 0.2,
			},
		},
		AdaptiveRouting: &AdaptiveRouting{
			FailoverAcrossPools: BoolPtr(false),
		},
		LocationStrategy: &LocationStrategy{
			PreferECS: "never",
			Mode:      "pop",
		},
		Proxied:     true,
		Persistence: "none",
		SessionAffinityAttributes: &SessionAffinityAttributes{
			SameSite:             "Strict",
			Secure:               "Always",
			ZeroDowntimeFailover: "sticky",
		},
	}

	actual, err := client.UpdateLoadBalancer(context.Background(), ZoneIdentifier(testZoneID), UpdateLoadBalancerParams{LoadBalancer: request})
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestUpdateLoadBalancer_AccountIsNotSupported(t *testing.T) {
	setup()
	defer teardown()

	_, err := client.UpdateLoadBalancer(context.Background(), AccountIdentifier(testAccountID), UpdateLoadBalancerParams{LoadBalancer: LoadBalancer{}})
	if assert.Error(t, err) {
		assert.Equal(t, fmt.Sprintf(errInvalidResourceContainerAccess, AccountRouteLevel), err.Error())
	}
}

func TestLoadBalancerPoolHealthDetails(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
            "success": true,
            "errors": [],
            "messages": [],
            "result": {
                "pool_id": "699d98642c564d2e855e9661899b7252",
                "pop_health": {
                    "Amsterdam, NL": {
                        "healthy": true,
                        "origins": [
                          {
                            "2001:DB8::5": {
                                "healthy": true,
                                "rtt": "12.1ms",
                                "failure_reason": "No failures",
                                "response_code": 401
                            }
                          }
                        ]
                    }
                }
            }
        }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/load_balancers/pools/699d98642c564d2e855e9661899b7252/health", handler)
	want := LoadBalancerPoolHealth{
		ID: "699d98642c564d2e855e9661899b7252",
		PopHealth: map[string]LoadBalancerPoolPopHealth{
			"Amsterdam, NL": {
				Healthy: true,
				Origins: []map[string]LoadBalancerOriginHealth{
					{
						"2001:DB8::5": {
							Healthy:       true,
							RTT:           Duration{12*time.Millisecond + 100*time.Microsecond},
							FailureReason: "No failures",
							ResponseCode:  401,
						},
					},
				},
			},
		},
	}

	actual, err := client.GetLoadBalancerPoolHealth(context.Background(), AccountIdentifier(testAccountID), "699d98642c564d2e855e9661899b7252")
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestLoadBalancerPoolHealthDetails_ZoneIsNotSupported(t *testing.T) {
	setup()
	defer teardown()

	_, err := client.GetLoadBalancerPoolHealth(context.Background(), ZoneIdentifier(testZoneID), "foo")
	if assert.Error(t, err) {
		assert.Equal(t, fmt.Sprintf(errInvalidResourceContainerAccess, ZoneRouteLevel), err.Error())
	}
}
