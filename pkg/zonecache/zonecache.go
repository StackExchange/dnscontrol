package zonecache

import (
	"errors"
	"maps"
	"sync"
)

func New[Zone any](fetchAll func() (map[string]Zone, error)) ZoneCache[Zone] {
	return ZoneCache[Zone]{fetchAll: fetchAll}
}

var ErrZoneNotFound = errors.New("zone not found")

type ZoneCache[Zone any] struct {
	mu       sync.Mutex
	cached   bool
	cache    map[string]Zone
	fetchAll func() (map[string]Zone, error)
}

func (c *ZoneCache[Zone]) ensureCached() error {
	if c.cached {
		return nil
	}
	zones, err := c.fetchAll()
	if err != nil {
		return err
	}
	if c.cache == nil {
		c.cache = make(map[string]Zone, len(zones))
	}
	maps.Copy(c.cache, zones)
	c.cached = true
	return nil
}

func (c *ZoneCache[Zone]) HasZone(name string) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureCached(); err != nil {
		return false, err
	}
	_, ok := c.cache[name]
	return ok, nil
}

func (c *ZoneCache[Zone]) GetZone(name string) (Zone, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureCached(); err != nil {
		var z Zone
		return z, err
	}
	z, ok := c.cache[name]
	if !ok {
		return z, ErrZoneNotFound
	}
	return z, nil
}

func (c *ZoneCache[Zone]) GetZoneNames() ([]string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureCached(); err != nil {
		return nil, err
	}
	names := make([]string, 0, len(c.cache))
	for name := range c.cache {
		names = append(names, name)
	}
	return names, nil
}

func (c *ZoneCache[Zone]) SetZone(name string, z Zone) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cache == nil {
		c.cache = make(map[string]Zone, 1)
	}
	c.cache[name] = z
}
