package cloudflare

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/goccy/go-json"

	"github.com/stretchr/testify/assert"
)

func TestListWorkerBindings(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/workers/scripts/my-script/bindings", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, listBindingsResponseData)
	})

	mux.HandleFunc("/accounts/"+testAccountID+"/workers/scripts/my-script/bindings/MY_WASM/content", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/wasm")
		_, _ = w.Write([]byte("mock multi-script wasm"))
	})

	res, err := client.ListWorkerBindings(context.Background(), AccountIdentifier(testAccountID), ListWorkerBindingsParams{
		ScriptName: "my-script",
	})
	assert.NoError(t, err)

	assert.Equal(t, successResponse, res.Response)
	assert.Equal(t, 9, len(res.BindingList))

	assert.Equal(t, res.BindingList[0], WorkerBindingListItem{
		Name: "MY_KV",
		Binding: WorkerKvNamespaceBinding{
			NamespaceID: "89f5f8fd93f94cb98473f6f421aa3b65",
		},
	})
	assert.Equal(t, WorkerKvNamespaceBindingType, res.BindingList[0].Binding.Type())

	assert.Equal(t, "MY_WASM", res.BindingList[1].Name)
	wasmBinding := res.BindingList[1].Binding.(WorkerWebAssemblyBinding)
	wasmModuleContent, err := io.ReadAll(wasmBinding.Module)
	assert.NoError(t, err)
	assert.Equal(t, []byte("mock multi-script wasm"), wasmModuleContent)
	assert.Equal(t, WorkerWebAssemblyBindingType, res.BindingList[1].Binding.Type())

	assert.Equal(t, res.BindingList[2], WorkerBindingListItem{
		Name: "MY_PLAIN_TEXT",
		Binding: WorkerPlainTextBinding{
			Text: "text",
		},
	})
	assert.Equal(t, WorkerPlainTextBindingType, res.BindingList[2].Binding.Type())

	assert.Equal(t, res.BindingList[3], WorkerBindingListItem{
		Name:    "MY_SECRET_TEXT",
		Binding: WorkerSecretTextBinding{},
	})
	assert.Equal(t, WorkerSecretTextBindingType, res.BindingList[3].Binding.Type())

	environment := "MY_ENVIRONMENT"
	assert.Equal(t, res.BindingList[4], WorkerBindingListItem{
		Name: "MY_SERVICE_BINDING",
		Binding: WorkerServiceBinding{
			Service:     "MY_SERVICE",
			Environment: &environment,
		},
	})
	assert.Equal(t, WorkerServiceBindingType, res.BindingList[4].Binding.Type())

	assert.Equal(t, res.BindingList[5], WorkerBindingListItem{
		Name:    "MY_NEW_BINDING",
		Binding: WorkerInheritBinding{},
	})
	assert.Equal(t, WorkerInheritBindingType, res.BindingList[5].Binding.Type())

	assert.Equal(t, res.BindingList[6], WorkerBindingListItem{
		Name: "MY_BUCKET",
		Binding: WorkerR2BucketBinding{
			BucketName: "bucket",
		},
	})
	assert.Equal(t, WorkerR2BucketBindingType, res.BindingList[6].Binding.Type())

	assert.Equal(t, res.BindingList[7], WorkerBindingListItem{
		Name: "MY_DATASET",
		Binding: WorkerAnalyticsEngineBinding{
			Dataset: "my_dataset",
		},
	})

	assert.Equal(t, WorkerAnalyticsEngineBindingType, res.BindingList[7].Binding.Type())

	assert.Equal(t, res.BindingList[8], WorkerBindingListItem{
		Name: "MY_DATABASE",
		Binding: WorkerD1DatabaseBinding{
			DatabaseID: "cef5331f-e5c7-4c8a-a415-7908ae45f92a",
		},
	})
	assert.Equal(t, WorkerD1DataseBindingType, res.BindingList[8].Binding.Type())
}

