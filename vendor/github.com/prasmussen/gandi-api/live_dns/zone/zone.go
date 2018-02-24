package zone

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/prasmussen/gandi-api/client"
	"github.com/prasmussen/gandi-api/live_dns/domain"
	"github.com/prasmussen/gandi-api/live_dns/record"
)

// Enable/disable debug output:
const debug = false

// Zone holds the zone client structure
type Zone struct {
	*client.Client
}

// New instanciates a new instance of a Zone client
func New(c *client.Client) *Zone {
	return &Zone{c}
}

// List accessible DNS zones.
func (z *Zone) List() (zones []*Info, err error) {
	_, err = z.Get("/zones", &zones)
	return
}

// InfoByUUID Gets zone information from its UUID
func (z *Zone) InfoByUUID(uuid uuid.UUID) (info *Info, err error) {
	_, err = z.Get(fmt.Sprintf("/zones/%s", uuid), &info)
	if debug {
		fmt.Printf("DEBUG: InfoByUUID returned SharingID=%v domain=%v\n", info.SharingID, info.Name)
	}
	return
}

// Info Gets zone information
func (z *Zone) Info(zoneInfo Info) (info *Info, err error) {
	if zoneInfo.UUID == nil {
		return nil, fmt.Errorf("can not get zone info %s without an id", zoneInfo.Name)
	}
	return z.InfoByUUID(*zoneInfo.UUID)
}

// Create creates a new zone
func (z *Zone) Create(zoneInfo Info) (status *CreateStatus, err error) {
	if debug {
		fmt.Printf("DEBUG: Create WILL SET SharingID=%v domain=%v\n", zoneInfo.SharingID, zoneInfo.Name)
	}
	_, err = z.Post(fmt.Sprintf("/zones?sharing_id=%s", zoneInfo.SharingID), zoneInfo, &status)
	return
}

// Update updates an existing zone
func (z *Zone) Update(zoneInfo Info) (status *Status, err error) {
	if zoneInfo.UUID == nil {
		return nil, fmt.Errorf("can not update zone %s without an id", zoneInfo.Name)
	}
	_, err = z.Patch(fmt.Sprintf("/zones/%s", zoneInfo.UUID), zoneInfo, &status)
	return
}

// Delete Deletes an existing zone
func (z *Zone) Delete(zoneInfo Info) (err error) {
	if zoneInfo.UUID == nil {
		return fmt.Errorf("can not update zone %s without an id", zoneInfo.Name)
	}
	_, err = z.Client.Delete(fmt.Sprintf("/zones/%s", zoneInfo.UUID), nil)
	return
}

// Domains lists all domains using a zone
func (z *Zone) Domains(zoneInfo Info) (domains []*domain.InfoBase, err error) {
	if zoneInfo.UUID == nil {
		return nil, fmt.Errorf("can not get domains on a zone %s without an id", zoneInfo.Name)
	}
	_, err = z.Get(fmt.Sprintf("/zones/%s/domains", zoneInfo.UUID), &domains)
	return

}

// Set the current zone of a domain
func (z *Zone) Set(domainName string, zoneInfo Info) (status *Status, err error) {
	if zoneInfo.UUID == nil {
		return nil, fmt.Errorf("can not attach a domain %s to a zone %s without an id", domainName, zoneInfo.Name)
	}
	if debug {
		fmt.Printf("DEBUG: Set WILL SET SharingID=%v domain=%s dn=%v\n", zoneInfo.SharingID, domainName, zoneInfo.Name)
	}
	_, err = z.Post(fmt.Sprintf("/zones/%s/domains/%s", zoneInfo.UUID, domainName), nil, &status)
	return
}

// Records gets a record client for the current zone
func (z *Zone) Records(zoneInfo Info) record.Manager {
	return record.New(z.Client, fmt.Sprintf("/zones/%s", zoneInfo.UUID))
}
