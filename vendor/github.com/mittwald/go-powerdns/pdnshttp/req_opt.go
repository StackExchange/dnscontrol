package pdnshttp

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
)

// RequestOption is a special type of function that can be passed to most HTTP
// request functions in this package; it is used to modify an HTTP request and
// to implement special request logic.
type RequestOption func(*http.Request) error

// WithJSONRequestBody adds a JSON body to a request. The input type can be
// anything, as long as it can be marshaled by "json.Marshal". This method will
// also automatically set the correct content type and content-length.
func WithJSONRequestBody(in interface{}) RequestOption {
	return func(req *http.Request) error {
		if in == nil {
			return nil
		}

		buf := bytes.Buffer{}
		enc := json.NewEncoder(&buf)
		err := enc.Encode(in)

		if err != nil {
			return err
		}

		rc := ioutil.NopCloser(&buf)

		copyBuf := buf.Bytes()

		req.Body = rc
		req.Header.Set("Content-Type", "application/json")
		req.ContentLength = int64(buf.Len())
		req.GetBody = func() (io.ReadCloser, error) {
			r := bytes.NewReader(copyBuf)
			return ioutil.NopCloser(r), nil
		}

		return nil
	}
}

// WithQueryValue adds a query parameter to a request's URL.
func WithQueryValue(key, value string) RequestOption {
	return func(req *http.Request) error {
		q := req.URL.Query()
		q.Set(key, value)

		req.URL.RawQuery = q.Encode()
		return nil
	}
}
