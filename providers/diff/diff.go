package diff

import (
	"fmt"
	"sort"
)

type Record interface {
	GetName() string
	GetType() string
	GetContent() string

	// Get relevant comparision data. Default implentation uses "ttl [mx priority]", but providers may insert
	// provider specific metadata if needed.
	GetComparisionData() string
}

type Correlation struct {
	Existing Record
	Desired  Record
}
type Changeset []Correlation

func IncrementalDiff(existing []Record, desired []Record) (unchanged, create, toDelete, modify Changeset) {
	unchanged = Changeset{}
	create = Changeset{}
	toDelete = Changeset{}
	modify = Changeset{}

	//	log.Printf("ID existing records: (%d)\n", len(existing))
	//	for i, d := range existing {
	//		log.Printf("\t%d\t%v\n", i, d)
	//	}
	//	log.Printf("ID desired records: (%d)\n", len(desired))
	//	for i, d := range desired {
	//		log.Printf("\t%d\t%v\n", i, d)
	//	}

	//sort existing and desired by name
	type key struct {
		name, rType string
	}
	existingByNameAndType := map[key][]Record{}
	desiredByNameAndType := map[key][]Record{}
	for _, e := range existing {
		k := key{e.GetName(), e.GetType()}
		existingByNameAndType[k] = append(existingByNameAndType[k], e)
	}
	for _, d := range desired {
		k := key{d.GetName(), d.GetType()}
		desiredByNameAndType[k] = append(desiredByNameAndType[k], d)
	}

	// Look through existing records. This will give us changes and deletions and some additions
	for key, existingRecords := range existingByNameAndType {
		desiredRecords := desiredByNameAndType[key]

		//first look through records that are the same content on both sides. Those are either modifications or unchanged

		for i := len(existingRecords) - 1; i >= 0; i-- {
			ex := existingRecords[i]
			for j, de := range desiredRecords {
				if de.GetContent() == ex.GetContent() {
					//they're either identical or should be a modification of each other
					if de.GetComparisionData() == ex.GetComparisionData() {
						unchanged = append(unchanged, Correlation{ex, de})
					} else {
						modify = append(modify, Correlation{ex, de})
					}
					// remove from both slices by index
					existingRecords = existingRecords[:i+copy(existingRecords[i:], existingRecords[i+1:])]
					desiredRecords = desiredRecords[:j+copy(desiredRecords[j:], desiredRecords[j+1:])]
					break
				}
			}
		}

		desiredLookup := map[string]Record{}
		existingLookup := map[string]Record{}
		// build index based on normalized value/ttl
		for _, ex := range existingRecords {
			normalized := fmt.Sprintf("%s %s", ex.GetContent(), ex.GetComparisionData())
			if existingLookup[normalized] != nil {
				panic(fmt.Sprintf("DUPLICATE E_RECORD FOUND: %s %s", key, normalized))
			}
			existingLookup[normalized] = ex
		}
		for _, de := range desiredRecords {
			normalized := fmt.Sprintf("%s %s", de.GetContent(), de.GetComparisionData())
			if desiredLookup[normalized] != nil {
				panic(fmt.Sprintf("DUPLICATE D_RECORD FOUND: %s %s", key, normalized))
			}
			desiredLookup[normalized] = de
		}
		// if a record is in both, it is unchanged
		for norm, ex := range existingLookup {
			if de, ok := desiredLookup[norm]; ok {
				unchanged = append(unchanged, Correlation{ex, de})
				delete(existingLookup, norm)
				delete(desiredLookup, norm)
			}
		}
		//sort records by normalized text. Keeps behaviour deterministic
		existingStrings, desiredStrings := []string{}, []string{}
		for norm := range existingLookup {
			existingStrings = append(existingStrings, norm)
		}
		for norm := range desiredLookup {
			desiredStrings = append(desiredStrings, norm)
		}
		sort.Strings(existingStrings)
		sort.Strings(desiredStrings)
		// Modifications. Take 1 from each side.
		for len(desiredStrings) > 0 && len(existingStrings) > 0 {
			modify = append(modify, Correlation{existingLookup[existingStrings[0]], desiredLookup[desiredStrings[0]]})
			existingStrings = existingStrings[1:]
			desiredStrings = desiredStrings[1:]
		}
		// If desired still has things they are additions
		for _, norm := range desiredStrings {
			rec := desiredLookup[norm]
			create = append(create, Correlation{nil, rec})
		}
		// if found , but not desired, delete it
		for _, norm := range existingStrings {
			rec := existingLookup[norm]
			toDelete = append(toDelete, Correlation{rec, nil})
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
			create = append(create, Correlation{nil, rec})
		}
	}
	return
}

func (c Correlation) String() string {
	if c.Existing == nil {
		return fmt.Sprintf("CREATE %s %s %s %s", c.Desired.GetType(), c.Desired.GetName(), c.Desired.GetContent(), c.Desired.GetComparisionData())
	}
	if c.Desired == nil {
		return fmt.Sprintf("DELETE %s %s %s %s", c.Existing.GetType(), c.Existing.GetName(), c.Existing.GetContent(), c.Existing.GetComparisionData())
	}
	return fmt.Sprintf("MODIFY %s %s: (%s %s) -> (%s %s)", c.Existing.GetType(), c.Existing.GetName(), c.Existing.GetContent(), c.Existing.GetComparisionData(), c.Desired.GetContent(), c.Desired.GetComparisionData())
}
