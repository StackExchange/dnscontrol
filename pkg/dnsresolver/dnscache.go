package dnsresolver

import (
	"sync"
)

// dnsCache implements a very simple DNS cache.
// It caches the entire answer (i.e. all TXT records), filtering
// out the non-SPF answers is done at a higher layer.
// At this time the only rtype is "TXT". Eventually we'll need
// to cache A/AAAA/CNAME records to to CNAME flattening.
type dnsCache struct {
	m map[string]map[string][]string // map[fqdn]map[rtype] -> answers
	sync.Mutex
}

func (c *dnsCache) get(label, rtype string) ([]string, bool) {
	c.Lock()
	defer c.Unlock()
	v1, ok := c.m[label]
	if !ok {
		return nil, false
	}
	v2, ok := v1[rtype]
	if !ok {
		return nil, false
	}
	return v2, true
}

func (c *dnsCache) put(label, rtype string, answers []string) {
	c.Lock()
	defer c.Unlock()
	_, ok := c.m[label]
	if !ok {
		c.m[label] = make(map[string][]string)
	}
	c.m[label][rtype] = answers
}
