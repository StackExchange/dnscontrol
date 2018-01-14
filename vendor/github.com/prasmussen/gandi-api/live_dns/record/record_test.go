package record

import (
	"net/http"
	"testing"

	"github.com/prasmussen/gandi-api/client"
	"github.com/prasmussen/gandi-api/live_dns/test_helpers"
	"github.com/stretchr/testify/assert"
)

func RunTest(t testing.TB, method, uri, requestBody, responseBody string, code int, call func(t testing.TB, r *Record)) {
	testHelpers.RunTest(t, method, uri, requestBody, responseBody, code, func(t testing.TB, c *client.Client) {
		call(t, New(c, "/zones/f05ac8b8-e447-11e7-8e33-00163ec31f40/"))
	})
}

func TestTooManyArgs(t *testing.T) {
	r := New(&client.Client{}, "")
	t.Run("Test Create", func(t *testing.T) {
		ret, err := r.Create(Info{}, "arg1", "arg2", "arg3")
		assert.Error(t, err)
		assert.Nil(t, ret)
	})
	t.Run("Test Update", func(t *testing.T) {
		ret, err := r.Update(Info{}, "arg1", "arg2", "arg3")
		assert.Error(t, err)
		assert.Nil(t, ret)
	})
	t.Run("Test List", func(t *testing.T) {
		ret, err := r.List("arg1", "arg2", "arg3")
		assert.Error(t, err)
		assert.Nil(t, ret)
	})
	t.Run("Test Delete", func(t *testing.T) {
		err := r.Delete("arg1", "arg2", "arg3")
		assert.Error(t, err)
	})
}

func TestList(t *testing.T) {
	tests := []struct {
		urlPattern string
		args       []string
	}{
		{
			urlPattern: "",
			args:       []string{},
		},
		{
			urlPattern: "/www",
			args:       []string{"www"},
		},
		{
			urlPattern: "/www/AAA",
			args:       []string{"www", "AAA"},
		},
	}
	for _, test := range tests {
		t.Run("with pattern "+test.urlPattern, func(t *testing.T) {
			RunTest(t,
				"GET", "/api/v5/zones/f05ac8b8-e447-11e7-8e33-00163ec31f40/records"+test.urlPattern,
				``,
				`[
					{
					"rrset_type": "A",
					"rrset_ttl": 10800,
					"rrset_name": "www",
					"rrset_href": "https://dns.api.gandi.net/api/v5/zones/f05ac8b8-e447-11e7-8e33-00163ec31f40/records/www/A",
					"rrset_values": [
						"127.0.0.1"
					]
					}
				]`,
				http.StatusOK,
				func(t testing.TB, r *Record) {
					recordInfos, err := r.List(test.args...)
					assert.NoError(t, err)
					assert.Equal(t, []*Info{
						{
							Name:   "www",
							TTL:    10800,
							Type:   A,
							Values: []string{"127.0.0.1"},
							Href:   "https://dns.api.gandi.net/api/v5/zones/f05ac8b8-e447-11e7-8e33-00163ec31f40/records/www/A",
						},
					}, recordInfos)
				},
			)
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		urlPattern string
		args       []string
	}{
		{
			urlPattern: "",
			args:       []string{},
		},
		{
			urlPattern: "/www",
			args:       []string{"www"},
		},
		{
			urlPattern: "/www/AAA",
			args:       []string{"www", "AAA"},
		},
	}
	for _, test := range tests {
		t.Run("with pattern "+test.urlPattern, func(t *testing.T) {
			RunTest(t,
				"DELETE", "/api/v5/zones/f05ac8b8-e447-11e7-8e33-00163ec31f40/records"+test.urlPattern,
				``,
				``,
				http.StatusNoContent,
				func(t testing.TB, r *Record) {
					err := r.Delete(test.args...)
					assert.NoError(t, err)
				},
			)
		})
	}
}

func TestCreate(t *testing.T) {
	tests := []struct {
		urlPattern string
		args       []string
	}{
		{
			urlPattern: "",
			args:       []string{},
		},
		{
			urlPattern: "/www",
			args:       []string{"www"},
		},
		{
			urlPattern: "/www/AAA",
			args:       []string{"www", "AAA"},
		},
	}
	for _, test := range tests {
		t.Run("with pattern "+test.urlPattern, func(t *testing.T) {
			RunTest(t,
				"POST", "/api/v5/zones/f05ac8b8-e447-11e7-8e33-00163ec31f40/records"+test.urlPattern,
				`{"rrset_name": "www","rrset_type": "A","rrset_ttl": 10800,"rrset_values": ["127.0.0.1"]}`,
				`{
					"message": "DNS Record Created"
				}`,
				http.StatusCreated,
				func(t testing.TB, z *Record) {
					recordInfo := Info{
						Name:   "www",
						TTL:    10800,
						Type:   A,
						Values: []string{"127.0.0.1"},
					}
					info, err := z.Create(recordInfo, test.args...)
					assert.NoError(t, err)
					assert.Equal(t, "DNS Record Created", info.Message)
				},
			)
		})
	}
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		urlPattern string
		args       []string
	}{
		{
			urlPattern: "",
			args:       []string{},
		},
		{
			urlPattern: "/www",
			args:       []string{"www"},
		},
		{
			urlPattern: "/www/AAA",
			args:       []string{"www", "AAA"},
		},
	}
	for _, test := range tests {
		t.Run("with pattern "+test.urlPattern, func(t *testing.T) {
			RunTest(t,
				"PUT", "/api/v5/zones/f05ac8b8-e447-11e7-8e33-00163ec31f40/records"+test.urlPattern,
				`{"rrset_name": "www","rrset_type": "A","rrset_ttl": 10800,"rrset_values": ["127.0.0.1"]}`,
				`{
					"message": "DNS Record Created"
				}`,
				http.StatusCreated,
				func(t testing.TB, z *Record) {
					recordInfo := Info{
						Name:   "www",
						TTL:    10800,
						Type:   A,
						Values: []string{"127.0.0.1"},
					}
					info, err := z.Update(recordInfo, test.args...)
					assert.NoError(t, err)
					assert.Equal(t, "DNS Record Created", info.Message)
				},
			)
		})
	}
}
