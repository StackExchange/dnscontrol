package misaka

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// pagedServer answers GET /dns/zones with one page per request, in order and
// whatever the offset, so a test can serve short pages.
func pagedServer(t *testing.T, count int, pages ...[]string) *httptest.Server {
	t.Helper()
	call := 0
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dns/zones" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if got := r.Header.Get("Authorization"); got != "Token test-token" {
			t.Errorf("Authorization = %q, want %q", got, "Token test-token")
		}
		if got := r.Header.Get("User-Agent"); !strings.HasPrefix(got, "dnscontrol/") {
			t.Errorf("User-Agent = %q, want dnscontrol/...", got)
		}
		var names []string
		if call < len(pages) {
			names = pages[call]
		}
		call++

		results := make([]string, 0, len(names))
		for _, n := range names {
			results = append(results, fmt.Sprintf(`{"name":%q}`, n))
		}
		_, _ = fmt.Fprintf(w, `{"results":[%s],"count":%d}`, strings.Join(results, ","), count)
	}))
}

func TestListAllPagination(t *testing.T) {
	tests := []struct {
		name    string
		count   int
		pages   [][]string
		want    []string
		wantErr string
	}{
		{
			name:  "empty account",
			count: 0,
			pages: [][]string{{}},
			want:  []string{},
		},
		{
			name:  "two pages are merged",
			count: 3,
			pages: [][]string{{"a.example", "b.example"}, {"c.example"}},
			want:  []string{"a.example", "b.example", "c.example"},
		},
		{
			name:  "a short intermediate page does not end the listing",
			count: 4,
			pages: [][]string{{"a.example"}, {"b.example", "c.example"}, {"d.example"}},
			want:  []string{"a.example", "b.example", "c.example", "d.example"},
		},
		{
			name:    "a premature empty page is an error",
			count:   3,
			pages:   [][]string{{"a.example", "b.example"}, {}},
			wantErr: "API returned 2 of 3 items",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := pagedServer(t, tt.count, tt.pages...)
			defer server.Close()

			zones, err := newAPIClient(server.URL, "test-token").listZones()
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("listZones() error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("listZones() error = %v", err)
			}

			got := make([]string, 0, len(zones))
			for _, z := range zones {
				got = append(got, z.Name)
			}
			if strings.Join(got, ",") != strings.Join(tt.want, ",") {
				t.Errorf("listZones() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListAllSendsOffset(t *testing.T) {
	var offsets []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		offsets = append(offsets, r.URL.Query().Get("limit")+"/"+r.URL.Query().Get("offset"))
		if r.URL.Query().Get("offset") == "0" {
			_, _ = w.Write([]byte(`{"results":[{"name":"a.example"},{"name":"b.example"}],"count":3}`))
			return
		}
		_, _ = w.Write([]byte(`{"results":[{"name":"c.example"}],"count":3}`))
	}))
	defer server.Close()

	if _, err := newAPIClient(server.URL, "test-token").listZones(); err != nil {
		t.Fatalf("listZones() error = %v", err)
	}
	if strings.Join(offsets, ",") != "100/0,100/2" {
		t.Errorf("limit/offset = %v, want [100/0 100/2]", offsets)
	}
}

func TestAPIErrors(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
		want   []string
	}{
		{
			name:   "invalid token",
			status: http.StatusUnauthorized,
			body:   `{"detail": "Invalid token."}`,
			want:   []string{"MISAKA", "GET /dns/zones/example.com/settings", "HTTP 401", "Invalid token."},
		},
		{
			name:   "team without DNS access",
			status: http.StatusForbidden,
			body:   `{"detail": "DNS service is not enabled for this Team or your role lacks access."}`,
			want:   []string{"HTTP 403", "DNS service is not enabled for this Team or your role lacks access."},
		},
		{
			name:   "missing zone",
			status: http.StatusNotFound,
			body:   `{"detail": "Zone not found"}`,
			want:   []string{"HTTP 404", "Zone not found"},
		},
		{
			name:   "validation error",
			status: http.StatusUnprocessableEntity,
			body:   `{"detail": [{"detail": "Invalid record"}, {"detail": "Ensure this value is less than or equal to 2147483647."}]}`,
			want:   []string{"HTTP 422", "Invalid record; Ensure this value is less than or equal to 2147483647."},
		},
		{
			name:   "body that is not the API's own",
			status: http.StatusForbidden,
			body:   `<!DOCTYPE html><title>Access denied | api.misaka.io used Cloudflare to restrict access</title>`,
			want:   []string{"HTTP 403", "<!DOCTYPE html><title>Access denied"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			err := newAPIClient(server.URL, "test-token").getZoneSettings("example.com")
			if err == nil {
				t.Fatal("getZoneSettings() error = nil")
			}
			for _, w := range tt.want {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("error %q does not contain %q", err, w)
				}
			}
			if got, want := isNotFound(err), tt.status == http.StatusNotFound; got != want {
				t.Errorf("isNotFound() = %v, want %v", got, want)
			}
		})
	}
}

