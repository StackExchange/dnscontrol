package cache

import "context"

// Client defines the interface for Cache operations.
type Client interface {
	// Flush flush a cache-entry by name
	Flush(ctx context.Context, serverID string, name string) (*FlushResult, error)
}