func TestListWorkerBindings_Wfp(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/workers/dispatch/namespaces/my-namespace/scripts/my-script/bindings", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, listBindingsResponseData)
	})

	res, err := client.ListWorkerBindings(context.Background(), AccountIdentifier(testAccountID), ListWorkerBindingsParams{
		ScriptName:        "my-script",
		DispatchNamespace: &[]string{"my-namespace"}[0],
	})
	assert.NoError(t, err)

	assert.Equal(t, successResponse, res.Response)
	assert.Equal(t, 9, len(res.BindingList))

	assert.Equal(t, res.BindingList[0], WorkerBindingListItem{
		Name: "MY_KV",
		Binding: WorkerKvNamespaceBinding{
			NamespaceID: "89f5f8fd93f94cb98473f6f421aa3b65",
		},
	})
	assert.Equal(t, WorkerKvNamespaceBindingType, res.BindingList[0].Binding.Type())

	// WASM binding - No binding content endpoint exists for WfP

	assert.Equal(t, res.BindingList[2], WorkerBindingListItem{
		Name: "MY_PLAIN_TEXT",
		Binding: WorkerPlainTextBinding{
			Text: "text",
		},
	})
	assert.Equal(t, WorkerPlainTextBindingType, res.BindingList[2].Binding.Type())

	assert.Equal(t, res.BindingList[3], WorkerBindingListItem{
		Name:    "MY_SECRET_TEXT",
		Binding: WorkerSecretTextBinding{},
	})
	assert.Equal(t, WorkerSecretTextBindingType, res.BindingList[3].Binding.Type())

	environment := "MY_ENVIRONMENT"
	assert.Equal(t, res.BindingList[4], WorkerBindingListItem{
		Name: "MY_SERVICE_BINDING",
		Binding: WorkerServiceBinding{
			Service:     "MY_SERVICE",
			Environment: &environment,
		},
	})
	assert.Equal(t, WorkerServiceBindingType, res.BindingList[4].Binding.Type())

	assert.Equal(t, res.BindingList[5], WorkerBindingListItem{
		Name:    "MY_NEW_BINDING",
		Binding: WorkerInheritBinding{},
	})
	assert.Equal(t, WorkerInheritBindingType, res.BindingList[5].Binding.Type())

	assert.Equal(t, res.BindingList[6], WorkerBindingListItem{
		Name: "MY_BUCKET",
		Binding: WorkerR2BucketBinding{
			BucketName: "bucket",
		},
	})
	assert.Equal(t, WorkerR2BucketBindingType, res.BindingList[6].Binding.Type())

	assert.Equal(t, res.BindingList[7], WorkerBindingListItem{
		Name: "MY_DATASET",
		Binding: WorkerAnalyticsEngineBinding{
			Dataset: "my_dataset",
		},
	})

	assert.Equal(t, WorkerAnalyticsEngineBindingType, res.BindingList[7].Binding.Type())

	assert.Equal(t, res.BindingList[8], WorkerBindingListItem{
		Name: "MY_DATABASE",
		Binding: WorkerD1DatabaseBinding{
			DatabaseID: "cef5331f-e5c7-4c8a-a415-7908ae45f92a",
		},
	})
	assert.Equal(t, WorkerD1DataseBindingType, res.BindingList[8].Binding.Type())
}

func ExampleUnsafeBinding() {
	pretty := func(meta workerBindingMeta) string {
		buf := bytes.NewBufferString("")
		encoder := json.NewEncoder(buf)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(meta); err != nil {
			fmt.Println("error:", err)
		}
		return buf.String()
	}

	binding_a := WorkerServiceBinding{
		Service: "foo",
	}
	meta_a, _, _ := binding_a.serialize("my_binding")
	meta_a_json := pretty(meta_a)
	fmt.Println(meta_a_json)

	binding_b := UnsafeBinding{
		"type":    "service",
		"service": "foo",
	}
	meta_b, _, _ := binding_b.serialize("my_binding")
	meta_b_json := pretty(meta_b)
	fmt.Println(meta_b_json)

	fmt.Println(meta_a_json == meta_b_json)
	// Output:
	// {
	//   "name": "my_binding",
	//   "service": "foo",
	//   "type": "service"
	// }
	//
	// {
	//   "name": "my_binding",
	//   "service": "foo",
	//   "type": "service"
	// }
	//
	// true
}
