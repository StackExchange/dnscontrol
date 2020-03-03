package main

import (
	"testing"
)

func TestARecords(t *testing.T) {
	dnsProviderTestCase(t,
		tc("Create an A record", a("@", "1.1.1.1")),
		tc("Change it", a("@", "1.2.3.4")),
		tc("Add another", a("@", "1.2.3.4"), a("www", "1.2.3.4")),
		tc("Add another(same name)", a("@", "1.2.3.4"), a("www", "1.2.3.4"), a("www", "5.6.7.8")),
		tc("Change a ttl", a("@", "1.2.3.4").ttl(1000), a("www", "1.2.3.4"), a("www", "5.6.7.8")),
		tc("Change single target from set", a("@", "1.2.3.4").ttl(1000), a("www", "2.2.2.2"), a("www", "5.6.7.8")),
		tc("Change all ttls", a("@", "1.2.3.4").ttl(500), a("www", "2.2.2.2").ttl(400), a("www", "5.6.7.8").ttl(400)),
		tc("Delete one", a("@", "1.2.3.4").ttl(500), a("www", "5.6.7.8").ttl(400)),
		tc("Add back and change ttl", a("www", "5.6.7.8").ttl(700), a("www", "1.2.3.4").ttl(700)),
		tc("Change targets and ttls", a("www", "1.1.1.1"), a("www", "2.2.2.2")),
		tc("Create wildcard", a("*", "1.2.3.4"), a("www", "1.1.1.1")),
		tc("Delete wildcard", a("www", "1.1.1.1")),
	)
}
