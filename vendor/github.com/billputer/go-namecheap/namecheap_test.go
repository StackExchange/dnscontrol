package namecheap

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

var (
	mux    *http.ServeMux
	client *Client
	server *httptest.Server
)

// This method of testing http client APIs is borrowed from
// Will Norris's work in go-github @ https://github.com/google/go-github
func setup() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	client = NewClient("anApiUser", "anToken", "anUser")
	client.BaseURL = server.URL + "/"
}

func teardown() {
	server.Close()
}

func testMethod(t *testing.T, r *http.Request, want string) {
	if want != r.Method {
		t.Errorf("Request method = %v, want %v", r.Method, want)
	}
}

func TestNewClient(t *testing.T) {
	c := NewClient("anApiUser", "anToken", "anUser")

	if c.BaseURL != defaultBaseURL {
		t.Errorf("NewClient BaseURL = %v, want %v", c.BaseURL, defaultBaseURL)
	}
}

// Verify that the MakeRequest function assembles the correct API URL
func TestMakeRequest(t *testing.T) {
	c := NewClient("anApiUser", "anToken", "anUser")
	c.BaseURL = "https://fake-api-server/"
	requestInfo := &ApiRequest{
		method:  "GET",
		command: "namecheap.domains.getList",
		params:  url.Values{},
	}
	req, _ := c.makeRequest(requestInfo)

	// correctly assembled URL
	outURL := "https://fake-api-server/?ApiKey=anToken&ApiUser=anApiUser&ClientIp=127.0.0.1&Command=namecheap.domains.getList&UserName=anUser"

	// test that URL was correctly assembled
	if req.URL.String() != outURL {
		t.Errorf("NewRequest() URL = %v, want %v", req.URL, outURL)
	}
}
