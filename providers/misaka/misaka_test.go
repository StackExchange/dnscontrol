package misaka

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DNSControl/dnscontrol/v5/models"
)

// zonesServer knows one zone, example.com, with an apex NS record set, and
// records every zone that is created.
func zonesServer(created *[]string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/dns/zones/example.com/settings":
			_, _ = w.Write([]byte(`{"default_ttl": 600, "managed_soa": true, "managed_acme": false}`))
		case r.Method == http.MethodGet && r.URL.Path == "/dns/zones/example.com/recordsets/@/NS":
			_, _ = w.Write([]byte(`{"name": "@", "type": "NS", "ttl": 600, "enabled": true, "protected": false, "alias": null,
				"records": [{"value": "a.m1ns.moe"}, {"value": "b.m1ns.uk."}, {"value": "c.m1ns.us"}], "filters": []}`))
		case r.Method == http.MethodGet && r.URL.Path == "/dns/zones":
			_, _ = w.Write([]byte(`{"results": [{"name": "example.com"}], "count": 1}`))
		case r.Method == http.MethodPost && r.URL.Path == "/dns/zones":
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			*created = append(*created, body["name"])
			_, _ = w.Write([]byte(`{"name": "new.example.com"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"detail": "Zone not found"}`))
		}
	}))
}

func TestGetNameservers(t *testing.T) {
	server := zonesServer(new([]string))
	defer server.Close()
	p := testProvider(server.URL)

	nss, err := p.GetNameservers("example.com")
	if err != nil {
		t.Fatalf("GetNameservers() error = %v", err)
	}
	got := models.NameserversToStrings(nss)
	want := []string{"a.m1ns.moe", "b.m1ns.uk", "c.m1ns.us"}
	if len(got) != len(want) {
		t.Fatalf("GetNameservers() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("GetNameservers() = %v, want %v", got, want)
		}
	}

	// A zone that does not exist yet is not an error: preview asks before it
	// knows that the zone will be created.
	nss, err = p.GetNameservers("new.example.com")
	if err != nil || len(nss) != 0 {
		t.Errorf("GetNameservers() of a missing zone = %v, %v; want none and no error", nss, err)
	}
}

func TestListZones(t *testing.T) {
	server := zonesServer(new([]string))
	defer server.Close()

	zones, err := testProvider(server.URL).ListZones()
	if err != nil {
		t.Fatalf("ListZones() error = %v", err)
	}
	if len(zones) != 1 || zones[0] != "example.com" {
		t.Errorf("ListZones() = %v, want [example.com]", zones)
	}
}

func TestEnsureZoneExists(t *testing.T) {
	var created []string
	server := zonesServer(&created)
	defer server.Close()
	p := testProvider(server.URL)

	if err := p.EnsureZoneExists(models.MustNewDomainConfig("example.com")); err != nil {
		t.Fatalf("EnsureZoneExists() error = %v", err)
	}
	if len(created) != 0 {
		t.Errorf("an existing zone was created again: %v", created)
	}

	if err := p.EnsureZoneExists(models.MustNewDomainConfig("new.example.com")); err != nil {
		t.Fatalf("EnsureZoneExists() error = %v", err)
	}
	if len(created) != 1 || created[0] != "new.example.com" {
		t.Errorf("created = %v, want [new.example.com]", created)
	}
}

func TestEnsureZoneExistsReportsOtherErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"detail": "DNS service is not enabled for this Team or your role lacks access."}`))
	}))
	defer server.Close()

	err := testProvider(server.URL).EnsureZoneExists(models.MustNewDomainConfig("example.com"))
	if err == nil || isNotFound(err) {
		t.Fatalf("EnsureZoneExists() error = %v, want the 403", err)
	}
}

func TestInitializeRequiresToken(t *testing.T) {
	var p misakaProvider
	if err := p.Initialize(map[string]string{}, nil, nil); err == nil {
		t.Fatal("Initialize() without api_token succeeded")
	}
	if err := p.Initialize(map[string]string{"api_token": "x"}, nil, nil); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}
	if p.client == nil || p.observer == nil {
		t.Error("Initialize() left the client or the observer unset")
	}
}
