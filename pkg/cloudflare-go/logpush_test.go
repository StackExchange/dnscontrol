package cloudflare

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"testing"

	"github.com/goccy/go-json"

	"time"

	"github.com/stretchr/testify/assert"
)

const (
	jobID                       = 1
	serverLogpushJobDescription = `{
	"id": %d,
	"dataset": "http_requests",
	"kind": "",
	"enabled": false,
	"name": "example.com",
	"logpull_options": "fields=RayID,ClientIP,EdgeStartTimestamp&timestamps=rfc3339",
	"destination_conf": "s3://mybucket/logs?region=us-west-2",
	"last_complete": "%[2]s",
	"last_error": "%[2]s",
	"error_message": "test",
	"frequency": "high",
	"max_upload_bytes": 5000000
  }
`
	serverLogpushJobWithOutputOptionsDescription = `{
	"id": %d,
	"dataset": "http_requests",
	"kind": "",
	"enabled": false,
	"name": "example.com",
	"output_options": {
		"field_names":[
			"RayID",
			"ClientIP",
			"EdgeStartTimestamp"
		],
		"timestamp_format": "rfc3339"
	},
	"destination_conf": "s3://mybucket/logs?region=us-west-2",
	"last_complete": "%[2]s",
	"last_error": "%[2]s",
	"error_message": "test",
	"frequency": "high",
	"max_upload_bytes": 5000000
  }
`
	serverEdgeLogpushJobDescription = `{
	"id": %d,
	"dataset": "http_requests",
	"kind": "edge",
	"enabled": true,
	"name": "example.com",
	"logpull_options": "fields=RayID,ClientIP,EdgeStartTimestamp&timestamps=rfc3339",
	"destination_conf": "s3://mybucket/logs?region=us-west-2",
	"last_complete": "%[2]s",
	"last_error": "%[2]s",
	"error_message": "test",
	"frequency": "high"
  }
`
	serverLogpushGetOwnershipChallengeDescription = `{
	"filename": "logs/challenge-filename.txt",
	"valid": true,
	"message": ""
  }
`
	serverLogpushGetOwnershipChallengeInvalidResponseDescription = `{
	"filename": "logs/challenge-filename.txt",
	"valid": false,
	"message": "destination is invalid"
  }
`
)

var (
	testLogpushTimestamp     = time.Now().UTC()
	expectedLogpushJobStruct = LogpushJob{
		ID:              jobID,
		Dataset:         "http_requests",
		Enabled:         false,
		Name:            "example.com",
		LogpullOptions:  "fields=RayID,ClientIP,EdgeStartTimestamp&timestamps=rfc3339",
		DestinationConf: "s3://mybucket/logs?region=us-west-2",
		LastComplete:    &testLogpushTimestamp,
		LastError:       &testLogpushTimestamp,
		ErrorMessage:    "test",
		Frequency:       "high",
		MaxUploadBytes:  5000000,
	}
	expectedLogpushJobWithOutputOptionsStruct = LogpushJob{
		ID:      jobID,
		Dataset: "http_requests",
		Enabled: false,
		Name:    "example.com",
		OutputOptions: &LogpushOutputOptions{
			FieldNames: []string{
				"RayID",
				"ClientIP",
				"EdgeStartTimestamp",
			},
			TimestampFormat: "rfc3339",
		},
		DestinationConf: "s3://mybucket/logs?region=us-west-2",
		LastComplete:    &testLogpushTimestamp,
		LastError:       &testLogpushTimestamp,
		ErrorMessage:    "test",
		Frequency:       "high",
		MaxUploadBytes:  5000000,
	}
	expectedEdgeLogpushJobStruct = LogpushJob{
		ID:              jobID,
		Dataset:         "http_requests",
		Kind:            "edge",
		Enabled:         true,
		Name:            "example.com",
		LogpullOptions:  "fields=RayID,ClientIP,EdgeStartTimestamp&timestamps=rfc3339",
		DestinationConf: "s3://mybucket/logs?region=us-west-2",
		LastComplete:    &testLogpushTimestamp,
		LastError:       &testLogpushTimestamp,
		ErrorMessage:    "test",
		Frequency:       "high",
	}
	expectedLogpushGetOwnershipChallengeStruct = LogpushGetOwnershipChallenge{
		Filename: "logs/challenge-filename.txt",
		Valid:    true,
		Message:  "",
	}
)

