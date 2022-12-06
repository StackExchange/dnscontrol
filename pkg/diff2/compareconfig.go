package diff2

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/prettyzone"
)

type ComparableFunc func(*models.RecordConfig) string

type CompareConfig struct {
	existing, desired models.Records
	ldata             []*labelConfig
	//
	origin          string // Domain zone
	compareableFunc ComparableFunc
	//
	labelMap map[string]bool
	keyMap   map[models.RecordKey]bool
}

type labelConfig struct {
	label string
	tdata []*rTypeConfig
}

type rTypeConfig struct {
	rType           string
	existingTargets []targetConfig
	desiredTargets  []targetConfig
	existingRecs    []*models.RecordConfig
	desiredRecs     []*models.RecordConfig
}

type targetConfig struct {
	compareable string
	rec         *models.RecordConfig
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
	cc.addRecords(existing, true)
	cc.addRecords(desired, false)
	cc.VerifyCNAMEAssertions()
	sort.Slice(cc.ldata, func(i, j int) bool {
		return prettyzone.LabelLess(cc.ldata[i].label, cc.ldata[j].label)
	})
	return cc
}

func (cc *CompareConfig) VerifyCNAMEAssertions() {

	// NB(tlim): This can be deleted. This should be probably not possible.
	// However, let's keep it around for a few iterations to be paranoid.

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

func comparable(rc *models.RecordConfig, f func(*models.RecordConfig) string) string {
	if f == nil {
		return rc.ToDiffable()
	}
	return rc.ToDiffable() + " " + f(rc)
}

func (cc *CompareConfig) addRecords(recs models.Records, storeInExisting bool) {

	// Sort, because sorted data is easier to work with.
	// NB(tlim): The actual sort order doesn't matter as long as all the records
	// of the same label+rtype are adjacent. We use PrettySort because it works,
	// has been extensively tested, and assures that the ChangeList will be a
	// pretty order.
	//for _, rec := range recs {
	z := prettyzone.PrettySort(recs, cc.origin, 0, nil)
	for _, rec := range z.Records {

		label := rec.NameFQDN
		rtype := rec.Type
		comp := comparable(rec, cc.compareableFunc)
		//fmt.Printf("DEBUG addRecords rec=%v:%v es=%v comp=%v\n", label, rtype, storeInExisting, comp)

		//fmt.Printf("BEFORE L: %v\n", len(cc.ldata))
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
			cc.ldata[labelIdx].tdata[rtIdx].existingTargets = append(cc.ldata[labelIdx].tdata[rtIdx].existingTargets, targetConfig{compareable: comp, rec: rec})
		} else {
			cc.ldata[labelIdx].tdata[rtIdx].desiredRecs = append(cc.ldata[labelIdx].tdata[rtIdx].desiredRecs, rec)
			cc.ldata[labelIdx].tdata[rtIdx].desiredTargets = append(cc.ldata[labelIdx].tdata[rtIdx].desiredTargets, targetConfig{compareable: comp, rec: rec})
		}
		//fmt.Printf("AFTER  L: %v\n", len(cc.ldata))
		//fmt.Printf("AFTER  E/D: %v/%v\n", len(td.existingRecs), len(td.desiredRecs))
		//fmt.Printf("\n")
	}
}
