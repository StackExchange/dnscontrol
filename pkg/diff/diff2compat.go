package diff

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff2"
)

// NewCompat is a constructor that uses the new pkg/diff2 system
// instead of pkg/diff.
//
// It is for backwards compatibility only. New providers should use
// pkg/diff2.  Older providers should use this to reduce their
// dependency on pkg/diff2 until they can move to the pkg/diff2/By*()
// functions.
//
// To use this simply change New() to NewCompat(). If that doesn't
// work please report a bug. The only exception is if you depend on
// the extraValues feature, which will not be supported. That
// parameter must be set to nil.
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

// differCompat meets the Differ interface but provides its service
// using pkg/diff2 instead of pkg/diff.
type differCompat struct {
	OldDiffer *differ // Store the backwards-compatible "d" for pkg/diff

	dc *models.DomainConfig
}

// IncrementalDiff generates the diff using the pkg/diff2 code.
// NOTE: While this attempts to be backwards compatible, it does not
// support all features of the old system:
//   - The IncrementalDiff() `unchanged` return value is always empty.
//     Most providers ignore this return value. If a provider depends on
//     that result, please consider one of the pkg/diff2/By*() functions
//     instead.  (ByZone() is likely to be what you need)
//   - The NewCompat() feature `extraValues` is not supported. That
//     parameter must be set to nil.  If you use that feature, consider
//     one of the pkg/diff2/By*() functions.
func (d *differCompat) IncrementalDiff(existing []*models.RecordConfig) (unchanged, toCreate, toDelete, toModify Changeset, err error) {
	instructions, err := diff2.ByRecord(existing, d.dc, nil)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	for _, inst := range instructions {
		cor := Correlation{d: d.OldDiffer}
		switch inst.Type {
		case diff2.REPORT:
			// Sadly the NewCompat function doesn't have an equivalent. We
			// just output the messages now.
			fmt.Println(inst.MsgsJoined)
		case diff2.CREATE:
			cor.Desired = inst.New[0]
			toCreate = append(toCreate, cor)
		case diff2.CHANGE, diff2.MODIFYTTL:
			cor.Existing = inst.Old[0]
			cor.Desired = inst.New[0]
			toModify = append(toModify, cor)
		case diff2.DELETE:
			cor.Existing = inst.Old[0]
			toDelete = append(toDelete, cor)
		default:
			panic(fmt.Sprintf("unhandled inst.Type %s", inst.Type))
		}
	}

	return
}

// ChangedGroups provides the same results as IncrementalDiff but grouped by key.
func (d *differCompat) ChangedGroups(existing []*models.RecordConfig) (map[models.RecordKey][]string, error) {
	changedKeys := map[models.RecordKey][]string{}
	_, toCreate, toDelete, toModify, err := d.IncrementalDiff(existing)
	if err != nil {
		return nil, err
	}
	for _, c := range toCreate {
		changedKeys[c.Desired.Key()] = append(changedKeys[c.Desired.Key()], c.String())
	}
	for _, d := range toDelete {
		changedKeys[d.Existing.Key()] = append(changedKeys[d.Existing.Key()], d.String())
	}
	for _, m := range toModify {
		changedKeys[m.Desired.Key()] = append(changedKeys[m.Desired.Key()], m.String())
	}
	return changedKeys, nil
}
