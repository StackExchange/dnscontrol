package domain

import (
	"github.com/prasmussen/gandi-api/client"
	"github.com/prasmussen/gandi-api/operation"
)

type Domain struct {
	*client.Client
}

func New(c *client.Client) *Domain {
	return &Domain{c}
}

// Check the availability of some domain
func (self *Domain) Available(name string) (string, error) {
	var result map[string]interface{}
	params := []interface{}{self.Key, name}
	if err := self.Call("domain.available", params, &result); err != nil {
		return "", err
	}
	return result[name].(string), nil
}

// Get domain information
func (self *Domain) Info(name string) (*DomainInfo, error) {
	var res map[string]interface{}
	params := []interface{}{self.Key, name}
	if err := self.Call("domain.info", params, &res); err != nil {
		return nil, err
	}
	return ToDomainInfo(res), nil
}

// List domains associated to the contact represented by apikey
func (self *Domain) List() ([]*DomainInfoBase, error) {
	var res []interface{}
	params := []interface{}{self.Key}
	if err := self.Call("domain.list", params, &res); err != nil {
		return nil, err
	}

	domains := make([]*DomainInfoBase, 0)
	for _, r := range res {
		domain := ToDomainInfoBase(r.(map[string]interface{}))
		domains = append(domains, domain)
	}
	return domains, nil
}

// Count domains associated to the contact represented by apikey
func (self *Domain) Count() (int64, error) {
	var result int64
	params := []interface{}{self.Key}
	if err := self.Call("domain.count", params, &result); err != nil {
		return -1, err
	}
	return result, nil
}

// Create a domain
func (self *Domain) Create(name, contactHandle string, years int64) (*operation.OperationInfo, error) {
	var res map[string]interface{}
	createArgs := map[string]interface{}{
		"admin":    contactHandle,
		"bill":     contactHandle,
		"owner":    contactHandle,
		"tech":     contactHandle,
		"duration": years,
	}
	params := []interface{}{self.Key, name, createArgs}
	if err := self.Call("domain.create", params, &res); err != nil {
		return nil, err
	}
	return operation.ToOperationInfo(res), nil
}
