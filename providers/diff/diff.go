package diff

import (
	"fmt"
	"sort"

	"github.com/gobwas/glob"

	"github.com/StackExchange/dnscontrol/v2/models"
	"github.com/StackExchange/dnscontrol/v2/pkg/printer"
)

// Correlation stores a difference between two domains.
type Correlation struct {
	d        *differ
	Existing *models.RecordConfig
	Desired  *models.RecordConfig
}

// Changeset stores many Correlation.
type Changeset []Correlation

// Differ is an interface for computing the difference between two zones.
type Differ interface {
	// IncrementalDiff performs a diff on a record-by-record basis, and returns a sets for which records need to be created, deleted, or modified.
	IncrementalDiff(existing []*models.RecordConfig) (unchanged, create, toDelete, modify Changeset)
	// ChangedGroups performs a diff more appropriate for providers with a "RecordSet" model, where all records with the same name and type are grouped.
	// Individual record changes are often not useful in such scenarios. Instead we return a map of record keys to a list of change descriptions within that group.
	ChangedGroups(existing []*models.RecordConfig) map[models.RecordKey][]string
}

// New is a constructor for a Differ.
func New(dc *models.DomainConfig, extraValues ...func(*models.RecordConfig) map[string]string) Differ {
	return &differ{
		dc:          dc,
		extraValues: extraValues,

		// compile IGNORE glob patterns
		compiledIgnoredLabels: compileIgnoredLabels(dc.IgnoredLabels),
	}
}

type differ struct {
	dc          *models.DomainConfig
	extraValues []func(*models.RecordConfig) map[string]string

	compiledIgnoredLabels []glob.Glob
}

// get normalized content for record. target, ttl, mxprio, and specified metadata
func (d *differ) content(r *models.RecordConfig) string {
	// NB(tlim): This function will eventually be replaced by calling
	// r.GetTargetDiffable().  In the meanwhile, this function compares
	// its output with r.GetTargetDiffable() to make sure the same
	// results are generated.  Once we have confidence, this function will go away.
	content := fmt.Sprintf("%v ttl=%d", r.GetTargetCombined(), r.TTL)
	if r.Type == "SOA" {
		content = fmt.Sprintf("%s %v %d %d %d %d ttl=%d", r.Target, r.SoaMbox, r.SoaRefresh, r.SoaRetry, r.SoaExpire, r.SoaMinttl, r.TTL) // SoaSerial is not used in comparison
	}
	var allMaps []map[string]string
	for _, f := range d.extraValues {
		// sort the extra values map keys to perform a deterministic
		// comparison since Golang maps iteration order is not guaranteed
		valueMap := f(r)
		allMaps = append(allMaps, valueMap)
		keys := make([]string, 0)
		for k := range valueMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := valueMap[k]
			content += fmt.Sprintf(" %s=%s", k, v)
		}
	}
	control := r.ToDiffable(allMaps...)
	if control != content {
		fmt.Printf("CONTROL=%q CONTENT=%q\n", control, content)
		panic("OOPS! control != content")
	}
	return content
}

