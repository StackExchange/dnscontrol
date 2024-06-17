package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSecondaryDNSTSIG(t *testing.T) {
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
				"id": "69cd1e104af3e6ed3cb344f263fd0d5a",
				"name": "tsig.customer.cf.",
				"secret": "caf79a7804b04337c9c66ccd7bef9190a1e1679b5dd03d8aa10f7ad45e1a9dab92b417896c15d4d007c7c14194538d2a5d0feffdecc5a7f0e1c570cfa700837c",
				"algo": "hmac-sha512."
			}
		}
		`)
	}

	mux.HandleFunc("/accounts/01a7362d577a6c3019a474fd6f485823/secondary_dns/tsigs/69cd1e104af3e6ed3cb344f263fd0d5a", handler)

	want := SecondaryDNSTSIG{
		ID:     "69cd1e104af3e6ed3cb344f263fd0d5a",
		Name:   "tsig.customer.cf.",
		Secret: "caf79a7804b04337c9c66ccd7bef9190a1e1679b5dd03d8aa10f7ad45e1a9dab92b417896c15d4d007c7c14194538d2a5d0feffdecc5a7f0e1c570cfa700837c",
		Algo:   "hmac-sha512.",
	}

	actual, err := client.GetSecondaryDNSTSIG(context.Background(), "01a7362d577a6c3019a474fd6f485823", "69cd1e104af3e6ed3cb344f263fd0d5a")
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestListSecondaryDNSTSIGs(t *testing.T) {
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
					"id": "69cd1e104af3e6ed3cb344f263fd0d5a",
					"name": "tsig.customer.cf.",
					"secret": "caf79a7804b04337c9c66ccd7bef9190a1e1679b5dd03d8aa10f7ad45e1a9dab92b417896c15d4d007c7c14194538d2a5d0feffdecc5a7f0e1c570cfa700837c",
					"algo": "hmac-sha512."
				}
			]
		}
		`)
	}

	mux.HandleFunc("/accounts/01a7362d577a6c3019a474fd6f485823/secondary_dns/tsigs", handler)

	want := []SecondaryDNSTSIG{{
		ID:     "69cd1e104af3e6ed3cb344f263fd0d5a",
		Name:   "tsig.customer.cf.",
		Secret: "caf79a7804b04337c9c66ccd7bef9190a1e1679b5dd03d8aa10f7ad45e1a9dab92b417896c15d4d007c7c14194538d2a5d0feffdecc5a7f0e1c570cfa700837c",
		Algo:   "hmac-sha512.",
	}}

	actual, err := client.ListSecondaryDNSTSIGs(context.Background(), "01a7362d577a6c3019a474fd6f485823")
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateSecondaryDNSTSIG(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "69cd1e104af3e6ed3cb344f263fd0d5a",
				"name": "tsig.customer.cf.",
				"secret": "caf79a7804b04337c9c66ccd7bef9190a1e1679b5dd03d8aa10f7ad45e1a9dab92b417896c15d4d007c7c14194538d2a5d0feffdecc5a7f0e1c570cfa700837c",
				"algo": "hmac-sha512."
			}
		}
		`)
	}

	mux.HandleFunc("/accounts/01a7362d577a6c3019a474fd6f485823/secondary_dns/tsigs", handler)

	want := SecondaryDNSTSIG{
		ID:     "69cd1e104af3e6ed3cb344f263fd0d5a",
		Name:   "tsig.customer.cf.",
		Secret: "caf79a7804b04337c9c66ccd7bef9190a1e1679b5dd03d8aa10f7ad45e1a9dab92b417896c15d4d007c7c14194538d2a5d0feffdecc5a7f0e1c570cfa700837c",
		Algo:   "hmac-sha512.",
	}

	actual, err := client.CreateSecondaryDNSTSIG(context.Background(), "01a7362d577a6c3019a474fd6f485823", SecondaryDNSTSIG{
		Name:   "tsig.customer.cf.",
		Secret: "caf79a7804b04337c9c66ccd7bef9190a1e1679b5dd03d8aa10f7ad45e1a9dab92b417896c15d4d007c7c14194538d2a5d0feffdecc5a7f0e1c570cfa700837c",
		Algo:   "hmac-sha512.",
	})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestUpdateSecondaryDNSTSIG(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "69cd1e104af3e6ed3cb344f263fd0d5a",
				"name": "tsig.customer.cf.",
				"secret": "caf79a7804b04337c9c66ccd7bef9190a1e1679b5dd03d8aa10f7ad45e1a9dab92b417896c15d4d007c7c14194538d2a5d0feffdecc5a7f0e1c570cfa700837c",
				"algo": "hmac-sha512."
			}
		}
		`)
	}

	mux.HandleFunc("/accounts/01a7362d577a6c3019a474fd6f485823/secondary_dns/tsigs/69cd1e104af3e6ed3cb344f263fd0d5a", handler)

	want := SecondaryDNSTSIG{
		ID:     "69cd1e104af3e6ed3cb344f263fd0d5a",
		Name:   "tsig.customer.cf.",
		Secret: "caf79a7804b04337c9c66ccd7bef9190a1e1679b5dd03d8aa10f7ad45e1a9dab92b417896c15d4d007c7c14194538d2a5d0feffdecc5a7f0e1c570cfa700837c",
		Algo:   "hmac-sha512.",
	}

	actual, err := client.UpdateSecondaryDNSTSIG(context.Background(), "01a7362d577a6c3019a474fd6f485823", want)

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestDeleteSecondaryDNSTSIG(t *testing.T) {
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
				"id": "269d8f4853475ca241c4e730be286b20"
			}
		}
		`)
	}

	mux.HandleFunc("/accounts/01a7362d577a6c3019a474fd6f485823/secondary_dns/tsigs/69cd1e104af3e6ed3cb344f263fd0d5a", handler)

	err := client.DeleteSecondaryDNSTSIG(context.Background(), "01a7362d577a6c3019a474fd6f485823", "69cd1e104af3e6ed3cb344f263fd0d5a")
	assert.NoError(t, err)
}
