package diff2

import (
	"fmt"
	"sort"

	"github.com/StackExchange/dnscontrol/v3/models"
)

func analyzeByRecordSet(cc *CompareConfig) ChangeList {
	var instructions ChangeList
	// For each label, for each type at that label, if there are changes generate an instruction.
	for _, lc := range cc.ldata {
		for _, rt := range lc.tdata {
			ets := rt.existingTargets
			dts := rt.desiredTargets
			msgs := genmsgs(ets, dts)
			if len(msgs) == 0 {
				// The records at this rset are the same. No work to be done.
				continue
			}
			if len(ets) == 0 { // Create a new label.
				instructions = append(instructions, add(lc.label, rt.rType, msgs, rt.desiredRecs))
			} else if len(dts) == 0 { // Delete that label and all its records.
				instructions = append(instructions, delete(lc.label, rt.rType, msgs))
			} else { // Change the records at that label
				instructions = append(instructions, change(lc.label, rt.rType, msgs, rt.desiredRecs))
			}
		}
	}
	return instructions
}

func analyzeByLabel(cc *CompareConfig) ChangeList {
	var instructions ChangeList
	// Accumulate if there are any changes and collect the info needed to generate instructions.
	for _, lc := range cc.ldata {
		label := lc.label
		hadExisting := false
		var accMsgs []string
		var accDesired models.Records
		for _, rt := range lc.tdata {
			ets := rt.existingTargets
			dts := rt.desiredTargets
			msgs := genmsgs(ets, dts)
			if len(msgs) != 0 { // If there were differences
				accMsgs = append(accMsgs, msgs...)                 // Accumulate the messages
				accDesired = append(accDesired, rt.desiredRecs...) // Accumulate what records should be at this label.
				if len(rt.existingRecs) != 0 {                     // Are we adding to an existing label?
					hadExisting = true
				}
			}
		}

		// We now know what changed (accMsgs), what records should exist at that
		// label (accDesired), and if any old records used to exist at this label
		// (hadExisting).  Based on that info, we can generate the instructions.

		if len(accMsgs) == 0 { // Nothing changed.
			return nil
		}
		if len(accDesired) == 0 { // No new records at the label? This must be a delete.
			instructions = append(instructions, delete(label, "", accMsgs))
		} else if !hadExisting { // No old records at the label? This must be a create/add.
			instructions = append(instructions, add(label, "", accMsgs, accDesired))
		} else { // Must be a change
			instructions = append(instructions, add(label, "", accMsgs, accDesired))
		}
	}

	return instructions
}

func analyzeByRecord(cc *CompareConfig) ChangeList {
	var instructions ChangeList
	// For each label, for each type at that label, see if there are any changes.
	for _, lc := range cc.ldata {
		for _, rt := range lc.tdata {
			ets := rt.existingTargets
			dts := rt.desiredTargets
			cs := diffTargets(ets, dts)
			instructions = append(instructions, cs...)
		}
	}
	return instructions
}

func analyzeByZone(cc *CompareConfig) []string {
	var accMsgs []string
	// For each label, for each type at that label, see if there are any changes.
	for _, lc := range cc.ldata {
		for _, rt := range lc.tdata {
			ets := rt.existingTargets
			dts := rt.desiredTargets
			msgs := genmsgs(ets, dts)
			accMsgs = append(accMsgs, msgs...) // Accumulate the messages
		}
	}
	return accMsgs
}

func add(l string, t string, msgs []string, recs models.Records) Change {
	c := Change{Type: ADD, Msgs: msgs}
	c.Key.NameFQDN = l
	c.Key.Type = t
	c.New = recs
	return c
}
func change(l string, t string, msgs []string, recs models.Records) Change {
	c := Change{Type: CHANGE, Msgs: msgs}
	c.Key.NameFQDN = l
	c.Key.Type = t
	c.New = recs
	return c
}
func delete(l string, t string, msgs []string) Change {
	c := Change{Type: DELETE, Msgs: msgs}
	c.Key.NameFQDN = l
	c.Key.Type = t
	return c
}
func deleteRec(l string, t string, msgs []string, rec *models.RecordConfig) Change {
	c := Change{Type: DELETE, Msgs: msgs}
	c.Key.NameFQDN = l
	c.Key.Type = t
	c.Old = models.Records{rec}
	return c
}

func diffTargets(existing, desired []targetConfig) ChangeList {

	// Nothing to do?
	if len(existing) == 0 && len(desired) == 0 {
		return nil
	}

	// Sort to make comparisons easier
	sort.Slice(existing, func(i, j int) bool { return existing[i].compareable < existing[j].compareable })
	sort.Slice(desired, func(i, j int) bool { return desired[i].compareable < desired[j].compareable })

	var instructions ChangeList

	ei := 0
	di := 0
	for {
		compe := existing[ei].compareable
		compd := desired[di].compareable
		rece := existing[ei].rec
		recd := desired[di].rec
		if compe < compd { // Delete existing
			instructions = append(instructions,
				delete(rece.NameFQDN, "", []string{
					fmt.Sprintf("DELETE %s %s %s",
						rece.GetLabelFQDN(),
						rece.Type,
						rece.GetTargetCombined(),
					)}))
			ei++
		} else if compe > compd { // Add desired
			instructions = append(instructions,
				add(recd.NameFQDN, "", []string{fmt.Sprintf("CREATE %s %s %s", recd.GetLabelFQDN(), recd.Type, recd.GetTargetCombined())},
					models.Records{recd},
				))
			di++
		} else { // They must be equal. Skip
			ei++
			di++
		}

		// Are we done?
		if ei == len(existing) && di == len(desired) { // Done.
			break
		} else if ei == len(existing) {
			// Existing is depleted. Create the remaining desired.
			for _, recd := range desired[di:] {
				r := recd.rec
				instructions = append(instructions,
					add(
						r.NameFQDN,
						r.Type, // Not needed?
						[]string{
							fmt.Sprintf("CREATE %s %s %s", r.GetLabelFQDN(), r.Type, r.GetTargetCombined()),
						},
						models.Records{r},
					),
				)
			}
		} else {
			// Desired is depleted. Delete the remaining existing.
			for _, rece := range existing[di:] {
				r := rece.rec
				instructions = append(instructions,
					deleteRec(
						r.NameFQDN,
						r.Type, // Not needed?
						[]string{
							fmt.Sprintf("DELETE %s %s %s", r.GetLabelFQDN(), r.Type, r.GetTargetCombined()),
						},
						r,
					),
				)
			}

		}
	}

	return nil
}

func genmsgs(existing, desired []targetConfig) []string {
	cl := diffTargets(existing, desired)
	var msgs []string
	for _, c := range cl {
		msgs = append(msgs, c.Msgs...)
	}
	return msgs
}
