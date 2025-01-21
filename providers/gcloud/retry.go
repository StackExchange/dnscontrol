package gcloud

import (
	"log"
	"time"

	"google.golang.org/api/googleapi"
)

const (
	initialBackoff = time.Second * 10 // First delay duration
	maxBackoff     = time.Minute * 3  // Maximum backoff delay
)

// backoff is the amount of time to sleep if a 429 or 504 is received.
// It is doubled after each use.
var (
	backoff    = initialBackoff
	backoff404 = false // Set if the last call requested a retry of a 404
	backoff502 = false // Set if the last call requested a retry of a 502
)

func retryNeeded(resp *googleapi.ServerResponse, err error) bool {
	if err == nil {
		return false // Not an error.
	}

	serr, ok := err.(*googleapi.Error)
	if !ok {
		return false // Not a google error.
	}

	if serr.Code == 200 {
		backoff = initialBackoff // Reset
		return false             // Success! No need to retry.
	}

	if serr.Code == 404 {
		// serr.Code == 404 happens occasionally when GCLOUD hasn't
		// finished updating the database yet.  We pause and retry
		// exactly once. There should be a better way to do this, such as
		// a callback that would tell us a transaction is complete.
		if backoff404 {
			backoff404 = false
			return false // Give up. We've done this already.
		}
		log.Printf("Special 404 pause-and-retry for GCLOUD: Pausing %s\n", backoff)
		time.Sleep(backoff)
		backoff404 = true
		return true // Request a retry.
	}
	backoff404 = false

	if serr.Code == 502 {
		// serr.Code == 502 happens occasionally when "The server
		// encountered a temporary error and could not complete your
		// request. Please try again in 30 seconds.  That’s all we know."
		// We pause and retry exactly once.
		if backoff502 {
			backoff502 = false
			return false // Give up. We've done this already.
		}
		log.Printf("Special 502 pause-and-retry for GCLOUD: Pausing %s\n", backoff)
		time.Sleep(31 * time.Second)
		backoff502 = true
		return true // Request a retry.
	}
	backoff502 = false

	if serr.Code != 429 && serr.Code != 503 {
		return false // Not an error that permits retrying.
	}

	// TODO(tlim): In theory, resp.Header has a header that says how
	// long to wait but I haven't been able to capture that header in
	// the wild. If you get these "RUNCHANGE HEAD" messages, please
	// file a bug with the contents!

	if resp != nil {
		log.Printf("NOTE: If you see this message, please file a bug with the output below:\n")
		log.Printf("RUNCHANGE CODE = %+v\n", resp.HTTPStatusCode)
		log.Printf("RUNCHANGE HEAD = %+v\n", resp.Header)
	}

	// a simple exponential back-off
	log.Printf("Pausing due to ratelimit: %v seconds\n", backoff)
	time.Sleep(backoff)
	backoff = backoff + (backoff / 2)
	if backoff > maxBackoff {
		backoff = maxBackoff
	}

	return true // Request the API call be re-tried.
}
