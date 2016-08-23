package zone

import (
	"github.com/prasmussen/gandi-api/client"
	"github.com/prasmussen/gandi-api/domain"
)

type Zone struct {
	*client.Client
}

func New(c *client.Client) *Zone {
	return &Zone{c}
}

// Counts accessible zones
func (self *Zone) Count() (int64, error) {
	var result int64
	params := []interface{}{self.Key}
	if err := self.Call("domain.zone.count", params, &result); err != nil {
		return -1, err
	}
	return result, nil
}

// Get zone information
func (self *Zone) Info(id int64) (*ZoneInfo, error) {
	var res map[string]interface{}
	params := []interface{}{self.Key, id}
	if err := self.Call("domain.zone.info", params, &res); err != nil {
		return nil, err
	}
	return ToZoneInfo(res), nil
}

// List accessible DNS zones.
func (self *Zone) List() ([]*ZoneInfoBase, error) {
	var res []interface{}
	params := []interface{}{self.Key}
	if err := self.Call("domain.zone.list", params, &res); err != nil {
		return nil, err
	}

	zones := make([]*ZoneInfoBase, 0)
	for _, r := range res {
		zone := ToZoneInfoBase(r.(map[string]interface{}))
		zones = append(zones, zone)
	}
	return zones, nil
}

// Create a zone
func (self *Zone) Create(name string) (*ZoneInfo, error) {
	var res map[string]interface{}
	createArgs := map[string]interface{}{"name": name}
	params := []interface{}{self.Key, createArgs}
	if err := self.Call("domain.zone.create", params, &res); err != nil {
		return nil, err
	}
	return ToZoneInfo(res), nil
}

// Delete a zone
func (self *Zone) Delete(id int64) (bool, error) {
	var res bool
	params := []interface{}{self.Key, id}
	if err := self.Call("domain.zone.delete", params, &res); err != nil {
		return false, err
	}
	return res, nil
}

// Set the current zone of a domain
func (self *Zone) Set(domainName string, id int64) (*domain.DomainInfo, error) {
	var res map[string]interface{}
	params := []interface{}{self.Key, domainName, id}
	if err := self.Call("domain.zone.set", params, &res); err != nil {
		return nil, err
	}
	return domain.ToDomainInfo(res), nil
}
