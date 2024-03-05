package diff2

// type recset struct {
// 	Key  models.RecordKey
// 	Recs []*models.RecordConfig
// }

// func groupbyRSet(recs models.Records, origin string) []recset {

// 	if len(recs) == 0 {
// 		return nil
// 	}

// 	// Sort the NameFQDN to a consistent order. The actual sort methodology
// 	// doesn't matter as long as equal values are adjacent.
// 	// Use the PrettySort ordering so that the records are extra pretty.
// 	pretty := prettyzone.PrettySort(recs, origin, 0, nil)
// 	recs = pretty.Records

// 	var result []recset
// 	var acc []*models.RecordConfig

// 	// Do the first element
// 	prevkey := recs[0].Key()
// 	acc = append(acc, recs[0])

// 	for i := 1; i < len(recs); i++ {
// 		curkey := recs[i].Key()
// 		if prevkey == curkey { // A run of equal keys.
// 			acc = append(acc, recs[i])
// 		} else { // New key. Append old data to result and start new acc.
// 			result = append(result, recset{Key: prevkey, Recs: acc})
// 			acc = []*models.RecordConfig{recs[i]}
// 		}
// 		prevkey = curkey
// 	}
// 	result = append(result, recset{Key: prevkey, Recs: acc}) // The remainder

// 	return result
// }