func TestLogpushJobs(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
		  "result": [
			%s
		  ],
		  "success": true,
		  "errors": null,
		  "messages": null,
		  "result_info": {
			"page": 1,
			"per_page": 25,
			"count": 1,
			"total_count": 1
		  }
		}
		`, fmt.Sprintf(serverLogpushJobDescription, jobID, testLogpushTimestamp.Format(time.RFC3339Nano)))
	}

	mux.HandleFunc("/zones/"+testZoneID+"/logpush/jobs", handler)
	want := []LogpushJob{expectedLogpushJobStruct}

	actual, err := client.ListLogpushJobs(context.Background(), ZoneIdentifier(testZoneID), ListLogpushJobsParams{})
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestGetLogpushJob(t *testing.T) {
	testCases := map[string]struct {
		result string
		want   LogpushJob
	}{
		"core logpush job": {
			result: serverLogpushJobDescription,
			want:   expectedLogpushJobStruct,
		},
		"core logpush job with output options": {
			result: serverLogpushJobWithOutputOptionsDescription,
			want:   expectedLogpushJobWithOutputOptionsStruct,
		},
		"edge logpush job": {
			result: serverEdgeLogpushJobDescription,
			want:   expectedEdgeLogpushJobStruct,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			setup()
			defer teardown()

			handler := func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
				w.Header().Set("content-type", "application/json")
				fmt.Fprintf(w, `{
				  "result": %s,
				  "success": true,
				  "errors": null,
				  "messages": null
				}
				`, fmt.Sprintf(tc.result, jobID, testLogpushTimestamp.Format(time.RFC3339Nano)))
			}

			mux.HandleFunc("/zones/"+testZoneID+"/logpush/jobs/"+strconv.Itoa(jobID), handler)

			actual, err := client.GetLogpushJob(context.Background(), ZoneIdentifier(testZoneID), jobID)
			if assert.NoError(t, err) {
				assert.Equal(t, tc.want, actual)
			}
		})
	}
}

func TestCreateLogpushJob(t *testing.T) {
	testCases := map[string]struct {
		newJob  CreateLogpushJobParams
		payload string
		result  string
		want    LogpushJob
	}{
		"core logpush job": {
			newJob: CreateLogpushJobParams{
				Dataset:          "http_requests",
				Enabled:          false,
				Name:             "example.com",
				LogpullOptions:   "fields=RayID,ClientIP,EdgeStartTimestamp&timestamps=rfc3339",
				DestinationConf:  "s3://mybucket/logs?region=us-west-2",
				MaxUploadRecords: 1000,
			},
			payload: `{
				"dataset": "http_requests",
				"enabled":false,
				"name":"example.com",
				"logpull_options":"fields=RayID,ClientIP,EdgeStartTimestamp&timestamps=rfc3339",
				"destination_conf":"s3://mybucket/logs?region=us-west-2",
				"max_upload_records": 1000
			}`,
			result: serverLogpushJobDescription,
			want:   expectedLogpushJobStruct,
		},
		"core logpush job with output options": {
			newJob: CreateLogpushJobParams{
				Dataset: "http_requests",
				Enabled: false,
				Name:    "example.com",
				OutputOptions: &LogpushOutputOptions{
					FieldNames: []string{
						"RayID",
						"ClientIP",
						"EdgeStartTimestamp",
					},
					TimestampFormat: "rfc3339",
				},
				DestinationConf:  "s3://mybucket/logs?region=us-west-2",
				MaxUploadRecords: 1000,
			},
			payload: `{
				"dataset": "http_requests",
				"enabled":false,
				"name":"example.com",
				"output_options": {
					"field_names":[
						"RayID",
						"ClientIP",
						"EdgeStartTimestamp"
					],
					"timestamp_format": "rfc3339"
				},
				"destination_conf":"s3://mybucket/logs?region=us-west-2",
				"max_upload_records": 1000
			}`,
			result: serverLogpushJobWithOutputOptionsDescription,
			want:   expectedLogpushJobWithOutputOptionsStruct,
		},
		"edge logpush job": {
			newJob: CreateLogpushJobParams{
				Dataset:         "http_requests",
				Enabled:         true,
				Name:            "example.com",
				Kind:            "edge",
				LogpullOptions:  "fields=RayID,ClientIP,EdgeStartTimestamp&timestamps=rfc3339",
				DestinationConf: "s3://mybucket/logs?region=us-west-2",
			},
			payload: `{
				"dataset": "http_requests",
				"enabled":true,
				"name":"example.com",
				"kind":"edge",
				"logpull_options":"fields=RayID,ClientIP,EdgeStartTimestamp&timestamps=rfc3339",
				"destination_conf":"s3://mybucket/logs?region=us-west-2"
			}`,
			result: serverEdgeLogpushJobDescription,
			want:   expectedEdgeLogpushJobStruct,
		},
		"filtered edge logpush job": {
			newJob: CreateLogpushJobParams{
				Dataset:         "http_requests",
				Enabled:         true,
				Name:            "example.com",
				Kind:            "edge",
				LogpullOptions:  "fields=RayID,ClientIP,EdgeStartTimestamp&timestamps=rfc3339",
				DestinationConf: "s3://mybucket/logs?region=us-west-2",
				Filter: &LogpushJobFilters{
					Where: LogpushJobFilter{Key: "ClientRequestHost", Operator: "eq", Value: "example.com"},
				},
			},
			payload: `{
				"dataset": "http_requests",
				"enabled":true,
				"name":"example.com",
				"kind":"edge",
				"logpull_options":"fields=RayID,ClientIP,EdgeStartTimestamp&timestamps=rfc3339",
				"destination_conf":"s3://mybucket/logs?region=us-west-2",
				"filter":"{\"where\":{\"key\":\"ClientRequestHost\",\"operator\":\"eq\",\"value\":\"example.com\"}}"
			}`,
			result: serverEdgeLogpushJobDescription,
			want:   expectedEdgeLogpushJobStruct,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			setup()
			defer teardown()

			handler := func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
				b, err := io.ReadAll(r.Body)
				defer r.Body.Close()

				if assert.NoError(t, err) {
					assert.JSONEq(t, tc.payload, string(b), "JSON payload not equal")
				}

				w.Header().Set("content-type", "application/json")
				fmt.Fprintf(w, `{
				"result": %s,
				"success": true,
				"errors": null,
				"messages": null
				}
				`, fmt.Sprintf(tc.result, jobID, testLogpushTimestamp.Format(time.RFC3339Nano)))
			}

			mux.HandleFunc("/zones/"+testZoneID+"/logpush/jobs", handler)

			actual, err := client.CreateLogpushJob(context.Background(), ZoneIdentifier(testZoneID), tc.newJob)
			if assert.NoError(t, err) {
				assert.Equal(t, tc.want, *actual)
			}
		})
	}
}

func TestUpdateLogpushJob(t *testing.T) {
	setup()
	defer teardown()
	updatedJob := UpdateLogpushJobParams{
		ID:              jobID,
		Enabled:         true,
		Name:            "updated.com",
		LogpullOptions:  "fields=RayID,ClientIP,EdgeStartTimestamp",
		DestinationConf: "gs://mybucket/logs",
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
		  "result": %s,
		  "success": true,
		  "errors": null,
		  "messages": null
		}
		`, fmt.Sprintf(serverLogpushJobDescription, jobID, testLogpushTimestamp.Format(time.RFC3339Nano)))
	}

	mux.HandleFunc("/zones/"+testZoneID+"/logpush/jobs/"+strconv.Itoa(jobID), handler)

	err := client.UpdateLogpushJob(context.Background(), ZoneIdentifier(testZoneID), updatedJob)
	assert.NoError(t, err)
}

