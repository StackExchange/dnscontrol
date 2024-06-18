package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testImagesVariantID = "hero"
)

func TestImageVariants_List(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, loadFixture("images_variants", "single_list"))
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/images/v1/variants", handler)

	want := ListImageVariantsResult{
		ImagesVariants: map[string]ImagesVariant{
			"hero": {
				ID:                     "hero",
				NeverRequireSignedURLs: BoolPtr(true),
				Options: ImagesVariantsOptions{
					Fit:      "scale-down",
					Height:   768,
					Width:    1366,
					Metadata: "none",
				},
			},
		},
	}

	got, err := client.ListImagesVariants(context.Background(), AccountIdentifier(testAccountID), ListImageVariantsParams{})
	if assert.NoError(t, err) {
		assert.Equal(t, want, got)
	}
}

func TestImageVariants_Delete(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method '%s', got %s", http.MethodDelete, r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {}
		}`)
	}

	url := fmt.Sprintf("/accounts/%s/images/v1/variants/%s", testAccountID, testImagesVariantID)
	mux.HandleFunc(url, handler)

	err := client.DeleteImagesVariant(context.Background(), AccountIdentifier(testAccountID), testImagesVariantID)
	assert.NoError(t, err)
}

func TestImagesVariants_Get(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method '%s', got %s", http.MethodGet, r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, loadFixture("images_variants", "single_full"))
	}

	url := fmt.Sprintf("/accounts/%s/images/v1/variants/%s", testAccountID, testImagesVariantID)
	mux.HandleFunc(url, handler)

	want := ImagesVariant{
		ID:                     "hero",
		NeverRequireSignedURLs: BoolPtr(true),
		Options: ImagesVariantsOptions{
			Fit:      "scale-down",
			Height:   768,
			Width:    1366,
			Metadata: "none",
		},
	}

	got, err := client.GetImagesVariant(context.Background(), AccountIdentifier(testAccountID), testImagesVariantID)
	if assert.NoError(t, err) {
		assert.Equal(t, want, got)
	}
}

func TestImagesVariants_Create(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method '%s', got %s", http.MethodPost, r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, loadFixture("images_variants", "single_full"))
	}

	url := fmt.Sprintf("/accounts/%s/images/v1/variants", testAccountID)
	mux.HandleFunc(url, handler)

	want := ImagesVariant{
		ID:                     "hero",
		NeverRequireSignedURLs: BoolPtr(true),
		Options: ImagesVariantsOptions{
			Fit:      "scale-down",
			Height:   768,
			Width:    1366,
			Metadata: "none",
		},
	}

	got, err := client.CreateImagesVariant(context.Background(), AccountIdentifier(testAccountID), CreateImagesVariantParams{
		ID:                     testImagesVariantID,
		NeverRequireSignedURLs: BoolPtr(true),
		Options: ImagesVariantsOptions{
			Fit:      "scale-down",
			Height:   768,
			Width:    1366,
			Metadata: "none",
		},
	})
	if assert.NoError(t, err) {
		assert.Equal(t, want, got)
	}
}

func TestImagesVariants_Update(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method '%s', got %s", http.MethodPatch, r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, loadFixture("images_variants", "single_full"))
	}

	url := fmt.Sprintf("/accounts/%s/images/v1/variants/%s", testAccountID, testImagesVariantID)
	mux.HandleFunc(url, handler)

	want := ImagesVariant{
		ID:                     "hero",
		NeverRequireSignedURLs: BoolPtr(true),
		Options: ImagesVariantsOptions{
			Fit:      "scale-down",
			Height:   768,
			Width:    1366,
			Metadata: "none",
		},
	}

	got, err := client.UpdateImagesVariant(context.Background(), AccountIdentifier(testAccountID), UpdateImagesVariantParams{
		ID:                     "hero",
		NeverRequireSignedURLs: BoolPtr(true),
		Options: ImagesVariantsOptions{
			Fit:      "scale-down",
			Height:   768,
			Width:    1366,
			Metadata: "none",
		},
	})

	if assert.NoError(t, err) {
		assert.Equal(t, want, got)
	}
}

func TestImageVariants_MissingAccountId(t *testing.T) {
	_, err := client.ListImagesVariants(context.Background(), AccountIdentifier(""), ListImageVariantsParams{})
	assert.Equal(t, ErrMissingAccountID, err)

	_, err = client.GetImagesVariant(context.Background(), AccountIdentifier(""), testImagesVariantID)
	assert.Equal(t, ErrMissingAccountID, err)

	_, err = client.CreateImagesVariant(context.Background(), AccountIdentifier(""), CreateImagesVariantParams{})
	assert.Equal(t, ErrMissingAccountID, err)

	err = client.DeleteImagesVariant(context.Background(), AccountIdentifier(""), testImagesVariantID)
	assert.Equal(t, ErrMissingAccountID, err)

	_, err = client.UpdateImagesVariant(context.Background(), AccountIdentifier(""), UpdateImagesVariantParams{})
	assert.Equal(t, ErrMissingAccountID, err)
}
