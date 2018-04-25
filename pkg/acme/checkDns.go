package acme

import (
	"log"
	"time"

	"github.com/xenolf/lego/acmev2"
)

func init() {
	// default record verification in the client library makes sure the authoritative nameservers
	// have the expected records.
	// Sometimes the Let's Encrypt verification fails anyway because records have not propagated the provider's network fully.
	// So we add an additional 20 second sleep just for safety.
	origCheck := acme.PreCheckDNS
	acme.PreCheckDNS = func(fqdn, value string) (bool, error) {
		start := time.Now()
		v, err := origCheck(fqdn, value)
		if err != nil {
			return v, err
		}
		log.Printf("DNS ok after %s. Waiting again for propagation", time.Now().Sub(start))
		time.Sleep(20 * time.Second)
		return v, err
	}
}

// Timeout increases the client-side polling check time to five minutes with one second waits in-between.
func (c *certManager) Timeout() (timeout, interval time.Duration) {
	return 5 * time.Minute, time.Second
}
