package cloudflare

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/goccy/go-json"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func timeMustParse(layout, value string) time.Time {
	t, err := time.Parse(layout, value)
	if err != nil {
		panic(err)
	}
	return t
}

var expectedImageStruct = Image{
	ID:       "ZxR0pLaXRldlBtaFhhO2FiZGVnaA",
	Filename: "avatar.png",
	Meta: map[string]interface{}{
		"meta": "metaID",
	},
	RequireSignedURLs: true,
	Variants: []string{
		"https://imagedelivery.net/MTt4OTd0b0w5aj/ZxR0pLaXRldlBtaFhhO2FiZGVnaA/hero",
		"https://imagedelivery.net/MTt4OTd0b0w5aj/ZxR0pLaXRldlBtaFhhO2FiZGVnaA/original",
		"https://imagedelivery.net/MTt4OTd0b0w5aj/ZxR0pLaXRldlBtaFhhO2FiZGVnaA/thumbnail",
	},
	Uploaded: timeMustParse(time.RFC3339, "2014-01-02T02:20:00Z"),
}

func TestUploadImage(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)

		u, err := parseImageMultipartUpload(r)
		if !assert.NoError(t, err) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		assert.Equal(t, u.RequireSignedURLs, true)
		assert.Equal(t, u.Metadata, map[string]interface{}{"meta": "metaID"})
		assert.Equal(t, u.File, []byte("this is definitely an image"))

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "ZxR0pLaXRldlBtaFhhO2FiZGVnaA",
				"filename": "avatar.png",
				"meta": {
					"meta": "metaID"
				},
				"requireSignedURLs": true,
				"variants": [
					"https://imagedelivery.net/MTt4OTd0b0w5aj/ZxR0pLaXRldlBtaFhhO2FiZGVnaA/hero",
					"https://imagedelivery.net/MTt4OTd0b0w5aj/ZxR0pLaXRldlBtaFhhO2FiZGVnaA/original",
					"https://imagedelivery.net/MTt4OTd0b0w5aj/ZxR0pLaXRldlBtaFhhO2FiZGVnaA/thumbnail"
				],
				"uploaded": "2014-01-02T02:20:00Z"
			}
		}
		`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/images/v1", handler)
	want := expectedImageStruct

	actual, err := client.UploadImage(context.Background(), AccountIdentifier(testAccountID), UploadImageParams{
		File: fakeFile{
			Buffer: bytes.NewBufferString("this is definitely an image"),
		},
		Name:              "avatar.png",
		RequireSignedURLs: true,
		Metadata: map[string]interface{}{
			"meta": "metaID",
		},
	})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestUploadImageByUrl(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)

		u, err := parseImageMultipartUpload(r)
		if !assert.NoError(t, err) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		assert.Equal(t, u.RequireSignedURLs, true)
		assert.Equal(t, u.Metadata, map[string]interface{}{"meta": "metaID"})
		assert.Equal(t, u.Url, "https://www.images-elsewhere.com/avatar.png")

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "ZxR0pLaXRldlBtaFhhO2FiZGVnaA",
				"filename": "avatar.png",
				"meta": {
					"meta": "metaID"
				},
				"requireSignedURLs": true,
				"variants": [
					"https://imagedelivery.net/MTt4OTd0b0w5aj/ZxR0pLaXRldlBtaFhhO2FiZGVnaA/hero",
					"https://imagedelivery.net/MTt4OTd0b0w5aj/ZxR0pLaXRldlBtaFhhO2FiZGVnaA/original",
					"https://imagedelivery.net/MTt4OTd0b0w5aj/ZxR0pLaXRldlBtaFhhO2FiZGVnaA/thumbnail"
				],
				"uploaded": "2014-01-02T02:20:00Z"
			}
		}
		`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/images/v1", handler)
	want := expectedImageStruct

	actual, err := client.UploadImage(context.Background(), AccountIdentifier(testAccountID), UploadImageParams{
		URL:               "https://www.images-elsewhere.com/avatar.png",
		RequireSignedURLs: true,
		Metadata: map[string]interface{}{
			"meta": "metaID",
		},
	})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestUpdateImage(t *testing.T) {
	setup()
	defer teardown()

	input := UpdateImageParams{
		RequireSignedURLs: true,
		Metadata: map[string]interface{}{
			"meta": "metaID",
		},
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method, "Expected method 'PATCH', got %s", r.Method)

		var v UpdateImageParams
		err := json.NewDecoder(r.Body).Decode(&v)
		require.NoError(t, err)
		assert.Equal(t, input, v)

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "ZxR0pLaXRldlBtaFhhO2FiZGVnaA",
				"filename": "avatar.png",
				"meta": {
					"meta": "metaID"
				},
				"requireSignedURLs": true,
				"variants": [
					"https://imagedelivery.net/MTt4OTd0b0w5aj/ZxR0pLaXRldlBtaFhhO2FiZGVnaA/hero",
					"https://imagedelivery.net/MTt4OTd0b0w5aj/ZxR0pLaXRldlBtaFhhO2FiZGVnaA/original",
					"https://imagedelivery.net/MTt4OTd0b0w5aj/ZxR0pLaXRldlBtaFhhO2FiZGVnaA/thumbnail"
				],
				"uploaded": "2014-01-02T02:20:00Z"
			}
		}
		`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/images/v1/ZxR0pLaXRldlBtaFhhO2FiZGVnaA", handler)
	want := expectedImageStruct

	actual, err := client.UpdateImage(context.Background(), AccountIdentifier(testAccountID), UpdateImageParams{
		ID:                "ZxR0pLaXRldlBtaFhhO2FiZGVnaA",
		RequireSignedURLs: true,
		Metadata: map[string]interface{}{
			"meta": "metaID",
		},
	})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateImageDirectUploadURL(t *testing.T) {
	setup()
	defer teardown()

	expiry := time.Now().UTC().Add(30 * time.Minute)
	input := CreateImageDirectUploadURLParams{
		Expiry: &expiry,
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)

		var v CreateImageDirectUploadURLParams
		err := json.NewDecoder(r.Body).Decode(&v)
		require.NoError(t, err)
		assert.Equal(t, input, v)

		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "ZxR0pLaXRldlBtaFhhO2FiZGVnaA",
				"uploadURL": "https://upload.imagedelivery.net/fgr33htrthytjtyereifjewoi338272s7w1383"
			}
		}
		`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/images/v1/direct_upload", handler)
	want := ImageDirectUploadURL{
		ID:        "ZxR0pLaXRldlBtaFhhO2FiZGVnaA",
		UploadURL: "https://upload.imagedelivery.net/fgr33htrthytjtyereifjewoi338272s7w1383",
	}

	actual, err := client.CreateImageDirectUploadURL(context.Background(), AccountIdentifier(testAccountID), input)

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestCreateImageConflictingTypes(t *testing.T) {
	setup()
	defer teardown()

	_, err := client.UploadImage(context.Background(), AccountIdentifier(testAccountID), UploadImageParams{
		URL: "https://example.com/foo.jpg",
		File: fakeFile{
			Buffer: bytes.NewBufferString("this is definitely an image"),
		},
	})

	assert.Error(t, err)
}

