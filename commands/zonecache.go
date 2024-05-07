package commands

import "github.com/StackExchange/dnscontrol/v4/providers"

// NewZoneCache creates a zoneCache.
func NewZoneCache() *zoneCache {
	return &zoneCache{}
}

func (zc *zoneCache) zoneList(name string, lister providers.ZoneLister) (*[]string, error) {
	zc.Lock()
	defer zc.Unlock()

	if zc.cache == nil {
		zc.cache = map[string]*[]string{}
	}

	if v, ok := zc.cache[name]; ok {
		return v, nil
	}

	zones, err := lister.ListZones()
	if err != nil {
		return nil, err
	}
	zc.cache[name] = &zones
	return &zones, nil
}
