package diff2

import "github.com/StackExchange/dnscontrol/v3/models"

// This module provides functions that "diff" the existing records
// against the desired records.

// Change stores information about a record change.
// If Old is nil, this is a creation.
// If New is nil, this is a deletion.
type Change struct {
	Old *models.RecordConfig
	New *models.RecordConfig
	Msg string // Human-friendly explaination of what this change is.
}

// ByRecord takes two lists of records (existing and desired)
// and returns instructions that allow you to turn existing into
// desired.
//
// The instructions are in a form of records to add, change, delete.
//
// This function is appropriate for a DNS provider with an API that
// updates one record at a time. This is the most typical situation.
func ByRecord(existing, desired models.Records) (creations, deletions, modifications []Change, err error) {
	toCreate, toDelete, toModify, err := analyze(existing, desired)
	if err != nil {
		return nil, nil, nil, err
	}
	return toCreate, toDelete, toModify, nil
}

// ByLabel takes two lists of records (existing and desired) and
// returns instructions that allow you to turn existing into desired.
//
// The instructions are in a form of which labels have records that
// changed and which labels should be deleted entirely.  It also
// returns a map indicating what records should be installed at a
// given label.
//
// This function is appropriate for a DNS provider with an API that
// updates one label at a time (i.e. the only update mechanism is to
// specify a label and all the DNS records at that label).
func ByLabel(existing, desired models.Records) (deletions models.Records, modifiedLabels []string, modifications map[string]models.Records, err error) {
	toCreate, toDelete, toModify, err := analyze(existing, desired)
	if err != nil {
		return nil, nil, nil, err
	}

	// Gather the items to be deleted
	for _, j := range toDelete {
		deletions = append(deletions, j.Old)
	}

	// Gather the labels that will need to be created/updated.
	modified := map[string]bool{}
	for _, c := range toCreate {
		modified[c.New.Name] = true
	}
	for _, c := range toModify {
		modified[c.New.Name] = true
	}

	// Gather the records to install at each label.
	modifications = map[string]models.Records{}
	for _, d := range desired {
		label := d.Name
		if _, ok := modified[label]; ok { // If the label was one that was modified...
			if _, ok := modifications[label]; !ok { // map key doesn't exist yet?
				modifications[label] = models.Records{} // Intialize it
			}
			modifications[label] = append(modifications[label], d)
		}
	}

	// Gather they label names as a list.
	modifiedLabels = make([]string, len(modifications))
	i := 0
	for k := range modifications {
		modifiedLabels[i] = k
		i++
	}

	return deletions, modifiedLabels, modifications, nil
}

// ByZone takes two lists of records (existing and desired) and
// returns text one would output to users describing the change.
//
// Some providers accepts updates that are "all or nothing". That is,
// each update uploads the entire zonefile. In this case, the user
// should see a list of changes. As an optimization, if existing is
// empty, we just output a 1-line message such as:
//
//	WRITING ZONEFILE: zones/example.com.zone
//
// This function is appropriate for a DNS provider with an API that
// updates entire zones at a time. For example BIND.
func ByZone(existing, desired models.Records, firstMsg string) (report []string, err error) {

	// If we are creating the zone from scratch, no need to list all the
	// details.
	if len(existing) == 0 {
		return []string{firstMsg}, nil
	}

	// Document the changes.
	toCreate, toDelete, toModify, err := analyze(existing, desired)
	if err != nil {
		return nil, err
	}

	for _, c := range toDelete {
		report = append(report, c.Msg)
	}
	for _, c := range toCreate {
		report = append(report, c.Msg)
	}
	for _, c := range toModify {
		report = append(report, c.Msg)
	}
	return report, nil
}
