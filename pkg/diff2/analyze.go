package diff2

import (
	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
)

func analyze(
	existing models.Records,
	desired models.Records,
) (
	creations, deletions, modifications []Change,
	err error,
) {

	/*

		// Handle hands_off:
		for each any HANDS_OFF() records:
			exts := query(existing, hands_off_args)
			// Modifying a hands-off record is an error.
			both := intersection(desired, exts)
			if len(both) != 0: error.
			// Add the query results to desired.
			desired = append(desired, exts)

		// Walk the records looking for changes.
		esets := groupbyRSet(existing).sort()
		dsets := groupbyRSet(desired).sort()
		walk esets/dsets in parallel:
			if e.Key < d.Key: delete e; e++; continue
			if e.Key > d.Key: add d; d++; continue
			if e.Key == d.Key: c = recdiff(e, d); if len(c) > 0: add c; continue
		If there are e leftovers, delete them.
		If there are d leftovers, add them.

		recdiff(es, ds):
			de := decorate(es).sort() // returns []struct{ content, orig }
			dd := decorate(es).sort() // returns []struct{ content, orig }
			delete de/dd if content matches
			else if len(de) < len(dd):
				generate changes for 0:len(de)
				generate add for len(de):len(dd)
			else if len(de) > len(dd):
				generate changes for 0:len(dd)
				generate deletes for len(dd):len(de)
			if len(de) == len(dd):
			  generate changes for 0:len(de)

		if no_purge:
			remove any DELETEs
			add a count as a Msgs.

	*/

	dc := models.DomainConfig{
		Records: desired,
	}

	differ := diff.New(&dc)
	_, changeCS, delCS, modCS, err := differ.IncrementalDiff(existing)
	if err != nil {
		return nil, nil, nil, err
	}

	return changeSetToChange(changeCS), changeSetToChange(delCS), changeSetToChange(modCS), nil
}

func changeSetToChange(csl diff.Changeset) []Change {

	cl := make([]Change, len(csl))
	for i, c := range csl {
		cl[i] = Change{
			Old: c.Existing,
			New: c.Desired,
			Msg: c.String(),
		}
	}

	return cl
}
