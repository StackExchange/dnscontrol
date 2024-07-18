package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	singleStreamResponse = `
{
  "success": true,
  "errors": [],
  "messages": [],
  "result": {
    "allowedOrigins": [
      "example.com"
    ],
    "created": "2014-01-02T02:20:00Z",
    "duration": 300.5,
    "input": {
      "height": 1080,
      "width": 1920
    },
    "maxDurationSeconds": 300,
    "meta": {
	  "name": "My First Stream Video"
	},
    "modified": "2014-01-02T02:20:00Z",
    "uploadExpiry": "2014-01-02T02:20:00Z",
    "playback": {
      "hls": "https://videodelivery.net/ea95132c15732412d22c1476fa83f27a/manifest/video.m3u8",
      "dash": "https://videodelivery.net/ea95132c15732412d22c1476fa83f27a/manifest/video.mpd"
    },
    "preview": "https://watch.cloudflarestream.com/ea95132c15732412d22c1476fa83f27a",
    "readyToStream": true,
    "requireSignedURLs": true,
    "size": 4190963,
    "status": {
      "state": "inprogress",
      "pctComplete": "51",
      "errorReasonCode": "ERR_NON_VIDEO",
      "errorReasonText": "The file was not recognized as a valid video file."
    },
    "thumbnail": "https://videodelivery.net/ea95132c15732412d22c1476fa83f27a/thumbnails/thumbnail.jpg",
    "thumbnailTimestampPct": 0.529241,
    "uid": "ea95132c15732412d22c1476fa83f27a",
    "creator": "creator-id_abcde12345",
    "liveInput": "fc0a8dc887b16759bfd9ad922230a014",
    "uploaded": "2014-01-02T02:20:00Z",
    "watermark": {
      "uid": "ea95132c15732412d22c1476fa83f27a",
      "size": 29472,
      "height": 600,
      "width": 400,
      "created": "2014-01-02T02:20:00Z",
      "downloadedFrom": "https://company.com/logo.png",
      "name": "Marketing Videos",
      "opacity": 0.75,
      "padding": 0.1,
      "scale": 0.1,
      "position": "center"
    },
    "nft": {
      "contract": "0x57f1887a8bf19b14fc0d912b9b2acc9af147ea85",
      "token": 5
    },
	"scheduledDeletion": "2014-01-02T02:20:00Z"
  }
}
`
	testVideoID = "ea95132c15732412d22c1476fa83f27a"
)

var (
	TestVideoStruct = createTestVideo()
)

func createTestVideo() StreamVideo {
	created, _ := time.Parse(time.RFC3339, "2014-01-02T02:20:00Z")
	modified, _ := time.Parse(time.RFC3339, "2014-01-02T02:20:00Z")
	uploadexpiry, _ := time.Parse(time.RFC3339, "2014-01-02T02:20:00Z")
	uploaded, _ := time.Parse(time.RFC3339, "2014-01-02T02:20:00Z")
	scheduledDuration, _ := time.Parse(time.RFC3339, "2014-01-02T02:20:00Z")

	return StreamVideo{
		AllowedOrigins:     []string{"example.com"},
		Created:            &created,
		Duration:           300.5,
		Input:              StreamVideoInput{Height: 1080, Width: 1920},
		MaxDurationSeconds: 300,
		Modified:           &modified,
		UploadExpiry:       &uploadexpiry,
		Playback:           StreamVideoPlayback{Dash: "https://videodelivery.net/ea95132c15732412d22c1476fa83f27a/manifest/video.mpd", HLS: "https://videodelivery.net/ea95132c15732412d22c1476fa83f27a/manifest/video.m3u8"},
		Preview:            "https://watch.cloudflarestream.com/ea95132c15732412d22c1476fa83f27a",
		ReadyToStream:      true,
		RequireSignedURLs:  true,
		Size:               4190963,
		Status: StreamVideoStatus{
			State:           "inprogress",
			PctComplete:     "51",
			ErrorReasonCode: "ERR_NON_VIDEO",
			ErrorReasonText: "The file was not recognized as a valid video file.",
		},
		Thumbnail:             "https://videodelivery.net/ea95132c15732412d22c1476fa83f27a/thumbnails/thumbnail.jpg",
		ThumbnailTimestampPct: 0.529241,
		UID:                   "ea95132c15732412d22c1476fa83f27a",
		Creator:               "creator-id_abcde12345",
		LiveInput:             "fc0a8dc887b16759bfd9ad922230a014",
		Uploaded:              &uploaded,
		Watermark: StreamVideoWatermark{
			UID:            "ea95132c15732412d22c1476fa83f27a",
			Size:           29472,
			Height:         600,
			Width:          400,
			Created:        &created,
			DownloadedFrom: "https://company.com/logo.png",
			Name:           "Marketing Videos",
			Opacity:        0.75,
			Padding:        0.1,
			Scale:          0.1,
			Position:       "center",
		},
		Meta: map[string]interface{}{
			"name": "My First Stream Video",
		},
		NFT: StreamVideoNFTParameters{
			Token:    5,
			Contract: "0x57f1887a8bf19b14fc0d912b9b2acc9af147ea85",
		},
		ScheduledDeletion: &scheduledDuration,
	}
}

