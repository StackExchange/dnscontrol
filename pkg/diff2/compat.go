package diff2

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v3/models"
)

// Provide an interface that is backwards compatible with pkg/diff.

type Differ struct {
	dc          *models.DomainConfig
	extraValues []func(*models.RecordConfig) map[string]string

	//compiledIgnoredNames   []ignoredName
	//compiledIgnoredTargets []glob.Glob
}

// Correlation stores a difference between two records.
type Correlation struct {
	d        *Differ
	Existing *models.RecordConfig
	Desired  *models.RecordConfig
}

// Changeset stores many Correlation.
type Changeset []Correlation

// New is a constructor for a Differ.
func New(dc *models.DomainConfig, extraValues ...func(*models.RecordConfig) map[string]string) *Differ {
	if len(extraValues) != 0 {
		panic("compat.New() does not implement extraValues")
	}
	return &Differ{
		dc:          dc,
		extraValues: extraValues,

		// compile IGNORE_NAME glob patterns
		//compiledIgnoredNames: compileIgnoredNames(dc.IgnoredNames),

		// compile IGNORE_TARGET glob patterns
		//compiledIgnoredTargets: compileIgnoredTargets(dc.IgnoredTargets),
	}
}

// IncrementalDiff provides an interface that is compatible with
// pkg/diff/IncrementalDiffCompat(). It not as efficient as converting
// to the diff2.By*() functions. However, using this is better than
// staying with pkg/diff.
func (d *Differ) IncrementalDiff(existing []*models.RecordConfig) (unchanged, create, toDelete, modify Changeset, err error) {
	unchanged = Changeset{}
	create = Changeset{}
	toDelete = Changeset{}
	modify = Changeset{}
	//desired := d.dc.Records

	instructions, err := ByRecord(existing, d.dc, nil)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	for _, inst := range instructions {
		cor := Correlation{d: d}
		switch inst.Type {
		case CREATE:
			cor.Desired = inst.New[0]
			create = append(create, cor)
		case CHANGE:
			cor.Existing = inst.Old[0]
			cor.Desired = inst.New[0]
			modify = append(modify, cor)
		case DELETE:
			cor.Existing = inst.Old[0]
			toDelete = append(toDelete, cor)
		}
	}

	return
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

// get normalized content for record. target, ttl, mxprio, and specified metadata
func (d *Differ) content(r *models.RecordConfig) string {

	// get the extra values maps to add to the comparison.
	var allMaps []map[string]string
	for _, f := range d.extraValues {
		valueMap := f(r)
		allMaps = append(allMaps, valueMap)
	}

	return r.ToDiffable(allMaps...)
}