func (d *differ) IncrementalDiff(existing []*models.RecordConfig) (unchanged, create, toDelete, modify Changeset) {
	unchanged = Changeset{}
	create = Changeset{}
	toDelete = Changeset{}
	modify = Changeset{}
	desired := d.dc.Records

	// sort existing and desired by name

	existingByNameAndType := map[models.RecordKey][]*models.RecordConfig{}
	desiredByNameAndType := map[models.RecordKey][]*models.RecordConfig{}
	for _, e := range existing {
		if d.matchIgnored(e.GetLabel()) {
			printer.Debugf("Ignoring record %s %s due to IGNORE\n", e.GetLabel(), e.Type)
		} else {
			k := e.Key()
			existingByNameAndType[k] = append(existingByNameAndType[k], e)
		}
	}
	for _, dr := range desired {
		if d.matchIgnored(dr.GetLabel()) {
			panic(fmt.Sprintf("Trying to update/add IGNOREd record: %s %s", dr.GetLabel(), dr.Type))
		} else {
			k := dr.Key()
			desiredByNameAndType[k] = append(desiredByNameAndType[k], dr)
		}
	}
	// if NO_PURGE is set, just remove anything that is only in existing.
	if d.dc.KeepUnknown {
		for k := range existingByNameAndType {
			if _, ok := desiredByNameAndType[k]; !ok {
				printer.Debugf("Ignoring record set %s %s due to NO_PURGE\n", k.Type, k.NameFQDN)
				delete(existingByNameAndType, k)
			}
		}
	}
	// Look through existing records. This will give us changes and deletions and some additions.
	// Each iteration is only for a single type/name record set
	for key, existingRecords := range existingByNameAndType {
		desiredRecords := desiredByNameAndType[key]

		// Very first, get rid of any identical records. Easy.
		for i := len(existingRecords) - 1; i >= 0; i-- {
			ex := existingRecords[i]
			for j, de := range desiredRecords {
				if d.content(de) != d.content(ex) {
					continue
				}
				unchanged = append(unchanged, Correlation{d, ex, de})
				existingRecords = existingRecords[:i+copy(existingRecords[i:], existingRecords[i+1:])]
				desiredRecords = desiredRecords[:j+copy(desiredRecords[j:], desiredRecords[j+1:])]
				break
			}
		}

		// Next, match by target. This will give the most natural modifications.
		for i := len(existingRecords) - 1; i >= 0; i-- {
			ex := existingRecords[i]
			for j, de := range desiredRecords {
				if de.GetTargetField() == ex.GetTargetField() {
					// two records share a target, but different content (ttl or metadata changes)
					modify = append(modify, Correlation{d, ex, de})
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
		// sort records by normalized text. Keeps behaviour deterministic
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

	// any name/type sets not already processed are pure additions
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

func (d *differ) ChangedGroups(existing []*models.RecordConfig) map[models.RecordKey][]string {
	changedKeys := map[models.RecordKey][]string{}
	_, create, delete, modify := d.IncrementalDiff(existing)
	for _, c := range create {
		changedKeys[c.Desired.Key()] = append(changedKeys[c.Desired.Key()], c.String())
	}
	for _, d := range delete {
		changedKeys[d.Existing.Key()] = append(changedKeys[d.Existing.Key()], d.String())
	}
	for _, m := range modify {
		changedKeys[m.Desired.Key()] = append(changedKeys[m.Desired.Key()], m.String())
	}
	return changedKeys
}

// DebugKeyMapMap debug prints the results from ChangedGroups.
func DebugKeyMapMap(note string, m map[models.RecordKey][]string) {
	// The output isn't pretty but it is useful.
	fmt.Println("DEBUG:", note)

	// Extract the keys
	var keys []models.RecordKey
	for k := range m {
		keys = append(keys, k)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		if keys[i].NameFQDN == keys[j].NameFQDN {
			return keys[i].Type < keys[j].Type
		}
		return keys[i].NameFQDN < keys[j].NameFQDN
	})

	// Pretty print the map:
	for _, k := range keys {
		fmt.Printf("   %v %v:\n", k.Type, k.NameFQDN)
		for _, s := range m[k] {
			fmt.Printf("      -- %q\n", s)
		}
	}
}

func (c Correlation) String() string {
	if c.Existing == nil {
		return fmt.Sprintf("CREATE %s %s %s", c.Desired.Type, c.Desired.GetLabelFQDN(), c.d.content(c.Desired))
	}
	if c.Desired == nil {
		return fmt.Sprintf("DELETE %s %s %s", c.Existing.Type, c.Existing.GetLabelFQDN(), c.d.content(c.Existing))
	}
	return fmt.Sprintf("MODIFY %s %s: (%s) -> (%s)", c.Existing.Type, c.Existing.GetLabelFQDN(), c.d.content(c.Existing), c.d.content(c.Desired))
}

func sortedKeys(m map[string]*models.RecordConfig) []string {
	s := []string{}
	for v := range m {
		s = append(s, v)
	}
	sort.Strings(s)
	return s
}

func compileIgnoredLabels(ignoredLabels []string) []glob.Glob {
	result := make([]glob.Glob, 0, len(ignoredLabels))

	for _, tst := range ignoredLabels {
		g, err := glob.Compile(tst, '.')
		if err != nil {
			panic(fmt.Sprintf("Failed to compile IGNORE pattern %q: %v", tst, err))
		}

		result = append(result, g)
	}

	return result
}

func (d *differ) matchIgnored(name string) bool {
	for _, tst := range d.compiledIgnoredLabels {
		if tst.Match(name) {
			return true
		}
	}
	return false
}