func TestStream_StreamUploadFromURL(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/stream/copy", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, singleStreamResponse)
	})

	// Make sure missing account ID is thrown
	_, err := client.StreamUploadFromURL(context.Background(), StreamUploadFromURLParameters{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	// Make sure missing upload URL is thrown
	_, err = client.StreamUploadFromURL(context.Background(), StreamUploadFromURLParameters{AccountID: testAccountID})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingUploadURL, err)
	}

	scheduledDuration, _ := time.Parse(time.RFC3339, "2014-01-02T02:20:00Z")

	want := TestVideoStruct
	input := StreamUploadFromURLParameters{
		AccountID: testAccountID,
		URL:       "https://example.com/myvideo.mp4",
		Meta: map[string]interface{}{
			"name": "My First Stream Video",
		},
		ScheduledDeletion: &scheduledDuration,
	}

	out, err := client.StreamUploadFromURL(context.Background(), input)
	if assert.NoError(t, err) {
		assert.Equal(t, out, want, "structs not equal")
	}
}

func TestStream_UploadVideoFile(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/stream", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, singleStreamResponse)
	})

	// Make sure missing account ID is thrown
	_, err := client.StreamUploadVideoFile(context.Background(), StreamUploadFileParameters{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	// Make sure missing file path is thrown
	_, err = client.StreamUploadVideoFile(context.Background(), StreamUploadFileParameters{AccountID: testAccountID})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingFilePath, err)
	}

	scheduledDuration, _ := time.Parse(time.RFC3339, "2014-01-02T02:20:00Z")

	input := StreamUploadFileParameters{
		AccountID:         testAccountID,
		VideoID:           testVideoID,
		FilePath:          "stream_test.go",
		ScheduledDeletion: &scheduledDuration,
	}

	out, err := client.StreamUploadVideoFile(context.Background(), input)

	want := TestVideoStruct
	if assert.NoError(t, err) {
		assert.Equal(t, out, want, "structs not equal")
	}
}

