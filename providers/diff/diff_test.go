package diff

import (
	"fmt"
	"strings"
	"testing"

	"github.com/miekg/dns/dnsutil"
)

type myRecord string //@ A 1 1.2.3.4

func (m myRecord) GetName() string {
	name := strings.SplitN(string(m), " ", 4)[0]
	return dnsutil.AddOrigin(name, "example.com")
}
func (m myRecord) GetType() string {
	return strings.SplitN(string(m), " ", 4)[1]
}
func (m myRecord) GetContent() string {
	return strings.SplitN(string(m), " ", 4)[3]
}
func (m myRecord) GetComparisionData() string {
	return fmt.Sprint(strings.SplitN(string(m), " ", 4)[2])
}

func TestAdditionsOnly(t *testing.T) {
	desired := []Record{
		myRecord("@ A 1 1.2.3.4"),
	}
	existing := []Record{}
	checkLengths(t, existing, desired, 0, 1, 0, 0)
}

func TestDeletionsOnly(t *testing.T) {
	existing := []Record{
		myRecord("@ A 1 1.2.3.4"),
	}
	desired := []Record{}
	checkLengths(t, existing, desired, 0, 0, 1, 0)
}

func TestModification(t *testing.T) {
	existing := []Record{
		myRecord("www A 1 1.1.1.1"),
		myRecord("@ A 1 1.2.3.4"),
	}
	desired := []Record{
		myRecord("@ A 32 1.2.3.4"),
		myRecord("www A 1 1.1.1.1"),
	}
	un, _, _, mod := checkLengths(t, existing, desired, 1, 0, 0, 1)
	if t.Failed() {
		return
	}
	if un[0].Desired != desired[1] || un[0].Existing != existing[0] {
		t.Error("Expected unchanged records to be correlated")
	}
	if mod[0].Desired != desired[0] || mod[0].Existing != existing[1] {
		t.Errorf("Expected modified records to be correlated")
	}
}

func TestUnchangedWithAddition(t *testing.T) {
	existing := []Record{
		myRecord("www A 1 1.1.1.1"),
	}
	desired := []Record{
		myRecord("www A 1 1.2.3.4"),
		myRecord("www A 1 1.1.1.1"),
	}
	un, _, _, _ := checkLengths(t, existing, desired, 1, 1, 0, 0)
	if un[0].Desired != desired[1] || un[0].Existing != existing[0] {
		t.Errorf("Expected unchanged records to be correlated")
	}
}

func TestOutOfOrderRecords(t *testing.T) {
	existing := []Record{
		myRecord("www A 1 1.1.1.1"),
		myRecord("www A 1 2.2.2.2"),
		myRecord("www A 1 3.3.3.3"),
	}
	desired := []Record{
		myRecord("www A 1 1.1.1.1"),
		myRecord("www A 1 2.2.2.2"),
		myRecord("www A 1 2.2.2.3"),
		myRecord("www A 10 3.3.3.3"),
	}
	_, _, _, mods := checkLengths(t, existing, desired, 2, 1, 0, 1)
	if mods[0].Desired != desired[3] || mods[0].Existing != existing[2] {
		t.Fatalf("Expected to match %s and %s, but matched %s and %s", existing[2], desired[3], mods[0].Existing, mods[0].Desired)
	}

}

func checkLengths(t *testing.T, existing, desired []Record, unCount, createCount, delCount, modCount int) (un, cre, del, mod Changeset) {
	un, cre, del, mod = IncrementalDiff(existing, desired)
	if len(un) != unCount {
		t.Errorf("Got %d unchanged records, but expected %d", len(un), unCount)
	}
	if len(cre) != createCount {
		t.Errorf("Got %d records to create, but expected %d", len(cre), createCount)
	}
	if len(del) != delCount {
		t.Errorf("Got %d records to delete, but expected %d", len(del), delCount)
	}
	if len(mod) != modCount {
		t.Errorf("Got %d records to modify, but expected %d", len(mod), modCount)
	}
	return
}
