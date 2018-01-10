package spflib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
)

// Resolver looks up spf txt records associated with a FQDN.
type Resolver interface {
	GetSPF(string) (string, error)
}

// LiveResolver simply queries DNS to resolve SPF records.
type LiveResolver struct{}

// GetSPF looks up the SPF record named "name".
func (l LiveResolver) GetSPF(name string) (string, error) {
	vals, err := net.LookupTXT(name)
	if err != nil {
		return "", err
	}
	spf := ""
	for _, v := range vals {
		if strings.HasPrefix(v, "v=spf1") {
			if spf != "" {
				return "", fmt.Errorf("%s has multiple SPF records", name)
			}
			spf = v
		}
	}
	if spf == "" {
		return "", fmt.Errorf("%s has no SPF record", name)
	}
	return spf, nil
}

// CachingResolver wraps a live resolver and adds caching to it.
// GetSPF will always return the cached value, if present.
// It will also query the inner resolver and compare results.
// If a given lookup has inconsistencies between cache and live,
// GetSPF will return the cached result.
// All records queries will be stored for the lifetime of the resolver,
// and can be flushed to disk at the end.
// All resolution errors from the inner resolver will be saved and can be retreived later.
type CachingResolver interface {
	Resolver
	ChangedRecords() []string
	ResolveErrors() []error
	Save(filename string) error
}

type cacheEntry struct {
	SPF string

	// value we have looked up this run
	resolvedSPF  string
	resolveError error
}

type cache struct {
	records map[string]*cacheEntry

	inner Resolver
}

// NewCache creates a new cache file named filename.
func NewCache(filename string) (CachingResolver, error) {
	f, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// doesn't exist, just make a new one
			return &cache{
				records: map[string]*cacheEntry{},
				inner:   LiveResolver{},
			}, nil
		}
		return nil, err
	}
	dec := json.NewDecoder(f)
	recs := map[string]*cacheEntry{}
	if err := dec.Decode(&recs); err != nil {
		return nil, err
	}
	return &cache{
		records: recs,
		inner:   LiveResolver{},
	}, nil
}

func (c *cache) GetSPF(name string) (string, error) {
	entry, ok := c.records[name]
	if !ok {
		entry = &cacheEntry{}
		c.records[name] = entry
	}
	if entry.resolvedSPF == "" && entry.resolveError == nil {
		entry.resolvedSPF, entry.resolveError = c.inner.GetSPF(name)
	}
	// return cached value
	if entry.SPF != "" {
		return entry.SPF, nil
	}
	// if not cached, return results of inner resolver
	return entry.resolvedSPF, entry.resolveError
}

func (c *cache) ChangedRecords() []string {
	names := []string{}
	for name, entry := range c.records {
		if entry.resolvedSPF != entry.SPF {
			names = append(names, name)
		}
	}
	return names
}

func (c *cache) ResolveErrors() (errs []error) {
	for _, entry := range c.records {
		if entry.resolveError != nil {
			errs = append(errs, entry.resolveError)
		}
	}
	return
}
func (c *cache) Save(filename string) error {
	outRecs := make(map[string]*cacheEntry, len(c.records))
	for k, entry := range c.records {
		// move resolved data into cached field
		// only take those we actually resolved
		if entry.resolvedSPF != "" {
			entry.SPF = entry.resolvedSPF
			outRecs[k] = entry
		}
	}
	dat, _ := json.MarshalIndent(outRecs, "", "  ")
	return ioutil.WriteFile(filename, dat, 0644)
}
