package diff2

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v3/models"
)

func analyzeByRecordSet(
	existing models.Records,
	desired models.Records,
	domain string,
) (
	ChangeList,
	error,
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
	*/

	/*
		// Walk the records looking for changes.
		esets := groupbyRSet(existing).sort()
		dsets := groupbyRSet(desired).sort()
		walk esets/dsets in parallel:
			if e.Key < d.Key: delete e; e++; continue
			if e.Key > d.Key: add d; d++; continue
			if e.Key == d.Key: c = recdiff(e, d); if len(c) > 0: add c; continue
		If there are e leftovers, delete them.
		If there are d leftovers, add them.
	*/
	esets := groupbyRSet(existing, domain)
	dsets := groupbyRSet(desired, domain)
	fmt.Printf("esets = %v\n", esets)
	fmt.Printf("dsets = %v\n", dsets)
	instructions := walkRSets(esets, dsets)

	return instructions, nil
}

func walkRSets(esets, dsets []recset) ChangeList {
	var instructions ChangeList

	ei := 0
	di := 0
	for {
		e := &esets[ei]
		d := &esets[di]
		if e.Key.String() < d.Key.String() {
			// generate delete e
			c := Change{Type: DELETE}
			c.Key = e.Key
			c.Old = e.Recs
			instructions = append(instructions, c)
			//
			ei++
		} else if e.Key.String() > d.Key.String() {
			// generate add d
			c := Change{Type: ADD}
			c.Key = d.Key
			c.New = d.Recs
			instructions = append(instructions, c)
			//
			di++
		} else { // equal
			cs := recdiff(esets[ei], dsets[di])
			instructions = append(instructions, cs...)
			//
			ei++
			di++
		}

		if ei == len(esets) && di == len(dsets) {
			// both lists completed at the same time.
			break
		}
		if ei == len(esets) { // e is done; add the remaining e
			// generate add dset[di:]
			for _, rec := range dsets[di:] {
				c := Change{Type: ADD}
				c.Key = rec.Key
				c.New = rec.Recs
				instructions = append(instructions, c)
			}
			break
		}
		if di == len(dsets) { // d is done; delete the remaining e
			// generate delete eset[ei:]
			for _, rec := range esets[ei:] {
				c := Change{Type: DELETE}
				c.Key = rec.Key
				c.Old = rec.Recs
				instructions = append(instructions, c)
			}
			break
		}

	}

	return instructions
}

func comparable(rc models.RecordConfig) string {

	return ""
}

func recdiff(e, d recset) ChangeList {
	var instructions ChangeList

	return instructions
}

/*

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

//	dc := models.DomainConfig{
//		Records: desired,
//	}
//
//	differ := diff.New(&dc)
//	_, changeCS, delCS, modCS, err := differ.IncrementalDiff(existing)
//	if err != nil {
//		return nil, nil, nil, err
//	}
//
//	return changeSetToChange(changeCS), changeSetToChange(delCS), changeSetToChange(modCS), nil

//func changeSetToChange(csl diff.Changeset) []Change {
//
//	cl := make([]Change, len(csl))
//	for i, c := range csl {
//		cl[i] = Change{
//			Old: c.Existing,
//			New: c.Desired,
//			Msg: c.String(),
//		}
//	}
//
//	return cl
