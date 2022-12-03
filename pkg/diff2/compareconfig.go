package diff2

import (
	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/prettyzone"
)

type ComparableFunc func(*models.RecordConfig) string

type CompareConfig struct {
	existing, desired models.Records
	ldata             []labelConfig
	//
	origin          string // Domain zone
	compareableFunc ComparableFunc
	//
	labelMap map[string]bool
	keyMap   map[models.RecordKey]bool
}

type labelConfig struct {
	label string
	tdata []rTypeConfig
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

func comparable(rc *models.RecordConfig, f func(*models.RecordConfig) string) string {
	if f == nil {
		return rc.ToDiffable()
	}
	return rc.ToDiffable() + " " + f(rc)
}

func (cc *CompareConfig) addRecords(recs models.Records, existing bool) {

	// Sort, because sorted data is easier to work with.
	// NB(tlim): The actual sort order doesn't matter as long as all the records
	// of the same label+rtype are adjacent. We use PrettySort because it works,
	// has been extensively tested, and assures that the ChangeList will be a
	// pretty order.
	z := prettyzone.PrettySort(recs, cc.origin, 0, nil)

	for _, rec := range z.Records {
		label := rec.NameFQDN
		rtype := rec.Type
		comp := comparable(rec, cc.compareableFunc)

		// Are we seeing this label for the first time?
		var labelIdx int
		if _, ok := cc.labelMap[label]; !ok {
			cc.labelMap[label] = true
			cc.ldata = append(cc.ldata, labelConfig{})
		}
		labelIdx = highest(cc.ldata)

		// Are we seeing this label+rtype for the first time?
		key := rec.Key()
		if _, ok := cc.keyMap[key]; !ok {
			cc.keyMap[key] = true
			cc.ldata[labelIdx].tdata = append(cc.ldata[labelIdx].tdata, rTypeConfig{})
		}
		rtIdx := highest(cc.ldata[labelIdx].tdata)

		// Now it is safe to add/modify the records.

		cc.ldata[labelIdx].label = label
		cc.ldata[labelIdx].tdata[rtIdx].rType = rtype
		if existing {
			cc.ldata[labelIdx].tdata[rtIdx].existingRecs = append(
				cc.ldata[labelIdx].tdata[rtIdx].existingRecs,
				rec)
			cc.ldata[labelIdx].tdata[rtIdx].existingTargets = append(
				cc.ldata[labelIdx].tdata[rtIdx].existingTargets,
				targetConfig{compareable: comp, rec: rec})
		} else {
			cc.ldata[labelIdx].tdata[rtIdx].desiredRecs = append(
				cc.ldata[labelIdx].tdata[rtIdx].desiredRecs,
				rec)
			cc.ldata[labelIdx].tdata[rtIdx].desiredTargets = append(
				cc.ldata[labelIdx].tdata[rtIdx].desiredTargets,
				targetConfig{compareable: comp, rec: rec})
		}
	}
}