func TestErrorDetailTruncatesRawBody(t *testing.T) {
	got := errorDetail([]byte(strings.Repeat("x", 2*maxErrorBody)))
	if len(got) != maxErrorBody+len("...") {
		t.Errorf("len(errorDetail()) = %d, want %d", len(got), maxErrorBody+len("..."))
	}
}

// dropFirstServer closes the connection of its first request without
// answering, and answers every later one with status and body.
func dropFirstServer(t *testing.T, hits *int, status int, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		*hits++
		if *hits == 1 {
			conn, _, err := w.(http.Hijacker).Hijack()
			if err != nil {
				t.Fatal(err)
			}
			_ = conn.Close()
			return
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
}

func TestDroppedConnection(t *testing.T) {
	req := recordSetRequest{TTL: 300, Records: []apiRecord{{Value: "192.0.2.1"}}, Filters: []any{}}
	tests := []struct {
		name     string
		status   int
		body     string
		call     func(c *apiClient) error
		wantHits int
		wantErr  bool
	}{
		{
			name:     "GET is sent again",
			status:   http.StatusOK,
			body:     `{"name": "www", "type": "A", "records": [{"value": "192.0.2.1"}]}`,
			call:     func(c *apiClient) error { _, err := c.getRecordSet("example.com", "www", "A"); return err },
			wantHits: 2,
		},
		{
			name:     "PUT is sent again",
			status:   http.StatusOK,
			body:     `{}`,
			call:     func(c *apiClient) error { return c.updateRecordSet("example.com", "www", "A", req) },
			wantHits: 2,
		},
		{
			name:     "DELETE already applied by the dropped request",
			status:   http.StatusNotFound,
			body:     `{"detail": "RecordSet not found"}`,
			call:     func(c *apiClient) error { return c.deleteRecordSet("example.com", "www", "A") },
			wantHits: 2,
		},
		{
			name:     "POST is not sent again",
			status:   http.StatusOK,
			body:     `{}`,
			call:     func(c *apiClient) error { return c.createRecordSet("example.com", "www", "A", req) },
			wantHits: 1,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var hits int
			server := dropFirstServer(t, &hits, tt.status, tt.body)
			defer server.Close()

			c := newAPIClient(server.URL, "test-token")
			var slept int
			c.sleep = func(time.Duration) { slept++ }

			err := tt.call(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, want error %v", err, tt.wantErr)
			}
			if hits != tt.wantHits {
				t.Errorf("server saw %d requests, want %d", hits, tt.wantHits)
			}
			if slept != tt.wantHits-1 {
				t.Errorf("paused %d times, want %d", slept, tt.wantHits-1)
			}
		})
	}
}

func TestDeleteNotFoundIsAnError(t *testing.T) {
	// Only a repeated DELETE may meet a 404 that it caused itself.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"detail": "RecordSet not found"}`))
	}))
	defer server.Close()

	err := newAPIClient(server.URL, "test-token").deleteRecordSet("example.com", "www", "A")
	if !isNotFound(err) {
		t.Fatalf("deleteRecordSet() error = %v, want the 404", err)
	}
}
