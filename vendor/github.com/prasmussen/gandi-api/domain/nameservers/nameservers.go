package zone

import (
	"github.com/prasmussen/gandi-api/client"
	"github.com/prasmussen/gandi-api/operation"
)

type Nameservers struct {
	*client.Client
}

func New(c *client.Client) *Nameservers {
	return &Nameservers{c}
}

// Set the current zone of a domain
func (self *Nameservers) Set(domainName string, nameservers []string) (*operation.OperationInfo, error) {
	var res map[string]interface{}
	params := []interface{}{self.Key, domainName, nameservers}
	if err := self.Call("domain.nameservers.set", params, &res); err != nil {
		return nil, err
	}
	return operation.ToOperationInfo(res), nil
}
