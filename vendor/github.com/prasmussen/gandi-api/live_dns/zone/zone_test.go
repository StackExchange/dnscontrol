package zone

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/prasmussen/gandi-api/client"
	"github.com/prasmussen/gandi-api/live_dns/domain"
	"github.com/prasmussen/gandi-api/live_dns/record"
	"github.com/prasmussen/gandi-api/live_dns/test_helpers"
	"github.com/stretchr/testify/assert"
)

func RunTest(t testing.TB, method, uri, requestBody, responseBody string, code int, call func(t testing.TB, z *Zone)) {
	testHelpers.RunTest(t, method, uri, requestBody, responseBody, code, func(t testing.TB, c *client.Client) {
		call(t, New(c))
	})
}
func TestNoID(t *testing.T) {
	zoneInfo := Info{
		Name: "example.com",
	}
	z := New(&client.Client{})
	t.Run("Test Info", func(t *testing.T) {
		infos, err := z.Info(zoneInfo)
		assert.Error(t, err)
		assert.Nil(t, infos)
	})
	t.Run("Test Create", func(t *testing.T) {
		infos, err := z.Create(zoneInfo)
		assert.Error(t, err)
		assert.Nil(t, infos)
	})
	t.Run("Test Update", func(t *testing.T) {
		infos, err := z.Update(zoneInfo)
		assert.Error(t, err)
		assert.Nil(t, infos)
	})
	t.Run("Test Delete", func(t *testing.T) {
		err := z.Delete(zoneInfo)
		assert.Error(t, err)
	})
	t.Run("Test Domains", func(t *testing.T) {
		infos, err := z.Domains(zoneInfo)
		assert.Error(t, err)
		assert.Nil(t, infos)
	})
	t.Run("Test Set domain", func(t *testing.T) {
		infos, err := z.Set("example.com", zoneInfo)
		assert.Error(t, err)
		assert.Nil(t, infos)
	})
}

func TestInfo(t *testing.T) {
	RunTest(t,
		"GET", "/api/v5/zones/f05ac8b8-e447-11e7-8e33-00163ec31f40",
		``,
		`{
			"retry": 3600,
			"uuid": "f05ac8b8-e447-11e7-8e33-00163ec31f40",
			"zone_href": "https://dns.api.gandi.net/api/v5/zones/f05ac8b8-e447-11e7-8e33-00163ec31f40",
			"minimum": 10800,
			"domains_href": "https://dns.api.gandi.net/api/v5/zones/f05ac8b8-e447-11e7-8e33-00163ec31f40/domains",
			"refresh": 10800,
			"zone_records_href": "https://dns.api.gandi.net/api/v5/zones/f05ac8b8-e447-11e7-8e33-00163ec31f40/records",
			"expire": 604800,
			"sharing_id": "d85976ac-16d8-11e7-bbe1-00163e61ef31",
			"serial": 1513638328,
			"email": "hostmaster.gandi.net.",
			"primary_ns": "ns1.gandi.net",
			"name": "example.com zone"
		  }`,
		http.StatusOK,
		func(t testing.TB, z *Zone) {
			id, err := uuid.Parse("f05ac8b8-e447-11e7-8e33-00163ec31f40")
			assert.NoError(t, err)
			sharingID, err := uuid.Parse("d85976ac-16d8-11e7-bbe1-00163e61ef31")
			assert.NoError(t, err)
			zoneInfo := Info{
				Retry:           3600,
				UUID:            &id,
				ZoneHref:        "https://dns.api.gandi.net/api/v5/zones/f05ac8b8-e447-11e7-8e33-00163ec31f40",
				DomainsHref:     "https://dns.api.gandi.net/api/v5/zones/f05ac8b8-e447-11e7-8e33-00163ec31f40/domains",
				Minimum:         10800,
				Refresh:         10800,
				ZoneRecordsHref: "https://dns.api.gandi.net/api/v5/zones/f05ac8b8-e447-11e7-8e33-00163ec31f40/records",
				Expire:          604800,
				SharingID:       &sharingID,
				Serial:          1513638328,
				Email:           "hostmaster.gandi.net.",
				PrimaryNS:       "ns1.gandi.net",
				Name:            "Something",
			}
			info, err := z.Info(zoneInfo)
			zoneInfo.Name = "example.com zone"
			assert.NoError(t, err)
			assert.Equal(t, zoneInfo.Retry, info.Retry)
			assert.Equal(t, zoneInfo.UUID, info.UUID)
			assert.Equal(t, zoneInfo.ZoneHref, info.ZoneHref)
			assert.Equal(t, zoneInfo.DomainsHref, info.DomainsHref)
			assert.Equal(t, zoneInfo.Minimum, info.Minimum)
			assert.Equal(t, zoneInfo.Refresh, info.Refresh)
			assert.Equal(t, zoneInfo.ZoneRecordsHref, info.ZoneRecordsHref)
			assert.Equal(t, zoneInfo.Expire, info.Expire)
			assert.Equal(t, zoneInfo.SharingID, info.SharingID)
			assert.Equal(t, zoneInfo.Serial, info.Serial)
			assert.Equal(t, zoneInfo.Email, info.Email)
			assert.Equal(t, zoneInfo.PrimaryNS, info.PrimaryNS)
			assert.Equal(t, zoneInfo.Name, info.Name)
		},
	)
}

