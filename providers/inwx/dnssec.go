package inwx

import (
	"golang.org/x/net/idna"
)

// zoneIsAutoDNSSECEnabled checks for active autodnssec configuration
func (api *inwxAPI) zoneIsAutoDNSSECEnabled(domain string) (bool, error) {

	resp, err := api.client.Dnssec.Info([]string{domain})
	if err != nil {
		return false, err
	}

	// https://www.inwx.com/en/help/apidoc/f/ch03.html#type.dnssecdomainstatus
	// claims status values can be 'DELETE_ALL', 'MANUAL', 'UPDATE', but
	// testing shows 'AUTO' is what to expect if the domain has automatic
	// DNSSEC enabled.
	return resp.Data[0].DNSSecStatus == "AUTO", nil
}

// enableAutoDNSSEC enables automatic management of DNSSEC
func (api *inwxAPI) enableAutoDNSSEC(domain string) error {
	// if the domain is IDN, it must be in Unicode - ACE encoding is not supported
	// in the INWX dnssec.enablednssec endpoint
	domain, err := idna.ToUnicode(domain)
	if err != nil {
		return err
	}

	err = api.client.Dnssec.Enable(domain)

	return err
}

// disableAutoDNSSEC disables automatic management of DNSSEC
func (api *inwxAPI) disableAutoDNSSEC(domain string) error {

	err := api.client.Dnssec.Disable(domain)

	return err
}
