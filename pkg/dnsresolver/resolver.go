package dnsresolver

import (
	"encoding/json"
	"io/ioutil"
	"net"

	"github.com/pkg/errors"
)

// This file includes all the DNS Resolvers used by package spf.

// DnsResolver looks up txt strings associated with a FQDN.
type DnsResolver interface {
	GetTxt(string) ([]string, error) // Given a DNS label, return the TXT values records.
}

// The "Live DNS" Resolver:

type dnsLive struct {
	filename string
	cache    *dnsCache
}

func NewResolverLive(filename string) *dnsLive {
	// Does live DNS lookups. Records them. Writes file on Close.
	c := &dnsLive{filename: filename}
	c.cache = &dnsCache{m: map[string]map[string][]string{}}
	return c
}

func (c *dnsLive) GetTxt(label string) ([]string, error) {
	// Try the cache.
	txts, ok := c.cache.get(label, "txt")
	if ok {
		return txts, nil
	}

	// Populate the cache:
	t, err := net.LookupTXT(label)
	if err == nil {
		c.cache.put(label, "txt", t)
	}

	return t, err
}

func (c *dnsLive) Close() {
	// Write out and close the file.
	m, _ := json.MarshalIndent(c.cache, "", "  ")
	m = append(m, "\n"...)
	ioutil.WriteFile(c.filename, m, 0666)
}

// The "Pre-Cached DNS" Resolver:

type dnsPreloaded struct {
	cache *dnsCache
}

func NewResolverPreloaded(filename string) (*dnsPreloaded, error) {
	c := &dnsPreloaded{}
	c.cache = &dnsCache{m: map[string]map[string][]string{}}
	j, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(j, &c.cache.m)
	return c, err
}

func (c *dnsPreloaded) DumpCache() *dnsCache {
	return c.cache
}

func (c *dnsPreloaded) GetTxt(label string) ([]string, error) {
	// Try the cache.
	txts, ok := c.cache.get(label, "txt")
	if ok {
		return txts, nil
	}
	return nil, errors.Errorf("No preloaded DNS entry for: %#v", label)
}