func TestList(t *testing.T) {
	RunTest(t,
		"GET", "/api/v5/zones",
		``,
		`[
			{
			  "retry": 3600,
			  "uuid": "f05ac8b8-e447-11e7-8e33-00163ec31f40",
			  "zone_href": "https://dns.api.gandi.net/api/v5/zones/f05ac8b8-e447-11e7-8e33-00163ec31f40",
			  "minimum": 10800,
			  "domains_href": "https://dns.api.gandi.net/api/v5/zones/f05ac8b8-e447-11e7-8e33-00163ec31f40/domains",
			  "refresh": 10800,
			  "zone_records_href": "https://dns.api.gandi.net/api/v5/zones/f05ac8b8-e447-11e7-8e33-00163ec31f40/records",
			  "expire": 604800,
			  "sharing_id": "d85976ac-16d8-11e7-bbe1-00163e61ef31",
			  "serial": 1513638328,
			  "email": "hostmaster.gandi.net.",
			  "primary_ns": "ns1.gandi.net",
			  "name": "example.com zone"
			}
		]`,
		http.StatusOK,
		func(t testing.TB, z *Zone) {
			info, err := z.List()
			assert.NoError(t, err)
			assert.Equal(t, 1, len(info))
			assert.Equal(t, "f05ac8b8-e447-11e7-8e33-00163ec31f40", info[0].UUID.String())
		},
	)
}

func TestCreate(t *testing.T) {
	RunTest(t,
		"POST", "/api/v5/zones",
		`{"name": "example.com Zone"}`,
		`{"message": "Zone Created", "uuid": "12bb7678-e43e-11e7-80c1-00163e6dc886"}`,
		http.StatusCreated,
		func(t testing.TB, z *Zone) {
			zoneInfo := Info{
				Name: "example.com Zone",
			}
			info, err := z.Create(zoneInfo)
			assert.NoError(t, err)
			assert.Equal(t, "12bb7678-e43e-11e7-80c1-00163e6dc886", info.UUID.String())
			assert.Equal(t, "Zone Created", info.Message)
		},
	)
}

