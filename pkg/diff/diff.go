package diff

import (
	"fmt"
	"sort"

	"github.com/gobwas/glob"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/printer"
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
	IncrementalDiff(existing []*models.RecordConfig) (unchanged, create, toDelete, modify Changeset, err error)
	// ChangedGroups performs a diff more appropriate for providers with a "RecordSet" model, where all records with the same name and type are grouped.
	// Individual record changes are often not useful in such scenarios. Instead we return a map of record keys to a list of change descriptions within that group.
	ChangedGroups(existing []*models.RecordConfig) (map[models.RecordKey][]string, error)
}

// New is a constructor for a Differ.
func New(dc *models.DomainConfig, extraValues ...func(*models.RecordConfig) map[string]string) Differ {
	return &differ{
		dc:          dc,
		extraValues: extraValues,

		// compile IGNORE_NAME glob patterns
		compiledIgnoredNames: compileIgnoredNames(dc.IgnoredNames),

		// compile IGNORE_TARGET glob patterns
		compiledIgnoredTargets: compileIgnoredTargets(dc.IgnoredTargets),
	}
}

type differ struct {
	dc          *models.DomainConfig
	extraValues []func(*models.RecordConfig) map[string]string

	compiledIgnoredNames   []glob.Glob
	compiledIgnoredTargets []glob.Glob
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

func (d *differ) IncrementalDiff(existing []*models.RecordConfig) (unchanged, create, toDelete, modify Changeset, err error) {
	unchanged = Changeset{}
	create = Changeset{}
	toDelete = Changeset{}
	modify = Changeset{}
	desired := d.dc.Records

	// sort existing and desired by name

	existingByNameAndType := map[models.RecordKey][]*models.RecordConfig{}
	desiredByNameAndType := map[models.RecordKey][]*models.RecordConfig{}
	for _, e := range existing {
		if d.matchIgnoredName(e.GetLabel()) {
			printer.Debugf("Ignoring record %s %s due to IGNORE_NAME\n", e.GetLabel(), e.Type)
		} else if d.matchIgnoredTarget(e.GetTargetField(), e.Type) {
			printer.Debugf("Ignoring record %s %s due to IGNORE_TARGET\n", e.GetLabel(), e.Type)
		} else {
			k := e.Key()
			existingByNameAndType[k] = append(existingByNameAndType[k], e)
		}
	}
	for _, dr := range desired {
		if d.matchIgnoredName(dr.GetLabel()) {
			return nil, nil, nil, nil, fmt.Errorf("trying to update/add IGNORE_NAMEd record: %s %s", dr.GetLabel(), dr.Type)
		} else if d.matchIgnoredTarget(dr.GetTargetField(), dr.Type) {
			return nil, nil, nil, nil, fmt.Errorf("trying to update/add IGNORE_TARGETd record: %s %s", dr.GetLabel(), dr.Type)
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
				return nil, nil, nil, nil, fmt.Errorf("DUPLICATE E_RECORD FOUND: %s %s", key, normalized)
			}
			existingLookup[normalized] = ex
		}
		for _, de := range desiredRecords {
			normalized := d.content(de)
			if desiredLookup[normalized] != nil {
				return nil, nil, nil, nil, fmt.Errorf("DUPLICATE D_RECORD FOUND: %s %s", key, normalized)
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
		// if found, but not desired, delete it
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

	// Sort the lists. This is purely cosmetic.
	sort.Slice(unchanged, func(i, j int) bool { return ChangesetLess(unchanged, i, j) })
	sort.Slice(create, func(i, j int) bool { return ChangesetLess(create, i, j) })
	sort.Slice(toDelete, func(i, j int) bool { return ChangesetLess(toDelete, i, j) })

	return
}

// ChangesetLess returns true if c[i] < c[j].
func ChangesetLess(c Changeset, i, j int) bool {
	var a, b string
	// Which fields are we comparing?
	// Usually only Desired OR Existing content exists (we're either
	// adding or deleting records).  In those cases, just use whichever
	// isn't nil.
	// In the case where both Desired AND Existing exist, it doesn't
	// matter which we use as long as we are consistent.  I flipped a
	// coin and picked to use Desired in that case.

	if c[i].Desired != nil {
		a = c[i].Desired.NameFQDN
	} else {
		a = c[i].Existing.NameFQDN
	}

	if c[j].Desired != nil {
		b = c[j].Desired.NameFQDN
	} else {
		b = c[j].Existing.NameFQDN
	}

	return a < b

	// TODO(tlim): This won't correctly sort:
	// []string{"example.com", "foo.example.com", "bar.example.com"}
	// A simple way to do that correctly is to split on ".", reverse the
	// elements, and sort on the result.
}

// CorrectionLess returns true when comparing corrections.
func CorrectionLess(c []*models.Correction, i, j int) bool {
	return c[i].Msg < c[j].Msg
}

func (d *differ) ChangedGroups(existing []*models.RecordConfig) (map[models.RecordKey][]string, error) {
	changedKeys := map[models.RecordKey][]string{}
	_, create, delete, modify, err := d.IncrementalDiff(existing)
	if err != nil {
		return nil, err
	}
	for _, c := range create {
		changedKeys[c.Desired.Key()] = append(changedKeys[c.Desired.Key()], c.String())
	}
	for _, d := range delete {
		changedKeys[d.Existing.Key()] = append(changedKeys[d.Existing.Key()], d.String())
	}
	for _, m := range modify {
		changedKeys[m.Desired.Key()] = append(changedKeys[m.Desired.Key()], m.String())
	}
	return changedKeys, nil
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

func compileIgnoredNames(ignoredNames []string) []glob.Glob {
	result := make([]glob.Glob, 0, len(ignoredNames))

	for _, tst := range ignoredNames {
		g, err := glob.Compile(tst, '.')
		if err != nil {
			panic(fmt.Sprintf("Failed to compile IGNORE_NAME pattern %q: %v", tst, err))
		}

		result = append(result, g)
	}

	return result
}

func compileIgnoredTargets(ignoredTargets []*models.IgnoreTarget) []glob.Glob {
	result := make([]glob.Glob, 0, len(ignoredTargets))

	for _, tst := range ignoredTargets {
		if tst.Type != "CNAME" {
			panic(fmt.Sprintf("Invalid rType for IGNORE_TARGET %v", tst.Type))
		}

		g, err := glob.Compile(tst.Pattern, '.')
		if err != nil {
			panic(fmt.Sprintf("Failed to compile IGNORE_TARGET pattern %q: %v", tst, err))
		}

		result = append(result, g)
	}

	return result
}

func (d *differ) matchIgnoredName(name string) bool {
	for _, tst := range d.compiledIgnoredNames {
		if tst.Match(name) {
			return true
		}
	}
	return false
}

func (d *differ) matchIgnoredTarget(target string, rType string) bool {
	if rType != "CNAME" {
		return false
	}

	for _, tst := range d.compiledIgnoredTargets {
		if tst.Match(target) {
			return true
		}
	}

	return false
}
