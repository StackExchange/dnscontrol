package diff2

import (
	"bytes"
	"fmt"

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
	return cc
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
			//fmt.Printf("DEBUG: APpending rtype=%v\n", rtype)
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
