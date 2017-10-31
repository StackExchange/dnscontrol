package normalize

import (
	"fmt"
	"strings"

	"github.com/miekg/dns/dnsutil"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/pkg/spflib"
)

// hasSpfRecords returns true if this record requests SPF unrolling.
func flattenSPFs(cfg *models.DNSConfig) []error {
	var cache spflib.CachingResolver
	var errs []error
	var err error
	for _, domain := range cfg.Domains {
		apexTXTs := domain.Records.Grouped()[models.RecordKey{Type: "TXT", Name: "@"}]
		// flatten all spf records that have the "flatten" metadata
		for _, txt := range apexTXTs {
			var rec *spflib.SPFRecord
			if txt.Metadata["flatten"] != "" || txt.Metadata["split"] != "" {
				if cache == nil {
					cache, err = spflib.NewCache("spfcache.json")
					if err != nil {
						return []error{err}
					}
				}
				rec, err = spflib.Parse(txt.Target, cache)
				if err != nil {
					errs = append(errs, err)
					continue
				}
			}
			if flatten, ok := txt.Metadata["flatten"]; ok && strings.HasPrefix(txt.Target, "v=spf1") {
				rec = rec.Flatten(flatten)
				txt.Target = rec.TXT()
			}
			// now split if needed
			if split, ok := txt.Metadata["split"]; ok {
				if !strings.Contains(split, "%d") {
					errs = append(errs, Warning{fmt.Errorf("Split format `%s` in `%s` is not proper format (should have %%d in it)", split, txt.NameFQDN)})
					continue
				}
				recs := rec.TXTSplit(split + "." + domain.Name)
				for k, v := range recs {
					if k == "@" {
						txt.Target = v
					} else {
						cp := txt.Copy()
						cp.Target = v
						cp.NameFQDN = k
						cp.Name = dnsutil.TrimDomainName(k, domain.Name)
						domain.Records = append(domain.Records, cp)
					}
				}
			}
		}
	}
	if cache == nil {
		return errs
	}
	// check if cache is stale
	for _, e := range cache.ResolveErrors() {
		errs = append(errs, Warning{fmt.Errorf("problem resolving SPF record: %s", e)})
	}
	if len(cache.ResolveErrors()) == 0 {
		changed := cache.ChangedRecords()
		if len(changed) > 0 {
			if err := cache.Save("spfcache.updated.json"); err != nil {
				errs = append(errs, err)
			} else {
				errs = append(errs, Warning{fmt.Errorf("%d spf record lookups are out of date with cache (%s). Wrote changes to spfcache.updated.json. Please rename and commit:\nmv spfcache.updated.json spfcache.json", len(changed), strings.Join(changed, ","))})
			}
		}
	}
	return errs
}
