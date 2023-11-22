package diff

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/diff2"
)

// NewCompat is a constructor that uses the new pkg/diff2 system
// instead of pkg/diff.
//
// It is for backwards compatibility only. New providers should use pkg/diff2.
//
// To use this simply change New() to NewCompat(). If that doesn't
// work please report a bug. The extraValues parameter is not supported.
func NewCompat(dc *models.DomainConfig, extraValues ...func(*models.RecordConfig) map[string]string) Differ {
	if len(extraValues) != 0 {
		panic("extraValues not supported")
	}

	return &differCompat{
		dc: dc,
	}
}

// differCompat meets the Differ interface but provides its service
// using pkg/diff2 instead of pkg/diff.
type differCompat struct {
	dc *models.DomainConfig
}

// IncrementalDiff usees pkg/diff2 to generate output compatible with systems
// still using NewCompat().
func (d *differCompat) IncrementalDiff(existing []*models.RecordConfig) (reportMsgs []string, toCreate, toDelete, toModify Changeset, err error) {
	instructions, err := diff2.ByRecord(existing, d.dc, nil)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	for _, inst := range instructions {
		cor := Correlation{}
		switch inst.Type {
		case diff2.REPORT:
			reportMsgs = append(reportMsgs, inst.Msgs...)
		case diff2.CREATE:
			cor.Desired = inst.New[0]
			toCreate = append(toCreate, cor)
		case diff2.CHANGE:
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

func GenerateMessageCorrections(msgs []string) (corrections []*models.Correction) {
	for _, msg := range msgs {
		corrections = append(corrections, &models.Correction{Msg: msg})
	}
	return
}

// ChangedGroups provides the same results as IncrementalDiff but grouped by key.
func (d *differCompat) ChangedGroups(existing []*models.RecordConfig) (map[models.RecordKey][]string, []string, error) {
	changedKeys := map[models.RecordKey][]string{}
	toReport, toCreate, toDelete, toModify, err := d.IncrementalDiff(existing)
	if err != nil {
		return nil, nil, err
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
	return changedKeys, toReport, nil
}
