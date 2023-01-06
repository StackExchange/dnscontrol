package diff2

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/prettyzone"
)

/*

Diffing the existing and desired DNS records is difficult. There are
many edge cases to consider.  The old pkg/diff system was successful
at handling all these edge cases but it made the code very complex.

pkg/diff2 was inspired by the intuition that the edge cases would
disappear if we simply stored the data in a way that was easier to
compare. The edge cases would disappear and the code would become more
simple.  Simple is better.

The struct CompareConfig is a data structure that stores all the
existing/desired RecordConfig items in a way that makes the
differencing engine easier to implement.

However, complexity never disappears it just moves elsewhere in the
system.  Converting our RecordConfigs to this datastructure is
complex. However by decoupling the conversion and the differencing, we
get two systems that can be tested independently. Thus, we have more
confidence in this system.

CompareConfig stores pointers to the original models.RecordConfig
grouped by label, and within each label grouped by rType (A, CNAME,
etc.).  In that final grouping the records are stored two ways. First,
as a list.  Second, as a list of (string,RecordConfig) tuples, where
the string is an opaque blob that can be used to compare for equality.
These lists are stored in an order that makes generating lists of
changes naturally be in the correct order for updates.

The structure also stores or pre-computes data that is needed by the
differencing engine, such as maps of which labels and RecordKeys
exist.
*/

type ComparableFunc func(*models.RecordConfig) string

type CompareConfig struct {
	// The primary data. Each record stored once, grouped by label then
	// by rType:
	existing, desired models.Records // The original Recs.
	ldata             []*labelConfig // The Recs, grouped by label.
	//
	// Pre-computed values stored for easy access.
	origin   string                    // Domain zone
	labelMap map[string]bool           // Which labels exist?
	keyMap   map[models.RecordKey]bool // Which RecordKey exists?
	//
	// A function that generates a string used to compare two
	// RecordConfigs for equality.  This is normally nil. If it is not
	// nil, the function is called and the resulting string is joined to
	// the existing compareable string. This enables (for example)
	// custom rtypes to have their own comparison text added to the
	// comparison string.
	compareableFunc ComparableFunc
	//
}

type labelConfig struct {
	label string         // The label
	tdata []*rTypeConfig // The records for that label, grouped by rType.
}

type rTypeConfig struct {
	rType string // The rType for all records in this group (A, CNAME, etc)
	// The records stored as lists:
	existingRecs []*models.RecordConfig
	desiredRecs  []*models.RecordConfig
	// The records stored as compareable/rec tuples:
	existingTargets []targetConfig
	desiredTargets  []targetConfig
}

type targetConfig struct {
	compareable string               // A string that can be used to compare two rec's for equality.
	rec         *models.RecordConfig // The RecordConfig itself.
}

func NewCompareConfig(origin string, existing, desired models.Records, compFn ComparableFunc) *CompareConfig {
	cc := &CompareConfig{
		existing: existing,
		desired:  desired,
		//
		origin:          origin,
		compareableFunc: compFn,
		//
		labelMap: map[string]bool{},
		keyMap:   map[models.RecordKey]bool{},
	}
	cc.addRecords(existing, true) // Must be called first so that CNAME manipulations happen in the correct order.
	cc.addRecords(desired, false)
	cc.VerifyCNAMEAssertions()
	sort.Slice(cc.ldata, func(i, j int) bool {
		return prettyzone.LabelLess(cc.ldata[i].label, cc.ldata[j].label)
	})
	return cc
}

func (cc *CompareConfig) VerifyCNAMEAssertions() {

	// In theory these assertions do not need to be tested as they test
	// something that can not happen. In my head I've proved this to be
	// true.  That said, a little paranoia is healthy.  Those familiar
	// with the Therac-25 accident will agree:
	// https://hackaday.com/2015/10/26/killed-by-a-machine-the-therac-25/

	// According to the RFCs if a label has a CNAME, it can not have any other
	// records at that label... even other CNAMEs.  Therefore, we need to be
	// careful with changes at a label that involve a CNAME.
	// Example 1:
	//   OLD: a.example.com CNAME b
	//   NEW: a.example.com A 1.2.3.4
	//   We must delete the CNAME record THEN create the A record.  If we
	//   blindly create the the A first, most APIs will reply with an error
	//   because there is already a CNAME at that label.
	// Example 2:
	//   OLD: a.example.com A 1.2.3.4
	//   NEW: a.example.com CNAME b
	//   We must delete the A record THEN create the CNAME.  If we blindly
	//   create the CNAME first, most APIs will reply with an error because
	//   there is already an A record at that label.
	//
	// To assure that DNS providers don't have to think about this, we order
	// the tdata items so that we generate the instructions in the best order.
	// In other words:
	//     If there is a CNAME in existing, it should be in front.
	//     If there is a CNAME in desired, it should be at the end.

	// That said, since we addRecords existing first, and desired last, the data
	// should already be in the right order.

	for _, ld := range cc.ldata {
		for j, td := range ld.tdata {
			if td.rType == "CNAME" {
				if len(td.existingTargets) != 0 {
					//fmt.Printf("DEBUG: cname in existing: index=%d\n", j)
					if j != 0 {
						panic("should not happen: (CNAME not in first position)")
					}
				}
				if len(td.desiredTargets) != 0 {
					//fmt.Printf("DEBUG: cname in desired: index=%d\n", j)
					if j != highest(ld.tdata) {
						panic("should not happen: (CNAME not in last position)")
					}
				}
			}
		}
	}

}

