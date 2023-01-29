package diff2

import (
	"fmt"
	"sort"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
)

func analyzeByRecordSet(cc *CompareConfig) ChangeList {
	var instructions ChangeList
	// For each label...
	for _, lc := range cc.ldata {
		// for each type at that label...
		for _, rt := range lc.tdata {
			// ...if there are changes generate an instruction.
			ets := rt.existingTargets
			dts := rt.desiredTargets
			msgs := genmsgs(ets, dts)
			if len(msgs) == 0 { // No differences?
				//fmt.Printf("DEBUG: done. Records are the same\n")
				// The records at this rset are the same. No work to be done.
				continue
			}
			if len(ets) == 0 { // Create a new label.
				//fmt.Printf("DEBUG: add\n")
				instructions = append(instructions, mkAdd(lc.label, rt.rType, msgs, rt.desiredRecs))
			} else if len(dts) == 0 { // Delete that label and all its records.
				//fmt.Printf("DEBUG: delete\n")
				instructions = append(instructions, mkDelete(lc.label, rt.rType, msgs, rt.existingRecs))
			} else { // Change the records at that label
				//fmt.Printf("DEBUG: change\n")
				instructions = append(instructions, mkChange(lc.label, rt.rType, msgs, rt.existingRecs, rt.desiredRecs))
			}
		}
	}
	return instructions
}

func analyzeByLabel(cc *CompareConfig) ChangeList {
	var instructions ChangeList
	//fmt.Printf("DEBUG: START: analyzeByLabel\n")
	// Accumulate if there are any changes and collect the info needed to generate instructions.
	for i, lc := range cc.ldata {
		//fmt.Printf("DEBUG: START LABEL = %q\n", lc.label)
		label := lc.label
		var accMsgs []string
		var accExisting models.Records
		var accDesired models.Records
		msgsByKey := map[models.RecordKey][]string{}
		for _, rt := range lc.tdata {
			//fmt.Printf("DEBUG: START RTYPE = %q\n", rt.rType)
			ets := rt.existingTargets
			dts := rt.desiredTargets
			msgs := genmsgs(ets, dts)
			msgsByKey[models.RecordKey{NameFQDN: label, Type: rt.rType}] = msgs
			//fmt.Printf("DEBUG:    appending msgs=%v\n", msgs)
			accMsgs = append(accMsgs, msgs...)                    // Accumulate the messages
			accExisting = append(accExisting, rt.existingRecs...) // Accumulate records existing at this label.
			accDesired = append(accDesired, rt.desiredRecs...)    // Accumulate records desired at this label.
		}

		// We now know what changed (accMsgs),
		// what records USED TO EXIST at that label (accExisting),
		// and what records SHOULD EXIST at that label (accDesired).
		// Based on that info, we can generate the instructions.

		if len(accMsgs) == 0 { // Nothing changed.
			//fmt.Printf("DEBUG: analyzeByLabel: %02d: no change\n", i)
		} else if len(accDesired) == 0 { // No new records at the label? This must be a delete.
			//fmt.Printf("DEBUG: analyzeByLabel: %02d: delete\n", i)
			instructions = append(instructions, mkDelete(label, "", accMsgs, accExisting))
		} else if len(accExisting) == 0 { // No old records at the label? This must be a change.
			//fmt.Printf("DEBUG: analyzeByLabel: %02d: create\n", i)
			//fmt.Printf("DEBUG: analyzeByLabel mkAdd msgs=%d\n", len(accMsgs))
			instructions = append(instructions, mkAddByLabel(label, "", accMsgs, accDesired))
		} else { // If we get here, it must be a change.
			_ = i
			// fmt.Printf("DEBUG: analyzeByLabel: %02d: change %d{%v} %d{%v} msgs=%v\n", i,
			// 	len(accExisting), accExisting,
			// 	len(accDesired), accDesired,
			// 	accMsgs,
			// )
			//fmt.Printf("DEBUG: analyzeByLabel mkchange msgs=%d\n", len(accMsgs))
			instructions = append(instructions, mkChangeLabel(label, "", accMsgs, accExisting, accDesired, msgsByKey))
		}
	}

	return instructions
}

func analyzeByRecord(cc *CompareConfig) ChangeList {
	//fmt.Printf("DEBUG: analyzeByRecord: cc=%v\n", cc)

	var instructions ChangeList
	// For each label, for each type at that label, see if there are any changes.
	for _, lc := range cc.ldata {
		//fmt.Printf("DEBUG: analyzeByRecord: next lc=%v\n", lc)
		for _, rt := range lc.tdata {
			ets := rt.existingTargets
			dts := rt.desiredTargets
			cs := diffTargets(ets, dts)
			//fmt.Printf("DEBUG: analyzeByRecord: cs=%v\n", cs)
			instructions = append(instructions, cs...)
		}
	}
	return instructions
}

// NB(tlim): there is no analyzeByZone.  ByZone calls anayzeByRecords().

