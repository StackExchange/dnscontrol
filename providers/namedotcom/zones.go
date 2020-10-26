package namedotcom

import (
	"github.com/namedotcom/go/namecom"
)

// ListZones returns all the zones in an account
func (c *namedotcomProvider) ListZones() ([]string, error) {
	var names []string
	var page int32

	for {
		n, err := c.client.ListDomains(&namecom.ListDomainsRequest{Page: page})
		if err != nil {
			return nil, err
		}
		page = n.NextPage

		for _, j := range n.Domains {
			names = append(names, j.DomainName)
		}

		if page == 0 {
			break
		}
	}

	return names, nil
}
