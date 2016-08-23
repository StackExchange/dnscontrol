package version

import "github.com/prasmussen/gandi-api/client"

type Version struct {
	*client.Client
}

func New(c *client.Client) *Version {
	return &Version{c}
}

// Count this zone versions
func (self *Version) Count(zoneId int64) (int64, error) {
	var result int64
	params := []interface{}{self.Key, zoneId}
	if err := self.Call("domain.zone.version.count", params, &result); err != nil {
		return -1, err
	}
	return result, nil
}

// List this zone versions, with their creation date
func (self *Version) List(zoneId int64) ([]*VersionInfo, error) {
	var res []interface{}
	params := []interface{}{self.Key, zoneId}
	if err := self.Call("domain.zone.version.list", params, &res); err != nil {
		return nil, err
	}

	versions := make([]*VersionInfo, 0)
	for _, r := range res {
		version := ToVersionInfo(r.(map[string]interface{}))
		versions = append(versions, version)
	}
	return versions, nil
}

// Create a new version from another version. This will duplicate the version’s records
func (self *Version) New(zoneId, version int64) (int64, error) {
	var res int64

	params := []interface{}{self.Key, zoneId, version}
	if err := self.Call("domain.zone.version.new", params, &res); err != nil {
		return -1, err
	}
	return res, nil
}

// Delete a specific version
func (self *Version) Delete(zoneId, version int64) (bool, error) {
	var res bool
	params := []interface{}{self.Key, zoneId, version}
	if err := self.Call("domain.zone.version.delete", params, &res); err != nil {
		return false, err
	}
	return res, nil
}

// Set the active version of a zone
func (self *Version) Set(zoneId, version int64) (bool, error) {
	var res bool
	params := []interface{}{self.Key, zoneId, version}
	if err := self.Call("domain.zone.version.set", params, &res); err != nil {
		return false, err
	}
	return res, nil
}