func TestDeleteLogpushJob(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
		  "result": null,
		  "success": true,
		  "errors": null,
		  "messages": null
		}
		`)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/logpush/jobs/"+strconv.Itoa(jobID), handler)

	err := client.DeleteLogpushJob(context.Background(), ZoneIdentifier(testZoneID), jobID)
	assert.NoError(t, err)
}

func TestGetLogpushOwnershipChallenge(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
		  "result": %s,
		  "success": true,
		  "errors": null,
		  "messages": null
		}
		`, serverLogpushGetOwnershipChallengeDescription)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/logpush/ownership", handler)

	want := &expectedLogpushGetOwnershipChallengeStruct

	actual, err := client.GetLogpushOwnershipChallenge(context.Background(), ZoneIdentifier(testZoneID), GetLogpushOwnershipChallengeParams{DestinationConf: "destination_conf"})
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestGetLogpushOwnershipChallengeWithInvalidResponse(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
		  "result": %s,
		  "success": true,
		  "errors": null,
		  "messages": null
		}
		`, serverLogpushGetOwnershipChallengeInvalidResponseDescription)
	}

	mux.HandleFunc("/zones/"+testZoneID+"/logpush/ownership", handler)
	_, err := client.GetLogpushOwnershipChallenge(context.Background(), ZoneIdentifier(testZoneID), GetLogpushOwnershipChallengeParams{DestinationConf: "destination_conf"})

	assert.Error(t, err)
}

func TestValidateLogpushOwnershipChallenge(t *testing.T) {
	testCases := map[string]struct {
		isValid bool
	}{
		"ownership is valid": {
			isValid: true,
		},
		"ownership is not valid": {
			isValid: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			setup()
			defer teardown()

			handler := func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
				w.Header().Set("content-type", "application/json")
				fmt.Fprintf(w, `{
				  "result": {
					"valid": %v
				  },
				  "success": true,
				  "errors": null,
				  "messages": null
				}
				`, tc.isValid)
			}

			mux.HandleFunc("/zones/"+testZoneID+"/logpush/ownership/validate", handler)

			actual, err := client.ValidateLogpushOwnershipChallenge(context.Background(), ZoneIdentifier(testZoneID), ValidateLogpushOwnershipChallengeParams{
				DestinationConf:    "destination_conf",
				OwnershipChallenge: "ownership_challenge",
			})
			if assert.NoError(t, err) {
				assert.Equal(t, tc.isValid, actual)
			}
		})
	}
}

func TestCheckLogpushDestinationExists(t *testing.T) {
	testCases := map[string]struct {
		exists bool
	}{
		"destination exists": {
			exists: true,
		},
		"destination does not exists": {
			exists: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			setup()
			defer teardown()

			handler := func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
				w.Header().Set("content-type", "application/json")
				fmt.Fprintf(w, `{
				  "result": {
					"exists": %v
				  },
				  "success": true,
				  "errors": null,
				  "messages": null
				}
				`, tc.exists)
			}

			mux.HandleFunc("/zones/"+testZoneID+"/logpush/validate/destination/exists", handler)

			actual, err := client.CheckLogpushDestinationExists(context.Background(), ZoneIdentifier(testZoneID), "destination_conf")
			if assert.NoError(t, err) {
				assert.Equal(t, tc.exists, actual)
			}
		})
	}
}

var (
	validFilter LogpushJobFilter = LogpushJobFilter{Key: "ClientRequestPath", Operator: Contains, Value: "static"}
)

var logpushJobFiltersTest = []struct {
	name                 string
	input                LogpushJobFilter
	haserror             bool
	expectedErrorMessage string
}{
	// Tests without And or Or
	{"Empty Filter", LogpushJobFilter{}, true, "Key is missing"},
	{"Missing Operator", LogpushJobFilter{Key: "ClientRequestPath"}, true, "Operator is missing"},
	{"Missing Value", LogpushJobFilter{Key: "ClientRequestPath", Operator: Contains}, true, "Value is missing"},
	{"Valid Basic Filter", validFilter, false, ""},
	// Tests with And
	{"Valid And Filter", LogpushJobFilter{And: []LogpushJobFilter{validFilter}}, false, ""},
	{"And and Or", LogpushJobFilter{And: []LogpushJobFilter{validFilter}, Or: []LogpushJobFilter{validFilter}}, true, "And can't be set with Or, Key, Operator or Value"},
	{"And and Key", LogpushJobFilter{And: []LogpushJobFilter{validFilter}, Key: "Key"}, true, "And can't be set with Or, Key, Operator or Value"},
	{"And and Operator", LogpushJobFilter{And: []LogpushJobFilter{validFilter}, Operator: Contains}, true, "And can't be set with Or, Key, Operator or Value"},
	{"And and Value", LogpushJobFilter{And: []LogpushJobFilter{validFilter}, Value: "Value"}, true, "And can't be set with Or, Key, Operator or Value"},
	{"And with nested error", LogpushJobFilter{And: []LogpushJobFilter{validFilter, {}}}, true, "element 1 in And is invalid: Key is missing"},
	// Tests with Or
	{"Valid Or Filter", LogpushJobFilter{Or: []LogpushJobFilter{validFilter}}, false, ""},
	{"Or and Key", LogpushJobFilter{Or: []LogpushJobFilter{validFilter}, Key: "Key"}, true, "Or can't be set with And, Key, Operator or Value"},
	{"Or and Operator", LogpushJobFilter{Or: []LogpushJobFilter{validFilter}, Operator: Contains}, true, "Or can't be set with And, Key, Operator or Value"},
	{"Or and Value", LogpushJobFilter{Or: []LogpushJobFilter{validFilter}, Value: "Value"}, true, "Or can't be set with And, Key, Operator or Value"},
	{"Or with nested error", LogpushJobFilter{Or: []LogpushJobFilter{validFilter, {}}}, true, "element 1 in Or is invalid: Key is missing"},
}

func TestLogpushJobFilter_Validate(t *testing.T) {
	for _, tt := range logpushJobFiltersTest {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.Validate()
			if tt.haserror {
				assert.ErrorContains(t, got, tt.expectedErrorMessage)
			} else {
				assert.NoError(t, got)
			}
		})
	}
}

func TestLogpushJob_Unmarshall(t *testing.T) {
	t.Run("Valid Filter", func(t *testing.T) {
		jsonstring := `{"filter":"{\"where\":{\"and\":[{\"key\":\"ClientRequestPath\",\"operator\":\"contains\",\"value\":\"/static\\\\\"},{\"key\":\"ClientRequestHost\",\"operator\":\"eq\",\"value\":\"example.com\"}]}}","dataset":"http_requests","enabled":false,"name":"example.com static assets","logpull_options":"fields=RayID,ClientIP,EdgeStartTimestamp\u0026timestamps=rfc3339\u0026CVE-2021-44228=true","destination_conf":"s3://\u003cBUCKET_PATH\u003e?region=us-west-2/"}`
		var job LogpushJob
		if err := json.Unmarshal([]byte(jsonstring), &job); err != nil {
			log.Fatal(err)
		}

		assert.Equal(t, LogpushJob{
			Name:            "example.com static assets",
			LogpullOptions:  "fields=RayID,ClientIP,EdgeStartTimestamp&timestamps=rfc3339&CVE-2021-44228=true",
			Dataset:         "http_requests",
			DestinationConf: "s3://<BUCKET_PATH>?region=us-west-2/",
			Filter: &LogpushJobFilters{
				Where: LogpushJobFilter{
					And: []LogpushJobFilter{
						{Key: "ClientRequestPath", Operator: Contains, Value: "/static\\"},
						{Key: "ClientRequestHost", Operator: Equal, Value: "example.com"},
					},
				},
			},
		}, job)
	})

	t.Run("Invalid Filter", func(t *testing.T) {
		jsonstring := `{"filter":"{\"where\":{\"and\":[{\"key\":\"ClientRequestPath\",\"operator\":\"contains\"},{\"key\":\"ClientRequestHost\",\"operator\":\"eq\",\"value\":\"example.com\"}]}}","dataset":"http_requests","enabled":false,"name":"example.com static assets","logpull_options":"fields=RayID,ClientIP,EdgeStartTimestamp\u0026timestamps=rfc3339\u0026CVE-2021-44228=true","destination_conf":"s3://\u003cBUCKET_PATH\u003e?region=us-west-2/"}`
		var job LogpushJob
		err := json.Unmarshal([]byte(jsonstring), &job)

		assert.ErrorContains(t, err, "element 0 in And is invalid: Value is missing")
	})

	t.Run("No Filter", func(t *testing.T) {
		jsonstring := `{"dataset":"http_requests","enabled":false,"name":"example.com static assets","logpull_options":"fields=RayID,ClientIP,EdgeStartTimestamp\u0026timestamps=rfc3339\u0026CVE-2021-44228=true","destination_conf":"s3://\u003cBUCKET_PATH\u003e?region=us-west-2/"}`
		var job LogpushJob
		if err := json.Unmarshal([]byte(jsonstring), &job); err != nil {
			log.Fatal(err)
		}

		assert.Equal(t, LogpushJob{
			Name:            "example.com static assets",
			LogpullOptions:  "fields=RayID,ClientIP,EdgeStartTimestamp&timestamps=rfc3339&CVE-2021-44228=true",
			Dataset:         "http_requests",
			DestinationConf: "s3://<BUCKET_PATH>?region=us-west-2/",
		}, job)
	})
}

func TestLogPushJob_Marshall(t *testing.T) {
	testCases := []struct {
		job  LogpushJob
		want string
	}{
		{
			job: LogpushJob{
				Dataset:         "http_requests",
				Name:            "valid filter",
				LogpullOptions:  "fields=RayID,ClientIP,EdgeStartTimestamp&timestamps=rfc3339",
				DestinationConf: "https://example.com",
				Filter: &LogpushJobFilters{
					Where: LogpushJobFilter{Key: "ClientRequestHost", Operator: Equal, Value: "example.com"},
				},
			},
			want: `{
				"dataset": "http_requests",
				"enabled": false,
				"name": "valid filter",
				"logpull_options": "fields=RayID,ClientIP,EdgeStartTimestamp&timestamps=rfc3339",
				"destination_conf": "https://example.com",
				"filter":"{\"where\":{\"key\":\"ClientRequestHost\",\"operator\":\"eq\",\"value\":\"example.com\"}}"
			}`,
		},
		{
			job: LogpushJob{
				Dataset:         "http_requests",
				Name:            "no filter",
				LogpullOptions:  "fields=RayID,ClientIP,EdgeStartTimestamp&timestamps=rfc3339",
				DestinationConf: "https://example.com",
			},
			want: `{
				"dataset": "http_requests",
				"enabled": false,
				"name": "no filter",
				"logpull_options": "fields=RayID,ClientIP,EdgeStartTimestamp&timestamps=rfc3339",
				"destination_conf": "https://example.com"
			}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.job.Name, func(t *testing.T) {
			got, err := json.Marshal(tc.job)

			if assert.NoError(t, err) {
				assert.JSONEq(t, tc.want, string(got))
			}
		})
	}
}