func TestStream_CreateVideoDirectURL(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/stream/direct_upload", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `
{
  "success": true,
  "errors": [],
  "messages": [],
  "result": {
    "uploadURL": "www.example.com/samplepath",
    "uid": "ea95132c15732412d22c1476fa83f27a",
    "watermark": {
      "uid": "ea95132c15732412d22c1476fa83f27a",
      "size": 29472,
      "height": 600,
      "width": 400,
      "created": "2014-01-02T02:20:00Z",
      "downloadedFrom": "https://company.com/logo.png",
      "name": "Marketing Videos",
      "opacity": 0.75,
      "padding": 0.1,
      "scale": 0.1,
      "position": "center"
    },
	"scheduledDeletion": "2014-01-02T02:20:00Z"
  }
}
`)
	})

	// Make sure AccountID is required
	_, err := client.StreamCreateVideoDirectURL(context.Background(), StreamCreateVideoParameters{})

	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	// Make sure MaxDuration is required
	_, err = client.StreamCreateVideoDirectURL(context.Background(), StreamCreateVideoParameters{AccountID: testAccountID})

	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingMaxDuration, err)
	}

	scheduledDuration, _ := time.Parse(time.RFC3339, "2014-01-02T02:20:00Z")

	input := StreamCreateVideoParameters{
		AccountID:          testAccountID,
		MaxDurationSeconds: 300,
		Meta: map[string]interface{}{
			"name": "My First Stream Video",
		},
		ScheduledDeletion: &scheduledDuration,
	}

	out, err := client.StreamCreateVideoDirectURL(context.Background(), input)

	created, _ := time.Parse(time.RFC3339, "2014-01-02T02:20:00Z")

	want := StreamVideoCreate{
		UploadURL: "www.example.com/samplepath",
		UID:       "ea95132c15732412d22c1476fa83f27a",
		Watermark: StreamVideoWatermark{
			UID:            "ea95132c15732412d22c1476fa83f27a",
			Size:           29472,
			Height:         600,
			Width:          400,
			Created:        &created,
			DownloadedFrom: "https://company.com/logo.png",
			Name:           "Marketing Videos",
			Opacity:        0.75,
			Padding:        0.1,
			Scale:          0.1,
			Position:       "center",
		},
		ScheduledDeletion: &scheduledDuration,
	}

	if assert.NoError(t, err) {
		assert.Equal(t, out, want, "structs not equal")
	}
}

func TestStream_ListVideos(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/stream", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `
{
  "success": true,
  "errors": [],
  "messages": [],
  "result": [{
    "allowedOrigins": [
      "example.com"
    ],
    "created": "2014-01-02T02:20:00Z",
    "duration": 300.5,
    "input": {
      "height": 1080,
      "width": 1920
    },
    "maxDurationSeconds": 300,
    "meta": {
	  "name": "My First Stream Video"
	},
    "modified": "2014-01-02T02:20:00Z",
    "uploadExpiry": "2014-01-02T02:20:00Z",
    "playback": {
      "hls": "https://videodelivery.net/ea95132c15732412d22c1476fa83f27a/manifest/video.m3u8",
      "dash": "https://videodelivery.net/ea95132c15732412d22c1476fa83f27a/manifest/video.mpd"
    },
    "preview": "https://watch.cloudflarestream.com/ea95132c15732412d22c1476fa83f27a",
    "readyToStream": true,
    "requireSignedURLs": true,
    "size": 4190963,
    "status": {
      "state": "inprogress",
      "pctComplete": "51",
      "errorReasonCode": "ERR_NON_VIDEO",
      "errorReasonText": "The file was not recognized as a valid video file."
    },
    "thumbnail": "https://videodelivery.net/ea95132c15732412d22c1476fa83f27a/thumbnails/thumbnail.jpg",
    "thumbnailTimestampPct": 0.529241,
    "uid": "ea95132c15732412d22c1476fa83f27a",
    "creator": "creator-id_abcde12345",
    "liveInput": "fc0a8dc887b16759bfd9ad922230a014",
    "uploaded": "2014-01-02T02:20:00Z",
    "watermark": {
      "uid": "ea95132c15732412d22c1476fa83f27a",
      "size": 29472,
      "height": 600,
      "width": 400,
      "created": "2014-01-02T02:20:00Z",
      "downloadedFrom": "https://company.com/logo.png",
      "name": "Marketing Videos",
      "opacity": 0.75,
      "padding": 0.1,
      "scale": 0.1,
      "position": "center"
    },
    "nft": {
      "contract": "0x57f1887a8bf19b14fc0d912b9b2acc9af147ea85",
      "token": 5
    },
 	"scheduledDeletion": "2014-01-02T02:20:00Z"
  }]
}
`)
	})

	// Make sure AccountID is required
	_, err := client.StreamListVideos(context.Background(), StreamListParameters{})

	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	out, err := client.StreamListVideos(context.Background(), StreamListParameters{AccountID: testAccountID})
	want := TestVideoStruct

	if assert.NoError(t, err) {
		assert.Equal(t, len(out), 1, "length of videos is not one")
		assert.Equal(t, out[0], want, "structs not equal")
	}
}

func TestStream_GetVideo(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/accounts/"+testAccountID+"/stream/"+testVideoID, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, singleStreamResponse)
	})

	// Make sure AccountID is required
	_, err := client.StreamGetVideo(context.Background(), StreamParameters{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	// Make sure VideoID is required
	_, err = client.StreamGetVideo(context.Background(), StreamParameters{AccountID: testAccountID})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingVideoID, err)
	}

	input := StreamParameters{AccountID: testAccountID, VideoID: testVideoID}
	out, err := client.StreamGetVideo(context.Background(), input)

	want := TestVideoStruct

	if assert.NoError(t, err) {
		assert.Equal(t, want, out, "structs not equal")
	}
}

func TestStream_DeleteVideo(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/stream/"+testVideoID, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method, "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"success": true,
			"errors": [],
			"messages": [],
			"result": {}
		}`)
	})

	// Make sure AccountID is required
	err := client.StreamDeleteVideo(context.Background(), StreamParameters{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	// Make sure VideoID is required
	err = client.StreamDeleteVideo(context.Background(), StreamParameters{AccountID: testAccountID})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingVideoID, err)
	}

	input := StreamParameters{AccountID: testAccountID, VideoID: testVideoID}
	err = client.StreamDeleteVideo(context.Background(), input)
	require.NoError(t, err)
}

