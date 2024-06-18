package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSecondaryDNSPrimary(t *testing.T) {
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
				"id": "23ff594956f20c2a721606e94745a8aa",
				"ip": "192.0.2.53",
				"port": 53,
				"ixfr_enable": false,
				"tsig_id": "69cd1e104af3e6ed3cb344f263fd0d5a",
				"name": "my-primary-1"
			}
		}
		`)
	}

	mux.HandleFunc("/accounts/01a7362d577a6c3019a474fd6f485823/secondary_dns/primaries/23ff594956f20c2a721606e94745a8aa", handler)
	want := SecondaryDNSPrimary{
		ID:         "23ff594956f20c2a721606e94745a8aa",
		IP:         "192.0.2.53",
		Port:       53,
		IxfrEnable: false,
		TsigID:     "69cd1e104af3e6ed3cb344f263fd0d5a",
		Name:       "my-primary-1",
	}

	actual, err := client.GetSecondaryDNSPrimary(context.Background(), "01a7362d577a6c3019a474fd6f485823", "23ff594956f20c2a721606e94745a8aa")
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestListSecondaryDNSPrimaries(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": [{
				"id": "23ff594956f20c2a721606e94745a8aa",
				"ip": "192.0.2.53",
				"port": 53,
				"ixfr_enable": false,
				"tsig_id": "69cd1e104af3e6ed3cb344f263fd0d5a",
				"name": "my-primary-1"
			}]
		}
		`)
	}

	mux.HandleFunc("/accounts/01a7362d577a6c3019a474fd6f485823/secondary_dns/primaries", handler)
	want := []SecondaryDNSPrimary{{
		ID:         "23ff594956f20c2a721606e94745a8aa",
		IP:         "192.0.2.53",
		Port:       53,
		IxfrEnable: false,
		TsigID:     "69cd1e104af3e6ed3cb344f263fd0d5a",
		Name:       "my-primary-1",
	}}

	actual, err := client.ListSecondaryDNSPrimaries(context.Background(), "01a7362d577a6c3019a474fd6f485823")
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateSecondaryDNSPrimary(t *testing.T) {
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
				"id": "23ff594956f20c2a721606e94745a8aa",
				"ip": "192.0.2.53",
				"port": 53,
				"ixfr_enable": false,
				"tsig_id": "69cd1e104af3e6ed3cb344f263fd0d5a",
				"name": "my-primary-1"
			}
		}
		`)
	}

	mux.HandleFunc("/accounts/01a7362d577a6c3019a474fd6f485823/secondary_dns/primaries", handler)
	want := SecondaryDNSPrimary{
		ID:         "23ff594956f20c2a721606e94745a8aa",
		IP:         "192.0.2.53",
		Port:       53,
		IxfrEnable: false,
		TsigID:     "69cd1e104af3e6ed3cb344f263fd0d5a",
		Name:       "my-primary-1",
	}

	actual, err := client.CreateSecondaryDNSPrimary(context.Background(), "01a7362d577a6c3019a474fd6f485823", want)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestUpdateSecondaryDNSPrimary(t *testing.T) {
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
				"id": "23ff594956f20c2a721606e94745a8aa",
				"ip": "192.0.2.53",
				"port": 53,
				"ixfr_enable": false,
				"tsig_id": "69cd1e104af3e6ed3cb344f263fd0d5a",
				"name": "my-primary-1"
			}
		}
		`)
	}

	mux.HandleFunc("/accounts/01a7362d577a6c3019a474fd6f485823/secondary_dns/primaries/23ff594956f20c2a721606e94745a8aa", handler)
	want := SecondaryDNSPrimary{
		ID:         "23ff594956f20c2a721606e94745a8aa",
		IP:         "192.0.2.53",
		Port:       53,
		IxfrEnable: false,
		TsigID:     "69cd1e104af3e6ed3cb344f263fd0d5a",
		Name:       "my-primary-1",
	}

	actual, err := client.UpdateSecondaryDNSPrimary(context.Background(), "01a7362d577a6c3019a474fd6f485823", want)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestDeleteSecondaryDNSPrimary(t *testing.T) {
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
				"id": "23ff594956f20c2a721606e94745a8aa"
			}
		}
		`)
	}

	mux.HandleFunc("/zones/01a7362d577a6c3019a474fd6f485823/secondary_dns/primaries/23ff594956f20c2a721606e94745a8aa", handler)

	err := client.DeleteSecondaryDNSPrimary(context.Background(), "01a7362d577a6c3019a474fd6f485823", "23ff594956f20c2a721606e94745a8aa")
	assert.NoError(t, err)
}

func TestValidateRequiredSecondaryDNSPrimaries(t *testing.T) {
	p1 := SecondaryDNSPrimary{}
	err1 := validateRequiredSecondaryDNSPrimaries(p1)
	assert.EqualError(t, err1, errSecondaryDNSInvalidPrimaryIP)

	p2 := SecondaryDNSPrimary{IP: "192.0.2.53"}
	err2 := validateRequiredSecondaryDNSPrimaries(p2)
	assert.EqualError(t, err2, errSecondaryDNSInvalidPrimaryPort)

	p3 := SecondaryDNSPrimary{IP: "192.0.2.53", Port: 53}
	err3 := validateRequiredSecondaryDNSPrimaries(p3)
	assert.NoError(t, err3)
}
