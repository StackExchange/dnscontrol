package diff2

// import (
// 	"github.com/StackExchange/dnscontrol/v3/models"
// 	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
// )

// // Provide an interface that is backwards compatible with pkg/diff.

// // // /diff/IncrementalDiffCompat(). It not as efficient as converting
// // to the diff2.By*() functions. However, using this is better than
// // staying with pkg/diff.
// func CompatIncrementalDiff(dc *models.DomainConfig, existing []*models.RecordConfig) (unchanged, create, toDelete, modify diff.Changeset, err error) {
// 	unchanged = diff.Changeset{}
// 	create = diff.Changeset{}
// 	toDelete = diff.Changeset{}
// 	modify = diff.Changeset{}

// 	instructions, err := ByRecord(existing, dc, nil)
// 	if err != nil {
// 		return nil, nil, nil, nil, err
// 	}

// 	d := diff.New(dc, nil)

// 	for _, inst := range instructions {
// 		//cor := Correlation{d: d}
// 		cor := diff.Correlation{d: d}
// 		switch inst.Type {
// 		case CREATE:
// 			cor.Desired = inst.New[0]
// 			create = append(create, cor)
// 		case CHANGE:
// 			cor.Existing = inst.Old[0]
// 			cor.Desired = inst.New[0]
// 			modify = append(modify, cor)
// 		case DELETE:
// 			cor.Existing = inst.Old[0]
// 			toDelete = append(toDelete, cor)
// 		}
// 	}

// 	return
// }

// // func (d *Differ) ChangedGroups(existing []*models.RecordConfig) (map[models.RecordKey][]string, error) {
// // 	changedKeys := map[models.RecordKey][]string{}
// // 	_, create, toDelete, modify, err := d.IncrementalDiff(existing)
// // 	if err != nil {
// // 		return nil, err
// // 	}
// // 	for _, c := range create {
// // 		changedKeys[c.Desired.Key()] = append(changedKeys[c.Desired.Key()], c.String())
// // 	}
// // 	for _, d := range toDelete {
// // 		changedKeys[d.Existing.Key()] = append(changedKeys[d.Existing.Key()], d.String())
// // 	}
// // 	for _, m := range modify {
// // 		changedKeys[m.Desired.Key()] = append(changedKeys[m.Desired.Key()], m.String())
// // 	}
// // 	return changedKeys, nil
// // }

// // func (c Correlation) String() string {
// // 	if c.Existing == nil {
// // 		return fmt.Sprintf("CREATE %s %s %s", c.Desired.Type, c.Desired.GetLabelFQDN(), c.d.content(c.Desired))
// // 	}
// // 	if c.Desired == nil {
// // 		return fmt.Sprintf("DELETE %s %s %s", c.Existing.Type, c.Existing.GetLabelFQDN(), c.d.content(c.Existing))
// // 	}
// // 	return fmt.Sprintf("MODIFY %s %s: (%s) -> (%s)", c.Existing.Type, c.Existing.GetLabelFQDN(), c.d.content(c.Existing), c.d.content(c.Desired))
// // }

// // // get normalized content for record. target, ttl, mxprio, and specified metadata
// // func (d *Differ) content(r *models.RecordConfig) string {

// // 	// get the extra values maps to add to the comparison.
// // 	var allMaps []map[string]string
// // 	for _, f := range d.extraValues {
// // 		valueMap := f(r)
// // 		allMaps = append(allMaps, valueMap)
// // 	}

// // 	return r.ToDiffable(allMaps...)
// // }
