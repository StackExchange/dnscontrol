package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const testPagesProjectName = "page-project"

var testCreatedOn, _ = time.Parse(time.RFC3339, "2017-01-01T00:00:00Z")

var testPagesDomain = PagesDomain{
	ID:     "8232210c-6818-4e34-8d95-cc386874b8d2",
	Name:   "example.com",
	Status: "pending",
	VerificationData: VerificationData{
		Status: "active",
	},
	ValidationData: ValidationData{
		Status: "active",
		Method: "http",
	},
	ZoneTag:   "023e105f4ecef8ad9ca31a8372d0c353",
	CreatedOn: &testCreatedOn,
}

func TestPages_GetDomains(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/pages/projects/"+testPagesProjectName+"/domains", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `
{
  "success": true,
  "errors": [],
  "messages": [],
  "result": [
    {
      "id": "8232210c-6818-4e34-8d95-cc386874b8d2",
      "name": "example.com",
      "status": "pending",
      "verification_data": {
        "status": "active"
      },
      "validation_data": {
        "status": "active",
        "method": "http"
      },
      "zone_tag": "023e105f4ecef8ad9ca31a8372d0c353",
      "created_on": "2017-01-01T00:00:00Z"
    }
  ]
}`)
	})

	// Make sure missing account ID is thrown
	_, err := client.GetPagesDomains(context.Background(), PagesDomainsParameters{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	// Make sure missing project name is thrown
	_, err = client.GetPagesDomains(context.Background(), PagesDomainsParameters{AccountID: testAccountID})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingProjectName, err)
	}

	out, err := client.GetPagesDomains(context.Background(), PagesDomainsParameters{AccountID: testAccountID, ProjectName: testPagesProjectName})
	if assert.NoError(t, err) {
		assert.Equal(t, 1, len(out), "Domains length not correct")
		assert.Equal(t, out[0], testPagesDomain, "structs not equal")
	}
}

func TestPages_GetDomain(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/pages/projects/"+testPagesProjectName+"/domains/example.com", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `
{
  "success": true,
  "errors": [],
  "messages": [],
  "result": {
      "id": "8232210c-6818-4e34-8d95-cc386874b8d2",
      "name": "example.com",
      "status": "pending",
      "verification_data": {
        "status": "active"
      },
      "validation_data": {
        "status": "active",
        "method": "http"
      },
      "zone_tag": "023e105f4ecef8ad9ca31a8372d0c353",
      "created_on": "2017-01-01T00:00:00Z"
    }
}`)
	})

	// Make sure missing account ID is thrown
	_, err := client.GetPagesDomain(context.Background(), PagesDomainParameters{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	// Make sure missing project name is thrown
	_, err = client.GetPagesDomain(context.Background(), PagesDomainParameters{AccountID: testAccountID})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingProjectName, err)
	}

	// Make sure missing domain is thrown
	_, err = client.GetPagesDomain(context.Background(), PagesDomainParameters{AccountID: testAccountID, ProjectName: testPagesProjectName})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingDomain, err)
	}

	out, err := client.GetPagesDomain(context.Background(), PagesDomainParameters{AccountID: testAccountID, ProjectName: testPagesProjectName, DomainName: "example.com"})
	if assert.NoError(t, err) {
		assert.Equal(t, out, testPagesDomain, "structs not equal")
	}
}

func TestPages_PatchDomain(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/pages/projects/"+testPagesProjectName+"/domains/example.com", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `
{
  "success": true,
  "errors": [],
  "messages": [],
  "result": {
      "id": "8232210c-6818-4e34-8d95-cc386874b8d2",
      "name": "example.com",
      "status": "pending",
      "verification_data": {
        "status": "active"
      },
      "validation_data": {
        "status": "active",
        "method": "http"
      },
      "zone_tag": "023e105f4ecef8ad9ca31a8372d0c353",
      "created_on": "2017-01-01T00:00:00Z"
    }
}`)
	})

	// Make sure missing account ID is thrown
	_, err := client.PagesPatchDomain(context.Background(), PagesDomainParameters{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	// Make sure missing project name is thrown
	_, err = client.PagesPatchDomain(context.Background(), PagesDomainParameters{AccountID: testAccountID})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingProjectName, err)
	}

	// Make sure missing domain is thrown
	_, err = client.PagesPatchDomain(context.Background(), PagesDomainParameters{AccountID: testAccountID, ProjectName: testPagesProjectName})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingDomain, err)
	}

	out, err := client.PagesPatchDomain(context.Background(), PagesDomainParameters{AccountID: testAccountID, ProjectName: testPagesProjectName, DomainName: "example.com"})
	if assert.NoError(t, err) {
		assert.Equal(t, out, testPagesDomain, "structs not equal")
	}
}

func TestPages_AddDomain(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/pages/projects/"+testPagesProjectName+"/domains", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `
{
  "success": true,
  "errors": [],
  "messages": [],
  "result": {
      "id": "8232210c-6818-4e34-8d95-cc386874b8d2",
      "name": "example.com",
      "status": "pending",
      "verification_data": {
        "status": "active"
      },
      "validation_data": {
        "status": "active",
        "method": "http"
      },
      "zone_tag": "023e105f4ecef8ad9ca31a8372d0c353",
      "created_on": "2017-01-01T00:00:00Z"
    }
}`)
	})

	// Make sure missing account ID is thrown
	_, err := client.PagesAddDomain(context.Background(), PagesDomainParameters{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	// Make sure missing project name is thrown
	_, err = client.PagesAddDomain(context.Background(), PagesDomainParameters{AccountID: testAccountID})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingProjectName, err)
	}

	// Make sure missing domain is thrown
	_, err = client.PagesAddDomain(context.Background(), PagesDomainParameters{AccountID: testAccountID, ProjectName: testPagesProjectName})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingDomain, err)
	}

	out, err := client.PagesAddDomain(context.Background(), PagesDomainParameters{AccountID: testAccountID, ProjectName: testPagesProjectName, DomainName: "example.com"})
	if assert.NoError(t, err) {
		assert.Equal(t, out, testPagesDomain, "structs not equal")
	}
}

func TestPages_DeleteDomain(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/pages/projects/"+testPagesProjectName+"/domains/example.com", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `
{
  "success": true,
  "errors": [],
  "messages": [],
  "result": null
}`)
	})

	// Make sure missing account ID is thrown
	err := client.PagesDeleteDomain(context.Background(), PagesDomainParameters{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	// Make sure missing project name is thrown
	err = client.PagesDeleteDomain(context.Background(), PagesDomainParameters{AccountID: testAccountID})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingProjectName, err)
	}

	// Make sure missing domain is thrown
	err = client.PagesDeleteDomain(context.Background(), PagesDomainParameters{AccountID: testAccountID, ProjectName: testPagesProjectName})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingDomain, err)
	}

	err = client.PagesDeleteDomain(context.Background(), PagesDomainParameters{AccountID: testAccountID, ProjectName: testPagesProjectName, DomainName: "example.com"})
	assert.NoError(t, err)
}
