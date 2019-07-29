package acme

import (
	"log"
	"time"

	"github.com/go-acme/lego/challenge/dns01"
)

func (c *certManager) preCheckDNS(domain, fqdn, value string, native dns01.PreCheckFunc) (bool, error) {
	// default record verification in the client library makes sure the authoritative nameservers
	// have the expected records.
	// Sometimes the Let's Encrypt verification fails anyway because records have not propagated the provider's network fully.
	// So we add an additional 60 second sleep just for safety.
	v, err := native(fqdn, value)
	if err != nil {
		return v, err
	}
	if !c.waitedOnce {
		log.Printf("DNS ok. Waiting another 60s to ensure stability.")
		time.Sleep(60 * time.Second)
		c.waitedOnce = true
	}
	log.Printf("DNS records seem to exist. Proceeding to request validation")
	return v, err
}

// Timeout increases the client-side polling check time to five minutes with one second waits in-between.
func (c *certManager) Timeout() (timeout, interval time.Duration) {
	return 5 * time.Minute, time.Second
}
