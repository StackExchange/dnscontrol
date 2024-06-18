package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testHyperdriveConfigId   = "6b7efc370ea34ded8327fa20698dfe3a"
	testHyperdriveConfigName = "example-hyperdrive"
)

func testHyperdriveConfig() HyperdriveConfig {
	return HyperdriveConfig{
		ID:   testHyperdriveConfigId,
		Name: testHyperdriveConfigName,
		Origin: HyperdriveConfigOrigin{
			Database: "postgres",
			Host:     "database.example.com",
			Port:     5432,
			Scheme:   "postgres",
			User:     "postgres",
		},
		Caching: HyperdriveConfigCaching{
			Disabled:             BoolPtr(false),
			MaxAge:               30,
			StaleWhileRevalidate: 15,
		},
	}
}

func TestHyperdriveConfig_List(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/hyperdrive/configs", testAccountID), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": [{
				"id": "6b7efc370ea34ded8327fa20698dfe3a",
				"caching": {
					"disabled": false,
					"max_age": 30,
					"stale_while_revalidate": 15
				},
				"name": "example-hyperdrive",
				"origin": {
					"database": "postgres",
					"host": "database.example.com",
					"port": 5432,
					"scheme": "postgres",
					"user": "postgres"
				}
			}]
		}`)
	})

	_, err := client.ListHyperdriveConfigs(context.Background(), AccountIdentifier(""), ListHyperdriveConfigParams{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	result, err := client.ListHyperdriveConfigs(context.Background(), AccountIdentifier(testAccountID), ListHyperdriveConfigParams{})
	if assert.NoError(t, err) {
		assert.Equal(t, 1, len(result))
		assert.Equal(t, testHyperdriveConfig(), result[0])
	}
}

func TestHyperdriveConfig_Get(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/hyperdrive/configs/%s", testAccountID, testHyperdriveConfigId), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "6b7efc370ea34ded8327fa20698dfe3a",
				"caching": {
					"disabled": false,
					"max_age": 30,
					"stale_while_revalidate": 15
				},
				"name": "example-hyperdrive",
				"origin": {
					"database": "postgres",
					"host": "database.example.com",
					"port": 5432,
					"scheme": "postgres",
					"user": "postgres"
				}
			}
		}`)
	})

	_, err := client.GetHyperdriveConfig(context.Background(), AccountIdentifier(""), "")
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	_, err = client.GetHyperdriveConfig(context.Background(), AccountIdentifier(testAccountID), "")
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingHyperdriveConfigID, err)
	}

	result, err := client.GetHyperdriveConfig(context.Background(), AccountIdentifier(testAccountID), testHyperdriveConfigId)
	if assert.NoError(t, err) {
		assert.Equal(t, testHyperdriveConfig(), result)
	}
}

func TestHyperdriveConfig_Create(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/hyperdrive/configs", testAccountID), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "6b7efc370ea34ded8327fa20698dfe3a",
				"caching": {
					"disabled": false,
					"max_age": 30,
					"stale_while_revalidate": 15
				},
				"name": "example-hyperdrive",
				"origin": {
					"database": "postgres",
					"host": "database.example.com",
					"port": 5432,
					"scheme": "postgres",
					"user": "postgres"
				}
			}
		}`)
	})

	_, err := client.CreateHyperdriveConfig(context.Background(), AccountIdentifier(""), CreateHyperdriveConfigParams{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	_, err = client.CreateHyperdriveConfig(context.Background(), AccountIdentifier(testAccountID), CreateHyperdriveConfigParams{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingHyperdriveConfigName, err)
	}

	result, err := client.CreateHyperdriveConfig(context.Background(), AccountIdentifier(testAccountID), CreateHyperdriveConfigParams{
		Name: "example-hyperdrive",
		Origin: HyperdriveConfigOrigin{
			Database: "postgres",
			Password: "password",
			Host:     "database.example.com",
			Port:     5432,
			Scheme:   "postgres",
			User:     "postgres",
		},
		Caching: HyperdriveConfigCaching{
			Disabled:             BoolPtr(false),
			MaxAge:               30,
			StaleWhileRevalidate: 15,
		},
	})

	if assert.NoError(t, err) {
		assert.Equal(t, testHyperdriveConfig(), result)
	}
}

func TestHyperdriveConfig_Delete(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/hyperdrive/configs/%s", testAccountID, testHyperdriveConfigId), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": null
		}`)
	})
	err := client.DeleteHyperdriveConfig(context.Background(), AccountIdentifier(""), "")
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	err = client.DeleteHyperdriveConfig(context.Background(), AccountIdentifier(testAccountID), "")
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingHyperdriveConfigID, err)
	}

	err = client.DeleteHyperdriveConfig(context.Background(), AccountIdentifier(testAccountID), testHyperdriveConfigId)
	assert.NoError(t, err)
}

func TestHyperdriveConfig_Update(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/accounts/%s/hyperdrive/configs/%s", testAccountID, testHyperdriveConfigId), func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "6b7efc370ea34ded8327fa20698dfe3a",
				"caching": {
					"disabled": false,
					"max_age": 30,
					"stale_while_revalidate": 15
				},
				"name": "example-hyperdrive",
				"origin": {
					"database": "postgres",
					"host": "database.example.com",
					"port": 5432,
					"scheme": "postgres",
					"user": "postgres"
				}
			}
		}`)
	})

	_, err := client.UpdateHyperdriveConfig(context.Background(), AccountIdentifier(""), UpdateHyperdriveConfigParams{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	_, err = client.UpdateHyperdriveConfig(context.Background(), AccountIdentifier(testAccountID), UpdateHyperdriveConfigParams{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingHyperdriveConfigID, err)
	}

	result, err := client.UpdateHyperdriveConfig(context.Background(), AccountIdentifier(testAccountID), UpdateHyperdriveConfigParams{
		HyperdriveID: "6b7efc370ea34ded8327fa20698dfe3a",
		Name:         "example-hyperdrive",
		Origin: HyperdriveConfigOrigin{
			Database: "postgres",
			Password: "password",
			Host:     "database.example.com",
			Port:     5432,
			Scheme:   "postgres",
			User:     "postgres",
		},
		Caching: HyperdriveConfigCaching{
			Disabled:             BoolPtr(false),
			MaxAge:               30,
			StaleWhileRevalidate: 15,
		},
	})

	if assert.NoError(t, err) {
		assert.Equal(t, testHyperdriveConfig(), result)
	}
}