func TestUpdate(t *testing.T) {
	RunTest(t,
		"PATCH", "/api/v5/zones/12bb7678-e43e-11e7-80c1-00163e6dc886",
		`{"name": "example.com","uuid":"12bb7678-e43e-11e7-80c1-00163e6dc886"}`,
		`{"message": "Request Accepted"}`,
		http.StatusAccepted,
		func(t testing.TB, z *Zone) {
			id, err := uuid.Parse("12bb7678-e43e-11e7-80c1-00163e6dc886")
			assert.NoError(t, err)
			zoneInfo := Info{
				Name: "example.com",
				UUID: &id,
			}
			info, err := z.Update(zoneInfo)
			assert.NoError(t, err)
			assert.Equal(t, "Request Accepted", info.Message)
		},
	)
}

func TestDelete(t *testing.T) {
	RunTest(t,
		"DELETE", "/api/v5/zones/12bb7678-e43e-11e7-80c1-00163e6dc886",
		``,
		``,
		http.StatusNoContent,
		func(t testing.TB, z *Zone) {
			id, err := uuid.Parse("12bb7678-e43e-11e7-80c1-00163e6dc886")
			assert.NoError(t, err)
			zoneInfo := Info{
				Name: "example.com",
				UUID: &id,
			}
			err = z.Delete(zoneInfo)
			assert.NoError(t, err)
		},
	)
}

func TestDeleteWrongCode(t *testing.T) {
	RunTest(t,
		"DELETE", "/api/v5/zones/12bb7678-e43e-11e7-80c1-00163e6dc886",
		``,
		``,
		http.StatusNotFound,
		func(t testing.TB, z *Zone) {
			id, err := uuid.Parse("12bb7678-e43e-11e7-80c1-00163e6dc886")
			assert.NoError(t, err)
			zoneInfo := Info{
				Name: "example.com",
				UUID: &id,
			}
			err = z.Delete(zoneInfo)
			assert.Error(t, err)
		},
	)
}

func TestDomains(t *testing.T) {
	RunTest(t,
		"GET", "/api/v5/zones/12bb7678-e43e-11e7-80c1-00163e6dc886/domains",
		``,
		`[
			{
			  "fqdn": "example.com",
			  "domain_records_href": "https://dns.api.gandi.net/api/v5/domains/example.com/records",
			  "domain_href": "https://dns.api.gandi.net/api/v5/domains/example.com"
			}
		  ]`,
		http.StatusOK,
		func(t testing.TB, z *Zone) {
			id, err := uuid.Parse("12bb7678-e43e-11e7-80c1-00163e6dc886")
			assert.NoError(t, err)
			zoneInfo := Info{
				Name: "example.com",
				UUID: &id,
			}
			domains, err := z.Domains(zoneInfo)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(domains))
			assert.Equal(t, &domain.InfoBase{
				Fqdn:              "example.com",
				DomainRecordsHref: "https://dns.api.gandi.net/api/v5/domains/example.com/records",
				DomainHref:        "https://dns.api.gandi.net/api/v5/domains/example.com",
			}, domains[0])
		},
	)
}

func TestRecords(t *testing.T) {
	id, err := uuid.Parse("12bb7678-e43e-11e7-80c1-00163e6dc886")
	assert.NoError(t, err)
	zoneInfo := Info{
		Name: "example.com",
		UUID: &id,
	}
	z := New(&client.Client{})
	records := z.Records(zoneInfo).(*record.Record)
	assert.Equal(t, "/zones/12bb7678-e43e-11e7-80c1-00163e6dc886", records.Prefix)
}

func TestSet(t *testing.T) {
	RunTest(t,
		"POST", "/api/v5/zones/12bb7678-e43e-11e7-80c1-00163e6dc886/domains/example.com",
		``,
		`{
			"message": "Domain Created"
		  }`,
		http.StatusCreated,
		func(t testing.TB, z *Zone) {
			id, err := uuid.Parse("12bb7678-e43e-11e7-80c1-00163e6dc886")
			assert.NoError(t, err)
			zoneInfo := Info{
				UUID: &id,
			}
			status, err := z.Set("example.com", zoneInfo)
			assert.NoError(t, err)
			assert.Equal(t, &Status{Message: "Domain Created"}, status)
		},
	)
}
