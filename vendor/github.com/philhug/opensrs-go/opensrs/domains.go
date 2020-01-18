package opensrs

import (
	"strconv"
)

// DomainsService handles communication with the domain related
// methods of the OpenSRS API.
//
type DomainsService struct {
	client *Client
}

// GetDomain fetches a domain.
//
func (s *DomainsService) GetDomain(domainIdentifier string, domainType string, limit int) (*OpsResponse, error) {
	opsResponse := OpsResponse{}
	opsRequestAttributes := OpsRequestAttributes{Domain: domainIdentifier, Limit: strconv.Itoa(limit), Type: domainType}

	resp, err := s.client.post("GET", "DOMAIN", opsRequestAttributes, &opsResponse)
	if err != nil {
		return nil, err
	}
	_ = resp
	return &opsResponse, nil
}

// UpdateDomainNameservers changes domain servers on a domain.

//
func (s *DomainsService) UpdateDomainNameservers(domainIdentifier string, newDs []string) (*OpsResponse, error) {
	opsResponse := OpsResponse{}

	opsRequestAttributes := OpsRequestAttributes{Domain: domainIdentifier, AssignNs: newDs, OpType: "assign"}

	resp, err := s.client.post("ADVANCED_UPDATE_NAMESERVERS", "DOMAIN", opsRequestAttributes, &opsResponse)
	if err != nil {
		return nil, err
	}
	_ = resp
	return &opsResponse, nil
}
