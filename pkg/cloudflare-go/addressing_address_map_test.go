package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	addressMapDesc       = "My Ecommerce zones"
	addressMapDefaultSNI = "*.example.com"
)

func TestListAddressMap(t *testing.T) {
	setup()
	defer teardown()

	expectedIP := "127.0.0.1"
	expectedCIDR := "127.0.0.1/24"

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)

		actualIP := r.URL.Query().Get("ip")
		assert.Equal(t, expectedIP, actualIP, "Expected ip %q, got %q", expectedIP, actualIP)

		actualCIDR := r.URL.Query().Get("cidr")
		assert.Equal(t, expectedCIDR, actualCIDR, "Expected cidr %q, got %q", expectedCIDR, actualCIDR)

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": [{
			  "id": "9a7806061c88ada191ed06f989cc3dac",
			  "description": "My Ecommerce zones",
			  "can_delete": true,
			  "can_modify_ips": true,
			  "default_sni": "*.example.com",
			  "created_at": "2023-01-01T05:20:00.12345Z",
			  "modified_at": "2023-01-05T05:20:00.12345Z",
			  "enabled": true,
			  "ips": [
				{
				  "ip": "192.0.2.1",
				  "created_at": "2023-01-02T05:20:00.12345Z"
				}
			  ],
			  "memberships": [
				{
				  "kind": "zone",
				  "identifier": "01a7362d577a6c3019a474fd6f485823",
				  "can_delete": true,
				  "created_at": "2023-01-03T05:20:00.12345Z"
				}
			  ]
			}]
		  }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/addressing/address_maps", handler)

	createdAt, _ := time.Parse(time.RFC3339, "2023-01-01T05:20:00.12345Z")
	modifiedAt, _ := time.Parse(time.RFC3339, "2023-01-05T05:20:00.12345Z")
	ipCreatedAt, _ := time.Parse(time.RFC3339, "2023-01-02T05:20:00.12345Z")
	membershipCreatedAt, _ := time.Parse(time.RFC3339, "2023-01-03T05:20:00.12345Z")

	want := []AddressMap{
		{
			ID:           "9a7806061c88ada191ed06f989cc3dac",
			CreatedAt:    createdAt,
			ModifiedAt:   modifiedAt,
			Description:  &addressMapDesc,
			Deletable:    BoolPtr(true),
			CanModifyIPs: BoolPtr(true),
			DefaultSNI:   &addressMapDefaultSNI,
			Enabled:      BoolPtr(true),
			IPs:          []AddressMapIP{{"192.0.2.1", ipCreatedAt}},
			Memberships: []AddressMapMembership{{
				Identifier: "01a7362d577a6c3019a474fd6f485823",
				Kind:       AddressMapMembershipZone,
				Deletable:  BoolPtr(true),
				CreatedAt:  membershipCreatedAt,
			}},
		},
	}

	actual, err := client.ListAddressMaps(context.Background(), AccountIdentifier(testAccountID), ListAddressMapsParams{&expectedIP, &expectedCIDR})
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestGetAddressMap(t *testing.T) {
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
			  "id": "9a7806061c88ada191ed06f989cc3dac",
			  "description": "My Ecommerce zones",
			  "can_delete": true,
			  "can_modify_ips": true,
			  "default_sni": "*.example.com",
			  "created_at": "2023-01-01T05:20:00.12345Z",
			  "modified_at": "2023-01-05T05:20:00.12345Z",
			  "enabled": true,
			  "ips": [
				{
				  "ip": "192.0.2.1",
				  "created_at": "2023-01-02T05:20:00.12345Z"
				}
			  ],
			  "memberships": [
				{
				  "kind": "zone",
				  "identifier": "01a7362d577a6c3019a474fd6f485823",
				  "can_delete": true,
				  "created_at": "2023-01-03T05:20:00.12345Z"
				}
			  ]
			}
		  }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/addressing/address_maps/9a7806061c88ada191ed06f989cc3dac", handler)

	createdAt, _ := time.Parse(time.RFC3339, "2023-01-01T05:20:00.12345Z")
	modifiedAt, _ := time.Parse(time.RFC3339, "2023-01-05T05:20:00.12345Z")
	ipCreatedAt, _ := time.Parse(time.RFC3339, "2023-01-02T05:20:00.12345Z")
	membershipCreatedAt, _ := time.Parse(time.RFC3339, "2023-01-03T05:20:00.12345Z")

	want := AddressMap{
		ID:           "9a7806061c88ada191ed06f989cc3dac",
		CreatedAt:    createdAt,
		ModifiedAt:   modifiedAt,
		Description:  &addressMapDesc,
		Deletable:    BoolPtr(true),
		CanModifyIPs: BoolPtr(true),
		DefaultSNI:   &addressMapDefaultSNI,
		Enabled:      BoolPtr(true),
		IPs:          []AddressMapIP{{"192.0.2.1", ipCreatedAt}},
		Memberships: []AddressMapMembership{{
			Identifier: "01a7362d577a6c3019a474fd6f485823",
			Kind:       AddressMapMembershipZone,
			Deletable:  BoolPtr(true),
			CreatedAt:  membershipCreatedAt,
		}},
	}

	actual, err := client.GetAddressMap(context.Background(), AccountIdentifier(testAccountID), "9a7806061c88ada191ed06f989cc3dac")
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestUpdateAddressMap(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
			  "id": "9a7806061c88ada191ed06f989cc3dac",
			  "description": "My Ecommerce zones",
			  "can_delete": true,
			  "can_modify_ips": true,
			  "default_sni": "*.example.com",
			  "created_at": "2023-01-01T05:20:00.12345Z",
			  "modified_at": "2023-01-05T05:20:00.12345Z",
			  "enabled": true,
			  "ips": [
				{
				  "ip": "192.0.2.1",
				  "created_at": "2023-01-02T05:20:00.12345Z"
				}
			  ],
			  "memberships": [
				{
				  "kind": "zone",
				  "identifier": "01a7362d577a6c3019a474fd6f485823",
				  "can_delete": true,
				  "created_at": "2023-01-03T05:20:00.12345Z"
				}
			  ]
			}
		  }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/addressing/address_maps/9a7806061c88ada191ed06f989cc3dac", handler)

	createdAt, _ := time.Parse(time.RFC3339, "2023-01-01T05:20:00.12345Z")
	modifiedAt, _ := time.Parse(time.RFC3339, "2023-01-05T05:20:00.12345Z")
	ipCreatedAt, _ := time.Parse(time.RFC3339, "2023-01-02T05:20:00.12345Z")
	membershipCreatedAt, _ := time.Parse(time.RFC3339, "2023-01-03T05:20:00.12345Z")

	want := AddressMap{
		ID:           "9a7806061c88ada191ed06f989cc3dac",
		CreatedAt:    createdAt,
		ModifiedAt:   modifiedAt,
		Description:  &addressMapDesc,
		Deletable:    BoolPtr(true),
		CanModifyIPs: BoolPtr(true),
		DefaultSNI:   &addressMapDefaultSNI,
		Enabled:      BoolPtr(true),
		IPs:          []AddressMapIP{{"192.0.2.1", ipCreatedAt}},
		Memberships: []AddressMapMembership{{
			Identifier: "01a7362d577a6c3019a474fd6f485823",
			Kind:       AddressMapMembershipZone,
			Deletable:  BoolPtr(true),
			CreatedAt:  membershipCreatedAt,
		}},
	}

	actual, err := client.UpdateAddressMap(context.Background(), AccountIdentifier(testAccountID), UpdateAddressMapParams{ID: "9a7806061c88ada191ed06f989cc3dac"})
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestDeleteAddressMap(t *testing.T) {
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
			  "id": "9a7806061c88ada191ed06f989cc3dac"
			}
		  }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/addressing/address_maps/9a7806061c88ada191ed06f989cc3dac", handler)

	err := client.DeleteAddressMap(context.Background(), AccountIdentifier(testAccountID), "9a7806061c88ada191ed06f989cc3dac")
	assert.NoError(t, err)
}