func TestStream_EmbedHTML(t *testing.T) {
	setup()
	defer teardown()

	streamHTML := `<stream id="ea95132c15732412d22c1476fa83f27a"></stream><script data-cfasync="false" defer type="text/javascript" src="https://embed.cloudflarestream.com/embed/we4g.fla9.latest.js"></script>`
	mux.HandleFunc("/accounts/"+testAccountID+"/stream/"+testVideoID+"/embed", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "text/html")
		fmt.Fprint(w, streamHTML)
	})

	// Make sure AccountID is required
	_, err := client.StreamEmbedHTML(context.Background(), StreamParameters{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	// Make sure VideoID is required
	_, err = client.StreamEmbedHTML(context.Background(), StreamParameters{AccountID: testAccountID})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingVideoID, err)
	}

	input := StreamParameters{AccountID: testAccountID, VideoID: testVideoID}
	out, err := client.StreamEmbedHTML(context.Background(), input)
	if assert.NoError(t, err) {
		assert.Equal(t, streamHTML, out, "bad html output")
	}
}

func TestStream_AssociateNFT(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/stream/"+testVideoID, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		fmt.Fprint(w, singleStreamResponse)
	})

	// Make sure AccountID is required
	_, err := client.StreamAssociateNFT(context.Background(), StreamVideoNFTParameters{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	// Make sure VideoID is required
	_, err = client.StreamAssociateNFT(context.Background(), StreamVideoNFTParameters{AccountID: testAccountID})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingVideoID, err)
	}

	input := StreamVideoNFTParameters{AccountID: testAccountID, VideoID: testVideoID, Token: 5, Contract: "0x57f1887a8bf19b14fc0d912b9b2acc9af147ea85"}
	out, err := client.StreamAssociateNFT(context.Background(), input)

	want := TestVideoStruct

	if assert.NoError(t, err) {
		assert.Equal(t, want, out, "structs not equal")
	}
}

