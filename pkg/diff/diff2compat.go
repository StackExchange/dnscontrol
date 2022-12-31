package diff

import (
	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff2"
)

// NewCompat is a constructor that uses the new Diff2 system but provides minimal backwards compatibility.
func NewCompat(dc *models.DomainConfig, extraValues ...func(*models.RecordConfig) map[string]string) Differ {
	if len(extraValues) != 0 {
		panic("extraValues not supported")
	}

	d := New(dc)
	return &differCompat{
		OldDiffer: d.(*differ),

		dc: dc,
	}
}

type differCompat struct {
	OldDiffer *differ // Store the backwards-compatible "d" for pkg/diff

	dc *models.DomainConfig
}

// IncrementalDiff generates the diff using the pkg/diff2 code. It is for
// backwards compatibility only. New providers should use pkg/diff2.  Older
// providers should use this to wean off the pkg/diff2 code until they can
func (d *differCompat) IncrementalDiff(existing []*models.RecordConfig) (unchanged, create, toDelete, modify Changeset, err error) {
	unchanged = Changeset{}
	create = Changeset{}
	toDelete = Changeset{}
	modify = Changeset{}

	instructions, err := diff2.ByRecord(existing, d.dc, nil)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	for _, inst := range instructions {
		cor := Correlation{d: d.OldDiffer}
		switch inst.Type {
		case diff2.CREATE:
			cor.Desired = inst.New[0]
			create = append(create, cor)
		case diff2.CHANGE:
			cor.Existing = inst.Old[0]
			cor.Desired = inst.New[0]
			modify = append(modify, cor)
		case diff2.DELETE:
			cor.Existing = inst.Old[0]
			toDelete = append(toDelete, cor)
		}
	}

	return
}

// ChangedGroups provides the same results as IncrementalDiff but grouped by key.
func (d *differCompat) ChangedGroups(existing []*models.RecordConfig) (map[models.RecordKey][]string, error) {
	changedKeys := map[models.RecordKey][]string{}
	_, create, toDelete, modify, err := d.IncrementalDiff(existing)
	if err != nil {
		return nil, err
	}
	for _, c := range create {
		changedKeys[c.Desired.Key()] = append(changedKeys[c.Desired.Key()], c.String())
	}
	for _, d := range toDelete {
		changedKeys[d.Existing.Key()] = append(changedKeys[d.Existing.Key()], d.String())
	}
	for _, m := range modify {
		changedKeys[m.Desired.Key()] = append(changedKeys[m.Desired.Key()], m.String())
	}
	return changedKeys, nil
}
