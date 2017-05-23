package spf

import (
	"fmt"
	"net"
)

// This file includes all the DNS Resolvers used by package spf.

// DnsResolver looks up txt strings associated with a FQDN.
type DnsResolver interface {
	GetTxt(string) []string // Given a DNS label, return the TXT values records.
}

// The "Live DNS" Resolver:

type dnsLive struct {
	filename string
	cache    dnsCache
}

func NewResolverLive(filename string) *dnsLive {
	// Does live DNS lookups. Records them. Writes file on Close.
	c := &dnsLive{filename: filename}
	return c
}

func (c *dnsLive) GetTxt(label string) ([]string, error) {
	// Try the cache.
	txts, ok := c.dnsGet(label, "txt")
	if ok {
		return txts, nil
	}

	// Populate the cache:
	t, err := net.LookupTXT(label)
	if err == nil {
		c.cache.dnsPut(label, "txt", t)
	}

	return t, err
}

func (c *dnsLive) Close() {
	// Write out and close the file.
	fmt.Printf("UNIMPLEMENTED: Create file %#v with %#v\n", c.filename, c.cache)
}

// The "Pre-Cached DNS" Resolver:

type dnsPreloaded struct {
	filename string
	cache    dnsCache
}

func NewResolverPreloaded(filename string) *dnsPreloaded {
	c := &dnsPreloaded{filename: filename}
	return c
}

func (c *dnsCache) GetTxt(label string) ([]string, error) {
	// If in c.cache, return it.
	// Otherwise: return error.
	return nil, nil
}

// Notes

// var m = dnsCache{
// 	"_spf.google.com": {
// 		"txt": "foo",
// 	},
// 	"mail.zendesk.com.google.com": {
// 		"txt": "bar",
// 	},
// }
