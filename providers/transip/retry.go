package transip

import (
	"log"
	"time"

	"gopkg.in/ns1/ns1-go.v2/rest"
)

const (
	initialBackoff = time.Second * 20 // First delay duration
	maxBackoff     = time.Minute * 4  // Maximum backoff delay
)

// backoff is the amount of time to sleep if a 429 (or similar) is received.
// It is doubled after each use.
var (
	backoff = initialBackoff
)

func retryNeeded(err error) bool {
	if err == nil {
		return false // Not an error.
	}

	serr, ok := err.(*rest.Error)
	if !ok {
		return false // Not an error we know how to work with.
	}

	if serr.StatusCode == 200 {
		backoff = initialBackoff // Reset
		return false             // Success! No need to retry.
	}

	if serr.StatusCode != 429 {
		return false
	}

	// a simple exponential back-off
	log.Printf("Pausing due to ratelimit (%03d): %v seconds\n", serr.StatusCode, backoff)
	time.Sleep(backoff)
	backoff = backoff + (backoff / 2)
	if backoff > maxBackoff {
		backoff = maxBackoff
	}

	return true // Request the API call be re-tried.
}