func TestAddIPAddressToAddressMap(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": []
		  }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/addressing/address_maps/9a7806061c88ada191ed06f989cc3dac/ips/192.0.2.1", handler)

	err := client.CreateIPAddressToAddressMap(context.Background(), AccountIdentifier(testAccountID), CreateIPAddressToAddressMapParams{"9a7806061c88ada191ed06f989cc3dac", "192.0.2.1"})
	assert.NoError(t, err)
}

func TestRemoveIPAddressFromAddressMap(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": []
		  }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/addressing/address_maps/9a7806061c88ada191ed06f989cc3dac/ips/192.0.2.1", handler)

	err := client.DeleteIPAddressFromAddressMap(context.Background(), AccountIdentifier(testAccountID), DeleteIPAddressFromAddressMapParams{"9a7806061c88ada191ed06f989cc3dac", "192.0.2.1"})
	assert.NoError(t, err)
}

func TestAddZoneToAddressMap(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": []
		  }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/addressing/address_maps/9a7806061c88ada191ed06f989cc3dac/zones/01a7362d577a6c3019a474fd6f485823", handler)

	err := client.CreateMembershipToAddressMap(context.Background(), AccountIdentifier(testAccountID), CreateMembershipToAddressMapParams{"9a7806061c88ada191ed06f989cc3dac", AddressMapMembershipContainer{"01a7362d577a6c3019a474fd6f485823", AddressMapMembershipZone}})
	assert.NoError(t, err)
}

func TestRemoveZoneFromAddressMap(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": []
		  }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/addressing/address_maps/9a7806061c88ada191ed06f989cc3dac/zones/01a7362d577a6c3019a474fd6f485823", handler)

	err := client.DeleteMembershipFromAddressMap(context.Background(), AccountIdentifier(testAccountID), DeleteMembershipFromAddressMapParams{"9a7806061c88ada191ed06f989cc3dac", AddressMapMembershipContainer{"01a7362d577a6c3019a474fd6f485823", AddressMapMembershipZone}})
	assert.NoError(t, err)
}

func TestAddAccountToAddressMap(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": []
		  }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/addressing/address_maps/9a7806061c88ada191ed06f989cc3dac/accounts/01a7362d577a6c3019a474fd6f485823", handler)

	err := client.CreateMembershipToAddressMap(context.Background(), AccountIdentifier(testAccountID), CreateMembershipToAddressMapParams{"9a7806061c88ada191ed06f989cc3dac", AddressMapMembershipContainer{"01a7362d577a6c3019a474fd6f485823", AddressMapMembershipAccount}})
	assert.NoError(t, err)
}

func TestRemoveAccountFromAddressMap(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": []
		  }`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/addressing/address_maps/9a7806061c88ada191ed06f989cc3dac/accounts/01a7362d577a6c3019a474fd6f485823", handler)

	err := client.DeleteMembershipFromAddressMap(context.Background(), AccountIdentifier(testAccountID), DeleteMembershipFromAddressMapParams{"9a7806061c88ada191ed06f989cc3dac", AddressMapMembershipContainer{"01a7362d577a6c3019a474fd6f485823", AddressMapMembershipAccount}})
	assert.NoError(t, err)
}
