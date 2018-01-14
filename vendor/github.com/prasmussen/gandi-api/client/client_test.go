package client

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testSerialized struct {
	Value string
	Other int `json:"other-value"`
}

func Test_Get(t *testing.T) {

	c := Client{
		Key: "abc",
	}
	calls := []struct {
		method  string
		code    int
		handler func(string, interface{}) (*http.Response, error)
		value   string
	}{
		{"GET", http.StatusOK, c.Get, "abc"},
		{"DELETE", http.StatusNoContent, c.Delete, "abcdef"},
	}

	for _, test := range calls {
		t.Run(test.method, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, test.method, r.Method)
				assert.Equal(t, "/something", r.RequestURI)
				assert.Equal(t, "", r.Header.Get("Content-Type"))
				assert.Equal(t, "", r.Header.Get("Accept"))
				assert.Equal(t, "abc", r.Header.Get("X-Api-Key"))
				b, err := ioutil.ReadAll(r.Body)
				assert.NoError(t, err)
				assert.Empty(t, b)
				w.WriteHeader(test.code)
			}
			s := httptest.NewServer(http.HandlerFunc(handler))
			c.Url = s.URL + "/"

			defer s.Close()
			resp, err := test.handler("/something", nil)
			assert.NoError(t, err)
			assert.Equal(t, test.code, resp.StatusCode)
			b, err := ioutil.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Empty(t, b)
		})

		t.Run(test.method+" with response data", func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				assert.Equal(t, test.method, r.Method)
				assert.Equal(t, "/something", r.RequestURI)
				assert.Equal(t, "", r.Header.Get("Content-Type"))
				assert.Equal(t, "application/json", r.Header.Get("Accept"))
				assert.Equal(t, "abc", r.Header.Get("X-Api-Key"))
				b, err := ioutil.ReadAll(r.Body)
				assert.NoError(t, err)
				assert.Empty(t, b)
				w.WriteHeader(test.code)
				w.Write([]byte(`{"value":"abc"}`))
			}
			s := httptest.NewServer(http.HandlerFunc(handler))
			c.Url = s.URL + "/"

			data := testSerialized{
				Value: "abcdef",
				Other: 4,
			}

			defer s.Close()
			_, err := test.handler("/something", &data)
			assert.NoError(t, err)
			assert.Equal(t, testSerialized{Value: test.value, Other: 4}, data)
		})
	}
}

func Test_CallWithData(t *testing.T) {

	c := Client{
		Key: "abc",
	}
	calls := []struct {
		method  string
		code    int
		handler func(string, interface{}, interface{}) (*http.Response, error)
	}{
		{"POST", http.StatusCreated, c.Post},
		{"PATCH", http.StatusAccepted, c.Patch},
		{"PUT", http.StatusOK, c.Put},
	}
	for _, test := range calls {
		data := testSerialized{
			Value: "abcdef",
			Other: 4,
		}
		t.Run(test.method, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, test.method, r.Method)
				assert.Equal(t, "/something", r.RequestURI)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "", r.Header.Get("Accept"))
				assert.Equal(t, "abc", r.Header.Get("X-Api-Key"))
				d := json.NewDecoder(r.Body)
				values := map[string]interface{}{}
				err := d.Decode(&values)
				assert.NoError(t, err)
				assert.Equal(t, map[string]interface{}{"Value": "abcdef", "other-value": float64(4)}, values)
				w.WriteHeader(test.code)
			}
			s := httptest.NewServer(http.HandlerFunc(handler))
			defer s.Close()
			c.Url = s.URL + "/"

			_, err := test.handler("/something", &data, nil)
			assert.NoError(t, err)
		})

		t.Run(test.method+" with response data", func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, test.method, r.Method)
				assert.Equal(t, "/something", r.RequestURI)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "application/json", r.Header.Get("Accept"))
				assert.Equal(t, "abc", r.Header.Get("X-Api-Key"))
				d := json.NewDecoder(r.Body)
				values := map[string]interface{}{}
				err := d.Decode(&values)
				assert.NoError(t, err)
				assert.Equal(t, map[string]interface{}{"Value": "abcdef", "other-value": float64(4)}, values)
				w.WriteHeader(test.code)
				w.Write([]byte(`{"value":"abc"}`))
			}
			s := httptest.NewServer(http.HandlerFunc(handler))
			defer s.Close()
			c.Url = s.URL + "/"

			data := testSerialized{
				Value: "abcdef",
				Other: 4,
			}

			_, err := test.handler("/something", &data, &data)
			assert.NoError(t, err)
			assert.Equal(t, testSerialized{Value: "abc", Other: 4}, data)
		})
	}
}
