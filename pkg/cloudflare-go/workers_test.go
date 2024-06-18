package cloudflare

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/goccy/go-json"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	deleteWorkerResponseData = `{
    "result": null,
    "success": true,
    "errors": [],
    "messages": []
}`

	updateWorkerRouteResponse = `{
    "result": {
        "id": "e7a57d8746e74ae49c25994dadb421b1",
        "pattern": "app3.example.com/*",
        "enabled": true
    },
    "success": true,
    "errors": [],
    "messages": []
}`
	updateWorkerRouteEntResponse = `{
    "result": {
        "id": "e7a57d8746e74ae49c25994dadb421b1",
        "pattern": "app3.example.com/*",
        "script": "test_script_1"
    },
    "success": true,
    "errors": [],
    "messages": []
}`
	createWorkerRouteResponse = `{
    "result": {
        "id": "e7a57d8746e74ae49c25994dadb421b1"
    },
    "success": true,
    "errors": [],
    "messages": []
}`
	listRouteResponseData = `{
    "result": [
        {
            "id": "e7a57d8746e74ae49c25994dadb421b1",
            "pattern": "app1.example.com/*",
            "enabled": true
        },
        {
            "id": "f8b68e9857f85bf59c25994dadb421b1",
            "pattern": "app2.example.com/*",
            "enabled": false
        }
    ],
    "success": true,
    "errors": [],
    "messages": []
}`
	listWorkerRouteResponse = `{
    "result": [
        {
            "id": "e7a57d8746e74ae49c25994dadb421b1",
            "pattern": "app1.example.com/*",
            "script": "test_script_1"
        },
        {
            "id": "f8b68e9857f85bf59c25994dadb421b1",
            "pattern": "app2.example.com/*",
            "script": "test_script_2"
        },
        {
            "id": "2b5bf4240cd34c77852fac70b1bf745a",
            "pattern": "app3.example.com/*",
			"script": "test_script_3"
        }
    ],
    "success": true,
    "errors": [],
    "messages": []
}`
	getRouteResponseData = `{
    "result": {
       "id": "e7a57d8746e74ae49c25994dadb421b1",
       "pattern": "app1.example.com/*",
       "script": "script-name"
    },
    "success": true,
    "errors": [],
    "messages": []
}`
	listBindingsResponseData = `{
		"result": [
			{
				"name": "MY_KV",
				"namespace_id": "89f5f8fd93f94cb98473f6f421aa3b65",
				"type": "kv_namespace"
			},
			{
				"name": "MY_WASM",
				"type": "wasm_module"
			},
			{
				"name": "MY_PLAIN_TEXT",
				"type": "plain_text",
				"text": "text"
			},
			{
				"name": "MY_SECRET_TEXT",
				"type": "secret_text"
			},
			{
				"name": "MY_SERVICE_BINDING",
				"type": "service",
				"service": "MY_SERVICE",
				"environment": "MY_ENVIRONMENT"
			},
			{
				"name": "MY_NEW_BINDING",
				"type": "some_imaginary_new_binding_type"
			},
			{
				"name": "MY_BUCKET",
				"type": "r2_bucket",
				"bucket_name": "bucket"
			},
			{
				"name": "MY_DATASET",
				"type": "analytics_engine",
				"dataset": "my_dataset"
			},
			{
				"name": "MY_DATABASE",
				"type": "d1",
				"database_id": "cef5331f-e5c7-4c8a-a415-7908ae45f92a"
			}
		],
		"success": true,
		"errors": [],
		"messages": []
	}`
	listWorkersResponseData = `{
  "result": [
    {
      "id": "bar",
      "created_on": "2018-04-22T17:10:48.938097Z",
      "modified_on": "2018-04-22T17:10:48.938097Z",
      "etag": "279cf40d86d70b82f6cd3ba90a646b3ad995912da446836d7371c21c6a43977a"
    },
    {
      "id": "baz",
      "created_on": "2018-04-22T17:10:48.938097Z",
      "modified_on": "2018-04-22T17:10:48.938097Z",
      "etag": "380dg51e97e80b82f6cd3ba90a646b3ad995912da446836d7371c21c6a43088b"
    }
  ],
  "success": true,
  "errors": [],
  "messages": []
}`
	workerMetadata = `{
		"id": "e7a57d8746e74ae49c25994dadb421b1",
		"etag": "279cf40d86d70b82f6cd3ba90a646b3ad995912da446836d7371c21c6a43977a",
		"logpush": true
	}`
	workerScript = `addEventListener('fetch', event => {
  event.passThroughOnException()
  event.respondWith(handleRequest(event.request))
})

async function handleRequest(request) {
  return fetch(request)
}`
	workerModuleScript = `export default {
  async fetch(request, env, event) {
    event.passThroughOnException()
    return fetch(request)
  }
}`
	workerModuleScriptDownloadResponse = `
--workermodulescriptdownload
Content-Disposition: form-data; name="worker.js"

