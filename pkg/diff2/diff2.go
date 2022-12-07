package diff2

//go:generate stringer -type=Verb

// This module provides functions that "diff" the existing records
// against the desired records.

import (
	"bytes"
	"fmt"

	"github.com/StackExchange/dnscontrol/v3/models"
)

type Verb int

const (
	COMMENT Verb = iota // No change. Just a txt message to output.
	CREATE              // Create a record/recordset/label where none existed before.
	CHANGE              // Change existing record/recordset/label
	DELETE              // Delete existing record/recordset/label
)

type ChangeList []Change

type Change struct {
	Type Verb // Add, Change, Delete

	Key  models.RecordKey // .Key.Type is "" unless using ByRecordSet
	Old  models.Records
	New  models.Records // any changed or added records at Key.
	Msgs []string       // Human-friendly explanation of what changed
}

// ByRecordSet takes two lists of records (existing and desired) and
// returns instructions for turning existing into desired.
//
// Use this with DNS providers whose API updates one recordset at a
// time. A recordset is all the records of a particular type at a
// label. For example, if www.example.com has 3 A records and a TXT
// record, if A records are added, changed, or removed, the API takes
// www.example.com, A, and a list of all the desired IP addresses.
//
// Examples include:
func ByRecordSet(existing, desired models.Records,
	origin string, compFunc ComparableFunc) (ChangeList, error) {

	existing = handsoff(existing, desired)
	// There will need to be error checking eventually.
	// if err != nil {
	// 	return nil, err
	// }

	cc := NewCompareConfig(origin, existing, desired, compFunc)
	instructions := analyzeByRecordSet(cc)

	return processPurge(instructions, true), nil
}

//// ByRecord takes two lists of records (existing and desired) and
//// returns instructions for turning existing into desired.
////
//// Use this with DNS providers whose API updates one record at a time.
////
//// Examples include: INWX
//func ByRecord(existing, desired models.Records) (instructions ChangeList, err error) {
//	existing = handsoff(existing, desired)
//
//	instructions, err := analyzeByRecord(existing, desired)
//	if err != nil {
//		return nil, err
//	}
//
//	return processPurge(instructions, config.NoPurge)
//}
//
//// ByLabel takes two lists of records (existing and desired) and
//// returns instructions for turning existing into desired.
////
//// Use this with DNS providers whose API updates one label at a
//// time. That is, updates are done by sending a list of DNS records
//// to be served at a particular label, or the label itself is deleted.
////
//// Examples include:
//func ByLabel(existing, desired models.Records) (instructions CHangeList, err error) {
//	existing = handsoff(existing, desired)
//
//	instructions, err := analyzeByLabel(existing, desired)
//	if err != nil {
//		return nil, err
//	}
//
//	return processPurge(instructions, config.NoPurge)
//}
//
//// ByZone takes two lists of records (existing and desired) and
//// returns text one would output to users describing the change.
////
//// Use this with DNS providers whose API updates the entire zone at a
//// time. That is, to make any change (1 record or many) the entire DNS
//// zone is uploaded.
////
//// The user should see a list of changes as if individual records were
//// updated.  However, as an optimization, if existing is empty, we
//// just output a 1-line message such as:
////
////	WRITING ZONEFILE: zones/example.com.zone
////
//// Example providers include: BIND
//func ByZone(existing, desired models.Records, firstMsg string) (report []string, err error) {
//	// Short-circuit if we are creating the zone from scratch:
//	if len(existing) == 0 {
//		// TODO(tlim): If no_purge is set, output a warning that this may
//		// be dangerous.
//		return []string{firstMsg}, nil
//	}
//
//	existing = handsoff(existing, desired)
//
//	instructions, err := analyzeByZone(existing, desired)
//	if err != nil {
//		return nil, err
//	}
//
//	return processPurge(instructions, config.NoPurge)
//}
//

func (c Change) String() string {
	var buf bytes.Buffer
	b := &buf

	fmt.Fprintf(b, "Change: verb=%v\n", c.Type)
	fmt.Fprintf(b, "    key=%v\n", c.Key)
	if len(c.Old) != 0 {
		fmt.Fprintf(b, "    old=%v\n", c.Old)
	}
	if len(c.New) != 0 {
		fmt.Fprintf(b, "    new=%v\n", c.New)
	}
	fmt.Fprintf(b, "    msg=%q\n", c.Msgs)

	return b.String()
}

func (cl ChangeList) String() string {
	var buf bytes.Buffer
	b := &buf

	fmt.Fprintf(b, "ChangeList: len=%d\n", len(cl))
	for i, j := range cl {
		fmt.Fprintf(b, "%02d: %s", i, j)
	}

	return b.String()
}
