package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetWorkersSecret(t *testing.T) {
	setup()
	defer teardown()

	response := `{
		"result": {
			"name" : "my-secret",
			"type": "secret_text"
		},
		"success": true,
		"errors": [],
		"messages": []
	}`

	mux.HandleFunc("/accounts/"+testAccountID+"/workers/scripts/test-script/secrets", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method, "Expected method 'PUT', got %s", r.Method)
		w.Header().Set("content-type", "application/javascript")
		fmt.Fprint(w, response)
	})
	req := &WorkersPutSecretRequest{
		Name: "my-secret",
		Text: "super-secret",
	}
	res, err := client.SetWorkersSecret(context.Background(), AccountIdentifier(testAccountID), SetWorkersSecretParams{ScriptName: "test-script", Secret: req})
	want := WorkersPutSecretResponse{
		successResponse,
		WorkersSecret{
			Name: "test",
			Type: "secret_text",
		},
	}

	if assert.NoError(t, err) {
		assert.Equal(t, want.Response, res.Response)
	}
}

func TestDeleteWorkersSecret(t *testing.T) {
	setup()
	defer teardown()

	response := `{
		"result": {
			"name" : "test",
			"type": "secret_text"
		},
		"success": true,
		"errors": [],
		"messages": []
	}`

	mux.HandleFunc("/accounts/"+testAccountID+"/workers/scripts/test-script/secrets/my-secret", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/javascript")
		fmt.Fprint(w, response)
	})

	res, err := client.DeleteWorkersSecret(context.Background(), AccountIdentifier(testAccountID), DeleteWorkersSecretParams{ScriptName: "test-script", SecretName: "my-secret"})
	want := successResponse

	if assert.NoError(t, err) {
		assert.Equal(t, want, res)
	}
}

func TestListWorkersSecret(t *testing.T) {
	setup()
	defer teardown()

	response := `{
		"result": [{
			"name" : "my-secret",
			"type": "secret_text"
		}],
		"success": true,
		"errors": [],
		"messages": []
	}`

	mux.HandleFunc("/accounts/"+testAccountID+"/workers/scripts/test-script/secrets", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/javascript")
		fmt.Fprint(w, response)
	})

	res, err := client.ListWorkersSecrets(context.Background(), AccountIdentifier(testAccountID), ListWorkersSecretsParams{ScriptName: "test-script"})
	want := WorkersListSecretsResponse{
		successResponse,
		[]WorkersSecret{
			{
				Name: "my-secret",
				Type: "secret_text",
			},
		},
	}

	if assert.NoError(t, err) {
		assert.Equal(t, want.Response, res.Response)
	}
}