func TestStream_CreateSignedURL(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/stream/"+testVideoID+"/token", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		fmt.Fprint(w, `{
  "success": true,
  "errors": [],
  "messages": [],
  "result": {
    "token": "eyJhbGciOiJSUzI1NiIsImtpZCI6ImU5ZGI5OTBhODI2NjZkZDU3MWM3N2Y5NDRhNWM1YzhkIn0.eyJzdWIiOiJlYTk1MTMyYzE1NzMyNDEyZDIyYzE0NzZmYTgzZjI3YSIsImtpZCI6ImU5ZGI5OTBhODI2NjZkZDU3MWM3N2Y5NDRhNWM1YzhkIiwiZXhwIjoiMTUzNzQ2MDM2NSIsIm5iZiI6IjE1Mzc0NTMxNjUifQ.OZhqOARADn1iubK6GKcn25hN3nU-hCFF5q9w2C4yup0C4diG7aMIowiRpP-eDod8dbAJubsiFuTKrqPcmyCKWYsiv0TQueukqbQlF7HCO1TV-oF6El5-7ldJ46eD-ZQ0XgcIYEKrQOYFF8iDQbqPm3REWd6BnjKZdeVrLzuRaiSnZ9qqFpGu5dfxIY9-nZKDubJHqCr3Imtb211VIG_b9MdtO92JjvkDS-rxT_pkEfTZSafl1OU-98A7KBGtPSJHz2dHORIrUiTA6on4eIXTj9aFhGiir4rSn-rn0OjPRTtJMWIDMoQyE_fwrSYzB7MPuzL2t82BWaEbHZTfixBm5A"
  }
}`)
	})

	// Make sure AccountID is required
	_, err := client.StreamCreateSignedURL(context.Background(), StreamSignedURLParameters{})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountID, err)
	}

	// Make sure VideoID is required
	_, err = client.StreamCreateSignedURL(context.Background(), StreamSignedURLParameters{AccountID: testAccountID})
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingVideoID, err)
	}

	input := StreamSignedURLParameters{AccountID: testAccountID, VideoID: testVideoID}
	out, err := client.StreamCreateSignedURL(context.Background(), input)

	want := "eyJhbGciOiJSUzI1NiIsImtpZCI6ImU5ZGI5OTBhODI2NjZkZDU3MWM3N2Y5NDRhNWM1YzhkIn0.eyJzdWIiOiJlYTk1MTMyYzE1NzMyNDEyZDIyYzE0NzZmYTgzZjI3YSIsImtpZCI6ImU5ZGI5OTBhODI2NjZkZDU3MWM3N2Y5NDRhNWM1YzhkIiwiZXhwIjoiMTUzNzQ2MDM2NSIsIm5iZiI6IjE1Mzc0NTMxNjUifQ.OZhqOARADn1iubK6GKcn25hN3nU-hCFF5q9w2C4yup0C4diG7aMIowiRpP-eDod8dbAJubsiFuTKrqPcmyCKWYsiv0TQueukqbQlF7HCO1TV-oF6El5-7ldJ46eD-ZQ0XgcIYEKrQOYFF8iDQbqPm3REWd6BnjKZdeVrLzuRaiSnZ9qqFpGu5dfxIY9-nZKDubJHqCr3Imtb211VIG_b9MdtO92JjvkDS-rxT_pkEfTZSafl1OU-98A7KBGtPSJHz2dHORIrUiTA6on4eIXTj9aFhGiir4rSn-rn0OjPRTtJMWIDMoQyE_fwrSYzB7MPuzL2t82BWaEbHZTfixBm5A"

	if assert.NoError(t, err) {
		assert.Equal(t, want, out, "structs not equal")
	}
}

func TestStream_TUSUploadMetadataToTUSCsv(t *testing.T) {
	md := TUSUploadMetadata{
		Name: "test.mp4",
	}
	csv, err := md.ToTUSCsv()
	assert.NoError(t, err)
	assert.Equal(t, "name dGVzdC5tcDQ=", csv)

	md.RequireSignedURLs = true
	csv, err = md.ToTUSCsv()
	assert.NoError(t, err)
	assert.Equal(t, "name dGVzdC5tcDQ=,requiresignedurls", csv)

	md.AllowedOrigins = "example.com"
	csv, err = md.ToTUSCsv()
	assert.NoError(t, err)
	assert.Equal(t, "name dGVzdC5tcDQ=,requiresignedurls,allowedorigins ZXhhbXBsZS5jb20=", csv)

	md.ThumbnailTimestampPct = 0.5
	csv, err = md.ToTUSCsv()
	assert.NoError(t, err)
	assert.Equal(t, "name dGVzdC5tcDQ=,requiresignedurls,allowedorigins ZXhhbXBsZS5jb20=,thumbnailtimestamppct MC41", csv)

	scheduleDeletion, _ := time.Parse(time.RFC3339, "2023-10-01T02:20:00Z")
	md.ScheduledDeletion = &scheduleDeletion
	csv, err = md.ToTUSCsv()
	assert.NoError(t, err)
	assert.Equal(t, "name dGVzdC5tcDQ=,requiresignedurls,allowedorigins ZXhhbXBsZS5jb20=,thumbnailtimestamppct MC41,scheduledDeletion MjAyMy0xMC0wMVQwMjoyMDowMFo=", csv)

	expiry, _ := time.Parse(time.RFC3339, "2023-09-25T02:45:00Z")
	md.Expiry = &expiry
	csv, err = md.ToTUSCsv()
	assert.NoError(t, err)
	assert.Equal(t, "name dGVzdC5tcDQ=,requiresignedurls,allowedorigins ZXhhbXBsZS5jb20=,thumbnailtimestamppct MC41,scheduledDeletion MjAyMy0xMC0wMVQwMjoyMDowMFo=,expiry MjAyMy0wOS0yNVQwMjo0NTowMFo=", csv)

	md.Watermark = "watermark-profile-uid"
	csv, err = md.ToTUSCsv()
	assert.NoError(t, err)
	assert.Equal(t, "name dGVzdC5tcDQ=,requiresignedurls,allowedorigins ZXhhbXBsZS5jb20=,thumbnailtimestamppct MC41,scheduledDeletion MjAyMy0xMC0wMVQwMjoyMDowMFo=,expiry MjAyMy0wOS0yNVQwMjo0NTowMFo=,watermark d2F0ZXJtYXJrLXByb2ZpbGUtdWlk", csv)

	md.MaxDurationSeconds = 300
	csv, err = md.ToTUSCsv()
	assert.NoError(t, err)
	assert.Equal(t, "name dGVzdC5tcDQ=,maxDurationSeconds MzAw,requiresignedurls,allowedorigins ZXhhbXBsZS5jb20=,thumbnailtimestamppct MC41,scheduledDeletion MjAyMy0xMC0wMVQwMjoyMDowMFo=,expiry MjAyMy0wOS0yNVQwMjo0NTowMFo=,watermark d2F0ZXJtYXJrLXByb2ZpbGUtdWlk", csv)

	// empty metadata should return empty string
	md = TUSUploadMetadata{}
	csv, err = md.ToTUSCsv()
	assert.NoError(t, err)
	assert.Equal(t, "", csv)
}

func TestStream_StreamInitiateTUSVideoUpload(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/accounts/"+testAccountID+"/stream", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Expected method 'POST', got %s", r.Method)
		// Make sure Tus-Resumable header is set
		assert.Equal(t, "1.0.0", r.Header.Get("Tus-Resumable"))
		// Make sure Upload-Length header is set
		assert.Equal(t, "123", r.Header.Get("Upload-Length"))
		// set the response headers
		// if query param direct_user=true, then return the direct url in the header
		if r.URL.Query().Get("direct_user") == "true" {
			w.Header().Set("Location", "https://upload.videodelivery.net/tus/90c68cb5cd4fd5350b1962279c90bec0?tusv2=true")
		} else {
			w.Header().Set("Location", "https://api.cloudflare.com/client/v4/accounts/"+testAccountID+"/media/278f2a7e763c73dedc064b965d2cfbed?tusv2=true")
		}

		w.Header().Set("stream-media-id", "278f2a7e763c73dedc064b965d2cfbed")
		w.Header().Set("Tus-Resumable", "1.0.0")
		w.WriteHeader(http.StatusCreated)
	})

	// Make sure Tus-Resumable header is set
	params := StreamInitiateTUSUploadParameters{}
	_, err := client.StreamInitiateTUSVideoUpload(context.Background(), AccountIdentifier(testAccountID), params)
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingTusResumable, err)
	}
	params.TusResumable = TusProtocolVersion1_0_0

	// Make sure Upload-Length header is set
	_, err = client.StreamInitiateTUSVideoUpload(context.Background(), AccountIdentifier(testAccountID), params)
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingUploadLength, err)
	}
	params.UploadLength = 123

	out, err := client.StreamInitiateTUSVideoUpload(context.Background(), AccountIdentifier(testAccountID), params)
	if assert.NoError(t, err) {
		assert.Equal(t, "https://api.cloudflare.com/client/v4/accounts/"+testAccountID+"/media/278f2a7e763c73dedc064b965d2cfbed?tusv2=true", out.ResponseHeaders.Get("Location"))
		assert.Equal(t, "278f2a7e763c73dedc064b965d2cfbed", out.ResponseHeaders.Get("stream-media-id"))
		assert.Equal(t, "1.0.0", out.ResponseHeaders.Get("Tus-Resumable"))
	}

	params.DirectUserUpload = true
	out, err = client.StreamInitiateTUSVideoUpload(context.Background(), AccountIdentifier(testAccountID), params)
	if assert.NoError(t, err) {
		assert.Equal(t, "https://upload.videodelivery.net/tus/90c68cb5cd4fd5350b1962279c90bec0?tusv2=true", out.ResponseHeaders.Get("Location"))
		assert.Equal(t, "278f2a7e763c73dedc064b965d2cfbed", out.ResponseHeaders.Get("stream-media-id"))
		assert.Equal(t, "1.0.0", out.ResponseHeaders.Get("Tus-Resumable"))
	}
}
