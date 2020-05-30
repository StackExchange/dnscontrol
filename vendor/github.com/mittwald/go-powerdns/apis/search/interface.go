package search

import "context"

// Client defines method for interacting with the PowerDNS "Search" endpoints
type Client interface {

	// ListServers lists all known servers
	Search(ctx context.Context, serverID, query string, max int, objectType ObjectType) (ResultList, error)
}
