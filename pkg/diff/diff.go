package diff

import (
	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/fatih/color"
)

// Correlation stores a difference between two records.
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

type differ struct {
}

// get normalized content for record. target, ttl, mxprio, and specified metadata
func (d *differ) content(r *models.RecordConfig) string {
	return r.ToDiffable()
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

func (c Correlation) String() string {
	if c.Existing == nil {
		return color.GreenString("+ CREATE %s %s %s", c.Desired.Type, c.Desired.GetLabelFQDN(), c.d.content(c.Desired))
	}
	if c.Desired == nil {
		return color.RedString("- DELETE %s %s %s", c.Existing.Type, c.Existing.GetLabelFQDN(), c.d.content(c.Existing))
	}
	return color.YellowString("± MODIFY %s %s: (%s) -> (%s)", c.Existing.Type, c.Existing.GetLabelFQDN(), c.d.content(c.Existing), c.d.content(c.Desired))
}
