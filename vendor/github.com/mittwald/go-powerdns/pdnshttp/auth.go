package pdnshttp

import (
	"net/http"
)

type ClientAuthenticator interface {
	OnRequest(*http.Request) error
	OnConnect(*http.Client) error
}

// NoopAuthenticator provides an "empty" implementation of the
// ClientAuthenticator interface.
type NoopAuthenticator struct{}

// OnRequest is applied each time a HTTP request is built.
func (NoopAuthenticator) OnRequest(*http.Request) error {
	return nil
}

// OnConnect is applied on the entire connection as soon as it is set up.
func (NoopAuthenticator) OnConnect(*http.Client) error {
	return nil
}