// String returns cc represented as a string. This is used for
// debugging and unit tests, as the structure may otherwise be
// difficult to compare.
func (cc *CompareConfig) String() string {
	var buf bytes.Buffer
	b := &buf

	fmt.Fprintf(b, "ldata:\n")
	for i, ld := range cc.ldata {
		fmt.Fprintf(b, "  ldata[%02d]: %s\n", i, ld.label)
		for j, t := range ld.tdata {
			fmt.Fprintf(b, "             tdata[%d]: %q e(%d, %d) d(%d, %d)\n", j, t.rType,
				len(t.existingTargets),
				len(t.existingRecs),
				len(t.desiredTargets),
				len(t.desiredRecs),
			)
		}
	}
	fmt.Fprintf(b, "labelMap: len=%d %v\n", len(cc.labelMap), cc.labelMap)
	fmt.Fprintf(b, "keyMap:   len=%d %v\n", len(cc.keyMap), cc.keyMap)
	fmt.Fprintf(b, "existing: %q\n", cc.existing)
	fmt.Fprintf(b, "desired: %q\n", cc.desired)
	fmt.Fprintf(b, "origin: %v\n", cc.origin)
	fmt.Fprintf(b, "compFn: %v\n", cc.compareableFunc)

	return b.String()
}

// Generate a string that can be used to compare this record to others
// for equality.
func comparable(rc *models.RecordConfig, f func(*models.RecordConfig) string) string {
	if f == nil {
		return rc.ToDiffable()
	}
	return rc.ToDiffable() + " " + f(rc)
}

func (cc *CompareConfig) addRecords(recs models.Records, storeInExisting bool) {
	// storeInExisting indicates if the records should be stored in the
	// cc.existing* fields (true) or the cc.desired* fields (false).

	// Sort, because sorted data is easier to work with.
	// NB(tlim): The actual sort order doesn't matter as long as all the records
	// of the same label+rtype are grouped. We use PrettySort because it works,
	// has been extensively tested, and assures that the ChangeList will
	// be in an order that is pretty to look at.
	z := prettyzone.PrettySort(recs, cc.origin, 0, nil)

	for _, rec := range z.Records {

		label := rec.NameFQDN
		rtype := rec.Type
		comp := comparable(rec, cc.compareableFunc)

		// Are we seeing this label for the first time?
		var labelIdx int
		if _, ok := cc.labelMap[label]; !ok {
			//fmt.Printf("DEBUG: I haven't see label=%v before. Adding.\n", label)
			cc.labelMap[label] = true
			cc.ldata = append(cc.ldata, &labelConfig{label: label})
			labelIdx = highest(cc.ldata)
		} else {
			// find label in cc.ldata:
			for k, v := range cc.ldata {
				if v.label == label {
					labelIdx = k
					break
				}
			}
		}

		// Are we seeing this label+rtype for the first time?
		key := rec.Key()
		if _, ok := cc.keyMap[key]; !ok {
			//fmt.Printf("DEBUG: I haven't see key=%v before. Adding.\n", key)
			cc.keyMap[key] = true
			x := cc.ldata[labelIdx]
			//fmt.Printf("DEBUG: appending rtype=%v\n", rtype)
			x.tdata = append(x.tdata, &rTypeConfig{rType: rtype})
		}
		var rtIdx int
		// find rtype in tdata:
		for k, v := range cc.ldata[labelIdx].tdata {
			if v.rType == rtype {
				rtIdx = k
				break
			}
		}
		//fmt.Printf("DEBUG: found rtype=%v at index %d\n", rtype, rtIdx)

		// Now it is safe to add/modify the records.

		//fmt.Printf("BEFORE E/D: %v/%v\n", len(td.existingRecs), len(td.desiredRecs))
		if storeInExisting {
			cc.ldata[labelIdx].tdata[rtIdx].existingRecs = append(cc.ldata[labelIdx].tdata[rtIdx].existingRecs, rec)
			cc.ldata[labelIdx].tdata[rtIdx].existingTargets = append(cc.ldata[labelIdx].tdata[rtIdx].existingTargets,
				targetConfig{compareable: comp, rec: rec})
		} else {
			cc.ldata[labelIdx].tdata[rtIdx].desiredRecs = append(cc.ldata[labelIdx].tdata[rtIdx].desiredRecs, rec)
			cc.ldata[labelIdx].tdata[rtIdx].desiredTargets = append(cc.ldata[labelIdx].tdata[rtIdx].desiredTargets,
				targetConfig{compareable: comp, rec: rec})
		}
		//fmt.Printf("AFTER  L: %v\n", len(cc.ldata))
		//fmt.Printf("AFTER  E/D: %v/%v\n", len(td.existingRecs), len(td.desiredRecs))
		//fmt.Printf("\n")
	}
}