func TestCreateImageDirectUploadURLV2(t *testing.T) {
	setup()
	defer teardown()

	exp := time.Now().UTC().Add(30 * time.Minute)
	metadata := map[string]interface{}{
		"metaKey1": "metaValue1",
		"metaKey2": "metaValue2",
	}
	requireSignedURLs := true
	input := CreateImageDirectUploadURLParams{
		Version:           ImagesAPIVersionV2,
		Expiry:            &exp,
		Metadata:          metadata,
		RequireSignedURLs: &requireSignedURLs,
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		require.Equal(t,
			fmt.Sprintf("multipart/form-data; boundary=%s", imagesMultipartBoundary),
			r.Header.Get("Content-Type"),
		)
		require.NoError(t, r.ParseMultipartForm(32<<20))
		require.Equal(t, exp.Format(time.RFC3339), r.Form.Get("expiry"))
		require.Equal(t, "true", r.Form.Get("requireSignedURLs"))
		marshalledMetadata, err := json.Marshal(metadata)
		require.NoError(t, err)
		require.Equal(t, string(marshalledMetadata), r.Form.Get("metadata"))
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "ZxR0pLaXRldlBtaFhhO2FiZGVnaA",
				"uploadURL": "https://upload.imagedelivery.net/fgr33htrthytjtyereifjewoi338272s7w1383"
			}
		}
		`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/images/v2/direct_upload", handler)
	want := ImageDirectUploadURL{
		ID:        "ZxR0pLaXRldlBtaFhhO2FiZGVnaA",
		UploadURL: "https://upload.imagedelivery.net/fgr33htrthytjtyereifjewoi338272s7w1383",
	}

	actual, err := client.CreateImageDirectUploadURL(context.Background(), AccountIdentifier(testAccountID), input)
	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestListImages(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"images": [
					{
						"id": "ZxR0pLaXRldlBtaFhhO2FiZGVnaA",
						"filename": "avatar.png",
						"meta": {
							"meta": "metaID"
						},
						"requireSignedURLs": true,
						"variants": [
							"https://imagedelivery.net/MTt4OTd0b0w5aj/ZxR0pLaXRldlBtaFhhO2FiZGVnaA/hero",
							"https://imagedelivery.net/MTt4OTd0b0w5aj/ZxR0pLaXRldlBtaFhhO2FiZGVnaA/original",
							"https://imagedelivery.net/MTt4OTd0b0w5aj/ZxR0pLaXRldlBtaFhhO2FiZGVnaA/thumbnail"
						],
						"uploaded": "2014-01-02T02:20:00Z"
					}
				]
			}
		}
		`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/images/v1", handler)
	want := []Image{expectedImageStruct}

	actual, err := client.ListImages(context.Background(), AccountIdentifier(testAccountID), ListImagesParams{})

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestImageDetails(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {
				"id": "ZxR0pLaXRldlBtaFhhO2FiZGVnaA",
				"filename": "avatar.png",
				"meta": {
					"meta": "metaID"
				},
				"requireSignedURLs": true,
				"variants": [
					"https://imagedelivery.net/MTt4OTd0b0w5aj/ZxR0pLaXRldlBtaFhhO2FiZGVnaA/hero",
					"https://imagedelivery.net/MTt4OTd0b0w5aj/ZxR0pLaXRldlBtaFhhO2FiZGVnaA/original",
					"https://imagedelivery.net/MTt4OTd0b0w5aj/ZxR0pLaXRldlBtaFhhO2FiZGVnaA/thumbnail"
				],
				"uploaded": "2014-01-02T02:20:00Z"
			}
		}
		`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/images/v1/ZxR0pLaXRldlBtaFhhO2FiZGVnaA", handler)
	want := expectedImageStruct

	actual, err := client.GetImage(context.Background(), AccountIdentifier(testAccountID), "ZxR0pLaXRldlBtaFhhO2FiZGVnaA")

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestBaseImage(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "image/png")
		_, _ = w.Write([]byte{})
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/images/v1/ZxR0pLaXRldlBtaFhhO2FiZGVnaA/blob", handler)
	want := []byte{}

	actual, err := client.GetBaseImage(context.Background(), AccountIdentifier(testAccountID), "ZxR0pLaXRldlBtaFhhO2FiZGVnaA")

	if assert.NoError(t, err) {
		assert.Equal(t, want, actual)
	}
}

func TestDeleteImage(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {}
		}
		`)
	}

	mux.HandleFunc("/accounts/"+testAccountID+"/images/v1/ZxR0pLaXRldlBtaFhhO2FiZGVnaA", handler)

	err := client.DeleteImage(context.Background(), AccountIdentifier(testAccountID), "ZxR0pLaXRldlBtaFhhO2FiZGVnaA")
	require.NoError(t, err)
}