export default {
  async fetch(request, env, event) {
    event.passThroughOnException()
    return fetch(request)
  }
}
--workermodulescriptdownload--
`
)

var (
	successResponse               = Response{Success: true, Errors: []ResponseInfo{}, Messages: []ResponseInfo{}}
	deleteWorkerRouteResponseData = createWorkerRouteResponse
	attachWorkerToDomainResponse  = fmt.Sprintf(`{
    "result": {
        "id": "e7a57d8746e74ae49c25994dadb421b1",
	"zone_id": "%s",
	"service":"test_script_1",
	"hostname":"api4.example.com",
	"environment":"production"
    },
    "success": true,
    "errors": [],
    "messages": []
}`, testZoneID)
)

type (
	WorkersTestScriptResponse struct {
		Script            string                 `json:"script"`
		UsageModel        string                 `json:"usage_model,omitempty"`
		Handlers          []string               `json:"handlers"`
		ID                string                 `json:"id,omitempty"`
		ETAG              string                 `json:"etag,omitempty"`
		Size              uint                   `json:"size,omitempty"`
		CreatedOn         string                 `json:"created_on,omitempty"`
		ModifiedOn        string                 `json:"modified_on,omitempty"`
		LastDeployedFrom  *string                `json:"last_deployed_from,omitempty"`
		DeploymentId      *string                `json:"deployment_id,omitempty"`
		CompatibilityDate *string                `json:"compatibility_date,omitempty"`
		Logpush           *bool                  `json:"logpush,omitempty"`
		TailConsumers     *[]WorkersTailConsumer `json:"tail_consumers,omitempty"`
		PlacementMode     *string                `json:"placement_mode,omitempty"`
	}
	workersTestResponseOpt func(r *WorkersTestScriptResponse)
)

var (
	expectedWorkersServiceWorkerScript = "addEventListener('fetch', event => {\n  event.passThroughOnException()\n  event.respondWith(handleRequest(event.request))\n})\n\nasync function handleRequest(request) {\n  return fetch(request)\n}"
	expectedWorkersModuleWorkerScript  = "export default {\n  async fetch(request, env, event) {\n    event.passThroughOnException()\n    return fetch(request)\n  }\n}"
	WorkersDefaultTestResponse         = WorkersTestScriptResponse{
		Script:            expectedWorkersServiceWorkerScript,
		Handlers:          []string{"fetch"},
		UsageModel:        "unbound",
		ID:                "e7a57d8746e74ae49c25994dadb421b1",
		ETAG:              "279cf40d86d70b82f6cd3ba90a646b3ad995912da446836d7371c21c6a43977a",
		Size:              191,
		LastDeployedFrom:  StringPtr("dash"),
		Logpush:           BoolPtr(false),
		CompatibilityDate: StringPtr("2022-07-12"),
	}
)

//nolint:unused
func withWorkerScript(content string) workersTestResponseOpt {
	return func(r *WorkersTestScriptResponse) { r.Script = content }
}

//nolint:unused
func withWorkerUsageModel(um string) workersTestResponseOpt {
	return func(r *WorkersTestScriptResponse) { r.UsageModel = um }
}

//nolint:unused
func withWorkerHandlers(h []string) workersTestResponseOpt {
	return func(r *WorkersTestScriptResponse) { r.Handlers = h }
}

//nolint:unused
func withWorkerID(id string) workersTestResponseOpt {
	return func(r *WorkersTestScriptResponse) { r.ID = id }
}

//nolint:unused
func withWorkerEtag(etag string) workersTestResponseOpt {
	return func(r *WorkersTestScriptResponse) { r.ETAG = etag }
}

//nolint:unused
func withWorkerSize(size uint) workersTestResponseOpt {
	return func(r *WorkersTestScriptResponse) { r.Size = size }
}

//nolint:unused
func withWorkerCreatedOn(co time.Time) workersTestResponseOpt {
	return func(r *WorkersTestScriptResponse) { r.CreatedOn = co.Format(time.RFC3339Nano) }
}

//nolint:unused
func withWorkerModifiedOn(mo time.Time) workersTestResponseOpt {
	return func(r *WorkersTestScriptResponse) { r.ModifiedOn = mo.Format(time.RFC3339Nano) }
}

//nolint:unused
func withWorkerLogpush(logpush *bool) workersTestResponseOpt {
	return func(r *WorkersTestScriptResponse) { r.Logpush = logpush }
}

//nolint:unused
func withWorkerPlacementMode(mode *string) workersTestResponseOpt {
	return func(r *WorkersTestScriptResponse) { r.PlacementMode = mode }
}

//nolint:unused
func withWorkerTailConsumers(consumers ...WorkersTailConsumer) workersTestResponseOpt {
	return func(r *WorkersTestScriptResponse) { r.TailConsumers = &consumers }
}

//nolint:unused
func withWorkerLastDeployedFrom(from *string) workersTestResponseOpt {
	return func(r *WorkersTestScriptResponse) { r.LastDeployedFrom = from }
}

//nolint:unused
func withWorkerDeploymentId(dID *string) workersTestResponseOpt {
	return func(r *WorkersTestScriptResponse) { r.DeploymentId = dID }
}

func workersScriptResponse(t testing.TB, opts ...workersTestResponseOpt) string {
	var responseConfig = WorkersDefaultTestResponse
	for _, opt := range opts {
		opt(&responseConfig)
	}

	bytes, err := json.Marshal(struct {
		Response
		Result WorkersTestScriptResponse `json:"result"`
	}{
		Response: Response{Success: true, Errors: []ResponseInfo{}, Messages: []ResponseInfo{}},
		Result:   responseConfig,
	})
	require.NoError(t, err)

	return string(bytes)
}

func getFormValue(r *http.Request, key string) ([]byte, error) {
	err := r.ParseMultipartForm(1024 * 1024)
	if err != nil {
		return nil, err
	}

	// In Go 1.10 there was a bug where field values with a content-type
	// but without a filename would end up in Form.File but in versions
	// before and after 1.10 they would be in form.Value. Here we check
	// both in order to handle both scenarios
	// https://golang.org/doc/go1.11#mime/multipart

	// pre/post v1.10
	if values, ok := r.MultipartForm.Value[key]; ok {
		return []byte(values[0]), nil
	}

	// v1.10
	if fileHeaders, ok := r.MultipartForm.File[key]; ok {
		file, err := fileHeaders[0].Open()
		if err != nil {
			return nil, err
		}
		return io.ReadAll(file)
	}

	return nil, fmt.Errorf("no value found for key %v", key)
}

func getFileDetails(r *http.Request, key string) (*multipart.FileHeader, error) {
	err := r.ParseMultipartForm(1024 * 1024)
	if err != nil {
		return nil, err
	}

	fileHeaders := r.MultipartForm.File[key]

	if len(fileHeaders) > 0 {
		return fileHeaders[0], nil
	}

	return nil, fmt.Errorf("no value found for key %v", key)
}

type multipartUpload struct {
	Script             string
	BindingMeta        map[string]workerBindingMeta
	Logpush            *bool
	CompatibilityDate  string
	CompatibilityFlags []string
	Placement          *Placement
	Tags               []string
}

func parseMultipartUpload(r *http.Request) (multipartUpload, error) {
	// Parse the metadata
	mdBytes, err := getFormValue(r, "metadata")
	if err != nil {
		return multipartUpload{}, err
	}

	var metadata struct {
		BodyPart           string              `json:"body_part,omitempty"`
		MainModule         string              `json:"main_module,omitempty"`
		Bindings           []workerBindingMeta `json:"bindings"`
		Logpush            *bool               `json:"logpush,omitempty"`
		CompatibilityDate  string              `json:"compatibility_date,omitempty"`
		CompatibilityFlags []string            `json:"compatibility_flags,omitempty"`
		Placement          *Placement          `json:"placement,omitempty"`
		Tags               []string            `json:"tags"`
	}
	err = json.Unmarshal(mdBytes, &metadata)
	if err != nil {
		return multipartUpload{}, err
	}

	// Get the script
	script, err := getFormValue(r, metadata.BodyPart)
	if err != nil {
		script, err = getFormValue(r, metadata.MainModule)

		if err != nil {
			return multipartUpload{}, err
		}
	}

	// Since bindings are specified in the Go API as a map but are uploaded as a
	// JSON array, the ordering of uploaded bindings is non-deterministic. To make
	// it easier to compare for equality without running into ordering issues, we
	// convert it back to a map
	bindingMeta := make(map[string]workerBindingMeta)
	for _, binding := range metadata.Bindings {
		bindingMeta[binding["name"].(string)] = binding
	}

	return multipartUpload{
		Script:             string(script),
		BindingMeta:        bindingMeta,
		Logpush:            metadata.Logpush,
		CompatibilityDate:  metadata.CompatibilityDate,
		CompatibilityFlags: metadata.CompatibilityFlags,
		Placement:          metadata.Placement,
		Tags:               metadata.Tags,
	}, nil
}

func TestDeleteWorker(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/workers/scripts/bar", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/javascript")
		fmt.Fprint(w, deleteWorkerResponseData)
	})

	err := client.DeleteWorker(context.Background(), AccountIdentifier(testAccountID), DeleteWorkerParams{ScriptName: "bar"})
	assert.NoError(t, err)
}

func TestDeleteNamespacedWorker(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/workers/dispatch/namespaces/foo/scripts/bar", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/javascript")
		fmt.Fprint(w, deleteWorkerResponseData)
	})

	err := client.DeleteWorker(context.Background(), AccountIdentifier(testAccountID), DeleteWorkerParams{
		ScriptName:        "bar",
		DispatchNamespace: &[]string{"foo"}[0],
	})
	assert.NoError(t, err)
}

func TestGetWorker(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/workers/scripts/foo", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/javascript")
		fmt.Fprint(w, workerScript)
	})
	res, err := client.GetWorker(context.Background(), AccountIdentifier(testAccountID), "foo")
	want := WorkerScriptResponse{
		successResponse,
		false,
		WorkerScript{
			Script: workerScript,
		}}
	if assert.NoError(t, err) {
		assert.Equal(t, want.Script, res.Script)
	}
}

func TestGetWorker_Module(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/workers/scripts/foo", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "multipart/form-data; boundary=workermodulescriptdownload")
		fmt.Fprint(w, workerModuleScriptDownloadResponse)
	})

	res, err := client.GetWorker(context.Background(), AccountIdentifier(testAccountID), "foo")
	want := WorkerScriptResponse{
		successResponse,
		true,
		WorkerScript{
			Script: workerModuleScript,
		},
	}

	if assert.NoError(t, err) {
		assert.Equal(t, want.Script, res.Script)
	}
}

func TestGetWorkerWithDispatchNamespace_Module(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/workers/dispatch/namespaces/bar/scripts/foo/content", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "multipart/form-data; boundary=workermodulescriptdownload")
		fmt.Fprint(w, workerModuleScriptDownloadResponse)
	})

	res, err := client.GetWorkerWithDispatchNamespace(context.Background(), AccountIdentifier(testAccountID), "foo", "bar")
	want := WorkerScriptResponse{
		successResponse,
		true,
		WorkerScript{
			Script: workerModuleScript,
		},
	}

	if assert.NoError(t, err) {
		assert.Equal(t, want.Script, res.Script)
	}
}

func TestGetWorkersScriptContent(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/workers/scripts/foo/content/v2", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/javascript")
		fmt.Fprint(w, workerScript)
	})

	res, err := client.GetWorkersScriptContent(context.Background(), AccountIdentifier(testAccountID), "foo")
	want := workerScript
	if assert.NoError(t, err) {
		assert.Equal(t, want, res)
	}
}

func TestUpdateWorkersScriptContent(t *testing.T) {
	setup()
	defer teardown()

	formattedTime, _ := time.Parse(time.RFC3339Nano, "2018-06-09T15:17:01.989141Z")
	mux.HandleFunc("/accounts/"+testAccountID+"/workers/scripts/foo/content", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		contentTypeHeader := r.Header.Get("content-type")
		assert.Equal(t, "application/javascript", contentTypeHeader, "Expected content-type request header to be 'application/javascript', got %s", contentTypeHeader)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, workersScriptResponse(t, withWorkerModifiedOn(formattedTime)))
	})

	res, err := client.UpdateWorkersScriptContent(context.Background(), AccountIdentifier(testAccountID), UpdateWorkersScriptContentParams{ScriptName: "foo", Script: workerScript})
	want := WorkerScriptResponse{
		successResponse,
		false,
		WorkerScript{
			Script: workerScript,
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, want.Script, res.Script)
	}
}

func TestGetWorkersScriptSettings(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/workers/scripts/foo/settings", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/javascript")
		fmt.Fprint(w, workerMetadata)
	})

	res, err := client.GetWorkersScriptSettings(context.Background(), AccountIdentifier(testAccountID), "foo")
	logpush := true
	want := WorkerScriptSettingsResponse{
		successResponse,
		WorkerMetaData{
			ID:      "e7a57d8746e74ae49c25994dadb421b1",
			ETAG:    "279cf40d86d70b82f6cd3ba90a646b3ad995912da446836d7371c21c6a43977a",
			Logpush: &logpush,
		}}
	if assert.NoError(t, err) {
		assert.Equal(t, want.WorkerMetaData, res.WorkerMetaData)
	}
}

func TestUpdateWorkersScriptSettings(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/workers/scripts/foo/settings", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)
		w.Header().Set("content-type", "application/javascript")
		fmt.Fprint(w, workerMetadata)
	})

	res, err := client.UpdateWorkersScriptSettings(context.Background(), AccountIdentifier(testAccountID), UpdateWorkersScriptSettingsParams{ScriptName: "foo"})
	logpush := true
	want := WorkerScriptSettingsResponse{
		successResponse,
		WorkerMetaData{
			ID:      "e7a57d8746e74ae49c25994dadb421b1",
			ETAG:    "279cf40d86d70b82f6cd3ba90a646b3ad995912da446836d7371c21c6a43977a",
			Logpush: &logpush,
		}}
	if assert.NoError(t, err) {
		assert.Equal(t, want.WorkerMetaData, res.WorkerMetaData)
	}
}

func TestListWorkers(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/workers/scripts", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, listWorkersResponseData)
	})

	res, _, err := client.ListWorkers(context.Background(), AccountIdentifier(testAccountID), ListWorkersParams{})
	sampleDate, _ := time.Parse(time.RFC3339Nano, "2018-04-22T17:10:48.938097Z")
	want := []WorkerMetaData{
		{
			ID:         "bar",
			ETAG:       "279cf40d86d70b82f6cd3ba90a646b3ad995912da446836d7371c21c6a43977a",
			CreatedOn:  sampleDate,
			ModifiedOn: sampleDate,
		},
		{
			ID:         "baz",
			ETAG:       "380dg51e97e80b82f6cd3ba90a646b3ad995912da446836d7371c21c6a43088b",
			CreatedOn:  sampleDate,
			ModifiedOn: sampleDate,
		},
	}
	if assert.NoError(t, err) {
		assert.Equal(t, want, res.WorkerList)
	}
}

func TestUploadWorker_Basic(t *testing.T) {
	setup()
	defer teardown()

	formattedTime, _ := time.Parse(time.RFC3339Nano, "2018-06-09T15:17:01.989141Z")
	mux.HandleFunc("/accounts/"+testAccountID+"/workers/scripts/foo", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		contentTypeHeader := r.Header.Get("content-type")
		assert.Equal(t, "application/javascript", contentTypeHeader, "Expected content-type request header to be 'application/javascript', got %s", contentTypeHeader)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, workersScriptResponse(t, withWorkerModifiedOn(formattedTime)))
	})
	res, err := client.UploadWorker(context.Background(), AccountIdentifier(testAccountID), CreateWorkerParams{ScriptName: "foo", Script: workerScript})
	want := WorkerScriptResponse{
		successResponse,
		false,
		WorkerScript{
			Script:     workerScript,
			UsageModel: "unbound",
			WorkerMetaData: WorkerMetaData{
				ID:               "e7a57d8746e74ae49c25994dadb421b1",
				ETAG:             "279cf40d86d70b82f6cd3ba90a646b3ad995912da446836d7371c21c6a43977a",
				Size:             191,
				ModifiedOn:       formattedTime,
				Logpush:          BoolPtr(false),
				LastDeployedFrom: StringPtr("dash"),
			},
		}}
	if assert.NoError(t, err) {
		assert.Equal(t, want, res)
	}
}

func TestUploadWorker_Module(t *testing.T) {
	setup()
	defer teardown()

	formattedCreatedTime, _ := time.Parse(time.RFC3339Nano, "2018-06-09T15:17:01.989141Z")
	mux.HandleFunc("/accounts/"+testAccountID+"/workers/scripts/foo", func(w http.ResponseWriter, r *http.Request) {
		mpUpload, err := parseMultipartUpload(r)
		assert.NoError(t, err)

		assert.Equal(t, workerModuleScript, mpUpload.Script)

		workerFileDetails, err := getFileDetails(r, "worker.mjs")
		if !assert.NoError(t, err) {
			assert.FailNow(t, "worker file not found in multipart form body")
		}
		contentTypeHeader := workerFileDetails.Header.Get("content-type")
		expectedContentType := "application/javascript+module"
		assert.Equal(t, expectedContentType, contentTypeHeader, "Expected content-type request header to be %s, got %s", expectedContentType, contentTypeHeader)

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, workersScriptResponse(t, withWorkerScript(expectedWorkersModuleWorkerScript), withWorkerCreatedOn(formattedCreatedTime)))
	})
	res, err := client.UploadWorker(context.Background(), AccountIdentifier(testAccountID), CreateWorkerParams{ScriptName: "foo", Script: workerModuleScript, Module: true})
	want := WorkerScriptResponse{
		Response: successResponse,
		Module:   false,
		WorkerScript: WorkerScript{
			Script:     workerModuleScript,
			UsageModel: "unbound",
			WorkerMetaData: WorkerMetaData{
				ID:               "e7a57d8746e74ae49c25994dadb421b1",
				ETAG:             "279cf40d86d70b82f6cd3ba90a646b3ad995912da446836d7371c21c6a43977a",
				Size:             191,
				CreatedOn:        formattedCreatedTime,
				Logpush:          BoolPtr(false),
				LastDeployedFrom: StringPtr("dash"),
			},
		}}
	if assert.NoError(t, err) {
		assert.Equal(t, want, res)
	}
}

func TestUploadWorker_WithDurableObjectBinding(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)

		mpUpload, err := parseMultipartUpload(r)
		assert.NoError(t, err)

		expectedBindings := map[string]workerBindingMeta{
			"b1": {
				"name":        "b1",
				"type":        "durable_object_namespace",
				"class_name":  "TheClass",
				"script_name": "the_script",
			},
		}
		assert.Equal(t, workerScript, mpUpload.Script)
		assert.Equal(t, expectedBindings, mpUpload.BindingMeta)

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, workersScriptResponse(t))
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/workers/scripts/bar", handler)

	_, err := client.UploadWorker(context.Background(), AccountIdentifier(testAccountID), CreateWorkerParams{
		ScriptName: "bar",
		Script:     workerScript,
		Bindings: map[string]WorkerBinding{
			"b1": WorkerDurableObjectBinding{
				ClassName:  "TheClass",
				ScriptName: "the_script",
			},
		},
	})

	assert.NoError(t, err)
}

func TestUploadWorker_WithInheritBinding(t *testing.T) {
	setup()
	defer teardown()

	formattedTime, _ := time.Parse(time.RFC3339Nano, "2018-06-09T15:17:01.989141Z")
	// Setup route handler for both single-script and multi-script
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)

		mpUpload, err := parseMultipartUpload(r)
		assert.NoError(t, err)

		expectedBindings := map[string]workerBindingMeta{
			"b1": {
				"name": "b1",
				"type": "inherit",
			},
			"b2": {
				"name":     "b2",
				"type":     "inherit",
				"old_name": "old_binding_name",
			},
		}
		assert.Equal(t, workerScript, mpUpload.Script)
		assert.Equal(t, expectedBindings, mpUpload.BindingMeta)

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, workersScriptResponse(t, withWorkerModifiedOn(formattedTime)))
	}
	mux.HandleFunc("/accounts/"+testAccountID+"/workers/scripts/bar", handler)

	want := WorkerScriptResponse{
		Response: successResponse,
		Module:   false,
		WorkerScript: WorkerScript{
			Script:     workerScript,
			UsageModel: "unbound",
			WorkerMetaData: WorkerMetaData{
				ID:               "e7a57d8746e74ae49c25994dadb421b1",
				ETAG:             "279cf40d86d70b82f6cd3ba90a646b3ad995912da446836d7371c21c6a43977a",
				Size:             191,
				ModifiedOn:       formattedTime,
				Logpush:          BoolPtr(false),
				LastDeployedFrom: StringPtr("dash"),
			},
		}}

	res, err := client.UploadWorker(context.Background(), AccountIdentifier(testAccountID), CreateWorkerParams{
		ScriptName: "bar",
		Script:     workerScript,
		Bindings: map[string]WorkerBinding{
			"b1": WorkerInheritBinding{},
			"b2": WorkerInheritBinding{
				OldName: "old_binding_name",
			},
		}})
	if assert.NoError(t, err) {
		assert.Equal(t, want, res)
	}
}

func TestUploadWorker_WithKVBinding(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)

		mpUpload, err := parseMultipartUpload(r)
		assert.NoError(t, err)

		expectedBindings := map[string]workerBindingMeta{
			"b1": {
				"name":         "b1",
				"type":         "kv_namespace",
				"namespace_id": "test-namespace",
			},
		}
		assert.Equal(t, workerScript, mpUpload.Script)
		assert.Equal(t, expectedBindings, mpUpload.BindingMeta)

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, workersScriptResponse(t))
	}
	mux.HandleFunc("/accounts/"+testAccountID+"/workers/scripts/bar", handler)

	_, err := client.UploadWorker(context.Background(), AccountIdentifier(testAccountID), CreateWorkerParams{
		ScriptName: "bar",
		Script:     workerScript,
		Bindings: map[string]WorkerBinding{
			"b1": WorkerKvNamespaceBinding{
				NamespaceID: "test-namespace",
			},
		}})
	assert.NoError(t, err)
}

func TestUploadWorker_WithWasmBinding(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)

		mpUpload, err := parseMultipartUpload(r)
		assert.NoError(t, err)

		partName := mpUpload.BindingMeta["b1"]["part"].(string)
		expectedBindings := map[string]workerBindingMeta{
			"b1": {
				"name": "b1",
				"type": "wasm_module",
				"part": partName,
			},
		}
		assert.Equal(t, workerScript, mpUpload.Script)
		assert.Equal(t, expectedBindings, mpUpload.BindingMeta)

		wasmContent, err := getFormValue(r, partName)
		assert.NoError(t, err)
		assert.Equal(t, []byte("fake-wasm"), wasmContent)

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, workersScriptResponse(t))
	}
	mux.HandleFunc("/accounts/"+testAccountID+"/workers/scripts/bar", handler)

	_, err := client.UploadWorker(context.Background(), AccountIdentifier(testAccountID), CreateWorkerParams{
		ScriptName: "bar",
		Script:     workerScript,
		Bindings: map[string]WorkerBinding{
			"b1": WorkerWebAssemblyBinding{
				Module: strings.NewReader("fake-wasm"),
			},
		},
	})

	assert.NoError(t, err)
}

func TestUploadWorker_WithPlainTextBinding(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)

		mpUpload, err := parseMultipartUpload(r)
		assert.NoError(t, err)

		expectedBindings := map[string]workerBindingMeta{
			"b1": {
				"name": "b1",
				"type": "plain_text",
				"text": "plain text value",
			},
		}
		assert.Equal(t, workerScript, mpUpload.Script)
		assert.Equal(t, expectedBindings, mpUpload.BindingMeta)

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, workersScriptResponse(t))
	}
	mux.HandleFunc("/accounts/"+testAccountID+"/workers/scripts/bar", handler)

	_, err := client.UploadWorker(context.Background(), AccountIdentifier(testAccountID), CreateWorkerParams{
		ScriptName: "bar",
		Script:     workerScript,
		Bindings: map[string]WorkerBinding{
			"b1": WorkerPlainTextBinding{
				Text: "plain text value",
			},
		},
	})

	assert.NoError(t, err)
}

func TestUploadWorker_ModuleWithPlainTextBinding(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/workers/scripts/bar", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)

		mpUpload, err := parseMultipartUpload(r)
		assert.NoError(t, err)

		expectedBindings := map[string]workerBindingMeta{
			"b1": {
				"name": "b1",
				"type": "plain_text",
				"text": "plain text value",
			},
		}
		assert.Equal(t, workerModuleScript, mpUpload.Script)
		assert.Equal(t, expectedBindings, mpUpload.BindingMeta)

		workerFileDetails, err := getFileDetails(r, "worker.mjs")
		if !assert.NoError(t, err) {
			assert.FailNow(t, "worker file not found in multipart form body")
		}
		contentDispositonHeader := workerFileDetails.Header.Get("content-disposition")
		expectedContentDisposition := fmt.Sprintf(`form-data; name="%s"; filename="%[1]s"`, "worker.mjs")
		assert.Equal(t, expectedContentDisposition, contentDispositonHeader, "Expected content-disposition request header to be %s, got %s", expectedContentDisposition, contentDispositonHeader)

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, workersScriptResponse(t, withWorkerScript(expectedWorkersModuleWorkerScript)))
	})

	_, err := client.UploadWorker(context.Background(), AccountIdentifier(testAccountID), CreateWorkerParams{
		ScriptName: "bar",
		Script:     workerModuleScript,
		Module:     true,
		Bindings: map[string]WorkerBinding{
			"b1": WorkerPlainTextBinding{
				Text: "plain text value",
			},
		},
	})

	assert.NoError(t, err)
}

func TestUploadWorker_WithSecretTextBinding(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)

		mpUpload, err := parseMultipartUpload(r)
		assert.NoError(t, err)

		expectedBindings := map[string]workerBindingMeta{
			"b1": {
				"name": "b1",
				"type": "secret_text",
				"text": "secret text value",
			},
		}
		assert.Equal(t, workerScript, mpUpload.Script)
		assert.Equal(t, expectedBindings, mpUpload.BindingMeta)

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, workersScriptResponse(t))
	}
	mux.HandleFunc("/accounts/"+testAccountID+"/workers/scripts/bar", handler)

	_, err := client.UploadWorker(context.Background(), AccountIdentifier(testAccountID), CreateWorkerParams{
		ScriptName: "bar",
		Script:     workerScript,
		Bindings: map[string]WorkerBinding{
			"b1": WorkerSecretTextBinding{
				Text: "secret text value",
			},
		},
	})
	assert.NoError(t, err)
}

func TestUploadWorker_WithServiceBinding(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)

		mpUpload, err := parseMultipartUpload(r)
		assert.NoError(t, err)

		expectedBindings := map[string]workerBindingMeta{
			"b1": {
				"name":    "b1",
				"type":    "service",
				"service": "the_service",
			},
			"b2": {
				"name":        "b2",
				"type":        "service",
				"service":     "the_service",
				"environment": "the_environment",
			},
		}
		assert.Equal(t, workerScript, mpUpload.Script)
		assert.Equal(t, expectedBindings, mpUpload.BindingMeta)

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, workersScriptResponse(t))
	}
	mux.HandleFunc("/accounts/"+testAccountID+"/workers/scripts/bar", handler)

	_, err := client.UploadWorker(context.Background(), AccountIdentifier(testAccountID), CreateWorkerParams{
		ScriptName: "bar",
		Script:     workerScript,
		Bindings: map[string]WorkerBinding{
			"b1": WorkerServiceBinding{
				Service: "the_service",
			},
			"b2": WorkerServiceBinding{
				Service:     "the_service",
				Environment: StringPtr("the_environment"),
			},
		},
	})
	assert.NoError(t, err)
}

func TestUploadWorker_WithLogpush(t *testing.T) {
	setup()
	defer teardown()

	var (
		formattedTime, _ = time.Parse(time.RFC3339Nano, "2018-06-09T15:17:01.989141Z")
		logpush          = BoolPtr(true)
	)
	mux.HandleFunc("/accounts/"+testAccountID+"/workers/scripts/foo", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		mpUpload, err := parseMultipartUpload(r)
		assert.NoError(t, err)

		expected := true
		assert.Equal(t, &expected, mpUpload.Logpush)

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, workersScriptResponse(t, withWorkerScript(expectedWorkersModuleWorkerScript), withWorkerLogpush(logpush), withWorkerModifiedOn(formattedTime)))
	})
	res, err := client.UploadWorker(context.Background(), AccountIdentifier(testAccountID), CreateWorkerParams{ScriptName: "foo", Script: workerScript, Logpush: logpush})
	want := WorkerScriptResponse{
		Response: successResponse,
		Module:   false,
		WorkerScript: WorkerScript{
			Script:     expectedWorkersModuleWorkerScript,
			UsageModel: "unbound",
			WorkerMetaData: WorkerMetaData{
				ID:               "e7a57d8746e74ae49c25994dadb421b1",
				ETAG:             "279cf40d86d70b82f6cd3ba90a646b3ad995912da446836d7371c21c6a43977a",
				Size:             191,
				ModifiedOn:       formattedTime,
				Logpush:          logpush,
				LastDeployedFrom: StringPtr("dash"),
			},
		}}
	if assert.NoError(t, err) {
		assert.Equal(t, want, res)
	}
}

func TestUploadWorker_WithCompatibilityFlags(t *testing.T) {
	setup()
	defer teardown()

	compatibilityDate := time.Now().Format("2006-01-02")
	compatibilityFlags := []string{"formdata_parser_supports_files"}

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)

		mpUpload, err := parseMultipartUpload(r)
		assert.NoError(t, err)

		assert.Equal(t, workerScript, mpUpload.Script)
		assert.Equal(t, compatibilityDate, mpUpload.CompatibilityDate)
		assert.Equal(t, compatibilityFlags, mpUpload.CompatibilityFlags)

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, workersScriptResponse(t))
	}
	mux.HandleFunc("/accounts/"+testAccountID+"/workers/scripts/bar", handler)

	_, err := client.UploadWorker(context.Background(), AccountIdentifier(testAccountID), CreateWorkerParams{
		ScriptName:         "bar",
		Script:             workerScript,
		CompatibilityDate:  compatibilityDate,
		CompatibilityFlags: compatibilityFlags,
	})
	assert.NoError(t, err)
}

func TestUploadWorker_WithQueueBinding(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)

		mpUpload, err := parseMultipartUpload(r)
		assert.NoError(t, err)

		expectedBindings := map[string]workerBindingMeta{
			"b1": {
				"name":       "b1",
				"type":       "queue",
				"queue_name": "test-queue",
			},
		}
		assert.Equal(t, workerScript, mpUpload.Script)
		assert.Equal(t, expectedBindings, mpUpload.BindingMeta)

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, workersScriptResponse(t))
	}
	mux.HandleFunc("/accounts/"+testAccountID+"/workers/scripts/bar", handler)

	_, err := client.UploadWorker(context.Background(), AccountIdentifier(testAccountID), CreateWorkerParams{
		ScriptName: "bar",
		Script:     workerScript,
		Bindings: map[string]WorkerBinding{
			"b1": WorkerQueueBinding{
				Binding: "b1",
				Queue:   "test-queue",
			},
		}})
	assert.NoError(t, err)
}

func TestUploadWorker_WithDispatchNamespaceBinding(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)

		mpUpload, err := parseMultipartUpload(r)
		assert.NoError(t, err)

		expectedBindings := map[string]workerBindingMeta{
			"b1": {
				"name":      "b1",
				"type":      "dispatch_namespace",
				"namespace": "n1",
				"outbound": map[string]interface{}{
					"worker": map[string]interface{}{
						"service":     "w1",
						"environment": "e1",
					},
					"params": []interface{}{
						map[string]interface{}{"name": "param1"},
					},
				},
			},
		}
		assert.Equal(t, workerScript, mpUpload.Script)
		assert.Equal(t, expectedBindings, mpUpload.BindingMeta)

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, workersScriptResponse(t))
	}
	mux.HandleFunc("/accounts/"+testAccountID+"/workers/scripts/bar", handler)

	environmentName := "e1"
	_, err := client.UploadWorker(context.Background(), AccountIdentifier(testAccountID), CreateWorkerParams{
		ScriptName: "bar",
		Script:     workerScript,
		Bindings: map[string]WorkerBinding{
			"b1": DispatchNamespaceBinding{
				Binding:   "b1",
				Namespace: "n1",
				Outbound: &NamespaceOutboundOptions{
					Worker: WorkerReference{
						Service:     "w1",
						Environment: &environmentName,
					},
					Params: []OutboundParamSchema{
						{
							Name: "param1",
						},
					},
				},
			},
		}})
	assert.NoError(t, err)
}

func TestUploadWorker_WithSmartPlacementEnabled(t *testing.T) {
	setup()
	defer teardown()

	placementMode := PlacementModeSmart
	response := workersScriptResponse(t, withWorkerScript(expectedWorkersModuleWorkerScript), withWorkerPlacementMode(StringPtr("smart")))

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)

		mpUpload, err := parseMultipartUpload(r)
		assert.NoError(t, err)

		assert.Equal(t, workerScript, mpUpload.Script)

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, response)
	}
	mux.HandleFunc("/accounts/"+testAccountID+"/workers/scripts/bar", handler)

	t.Run("Test enabling Smart Placement", func(t *testing.T) {
		worker, err := client.UploadWorker(context.Background(), AccountIdentifier(testAccountID), CreateWorkerParams{
			ScriptName: "bar",
			Script:     workerScript,
			Placement: &Placement{
				Mode: placementMode,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, placementMode, *worker.PlacementMode)
	})

	t.Run("Test disabling placement", func(t *testing.T) {
		placementMode = PlacementModeOff
		response = workersScriptResponse(t, withWorkerScript(expectedWorkersModuleWorkerScript))

		worker, err := client.UploadWorker(context.Background(), AccountIdentifier(testAccountID), CreateWorkerParams{
			ScriptName: "bar",
			Script:     workerScript,
			Placement: &Placement{
				Mode: placementMode,
			},
		})
		assert.NoError(t, err)
		assert.Nil(t, worker.PlacementMode)
	})
}

func TestUploadWorker_WithTailConsumers(t *testing.T) {
	setup()
	defer teardown()

	response := workersScriptResponse(t,
		withWorkerScript(expectedWorkersModuleWorkerScript))
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)

		mpUpload, err := parseMultipartUpload(r)
		assert.NoError(t, err)

		assert.Equal(t, workerScript, mpUpload.Script)

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, response)
	}
	mux.HandleFunc("/accounts/"+testAccountID+"/workers/scripts/bar", handler)

	t.Run("adds tail consumers", func(t *testing.T) {
		tailConsumers := []WorkersTailConsumer{
			{Service: "my-service-a"},
			{Service: "my-service-b", Environment: StringPtr("production")},
			{Service: "a-namespaced-service", Namespace: StringPtr("a-dispatch-namespace")},
		}
		response = workersScriptResponse(t,
			withWorkerScript(expectedWorkersModuleWorkerScript),
			withWorkerTailConsumers(tailConsumers...))

		worker, err := client.UploadWorker(context.Background(), AccountIdentifier(testAccountID), CreateWorkerParams{
			ScriptName:    "bar",
			Script:        workerScript,
			TailConsumers: &tailConsumers,
		})
		assert.NoError(t, err)
		require.NotNil(t, worker.TailConsumers)
		assert.Len(t, *worker.TailConsumers, 3)
	})
}

func TestUploadWorker_ToDispatchNamespace(t *testing.T) {
	setup()
	defer teardown()

	namespaceName := "n1"

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)

		mpUpload, err := parseMultipartUpload(r)
		require.NoError(t, err)

		assert.Equal(t, workerScript, mpUpload.Script)

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, workersScriptResponse(t))
	}
	mux.HandleFunc(
		fmt.Sprintf("/accounts/"+testAccountID+"/workers/dispatch/namespaces/%s/scripts/bar", namespaceName),
		handler,
	)

	_, err := client.UploadWorker(context.Background(), AccountIdentifier(testAccountID), CreateWorkerParams{
		ScriptName:            "bar",
		Script:                workerScript,
		DispatchNamespaceName: &namespaceName,
		Bindings: map[string]WorkerBinding{
			"b1": WorkerPlainTextBinding{
				Text: "hello",
			},
		},
	})
	assert.NoError(t, err)
}

func TestUploadWorker_ToDispatchNamespace_Tags(t *testing.T) {
	setup()
	defer teardown()

	namespaceName := "n1"
	tags := []string{"hello=there", "another-tag"}

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)

		mpUpload, err := parseMultipartUpload(r)
		require.NoError(t, err)

		assert.Equal(t, workerScript, mpUpload.Script)

		assert.EqualValues(t, tags, mpUpload.Tags)

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, workersScriptResponse(t))
	}
	mux.HandleFunc(
		fmt.Sprintf("/accounts/"+testAccountID+"/workers/dispatch/namespaces/%s/scripts/bar", namespaceName),
		handler,
	)

	_, err := client.UploadWorker(context.Background(), AccountIdentifier(testAccountID), CreateWorkerParams{
		ScriptName:            "bar",
		Script:                workerScript,
		DispatchNamespaceName: &namespaceName,
		Tags:                  tags,
	})
	assert.NoError(t, err)
}

func TestUploadWorker_UnsafeBinding(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)

		mpUpload, err := parseMultipartUpload(r)
		require.NoError(t, err)

		assert.Equal(t, workerScript, mpUpload.Script)

		require.Contains(t, mpUpload.BindingMeta, "b1")
		assert.Contains(t, mpUpload.BindingMeta["b1"], "name")
		assert.Equal(t, "b1", mpUpload.BindingMeta["b1"]["name"])
		assert.Contains(t, mpUpload.BindingMeta["b1"], "type")
		assert.Equal(t, "dynamic_dispatch", mpUpload.BindingMeta["b1"]["type"])

		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, workersScriptResponse(t))
	}
	mux.HandleFunc("/accounts/"+testAccountID+"/workers/scripts/bar", handler)

	_, err := client.UploadWorker(context.Background(), AccountIdentifier(testAccountID), CreateWorkerParams{
		ScriptName: "bar",
		Script:     workerScript,
		Bindings: map[string]WorkerBinding{
			"b1": UnsafeBinding{
				"type": "dynamic_dispatch",
			},
		},
	})
	assert.NoError(t, err)
}