func mkAdd(l string, t string, msgs []string, recs models.Records) Change {
	//fmt.Printf("DEBUG mkAdd called: (%v, %v, %v, %v)\n", l, t, msgs, recs)
	c := Change{Type: CREATE, Msgs: msgs, MsgsJoined: strings.Join(msgs, "\n")}
	c.Key.NameFQDN = l
	c.Key.Type = t
	c.New = recs
	return c
}

// TODO(tlim): Clean these up. Some of them are exact duplicates!

func mkAddByLabel(l string, t string, msgs []string, newRecs models.Records) Change {
	//fmt.Printf("DEBUG mkAddByLabel called: (%v, %v, %v, %v)\n", l, t, msgs, newRecs)
	c := Change{Type: CREATE, Msgs: msgs, MsgsJoined: strings.Join(msgs, "\n")}
	c.Key.NameFQDN = l
	c.Key.Type = t
	c.New = newRecs
	return c
}

func mkChange(l string, t string, msgs []string, oldRecs, newRecs models.Records) Change {
	//fmt.Printf("DEBUG mkChange called: (%v, %v, %v, %v, %v)\n", l, t, msgs, oldRecs, newRecs)
	c := Change{Type: CHANGE, Msgs: msgs, MsgsJoined: strings.Join(msgs, "\n")}
	c.Key.NameFQDN = l
	c.Key.Type = t
	c.Old = oldRecs
	c.New = newRecs
	return c
}

func mkChangeLabel(l string, t string, msgs []string, oldRecs, newRecs models.Records, msgsByKey map[models.RecordKey][]string) Change {
	//fmt.Printf("DEBUG mkChangeLabel called: (%v, %v, %v, %v, %v, %v)\n", l, t, msgs, oldRecs, newRecs, msgsByKey)
	c := Change{Type: CHANGE, Msgs: msgs, MsgsJoined: strings.Join(msgs, "\n")}
	c.Key.NameFQDN = l
	c.Key.Type = t
	c.Old = oldRecs
	c.New = newRecs
	c.MsgsByKey = msgsByKey
	return c
}

func mkDelete(l string, t string, msgs []string, oldRecs models.Records) Change {
	//fmt.Printf("DEBUG mkDelete called: (%v, %v, %v, %v)\n", l, t, msgs, oldRecs)
	if len(msgs) == 0 {
		panic("mkDelete with no msg")
	}
	c := Change{Type: DELETE, Msgs: msgs, MsgsJoined: strings.Join(msgs, "\n")}
	c.Key.NameFQDN = l
	c.Key.Type = t
	c.Old = oldRecs
	return c
}
func mkDeleteRec(l string, t string, msgs []string, rec *models.RecordConfig) Change {
	//fmt.Printf("DEBUG mkDeleteREc called: (%v, %v, %v, %v)\n", l, t, msgs, rec)
	c := Change{Type: DELETE, Msgs: msgs, MsgsJoined: strings.Join(msgs, "\n")}
	c.Key.NameFQDN = l
	c.Key.Type = t
	c.Old = models.Records{rec}
	return c
}

func removeCommon(existing, desired []targetConfig) ([]targetConfig, []targetConfig) {

	// NB(tlim): We could probably make this faster. Some ideas:
	// * pre-allocate newexisting/newdesired and assign to indexed elements instead of appending.
	// * iterate backwards (len(d) to 0) and delete items that are the same.
	// On the other hand, this function typically receives lists of 1-3 elements
	// and any optimization is probably fruitless.

	eKeys := map[string]*targetConfig{}
	for _, v := range existing {
		v := v
		eKeys[v.compareable] = &v
	}
	dKeys := map[string]*targetConfig{}
	for _, v := range desired {
		v := v
		dKeys[v.compareable] = &v
	}

	return filterBy(existing, dKeys), filterBy(desired, eKeys)
}

// Find the changes that are exclusively changes in TTL.
func splitTTLOnly(existing, desired []targetConfig) (
	existDiff []targetConfig, desireDiff []targetConfig,
	existTTL models.Records, desireTTL models.Records,
) {
	ei := 0
	di := 0

	for (ei < len(existing)) && (di < len(desired)) {
		er := existing[ei].rec
		dr := desired[di].rec
		ecomp := er.GetTargetCombined()
		dcomp := dr.GetTargetCombined()

		if ecomp == dcomp && er.TTL == dr.TTL {
			panic("Should not happen. There should be some difference!")
		}

		//fmt.Printf("DEBUG ecomp=%q dcomp=%q ettl=%d dttl=%d\n", ecomp, dcomp, er.TTL, dr.TTL)
		if ecomp == dcomp && er.TTL != dr.TTL {
			//fmt.Printf("DEBUG: equal\n")
			existTTL = append(existTTL, er)
			desireTTL = append(desireTTL, dr)
			ei++
			di++
		} else if ecomp < dcomp {
			//fmt.Printf("DEBUG: less\n")
			existDiff = append(existDiff, existing[ei])
			ei++
		} else if ecomp > dcomp {
			//fmt.Printf("DEBUG: greater\n")
			desireDiff = append(desireDiff, desired[di])
			di++
		} else {
			panic("Should not happen. e and d can not be both equal, less and greater")
		}

	}

	// Any remainder goes to the *Diff result:
	if ei < len(existing) {
		//fmt.Printf("DEBUG: append e len()=%d\n", ei)
		existDiff = append(existDiff, existing[ei:]...)
	}
	if di < len(desired) {
		//fmt.Printf("DEBUG: append d len()=%d\n", di)
		desireDiff = append(desireDiff, desired[di:]...)
	}

	return
}

