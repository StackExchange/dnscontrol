package diff2

//go:generate stringer -type=Verb

// This module provides functions that "diff" the existing records
// against the desired records.

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
)

type Verb int

const (
	_      Verb = iota // Skip the first value of 0
	CREATE             // Create a record/recordset/label where none existed before.
	CHANGE             // Change existing record/recordset/label
	DELETE             // Delete existing record/recordset/label
	REPORT             // No change, but boy do I have something to say!
)

type ChangeList []Change

type Change struct {
	Type Verb // Add, Change, Delete

	Key        models.RecordKey // .Key.Type is "" unless using ByRecordSet
	Old        models.Records
	New        models.Records                // any changed or added records at Key.
	Msgs       []string                      // Human-friendly explanation of what changed
	MsgsJoined string                        // strings.Join(Msgs, "\n")
	MsgsByKey  map[models.RecordKey][]string // Messages for a given key
}

/*
General instructions:

  changes, err := diff2.ByRecord(existing, dc, nil)
  //changes, err := diff2.ByRecordSet(existing, dc, nil)
  //changes, err := diff2.ByLabel(existing, dc, nil)
  if err != nil {
    return nil, err
  }

  var corrections []*models.Correction

  for _, change := range changes {
    switch change.Type {
    case diff2.REPORT:
      corr = &models.Correction{Msg: change.MsgsJoined}
    case diff2.CREATE:
      corr = &models.Correction{
        Msg: change.MsgsJoined,
        F: func() error {
          return c.createRecord(FILL_IN)
        },
      }
    case diff2.CHANGE:
      corr = &models.Correction{
        Msg: change.MsgsJoined,
        F: func() error {
          return c.modifyRecord(FILL_IN)
        },
      }
    case diff2.DELETE:
      corr = &models.Correction{
        Msg: change.MsgsJoined,
        F: func() error {
          return c.deleteRecord(FILL_IN)
        },
      }
	  default:
		  panic("unhandled change.TYPE %s", change.Type)
    }

    corrections = append(corrections, corr)
  }

  return corrections, nil
}


*/

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
	return byHelper(analyzeByRecordSet, existing, dc, compFunc)
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
	return byHelper(analyzeByLabel, existing, dc, compFunc)
}

// ByRecord takes two lists of records (existing and desired) and
// returns instructions for turning existing into desired.
//
// Use this with DNS providers whose API updates one record at a time.
//
// Note: The .Old and .New field are lists ([]models.RecordConfig) but
// when using ByRecord() they will never have more than one entry.
// A create always has exactly 1 new: .New[0]
// A change always has exactly 1 old and 1 new: .Old[0] and .New[0]
// A delete always has exactly 1 old: .Old[0]
//
// Examples include: INWX
func ByRecord(existing models.Records, dc *models.DomainConfig, compFunc ComparableFunc) (ChangeList, error) {
	return byHelper(analyzeByRecord, existing, dc, compFunc)
}

// ByZone takes two lists of records (existing and desired) and
// returns text to output to users describing the change, a bool
// indicating if there were any changes, and a possible err value.
//
// Use this with DNS providers whose API updates the entire zone at a
// time. That is, to make any change (even just 1 record) the entire DNS
// zone is uploaded.
//
// The user should see a list of changes as if individual records were updated.
//
// Example usage:
//
//	msgs, changes, err := diff2.ByZone(foundRecords, dc, nil)
//	if err != nil {
//	  return nil, err
//	}
//
//	if changes {
//		// Generate a "correction" that uploads the entire zone.
//		// (dc.Records are the new records for the zone).
//	}
//
// Example providers include: BIND
func ByZone(existing models.Records, dc *models.DomainConfig, compFunc ComparableFunc) ([]string, bool, error) {

	if len(existing) == 0 {
		// Nothing previously existed. No need to output a list of individual changes.
		return nil, true, nil
	}

	// Only return the messages.
	instructions, err := byHelper(analyzeByRecord, existing, dc, compFunc)
	return justMsgs(instructions), len(instructions) != 0, err
}

//

// byHelper does 90% of the work for the By*() calls.
func byHelper(fn func(cc *CompareConfig) ChangeList, existing models.Records, dc *models.DomainConfig, compFunc ComparableFunc) (ChangeList, error) {

	// Process NO_PURGE/ENSURE_ABSENT and UNMANAGED/IGNORE_*.
	desired, msgs, err := handsoff(
		dc.Name,
		existing, dc.Records, dc.EnsureAbsent,
		dc.Unmanaged,
		dc.UnmanagedUnsafe,
		dc.KeepUnknown,
	)
	if err != nil {
		return nil, err
	}

	// Regroup existing/desiredd for easy comparison:
	cc := NewCompareConfig(dc.Name, existing, desired, compFunc)

	// Analyze and generate the instructions:
	instructions := fn(cc)

	// If we have msgs, create a change to output them:
	if len(msgs) != 0 {
		chg := Change{
			Type:       REPORT,
			Msgs:       msgs,
			MsgsJoined: strings.Join(msgs, "\n"),
		}
		_ = chg
		instructions = append([]Change{chg}, instructions...)
	}

	return instructions, nil
}

// Stringify the datastructures for easier debugging

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
