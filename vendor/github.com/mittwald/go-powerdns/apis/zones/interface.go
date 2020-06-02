package zones

import "context"

// Client defines the interface for Zone operations.
type Client interface {
	// ListZones lists known zones for a given serverID
	ListZones(ctx context.Context, serverID string) ([]Zone, error)

	// ListZone list known zone for a given serverID and zoneID
	ListZone(ctx context.Context, serverID string, zoneID string) ([]Zone, error)

	// CreateZone creates a new zone for a given server.
	CreateZone(ctx context.Context, serverID string, zone Zone) (*Zone, error)

	// GetZone returns an existing zone by ID. If not found, the first returned value
	// will be nil, and the error return value will be an instance of "pdnshttp.ErrNotFound".
	GetZone(ctx context.Context, serverID string, zoneID string) (*Zone, error)

	// DeleteZone deletes a zone. No shit.
	DeleteZone(ctx context.Context, serverID string, zoneID string) error

	// AddRecordSetToZone will add a new set of records to a zone. Existing record sets for
	// the exact name/type combination will be replaced.
	AddRecordSetToZone(ctx context.Context, serverID string, zoneID string, set ResourceRecordSet) error

	// RemoveRecordSetFromZone removes a record set from a zone. The record set is matched
	// by name and type.
	RemoveRecordSetFromZone(ctx context.Context, serverID string, zoneID string, name string, recordType string) error

	// RetrieveFromMaster retrieves a slave zone from its master
	RetrieveFromMaster(ctx context.Context, serverID string, zoneID string) error

	// NotifySlaves sends a DNS NOTIFY to all slaves
	NotifySlaves(ctx context.Context, serverID string, zoneID string) error

	// ExportZone exports the entire zone in AXFR format
	ExportZone(ctx context.Context, serverID string, zoneID string) ([]byte, error)

	// VerifyZone verifies a zone's configuration
	VerifyZone(ctx context.Context, serverID string, zoneID string) error

	// RectifyZone rectifies the zone data
	RectifyZone(ctx context.Context, serverID string, zoneID string) error
}