// Return s but remove any items that can be found in m.
func filterBy(s []targetConfig, m map[string]*targetConfig) []targetConfig {
	// fmt.Printf("DEBUG: filterBy called with %v\n", s)
	// for k := range m {
	// 	fmt.Printf("DEBUG:    map %q\n", k)
	// }
	i := 0 // output index
	for _, x := range s {
		if _, ok := m[x.compareable]; !ok {
			//fmt.Printf("DEBUG: comp %q NO\n", x.compareable)
			// copy and increment index
			s[i] = x
			i++
		}
	}
	// // Prevent memory leak by erasing truncated values
	// // (not needed if values don't contain pointers, directly or indirectly)
	// for j := i; j < len(s); j++ {
	// 	s[j] = nil
	// }
	s = s[:i]
	// fmt.Printf("DEBUG: filterBy returns %v\n", s)
	return s
}

func humanDiff(a, b *models.RecordConfig) string {
	acombined := a.GetTargetCombined()
	bcombined := b.GetTargetCombined()
	combinedDiff := acombined != bcombined
	ttlDiff := a.TTL != b.TTL
	// TODO(tlim): It would be nice if we included special cases for MX
	// records and others.
	if combinedDiff && ttlDiff {
		return fmt.Sprintf("(%s ttl=%d) -> (%s ttl=%d)", acombined, a.TTL, bcombined, b.TTL)
	}
	if combinedDiff {
		return fmt.Sprintf("(%s) -> (%s)", acombined, bcombined)
	}
	return fmt.Sprintf("%s (ttl %d->%d)", acombined, a.TTL, b.TTL)
}

func diffTargets(existing, desired []targetConfig) ChangeList {
	//fmt.Printf("DEBUG: diffTargets called with len(e)=%d len(d)=%d\n", len(existing), len(desired))

	// Nothing to do?
	if len(existing) == 0 && len(desired) == 0 {
		//fmt.Printf("DEBUG: diffTargets: nothing to do\n")
		return nil
	}

	// Sort to make comparisons easier
	sort.Slice(existing, func(i, j int) bool { return existing[i].compareable < existing[j].compareable })
	sort.Slice(desired, func(i, j int) bool { return desired[i].compareable < desired[j].compareable })

	var instructions ChangeList

	// remove the exact matches.
	existing, desired = removeCommon(existing, desired)

	// At this point the exact matches are removed. However there may be
	// records that have the same GetTargetCombined() but different
	// TTLs.

	// Find TTL changes:
	existing, desired, existingTTL, desiredTTL := splitTTLOnly(existing, desired)
	for i := range desiredTTL {
		er := existingTTL[i]
		dr := desiredTTL[i]

		m := fmt.Sprintf("CHANGE %s %s ", dr.NameFQDN, dr.Type) + humanDiff(er, dr)

		instructions = append(instructions, mkChange(dr.NameFQDN, dr.Type, []string{m},
			models.Records{er},
			models.Records{dr},
		))
	}

	// the common chunk are changes (regardless of TTL)
	mi := min(len(existing), len(desired))
	//fmt.Printf("DEBUG: min=%d\n", mi)
	for i := 0; i < mi; i++ {
		//fmt.Println(i, "CHANGE")
		er := existing[i].rec
		dr := desired[i].rec

		m := fmt.Sprintf("CHANGE %s %s ", dr.NameFQDN, dr.Type) + humanDiff(existing[i].rec, desired[i].rec)

		instructions = append(instructions, mkChange(dr.NameFQDN, dr.Type, []string{m},
			models.Records{er},
			models.Records{dr},
		))
	}

	// any left-over existing are deletes
	for i := mi; i < len(existing); i++ {
		//fmt.Println(i, "DEL")
		er := existing[i].rec
		m := fmt.Sprintf("DELETE %s %s %s", er.NameFQDN, er.Type, er.GetTargetCombined())
		instructions = append(instructions, mkDeleteRec(er.NameFQDN, er.Type, []string{m}, er))
	}

	// any left-over desired are creates
	for i := mi; i < len(desired); i++ {
		//fmt.Println(i, "CREATE")
		dr := desired[i].rec
		m := fmt.Sprintf("CREATE %s %s %s", dr.NameFQDN, dr.Type, dr.GetTargetCombined())
		instructions = append(instructions, mkAdd(dr.NameFQDN, dr.Type, []string{m}, models.Records{dr}))
	}

	return instructions
}

func genmsgs(existing, desired []targetConfig) []string {
	cl := diffTargets(existing, desired)
	return justMsgs(cl)
}

func justMsgs(cl ChangeList) []string {
	var msgs []string
	for _, c := range cl {
		msgs = append(msgs, c.Msgs...)
	}
	return msgs
}

func justMsgString(cl ChangeList) string {
	msgs := justMsgs(cl)
	return strings.Join(msgs, "\n")
}
