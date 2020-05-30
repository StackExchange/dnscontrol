package servers

import "context"

// Client defines method for interacting with the PowerDNS "Servers" endpoints
type Client interface {

	// ListServers lists all known servers
	ListServers(ctx context.Context) ([]Server, error)

	// GetServer returns a specific server. If the server with the given "serverID" does
	// not exist, the error return value will contain a pdnshttp.ErrNotFound error (see example)
	GetServer(ctx context.Context, serverID string) (*Server, error)
}
