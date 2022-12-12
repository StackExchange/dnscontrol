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
	_      Verb = iota // Skip the first value of 0
	CREATE             // Create a record/recordset/label where none existed before.
	CHANGE             // Change existing record/recordset/label
	DELETE             // Delete existing record/recordset/label
)

type ChangeList []Change

type Change struct {
	Type Verb // Add, Change, Delete

	Key       models.RecordKey // .Key.Type is "" unless using ByRecordSet
	Old       models.Records
	New       models.Records                // any changed or added records at Key.
	Msgs      []string                      // Human-friendly explanation of what changed
	MsgsByKey map[models.RecordKey][]string // Messages for a given key
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
func ByRecordSet(existing models.Records, dc *models.DomainConfig, compFunc ComparableFunc) (ChangeList, error) {
	// dc stores the desired state.

	desired := dc.Records
	var err error
	desired, err = handsoff(existing, desired, dc.Unmanaged, dc.UnmanagedUnsafe) // Handle UNMANAGED()
	if err != nil {
		return nil, err
	}

	cc := NewCompareConfig(dc.Name, existing, desired, compFunc)
	instructions := analyzeByRecordSet(cc)
	return processPurge(instructions, !dc.KeepUnknown), nil
}

// ByLabel takes two lists of records (existing and desired) and
// returns instructions for turning existing into desired.
//
// Use this with DNS providers whose API updates one label at a
// time. That is, updates are done by sending a list of DNS records
// to be served at a particular label, or the label itself is deleted.
//
// Examples include:
func ByLabel(existing models.Records, dc *models.DomainConfig, compFunc ComparableFunc) (ChangeList, error) {
	// dc stores the desired state.

	desired := dc.Records
	var err error
	desired, err = handsoff(existing, desired, dc.Unmanaged, dc.UnmanagedUnsafe) // Handle UNMANAGED()
	if err != nil {
		return nil, err
	}

	cc := NewCompareConfig(dc.Name, existing, desired, compFunc)
	instructions := analyzeByLabel(cc)
	return processPurge(instructions, !dc.KeepUnknown), nil
}

// ByRecord takes two lists of records (existing and desired) and
// returns instructions for turning existing into desired.
//
// Use this with DNS providers whose API updates one record at a time.
//
// Examples include: INWX
func ByRecord(existing models.Records, dc *models.DomainConfig, compFunc ComparableFunc) (ChangeList, error) {
	// dc stores the desired state.

	desired := dc.Records
	var err error
	desired, err = handsoff(existing, desired, dc.Unmanaged, dc.UnmanagedUnsafe) // Handle UNMANAGED()
	if err != nil {
		return nil, err
	}

	cc := NewCompareConfig(dc.Name, existing, desired, compFunc)
	instructions := analyzeByRecord(cc)
	return processPurge(instructions, !dc.KeepUnknown), nil
}

// ByZone takes two lists of records (existing and desired) and
// returns text one would output to users describing the change.
//
// Use this with DNS providers whose API updates the entire zone at a
// time. That is, to make any change (1 record or many) the entire DNS
// zone is uploaded.
//
// The user should see a list of changes as if individual records were
// updated.
//
// The caller of this function should:
//
//	changed, msgs := diff2.ByZone(existing, desired, origin, nil
//		fmt.Sprintf("CREATING ZONEFILE FOR THE FIRST TIME: dir/example.com.zone"))
//	if changed {
//		// output msgs
//		// generate the zone using the "desired" records
//	}
//
// Example providers include: BIND
func ByZone(existing models.Records, dc *models.DomainConfig, compFunc ComparableFunc) ([]string, bool, error) {
	// dc stores the desired state.

	if len(existing) == 0 {
		// Nothing previously existed. No need to output a list of individual changes.
		return nil, true, nil
	}

	desired := dc.Records
	var err error
	desired, err = handsoff(existing, desired, dc.Unmanaged, dc.UnmanagedUnsafe) // Handle UNMANAGED()
	if err != nil {
		return nil, false, err
	}

	cc := NewCompareConfig(dc.Name, existing, desired, compFunc)
	instructions := analyzeByRecord(cc)
	instructions = processPurge(instructions, !dc.KeepUnknown)
	return justMsgs(instructions), len(instructions) != 0, nil
}

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
