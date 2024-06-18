package cloudflare

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	testD1DatabaseID = "480f4f691a284fdd92401ed29f0ac1df"
	testD1Result     = `{
		  "created_at": "2014-01-01T05:20:00.12345Z",
		  "name": "my-database",
		  "uuid": "480f4f691a284fdd92401ed29f0ac1df",
		  "version": "beta",
		  "num_tables": 1,
		  "file_size": 100
		}
`
)

var (
	d1CreatedAt, _ = time.Parse(time.RFC3339, "2014-01-01T05:20:00.12345Z")
	testD1Database = D1Database{
		Name:      "my-database",
		NumTables: 1,
		UUID:      testD1DatabaseID,
		Version:   "beta",
		CreatedAt: &d1CreatedAt,
		FileSize:  100,
	}
)

func TestListD1Databases(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/d1/database", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
            "success": true,
            "errors": [],
            "messages": [],
            "result": [
                %s
            ],
            "result_info": {
                "page": 1,
                "per_page": 100,
                "count": 1,
                "total_count": 1
            }
        }`, testD1Result)
	})

	_, _, err := client.ListD1Databases(context.Background(), AccountIdentifier(""), ListD1DatabasesParams{})
	if assert.Error(t, err, "Didn't get error for missing Account ID get listing D1 Database") {
		assert.Equal(t, err.Error(), errMissingAccountID)
	}
	actual, _, err := client.ListD1Databases(context.Background(), testAccountRC, ListD1DatabasesParams{})
	if assert.NoError(t, err, "ListD1Databases returned error: %v", err) {
		expected := []D1Database{testD1Database}
		if !assert.Equal(t, expected, actual) {
			t.Errorf("ListD1Databases returned %+v, expected %+v", actual, expected)
		}
	}
}

func TestGetD1Databases(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/d1/database/"+testD1DatabaseID, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
            "success": true,
            "errors": [],
            "messages": [],
            "result": %s
        }`, testD1Result)
	})

	_, err := client.GetD1Database(context.Background(), AccountIdentifier(""), "")
	if assert.Error(t, err, "Didn't get error for missing Account ID get getting D1 Database") {
		assert.Equal(t, err.Error(), errMissingAccountID)
	}
	actual, err := client.GetD1Database(context.Background(), testAccountRC, testD1DatabaseID)
	if assert.NoError(t, err, "GetD1Database returned error: %v", err) {
		if !assert.Equal(t, testD1Database, actual) {
			t.Errorf("GetD1Database returned %+v, expected %+v", actual, testD1Database)
		}
	}
}

func TestCreateD1Database(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/d1/database", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": %s
		}`, testD1Result)
	})

	_, err := client.CreateD1Database(context.Background(), AccountIdentifier(""), CreateD1DatabaseParams{})
	if assert.Error(t, err, "Didn't get error for missing Account ID get creating D1 Database") {
		assert.Equal(t, err.Error(), errMissingAccountID)
	}
	actual, err := client.CreateD1Database(context.Background(), testAccountRC, CreateD1DatabaseParams{
		Name: "my-database",
	})
	if assert.NoError(t, err, "CreateD1Database returned error: %v", err) {
		if !assert.Equal(t, testD1Database, actual) {
			t.Errorf("CreateD1Database returned %+v, expected %+v", actual, testD1Database)
		}
	}
}

func TestDeleteD1Database(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/d1/database/"+testD1DatabaseID, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": []
		}`)
	})

	err := client.DeleteD1Database(context.Background(), AccountIdentifier(""), "")
	if assert.Error(t, err, "Didn't get error for missing Account ID get deleting D1 Database") {
		assert.Equal(t, err.Error(), errMissingAccountID)
	}
	err = client.DeleteD1Database(context.Background(), testAccountRC, testD1DatabaseID)
	assert.NoError(t, err, "DeleteD1Database returned error: %v", err)
}

func TestQueryD1Database(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/d1/database/"+testD1DatabaseID+"/query", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Error reading request body: %v", err)
		}
		if got := string(b); got != `{"sql":"SELECT * FROM my-database","params":["param1","param2"]}` {
			t.Errorf("request Body is %s, want %s", got, `{"sql":"SELECT * FROM my-database","params":["param1","param2"]}`)
		}
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": [
				{
					"success": true,
					"meta": {
						"changed_db": false,
						"changes": 0,
						"duration": 3.3,
						"last_row_id": 3,
						"rows_read": 3,
						"rows_written": 0,
						"size_after": 10
					},
					"results": [
						{
							"id": 1,
							"name": "test user"
						},
						{
							"id": 2,
							"name": "test user 2"
						}
					]
				}
			]
		}`)
	})

	_, err := client.QueryD1Database(context.Background(), AccountIdentifier(""), QueryD1DatabaseParams{})
	if assert.Error(t, err, "Didn't get error for missing Account ID get querying D1 Database") {
		assert.Equal(t, err, ErrMissingAccountID)
	}
	_, err = client.QueryD1Database(context.Background(), testAccountRC, QueryD1DatabaseParams{})
	if assert.Error(t, err, "Didn't get error for missing D1 Database ID get querying D1 Database") {
		assert.Equal(t, err, ErrMissingDatabaseID)
	}
	actual, err := client.QueryD1Database(context.Background(), testAccountRC, QueryD1DatabaseParams{
		DatabaseID: testD1DatabaseID,
		SQL:        "SELECT * FROM my-database",
		Parameters: []string{"param1", "param2"},
	})
	if assert.NoError(t, err, "QueryD1Database returned error: %v", err) {
		expected := D1Result{
			Success: BoolPtr(true),
			Meta: D1DatabaseMetadata{
				ChangedDB:   BoolPtr(false),
				Changes:     0,
				Duration:    3.3,
				LastRowID:   3,
				RowsRead:    3,
				RowsWritten: 0,
				SizeAfter:   10,
			},
			Results: []map[string]any{
				{
					"id":   float64(1),
					"name": "test user",
				},
				{
					"id":   float64(2),
					"name": "test user 2",
				},
			},
		}
		if !assert.Equal(t, expected, actual[0]) {
			t.Errorf("QueryD1Database returned %+v, expected %+v", actual, expected)
		}
	}
}
