package domain

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/prasmussen/gandi-api/client"
	"github.com/prasmussen/gandi-api/live_dns/record"
	"github.com/prasmussen/gandi-api/live_dns/test_helpers"
	"github.com/stretchr/testify/assert"
)

func RunTest(t testing.TB, method, uri, requestBody, responseBody string, code int, call func(t testing.TB, d *Domain)) {
	testHelpers.RunTest(t, method, uri, requestBody, responseBody, code, func(t testing.TB, c *client.Client) {
		call(t, New(c))
	})
}

func TestList(t *testing.T) {
	RunTest(t,
		"GET", "/api/v5/domains",
		``,
		`[
			{
			  "fqdn": "example.com",
			  "domain_records_href": "https://dns.api.gandi.net/api/v5/domains/example.com/records",
			  "domain_href": "https://dns.api.gandi.net/api/v5/domains/example.com"
			},
			{
			  "fqdn": "example.fr",
			  "domain_records_href": "https://dns.api.gandi.net/api/v5/domains/example.fr/records",
			  "domain_href": "https://dns.api.gandi.net/api/v5/domains/example.fr"
			},
			{
			  "fqdn": "example.cat",
			  "domain_records_href": "https://dns.api.gandi.net/api/v5/domains/example.cat/records",
			  "domain_href": "https://dns.api.gandi.net/api/v5/domains/example.cat"
			}
		  ]`,
		http.StatusOK,
		func(t testing.TB, d *Domain) {
			info, err := d.List()
			assert.NoError(t, err)
			assert.Equal(t, []*InfoBase{
				&InfoBase{
					Fqdn:              "example.com",
					DomainRecordsHref: "https://dns.api.gandi.net/api/v5/domains/example.com/records",
					DomainHref:        "https://dns.api.gandi.net/api/v5/domains/example.com",
				},
				&InfoBase{
					Fqdn:              "example.fr",
					DomainRecordsHref: "https://dns.api.gandi.net/api/v5/domains/example.fr/records",
					DomainHref:        "https://dns.api.gandi.net/api/v5/domains/example.fr",
				},
				&InfoBase{
					Fqdn:              "example.cat",
					DomainRecordsHref: "https://dns.api.gandi.net/api/v5/domains/example.cat/records",
					DomainHref:        "https://dns.api.gandi.net/api/v5/domains/example.cat",
				},
			}, info)
		},
	)
}

func TestInfo(t *testing.T) {
	RunTest(t,
		"GET", "/api/v5/domains/example.com",
		``,
		`{
			"zone_uuid": "f05ac8b8-e447-11e7-8e33-00163ec31f40",
			"domain_keys_href": "https://dns.api.gandi.net/api/v5/domains/example.com/keys",
			"fqdn": "example.com",
			"zone_href": "https://dns.api.gandi.net/api/v5/zones/f05ac8b8-e447-11e7-8e33-00163ec31f40",
			"zone_records_href": "https://dns.api.gandi.net/api/v5/zones/f05ac8b8-e447-11e7-8e33-00163ec31f40/records",
			"domain_records_href": "https://dns.api.gandi.net/api/v5/domains/example.com/records",
			"domain_href": "https://dns.api.gandi.net/api/v5/domains/example.com"
		  }`,
		http.StatusOK,
		func(t testing.TB, d *Domain) {
			info, err := d.Info("example.com")
			assert.NoError(t, err)
			id, err := uuid.Parse("f05ac8b8-e447-11e7-8e33-00163ec31f40")
			assert.NoError(t, err)
			assert.Equal(t, &Info{
				&InfoBase{
					Fqdn:              "example.com",
					DomainRecordsHref: "https://dns.api.gandi.net/api/v5/domains/example.com/records",
					DomainHref:        "https://dns.api.gandi.net/api/v5/domains/example.com",
				},
				&InfoExtra{
					ZoneUUID:        &id,
					ZoneHref:        "https://dns.api.gandi.net/api/v5/zones/f05ac8b8-e447-11e7-8e33-00163ec31f40",
					ZoneRecordsHref: "https://dns.api.gandi.net/api/v5/zones/f05ac8b8-e447-11e7-8e33-00163ec31f40/records",
					DomainKeysHref:  "https://dns.api.gandi.net/api/v5/domains/example.com/keys",
				},
			}, info)
		},
	)
}

func TestRecords(t *testing.T) {
	d := New(&client.Client{})
	records := d.Records("example.com").(*record.Record)
	assert.Equal(t, "/domains/example.com", records.Prefix)
}
