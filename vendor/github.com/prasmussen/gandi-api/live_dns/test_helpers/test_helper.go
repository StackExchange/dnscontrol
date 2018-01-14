package testHelpers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prasmussen/gandi-api/client"
	"github.com/stretchr/testify/assert"
)

// RunTest starts an http, asserts calls provided as arguments and writes the response
func RunTest(t testing.TB, method, uri, requestBody, responseBody string, code int, call func(t testing.TB, c *client.Client)) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := r.Body.Close()
			assert.NoError(t, err)
		}()
		assert.Equal(t, uri, r.RequestURI)
		if len(requestBody) > 0 {
			var body map[string]interface{}
			assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))

			var expectedBody map[string]interface{}
			assert.NoError(t, json.NewDecoder(strings.NewReader(requestBody)).Decode(&expectedBody))
			assert.Equal(t, expectedBody, body)
		} else {
			b, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(b))
		}

		w.WriteHeader(code)
		_, err := w.Write([]byte(responseBody))
		assert.NoError(t, err)
	}
	s := httptest.NewServer(http.HandlerFunc(handler))
	defer s.Close()
	c := &client.Client{
		Key: "test",
		Url: s.URL + "/api/v5",
	}
	call(t, c)
}
