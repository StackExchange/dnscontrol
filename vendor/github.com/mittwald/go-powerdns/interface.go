package pdns

import (
	"context"

	"github.com/mittwald/go-powerdns/apis/cache"
	"github.com/mittwald/go-powerdns/apis/search"
	"github.com/mittwald/go-powerdns/apis/servers"
	"github.com/mittwald/go-powerdns/apis/zones"
)

// Client is the root-level interface for interacting with the PowerDNS API.
// You can instantiate an implementation of this interface using the "New" function.
type Client interface {

	// Status checks if the PowerDNS API is reachable. This does a simple HTTP connection check;
	// it will NOT check if your authentication is set up correctly (except you're using TLS client
	// authentication.
	Status() error

	// WaitUntilUp will block until the PowerDNS API accepts HTTP requests. You can use the "ctx"
	// parameter to make this method wait only for (or until) a certain time (see examples).
	WaitUntilUp(ctx context.Context) error

	// Servers returns a specialized API for interacting with PowerDNS servers
	Servers() servers.Client

	// Zones returns a specialized API for interacting with PowerDNS zones
	Zones() zones.Client

	// Search returns a specialized API for searching
	Search() search.Client

	// Cache returns a specialized API for caching
	Cache() cache.Client
}