type fakeFile struct {
	*bytes.Buffer
}

func (f fakeFile) Close() error {
	return nil
}

type imageMultipartUpload struct {
	// this is for testing, never read an entire file into memory,
	// especially when being done on a per-http request basis.
	File              []byte
	Url               string
	RequireSignedURLs bool
	Metadata          map[string]interface{}
}

func parseImageMultipartUpload(r *http.Request) (imageMultipartUpload, error) {
	var u imageMultipartUpload
	mdBytes, err := getImageFormValue(r, "metadata")
	if err != nil {
		if !strings.HasPrefix(err.Error(), "no value found for key") {
			return u, err
		}
	}
	if mdBytes != nil {
		err = json.Unmarshal(mdBytes, &u.Metadata)
		if err != nil {
			return u, err
		}
	}

	rsuBytes, err := getImageFormValue(r, "requireSignedURLs")
	if err != nil {
		if !strings.HasPrefix(err.Error(), "no value found for key") {
			return u, err
		}
	}
	if rsuBytes != nil {
		if bytes.Equal(rsuBytes, []byte("true")) {
			u.RequireSignedURLs = true
		}
	}

	if _, ok := r.MultipartForm.Value["url"]; ok {
		urlBytes, err := getImageFormValue(r, "url")
		if err != nil {
			if !strings.HasPrefix(err.Error(), "no value found for key") {
				return u, err
			}
		}
		if urlBytes != nil {
			u.Url = string(urlBytes)
		}
	} else {
		f, _, err := r.FormFile("file")
		if err != nil {
			return u, err
		}
		defer f.Close()

		u.File, err = io.ReadAll(f)
		if err != nil {
			return u, err
		}
	}

	return u, nil
}

// See getFormValue for more information, the only difference between
// getFormValue and this one is the max memory.
func getImageFormValue(r *http.Request, key string) ([]byte, error) {
	err := r.ParseMultipartForm(10 * 1024 * 1024)
	if err != nil {
		return nil, err
	}

	if values, ok := r.MultipartForm.Value[key]; ok {
		return []byte(values[0]), nil
	}

	if fileHeaders, ok := r.MultipartForm.File[key]; ok {
		file, err := fileHeaders[0].Open()
		if err != nil {
			return nil, err
		}
		return io.ReadAll(file)
	}

	return nil, fmt.Errorf("no value found for key %v", key)
}
