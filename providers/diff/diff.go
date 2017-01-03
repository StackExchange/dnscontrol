package diff

import (
	"fmt"
	"sort"

	"github.com/StackExchange/dnscontrol/models"
)

type Correlation struct {
	d        *differ
	Existing *models.RecordConfig
	Desired  *models.RecordConfig
}
type Changeset []Correlation

type Differ interface {
	IncrementalDiff(existing []*models.RecordConfig) (unchanged, create, toDelete, modify Changeset)
}

func New(dc *models.DomainConfig, metadataKeys ...string) Differ {
	return &differ{
		dc:           dc,
		metadataKeys: metadataKeys,
	}
}

type differ struct {
	dc           *models.DomainConfig
	metadataKeys []string
}

// get normalized content for record. target, ttl, mxprio, and specified metadata
func (d *differ) content(r *models.RecordConfig) string {
	content := fmt.Sprintf("%s %d", r.Target, r.TTL)
	if r.Type == "MX" {
		content += fmt.Sprintf(" priority=%d", r.Priority)
	}

	for _, key := range d.metadataKeys {
		val := ""
		if r.Metadata != nil {
			val = r.Metadata[key]
		}
		content += fmt.Sprintf(" %s=%s", key, val)
	}
	return content
}

func (d *differ) IncrementalDiff(existing []*models.RecordConfig) (unchanged, create, toDelete, modify Changeset) {
	unchanged = Changeset{}
	create = Changeset{}
	toDelete = Changeset{}
	modify = Changeset{}
	desired := d.dc.Records

	//sort existing and desired by name
	type key struct {
		name, rType string
	}
	existingByNameAndType := map[key][]*models.RecordConfig{}
	desiredByNameAndType := map[key][]*models.RecordConfig{}
	for _, e := range existing {
		k := key{e.NameFQDN, e.Type}
		existingByNameAndType[k] = append(existingByNameAndType[k], e)
	}
	for _, d := range desired {
		k := key{d.NameFQDN, d.Type}
		desiredByNameAndType[k] = append(desiredByNameAndType[k], d)
	}
	// Look through existing records. This will give us changes and deletions and some additions.
	// Each iteration is only for a single type/name record set
	for key, existingRecords := range existingByNameAndType {
		desiredRecords := desiredByNameAndType[key]
		//first look through records that are the same target on both sides. Those are either modifications or unchanged
		for i := len(existingRecords) - 1; i >= 0; i-- {
			ex := existingRecords[i]
			for j, de := range desiredRecords {
				if de.Target == ex.Target {
					//they're either identical or should be a modification of each other (ttl or metadata changes)
					if d.content(de) == d.content(ex) {
						unchanged = append(unchanged, Correlation{d, ex, de})
					} else {
						modify = append(modify, Correlation{d, ex, de})
					}
					// remove from both slices by index
					existingRecords = existingRecords[:i+copy(existingRecords[i:], existingRecords[i+1:])]
					desiredRecords = desiredRecords[:j+copy(desiredRecords[j:], desiredRecords[j+1:])]
					break
				}
			}
		}

		desiredLookup := map[string]*models.RecordConfig{}
		existingLookup := map[string]*models.RecordConfig{}
		// build index based on normalized content data
		for _, ex := range existingRecords {
			normalized := d.content(ex)
			if existingLookup[normalized] != nil {
				panic(fmt.Sprintf("DUPLICATE E_RECORD FOUND: %s %s", key, normalized))
			}
			existingLookup[normalized] = ex
		}
		for _, de := range desiredRecords {
			normalized := d.content(de)
			if desiredLookup[normalized] != nil {
				panic(fmt.Sprintf("DUPLICATE D_RECORD FOUND: %s %s", key, normalized))
			}
			desiredLookup[normalized] = de
		}
		// if a record is in both, it is unchanged
		for norm, ex := range existingLookup {
			if de, ok := desiredLookup[norm]; ok {
				unchanged = append(unchanged, Correlation{d, ex, de})
				delete(existingLookup, norm)
				delete(desiredLookup, norm)
			}
		}
		//sort records by normalized text. Keeps behaviour deterministic
		existingStrings, desiredStrings := sortedKeys(existingLookup), sortedKeys(desiredLookup)
		// Modifications. Take 1 from each side.
		for len(desiredStrings) > 0 && len(existingStrings) > 0 {
			modify = append(modify, Correlation{d, existingLookup[existingStrings[0]], desiredLookup[desiredStrings[0]]})
			existingStrings = existingStrings[1:]
			desiredStrings = desiredStrings[1:]
		}
		// If desired still has things they are additions
		for _, norm := range desiredStrings {
			rec := desiredLookup[norm]
			create = append(create, Correlation{d, nil, rec})
		}
		// if found , but not desired, delete it
		for _, norm := range existingStrings {
			rec := existingLookup[norm]
			toDelete = append(toDelete, Correlation{d, rec, nil})
		}
		// remove this set from the desired list to indicate we have processed it.
		delete(desiredByNameAndType, key)
	}

	//any name/type sets not already processed are pure additions
	for name := range existingByNameAndType {
		delete(desiredByNameAndType, name)
	}
	for _, desiredList := range desiredByNameAndType {
		for _, rec := range desiredList {
			create = append(create, Correlation{d, nil, rec})
		}
	}
	return
}

func (c Correlation) String() string {
	if c.Existing == nil {
		return fmt.Sprintf("CREATE %s %s %s", c.Desired.Type, c.Desired.NameFQDN, c.d.content(c.Desired))
	}
	if c.Desired == nil {
		return fmt.Sprintf("DELETE %s %s %s", c.Existing.Type, c.Existing.NameFQDN, c.d.content(c.Existing))
	}
	return fmt.Sprintf("MODIFY %s %s: (%s) -> (%s)", c.Existing.Type, c.Existing.NameFQDN, c.d.content(c.Existing), c.d.content(c.Desired))
}

func sortedKeys(m map[string]*models.RecordConfig) []string {
	s := []string{}
	for v := range m {
		s = append(s, v)
	}
	sort.Strings(s)
	return s
}
