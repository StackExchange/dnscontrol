package powerdns

import (
	"context"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/mittwald/go-powerdns/apis/cryptokeys"
)

// getDNSSECCorrections returns corrections that update a domain's DNSSEC state.
func (api *powerdnsProvider) getDNSSECCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	cryptokeys, getErr := api.client.Cryptokeys().ListCryptokeys(context.Background(), api.ServerName, dc.Name)
	if getErr != nil {
		return nil, getErr
	}

	// check if any of the avail. key is active and published
	hasEnabledKey := false
	var keyID int

	if len(cryptokeys) > 0 {
		for _, cryptoKey := range cryptokeys {
			if cryptoKey.Active && cryptoKey.Published {
				hasEnabledKey = true
				keyID = cryptoKey.ID
				break
			}
		}
	}

	// dnssec is enabled, we want it to be disabled
	if hasEnabledKey && dc.AutoDNSSEC == "off" {
		return []*models.Correction{
			{
				Msg: "Disable DNSSEC",
				F:   func() error { _, err := api.removeDnssec(dc.Name, keyID); return err },
			},
		}, nil
	}

	// dnssec is disabled, we want it to be enabled
	if !hasEnabledKey && dc.AutoDNSSEC == "on" {
		return []*models.Correction{
			{
				Msg: "Enable DNSSEC",
				F:   func() error { _, err := api.enableDnssec(dc.Name); return err },
			},
		}, nil
	}

	return nil, nil
}

// enableDnssec creates a active and published cryptokey on this domain
func (api *powerdnsProvider) enableDnssec(domain string) (bool, error) {
	// if there is now key, create one and enable it
	_, err := api.client.Cryptokeys().CreateCryptokey(context.Background(), api.ServerName, domain, cryptokeys.Cryptokey{
		KeyType:   "csk",
		Active:    true,
		Published: true,
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

// removeDnssec removes the cryptokey from this zone
func (api *powerdnsProvider) removeDnssec(domain string, keyID int) (bool, error) {
	err := api.client.Cryptokeys().DeleteCryptokey(context.Background(), api.ServerName, domain, keyID)
	if err != nil {
		return false, err
	}
	return true, nil
}
