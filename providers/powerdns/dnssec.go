package powerdns

import (
	"context"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/mittwald/go-powerdns/apis/cryptokeys"
)

// getDNSSECCorrections returns corrections that update a domain's DNSSEC state.
func (dsp *powerdnsProvider) getDNSSECCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	zoneCryptokeys, getErr := dsp.client.Cryptokeys().ListCryptokeys(context.Background(), dsp.ServerName, dc.Name)
	if getErr != nil {
		return nil, getErr
	}

	// check if any of the avail. key is active and published
	hasEnabledKey := false
	var keyID int

	if len(zoneCryptokeys) > 0 {
		for _, cryptoKey := range zoneCryptokeys {
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
				F: func() error {
					return dsp.client.Cryptokeys().DeleteCryptokey(context.Background(), dsp.ServerName, dc.Name, keyID)
				},
			},
		}, nil
	}

	// dnssec is disabled, we want it to be enabled
	if !hasEnabledKey && dc.AutoDNSSEC == "on" {
		return []*models.Correction{
			{
				Msg: "Enable DNSSEC",
				F: func() (err error) {
					_, err = dsp.client.Cryptokeys().CreateCryptokey(context.Background(), dsp.ServerName, dc.Name, cryptokeys.Cryptokey{
						KeyType:   "csk",
						Active:    true,
						Published: true,
					})
					return
				},
			},
		}, nil
	}

	return nil, nil
}
